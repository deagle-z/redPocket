package utils

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"os"
	"strings"
	"time"
)

var Db *gorm.DB

var dbPool = make(map[string]*gorm.DB)

func NewPrefixDb(prefix string) (db *gorm.DB) {
	if existingDb, ok := dbPool[prefix]; ok {
		return existingDb
	}
	db = Db.Session(&gorm.Session{
		NewDB: true,
	})
	err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", prefix)).Error
	if err != nil {
		log.Printf("创建数据库错误 %s", err.Error())
		return nil
	}
	masterStr := fmt.Sprintf(GlobalConfig.Mysql.Master, prefix)
	slaveStr := fmt.Sprintf(GlobalConfig.Mysql.Slave, prefix)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			//LogLevel:      logger.Info, // Log level
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,         // Don't include params in the SQL log
			Colorful:                  true,         // Disable color
		},
	)
	newDb, err := gorm.Open(mysql.Open(masterStr), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("连接数据库错误 %s", err.Error())
		return nil
	}
	err = newDb.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(masterStr)}, // 主库，写操作
		Replicas: []gorm.Dialector{mysql.Open(slaveStr)},  // 从库，读操作
		Policy:   dbresolver.RandomPolicy{},               // 读库负载均衡策略
	}).SetConnMaxIdleTime(30 * time.Second).
		SetConnMaxLifetime(5 * time.Minute).
		SetMaxIdleConns(2).
		SetMaxOpenConns(5))
	if err != nil {
		panic(err)
		return
	}
	sqlDB, err := newDb.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(2)                   // 设置最大空闲连接数
	sqlDB.SetMaxOpenConns(5)                   // 设置最大连接数
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 设置连接保持时间
	sqlDB.SetConnMaxIdleTime(30 * time.Second) // 设置闲置保持时间
	ctx := context.WithValue(context.Background(), KeyDbPrefix, prefix)
	newDb = newDb.WithContext(ctx)
	dbPool[prefix] = newDb
	return newDb
}

func GetDbPrefix(db *gorm.DB) (prefix string) {
	prefixObj := db.Statement.Context.Value(KeyDbPrefix)
	if prefixObj != nil && !strings.HasPrefix(db.Statement.Table, prefixObj.(string)) {
		prefix = prefixObj.(string)
	}
	return prefix
}

func AddTablePrefix(db *gorm.DB, username string) (result *gorm.DB) {
	ctx := context.WithValue(context.Background(), KeyTablePrefix, username)
	ctx = context.WithValue(ctx, KeyDbPrefix, GetDbPrefix(db))
	return db.WithContext(ctx)
}

func LimitPrefix(prefixBack func(prefix string)) {
	runKeys := make([]string, 0)
	hostInfos := GetTempHostInfos()
	for _, hostInfo := range hostInfos {
		if InStrings(runKeys, hostInfo.TablePrefix) {
			continue
		}
		runKeys = append(runKeys, hostInfo.TablePrefix)
		prefixBack(hostInfo.TablePrefix)
	}
}
