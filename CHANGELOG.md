# Changelog

本文档记录了 Saurfang V2 Fiber 项目的所有重要更改。

## [Unreleased]

### Added
- ✨ Consul 多节点集群支持与自动故障转移
- ✨ Nomad 多节点集群支持与自动故障转移
- ✨ 增强的日志记录和错误追踪
- ✨ 自动化发布脚本（Windows BAT）
- 📝 完整的 Consul 配置指南和环境变量示例

### Changed
- 🔧 优化 Consul 客户端初始化逻辑
- 🔧 改进 Nomad 连接重试机制
- 🔧 增强连接健康检查功能
- 📚 更新部署文档和配置说明

### Fixed
- 🐛 修复 Consul 节点故障后全局变量无法自动更新的问题
- 🐛 修复 Nomad 轮询重连失败的问题
- 🐛 修复多个 linter 警告和代码质量问题

## [2.0.0] - 2024-XX-XX

### Added
- ✨ 基于 Fiber v3 全新重构
- ✨ 新增自定义任务系统
- ✨ 增强的权限管理系统
- ✨ 改进的监控面板
- ✨ 操作消息通知订阅

### Changed
- 🔄 Web 框架从 Fiber v2 升级到 v3
- 🔄 优化数据库连接池配置

### Fixed
- 🐛 修复任务调度时区问题
- 🐛 修复文件上传大小限制

---

更多详细信息请查看 [GitHub Releases](https://github.com/YOUR_USERNAME/saurfang/releases)

