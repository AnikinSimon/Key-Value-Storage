package storage

import (
	"errors"
	"log"
	"math"
	"slices"

	"go.uber.org/zap"
)

type value struct {
	Val any  `json:"value"`
	Kin Kind `json:"type"`
}

func newValue(val any) (value, error) {
	switch k := getType(val); k {
	case kindInt, kindString:
		return value{
			Val: val,
			Kin: k,
		}, nil
	}
	return value{}, errors.New("UndefinedValueType")
}

type StorageCondition struct {
	InnerScalar map[string]value   `json:"innerscalar"`
	InnerArray  map[string][]value `json:"innerarray"`
}

type Kind string

const (
	kindInt      Kind = "D"
	kindString   Kind = "S"
	kindUndefind Kind = "UND"
)

type Storage struct {
	innerScalar map[string]value
	innerArray  map[string]*Treap
	logger      *zap.Logger
}

type StorageOption func(*Storage)

func WithoutLogging() StorageOption {
	return func(st *Storage) {
		st.logger = zap.NewNop()
	}
}

func NewStorage(opts ...StorageOption) (Storage, error) {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Panic(err)
	}

	resStorage := Storage{
		innerScalar: make(map[string]value),
		innerArray:  make(map[string]*Treap),
		logger:      logger,
	}

	for _, opt := range opts {
		opt(&resStorage)
	}

	return resStorage, nil
}

func (r Storage) SET(key string, val any) {
	new_val, err := newValue(val)
	if err != nil {
		r.logger.Error(err.Error())
		return
	}
	r.innerScalar[key] = new_val
}

func (r Storage) GET(key string) *any {
	res, ok := r.get(key)

	if !ok {
		r.logger.Error("KeyError",
			zap.String("Wrong key", key),
		)
		return nil
	}
	return &res.Val
}

func (r Storage) GetKind(key string) (Kind, bool) {
	res, ok := r.get(key)
	if !ok {
		return kindUndefind, false
	}
	return res.Kin, true
}

func (r Storage) get(key string) (value, bool) {
	res, ok := r.innerScalar[key]
	if !ok {
		return value{}, false
	}
	return res, true
}

func getType(val any) Kind {
	switch val.(type) {
	case int:
		return kindInt
	case float64:
		if isFloatInt(val) {
			return kindInt
		}
		return kindUndefind
	case string:
		return kindString
	default:
		return kindUndefind
	}
}

func (r Storage) LPUSH(key string, args []any) {
	if len(args) == 0 {
		return
	}

	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = NewTreap()
	}

	for _, arg := range args {
		r.innerArray[key].PushFront(arg)
	}
}

func (r Storage) RPUSH(key string, args []any) {
	if len(args) == 0 {
		return
	}

	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = NewTreap()
	}

	for _, arg := range args {
		r.innerArray[key].PushBack(arg)
	}
}

func (r Storage) RADDTOSET(key string, args []any) {
	if len(args) == 0 {
		return
	}

	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = NewTreap()
	}

	for _, arg := range args {
		r.innerArray[key].PushBackToSet(arg)
	}
}

func (r Storage) LPOP(key string, args []any) ([]any, error) {
	if len(args) > 2 || !allArgsAreInt(args) {
		return nil, errors.New("WrongArgs")
	}
	trp, ok := r.innerArray[key]
	if !ok {
		r.logger.Error("KeyError", zap.String("Key doesn't exist", key))
		return nil, errors.New("KeyError")
	}

	rt, lf, err := trp.ValidateEraseSlice(args, true)
	if err != nil {
		return nil, err
	}
	nodes := trp.EraseSection(rt, lf)

	return nodes, nil
}

func (r Storage) RPOP(key string, args []any) ([]any, error) {
	if len(args) > 2 || !allArgsAreInt(args) {
		return nil, errors.New("WrongArgs")
	}
	trp, ok := r.innerArray[key]
	if !ok {
		r.logger.Error("KeyError", zap.String("Key doesn't exist", key))
		return nil, errors.New("KeyError")
	}

	rt, lf, err := trp.ValidateEraseSlice(args, false)
	if err != nil {
		return nil, err
	}

	nodes := trp.EraseSection(rt, lf)
	slices.Reverse(nodes)
	return nodes, nil
}

func (r Storage) LSET(key string, args []any) error {
	if len(args) != 2 {
		return errors.New("WrongArgs")
	}
	if getType(args[0]) != kindInt {
		return errors.New("WrongArgs")
	}

	trp, ok := r.innerArray[key]
	if !ok {
		r.logger.Error("KeyError", zap.String("Key doesn't exist", key))
		return errors.New("KeyError")
	}
	if trp.Set(args[0].(int), args[1]) {
		return nil
	}

	return errors.New("IndexOutOfRange")
}

func (r Storage) LGET(key string, args []any) (any, error) {
	if len(args) != 1 || !allArgsAreInt(args) {
		return -1, errors.New("WrongArgs")
	}

	trp, ok := r.innerArray[key]
	if !ok {
		r.logger.Error("KeyError", zap.String("Key doesn't exist", key))
		return -1, errors.New("KeyError")
	}

	ans, ok := trp.Get(args[0].(int))
	if !ok {
		return -1, errors.New("IndexOutOfRange")
	}

	return ans, nil
}

func (r Storage) HandlerTask(tasks []Task) {
	for _, iter := range tasks {
		switch iter.Command {
		case "GET":
			r.GET(iter.Key)
		case "SET":
			r.SET(iter.Key, iter.Args[0])
		case "LPUSH":
			r.LPUSH(iter.Key, iter.Args)
		case "RPUSH":
			r.RPUSH(iter.Key, iter.Args)
		case "RADDTOSET":
			r.RADDTOSET(iter.Key, iter.Args)
		case "LPOP":
			r.LPOP(iter.Key, iter.Args)
		case "RPOP":
			r.RPOP(iter.Key, iter.Args)
		case "LSET":
			r.LSET(iter.Key, iter.Args)
		case "LGET":
			r.LGET(iter.Key, iter.Args)
		}
	}
}

func (r Storage) getState() StorageCondition {
	inArr := make(map[string][]value)
	for k, v := range r.innerArray {
		inArr[k] = v.GetAllValues()
	}

	toIncode := StorageCondition{
		InnerScalar: r.innerScalar,
		InnerArray:  inArr,
	}
	return toIncode
}

func (r Storage) recoverFromCondition(state StorageCondition) {
	innerScalarState := state.InnerScalar
	for key, val := range innerScalarState {
		r.SET(key, val.Val)
	}
	innerArrayState := state.InnerArray

	for key, vals := range innerArrayState {
		toPush := []any{}
		for _, val := range vals {
			toPush = append(toPush, val.Val)
		}
		r.RPUSH(key, toPush)
	}
}

func allArgsAreInt(args []any) bool {
	for _, arg := range args {
		switch arg.(type) {
		case int:
		case float64:
			if !isFloatInt(arg) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func isFloatInt(num any) bool {
	return num.(float64) == math.Trunc(num.(float64))
}
