# MIGRATIONS

数据库迁移脚本目录（PostgreSQL）。

## 文件顺序

- `001_identity_and_infra.sql`
- `002_content_core.sql`
- `003_interactions.sql`

## 执行建议

- 按序执行，不要跳号。
- 先在本地验证，再进入共享环境。
- 新增迁移使用递增编号并保持幂等检查。

## 关联文档

- [ARCHITECTURE](../docs/ARCHITECTURE.md)
- [EVENT_CONTRACT](../docs/EVENT_CONTRACT.md)
