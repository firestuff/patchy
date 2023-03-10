package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/firestuff/patchy/metadata"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(conn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", conn)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{
		db: db,
	}, nil
}

func (sls *SQLiteStore) Close() {
	sls.db.Close()
}

func (sls *SQLiteStore) Write(t string, obj any) error {
	id := metadata.GetMetadata(obj).ID

	js, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = sls.exec("INSERT INTO `%s` (id, obj) VALUES (?,?) ON CONFLICT(id) DO UPDATE SET obj=?;", t, id, js, js)
	if err != nil {
		return err
	}

	return nil
}

func (sls *SQLiteStore) Delete(t string, id string) error {
	_, err := sls.exec("DELETE FROM `%s` WHERE id=?", t, id)
	if err != nil {
		return err
	}

	return nil
}

func (sls *SQLiteStore) Read(t string, id string, factory func() any) (any, error) {
	rows, err := sls.query("SELECT obj FROM `%s` WHERE id=?;", t, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var js []byte

		err = rows.Scan(&js)
		if err != nil {
			return nil, err
		}

		obj := factory()

		err = json.Unmarshal(js, obj)
		if err != nil {
			return nil, err
		}

		return obj, nil
	}

	return nil, nil
}

func (sls *SQLiteStore) List(t string, factory func() any) ([]any, error) {
	rows, err := sls.query("SELECT obj FROM `%s`;", t)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ret := []any{}

	for rows.Next() {
		var js []byte

		err = rows.Scan(&js)
		if err != nil {
			return nil, err
		}

		obj := factory()

		err = json.Unmarshal(js, obj)
		if err != nil {
			return nil, err
		}

		ret = append(ret, obj)
	}

	return ret, nil
}

func (sls *SQLiteStore) exec(query, t string, args ...any) (sql.Result, error) {
	query = fmt.Sprintf(query, t)

	result, err := sls.db.Exec(query, args...)
	if err == nil {
		return result, nil
	}

	_, err = sls.db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (id TEXT NOT NULL PRIMARY KEY, obj TEXT NOT NULL);", t))
	if err != nil {
		return nil, err
	}

	return sls.db.Exec(query, args...)
}

func (sls *SQLiteStore) query(query, t string, args ...any) (*sql.Rows, error) {
	query = fmt.Sprintf(query, t)

	rows, err := sls.db.Query(query, args...)
	if err == nil {
		return rows, nil
	}

	_, err = sls.db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (id TEXT NOT NULL PRIMARY KEY, obj TEXT NOT NULL);", t))
	if err != nil {
		return nil, err
	}

	return sls.db.Query(query, args...)
}
