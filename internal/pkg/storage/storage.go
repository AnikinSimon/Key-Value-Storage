package storage

import (
	"strconv"

	"go.uber.org/zap"
)

type Storage struct {
	inner  map[string]string
	logger *zap.Logger
}

func NewStorage() (Storage, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return Storage{}, err
	}

	defer logger.Sync()
	logger.Info("New storage created")

	return Storage{
		inner:  make(map[string]string),
		logger: logger,
	}, nil
}

func (r Storage) Set(key, value string) {
	r.inner[key] = value

	r.logger.Info("New key and value added", zap.String("Key", key),
		zap.String("Value", value))
	r.logger.Sync()
}

func (r Storage) Get(key string) *string {
	res, ok := r.inner[key]
	if !ok {
		r.logger.Warn("KeyError", zap.String("Wrong key", key))
		r.logger.Sync()
		return nil
	}

	r.logger.Info("Key obtained", zap.String("Key", key),
		zap.String("Value", res))
	r.logger.Sync()

	return &res
}

func (r Storage) GetKind(key string) string {
	var value *string = r.Get(key)
	if value == nil {
		return "KeyError"
	}
	num, ok := strconv.Atoi(*value)
	if ok == nil {
		r.logger.Info("Value is Interger", zap.String("Key", key), zap.Int("Value", num))
		r.logger.Sync()
		return "D"
	}
	r.logger.Info("Value is String", zap.String("Key", key), zap.String("Value", *value))
	r.logger.Sync()
	return "S"
}
