// 阿里云ECS数据结构
package aliyun

// ECS 阿里云ECS
type ECS struct {
	Instances Instances `json:"Instances"`
}
type Instances struct {
	Instance []Instance `json:"Instance"`
}
type Instance struct {
	Memory          int             `json:"Memory"` //Mb
	Cpu             int             `json:"Cpu"`
	OSName          string          `json:"OSName"`
	ExpiredTime     string          `json:"ExpiredTime"`
	InstanceId      string          `json:"InstanceId"`
	VpcAttributes   VpcAttributes   `json:"VpcAttributes"`
	InstanceName    string          `json:"InstanceName"` // 类似于主机名
	PublicIpAddress PublicIpAddress `json:"PublicIpAddress"`
	InstanceType    string          `json:"InstanceType"` //ecs.ic5.xlarge
	RegionId        string          `json:"RegionId"`     //
}

// VpcAttributes 内网vpc
type VpcAttributes struct {
	PrivateIpAddress PrivateIpAddress `json:"PrivateIpAddress"`
}

// PrivateIpAddress 内网ip
type PrivateIpAddress struct {
	IPAddress []string `json:"IPAddress"`
}
type PublicIpAddress struct {
	IpAddress []string `json:"IpAddress"`
}
