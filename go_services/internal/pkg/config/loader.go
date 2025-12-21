package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Loader 是 viper 的简单封装，用于统一 monorepo 中服务的配置加载方式
//
// 约定：
//   - 配置文件格式：YAML
//   - 支持环境变量覆盖：启用
//   - 环境变量前缀：BIFROST（可覆盖）
//   - 键名替换规则："." -> "_"（例如 app.grpc_port => BIFROST_APP_GRPC_PORT）
//
// 这确保了所有服务的配置加载方式一致
type Loader struct {
	v *viper.Viper
}

// Option 配置加载器选项的函数类型
type Option func(*Loader)

// WithEnvPrefix 设置环境变量前缀（默认：BIFROST）
func WithEnvPrefix(prefix string) Option {
	return func(l *Loader) {
		l.v.SetEnvPrefix(prefix)
	}
}

// WithDefaults 设置配置默认值
func WithDefaults(defaults map[string]any) Option {
	return func(l *Loader) {
		for k, v := range defaults {
			l.v.SetDefault(k, v)
		}
	}
}

// NewLoader 创建新的配置加载器
func NewLoader(opts ...Option) *Loader {
	v := viper.New()
	v.SetConfigType("yaml")

	// 环境变量覆盖设置
	v.SetEnvPrefix("BIFROST")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	l := &Loader{v: v}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// LoadFile 读取 YAML 配置文件并解析到 out 参数指向的结构体中
//
// 如果提供了 validate 函数，它会在解析完成后被调用
func (l *Loader) LoadFile(path string, out any, validate func() error) error {
	l.v.SetConfigFile(path)
	if err := l.v.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}
	if err := l.v.Unmarshal(out); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	if validate != nil {
		if err := validate(); err != nil {
			return err
		}
	}
	return nil
}

// Viper 返回底层的 viper 实例，用于高级用法
func (l *Loader) Viper() *viper.Viper {
	return l.v
}
