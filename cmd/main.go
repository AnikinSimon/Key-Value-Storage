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
	ans1 := s.GetKind("key1")
	ans2 := s.GetKind("key2")
	ans3 := s.GetKind("key3")
	fmt.Println(ans1)
	fmt.Println(ans2)
	fmt.Println(ans3)
}
