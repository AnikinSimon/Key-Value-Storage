package storage_test

import (
	"golangProject/internal/pkg/storage"
	"strconv"
	"testing"
)

func BenchmarkGet(b *testing.B) {
	s, err := storage.NewStorage()
	if err != nil {
		return
	}

	s.SetLoggerLevel("fatal")

	for i := 0; i < b.N; i++ {
		s.Set(strconv.Itoa(i), strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Get(strconv.Itoa(i))
	}
}

// Отдельный бенчмарк для приватного гета лежит в файле storage_test.go в папке storage

func BenchmarkGetKind(b *testing.B) {
	s, err := storage.NewStorage()
	if err != nil {
		return
	}
	s.SetLoggerLevel("fatal")

	for i := 0; i < b.N; i++ {
		s.Set(strconv.Itoa(i), strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.GetKind(strconv.Itoa(i))
	}

}

func BenchmarkSet(b *testing.B) {
	s, err := storage.NewStorage()
	if err != nil {
		return
	}
	s.SetLoggerLevel("fatal")

	for i := 0; i < b.N; i++ {
		keyVal := strconv.Itoa(i)
		s.Set(keyVal, keyVal)
	}
}

func BenchmarkGetSet(b *testing.B) {
	s, err := storage.NewStorage()
	if err != nil {
		return
	}
	s.SetLoggerLevel("fatal")

	for i := 0; i < b.N; i++ {
		keyVal := strconv.Itoa(i)
		s.Set(keyVal, keyVal)
		s.Get(keyVal)
	}
}
