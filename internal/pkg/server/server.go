package server

import (
	"encoding/json"
	"fmt"
	"golangProject/internal/pkg/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host  string
	store *storage.Storage
}

type Entry struct {
	Value any `json:"value"`
}

type EntrySet struct {
	Value any    `json:"value"`
	Ex    uint32 `json:"ex,omitempty"`
}

type EntryArray struct {
	Value []any `json:"value"`
}

func New(host string, st *storage.Storage) *Server {
	s := &Server{
		host:  host,
		store: st,
	}

	return s
}

func (r *Server) newAPI() *gin.Engine {
	engine := gin.New()

	engine.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "OK")
	})

	engine.POST("/scalar/set/:key", r.handlerSet)
	engine.GET("/scalar/get/:key", r.handlerGet)

	engine.POST("/hash/set/:key/:field", r.handlerHSET)
	engine.GET("/hash/get/:key/:field", r.handlerHGET)

	engine.POST("array/rpush/:key", r.handlerRPUSH)
	engine.POST("array/raddtoset/:key", r.handlerRADDTOSET)
	engine.GET("array/rpop/:key", r.handlerRPOP)

	engine.POST("array/lpush/:key", r.handlerLPUSH)
	engine.GET("array/lpop/:key", r.handleLPOP)

	engine.POST("array/lset/:key", r.handlerLSET)
	engine.GET("array/lget/:key", r.handleLGET)

	engine.POST("/expire/:key", r.handlerExpire)

	return engine
}

func (r *Server) handlerSet(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntrySet
	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	err := r.store.SET(key, v.Value, int64(v.Ex))
	if err != nil {
		fmt.Println(err)
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
}

func (r *Server) handlerGet(ctx *gin.Context) {
	key := ctx.Param("key")

	v := r.store.GET(key)
	if v == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, Entry{
		Value: *v,
	})
}

func (r *Server) handlerHSET(ctx *gin.Context) {
	key := ctx.Param("key")
	field := ctx.Param("field")

	var v Entry
	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	err := r.store.HSET(key, field, v.Value)
	if err != nil {
		fmt.Println(err)
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
}

func (r *Server) handlerHGET(ctx *gin.Context) {
	key := ctx.Param("key")
	field := ctx.Param("field")

	v := r.store.HGET(key, field)
	if v == nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, Entry{
		Value: *v,
	})
}

func (r *Server) handlerRPUSH(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	err := r.store.RPUSH(key, v.Value)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusOK)
}

func (r *Server) handlerRADDTOSET(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	err := r.store.RADDTOSET(key, v.Value)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusOK)
}

func (r *Server) handlerRPOP(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	vals, err := r.store.RPOP(key, v.Value)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Entry{
		Value: vals,
	})
}

func (r *Server) handlerLPUSH(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	err := r.store.LPUSH(key, v.Value)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusOK)
}

func (r *Server) handleLPOP(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	vals, err := r.store.LPOP(key, v.Value)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Entry{
		Value: vals,
	})
}

func (r *Server) handlerLSET(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	err := r.store.LSET(key, v.Value)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusOK)
}

func (r *Server) handleLGET(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	vals, err := r.store.LGET(key, v.Value)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, Entry{
		Value: vals,
	})
}

func (r *Server) handlerExpire(ctx *gin.Context) {
	key := ctx.Param("key")

	var v Entry
	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	expireCode := r.store.Expire(key, int64(v.Value.(float64)))
	
	ctx.JSON(http.StatusOK, Entry{
		Value: expireCode,
	})
}

func (r *Server) Start() {
	r.newAPI().Run(r.host)
}
