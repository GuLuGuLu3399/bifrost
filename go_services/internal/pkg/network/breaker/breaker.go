package breaker

import (
	"errors"
	"time"

	"github.com/sony/gobreaker"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// State 熔断器状态
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// String 返回状态的字符串表示
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config 熔断器配置
type Config struct {
	Name          string          `yaml:"name"`
	MaxRequests   uint32          `yaml:"max_requests"`
	Interval      time.Duration   `yaml:"interval"`
	Timeout       time.Duration   `yaml:"timeout"`
	ReadyToTrip   CountTripFunc   `yaml:"-"`
	IsSuccessful  SuccessFunc     `yaml:"-"`
	OnStateChange StateChangeFunc `yaml:"-"`
}

// CountTripFunc 判断是否应该打开熔断器的函数
type CountTripFunc func(counts gobreaker.Counts) bool

// SuccessFunc 判断请求是否成功的函数
type SuccessFunc func(err error) bool

// StateChangeFunc 状态变化回调函数
type StateChangeFunc func(name string, from, to State)

// DefaultConfig 返回默认配置
func DefaultConfig(name string) *Config {
	return &Config{
		Name:        name,
		MaxRequests: 1,
		Interval:    0, // 不周期性清零计数器
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 连续失败 5 次后打开熔断器
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: func(err error) bool {
			// 默认只要没有错误就算成功
			return err == nil
		},
		OnStateChange: func(name string, from, to State) {
			// 默认空实现
		},
	}
}

// CircuitBreaker 熔断器封装
type CircuitBreaker struct {
	cb     *gobreaker.CircuitBreaker
	config *Config
}

// New 创建新的熔断器
func New(config *Config) *CircuitBreaker {
	if config == nil {
		return nil
	}

	st := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: config.ReadyToTrip,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			if config.OnStateChange != nil {
				config.OnStateChange(name, State(from), State(to))
			}
		},
		IsSuccessful: config.IsSuccessful,
	}

	return &CircuitBreaker{
		cb:     gobreaker.NewCircuitBreaker(st),
		config: config,
	}
}

// NewDefault 使用默认配置创建熔断器
func NewDefault(name string) *CircuitBreaker {
	return New(DefaultConfig(name))
}

// Execute 执行受熔断器保护的函数
func (c *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	result, err := c.cb.Execute(fn)
	if err != nil {
		// 检查是否是熔断器打开的错误
		if errors.Is(err, gobreaker.ErrOpenState) {
			return nil, xerr.New(xerr.CodeServiceUnavailable, "circuit breaker is open")
		}
		if errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, xerr.New(xerr.CodeServiceUnavailable, "too many requests in half-open state")
		}
		// 其他错误直接返回
		return nil, err
	}
	return result, nil
}

// ExecuteWithResult 执行受熔断器保护的函数，支持泛型
func ExecuteWithResult[T any](c *CircuitBreaker, fn func() (T, error)) (T, error) {
	var zero T

	result, err := c.Execute(func() (interface{}, error) {
		return fn()
	})

	if err != nil {
		return zero, err
	}

	return result.(T), nil
}

// State 返回当前熔断器状态
func (c *CircuitBreaker) State() State {
	return State(c.cb.State())
}

// Counts 返回当前计数器
func (c *CircuitBreaker) Counts() gobreaker.Counts {
	return c.cb.Counts()
}

// Reset 重置熔断器状态
// 注意：gobreaker 库不提供公开的 Reset 方法，这里仅作为接口占位
func (c *CircuitBreaker) Reset() {
	// gobreaker 不提供公开的 Reset 方法
	// 如需重置，需要创建新的熔断器实例
}

// Name 返回熔断器名称
func (c *CircuitBreaker) Name() string {
	return c.config.Name
}

// Manager 熔断器管理器
type Manager struct {
	breakers map[string]*CircuitBreaker
}

// NewManager 创建熔断器管理器
func NewManager() *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Get 获取或创建熔断器
func (m *Manager) Get(name string, config *Config) *CircuitBreaker {
	if cb, exists := m.breakers[name]; exists {
		return cb
	}

	cb := New(config)
	m.breakers[name] = cb
	return cb
}

// GetDefault 获取或创建使用默认配置的熔断器
func (m *Manager) GetDefault(name string) *CircuitBreaker {
	return m.Get(name, DefaultConfig(name))
}

// Remove 移除熔断器
func (m *Manager) Remove(name string) {
	delete(m.breakers, name)
}

// List 列出所有熔断器名称
func (m *Manager) List() []string {
	var names []string
	for name := range m.breakers {
		names = append(names, name)
	}
	return names
}
