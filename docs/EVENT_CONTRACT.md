# BIFROST EVENT CONTRACT

> 最后更新：2026-03-14

## 目录

- [主题约定](#主题约定)
- [Stream 与 Consumer](#stream-与-consumer)
- [消息载荷](#消息载荷)
- [Feature 开关](#feature-开关)

## 主题约定

- `content.post.created`
- `content.post.updated`
- `content.post.deleted`
- `content.post.>`
- `content.>`

## Stream 与 Consumer

- Stream: `BIFROST_CONTENT`
- Durable Consumer: `mirror_indexer`
- 默认过滤: `content.post.>`

说明：服务启动时会执行拓扑 ensure，不要求严格启动顺序。

## 消息载荷

### 标准示例

```json
{
  "id": 123,
  "slug": "hello-bifrost",
  "title": "Hello Bifrost",
  "summary": "A quick introduction",
  "status": 2,
  "published_at": 1760000000
}
```

### 字段定义

- `id`: int64，必填
- `slug`: string，建议提供
- `title`: string，建议提供（索引需要）
- `summary`: string，可选
- `status`: int32，可选（1=draft, 2=published, 3=archived）
- `published_at`: int64（Unix 秒），可选

### 最小兼容示例

```json
{
  "id": 123
}
```

说明：Mirror 对关键字段不足的消息会 ACK 并跳过索引，不阻塞消费。

## Feature 开关

| 维度 | 变量 | 默认值 | 说明 |
| --- | --- | --- | --- |
| Go | `BIFROST_FEATURES_ENABLE_MESSENGER` | `true` | 启用 NATS 发布/消费 |
| Go | `BIFROST_FEATURES_ENABLE_SEARCH` | `true` | 启用 Gjallar 搜索代理 |
| Go | `BIFROST_FEATURES_ENABLE_STORAGE` | `true` | 启用 Nexus 存储能力 |
| Rust(Mirror) | `APP_MIRROR__FEATURES__ENABLE_NATS_WORKER` | `true` | 启用 Mirror NATS Worker |
| Rust(Mirror) | `APP_MIRROR__NATS__FILTER_SUBJECT` | `content.post.>` | 订阅过滤 |
| Rust(Mirror) | `APP_MIRROR__NATS__STREAM_NAME` | `BIFROST_CONTENT` | Stream 名称 |
| Rust(Mirror) | `APP_MIRROR__NATS__CONSUMER_NAME` | `mirror_indexer` | Consumer 名称 |
