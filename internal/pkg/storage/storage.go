package storage

import (
	"errors"
	"golangProject/internal/pkg/treap"
	"log"
	"slices"
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
	innerScalar map[string]value
	innerArray  map[string]*treap.Treap
	logger      *zap.Logger
	loggerCheck bool
}

func NewStorage() (Storage, error) {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Panic(err)
	}

	return Storage{
		innerScalar: make(map[string]value),
		innerArray:  make(map[string]*treap.Treap),
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
		r.innerScalar[key] = value{
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
	res, ok := r.innerScalar[key]
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

func (r Storage) LPUSH(key string, args []int) {
	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = treap.NewTreap()
	}
	for _, arg := range args {
		r.innerArray[key].PushFront(arg)
	}
}

func (r Storage) RPUSH(key string, args []int) {
	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = treap.NewTreap()
	}
	for _, arg := range args {
		r.innerArray[key].PushBack(arg)
	}
}

func (r Storage) RADDTOSET(key string, args []int) {
	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = treap.NewTreap()
	}
	for _, arg := range args {
		r.innerArray[key].PushBackToSet(arg)
	}
	r.innerArray[key].Print()
}

func (r Storage) LPOP(key string, args []int) ([]int, error) {

	trp, ok := r.innerArray[key]
	if r.loggerCheck {
		defer r.logger.Sync()
	}
	if !ok {
		if r.loggerCheck {
			r.logger.Warn("KeyError", zap.String("Key doesn't exist", key))
		}
		return nil, errors.New("KeyError")
	}

	switch len(args) {
	case 2:
		if r.loggerCheck {
			r.logger.Info("LPOP 2 Args")
		}
		rt := trp.ValidateSlice(args[0])
		lf := trp.ValidateSlice(args[1])
		if r.loggerCheck {
			r.logger.Info("Bounds", zap.Int("RT", rt), zap.Int("LFT", lf))
		}
		if rt > lf {
			if r.loggerCheck {
				r.logger.Warn("Wrong Bounds")
			}
			return nil, errors.New("IndexOutOfRange")
		}
		nodes := trp.EraseSection(rt, lf)
		return nodes, nil
	case 1:
		if r.loggerCheck {
			r.logger.Info("LPOP 1 Arg")
		}
		ans := make([]int, 0)
		if args[0] < 0 {
			return nil, errors.New("IndexOutOfRange")
		} else {
			for i := 0; i < args[0]; i++ {
				ans = append(ans, trp.PopFront())
			}
			return ans, nil
		}
	case 0:
		if r.loggerCheck {
			r.logger.Info("LPOP 0 Args")
		}
		ans := make([]int, 0)
		ans = append(ans, trp.PopFront())
		return ans, nil
	}
	return nil, errors.New("WronnArgs")
}

func (r Storage) RPOP(key string, args []int) ([]int, error) {

	trp, ok := r.innerArray[key]

	if r.loggerCheck {
		defer r.logger.Sync()
	}

	if !ok {
		if r.loggerCheck {
			r.logger.Warn("KeyError", zap.String("Key doesn't exist", key))
		}
		return nil, errors.New("KeyError")
	}

	switch len(args) {
	case 2:
		if r.loggerCheck {
			r.logger.Info("LPOP 2 Args")
		}
		rt := trp.ValidateSlice(args[0])
		lf := trp.ValidateSlice(args[1])
		if r.loggerCheck {
			r.logger.Info("Bounds", zap.Int("RT", rt), zap.Int("LFT", lf))
		}
		if rt > lf {
			if r.loggerCheck {
				r.logger.Warn("Wrong Bounds")
			}
			return nil, errors.New("IndexOutOfRange")
		}
		nodes := trp.EraseSection(rt, lf)
		slices.Reverse(nodes)
		return nodes, nil
	case 1:
		if r.loggerCheck {
			r.logger.Info("LPOP 1 Arg1")
		}
		ans := make([]int, 0)
		if args[0] < 0 {
			return nil, errors.New("IndexOutOfRange")
		} else {
			for i := 0; i < args[0]; i++ {
				ans = append(ans, trp.PopBack())
			}
			return ans, nil
		}
	case 0:
		if r.loggerCheck {
			r.logger.Info("LPOP 0 Args")
		}
		ans := make([]int, 0)
		ans = append(ans, trp.PopBack())
		return ans, nil
	}
	return nil, errors.New("WrongArgs")
}

func (r Storage) LSET(key string, args []int) error {
	trp, ok := r.innerArray[key]
	if !ok {
		if r.loggerCheck {
			r.logger.Warn("KeyError", zap.String("Key doesn't exist", key))
		}
		return errors.New("KeyError")
	}
	if len(args) != 2 {
		return errors.New("WrongArgs")
	}

	if trp.Set(args[0], args[1]) {
		return nil
	}

	return errors.New("IndexOutOfRange")
}

func (r Storage) LGET(key string, args []int) (int, error) {
	trp, ok := r.innerArray[key]
	if !ok {
		if r.loggerCheck {
			r.logger.Warn("KeyError", zap.String("Key doesn't exist", key))
		}
		return -1, errors.New("KeyError")
	}
	if len(args) != 1 {
		return -1, errors.New("WrongArgs")
	}
	ans, ok := trp.Get(args[0])
	if !ok {
		return -1, errors.New("IndexOutOfRange")
	}

	return ans, nil
}
