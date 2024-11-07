package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"

	"go.uber.org/zap"
)

const accessCode fs.FileMode = 0o777

const (
	queryCreateCoreTable = `CREATE TABLE IF NOT EXISTS core (
		version bigserial PRIMARY KEY,
		timestamp bigint NOT NULL,
		payload JSONB NOT NULL
	)`
	querySaveState = `INSERT INTO core (timestamp, payload) 
	VALUES ($1, $2) RETURNING version`
	queryDeleteState = `DELETE FROM core
	WHERE version = $1
	`
	queryGetLastState = `SELECT * FROM core ORDER BY version DESC LIMIT 1`
	queryGetALL       = `SELECT * FROM core ORDER BY version DESC LIMIT 1`
)

type Task struct {
	Command string `json:"command"`
	Key     string `json:"key"`
	Args    []any  `json:"args,omitempty"`
}

type DbState struct {
	Version   int64            `json:"version"`
	Timestamp int64            `json:"timestamp"`
	Payload   StorageCondition `json:"payload"`
}

func stateFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(fmt.Errorf("ErrorGettingCurrentWorkingDirectory: %w", err))
		return ""
	}

	return path.Join(cwd, "../storage_state/storage_state.json")
}

func (r *Storage) InitializeDb() error {
	postgresURI := os.Getenv("POSTGRES")
	db, err := sql.Open("postgres", postgresURI)
	if err != nil {
		log.Panic("open", err)
		return errors.New("Initalize errors")
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Panic("ping ", err)
		return errors.New("Ping errors")
	}

	_, err = db.Exec(queryCreateCoreTable)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (r *Storage) ReadStateFromDB() error {
	postgresURI := os.Getenv("POSTGRES")
	db, err := sql.Open("postgres", postgresURI)
	if err != nil {
		log.Panic("open", err)
		return errors.New("Initalize errors")
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Panic("ping", err)
		return errors.New("Ping errors")
	}

	var version int64
	var timestamp int64
	var payloadJSON []byte

	err = db.QueryRow(queryGetLastState).Scan(
		&version,
		&timestamp,
		&payloadJSON,
	)
	if err != nil {
		return nil
	}

	var state StorageCondition

	err = json.Unmarshal(payloadJSON, &state)
	if err != nil {
		r.logger.Error("Error decoding file:", zap.Error(err))
		return err
	}

	r.recoverFromCondition(state)

	return nil
}

func (r *Storage) WriteStateToDB() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	postgresURI := os.Getenv("POSTGRES")
	db, err := sql.Open("postgres", postgresURI)
	if err != nil {
		log.Panic("open", err)
		return errors.New("Initalize errors")
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Panic("ping", err)
		return errors.New("Ping errors")
	}

	state := r.getState()
	fmt.Println(state)
	encodedState, err := json.Marshal(state)
	if err != nil {
		r.logger.Error("Json encoding error", zap.Error(err))
		return err
	}

	var lastVersion int64

	timestamp := time.Now().Unix()
	err = db.QueryRow(querySaveState, timestamp, encodedState).Scan(&lastVersion)

	if err != nil {
		r.logger.Error("Write to db error", zap.Error(err))
		return err
	}

	_, err = db.Exec(queryDeleteState, lastVersion)

	if err != nil {
		r.logger.Error("Delete expire state error", zap.Error(err))
		return err
	}

	// all, err := db.Exec(queryGetALL)
	// if err != nil {
	// 	r.logger.Error("Get all error", zap.Error(err))
	// 	return err
	// }

	// fmt.Println(all)

	return nil
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
