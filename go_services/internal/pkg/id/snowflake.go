package id

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// Generator ID 生成器接口
type Generator interface {
	// GenerateInt64 生成 int64 类型的 ID
	GenerateInt64() int64
	// GenerateString 生成字符串类型的 ID
	GenerateString() string
	// ParseString 解析字符串 ID 为 snowflake.ID
	ParseString(id string) (snowflake.ID, error)
}

// SnowflakeGenerator 雪花算法 ID 生成器
type SnowflakeGenerator struct {
	node *snowflake.Node
}

var (
	// 默认全局生成器
	defaultGenerator Generator
	once             sync.Once
)

// NewSnowflakeGenerator 创建新的雪花算法生成器
func NewSnowflakeGenerator(nodeID int64) (*SnowflakeGenerator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, fmt.Sprintf("failed to create snowflake node with ID %d", nodeID))
	}

	return &SnowflakeGenerator{
		node: node,
	}, nil
}

// GenerateInt64 生成 int64 类型的 ID
func (g *SnowflakeGenerator) GenerateInt64() int64 {
	if g.node == nil {
		panic("snowflake generator not initialized")
	}
	return g.node.Generate().Int64()
}

// GenerateString 生成字符串类型的 ID
func (g *SnowflakeGenerator) GenerateString() string {
	if g.node == nil {
		panic("snowflake generator not initialized")
	}
	return g.node.Generate().String()
}

// ParseString 解析字符串 ID 为 snowflake.ID
func (g *SnowflakeGenerator) ParseString(id string) (snowflake.ID, error) {
	return snowflake.ParseString(id)
}

// InitDefaultGenerator 初始化默认全局生成器
func InitDefaultGenerator(nodeID int64) error {
	var err error
	once.Do(func() {
		gen, e := NewSnowflakeGenerator(nodeID)
		if e != nil {
			err = e
			return
		}
		defaultGenerator = gen
	})
	return err
}

// ensureDefaultGenerator 确保默认生成器已初始化
func ensureDefaultGenerator() Generator {
	if defaultGenerator == nil {
		// 安全后备机制：如果忘记初始化，使用默认节点ID=1
		if err := InitDefaultGenerator(1); err != nil {
			panic(fmt.Sprintf("failed to initialize default generator: %v", err))
		}
	}
	return defaultGenerator
}

// GenerateInt64 使用默认生成器生成 int64 ID
func GenerateInt64() int64 {
	return ensureDefaultGenerator().GenerateInt64()
}

// GenerateString 使用默认生成器生成字符串 ID
func GenerateString() string {
	return ensureDefaultGenerator().GenerateString()
}

// ParseString 使用默认生成器解析字符串 ID
func ParseString(id string) (snowflake.ID, error) {
	// 确保已初始化（虽然 ParseString 是静态方法，但保持行为一致）
	_ = ensureDefaultGenerator()
	return snowflake.ParseString(id)
}

// ExtractTimestamp 从雪花 ID 中提取时间戳
func ExtractTimestamp(id int64) time.Time {
	// 手动位运算提取，替代已弃用的 snowflake.ID.Time()
	// Timestamp = (ID >> (NodeBits + StepBits)) + Epoch
	shift := snowflake.NodeBits + snowflake.StepBits
	timestamp := (id >> shift) + snowflake.Epoch
	return time.UnixMilli(timestamp)
}

// ExtractNodeID 从雪花 ID 中提取节点 ID
func ExtractNodeID(id int64) int64 {
	// 手动位运算提取，替代已弃用的 snowflake.ID.Node()
	// NodeID = (ID >> StepBits) & ((1 << NodeBits) - 1)
	mask := int64((1 << snowflake.NodeBits) - 1)
	return (id >> snowflake.StepBits) & mask
}

// ExtractSequence 从雪花 ID 中提取序列号
func ExtractSequence(id int64) int64 {
	// 手动位运算提取，替代已弃用的 snowflake.ID.Step()
	// Sequence = ID & ((1 << StepBits) - 1)
	mask := int64((1 << snowflake.StepBits) - 1)
	return id & mask
}

// IsValidSnowflakeID 检查是否是有效的雪花 ID
func IsValidSnowflakeID(id string) bool {
	_, err := snowflake.ParseString(id)
	return err == nil
}
