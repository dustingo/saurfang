# Saurfang

一个基于 Go Fiber 框架的游戏运维管理平台，集成了 Nomad、Consul、Redis 等现代化基础设施组件。

## 🚀 功能特性

### 核心功能
- **游戏服务器管理**: 支持游戏服务器配置、部署和监控
- **CMDB 资产管理**: 主机、分组、渠道等基础设施管理
- **任务调度系统**: 基于 Nomad 的执行
- **自定义任务**: 支持 Python 和 Shell 脚本的自定义任务执行
- **用户权限管理**: 完整的用户认证和权限控制系统
- **监控面板**: 实时资源监控和统计图表
- **API支持**: 支持使用Token调用API

### 技术栈
- **Web 框架**: Go Fiber v3
- **数据库**: MySQL + GORM
- **缓存**: Redis
- **任务调度**: HashiCorp Nomad
- **服务发现**: HashiCorp Consul
- **异步任务**: Asynq
- **云服务**: OSS 支持华为云，阿里云
- **配置管理**: 环境变量 + .env

### Nomad 任务调度说明
Nomad 是一个分布式的任务调度工具，但在游戏运维场景中需要特别注意以下配置：

#### 节点约束 (Node Constraints)
- **目的**: 游戏程序通常需要指定在特定服务器上运行
- **实现**: 通过 `constraint` 配置限制任务只能在指定的节点上运行
- **约束条件**: 以节点的 `hostname` 为主要约束条件
- **示例**: 
  ```hcl
  constraint {
    attribute = "${attr.unique.hostname}"
    operator  = "regexp"
    value     = "(server1|server2|server3)"
  }
  ```

#### 业务逻辑分组 (Group)
- **设计理念**: 每个 `group` 代表一个独立的业务逻辑单元
- **作用**: 便于管理、监控和扩展相关的任务组
- **示例**: 游戏服务器组、数据库组、监控组等

#### 资源限制 (Resources)
- **必要性**: 在 `driver = "raw_exec"` 情况下，`resources` 配置是必须的
- **限制行为**: 当业务超出资源限制时，Nomad 不会主动做出响应或限制
- **建议**: 根据实际业务需求合理设置 CPU 和内存限制
- **示例**:
  ```hcl
  resources {
    cpu    = 500    # 0.5 CPU 核心
    memory = 512    # 512MB 内存
  }
  ```

## 📋 系统要求

- Go 1.23.8+
- MySQL 8.0+
- Redis 6.0+
- Nomad 1.0+
- Consul 1.0+

## 🛠️ 安装配置

### 1. 克隆项目
```bash
git clone <repository-url>
cd saurfang_v2_fiber
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 环境配置
复制环境变量模板并配置：
```bash
cp env.example .env
```

编辑 `.env` 文件，配置以下必要参数：

```env
# 应用配置
APP_PORT=8080
APP_TRUST_PROXY=127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=saurfang

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Consul配置
CONSUL_HOST=localhost
CONSUL_PORT=8500

# Nomad配置
NOMAD_HOST=localhost
NOMAD_PORT=4646

# 其他配置
SERVER_PACKAGE_SRC_PATH=/path/to/source
SERVER_PACKAGE_DEST_PATH=/path/to/destination
GAME_NOMAD_JOB_NAMESPACE=game
```

### 4. 数据库迁移
```bash
go run main.go --migrate
```

### 5. 启动服务
```bash
go run main.go --serve
```

#### Nomad Job 配置示例
```hcl
job "custom-task-123" {
  datacenters = ["dc1"]
  type = "batch"
  
  # 节点约束 - 指定在特定主机上运行
  constraint {
    attribute = "${attr.unique.hostname}"
    operator  = "regexp"
    value     = "(game-server-01|game-server-02)"
  }
  
  group "game-server-group" {
    count = 1
    
    task "game-server-task" {
      driver = "raw_exec"
      
      config {
        command = "/usr/bin/python3"
        args = ["-c", "print('Hello from game server')"]
      }
      
      # 资源限制 - raw_exec 驱动下必须配置
      resources {
        cpu    = 500    # 0.5 CPU 核心
        memory = 512    # 512MB 内存
      }
      
      # 超时配置
      kill_timeout = "30s"
    }
  }
}
```

## 🚀 部署

### Docker 部署
```bash
# 构建镜像
docker build -t saurfang:v2 .

# 运行容器
docker run -d \
  --name saurfang \
  -p 8080:8080 \
  --env-file .env \
  saurfang:v2
```

### 生产环境部署
1. 确保所有依赖服务（MySQL、Redis、Nomad、Consul）正常运行
2. 配置生产环境的环境变量
3. 使用进程管理器（如 systemd）管理应用
4. 配置反向代理（如 Nginx）

## 🧪 测试

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/handler/...

# 运行测试并生成覆盖率报告
go test -cover ./...
```

### 集成测试
```bash
# 运行集成测试
go test -tags=integration ./...
```

## 🤝 贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🆘 支持

如果您遇到问题或有疑问，请：

1. 查看 [Issues](../../issues) 页面
2. 创建新的 Issue 描述问题
3. 联系项目维护者

## 🔄 更新日志

### v2.0.0 (最新)
- 🎉 重大架构重构
- 🔧 移除服务层，直接使用仓储模式
- 🚀 集成 HashiCorp Nomad 和 Consul
- 📊 添加自定义任务执行系统
- 🔐 改进用户认证和权限管理
- 📈 优化性能和可扩展性

---

**Saurfang  Fiber** - 现代化的游戏运维管理平台 🎮