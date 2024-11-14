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
)

const jsonFolder string = "../storage_state/storage_state.json"

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

type appConfig struct {
	serverCFG serverConfig
	dbCFG     dbConfig
}

type serverConfig struct {
	Port string
}

type dbConfig struct {
	ConnectionString string
}

func getConfig() (*appConfig, error) {
	serverPort, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		return nil, errors.New("NoServerPort")
	}
	postgresUrl, ok := os.LookupEnv("POSTGRES")
	if !ok {
		return nil, errors.New("NoDbConnection")
	}
	appCfg := &appConfig{
		serverCFG: serverConfig{
			Port: serverPort,
		},
		dbCFG: dbConfig{
			ConnectionString: postgresUrl,
		},
	}
	return appCfg, nil
}

func ErrorHandler(err error) {
	log.Panic(fmt.Errorf("Error:%w", err))
	os.Exit(1)
}

func stateFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panic(fmt.Errorf("ErrorGettingCurrentWorkingDirectory: %w", err))
		return ""
	}

	return path.Join(cwd, jsonFolder)
}

func InitializeDb(appCfg *appConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", appCfg.dbCFG.ConnectionString)
	if err != nil {
		log.Panic("open", err)
		db.Close()
		return db, errors.New("Initalize errors")
	}

	if err := db.Ping(); err != nil {
		log.Panic("ping ", err)
		db.Close()
		return db, errors.New("Ping errors")
	}

	_, err = db.Exec(queryCreateCoreTable)
	if err != nil {
		log.Panic(err)
		db.Close()
		return db, err
	}

	return db, nil
}

func (r *Storage) ReadStateFromDB() error {

	if err := r.dbConnection.Ping(); err != nil {
		log.Panic("ping", err)
		return errors.New("Ping errors")
	}

	var version int64
	var timestamp int64
	var payloadJSON []byte

	err := r.dbConnection.QueryRow(queryGetLastState).Scan(
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

	state := r.getState()
	encodedState, err := json.Marshal(state)
	if err != nil {
		r.logger.Error("Json encoding error", zap.Error(err))
		return err
	}

	var lastVersion int64

	timestamp := time.Now().Unix()
	err = r.dbConnection.QueryRow(querySaveState, timestamp, encodedState).Scan(&lastVersion)

	if err != nil {
		r.logger.Error("Write to db error", zap.Error(err))
		return err
	}

	_, err = r.dbConnection.Exec(queryDeleteState, lastVersion)

	if err != nil {
		r.logger.Error("Delete expire state error", zap.Error(err))
		return err
	}

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

func (r *Storage) GracefulShutdown() {
	defer r.dbConnection.Close()
	r.WriteStateToDB()
	r.WriteStateToFile()
}
