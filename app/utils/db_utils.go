package utils

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"os"
	"time"
)

func InitDb() (firstInit bool, err error) {
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
	noSchemataMaster := fmt.Sprintf(utils.GlobalConfig.Mysql.Master, "")
	utils.Db, err = gorm.Open(mysql.Open(noSchemataMaster), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("连接数据库错误 %s;noSchemataMaster=%s", err.Error(), noSchemataMaster)
		panic(err)
		return
	}
	err = utils.Db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		utils.CsConfig.DefaultHost.TablePrefix)).Error
	if err != nil {
		log.Printf("创建数据库错误 %s", err.Error())
		panic(err)
		return
	}
	masterStr := fmt.Sprintf(utils.GlobalConfig.Mysql.Master, utils.CsConfig.DefaultHost.TablePrefix)
	slaveStr := fmt.Sprintf(utils.GlobalConfig.Mysql.Slave, utils.CsConfig.DefaultHost.TablePrefix)
	err = utils.Db.Use(dbresolver.Register(dbresolver.Config{
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
	sqlDB, _ := utils.Db.DB()
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Second)
	firstInit, err = InitTables(utils.CsConfig.DefaultHost.TablePrefix)
	_ = utils.Db.AutoMigrate(&pojo.HostInfo{})
	return firstInit, nil
}

func InitTables(prefix string) (firstInit bool, err error) {
	db := utils.NewPrefixDb(prefix)
	if !db.Migrator().HasTable(&pojo.SysUser{}) {
		firstInit = true
		err = db.AutoMigrate(
			&pojo.SysUser{},
			&pojo.SysRole{},
			&pojo.SysMenu{},
			&pojo.ManageLog{},
			&pojo.LuckyMoney{},
			&pojo.LuckyHistory{},
			&pojo.AuthGroup{},
			&pojo.TgUser{},
		)
		if err != nil {
			panic(err)
		}
	}
	InitShardingHook(db)
	// 兼容历史数据：lucky_money_item.thunder 由可空升级为 not null 时，
	// 先将旧数据中的 NULL 回填为 0，避免 ALTER TABLE 报错 1138。
	if db.Migrator().HasTable(&pojo.LuckyMoneyItem{}) && db.Migrator().HasColumn(&pojo.LuckyMoneyItem{}, "thunder") {
		_ = db.Exec("UPDATE `lucky_money_item` SET `thunder` = 0 WHERE `thunder` IS NULL").Error
	}
	// 兼容历史数据：金额/手续费字段升级为 decimal(18,2) not null 时，先归一旧脏值，避免 1265。
	if db.Migrator().HasTable(&pojo.LuckyMoneyItem{}) {
		if db.Migrator().HasColumn(&pojo.LuckyMoneyItem{}, "thunder_amount") {
			_ = db.Exec("UPDATE `lucky_money_item` SET `thunder_amount` = 0").Error
		}
		if db.Migrator().HasColumn(&pojo.LuckyMoneyItem{}, "thunder_fee") {
			_ = db.Exec("UPDATE `lucky_money_item` SET `thunder_fee` = 0").Error
		}
		if db.Migrator().HasColumn(&pojo.LuckyMoneyItem{}, "win_fee") {
			_ = db.Exec("UPDATE `lucky_money_item` SET `win_fee` = 0").Error
		}
	}
	// 兼容增量字段/新表变更
	if err = db.AutoMigrate(&pojo.LuckyMoney{}, &pojo.LuckyMoneyItem{}); err != nil {
		panic(err)
	}
	if !db.Migrator().HasTable(pojo.AllCashHistoryShardingName) {
		err = InitShardingDataBase(db, pojo.CashHistory{}, pojo.CashHistoryTableName, pojo.CashHistoryShards)
		if err != nil {
			panic(fmt.Sprintf("Failed to init table: %v", err))
		}
		CreateView(uint(pojo.CashHistoryShards), pojo.AllCashHistoryShardingName, pojo.CashHistoryTableName)
		log.Printf("Init cash history success...\n")
	}
	return firstInit, nil
}
