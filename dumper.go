package ipdb

import (
	"fmt"
	"net"
	"os"
	"strings"
	"unsafe"
)

type node_t struct {
	l int
	r int
}

type Dumper struct {
	reader *reader
	nodes  []node_t

	start net.IP
	end   net.IP
	inode int

	language string
}

//Inc increases the IP by one this returns a new []byte for the IP
func Inc(IP net.IP) net.IP {
	incIP := make([]byte, len(IP))
	copy(incIP, IP)
	for j := len(incIP) - 1; j >= 0; j-- {
		incIP[j]++
		if incIP[j] > 0 {
			break
		}
	}
	return incIP
}

//Dec decreases the IP by one this returns a new []byte for the IP
func Dec(IP net.IP) net.IP {
	decIP := make([]byte, len(IP))
	copy(decIP, IP)
	for j := len(decIP) - 1; j >= 0; j-- {
		decIP[j]--
		if decIP[j] < 255 {
			break
		}
	}
	return decIP
}

func Dupip(ip net.IP) net.IP {
	dup := make([]byte, net.IPv6len)
	copy(dup, ip)
	return dup
}
func ClearBit(ip net.IP, whichbit int) {
	ip[whichbit>>3] &= ^(1 << uint(7-whichbit%8))
}
func SetBit(ip net.IP, whichbit int) {
	ip[whichbit>>3] |= (1 << uint(7-whichbit%8))
}

func (dumper *Dumper) get_ip_info_by_inode(inode int, language string) ([]string, error) {
	db := dumper.reader

	off, ok := db.meta.Languages[language]
	if !ok {
		return nil, ErrNoSupportLanguage
	}

	body, err := db.resolve(inode)
	if err != nil {
		return nil, err
	}

	str := (*string)(unsafe.Pointer(&body))
	tmp := strings.Split(*str, "\t")

	if (off + len(db.meta.Fields)) > len(tmp) {
		return nil, ErrDatabaseError
	}

	return tmp[off : off+len(db.meta.Fields)], nil
}

func (dumper *Dumper) dumpnode(bits, index, inode int) {
	end := Dupip(dumper.end)

	if index == 0 {
		ClearBit(end, bits)
	} else {
		SetBit(end, bits)
	}
	for i := bits + 1; i < 128; i++ {
		SetBit(end, i)
	}

	if dumper.inode == inode {
		dumper.end = end
		return
	}

	//fmt.Fprintf(os.Stdout, "bits: %d --> dump cidr: [%s - %s], inode: %d\n", bits, dumper.start, end, inode)

	arr, _ := dumper.get_ip_info_by_inode(inode, dumper.language)
	fmt.Fprintf(os.Stdout, "%s\t%s\t%s\n", dumper.start, end, strings.Join(arr, "\t"))

	dumper.inode = inode
	dumper.start = Inc(end)
}

func (dumper *Dumper) preOrder(idx int, bits int) {

	db := dumper.reader

	if idx > db.nodeCount {
		fmt.Fprintf(os.Stderr, "idx exception, out of index: %d, max is %d\n", idx, db.nodeCount)
		os.Exit(-1)
	}
	if bits >= 128 {
		fmt.Fprintf(os.Stderr, "bits exception, out of bits, current: %d\n", bits)
		os.Exit(-1)
	}

	node := dumper.nodes[idx]

	//fmt.Fprintf(os.Stdout, "idx: %d, bits: %d, node: %d, left<-\n", idx, bits, node)
	if node.l < db.nodeCount {
		ClearBit(dumper.end, bits)

		dumper.preOrder(node.l, bits+1)
	} else {
		dumper.dumpnode(bits, 0, node.l)
	}

	//fmt.Fprintf(os.Stdout, "idx: %d, bits: %d, node: %d, ->right\n", idx, bits, node)
	if node.r < db.nodeCount {
		SetBit(dumper.end, bits)

		dumper.preOrder(node.r, bits+1)
	} else {
		dumper.dumpnode(bits, 1, node.r)
	}
}

func (dumper *Dumper) DumpNodes(language string) {
	db := dumper.reader
	//fmt.Fprintf(os.Stdout, "total nodeCode: %d\n", db.nodeCount)

	dumper.nodes = make([]node_t, db.nodeCount)
	for i := 0; i < db.nodeCount; i++ {
		dumper.nodes[i] = node_t{
			l: db.readNode(i, 0),
			r: db.readNode(i, 1),
		}
	}

	dumper.language = language
	dumper.start = net.IPv6loopback
	dumper.end = make([]byte, net.IPv6len)
	dumper.inode = 0
	dumper.preOrder(0, 0)

	return

}

func (dumper *Dumper) PrintNodeData(inode int, language string) {
	db := dumper.reader

	off, ok := db.meta.Languages[language]
	if !ok {
		fmt.Fprintf(os.Stderr, "unexpected language: %s\n", language)
		os.Exit(-1)
	}

	body, err := db.resolve(inode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve node[%d] failed\n", inode)
		os.Exit(-1)
	}

	str := (*string)(unsafe.Pointer(&body))
	tmp := strings.Split(*str, "\t")

	if (off + len(db.meta.Fields)) > len(tmp) {
		fmt.Fprintf(os.Stderr, "data offset too long: %+v\n", off)
		os.Exit(-1)
	}

	data := tmp[off : off+len(db.meta.Fields)]

	info := make(map[string]string, len(db.meta.Fields))
	for k, v := range data {
		info[db.meta.Fields[k]] = v
	}

	fmt.Println(inode, info)
}
