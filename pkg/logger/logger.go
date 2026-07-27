package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// CtxKey 上下文键类型
type CtxKey string

const (
	// TraceIDKey 用于在 context 中传递 Trace ID
	TraceIDKey CtxKey = "trace_id"
	// ModuleKey 用于在 context 中传递模块名
	ModuleKey CtxKey = "module"
)

// Logger 封装 zerolog，提供结构化日志能力
type Logger struct {
	logger zerolog.Logger
	module string
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string `mapstructure:"level"`         // debug, info, warn, error
	Format        string `mapstructure:"format"`        // json, text
	SlowThreshold int64  `mapstructure:"slowThreshold"` // 慢请求阈值（毫秒）
}

// New 创建 Logger 实例
func New(cfg LogConfig) *Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var output io.Writer = os.Stdout
	if cfg.Format == "text" {
		output = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = time.DateTime
			w.Out = os.Stdout
		})
	}

	zlog := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &Logger{logger: zlog}
}

// WithModule 返回绑定模块名的 Logger
func (l *Logger) WithModule(module string) *Logger {
	child := l.logger.With().Str("module", module).Logger()
	return &Logger{logger: child, module: module}
}

// WithContext 从 context 中提取 trace_id 并绑定到日志
func (l *Logger) WithContext(ctx context.Context) *Logger {
	child := l.logger.With().Logger()

	if ctx != nil {
		if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
			child = child.With().Str("trace_id", traceID).Logger()
		}
	}

	if l.module != "" {
		child = child.With().Str("module", l.module).Logger()
	}

	return &Logger{logger: child}
}

// Debug 记录 Debug 级别日志
func (l *Logger) Debug(msg string, fields ...map[string]any) {
	evt := l.logger.Debug()
	addFields(evt, fields...)
	evt.Msg(msg)
}

// Info 记录 Info 级别日志
func (l *Logger) Info(msg string, fields ...map[string]any) {
	evt := l.logger.Info()
	addFields(evt, fields...)
	evt.Msg(msg)
}

// Warn 记录 Warn 级别日志
func (l *Logger) Warn(msg string, fields ...map[string]any) {
	evt := l.logger.Warn()
	addFields(evt, fields...)
	evt.Msg(msg)
}

// Error 记录 Error 级别日志
func (l *Logger) Error(msg string, err error, fields ...map[string]any) {
	evt := l.logger.Error()
	if err != nil {
		evt = evt.Err(err)
	}
	addFields(evt, fields...)
	evt.Msg(msg)
}

// Fatal 记录 Fatal 级别日志并退出
func (l *Logger) Fatal(msg string, err error, fields ...map[string]any) {
	evt := l.logger.Fatal()
	if err != nil {
		evt = evt.Err(err)
	}
	addFields(evt, fields...)
	evt.Msg(msg)
}

// WithSlowThreshold 创建带慢请求阈值检测的日志包装
// 返回的 Logger 会记录耗时，超过阈值的自动提升为 Warn 级别
func (l *Logger) WithSlowThreshold(threshold time.Duration) *SlowLogger {
	return &SlowLogger{
		logger:    l,
		threshold: threshold,
	}
}

// SlowLogger 支持慢请求检测的日志包装
type SlowLogger struct {
	logger    *Logger
	threshold time.Duration
}

// LogDuration 记录耗时日志，超过阈值自动提升级别
func (s *SlowLogger) LogDuration(msg string, start time.Time, fields ...map[string]any) {
	duration := time.Since(start)
	merged := map[string]any{
		"duration_ms": duration.Milliseconds(),
	}
	for _, f := range fields {
		for k, v := range f {
			merged[k] = v
		}
	}

	if duration >= s.threshold {
		s.logger.Warn(msg+" (slow)", merged)
	} else {
		s.logger.Info(msg, merged)
	}
}

// addFields 将字段添加到日志事件
func addFields(evt *zerolog.Event, fields ...map[string]any) {
	for _, f := range fields {
		for k, v := range f {
			evt = evt.Interface(k, v)
		}
	}
}

// -- 全局默认 Logger（方便非 DI 场景使用） --

var defaultLogger *Logger

func init() {
	defaultLogger = New(LogConfig{
		Level:  "debug",
		Format: "text",
	})
}

// SetDefault 设置全局默认 Logger
func SetDefault(l *Logger) {
	defaultLogger = l
}

// L 获取全局默认 Logger
func L() *Logger {
	return defaultLogger
}

// Debug 全局 Debug 日志
func Debug(msg string, fields ...map[string]any) {
	defaultLogger.Debug(msg, fields...)
}

// Info 全局 Info 日志
func Info(msg string, fields ...map[string]any) {
	defaultLogger.Info(msg, fields...)
}

// Warn 全局 Warn 日志
func Warn(msg string, fields ...map[string]any) {
	defaultLogger.Warn(msg, fields...)
}

// Error 全局 Error 日志
func Error(msg string, err error, fields ...map[string]any) {
	defaultLogger.Error(msg, err, fields...)
}

// Fatal 全局 Fatal 日志
func Fatal(msg string, err error, fields ...map[string]any) {
	defaultLogger.Fatal(msg, err, fields...)
}