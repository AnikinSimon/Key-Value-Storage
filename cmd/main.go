package main

import (
	"golangProject/internal/pkg/server"
	"fmt"
	"golangProject/internal/pkg/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		log.Panic(fmt.Errorf("InitializingError: %w", err))
	}
	store.ReadStateFromFile() // считывание состояния бд из .json
	serve := server.New(":8090", &store)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		serve.Start()
	}()
	<-sigChan

	store.WriteStateToFile()
}
