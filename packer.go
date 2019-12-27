package ipdb

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Meta struct {
	Build     int64          `json:"build"`
	IpVersion uint16         `json:"ip_version"`
	Languages map[string]int `json:"languages"`
	Fields    []string       `json:"fields"`
	NodeCount int            `json:"node_count"`
	TotalSize int            `json:"total_size"`
}

type Packer struct {
	nodes     [][2]int
	data      []byte
	data_hash map[string]int
	data_size int
	meta      Meta

	debug             bool
	null_fields_value []string
}

func NewPacker(is_ipv6 bool) *Packer {
	var ip_version uint16 = 0x03
	if false == is_ipv6 {
		ip_version = 0x01
	}

	nodes := make([][2]int, 1)
	nodes[0] = [2]int{-1, -1}

	p := &Packer{
		nodes:     nodes,
		data_size: 0,
		data_hash: make(map[string]int),
		meta: Meta{
			Build:     time.Now().Unix(),
			IpVersion: ip_version,
			Languages: map[string]int{
				"CN": 0,
			},
		},

		debug: false,
	}

	return p
}

func (p *Packer) SetDebug(is_debug bool) {
	p.debug = is_debug
}

func (p *Packer) SetFields(fields []string, null_fields_value []string) {
	if len(fields) == 0 || len(fields) != len(null_fields_value) {
		log.Fatalf("invalid fields or null_fields_value: fields item and fields values are not aligned\n")
	}
	p.meta.Fields = fields

	p.null_fields_value = null_fields_value
}

func (p *Packer) Insert(cidr string, data []string) int {
	if len(p.meta.Fields) == 0 {
		log.Fatalf("empty fields items, use SetFields first\n")
	}

	node, bit := p.getNode(cidr)
	offset := p.getOffset(data)

	p.nodes[node][bit] = -offset
	if p.debug {
		fmt.Printf("cidr[%s], node idx[%d], set bit[%d], and the node is %+v\n", cidr, node, bit, p.nodes[node])
	}

	return len(p.nodes)
}

func (p *Packer) Output() []byte {

	node_count := len(p.nodes)
	p.meta.NodeCount = node_count
	if p.debug {
		fmt.Printf("do output, node count: %d\n", node_count)
	}

	// fix null node
	null_count := 0
	null_offset := p.getOffset(p.null_fields_value)
	for i, _ := range p.nodes {
		for j := 0; j < 2; j++ {
			if p.nodes[i][j] == -1 {
				null_count++
				p.nodes[i][j] = -null_offset
			}

			if p.nodes[i][j] <= 0 {
				p.nodes[i][j] = node_count - p.nodes[i][j]
			}
		}
	}
	if p.debug {
		fmt.Printf("null_offset: %d, total null_count is %d\n", null_offset, null_count)
	}

	// node chunk
	node_chunk := make([]byte, node_count*8)
	for i := 0; i < len(p.nodes); i++ {
		if p.debug {
			fmt.Printf("%d, node: %+v\n", i, p.nodes[i])
		}

		node := p.nodes[i]
		off := i * 8
		binary.BigEndian.PutUint32(node_chunk[off:off+4], uint32(node[0]))
		binary.BigEndian.PutUint32(node_chunk[off+4:off+8], uint32(node[1]))
	}
	p.meta.TotalSize = len(node_chunk) + len(p.data)

	// header chunk
	header, err := json.Marshal(p.meta)
	if err != nil || len(header) == 0 {
		log.Fatalf("json encode Meta[%+v] error: %s\n", p.meta, err)
	}
	header_chunk := make([]byte, 4)
	binary.BigEndian.PutUint32(header_chunk, uint32(len(header)))

	// header + nodes + data
	db_chunk := append(header_chunk, header...)
	db_chunk = append(db_chunk, node_chunk...)
	db_chunk = append(db_chunk, p.data...)

	if p.debug {
		fmt.Printf("header: %s, len: %d\n", header, len(header))
		fmt.Printf("node count: %d, len: %d\n", node_count, len(node_chunk))
		fmt.Printf("data len: %d\n", len(p.data))
	}

	return db_chunk
}

func (p *Packer) getNode(cidr string) (int, int) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Fatalf("invalid cidr [%s], parse error: %s\n", cidr, err)
	}

	ip := ipnet.IP
	ones, _ := ipnet.Mask.Size()
	if p.debug {
		fmt.Printf("cidr[%s], ip[%s], mask len[%d]\n", cidr, ip, ones)
	}

	idx := 0
	for i := 0; i < ones-1; i++ {
		bit := ((0xFF & int(ip[i>>3])) >> uint(7-i%8)) & 1

		if p.nodes[idx][bit] == -1 {
			p.nodes = append(p.nodes, [2]int{-1, -1})
			p.nodes[idx][bit] = len(p.nodes) - 1
		}
		//fmt.Printf("\tbit %d, val:%d, current idx %d, node: %+v\n", i, bit, idx, p.nodes[idx])

		idx = p.nodes[idx][bit]
	}

	i := ones - 1
	last_bit := ((0xFF & int(ip[i>>3])) >> uint(7-i%8)) & 1

	if p.debug {
		fmt.Printf("\tget node for cidr[%s], ip[%s], mask len[%d], node idx[%d], last bit val[%d]\n", cidr, ip, ones, idx, last_bit)
	}

	return idx, last_bit
}

func (p *Packer) getOffset(data []string) int {
	// ip信息，字符串
	str := strings.Join(data, "\t")
	if offset_in_data, ok := p.data_hash[str]; ok {
		return offset_in_data
	}

	// ip信息，字节
	data_bytes := []byte(str)

	// ip信息的长度
	len_bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(len_bytes, uint16(len(data_bytes)))

	p.data_hash[str] = p.data_size
	p.data = append(p.data, len_bytes...)
	p.data = append(p.data, data_bytes...)
	p.data_size += len(len_bytes) + len(data_bytes)

	if p.debug {
		fmt.Printf("append data[%s], offset: %d, and offset forward to [%d]\n", str, p.data_hash[str], p.data_size)
	}

	return p.data_hash[str]
}
