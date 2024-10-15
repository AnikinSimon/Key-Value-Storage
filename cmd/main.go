package main

import (
	"encoding/json"
	"golangProject/internal/pkg/storage"
	"golangProject/internal/pkg/task"
	"log"
	"os"
)

func main() {

	fromflie, err := os.ReadFile("tasks.json")
	if err != nil {
		log.Panic("Read Fail")
	}

	var decoded []task.Task
	err = json.Unmarshal(fromflie, &decoded)
	if err != nil {
		log.Panic("Unmarshal Fail")
	}

	store, err := storage.NewStorage()
	if err != nil {
		log.Panic("Initialization failure!")
	}

	results := make([]task.TaskRes, 0)
	for _, iter := range decoded {
		switch iter.Command {
		case "LPUSH":
			store.LPUSH(iter.Key, iter.Args)
		case "RPUSH":
			store.RPUSH(iter.Key, iter.Args)
		case "RADDTOSET":
			store.RADDTOSET(iter.Key, iter.Args)
		case "LPOP":
			vals, err := store.LPOP(iter.Key, iter.Args)
			if err != nil {
				results = append(results, task.TaskRes{
					Result: make([]int, 0),
					Errors: err.Error(),
				})
			} else {
				results = append(results, task.TaskRes{
					Result: vals,
					Errors: "",
				})
			}
		case "RPOP":
			vals, err := store.RPOP(iter.Key, iter.Args)
			if err != nil {
				results = append(results, task.TaskRes{
					Result: make([]int, 0),
					Errors: err.Error(),
				})
			} else {
				results = append(results, task.TaskRes{
					Result: vals,
					Errors: "",
				})
			}
		case "LSET":
			err := store.LSET(iter.Key, iter.Args)
			if err != nil {
				results = append(results, task.TaskRes{
					Result: make([]int, 0),
					Errors: err.Error(),
				})
			} else {
				results = append(results, task.TaskRes{
					Result: make([]int, 0),
					Errors: "OK",
				})
			}
		case "LGET":
			val, err := store.LGET(iter.Key, iter.Args)
			if err != nil {
				results = append(results, task.TaskRes{
					Result: make([]int, 0),
					Errors: err.Error(),
				})
			} else {
				res := make([]int, 0, val)
				res = append(res, val)
				results = append(results, task.TaskRes{
					Result: res,
					Errors: "OK",
				})
			}

		}
	}
	b, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		log.Panic("Marshal Fail")
	}

	err = os.WriteFile("results.json", b, 0o777)
	if err != nil {
		log.Panic("Write File Error")
	}
}
