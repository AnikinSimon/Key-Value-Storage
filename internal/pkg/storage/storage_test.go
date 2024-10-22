package storage

import (
	"strconv"
	"testing"
)

func TestGet(t *testing.T) {
	s, err := NewStorage()
	if err != nil {
		t.Errorf("Initialize error")
	}

	keys := []string{"key1", "key2", "key3"}
	vals := []string{"val1", "123", "234.05"}
	expectedVals := []string{"val1", "123", "234.05"}
	for i, k := range keys {
		s.SET(k, vals[i])
		if actualVal := *s.GET(keys[i]); actualVal != expectedVals[i] {
			t.Errorf("Wrong Value by key. Actual: %s. Expected: %s", actualVal, expectedVals[i])
		}
	}
}

func TestWrongGet(t *testing.T) {
	s, err := NewStorage()
	if err != nil {
		t.Errorf("Initialize error")
	}

	keys := []string{"key1", "key2", "key3"}
	vals := []string{"val1", "123", "234.05"}
	wrongKeys := []string{"key6", "key4", "key5"}
	for i, k := range keys {
		s.SET(k, vals[i])
		if actualVal := s.GET(wrongKeys[i]); actualVal != nil {
			t.Errorf("Get value for unexisting key %s", k)
		}
	}
}

func TestGetKind(t *testing.T) {
	s, err := NewStorage()
	if err != nil {
		t.Errorf("Initialize error")
	}

	keys := []string{"key1", "key2", "key3"}
	vals := []string{"val1", "123", "234.05"}
	expectedKinds := []string{"S", "D", "S"}
	for i, k := range keys {
		s.SET(k, vals[i])
		if actualKind, _ := s.GetKind(keys[i]); string(actualKind) != expectedKinds[i] {
			t.Errorf("Wrong Kind by key Actual: %s. Expected: %s", actualKind, expectedKinds[i])
		}
	}
}

func BenchmarkPrivateGet(b *testing.B) {
	s, err := NewStorage()
	if err != nil {
		return
	}

	for i := 0; i < b.N; i++ {
		s.SET(strconv.Itoa(i), strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.get(strconv.Itoa(i))
	}

}
