// 华为云ECS数据结构
package huaweiyun

type ECS struct {
	Servers      []Servers      `json:"servers"`
	ServersLinks []ServersLinks `json:"servers_links"`
}
type ServersLinks struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}
type Servers struct {
	Id               string                 `json:"id"`   // instancdid
	Name             string                 `json:"name"` //主机名
	AvailabilityZone string                 `json:"availability_zone"`
	Addresses        map[string][]IPAddress `json:"addresses"`
	MetaData         MetaData               `json:"metadata"` //主要是系统名称 ubuntu/centos
	Flavor           Flavor                 `json:"flavor"`
	MarketInfo       MarketInfo             `json:"market_info"` //到期时间
}
type IPAddress struct {
	Addr         string `json:"addr"`
	OSEXTIPStype string `json:"OS-EXT-IPS:type"`
}

type MetaData struct {
	ImageName string `json:"image_name"`
	VpcID     string `json:"vpc_id"` //此处可以用于获取Address
}
type Flavor struct {
	Name  string `json:"name"`  // m7.xlarge.8
	Vcpus int    `json:"vcpus"` // cpu
	Ram   int    `json:"ram"`   //ram
}

// 付费信息：到期时间
type MarketInfo struct {
	PrepaidInfo PrepaidInfo `json:"prepaid_info"`
}
type PrepaidInfo struct {
	ExpiredTime string `json:"expired_time"`
}
