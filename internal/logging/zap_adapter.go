package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger returns a zap logger adapter to be used with `Logging` interface.
func NewZapLogger(lvl string) ZapWrapper {
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(lvl)); err != nil {
		logLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	core := zapcore.NewCore(zapcore.NewJSONEncoder(config), zapcore.Lock(os.Stderr), logLevel)
	logger := zap.New(core, zap.WithCaller(false))
	return ZapWrapper{Logger: logger}
}

// ZapWrapper is an adapter between service defined logging interface and zap logging implementation.
type ZapWrapper struct {
	*zap.Logger
}

func (zw ZapWrapper) WithString(key, val string) Logger {
	return ZapWrapper{Logger: zw.Logger.With(zap.String(key, val))}
}

func (zw ZapWrapper) WithUint64(key string, val uint64) Logger {
	return ZapWrapper{Logger: zw.Logger.With(zap.Uint64(key, val))}
}

func (zw ZapWrapper) WithInt64(key string, val int64) Logger {
	return ZapWrapper{Logger: zw.Logger.With(zap.Int64(key, val))}
}

func (zw ZapWrapper) WithInt(key string, val int) Logger {
	return ZapWrapper{Logger: zw.Logger.With(zap.Int(key, val))}
}

func (zw ZapWrapper) WithError(err error) Logger {
	return ZapWrapper{Logger: zw.Logger.With(zap.Error(err))}
}

func (zw ZapWrapper) Debug(msg string) {
	zw.Logger.Debug(msg)
}

func (zw ZapWrapper) Info(msg string) {
	zw.Logger.Info(msg)
}

func (zw ZapWrapper) Error(msg string) {
	zw.Logger.Error(msg)
}

func (zw ZapWrapper) IsDebug() bool {
	return zw.Logger.Core().Enabled(zapcore.DebugLevel)
}
