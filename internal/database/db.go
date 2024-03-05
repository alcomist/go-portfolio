// Copyright 2024 30K Dev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package database

import (
	"fmt"
	"github.com/alcomist/go-portfolio/internal/config"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
)

type DBConfig struct {
	mu     sync.Mutex
	config map[string]*mysql.Config
}

var dbConfig DBConfig

func init() {
	dbConfig.config = make(map[string]*mysql.Config)
}

func (c *DBConfig) Config(s string) *mysql.Config {

	c.mu.Lock()
	defer c.mu.Unlock()

	mysqlConfig, ok := c.config[s]
	if ok && mysqlConfig != nil {
		return mysqlConfig
	}

	mysqlConfig = mysql.NewConfig()

	section := config.MustGet(s)

	port := 3306
	host := section.Key("host").String()

	p, err := section.Key("port").Int()
	if err == nil {
		port = p
	}

	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = fmt.Sprintf("%s:%d", host, port)
	mysqlConfig.User = section.Key("username").String()
	mysqlConfig.Passwd = section.Key("password").String()
	mysqlConfig.DBName = section.Key("dbname").String()

	mysqlConfig.Params = make(map[string]string)
	mysqlConfig.Params["charset"] = section.Key("charset").String()

	c.config[s] = mysqlConfig

	return mysqlConfig
}

type DB struct {
	*sqlx.DB
}

type MysqlDB struct {
	mu sync.Mutex
	db map[string]*DB
}

var mysqlDB MysqlDB

func init() {
	mysqlDB.db = make(map[string]*DB)
}

func MustGet(s string) *DB {

	mysqlDB.mu.Lock()
	defer mysqlDB.mu.Unlock()

	db, ok := mysqlDB.db[s]
	if ok && db != nil {
		return db
	}

	cfg := dbConfig.Config(s)

	dsn := cfg.FormatDSN()

	sqlxDB, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	mysqlDB.db[s] = &DB{sqlxDB}
	return mysqlDB.db[s]
}
