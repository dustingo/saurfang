# Signature 签名工具

### 签名算法

**算法**: HMAC-SHA256  
**编码**: 十六进制(Hex)编码  
**签名字符串格式**: `{accessKey}-{method}-{path}-{timestamp}`

### 输出格式

工具会输出三个HTTP头部字段：
- `X-Access-Key`: 访问密钥
- `X-Timestamp`: Unix时间戳
- `X-Signature`: 生成的签名

### 命令行参数

| 参数 | 简写 | 类型 | 必需 | 描述 |
|------|------|------|------|------|
| `--key` | `-key` | string | 是 | API访问密钥(Access Key) |
| `--secret` | `-secret` | string | 是 | API密钥(Secret Key) |
| `--request-path` | `-request-path` | string | 是 | 请求路径(非全路径，请参考主程序的路由组,如: /api/v1/cmdb) |
| `--request-method` | `-request-method` | string | 是 | HTTP请求方法(大写，如: GET, POST) |

### 使用示例

#### 基本用法

```bash
# 为GET请求生成签名
./signature --key="your-access-key" --secret="your-secret-key" --request-path="/api/v1/cmdb" --request-method="GET"
```

#### 输出示例

```
canonicaRequest: your-access-key-GET-/api/v1/users-1703123456
X-Access-Key: your-access-key
X-Timestamp: 1703123456
X-Signature: a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456
```

## 技术实现

### 核心算法

```go
func generateSignature(accessKey, secretKey, method, path string, timestamp string) (string, error) {
    // 1. 构建规范化请求字符串
    canonicalRequest := fmt.Sprintf(
        "%s-%s-%s-%s",
        accessKey,
        method,
        path,
        timestamp,
    )
    
    // 2. 使用HMAC-SHA256生成签名
    h := hmac.New(sha256.New, []byte(secretKey))
    h.Write([]byte(canonicalRequest))
    signature := hex.EncodeToString(h.Sum(nil))
    
    return signature, nil
}
```
### HTTP请求头设置

使用生成的签名时，需要在HTTP请求中设置以下头部：

```http
X-Access-Key: your-access-key
X-Timestamp: 1703123456
X-Signature: generated-signature-hash
```

**注意**: 请妥善保管您的密钥信息，避免在公共场所或版本控制系统中暴露敏感信息。