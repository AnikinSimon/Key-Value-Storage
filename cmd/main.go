package main

import (
	"GolangCourse/internal/pkg/storage"
	"fmt"
	"log"
)

func main() {
	var s, err = storage.NewStorage()

	if err != nil {
		log.Fatal(err)
	}

	s.Set("key1", "val1")
	s.Set("key2", "123")
	type1 := s.GetKind("key1")
	type2 := s.GetKind("key2")
	type3 := s.GetKind("key3")
	fmt.Println(type1)
	fmt.Println(type2)
	fmt.Println(type3)
	val1 := s.Get("key1")
	if val1 != nil {
		fmt.Println(*val1)
	} else {
		fmt.Println("KeyError")
	}
	val2 := s.Get("key3")
	if val2 != nil {
		fmt.Println(*val2)
	} else {
		fmt.Println("KeyError")
	}
}
