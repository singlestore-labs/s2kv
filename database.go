package s2kv

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type SingleStore struct {
	db *sqlx.DB
}

func NewSingleStore(config DatabaseConfig) (*SingleStore, error) {
	// We use NewConfig here to set default values. Then we override what we need to.
	mysqlConf := mysql.NewConfig()
	mysqlConf.User = config.Username
	mysqlConf.Passwd = config.Password
	mysqlConf.DBName = config.Database
	mysqlConf.Addr = fmt.Sprintf("%s:%s", config.Host, config.Port)
	mysqlConf.ParseTime = true
	mysqlConf.Timeout = 10 * time.Second
	mysqlConf.InterpolateParams = true
	mysqlConf.AllowNativePasswords = true
	mysqlConf.MultiStatements = false

	mysqlConf.Params = map[string]string{
		"collation_server":    "utf8_general_ci",
		"sql_select_limit":    "18446744073709551615",
		"compile_only":        "false",
		"enable_auto_profile": "false",
		"sql_mode":            "'STRICT_ALL_TABLES'",
	}

	log.Println("Connecting to SingleStore database...", mysqlConf.Addr)
	connector, err := mysql.NewConnector(mysqlConf)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)

	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(20)

	return &SingleStore{db: sqlx.NewDb(db, "mysql")}, nil
}

func (s *SingleStore) Close() error {
	return s.db.Close()
}

func (s *SingleStore) FlushAll() error {
	_, err := s.db.Exec("call flushAll()")
	return err
}

func (s *SingleStore) KeyExists(k string) (bool, error) {
	var out bool
	err := s.db.Get(&out, "select * from keyExists(?)", k)
	if err != nil {
		return false, err
	}
	return out, nil
}

func (s *SingleStore) Keys(pattern string) ([][]byte, error) {
	var out [][]byte
	err := s.db.Select(&out, "select k from getKeys(?)", pattern)
	return out, err
}

func (s *SingleStore) KeyDelete(k string) (bool, error) {
	var out bool
	err := s.db.Get(&out, "echo keyDelete(?)", k)
	if err != nil {
		return false, err
	}
	return out, nil
}

func (s *SingleStore) BlobSet(k string, v []byte) error {
	_, err := s.db.Exec("call blobSet(?, ?)", k, v)
	return err
}

func (s *SingleStore) BlobGet(k string) ([]byte, error) {
	var out []byte
	err := s.db.Get(&out, "select v from blobGet(?)", k)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SingleStore) IncrBy(k string, v int64) (int64, error) {
	var out int64
	err := s.db.Get(&out, "echo incrBy(?, ?)", k, v)
	if err != nil {
		return 0, err
	}
	return out, nil
}

func (s *SingleStore) ListAppend(k string, v []byte) error {
	_, err := s.db.Exec("call listAppend(?, ?)", k, v)
	return err
}

func (s *SingleStore) ListRemove(k string, v []byte) (int64, error) {
	var out int64
	err := s.db.Get(&out, "echo listRemove(?, ?)", k, v)
	if err != nil {
		return 0, err
	}
	return out, nil
}

func (s *SingleStore) ListGet(k string) ([][]byte, error) {
	var out [][]byte
	err := s.db.Select(&out, "select v from listGet(?)", k)
	return out, err
}

func (s *SingleStore) ListRange(k string, start, end int) ([][]byte, error) {
	var out [][]byte
	err := s.db.Select(&out, "select v from listRange(?, ?, ?)", k, start, end)
	return out, err
}

func (s *SingleStore) SetAdd(k string, v []byte) error {
	_, err := s.db.Exec("call setAdd(?, ?)", k, v)
	return err
}

func (s *SingleStore) SetRemove(k string, v []byte) (int64, error) {
	var out int64
	err := s.db.Get(&out, "echo setRemove(?, ?)", k, v)
	if err != nil {
		return 0, err
	}
	return out, nil
}

func (s *SingleStore) SetGet(k string) ([][]byte, error) {
	var out [][]byte
	err := s.db.Select(&out, "select v from setGet(?)", k)
	return out, err
}

func (s *SingleStore) SetUnion(keys ...string) ([][]byte, error) {
	var out [][]byte

	query, args, err := sqlx.In("echo setUnion([?])", keys)
	if err != nil {
		return out, err
	}

	err = s.db.Select(&out, query, args...)
	return out, err
}

func (s *SingleStore) SetIntersect(keys ...string) ([][]byte, error) {
	var out [][]byte

	query, args, err := sqlx.In("echo setIntersect([?])", keys)
	if err != nil {
		return out, err
	}

	err = s.db.Select(&out, query, args...)
	return out, err
}

func (s *SingleStore) SetsWithMember(v []byte) ([]string, error) {
	var out []string
	err := s.db.Select(&out, "select k from setsWithMember(?)", v)
	return out, err
}

func (s *SingleStore) SetCardinality(k string) (int64, error) {
	var out int64
	err := s.db.Get(&out, "select * from setCardinality(?)", k)
	if err != nil {
		return 0, err
	}
	return out, nil
}

func (s *SingleStore) SetIntersectCardinality(keys ...string) (int64, error) {
	var out int64

	query, args, err := sqlx.In("echo setIntersectCardinality([?])", keys)
	if err != nil {
		return out, err
	}

	err = s.db.Get(&out, query, args...)
	if err != nil {
		return 0, err
	}
	return out, nil
}
