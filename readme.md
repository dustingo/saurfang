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

### 3. 安装依赖
```bash
go mod tidy
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
- **性能监控**: http://localhost:8080/debug/pprof/(需要main.go中开启)

## 🚀 部署指南

### 二进制 部署 (推荐)
```bash
# 构建二进制
build.bat
go build -ldflags "-w -s"
```
```bash
# 生成表结构
./saurfang  --migrate
# 生成注册邀请码
./saurfang admin codegen
# 设置管理员权限
./saurfang admin set-perms
# 添加管理员
./saurfang admin set-admin --name xxx
# 启动服务
./saurfang --serve
```

### 生产环境部署
1. 使用反向代理 (Nginx/caddy)
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
- 增加了操作消息通知订阅

---

**Saurfang V2 Fiber** - 让游戏运维更简单、更高效！