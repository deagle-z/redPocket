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
	"strings"
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
	log.Print("init mysql: open base connection...\n")
	utils.Db, err = gorm.Open(mysql.Open(noSchemataMaster), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("连接数据库错误 %s;noSchemataMaster=%s", err.Error(), noSchemataMaster)
		panic(err)
		return
	}
	log.Print("init mysql: base connection opened\n")
	log.Print("init mysql: create database if not exists...\n")
	err = utils.Db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		utils.CsConfig.DefaultHost.TablePrefix)).Error
	if err != nil {
		log.Printf("创建数据库错误 %s", err.Error())
		panic(err)
		return
	}
	log.Print("init mysql: database ready\n")
	masterStr := fmt.Sprintf(utils.GlobalConfig.Mysql.Master, utils.CsConfig.DefaultHost.TablePrefix)
	slaveStr := fmt.Sprintf(utils.GlobalConfig.Mysql.Slave, utils.CsConfig.DefaultHost.TablePrefix)
	log.Print("init mysql: register resolver...\n")
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
	log.Print("init mysql: init tables...\n")
	firstInit, err = InitTables(utils.CsConfig.DefaultHost.TablePrefix)
	if shouldSkipAutoMigrate() {
		log.Print("init mysql: skip host_info automigrate by BGU_SKIP_AUTO_MIGRATE\n")
	} else {
		log.Print("init mysql: automigrate host_info...\n")
		_ = utils.Db.AutoMigrate(&pojo.HostInfo{})
	}
	log.Print("init mysql: done\n")
	return firstInit, nil
}

func InitTables(prefix string) (firstInit bool, err error) {
	db := utils.NewPrefixDb(prefix)
	if shouldSkipAutoMigrate() {
		log.Print("init tables: skip automigrate by BGU_SKIP_AUTO_MIGRATE\n")
		InitShardingHook(db)
		return false, nil
	}
	log.Print("init tables: check first init...\n")
	if !db.Migrator().HasTable(&pojo.SysUser{}) {
		firstInit = true
		log.Print("init tables: first init automigrate...\n")
		err = db.AutoMigrate(
			&pojo.SysUser{},
			&pojo.SysRole{},
			&pojo.SysMenu{},
			&pojo.ManageLog{},
			&pojo.SysSourceChannel{},
			&pojo.SysBanner{},
			&pojo.SysBannerI18n{},
			&pojo.SysBannerCountryRel{},
			&pojo.SysTenant{},
			&pojo.SysTenantUser{},
			&pojo.AttributionEvent{},
			&pojo.LuckyMoney{},
			&pojo.LuckyHistory{},
			&pojo.LuckyMoneyItem{},
			&pojo.TrialLuckyMoney{},
			&pojo.TrialLuckyMoneyItem{},
			&pojo.TrialLuckyHistory{},
			&pojo.TrialLuckyFlowLotteryReward{},
			&pojo.TrialBotUser{},
			&pojo.TrialCashHistory{},
			&pojo.AuthGroup{},
			&pojo.TgUser{},
			&pojo.TgChannelMember{},
			&pojo.TgUserWithdrawLimitState{},
			&pojo.TgUserWithdrawActivityCycle{},
			&pojo.TgUserCheckInRecord{},
			&pojo.RechargeOrder{},
			&pojo.WithdrawOrderBr{},
			&pojo.TgUserRebateRecord{},
			&pojo.PlatformProfitLedger{},
			&pojo.SysTenantPrizePool{},
			&pojo.SysTenantPrizePoolRecord{},
			&pojo.SysTenantPrizePoolConfig{},
			&pojo.UserLotteryRecord{},
		)
		if err != nil {
			panic(err)
		}
		log.Print("init tables: first init automigrate done\n")
	}
	log.Print("init tables: init sharding hook...\n")
	InitShardingHook(db)
	if !db.Migrator().HasTable(pojo.CashHistoryTableName + "_0") {
		log.Print("init tables: init cash_history sharding tables...\n")
		err = InitShardingDataBase(db, pojo.CashHistory{}, pojo.CashHistoryTableName, pojo.CashHistoryShards)
		if err != nil {
			panic(fmt.Sprintf("Failed to init table: %v", err))
		}
		log.Print("init tables: init cash_history sharding tables done\n")
	} else {
		log.Print("init tables: cash_history sharding tables exist, skip init\n")
	}
	if !db.Migrator().HasTable(pojo.AllCashHistoryShardingName) {
		log.Print("init tables: create all_cash_history view...\n")
		CreateView(uint(pojo.CashHistoryShards), pojo.AllCashHistoryShardingName, pojo.CashHistoryTableName)
	} else {
		log.Print("init tables: all_cash_history view exists, skip create\n")
	}
	log.Print("init tables: done\n")
	return firstInit, nil
}

func shouldSkipAutoMigrate() bool {
	value := strings.TrimSpace(os.Getenv("BGU_SKIP_AUTO_MIGRATE"))
	return value == "1" || strings.EqualFold(value, "true")
}
