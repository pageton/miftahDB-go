package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pageton/miftahDB-go/encoding"
	"github.com/pageton/miftahDB-go/types"
)

type BaseMiftahDB struct {
	db *sql.DB
}

func NewBaseMiftahDB(path string) (*BaseMiftahDB, error) {
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	miftahDB := &BaseMiftahDB{db: database}
	miftahDB.initDatabase()
	return miftahDB, nil
}

func (db *BaseMiftahDB) initDatabase() {
	_, err := db.db.Exec(SQLStatements.CreateTable)
	if err != nil {
		fmt.Println("Error creating table:", err)
	}
}

func (db *BaseMiftahDB) Get(key string) (*types.Entry, error) {
	row := db.db.QueryRow(SQLStatements.Get, key)
	var encodedValue []byte
	var expiresAt int64
	err := row.Scan(&encodedValue, &expiresAt)
	if err != nil {
		return nil, err
	}
	if expiresAt != 0 && expiresAt <= time.Now().Unix() {
		db.Delete(key)
		return nil, nil
	}
	value, err := encoding.DecodeValue[string](encodedValue)
	if err != nil {
		return nil, err
	}
	entry := &types.Entry{
		Key:       key,
		Value:     value,
		ExpiresAt: &time.Time{},
	}
	if expiresAt != 0 {
		expiresAtTime := time.Unix(expiresAt, 0)
		entry.ExpiresAt = &expiresAtTime
	}
	return entry, nil
}

func (db *BaseMiftahDB) Set(key string, value types.MiftahValue, expiresAt *time.Time) error {
	encodedValue, err := encoding.EncodeValue(value)
	if err != nil {
		return err
	}

	var exp int64
	if expiresAt != nil {
		exp = expiresAt.Unix()
	}
	_, err = db.db.Exec(SQLStatements.Set, key, encodedValue, exp)
	return err
}

func (db *BaseMiftahDB) Exists(key string) bool {
	var exists int
	row := db.db.QueryRow(SQLStatements.Exists, key)
	err := row.Scan(&exists)
	if err != nil {
		return false
	}
	return exists > 0
}

func (db *BaseMiftahDB) Delete(key string) error {
	_, err := db.db.Exec(SQLStatements.Delete, key)
	return err
}

func (db *BaseMiftahDB) Rename(oldKey, newKey string) error {
	_, err := db.db.Exec(SQLStatements.Rename, newKey, oldKey)
	return err
}

func (db *BaseMiftahDB) SetExpire(key string, expiresAt time.Time) error {
	expiresAtMs := expiresAt.Unix()
	_, err := db.db.Exec(SQLStatements.SetExpire, expiresAtMs, key)
	return err
}

func (db *BaseMiftahDB) GetExpire(key string) (*time.Time, error) {
	var expiresAt int64
	row := db.db.QueryRow(SQLStatements.GetExpire, key)
	err := row.Scan(&expiresAt)
	if err != nil {
		return nil, err
	}
	if expiresAt == 0 {
		return nil, nil
	}
	expiresAtTime := time.Unix(expiresAt, 0)
	return &expiresAtTime, nil
}

func (db *BaseMiftahDB) Keys(pattern string) ([]string, error) {
	rows, err := db.db.Query(SQLStatements.Keys, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (db *BaseMiftahDB) Pagination(limit, page int, pattern string) ([]string, error) {
	offset := (page - 1) * limit
	rows, err := db.db.Query(SQLStatements.Pagination, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (db *BaseMiftahDB) Count(pattern string) (int, error) {
	var count int
	row := db.db.QueryRow(SQLStatements.CountKeys, pattern)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *BaseMiftahDB) CountExpired(pattern string) (int, error) {
	var count int
	row := db.db.QueryRow(SQLStatements.CountExpired, pattern)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *BaseMiftahDB) MultiGet(keys []string) (map[string]*types.Entry, error) {
	result := make(map[string]*types.Entry)
	for _, key := range keys {
		entry, err := db.Get(key)
		if err != nil {
			return nil, err
		}
		result[key] = entry
	}
	return result, nil
}

func (db *BaseMiftahDB) MultiSet(entries []types.Entry) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(SQLStatements.Set)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		encodedValue, err := encoding.EncodeValue(entry.Value)
		if err != nil {
			return err
		}

		var exp int64
		if entry.ExpiresAt != nil {
			exp = entry.ExpiresAt.Unix()
		}
		_, err = stmt.Exec(entry.Key, encodedValue, exp)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (db *BaseMiftahDB) MultiDelete(keys []string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(SQLStatements.Delete)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, key := range keys {
		_, err = stmt.Exec(key)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (db *BaseMiftahDB) Vacuum() error {
	_, err := db.db.Exec(SQLStatements.Vacuum)
	return err
}

func (db *BaseMiftahDB) Close() error {
	return db.db.Close()
}

func (db *BaseMiftahDB) Flush() error {
	_, err := db.db.Exec(SQLStatements.Flush)
	return err
}

func (db *BaseMiftahDB) Cleanup() error {
	_, err := db.db.Exec(SQLStatements.Cleanup, time.Now().Unix())
	return err
}

func (db *BaseMiftahDB) Backup(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = db.db.Exec("BEGIN IMMEDIATE")
	if err != nil {
		return err
	}

	rows, err := db.db.Query("SELECT name, sql FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, sql string
		err = rows.Scan(&name, &sql)
		if err != nil {
			return err
		}
		file.WriteString(sql + ";\n")
	}

	err = tx.Commit()
	return err
}
