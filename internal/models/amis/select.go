// Package amis 用于amis select选择菜单的数据结构
package amis

// AmisOptions amis select 使用
type AmisOptions struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

/*
   "options": [
            {
                "label": "正式",
                "value": 1
            },
            {
                "label": "测试",
                "value": 2
            }
        ]
*/

// AmisOptionsString amis 字符串选择菜单
type AmisOptionsString struct {
	Label      string `json:"label"`
	SelectMode string `json:"selectMode"`
	Value      string `json:"value"`
}

/*
   "options": [
            {
                "label": "ROG",
                "selectMode": "tree",
                "value": "阿里云-ROG"
            }
        ]
*/
