package response

// Response 是API的通用响应结构
type Response struct {
	//@name Status
	//@description 状态码
	//@type string
	Status string `json:"status"`
	//@name Message
	//@description 消息
	//@type string
	Message string `json:"message"`
	//@name Data
	//@description 数据
	//@type interface{}
	Data interface{} `json:"data,omitempty"`
}

// PagedResult 是分页查询的结果结构
type PagedResult struct {
	// @name Data
	// @description 数据
	// @type interface{}
	Data interface{} `json:"data"`
	// @name Total
	// @description 总数
	// @type int64
	Total int64 `json:"total"`
	// @name Page
	// @description 页码
	// @type int
	Page int `json:"page"`
	// @name PageSize
	// @description 每页数量
	// @type int
	PageSize int `json:"pageSize"`
}
