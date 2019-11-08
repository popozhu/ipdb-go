# ipdb-go
[![TravisCI Build Status](https://travis-ci.org/ipipdotnet/ipdb-go.svg?branch=master)](https://travis-ci.org/ipipdotnet/ipdb-go)
[![Coverage Status](https://coveralls.io/repos/github/ipipdotnet/ipdb-go/badge.svg?branch=master)](https://coveralls.io/github/ipipdotnet/ipdb-go?branch=master)
[![IPDB Database API Document](https://godoc.org/github.com/ipipdotnet/ipdb-go?status.svg)](https://godoc.org/github.com/ipipdotnet/ipdb-go)

IPIP.net officially supported IP database ipdb format parsing library

# Installing
<code>
    go get github.com/popozhu/ipdb-go
</code>

# Code Example

## 支持IPDB格式地级市精度IP离线库(免费版，每周高级版，每日标准版，每日高级版，每日专业版，每日旗舰版)
<pre>
<code>
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    
    ipdb "github.com/popozhu/ipdb-go"
)

func usage() {
    fmt.Fprintf(os.Stderr, "usage: \n\t%s [17monipdb_v6_vipday2.ipdb]\n\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(2)
}

func main() {
    if len(os.Args) == 1 {
            usage()
    }

    ipdb_file := os.Args[1]

    db, err := ipdb.NewCity(ipdb_file)
    if err != nil {
            log.Fatal(err)
    }

    dumper := ipdb.NewDumper(db)
    dumper.DumpNodes("CN")
    return
}

</code>
</pre>
## 地级市精度库数据字段说明
<pre>
country_name : 国家名字 （每周高级版及其以上版本包含）
region_name  : 省名字   （每周高级版及其以上版本包含）
city_name    : 城市名字 （每周高级版及其以上版本包含）
owner_domain : 所有者   （每周高级版及其以上版本包含）
isp_domain  : 运营商 （每周高级版与每日高级版及其以上版本包含）
latitude  :  纬度   （每日标准版及其以上版本包含）
longitude : 经度    （每日标准版及其以上版本包含）
timezone : 时区     （每日标准版及其以上版本包含）
utc_offset : UTC时区    （每日标准版及其以上版本包含）
china_admin_code : 中国行政区划代码 （每日标准版及其以上版本包含）
idd_code : 国家电话号码前缀 （每日标准版及其以上版本包含）
country_code : 国家2位代码  （每日标准版及其以上版本包含）
continent_code : 大洲代码   （每日标准版及其以上版本包含）
idc : IDC |  VPN   （每日专业版及其以上版本包含）
base_station : 基站 | WIFI （每日专业版及其以上版本包含）
country_code3 : 国家3位代码 （每日专业版及其以上版本包含）
european_union : 是否为欧盟成员国： 1 | 0 （每日专业版及其以上版本包含）
currency_code : 当前国家货币代码    （每日旗舰版及其以上版本包含）
currency_name : 当前国家货币名称    （每日旗舰版及其以上版本包含）
anycast : ANYCAST       （每日旗舰版及其以上版本包含）
</pre>
## 适用于IPDB格式的中国地区 IPv4 区县库
<pre>
db, err := ipdb.NewDistrict("/path/to/quxian.ipdb")
if err != nil {
	log.Fatal(err)
}
fmt.Println(db.IsIPv4())    // check database support ip type
fmt.Println(db.IsIPv6())    // check database support ip type
fmt.Println(db.Languages()) // database support language
fmt.Println(db.Fields())    // database support fields

fmt.Println(db.Find("1.12.7.255", "CN"))
fmt.Println(db.FindMap("2001:250:200::", "CN"))
fmt.Println(db.FindInfo("1.12.7.255", "CN"))

fmt.Println()
</pre>

## 适用于IPDB格式的基站 IPv4 库
<pre>
db, err := ipdb.NewBaseStation("/path/to/station_ip.ipdb")
if err != nil {
	log.Fatal(err)
}

fmt.Println(db.FindMap("223.220.223.255", "CN"))
</pre>
