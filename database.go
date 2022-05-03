package s2redis

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

type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
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

func (s *SingleStore) KeyExists(k string) (bool, error) {
	var out bool
	err := s.db.Get(&out, "echo keyExists(?)", k)
	if err != nil {
		return false, err
	}
	return out, nil
}

func (s *SingleStore) KeyDelete(k string) (bool, error) {
	var out bool
	err := s.db.Get(&out, "echo keyDelete(?)", k)
	if err != nil {
		return false, err
	}
	return out, nil
}

func (s *SingleStore) BlobSet(k string, v string) error {
	_, err := s.db.Exec("call blobSet(?, ?)", k, v)
	return err
}

func (s *SingleStore) BlobGet(k string) (string, error) {
	var out string
	err := s.db.Get(&out, "echo blobGet(?)", k)
	if err != nil {
		if sql.ErrNoRows == err {
			return "", nil
		}
		return "", err
	}
	return out, nil
}
