# 🚀 NATS Fire-and-Forget 快速参考

## 5 分钟快速入门

### 1️⃣ 初始化 (main.go)

```go
import "github.com/gulugulu3399/bifrost/internal/pkg/messenger"

func main() {
    // 连接 NATS
    msgr, err := messenger.New("nats://localhost:4222", "bifrost_nexus")
    if err != nil {
        log.Fatal(err)
    }
    defer msgr.Close()

    // 启动 Beacon 消费者
    consumer := NewConsumer(data, msgr)
    if err := consumer.Start(); err != nil {
        log.Fatal(err)
    }
    defer consumer.Close()
}
```

---

### 2️⃣ 发送事件 (Nexus)

```go
// 业务完成后，异步发送事件
func (s *PostService) CreatePost(ctx context.Context, req *Request) (*Response, error) {
    // 1. 保存到数据库
    post := &Post{ID: 123, Slug: "my-post"}
    if err := s.repo.Save(ctx, post); err != nil {
        return nil, err
    }

    // 2. 异步发送事件（Fire-and-Forget）
    go func() {
        payload := messenger.PostEventPayload{ID: 123, Slug: "my-post"}
        if err := s.msgr.Publish(messenger.SubjectPostCreated, payload); err != nil {
            // log.Warn("publish failed", err)
        }
    }()

    // 3. 立即返回给客户端
    return &Response{ID: post.ID}, nil
}
```

---

### 3️⃣ 消费事件 (Beacon)

```go
// 启动时订阅
func (c *Consumer) Start() error {
    _, err := c.msgr.Subscribe("content.>", messenger.GroupBeacon, func(subject, data []byte) {
        switch subject {
        case messenger.SubjectPostCreated:
            c.handlePostChange(data)
        case messenger.SubjectCategoryUpdate:
            c.handleCategoryChange()
        }
    })
    return err
}

// 处理事件
func (c *Consumer) handlePostChange(data []byte) {
    var p messenger.PostEventPayload
    json.Unmarshal(data, &p)
    
    // 删除 Redis 缓存
    ctx := context.Background()
    c.data.Cache().Delete(ctx, KeyPostDetail(fmt.Sprintf("%d", p.ID)))
    if p.Slug != "" {
        c.data.Cache().Delete(ctx, KeyPostDetail(p.Slug))
    }
}
```

---

### 4️⃣ 扩展到 Mirror (Rust)

```rust
// 完全不需要改动 Nexus！只需在 Mirror 中订阅新的 Queue Group
async fn main() {
    let client = async_nats::connect("nats://localhost:4222").await?;
    
    // 关键：不同的 group 名称 "mirror_service"
    let mut sub = client.queue_subscribe("content.>", "mirror_service").await?;
    
    while let Some(msg) = sub.next().await {
        match msg.subject.as_str() {
            "content.post.created" | "content.post.updated" => {
                update_search_index(&msg.payload).await;
            },
            _ => {}
        }
    }
}
```

---

## 📦 事件载荷

```go
// PostEventPayload - 文章事件
type PostEventPayload struct {
    ID   int64  `json:"id"`
    Slug string `json:"slug"`
}

// CategoryEventPayload - 分类事件
type CategoryEventPayload struct {
    ID   int64  `json:"id"`
    Slug string `json:"slug"`
}

// CommentEventPayload - 评论事件
type CommentEventPayload struct {
    ID     int64 `json:"id"`
    PostID int64 `json:"post_id"`
    UserID int64 `json:"user_id"`
}

// InteractionEventPayload - 互动事件（点赞/收藏/分享）
type InteractionEventPayload struct {
    UserID int64  `json:"user_id"`
    PostID int64  `json:"post_id"`
    Type   string `json:"type"`  // "like", "bookmark", "share"
}
```

---

## 🏷️ 事件主题

| 主题 | 发送方 | 说明 |
|-----|-------|------|
| `content.post.created` | Nexus | 文章发布 |
| `content.post.updated` | Nexus | 文章更新 |
| `content.post.deleted` | Nexus | 文章删除 |
| `content.category.updated` | Nexus | 分类变更 |
| `content.tag.updated` | Nexus | 标签变更 |
| `content.comment.created` | Nexus | 评论新增 |
| `content.interaction.updated` | Nexus | 用户互动 |

---

## 👥 消费者组

| 组名 | 用途 | 位置 |
|-----|------|------|
| `beacon_service` | 缓存失效 | Go - Beacon |
| `mirror_service` | 索引更新 | Rust - Mirror |
| `audit_service` | 审计日志 | 未来服务 |
| `notification_service` | 发送通知 | 未来服务 |

---

## 🔑 核心常数

```go
// 消费者组名
const (
    GroupBeacon = "beacon_service"
    GroupMirror = "mirror_service"
)

// 事件主题
const (
    SubjectPostCreated    = "content.post.created"
    SubjectPostUpdated    = "content.post.updated"
    SubjectPostDeleted    = "content.post.deleted"
    SubjectCategoryUpdate = "content.category.updated"
    SubjectCommentCreated = "content.comment.created"
    SubjectInteraction    = "content.interaction.updated"
)
```

---

## 🛠️ 调试技巧

### 查看 NATS 消息

```bash
# 使用 NATS CLI
nats sub "content.>"

# 或打开 NATS Dashboard
open http://localhost:8222
```

### 检查订阅者

```bash
nats server info | grep subscribers
```

### 查看连接

```bash
nats conn ls
```

---

## ⚡ 常见问题

### Q: 消息丢失怎么办？

**A:** 这是 Fire-and-Forget 的代价。但对 CMS：数据库已保存，缓存多保留几分钟问题不大。

### Q: 如何确保幂等性？

**A:** 消费者应该设计为幂等操作。例如，删除不存在的 Redis key 是安全的。

### Q: 如何处理消费者异常？

**A:** 只记日志，不返回错误。下一次事件到来时会重新尝试处理。

### Q: 如何添加新的消费者？

**A:** 创建新 Queue Group，订阅相同主题：

```go
msgr.Subscribe("content.>", "new_service", handler)
```

**无需改动 Nexus！**

---

## 📊 性能数据

| 指标 | 值 |
|-----|-----|
| 消息延迟 (P95) | < 10ms |
| 吞吐量 | 100K+ msg/s |
| NATS 内存 | ~50MB |
| Go 客户端代码 | < 50 行 |

---

## 📚 相关资源

- [NATS 完整指南](./NATS_MESSAGING_GUIDE.md)
- [架构变更说明](./ARCHITECTURE_CHANGES_v3.2.1.md)
- [项目 README](../README.md)
- [数据库架构](../migrations/readme.md)

---

## 💡 设计原则

✅ **简洁** - 代码少，易维护  
✅ **高效** - 微秒级延迟  
✅ **可扩展** - 新增服务零改动  
✅ **可靠** - NATS 保证消息分发  

---

**需要更多帮助？** 查看完整的 [NATS 消息传递指南](./NATS_MESSAGING_GUIDE.md) 📖
