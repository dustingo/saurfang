# Saurfang V2 Fiber

一个基于 Go Fiber 框架的现代化游戏运维管理平台，集成了 HashiCorp Nomad、Consul、Redis 等云原生基础设施组件，为游戏服务器提供全生命周期的运维管理解决方案。

## 🚀 项目概述

Saurfang V2 Fiber 是一个专为游戏运维场景设计的综合性管理平台，提供了从基础设施管理到任务调度、从用户权限控制到监控告警的完整运维工具链。

## ✨ 核心功能

### 🎮 游戏服务器管理
- **游戏服务器配置**: 支持多游戏、多渠道的服务器配置管理
- **逻辑服务器管理**: 游戏逻辑服务器的创建、配置和监控
- **游戏频道管理**: 多渠道游戏发布和管理
- **游戏组管理**: 服务器分组和批量操作
- **服务器配置**: 动态配置管理和热更新

### 🖥️ CMDB 资产管理
- **主机管理**: 云服务器实例的自动发现和管理
- **分组管理**: 主机分组和标签管理
- **自动同步**: 支持阿里云、华为云等云平台的资源自动同步
- **资产监控**: 实时监控主机状态和资源使用情况

### ⚡ 任务调度系统
- **自定义任务**: 支持 Python、Shell、Bash 脚本的自定义任务
- **定时任务**: 基于 Cron 表达式的定时任务调度
- **部署任务**: 游戏服务器的自动化部署和更新
- **任务模板**: 可复用的脚本模板库
- **执行监控**: 任务执行状态和日志的实时监控

### 👥 用户权限管理
- **用户认证**: JWT Token 认证机制
- **角色管理**: 基于角色的权限控制 (RBAC)
- **权限分组**: 细粒度的权限管理
- **邀请码系统**: 用户注册邀请机制

### 📊 监控与告警
- **实时监控**: 系统资源和服务状态监控
- **统计面板**: 可视化的数据统计和图表
- **通知系统**: 支持多种通知渠道 (钉钉、邮件等)
- **告警规则**: 自定义告警规则和阈值

### 🔧 系统管理
- **配置管理**: 集中化的系统配置管理
- **文件上传**: 支持游戏资源文件的上传和管理
- **数据源管理**: 多数据源的统一管理
- **API 接口**: 完整的 RESTful API 支持

## 🛠️ 技术栈

### 后端技术
- **Web 框架**: [Go Fiber v3](https://gofiber.io/) - 高性能的 Go Web 框架
- **数据库**: MySQL + [GORM](https://gorm.io/) - 对象关系映射
- **缓存**: [Redis](https://redis.io/) - 内存数据库
- **任务队列**: [Asynq](https://github.com/hibiken/asynq) - 异步任务处理

### 基础设施
- **任务调度**: [HashiCorp Nomad](https://www.nomadproject.io/) - 分布式任务调度器
- **服务发现**: [HashiCorp Consul](https://www.consul.io/) - 服务网格和配置管理
- **配置管理**: [HashiCorp Consul](https://www.consul.io/) - 分布式键值存储

### 云服务集成
- **阿里云**: ECS 实例管理和 OSS 对象存储
- **华为云**: ECS 实例管理和 OBS 对象存储
- **通知服务**: 钉钉、邮件等多种通知渠道

### 开发工具
- **依赖管理**: Go Modules
- **配置管理**: 环境变量 + .env 文件
- **日志系统**: 结构化日志记录
- **性能监控**: pprof 性能分析

## 📋 系统要求

### 运行环境
- **Go**: 1.23.8 或更高版本
- **MySQL**: 5.7 或更高版本
- **Redis**: 6.0 或更高版本

### 可选组件
- **HashiCorp Nomad**: 1.6+ (用于任务调度)
- **HashiCorp Consul**: 1.16+ (用于服务发现及配置管理)

## 🚀 快速开始

### 1. 克隆项目
```bash
git clone <repository-url>
cd saurfang_v2_fiber
```

### 2. 环境配置
复制环境变量模板并配置：
```bash
cp .env.examples .env
```

编辑 `.env` 文件，配置以下关键参数：
```env
# 数据库配置
MYSQL_DSN=user:password@tcp(localhost:3306)/saurfang?charset=utf8mb4&parseTime=True&loc=Local

# Redis 配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Nomad 配置
NOMAD_HTTP_API_ADDR=http://localhost:4646

# Consul 配置
CONSUL_HTTP_ADDR=http://localhost:8500

# JWT 密钥
JWT_SECRET=your-secret-key

# 服务端口
PORT=8080
```

### 3. 安装依赖
```bash
go mod download
```

### 4. 数据库初始化
确保 MySQL 服务运行，系统会自动创建所需的数据表。

### 5. 启动服务
```bash
# 开发模式
go run main.go

# 或使用构建脚本
.\build.bat
```

### 6. 访问系统
- **Web 界面**: http://localhost:8080
- **API 文档**: http://localhost:8080/api/docs (如果配置了)
- **性能监控**: http://localhost:8080/debug/pprof/

## 🏗️ 项目结构

```
saurfang_v2_fiber/
├── internal/                 # 内部包
│   ├── config/              # 配置管理
│   │   ├── mysql.go         # MySQL 配置
│   │   ├── nomad.go         # Nomad 配置
│   │   ├── consul.go        # Consul 配置
│   │   └── ...
│   ├── handler/             # HTTP 处理器
│   │   ├── userhandler/     # 用户管理
│   │   ├── gamehandler/     # 游戏管理
│   │   ├── taskhandler/     # 任务管理
│   │   ├── cmdbhandler/     # CMDB 管理
│   │   └── ...
│   ├── models/              # 数据模型
│   │   ├── user/            # 用户模型
│   │   ├── task/            # 任务模型
│   │   ├── gameserver/      # 游戏服务器模型
│   │   └── ...
│   ├── middleware/          # 中间件
│   │   └── auth.go          # 认证中间件
│   ├── route/               # 路由配置
│   │   ├── routes.go        # 路由注册
│   │   ├── user.go          # 用户路由
│   │   ├── game.go          # 游戏路由
│   │   └── ...
│   ├── repository/          # 数据访问层
│   │   └── base/            # 基础仓库
│   └── tools/               # 工具包
│       ├── pkg/             # 通用工具
│       ├── ntfy/            # 通知工具
│       └── ...
├── web/                     # 前端资源
├── signature/               # 签名工具
├── main.go                  # 程序入口
├── go.mod                   # Go 模块文件
├── build.bat                # 构建脚本
└── README.md                # 项目文档
```

## 🔧 配置说明

### Nomad 任务调度配置

Nomad 是本系统的核心任务调度组件，特别适用于游戏运维场景：

#### 节点约束 (Node Constraints)
游戏程序通常需要在特定服务器上运行，通过节点约束实现：
```hcl
constraint {
  attribute = "${attr.unique.hostname}"
  operator  = "regexp"
  value     = "(server1|server2|server3)"
}
```

#### 业务逻辑分组 (Group)
每个 `group` 代表一个独立的业务逻辑单元，便于管理和监控。

#### 资源限制 (Resources)
在 `driver = "raw_exec"` 模式下，资源配置是必需的：
```hcl
resources {
  cpu    = 500
  memory = 512
}
```

### 云服务配置

#### 阿里云配置
```env
ALIYUN_ACCESS_KEY_ID=your-access-key
ALIYUN_ACCESS_KEY_SECRET=your-secret-key
ALIYUN_REGION=cn-hangzhou
```

#### 华为云配置
```env
HUAWEI_ACCESS_KEY=your-access-key
HUAWEI_SECRET_KEY=your-secret-key
HUAWEI_REGION=cn-north-1
```

## 📚 API 文档

### 认证接口
- `POST /api/user/login` - 用户登录
- `POST /api/user/register` - 用户注册
- `POST /api/user/logout` - 用户登出

### 游戏管理接口
- `GET /api/game/servers` - 获取游戏服务器列表
- `POST /api/game/servers` - 创建游戏服务器
- `PUT /api/game/servers/:id` - 更新游戏服务器
- `DELETE /api/game/servers/:id` - 删除游戏服务器

### 任务管理接口
- `GET /api/tasks/custom` - 获取自定义任务列表
- `POST /api/tasks/custom` - 创建自定义任务
- `POST /api/tasks/custom/:id/execute` - 执行自定义任务
- `GET /api/tasks/executions/:id/status` - 获取任务执行状态

### CMDB 接口
- `GET /api/cmdb/hosts` - 获取主机列表
- `POST /api/cmdb/hosts/sync` - 同步云主机
- `GET /api/cmdb/groups` - 获取主机分组

## 🔍 监控和日志

### 性能监控
系统集成了 pprof 性能分析工具：
- CPU 分析: http://localhost:8080/debug/pprof/profile
- 内存分析: http://localhost:8080/debug/pprof/heap
- 协程分析: http://localhost:8080/debug/pprof/goroutine

### 日志系统
使用结构化日志记录，支持不同级别的日志输出：
```go
slog.Info("Server started", "port", 8080)
slog.Error("Database connection failed", "error", err)
```

## 🚀 部署指南

### Docker 部署 (推荐)
```bash
# 构建镜像
docker build -t saurfang-v2-fiber .

# 运行容器
docker run -d \
  --name saurfang \
  -p 8080:8080 \
  -e MYSQL_DSN="user:pass@tcp(mysql:3306)/saurfang" \
  -e REDIS_ADDR="redis:6379" \
  saurfang-v2-fiber
```

### 生产环境部署
1. 使用反向代理 (Nginx/Traefik)
2. 配置 HTTPS 证书
3. 设置数据库连接池
4. 配置日志轮转
5. 设置监控告警

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持与帮助

如果您在使用过程中遇到问题，可以通过以下方式获取帮助：

1. 查看项目文档和 FAQ
2. 提交 Issue 描述问题
3. 参与社区讨论

## 🔄 更新日志

### v2.0.0
- 基于 Fiber v3 重构
- 新增自定义任务系统
- 优化 Nomad 集成
- 增强权限管理
- 改进监控面板

---

**Saurfang V2 Fiber** - 让游戏运维更简单、更高效！