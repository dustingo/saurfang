// 游戏服配置文件结构体
package serverconfig

// Node 树节点
type Node struct {
	Label      string  `json:"label"`
	SelectMode string  `json:"selectMode"`
	Value      string  `json:"value,omitempty"` // 仅在 IP 层次设置 Value
	Children   []*Node `json:"children,omitempty"`
}
type ServerListNode struct {
	Label      string      `json:"label"`
	SelectMode string      `json:"selectMode"`
	Value      string      `json:"value,omitempty"` // server name
	Children   []*Children `json:"children,omitempty"`
}
type Children struct {
	Label string `json:"label"`
	Value string `json:"value,omitempty"`
}

//游戏服配置文件结构体

type GameConfigs struct {
	Configs []Configs `json:"configs"`
}
type Configs struct {
	SvcName    string                 `json:"svc_name"`  // 服务名称
	ServerId   string                 `json:"server_id"` // 游戏服sercerid
	Prefix     string                 `json:"prefix"`    // game服务器于对象存储中的路径;如果下方"应用发布-数据源"中的路径(版本)不是根目录"/",那么prefix应该以"/"开头,如:/game
	ConfigDir  string                 `json:"config_dir"`
	ConfigFile string                 `json:"config_file,omitempty"`
	Start      string                 `json:"start"` //服务的启动命令
	Stop       string                 `json:"stop"`  // 服务的关闭命令
	Vars       map[string]interface{} `json:"vars,omitempty"`
	User       string                 `json:"user"`
	Port       int                    `json:"port"`
	IP         string                 `json:"ip"`
}

// GameConfig 游戏服配置文件结构体
type GameConfig struct {
	Key     string `json:"key"`
	Setting string `json:"setting"`
}
