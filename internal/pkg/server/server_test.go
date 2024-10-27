package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golangProject/internal/pkg/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthPage(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New(":8090", store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	serve.newAPI().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSET(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := []any{123, "val2", 123.05}
	expectedCodes := []any{http.StatusOK, http.StatusOK, http.StatusBadGateway}
	for idx, key := range testkeys {
		testVal := Entry{
			Value: testVals[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/scalar/set/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		fmt.Print(w)
		assert.Equal(t, expectedCodes[idx], w.Code)
	}
}

func TestGET(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := []any{float64(123), "val2", float64(1234)}

	for idx, key := range testkeys {
		testVal := Entry{
			Value: testVals[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/scalar/set/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)

		req, _ = http.NewRequest(http.MethodGet, "/scalar/get/"+key, nil)
		serve.newAPI().ServeHTTP(w, req)

		var val Entry
		json.Unmarshal(w.Body.Bytes(), &val)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, val.Value, testVals[idx])
	}
}

func TestHSET(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := []any{123, "val2", 123.05}
	expectedCodes := []any{http.StatusOK, http.StatusOK, http.StatusBadGateway}
	for idx, key := range testkeys {
		testVal := Entry{
			Value: testVals[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/hash/set/"+key+"/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		fmt.Print(w)
		assert.Equal(t, expectedCodes[idx], w.Code)
	}
}

func TestHGET(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := []any{float64(123), "val2", float64(1234)}

	for idx, key := range testkeys {
		testVal := Entry{
			Value: testVals[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/hash/set/"+key+"/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)

		req, _ = http.NewRequest(http.MethodGet, "/hash/get/"+key+"/"+key, nil)
		serve.newAPI().ServeHTTP(w, req)

		var val Entry
		json.Unmarshal(w.Body.Bytes(), &val)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, val.Value, testVals[idx])
	}
}

func TestLPUSH(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3", "key4"}
	testVals := [][]any{
		{1, 2, 3, 4, 5, 6},
		{"1", "2", "arr", "3"},
		{1, 2, "3", "4.04", "5", 6},
		{1, "2", "4.04", 4.04, 5, 6},
	}
	expectedCodes := []any{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusBadGateway}
	for idx, key := range testkeys {
		testVal := EntryArray{
			Value: testVals[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/array/lpush/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		fmt.Print(w)
		assert.Equal(t, expectedCodes[idx], w.Code)
	}
}

func TestRPUSH(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3", "key4"}
	testVals := [][]any{
		{1, 2, 3, 4, 5, 6},
		{"1", "2", "arr", "3"},
		{1, 2, "3", "4.04", "5", 6},
		{1, "2", "4.04", 4.04, 5, 6},
	}
	expectedCodes := []any{http.StatusOK, http.StatusOK, http.StatusOK, http.StatusBadGateway}
	for idx, key := range testkeys {
		testVal := EntryArray{
			Value: testVals[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/array/rpush/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		fmt.Print(w)
		assert.Equal(t, expectedCodes[idx], w.Code)
	}
}

func TestLPOP(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := [][]any{
		{1, 2, 3, 4, 5, 6},
		{"1", "2", "arr", "3", "1234", "2345"},
		{1, 2, "3", "4.04", "5", 6},
	}
	testSlices := [][]any{
		{2},
		{2, -2},
		{},
	}
	expectedVals := [][]any{
		{float64(1), float64(2)},
		{"arr", "3", "1234"},
		{float64(1)},
	}
	for idx, key := range testkeys {
		testVal := EntryArray{
			Value: testVals[idx],
		}
		testSlice := EntryArray{
			Value: testSlices[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/array/rpush/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		jsonVal, _ = json.MarshalIndent(testSlice, "", "\t")
		req, _ = http.NewRequest(http.MethodGet, "/array/lpop/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)

		var val Entry
		json.Unmarshal(w.Body.Bytes(), &val)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedVals[idx], val.Value)
	}
}

func TestRPOP(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := [][]any{
		{1, 2, 3, 4, 5, 6},
		{"1", "2", "arr", "3", "1234", "2345"},
		{1, 2, "3", "4.04", "5", 6},
	}
	testSlices := [][]any{
		{2},
		{2, -2},
		{},
	}
	expectedVals := [][]any{
		{float64(6), float64(5)},
		{"1234", "3", "arr"},
		{float64(6)},
	}
	for idx, key := range testkeys {
		testVal := EntryArray{
			Value: testVals[idx],
		}
		testSlice := EntryArray{
			Value: testSlices[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/array/rpush/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		jsonVal, _ = json.MarshalIndent(testSlice, "", "\t")
		req, _ = http.NewRequest(http.MethodGet, "/array/rpop/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)

		var val Entry
		json.Unmarshal(w.Body.Bytes(), &val)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedVals[idx], val.Value)
	}
}

func TestRADDTOSET(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := [][]any{
		{1, 2, 3, 4, 5, 6},
		{"1", "2", "arr", "3", "1234", "2345"},
		{1, 2, "3", "4.04", "5", 6},
	}
	testSlices := [][]any{
		{2},
		{2, -2},
		{},
	}
	expectedVals := [][]any{
		{float64(6), float64(5)},
		{"1234", "3", "arr"},
		{float64(6)},
	}
	for idx, key := range testkeys {
		testVal := EntryArray{
			Value: testVals[idx],
		}
		testSlice := EntryArray{
			Value: testSlices[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/array/rpush/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		serve.newAPI().ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		jsonVal, _ = json.MarshalIndent(testSlice, "", "\t")
		req, _ = http.NewRequest(http.MethodGet, "/array/rpop/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)

		var val Entry
		json.Unmarshal(w.Body.Bytes(), &val)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedVals[idx], val.Value)
	}
}

func TestLSET(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := [][]any{
		{1, 2, 3, 4, 5, 6},
		{"1", "2", "arr", "3"},
		{1, 2, "3"},
	}
	testArgs := [][]any{
		{1, "1"},
		{3, 1},
		{2, 0},
	}

	for idx, key := range testkeys {
		testVal := EntryArray{
			Value: testVals[idx],
		}
		testArg := EntryArray{
			Value: testArgs[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/array/lpush/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		jsonVal, _ = json.MarshalIndent(testArg, "", "\t")
		req, _ = http.NewRequest(http.MethodPost, "/array/lset/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestLGET(t *testing.T) {
	store, err := storage.NewStorage(storage.WithoutLogging())
	if err != nil {
		t.Errorf("Initialize error")
	}
	serve := New("8080", store)

	testkeys := []string{"key1", "key2", "key3"}
	testVals := [][]any{
		{1, 2, 3, 4, 5, 6},
		{"1", "2", "arr", "3"},
		{1, 2, "3"},
	}
	testArgs := [][]any{
		{1, "1"},
		{3, float64(1)},
		{2, float64(0)},
	}
	testArgsGet := [][]any{
		{1},
		{3},
		{2},
	}

	for idx, key := range testkeys {
		testVal := EntryArray{
			Value: testVals[idx],
		}
		testArg := EntryArray{
			Value: testArgs[idx],
		}
		testArgGet := EntryArray{
			Value: testArgsGet[idx],
		}
		jsonVal, _ := json.MarshalIndent(testVal, "", "\t")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/array/lpush/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		jsonVal, _ = json.MarshalIndent(testArg, "", "\t")
		req, _ = http.NewRequest(http.MethodPost, "/array/lset/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		jsonVal, _ = json.MarshalIndent(testArgGet, "", "\t")
		req, _ = http.NewRequest(http.MethodGet, "/array/lget/"+key, bytes.NewBuffer(jsonVal))
		serve.newAPI().ServeHTTP(w, req)

		var val Entry
		json.Unmarshal(w.Body.Bytes(), &val)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testArgs[idx][1], val.Value)
	}
}
