package amis

// ResourceStatistics 资源统计面板
type ResourceStatistics struct {
	Channels int64 `json:"channels"`
	Hosts    int64 `json:"hosts"`
	Groups   int64 `json:"groups"`
	Games    int64 `json:"games"`
	Users    int64 `json:"users"`
}
