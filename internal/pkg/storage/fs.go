package storage

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"go.uber.org/zap"
)

const accessCode fs.FileMode = 0o777

type Task struct {
	Command string `json:"command"`
	Key     string `json:"key"`
	Args    []any  `json:"args,omitempty"`
}

func cmdFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(fmt.Errorf("ErrorGettingCurrentWorkingDirectory: %w", err))
		return ""
	}

	return path.Join(cwd, "commands.json")
}

func stateFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(fmt.Errorf("ErrorGettingCurrentWorkingDirectory: %w", err))
		return ""
	}

	return path.Join(cwd, "storage_state.json")
}

func (r *Storage) ReadTasksFromFile() ([]Task, error) {
	filePath := cmdFilePath()
	// Check if file exists
	// if doesn't exist, return empty list and create the file
	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		r.logger.Warn("File does not exist. Creating file...")
		file, err := os.Create(filePath)
		os.WriteFile(filePath, []byte("[]"), os.ModeAppend.Perm())

		if err != nil {
			r.logger.Error("Error creating file: ", zap.Error(err))
			return nil, err
		}

		defer file.Close()

		return []Task{}, nil
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		r.logger.Error("Error opening file:", zap.Error(err))
		return nil, err
	}

	tasks := []Task{}
	err = json.Unmarshal(file, &tasks)
	if err != nil {
		r.logger.Error("Error decoding file:", zap.Error(err))
		return nil, err
	}

	return tasks, nil
}

func (r *Storage) ReadStateFromFile() error {
	filePath := stateFilePath()

	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		r.logger.Warn("File does not exist.")
		return err
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		r.logger.Error("Error opening file:", zap.Error(err))
		return err
	}

	state := StorageCondition{}
	err = json.Unmarshal(file, &state)
	if err != nil {
		r.logger.Error("Error decoding file:", zap.Error(err))
		return err
	}

	r.recoverFromCondition(state)

	return nil
}

func writeAtomic(path string, b []byte) error {
	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	tmpPathName := filepath.Join(dir, filename+".tmp")
	err := os.WriteFile(tmpPathName, b, accessCode)
	if err != nil {
		return err
	}

	defer func() {
		os.Remove(tmpPathName)
	}()

	return os.Rename(tmpPathName, path)
}

func (r *Storage) WriteStateToFile() error {
	filePath := stateFilePath()

	state := r.getState()
	encodedState, err := json.MarshalIndent(state, "", "\t")
	if err != nil {
		r.logger.Error("Json encoding error", zap.Error(err))

		return err
	}

	err = writeAtomic(filePath, encodedState)
	if err != nil {
		r.logger.Panic("Write File Error", zap.Error(err))

		return err
	}

	return nil
}
