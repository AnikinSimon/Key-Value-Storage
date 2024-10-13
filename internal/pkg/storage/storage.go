package storage

import (
	"log"
	"strconv"

	"go.uber.org/zap"
)

type value struct {
	v string
	k Kind
}

type Kind string

const (
	kindInt      Kind = "D"
	kindString   Kind = "S"
	kindUndefind Kind = "UND"
)

type Storage struct {
	inner       map[string]value
	logger      *zap.Logger
	loggerCheck bool
}

func NewStorage() (Storage, error) {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Panic(err)
	}

	return Storage{
		inner:       make(map[string]value),
		logger:      logger,
		loggerCheck: false,
	}, nil
}

func (r Storage) SwitchTestLogger() {
	(&r).loggerCheck = !r.loggerCheck
}

func (r Storage) Set(key, val string) {
	if r.loggerCheck {
		defer r.logger.Sync()
	}
	switch k := getType(val); k {
	case kindInt, kindString:
		r.inner[key] = value{
			v: val,
			k: k,
		}
	case kindUndefind:
		if r.loggerCheck {
			r.logger.Warn("Undefined type of value")
		}
	}
	if r.loggerCheck {
		r.logger.Info("New key and value added",
			zap.String("Key", key),
			zap.String("Value", val),
		)
	}
}

func (r Storage) Get(key string) *string {
	res, ok := r.get(key)
	if r.loggerCheck {
		defer r.logger.Sync()
	}

	if !ok {
		if r.loggerCheck {
			r.logger.Warn("KeyError",
				zap.String("Wrong key", key),
			)
		}
		return nil
	}

	if r.loggerCheck {
		r.logger.Info("Key obtained",
			zap.String("Key", key),
			zap.String("Value", res.v),
		)
	}

	return &res.v
}

func (r Storage) GetKind(key string) (Kind, bool) {
	res, ok := r.get(key)
	if !ok {
		return kindUndefind, false
	}

	if r.loggerCheck {
		defer r.logger.Sync()

		r.logger.Info("Kind obtained",
			zap.String("Value", res.v),
			zap.String("Type", string(res.k)),
		)
	}

	return res.k, true
}

func (r Storage) get(key string) (value, bool) {
	res, ok := r.inner[key]
	if !ok {
		return value{}, false
	}
	return res, true
}

func getType(val string) Kind {
	var conv any
	conv, err := strconv.Atoi(val)
	if err != nil {
		return kindString
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
