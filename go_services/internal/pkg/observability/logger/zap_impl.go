//此文件实现了基于 Uber Zap 的日志记录器，符合 Logger 接口。

package logger

import (
	"context"
	"io"
	"os"

	"github.com/gulugulu3399/bifrost/internal/pkg/contextx"
	"github.com/gulugulu3399/bifrost/internal/pkg/observability/tracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	z     *zap.Logger
	cfg   *Config
	level Level
}

// NewZap 创建基于 Zap 的实现
func NewZap(cfg *Config, serviceName string, env string) Logger {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 1. 配置日志级别
	var zapLevel zapcore.Level
	switch cfg.Level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	case FatalLevel:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 2. 配置编码器
	var encoder zapcore.Encoder
	encoderConfig := zapcore.EncoderConfig{
		// 键名标准化：使用简短通用的键名 (可选，但我推荐这样改)
		// 很多日志系统默认识别 'ts'/'time', 'msg', 'level'
		TimeKey:       "ts",     // 原来是 timestamp
		LevelKey:      "level",  // 保持不变
		NameKey:       "logger", // 保持不变
		CallerKey:     "caller", // 保持不变
		FunctionKey:   "",       // 留空通常更好，除非你非常需要函数名，否则会增加杂讯
		MessageKey:    "msg",    // 原来是 message，'msg' 是 Go 社区(如 zerolog)事实标准
		StacktraceKey: "stack",  // 简短一点

		LineEnding: zapcore.DefaultLineEnding,

		EncodeLevel: zapcore.LowercaseLevelEncoder,

		// 时间格式：RFC3339 是最标准的互联网时间格式，带时区，Dozzle 解析最准
		EncodeTime: zapcore.RFC3339TimeEncoder,

		// 耗时格式：在UI里看 "1.5s" 比 "1.5" 更舒服
		EncodeDuration: zapcore.StringDurationEncoder,

		//  调用者：只显示文件名和行号 (biz/user.go:123)，不显示全路径
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	if cfg.Format == "console" {
		// 控制台模式保持颜色和易读性
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// 控制台模式下，时间也可以看简短一点
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		// 控制台模式通常用 ConsoleEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		// JSON 模式（生产环境/Dozzle）
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		if cfg.EnableColor {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 3. 配置输出
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	case "file":
		if cfg.Filename == "" {
			writeSyncer = zapcore.AddSync(os.Stdout)
		} else {
			// Lightweight mode: no rotation. Just append to the file.
			f, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
			if err != nil {
				writeSyncer = zapcore.AddSync(os.Stdout)
			} else {
				writeSyncer = zapcore.AddSync(f)
			}
		}
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 4. 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	// 5. 创建日志器
	z := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip))

	// 6. 注入通用字段
	if serviceName != "" {
		z = z.With(zap.String("service", serviceName))
	}
	if env != "" {
		z = z.With(zap.String("env", env))
	}

	return &zapLogger{
		z:     z,
		cfg:   cfg,
		level: cfg.Level,
	}
}

// 核心：将我们的 Field 转为 zap.Field
func (l *zapLogger) toZapFields(fields []Field) []zap.Field {
	zf := make([]zap.Field, len(fields))
	for i, f := range fields {
		switch f.Type {
		case fieldTypeString:
			zf[i] = zap.String(f.Key, f.Str)
		case fieldTypeInt:
			zf[i] = zap.Int64(f.Key, f.Int)
		case fieldTypeBool:
			zf[i] = zap.Bool(f.Key, f.Int == 1)
		case fieldTypeError:
			if err, ok := f.Interface.(error); ok {
				zf[i] = zap.NamedError(f.Key, err)
			} else {
				zf[i] = zap.Any(f.Key, f.Interface)
			}
		case fieldTypeAny:
			zf[i] = zap.Any(f.Key, f.Interface)
		default:
			zf[i] = zap.Any(f.Key, f.Interface)
		}
	}
	return zf
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.z.Debug(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.z.Info(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.z.Warn(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.z.Error(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.z.Fatal(msg, l.toZapFields(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		z:     l.z.With(l.toZapFields(fields)...),
		cfg:   l.cfg,
		level: l.level,
	}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	// 1. 收集上下文中的关键字段
	var fields []zap.Field

	// A. 提取 TraceID (来自 OpenTelemetry)
	if tid := tracing.TraceIDFromContext(ctx); tid != "" {
		fields = append(fields, zap.String("trace_id", tid))
	}

	// B. 提取 RequestID (来自网关/Metadata)
	if rid := contextx.RequestIDFromContext(ctx); rid != "" {
		fields = append(fields, zap.String("request_id", rid))
	}

	// C. 提取 UserID (来自认证中间件)
	if uid := contextx.UserIDFromContext(ctx); uid != 0 {
		fields = append(fields, zap.Int64("user_id", uid))
	}

	// 如果没有任何新字段，直接返回原 Logger，避免不必要的内存分配
	if len(fields) == 0 {
		return l
	}

	// 返回带有新字段的 Logger 副本
	return &zapLogger{
		z:     l.z.With(fields...),
		cfg:   l.cfg,
		level: l.level,
	}
}
func (l *zapLogger) SetLevel(level Level) {
	l.level = level
	// 创建新的核心并替换
	// 注意：zap 的原子级别需要在创建时设置，这里简化处理
}

func (l *zapLogger) GetLevel() Level {
	return l.level
}

func (l *zapLogger) Sync() error {
	return l.z.Sync()
}

// NewZapDevelopment 创建开发环境的 Zap 日志器
func NewZapDevelopment(serviceName string) Logger {
	cfg := DefaultConfig()
	cfg.Level = DebugLevel
	cfg.Format = "console"
	cfg.EnableColor = true
	return NewZap(cfg, serviceName, "dev")
}

// NewZapProduction 创建生产环境的 Zap 日志器
func NewZapProduction(serviceName string) Logger {
	cfg := DefaultConfig()
	cfg.Level = InfoLevel
	cfg.Format = "json"
	cfg.EnableColor = false
	return NewZap(cfg, serviceName, "prod")
}

// NewZapWithFile 创建输出到文件的 Zap 日志器
func NewZapWithFile(filename string, serviceName string) Logger {
	cfg := DefaultConfig()
	cfg.Output = "file"
	cfg.Filename = filename
	cfg.Format = "json"
	return NewZap(cfg, serviceName, "prod")
}

// MultiWriter 多输出写入器
type MultiWriter struct {
	writers []io.Writer
}

func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}

// NewZapWithMultiOutput 创建多输出的 Zap 日志器
func NewZapWithMultiOutput(outputs []string, filename string, serviceName string) Logger {
	cfg := DefaultConfig()
	cfg.Format = "json"

	var writers []io.Writer
	for _, output := range outputs {
		switch output {
		case "stdout":
			writers = append(writers, os.Stdout)
		case "stderr":
			writers = append(writers, os.Stderr)
		case "file":
			if filename != "" {
				// Lightweight mode: no rotation. Append only.
				f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
				if err == nil {
					writers = append(writers, f)
				}
			}
		}
	}

	if len(writers) > 0 {
		multiWriter := NewMultiWriter(writers...)
		return NewZapWithWriter(multiWriter, serviceName)
	}

	return NewZapProduction(serviceName)
}

// NewZapWithWriter 使用自定义写入器创建 Zap 日志器
func NewZapWithWriter(writer io.Writer, serviceName string) Logger {
	cfg := DefaultConfig()

	var zapLevel zapcore.Level
	switch cfg.Level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	case FatalLevel:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "function",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	writeSyncer := zapcore.AddSync(writer)
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	z := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(cfg.CallerSkip))

	if serviceName != "" {
		z = z.With(zap.String("service", serviceName))
	}

	return &zapLogger{
		z:     z,
		cfg:   cfg,
		level: cfg.Level,
	}
}
