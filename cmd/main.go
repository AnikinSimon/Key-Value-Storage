package main

import (
	"fmt"
	"golangProject/internal/pkg/storage"
	"log"
)

func main() {
	var s, err = storage.NewStorage()

	if err != nil {
		log.Fatal(err)
	}

	s.SetLoggerLevel("fatal")

	s.Set("key1", "val1")
	s.Set("key2", "123")

	fmt.Println(s.Get("key1"))
	fmt.Println(s.Get("key2"))
	fmt.Println(s.Get("key3"))

	fmt.Println(s.GetKind("key1"))
	fmt.Println(s.GetKind("key2"))
	fmt.Println(s.GetKind("key3"))

}
