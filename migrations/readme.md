# MIGRATIONS

> PostgreSQL 数据库迁移脚本目录。

## 目录

- [迁移顺序](#迁移顺序)
- [执行建议](#执行建议)
- [相关文档](#相关文档)

## 迁移顺序

1. `001_identity_and_infra.sql`
2. `002_content_core.sql`
3. `003_interactions.sql`

## 执行建议

- 严格按顺序执行，不跳号。
- 先在本地验证，再进入共享环境。
- 新增迁移保持编号递增，并包含幂等保护。

## 相关文档

- [ARCHITECTURE](../docs/ARCHITECTURE.md)
- [EVENT_CONTRACT](../docs/EVENT_CONTRACT.md)
