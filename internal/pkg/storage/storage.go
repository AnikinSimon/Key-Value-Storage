package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"sync"
	"time"

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
	InnerScalar map[string]value            `json:"innerscalar"`
	InnerArray  map[string][]value          `json:"innerarray"`
	InnerMap    map[string]map[string]value `json:"innermap"`
	InnerExpire map[string]int64            `json:"innerexpire"`
}

type Kind string

const (
	kindInt      Kind = "D"
	kindString   Kind = "S"
	kindUndefind Kind = "UND"
)

type StructKind string

const (
	kindScalar   StructKind = "SCALAR"
	kindArray    StructKind = "ARRAY"
	kindMap      StructKind = "MAP"
	kindNoStruct StructKind = "NOSTRUCTURE"
)

type Storage struct {
	innerScalar  map[string]value
	innerArray   map[string]*Treap
	innerMap     map[string]map[string]value
	innerKeys    map[string]StructKind
	innerExpire  map[string]int64
	mutex        *sync.RWMutex
	logger       *zap.Logger
	dbConnection *sql.DB
	appCfg       *appConfig
}

type StorageOption func(*Storage)

func WithoutLogging() StorageOption {
	return func(st *Storage) {
		st.logger = zap.NewNop()
	}
}

func NewStorage(opts ...StorageOption) (*Storage, error) {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Panic(err)
	}

	appConfig, err := getConfig()
	if err != nil {
		return nil, err
	}

	db, err := InitializeDb(appConfig)
	if err != nil {
		return nil, err
	}

	resStorage := &Storage{
		innerScalar:  make(map[string]value),
		innerArray:   make(map[string]*Treap),
		innerKeys:    make(map[string]StructKind),
		innerMap:     make(map[string]map[string]value),
		innerExpire:  make(map[string]int64),
		mutex:        new(sync.RWMutex),
		logger:       logger,
		dbConnection: db,
		appCfg:       appConfig,
	}

	for _, opt := range opts {
		opt(resStorage)
	}

	closeChan := make(chan struct{})
	go resStorage.startExpirationChecker(closeChan, time.Second*60)

	return resStorage, nil
}

func (r *Storage) GetConfigPort() string {
	return r.appCfg.serverCFG.Port
}

func (r *Storage) getStruct(key string) StructKind {
	sKind, ok := r.innerKeys[key]
	if !ok {
		return kindNoStruct
	}
	return sKind
}

func (r *Storage) HSET(key string, field string, val any) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	struct_kind := r.getStruct(key)
	if struct_kind == kindArray || struct_kind == kindScalar {
		return errors.New("KeyError: this key already exists and has different type")
	}

	new_val, err := newValue(val)
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}

	_, ok := r.innerMap[key]
	if !ok {
		r.innerMap[key] = make(map[string]value)
	}
	r.innerMap[key][field] = new_val
	r.innerExpire[key] = 0
	return nil
}

func (r *Storage) HGET(key string, field string) *any {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	res, ok := r.hget(key, field)

	if !ok {
		r.logger.Error("KeyError",
			zap.String("Wrong key", key),
		)
		return nil
	}

	if r.isExpired(key) {
		r.deleteKey(key, kindMap)
		return nil
	}

	return &res.Val
}

func (r *Storage) hget(key string, field string) (value, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	res, ok := r.innerMap[key][field]
	if !ok {
		return value{}, false
	}
	return res, true
}

func (r *Storage) SET(key string, val any, expireAt int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	struct_kind := r.getStruct(key)
	if struct_kind == kindArray || struct_kind == kindMap {
		return errors.New("KeyError: this key already exists and has different type")
	}
	new_val, err := newValue(val)
	if err != nil {
		r.logger.Error(err.Error())
		return err
	}
	r.innerScalar[key] = new_val
	r.innerKeys[key] = kindScalar
	r.Expire(key, expireAt)

	return nil
}

func (r *Storage) GET(key string) *any {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	res, ok := r.get(key)

	if !ok {
		r.logger.Error("KeyError",
			zap.String("Wrong key", key),
		)
		return nil
	}
	return &res.Val
}

func (r *Storage) GetKind(key string) (Kind, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	res, ok := r.get(key)
	if !ok {
		return kindUndefind, false
	}
	return res.Kin, true
}

func (r *Storage) get(key string) (value, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	res, ok := r.innerScalar[key]
	if !ok {
		return value{}, false
	}
	if r.isExpired(key) {
		r.deleteKey(key, kindScalar)
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

func (r *Storage) LPUSH(key string, args []any) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(args) == 0 {
		return errors.New("WrongArgs")
	}

	struct_kind := r.getStruct(key)
	if struct_kind == kindScalar || struct_kind == kindMap {
		return errors.New("KeyError: this key already exists and has different type")
	}

	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = NewTreap()
		r.innerExpire[key] = 0
	}

	if r.isExpired(key) {
		r.deleteKey(key, kindArray)
		return errors.New("KeyExpired")
	}

	for _, arg := range args {
		err := r.innerArray[key].PushFront(arg)
		if err != nil {
			return err
		}
	}
	r.innerKeys[key] = kindArray

	return nil
}

func (r *Storage) RPUSH(key string, args []any) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(args) == 0 {
		return errors.New("WrongArgs")
	}

	struct_kind := r.getStruct(key)
	if struct_kind == kindScalar || struct_kind == kindMap {
		return errors.New("KeyError: this key already exists and has different type")
	}

	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = NewTreap()
		r.innerExpire[key] = 0
	}

	if r.isExpired(key) {
		r.deleteKey(key, kindArray)
		return errors.New("KeyExpired")
	}

	for _, arg := range args {
		err := r.innerArray[key].PushBack(arg)
		if err != nil {
			return err
		}
	}
	r.innerKeys[key] = kindArray

	return nil
}

func (r *Storage) RADDTOSET(key string, args []any) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(args) == 0 {
		return errors.New("WrongArgs")
	}

	struct_kind := r.getStruct(key)
	if struct_kind == kindScalar || struct_kind == kindMap {
		return errors.New("KeyError: this key already exists and has different type")
	}

	if _, ok := r.innerArray[key]; !ok {
		r.innerArray[key] = NewTreap()
		r.innerExpire[key] = 0
	}

	if r.isExpired(key) {
		r.deleteKey(key, kindArray)
		return errors.New("KeyExpired")
	}

	for _, arg := range args {
		err := r.innerArray[key].PushBackToSet(arg)
		if err != nil {
			return err
		}
	}
	r.innerKeys[key] = kindArray

	return nil
}

func (r *Storage) LPOP(key string, args []int) ([]any, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(args) > 2 {
		return nil, errors.New("WrongArgs")
	}

	if r.isExpired(key) {
		r.deleteKey(key, kindArray)
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

func (r *Storage) RPOP(key string, args []int) ([]any, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(args) > 2 {
		return nil, errors.New("WrongArgs")
	}

	if r.isExpired(key) {
		r.deleteKey(key, kindArray)
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

func (r *Storage) LSET(key string, index int, val any) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isExpired(key) {
		r.deleteKey(key, kindArray)
		return errors.New("KeyExpired")
	}

	trp, ok := r.innerArray[key]
	if !ok {
		r.logger.Error("KeyError", zap.String("Key doesn't exist", key))
		return errors.New("KeyError")
	}
	if trp.Set(index, val) {
		return nil
	}

	return errors.New("IndexOutOfRange")
}

func (r *Storage) LGET(key string, index int) (any, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isExpired(key) {
		r.deleteKey(key, kindArray)
		return nil, errors.New("KeyExpired")
	}

	trp, ok := r.innerArray[key]
	if !ok {
		r.logger.Error("KeyError", zap.String("Key doesn't exist", key))
		return nil, errors.New("KeyError")
	}

	ans, ok := trp.Get(index)
	if !ok {
		return nil, errors.New("IndexOutOfRange")
	}

	return ans, nil
}

func (r *Storage) Expire(key string, secs int64) int {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	valKind := r.getStruct(key)
	if valKind == kindNoStruct {
		return 0
	}
	if secs == 0 {
		r.innerExpire[key] = 0
	} else {
		r.innerExpire[key] = time.Now().Add(time.Duration(secs * int64(time.Second))).UnixMilli()
	}
	return 1
}

func (r *Storage) getState() StorageCondition {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	inArr := make(map[string][]value)
	for k, v := range r.innerArray {
		inArr[k] = v.GetAllValues()
	}

	toIncode := StorageCondition{
		InnerScalar: r.innerScalar,
		InnerArray:  inArr,
		InnerMap:    r.innerMap,
		InnerExpire: r.innerExpire,
	}
	return toIncode
}

func (r *Storage) recoverFromCondition(state StorageCondition) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	fmt.Println(state)
	r.innerExpire = state.InnerExpire

	innerScalarState := state.InnerScalar
	for key, val := range innerScalarState {
		if r.isExpired(key) {
			delete(r.innerExpire, key)
		} else {
			tempExp := r.innerExpire[key]
			r.SET(key, val.Val, 0)
			r.innerExpire[key] = tempExp
		}
	}
	innerArrayState := state.InnerArray

	for key, vals := range innerArrayState {
		if r.isExpired(key) {
			delete(r.innerExpire, key)
			continue
		}
		tempExp := r.innerExpire[key]
		toPush := []any{}
		for _, val := range vals {
			toPush = append(toPush, val.Val)
		}
		r.RPUSH(key, toPush)
		r.innerExpire[key] = tempExp
	}

	for key, inHash := range state.InnerMap {
		if r.isExpired(key) {
			delete(r.innerExpire, key)
			continue
		}
		tempExp := r.innerExpire[key]
		for field, val := range inHash {
			r.HSET(key, field, val.Val)
		}
		r.innerExpire[key] = tempExp
	}
}

func (r *Storage) isExpired(key string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	expireAt := r.innerExpire[key]
	if expireAt == 0 {
		return false
	}
	return expireAt < time.Now().UnixMilli()
}

func (r *Storage) deleteKey(key string, valKind StructKind) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	switch valKind {
	case kindScalar:
		delete(r.innerScalar, key)
	case kindArray:
		delete(r.innerArray, key)
	case kindMap:
		delete(r.innerMap, key)
	}
	delete(r.innerKeys, key)
	delete(r.innerExpire, key)
}

func isFloatInt(num any) bool {
	return num.(float64) == math.Trunc(num.(float64))
}

func (r *Storage) garbageCollector() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for key := range r.innerExpire {
		if r.isExpired(key) {
			r.deleteKey(key, r.innerKeys[key])
		}
	}
}

func (r *Storage) startExpirationChecker(closeChan chan struct{}, tm time.Duration) {
	for {
		select {
		case <-closeChan:
			return
		case <-time.After(tm):
			r.garbageCollector()
			r.WriteStateToDB()
		}
	}
}
