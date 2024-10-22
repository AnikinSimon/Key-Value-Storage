package server

import (
	"encoding/json"
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

	engine.PUT("/scalar/set/:key", r.handlerSet)
	engine.GET("/scalar/get/:key", r.handlerGet)

	engine.PUT("array/rpush/:key", r.handlerRPUSH)
	engine.GET("array/rpop/:key", r.handlerRPOP)
	return engine
}

func (r *Server) handlerSet(ctx *gin.Context) {
	key := ctx.Param("key")

	var v Entry

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	r.store.SET(key, v.Value)

	ctx.Status(http.StatusOK)

}

func (r *Server) handlerRPUSH(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	r.store.RPUSH(key, v.Value)

	ctx.Status(http.StatusOK)

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

func (r *Server) handlerRPOP(ctx *gin.Context) {
	key := ctx.Param("key")

	var v EntryArray

	if err := json.NewDecoder(ctx.Request.Body).Decode(&v); err != nil {
		ctx.AbortWithStatus(http.StatusBadGateway)
		return
	}

	vals, _ := r.store.RPOP(key, v.Value)

	ctx.JSON(http.StatusOK, Entry{
		Value: vals,
	})

}

func (r *Server) Start() {
	r.newAPI().Run()
}
