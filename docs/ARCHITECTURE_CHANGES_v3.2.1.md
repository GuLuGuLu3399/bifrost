# 架构变更总结：v3.2.1 Outbox → NATS Fire-and-Forget

## 📅 变更时间

**2025 年 12 月 23 日**

## 🔄 变更内容

### 移除

- ❌ **Outbox 表** (`outbox_events`) - PostgreSQL 中的事务性发件箱表
- ❌ **Outbox Relayer** - 后台轮询器，维护 Outbox 表的状态
- ❌ **重试机制** - Outbox 的重试逻辑和错误追踪
- ❌ **分布式事务复杂性** - 两阶段提交、租约锁等

### 添加

- ✅ **NATS Fire-and-Forget** - 轻量级消息分发
- ✅ **Queue Groups** - 自动负载均衡的消费者组
- ✅ **Messenger 包** - 简洁的 NATS 客户端封装
- ✅ **事件载荷定义** - 统一的事件格式
- ✅ **消费者模式** - Beacon / Mirror / 未来服务的事件处理

## 🎯 设计原理

### 为什么从 Outbox 迁移到 NATS Fire-and-Forget？

| 维度 | Outbox 模式 | Fire-and-Forget |
|-----|---------|----------------|
| **复杂度** | 高（表+轮询+锁+重试） | 低（直接发送） |
| **数据库压力** | 高（维护表+索引） | 无 |
| **一致性** | 强（事务级） | 弱（事件可能丢失） |
| **适用场景** | 金融系统（账目对账） | CMS（缓存失效） |
| **代码行数** | 500+ | 50 |

### CMS 为什么不需要强一致性？

对于内容管理系统：

- 📝 **业务数据已持久化** → PostgreSQL 事务成功
- 📊 **缓存失效可以容忍延迟** → 用户多看几分钟旧文章问题不大
- 🔍 **索引更新可以异步进行** → 搜索结果晚一会儿更新可接受
- 💰 **不涉及账目对账** → 无金融级一致性要求

**结论**：宁可"错杀"（多删缓存），不可"放过"（缓存不一致） → Fire-and-Forget 完全满足需求。

---

## 📁 文件变更清单

### 新增文件

| 路径 | 说明 |
|-----|------|
| `go_services/internal/pkg/messenger/messenger.go` | NATS 客户端核心实现 |
| `go_services/internal/pkg/messenger/events.go` | 事件载荷定义 + 主题常量 |
| `go_services/internal/beacon/data/consumer.go` | Beacon 消费者（更新版本） |
| `go_services/internal/nexus/data/events.go` | Nexus 发布方法示例 |
| `docs/NATS_MESSAGING_GUIDE.md` | 完整的使用指南 |

### 修改文件

| 路径 | 变更 |
|-----|------|
| `migrations/readme.md` | 删除 Outbox 表描述，更新设计原则 |
| `README.md` | 更新架构图，修改分布式事务说明 |
| 数据库迁移脚本 | 取消 Outbox 表创建（v3.2.1+） |

### 删除文件

- 无（向后兼容：旧 Outbox 表可以保留或逐步清理）

---

## 🚀 使用示例

### 发送事件（Nexus 侧）

```go
// 业务完成后，异步发送事件（不阻塞主流程）
go func() {
    payload := messenger.PostEventPayload{
        ID:   post.ID,
        Slug: post.Slug,
    }
    if err := msgr.Publish(messenger.SubjectPostUpdated, payload); err != nil {
        log.Warn("publish failed", err)
    }
}()
```

### 消费事件（Beacon 侧）

```go
// 启动时订阅，使用 Queue Groups 自动负载均衡
sub, err := msgr.Subscribe("content.>", messenger.GroupBeacon, func(subject, data []byte) {
    switch subject {
    case messenger.SubjectPostUpdated:
        c.handlePostChange(data)
    }
})
```

### 扩展（Mirror / Audit 侧）

```rust
// Rust 中只需更改 Queue Group 名称，无需改动 Nexus
sub = client.queue_subscribe("content.>", "mirror_service").await?;
```

---

## ✅ 验证清单

- [ ] Messenger 包编译成功
- [ ] Beacon Consumer 正常启动
- [ ] 文章发布后事件正确发送
- [ ] Redis 缓存正确失效
- [ ] 多个 Beacon 副本不重复处理
- [ ] NATS 服务器日志无错误
- [ ] 数据库中不存在 Outbox 相关表

---

## 📖 相关文档

- [NATS 消息传递指南](./docs/NATS_MESSAGING_GUIDE.md)
- [数据库架构说明](./migrations/readme.md)
- [项目 README](./README.md)

---

## ⚠️ 注意事项

1. **不再保证消息投递**：极少数网络故障可能导致消息丢失，但这对 CMS 可接受
2. **异步发送**：使用 `go func()` 不阻塞主业务，缓存失效由后台处理
3. **幂等性**：消费者应该设计为幂等操作（多次删除同一缓存无害）
4. **可观测性**：利用 NATS Dashboard 或 CLI 调试消息流量

---

## 🎓 导师总结

> 这个方案为什么适合现在的你？
>
> 1. **代码量极少**：`messenger` 包不到 50 行代码。
> 2. **认知负担低**：没有 Outbox 表，没有 Relayer 协程，不需要懂 JetStream 的 ACK 机制。
> 3. **架构扩展性强**：虽然实现简单，但利用 **NATS Subject** 和 **Queue Group** 完美实现了微服务解耦。你以后加 `Audit`（审计服务）、`Notification`（通知服务），都只需要加一个新的 Group 订阅即可，Nexus 不需要改一行代码。
>
> 这就叫**"以简单御复杂"**。

---

## 🔗 后续步骤

1. **立即实施**：
   - [x] 更新 migrations/readme.md
   - [x] 创建 messenger 包
   - [x] 更新 Beacon consumer
   - [x] 创建使用指南

2. **近期计划**：
   - [ ] Rust Mirror 集成 NATS 消费（参考 Rust 伪代码）
   - [ ] 性能测试（吞吐量、延迟）
   - [ ] 灰度上线

3. **长期优化**：
   - [ ] 添加 Audit 服务
   - [ ] 添加 Notification 服务
   - [ ] 性能监控与告警
