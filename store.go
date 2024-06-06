package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

var ErrNotFound = errors.New("no long url associated to this short url")
var errNoRowsMsg = "no rows in result set" // can't use sql.ErrNoRows because it has a prefix "sql:" which is absent somehow

type store interface {
	Save(short string, long string) error
	Long(short string) (string, error)
	Exist(short string) (bool, error)
	Close() error
}

type mem struct {
	longs map[string]string
}

func newMem() *mem {
	return &mem{make(map[string]string)}
}

func (m mem) Save(short string, long string) error {
	m.longs[short] = long
	return nil
}

func (m mem) Long(short string) (string, error) {
	long, exist := m.longs[short]
	if !exist {
		return "", ErrNotFound
	}
	return long, nil
}

func (m mem) Exist(short string) (bool, error) {
	_, exist := m.longs[short]
	return exist, nil
}

func (m mem) Close() error {
	return nil
}

type postgres struct {
	pool *pgxpool.Pool
}

func newPostgres(db string) (*postgres, error) {
	pool, err := pgxpool.New(context.Background(), db)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	return &postgres{
		pool: pool,
	}, nil
}

func (db postgres) Save(short string, long string) error {
	_, err := db.pool.Exec(context.Background(), `insert into urls(short, long) values ($1, $2)`, short, long)
	if err != nil {
		return fmt.Errorf("cannot save url: %v", err)
	}
	return nil
}

func (db postgres) Long(short string) (string, error) {
	long, err := db.long(short)
	if err != nil {
		if strings.Contains(err.Error(), errNoRowsMsg) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("cannot get url: %v", err)
	}
	return long, nil
}

func (db postgres) long(short string) (string, error) {
	var long string
	err := db.pool.QueryRow(context.Background(), "select long from urls where short=$1", short).Scan(&long)
	if err != nil {
		return "", err
	}
	return long, nil
}

func (db postgres) Exist(short string) (bool, error) {
	_, err := db.long(short)
	if err != nil {
		if strings.Contains(err.Error(), errNoRowsMsg) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (db postgres) Close() error {
	db.pool.Close()
	return nil
}
