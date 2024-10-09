package storage

import (
	"strconv"
	"testing"
)

func TestGet(t *testing.T) {
	var s, err = NewStorage()

	if err != nil {
		t.Errorf("Initialize error")
	}

	keys := []string{"key1", "key2", "key3"}
	vals := []string{"val1", "123", "234.05"}
	expectedVals := []string{"val1", "123", "234.05"}
	for i, k := range keys {
		s.Set(k, vals[i])
		if actualVal := *s.Get(keys[i]); actualVal != expectedVals[i] {
			t.Errorf("Wrong Value by key. Actual: %s. Expected: %s", actualVal, expectedVals[i])
		}
	}
	wrongKeys := []string{"key6", "key4", "key5"}
	for _, k := range wrongKeys {
		if actualVal := s.Get(k); actualVal != nil {
			t.Errorf("Get value for unexisting key %s", k)
		}
	}
}

func TestGetKind(t *testing.T) {
	var s, err = NewStorage()

	if err != nil {
		t.Errorf("Initialize error")
	}

	keys := []string{"key1", "key2", "key3"}
	vals := []string{"val1", "123", "234.05"}
	expectedKinds := []string{"S", "D", "S"}
	for i, k := range keys {
		s.Set(k, vals[i])
		if actualKind := s.GetKind(keys[i]); actualKind != expectedKinds[i] {
			t.Errorf("Wrong Kind by key Actual: %s. Expected: %s", actualKind, expectedKinds[i])
		}
	}
}

func BenchmarkPrivateGet(b *testing.B) {
	s, err := NewStorage()
	if err != nil {
		return
	}

	s.SetLoggerLevel("fatal")

	for i := 0; i < b.N; i++ {
		s.Set(strconv.Itoa(i), strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.get(strconv.Itoa(i))
	}

}
