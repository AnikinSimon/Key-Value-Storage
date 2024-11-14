package storage_test

import (
	"golangProject/internal/pkg/storage"
	"strconv"
	"testing"
)

func BenchmarkGet(b *testing.B) {
	s, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		return
	}

	for i := 0; i < b.N; i++ {
		s.SET(strconv.Itoa(i), strconv.Itoa(i), 0)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.GET(strconv.Itoa(i))
	}
}

func BenchmarkGetKind(b *testing.B) {
	s, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		return
	}

	for i := 0; i < b.N; i++ {
		s.SET(strconv.Itoa(i), strconv.Itoa(i), 0)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.GetKind(strconv.Itoa(i))
	}

}

func BenchmarkSet(b *testing.B) {
	s, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		return
	}

	for i := 0; i < b.N; i++ {
		keyVal := strconv.Itoa(i)
		s.SET(keyVal, keyVal, 0)
	}
}

func BenchmarkGetSet(b *testing.B) {
	s, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		return
	}

	for i := 0; i < b.N; i++ {
		keyVal := strconv.Itoa(i)
		s.SET(keyVal, keyVal, 0)
		s.GET(keyVal)
	}
}
