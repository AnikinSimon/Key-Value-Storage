package storage

import (
	"log"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type value struct {
	v string
	k kind
}

type kind string

const (
	kindInt      kind = "D"
	kindString   kind = "S"
	kindUndefind kind = "UND"
)

var atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

type Storage struct {
	inner  map[string]value
	logger *zap.Logger
}

func NewStorage() (Storage, error) {
	loggerCfg := zap.Config{
		Level:            atomicLevel,
		Development:      true,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := loggerCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	return Storage{
		inner:  make(map[string]value),
		logger: logger,
	}, nil
}

func (r Storage) SetLoggerLevel(lvl string) {
	switch lvl {
	case "debug":
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case "info":
		atomicLevel.SetLevel(zapcore.InfoLevel)
	case "warn":
		atomicLevel.SetLevel(zapcore.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zapcore.ErrorLevel)
	case "fatal":
		atomicLevel.SetLevel(zapcore.FatalLevel)
	default:
		return
	}
}

func (r Storage) Set(key, val string) {
	// defer r.logger.Sync()
	switch k := getType(val); k {
	case kindInt, kindString:
		r.inner[key] = value{
			v: val,
			k: k,
		}
	case kindUndefind:
		r.logger.Warn("Undefined type of value")
	}

	// r.logger.Info("New key and value added",
	// 	zap.String("Key", key),
	// 	zap.String("Value", val),
	// )
}

func (r Storage) Get(key string) *string {
	res, ok := r.get(key)
	// defer r.logger.Sync()
	if !ok {
		r.logger.Warn("KeyError",
			zap.String("Wrong key", key),
		)
		return nil
	}

	// r.logger.Info("Key obtained",
	// 	zap.String("Key", key),
	// 	zap.String("Value", res.v),
	// )

	return &res.v
}

func (r Storage) GetKind(key string) string {
	res, ok := r.get(key)
	if !ok {
		return "KeyError"
	}
	// defer r.logger.Sync()

	// r.logger.Info("Kind obtained",
	// 	zap.String("Value", res.v),
	// 	zap.String("Type", string(res.k)),
	// )

	return string(res.k)
}

func (r Storage) get(key string) (value, bool) {
	res, ok := r.inner[key]
	if !ok {
		return value{}, false
	}
	return res, true
}

func getType(val string) kind {
	var conv any
	conv, ok := strconv.Atoi(val)
	if ok != nil {
		conv = val
	}
	switch conv.(type) {
	case int:
		return kindInt
	case string:
		return kindString
	default:
		return kindUndefind
	}
}
