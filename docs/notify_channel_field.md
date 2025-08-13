# 通知配置 Channel 字段说明

## 概述

在 `NotifyConfig` 结构体中新增了 `channel` 字段，用于标识通知配置属于哪种通知渠道类型。

## 字段定义

```go
type NotifyConfig struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"type:text;not null" json:"name,omitempty"`
    Channel   string    `gorm:"type:text;not null" json:"channel,omitempty"`  // 新增字段
    Config    string    `gorm:"type:text;not null" json:"config,omitempty"`
    Status    string    `gorm:"type:text;not null" json:"status,omitempty"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

## 支持的渠道类型

系统支持以下通知渠道类型：

| 渠道值 | 渠道名称 | 说明 |
|--------|----------|------|
| `dingtalk` | 钉钉 | 钉钉机器人通知 |
| `wechat` | 企业微信 | 企业微信应用通知 |
| `email` | 邮件 | 邮件通知 |
| `http` | HTTP | HTTP Webhook 通知 |
| `lark` | 飞书 | 飞书机器人通知 |

## 渠道常量定义

```go
const (
    ChannelDingTalk string = "dingtalk"
    ChannelWeChat   string = "wechat"
    ChannelEmail    string = "email"
    ChannelHTTP     string = "http"
    ChannelLark     string = "lark"
)
```

## 验证方法

### 渠道有效性验证

```go
// 检查渠道是否有效
func (nc *NotifyConfig) IsValidChannel() bool

// 使用示例
config := &NotifyConfig{Channel: "dingtalk"}
if config.IsValidChannel() {
    // 渠道有效
}
```

### 获取渠道显示名称

```go
// 获取渠道的显示名称
func (nc *NotifyConfig) GetChannelDisplayName() string

// 使用示例
config := &NotifyConfig{Channel: "dingtalk"}
displayName := config.GetChannelDisplayName() // 返回 "钉钉"
```

## API 接口

### 1. 创建通知渠道

**POST** `/api/v1/notify/channel/create`

请求体示例：
```json
{
    "name": "钉钉通知配置",
    "channel": "dingtalk",
    "config": "{\"token\": \"your_token\", \"secret\": \"your_secret\"}"
}
```

### 2. 更新通知渠道

**PUT** `/api/v1/notify/channel/update/:id`

请求体示例：
```json
{
    "name": "钉钉通知配置",
    "channel": "dingtalk",
    "config": "{\"token\": \"new_token\", \"secret\": \"new_secret\"}"
}
```

### 3. 根据渠道类型获取配置

**GET** `/api/v1/notify/channel/by-channel?channel=dingtalk`

响应示例：
```json
{
    "code": 0,
    "message": "Notify configs listed successfully",
    "data": [
        {
            "id": 1,
            "name": "钉钉通知配置",
            "channel": "dingtalk",
            "config": "{\"token\": \"your_token\", \"secret\": \"your_secret\"}",
            "status": "0"
        }
    ]
}
```

## 缓存支持

系统会自动为 `channel` 字段建立缓存索引，支持以下查询：

- 根据渠道类型获取配置列表
- 根据配置ID获取详细信息
- 根据名称获取配置
- 根据状态获取配置

## 配置示例

### 钉钉配置
```json
{
    "name": "钉钉通知配置",
    "channel": "dingtalk",
    "config": "{\"token\": \"dddd\", \"secret\": \"xxx\"}"
}
```

### 企业微信配置
```json
{
    "name": "企业微信通知配置",
    "channel": "wechat",
    "config": "{\"AppID\": \"abcdefghi\", \"AppSecret\": \"jklmnopqr\", \"Token\": \"mytoken\", \"EncodingAESKey\": \"IGNORED-IN-SANDBOX\", \"ToReceivers\": \"user1,user2\"}"
}
```

### 邮件配置
```json
{
    "name": "邮件通知配置",
    "channel": "email",
    "config": "{\"to\": \"admin@example.com,user@example.com\"}"
}
```

### HTTP配置
```json
{
    "name": "HTTP通知配置",
    "channel": "http",
    "config": "{\"URL\": \"http://localhost:8080/webhook\", \"Header\": {}, \"ContentType\": \"application/json\", \"Method\": \"POST\"}"
}
```

### 飞书配置
```json
{
    "name": "飞书通知配置",
    "channel": "lark",
    "config": "{\"webHookURL\": \"https://open.feishu.cn/open-apis/bot/v2/hook/your_webhook_url\"}"
}
```

## 注意事项

1. `channel` 字段为必填字段，不能为空
2. 只支持预定义的渠道类型，不支持自定义渠道
3. 创建和更新时会自动验证渠道类型的有效性
4. 缓存系统会自动为渠道字段建立索引，提高查询性能
5. 建议在生产环境中使用 HTTPS 和适当的认证机制 