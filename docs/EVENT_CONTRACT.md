# BIFROST EVENT CONTRACT

> 最后更新：2026-03-14

本文档定义 Go 与 Rust 服务之间的 NATS 事件契约。

## 1. Subject 约定

- `content.post.created`
- `content.post.updated`
- `content.post.deleted`
- `content.post.>`（文章事件通配）
- `content.>`（内容通配）

## 2. Stream / Consumer

- Stream：`BIFROST_CONTENT`
- Mirror Durable Consumer：`mirror_indexer`
- 默认过滤：`content.post.>`

服务启动时会执行拓扑 ensure，不强依赖固定启动顺序。

## 3. 事件载荷

```json
{
  "id": 123,
  "slug": "hello-bifrost",
  "title": "Hello Bifrost",
  "summary": "A quick introduction",
  "status": "published",
  "published_at": 1760000000
}
```

字段说明：

- `id`：int64，必填
- `slug`：string，建议填
- `title`：string，建议填（Mirror 建索引需要）
- `summary`：string，可选
- `status`：string，可选（如 `draft` / `published` / `archived`）
- `published_at`：Unix 秒时间戳，可选

最小兼容载荷：

```json
{
  "id": 123
}
```

Mirror 在关键字段不足时会 ACK 并跳过，避免阻塞消费。

## 4. Feature 开关

| 维度 | 变量 | 默认值 | 说明 |
| --- | --- | --- | --- |
| Go | `BIFROST_FEATURES_ENABLE_MESSENGER` | `true` | 启用 NATS 发布/消费 |
| Go | `BIFROST_FEATURES_ENABLE_SEARCH` | `true` | 启用 Gjallar 搜索代理 |
| Go | `BIFROST_FEATURES_ENABLE_STORAGE` | `true` | 启用 Nexus 存储能力 |
| Rust(Mirror) | `APP_MIRROR__FEATURES__ENABLE_NATS_WORKER` | `true` | 启用 Mirror NATS Worker |
| Rust(Mirror) | `APP_MIRROR__NATS__FILTER_SUBJECT` | `content.post.>` | 订阅过滤 |
| Rust(Mirror) | `APP_MIRROR__NATS__STREAM_NAME` | `BIFROST_CONTENT` | Stream 名称 |
| Rust(Mirror) | `APP_MIRROR__NATS__CONSUMER_NAME` | `mirror_indexer` | Durable Consumer 名称 |

## 5. 建议

- 生产环境显式配置以上开关，避免默认值漂移。
- 事件 schema 变更时先更新本文件，再同步 Go/Rust 解析逻辑。
- 保持向后兼容：新增字段可选，避免删除既有字段。
