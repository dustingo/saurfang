# 单元测试文档

## 概述

本项目为 `DeployHandler` 创建了全面的单元测试，覆盖了各种场景和边界情况。

## 测试文件

- `internal/handler/taskhandler/deploy_test.go` - DeployHandler 的单元测试

## 测试覆盖场景

### 1. 成功场景测试
- **TestDeployHandler_CreateDeployTask_Success**: 测试正常创建部署任务
  - 验证正确的请求参数处理
  - 验证数据源查询
  - 验证任务创建
  - 验证成功响应

### 2. 错误场景测试
- **TestDeployHandler_CreateDeployTask_InvalidID**: 测试无效ID参数
  - 验证非数字ID的错误处理
  - 验证400状态码返回

- **TestDeployHandler_CreateDeployTask_DataSourceNotFound**: 测试数据源不存在
  - 验证数据源查询失败的处理
  - 验证错误消息返回

- **TestDeployHandler_CreateDeployTask_InvalidBody**: 测试无效请求体
  - 验证JSON解析错误处理
  - 验证400状态码返回

- **TestDeployHandler_CreateDeployTask_CreateTaskError**: 测试创建任务失败
  - 验证数据库错误处理
  - 验证500状态码返回

### 3. 边界情况测试
- **TestDeployHandler_CreateDeployTask_EmptyPayload**: 测试空负载
  - 验证空JSON对象的处理
  - 验证默认值设置

### 4. 构造函数测试
- **TestNewDeployHandler**: 测试构造函数
  - 验证对象正确创建
  - 验证依赖注入

## 运行测试

### 运行所有测试
```bash
go test ./... -v
```

### 运行特定测试
```bash
go test ./internal/handler/taskhandler -v
```

### 运行特定测试函数
```bash
go test ./internal/handler/taskhandler -v -run TestDeployHandler_CreateDeployTask_Success
```

### 在Windows上运行
```bash
test.bat
```

## 测试依赖

- `github.com/stretchr/testify/assert` - 断言库
- `github.com/stretchr/testify/mock` - 模拟库
- `github.com/gofiber/fiber/v3` - Web框架
- `net/http/httptest` - HTTP测试工具

## 模拟对象

测试使用了以下模拟对象：

1. **MockDeployService**: 模拟部署服务
   - `Create(task *task.SaurfangPublishtasks) error`

2. **MockDataSourceService**: 模拟数据源服务
   - `Service_ShowDataSourceByID(id uint) (*datasource.SaurfangDatasources, error)`

## 测试数据

测试使用了以下测试数据结构：

```go
// 请求参数
payload := task.PublishTaskParams{
    Become:     1,
    BecomeUser: "testuser",
    Comment:    "test comment",
}

// 期望的数据源
expectedDataSource := &datasource.SaurfangDatasources{
    ID:    1,
    Label: "test-label",
}

// 期望的任务
expectedTask := task.SaurfangPublishtasks{
    ID:          1,
    SourceLabel: "test-label",
    Become:      1,
    BecomeUser:  "testuser",
    Comment:     "test comment",
}
```

## 测试最佳实践

1. **隔离性**: 每个测试都是独立的，不依赖其他测试
2. **可重复性**: 测试可以在任何环境下重复运行
3. **快速性**: 测试运行速度快，不依赖外部服务
4. **覆盖率**: 覆盖了主要的成功和失败场景
5. **可读性**: 测试代码清晰易懂，有详细的注释

## 扩展测试

如果需要添加更多测试，可以考虑：

1. **性能测试**: 测试高并发场景
2. **集成测试**: 测试与真实数据库的交互
3. **API测试**: 测试完整的API端点
4. **安全测试**: 测试输入验证和安全边界

## 故障排除

如果测试失败，请检查：

1. 依赖是否正确安装
2. 导入路径是否正确
3. 模拟对象设置是否正确
4. 测试数据是否匹配实际代码逻辑 