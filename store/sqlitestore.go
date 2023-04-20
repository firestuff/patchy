package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gopatchy/metadata"
	// Register sqlite3 db handler.
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

	// TODO: Keep a set of prepared statements with PrepareContext()
	// TODO: Consider tuning per https://phiresky.github.io/blog/2020/sqlite-performance-tuning/

	return &SQLiteStore{
		db: db,
	}, nil
}

func (sls *SQLiteStore) Close() {
	sls.db.Close()
}

func (sls *SQLiteStore) Write(ctx context.Context, t string, obj any) error {
	id := metadata.GetMetadata(obj).ID

	js, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	err = sls.exec(ctx, "INSERT INTO `%s` (id, obj) VALUES (?,?) ON CONFLICT(id) DO UPDATE SET obj=?;", t, id, js, js)
	if err != nil {
		return err
	}

	return nil
}

func (sls *SQLiteStore) Delete(ctx context.Context, t, id string) error {
	err := sls.exec(ctx, "DELETE FROM `%s` WHERE id=?", t, id)
	if err != nil {
		return err
	}

	return nil
}

func (sls *SQLiteStore) Read(ctx context.Context, t, id string, factory func() any) (any, error) {
	rows, err := sls.query(ctx, "SELECT obj FROM `%s` WHERE id=?;", t, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

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

func (sls *SQLiteStore) List(ctx context.Context, t string, factory func() any) ([]any, error) {
	rows, err := sls.query(ctx, "SELECT obj FROM `%s`;", t)
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

func (sls *SQLiteStore) exec(ctx context.Context, query, t string, args ...any) error {
	query = fmt.Sprintf(query, t)

	_, err := sls.db.ExecContext(ctx, query, args...)
	if err == nil {
		return nil
	}

	_, err = sls.db.ExecContext(ctx, sls.tableSQL(t))
	if err != nil {
		return err
	}

	_, err = sls.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (sls *SQLiteStore) query(ctx context.Context, query, t string, args ...any) (*sql.Rows, error) {
	query = fmt.Sprintf(query, t)

	rows, err := sls.db.QueryContext(ctx, query, args...)
	if err == nil {
		return rows, nil
	}

	_, err = sls.db.ExecContext(ctx, sls.tableSQL(t))
	if err != nil {
		return nil, err
	}

	return sls.db.QueryContext(ctx, query, args...)
}

func (sls *SQLiteStore) tableSQL(t string) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (id TEXT NOT NULL PRIMARY KEY, obj TEXT NOT NULL);", t)
}
