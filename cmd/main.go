package main

import (
	"fmt"
	"golangProject/internal/pkg/server"
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
		os.Exit(1)
	}

	err = store.InitializeDb()

	if err != nil {
		log.Panic(fmt.Errorf("InitializingError: %w", err))
		os.Exit(1)
	}

	store.ReadStateFromDB() // считывание состояния бд

	server_port, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		log.Panic(fmt.Errorf("NoServerPort: %w", err))
		os.Exit(1)
	}

	serve := server.New(":"+server_port, store)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		serve.Start()
	}()
	<-sigChan

	store.WriteStateToDB()
	store.WriteStateToFile()
}
