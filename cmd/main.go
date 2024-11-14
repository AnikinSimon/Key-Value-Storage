package main

import (
	"golangProject/internal/pkg/server"
	"golangProject/internal/pkg/storage"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		storage.ErrorHandler(err)
	}

	store.ReadStateFromDB() 

	serve := server.New(store)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		serve.Start()
	}()
	<-sigChan

	store.GracefulShutdown()
}
