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
	// 连接池参数（可在 core.yaml mysql 下配置；0 用默认）。原值 5/2 过小，高并发下请求排队等连接导致接口变慢。
	maxOpen := GlobalConfig.Mysql.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 30
	}
	maxIdle := GlobalConfig.Mysql.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}
	connMaxLifetime := time.Duration(GlobalConfig.Mysql.ConnMaxLifetimeMinutes) * time.Minute
	if connMaxLifetime <= 0 {
		connMaxLifetime = 30 * time.Minute
	}
	connMaxIdleTime := time.Duration(GlobalConfig.Mysql.ConnMaxIdleTimeMinutes) * time.Minute
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = 5 * time.Minute
	}

	err = newDb.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{mysql.Open(masterStr)}, // 主库，写操作
		Replicas: []gorm.Dialector{mysql.Open(slaveStr)},  // 从库，读操作
		Policy:   dbresolver.RandomPolicy{},               // 读库负载均衡策略
	}).SetConnMaxIdleTime(connMaxIdleTime).
		SetConnMaxLifetime(connMaxLifetime).
		SetMaxIdleConns(maxIdle).
		SetMaxOpenConns(maxOpen))
	if err != nil {
		panic(err)
		return
	}
	sqlDB, err := newDb.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(maxIdle)              // 设置最大空闲连接数
	sqlDB.SetMaxOpenConns(maxOpen)             // 设置最大连接数
	sqlDB.SetConnMaxLifetime(connMaxLifetime)  // 设置连接保持时间
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)  // 设置闲置保持时间
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
