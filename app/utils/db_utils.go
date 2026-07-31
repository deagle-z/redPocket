package utils

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"os"
	"strings"
	"time"
)

const cryptoRechargeSchemaMigrationKey = "crypto_recharge_hardening"
const cryptoRechargeSchemaVersion = 1

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
	log.Print("init mysql: register prefix db sharding hook...\n")
	utils.RegisterPrefixDbInitializer("cash_history_sharding", InitShardingHook)
	utils.RegisterPrefixDbInitializer("crypto_recharge_schema", func(prefixDB *gorm.DB) {
		if shouldSkipAutoMigrate() {
			if verifyErr := verifyCryptoRechargeSchema(prefixDB); verifyErr != nil {
				panic(verifyErr)
			}
		} else {
			if migrateErr := ensureUsdtRechargeSchema(prefixDB); migrateErr != nil {
				panic(migrateErr)
			}
		}
	})
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
			&pojo.TgUserWithdrawFlowBatch{},
			&pojo.TgUserWithdrawFlowEvent{},
			&pojo.TgUserWithdrawFlowAllocation{},
			&pojo.TgUserWithdrawFlowOutbox{},
			&pojo.TgUserCheckInRecord{},
			&pojo.TgUserInviteRewardLog{},
			&pojo.ExchangeCode{},
			&pojo.ExchangeCodeRedeem{},
			&pojo.RechargeOrder{},
			&pojo.WithdrawOrderBr{},
			&pojo.TgUserRebateRecord{},
			&pojo.TgTaskActivityConfig{},
			&pojo.TgTaskActivityRecord{},
			&pojo.PlatformProfitLedger{},
			&pojo.SysTenantPrizePool{},
			&pojo.SysTenantPrizePoolRecord{},
			&pojo.SysTenantPrizePoolConfig{},
			&pojo.UserLotteryRecord{},
			&pojo.AppGamePlatform{},
			&pojo.AppGame{},
			&pojo.AppUserBetRecord{},
		)
		if err != nil {
			panic(err)
		}
		log.Print("init tables: first init automigrate done\n")
	}
	log.Print("init tables: ensure exchange_code tables...\n")
	if err = db.AutoMigrate(&pojo.ExchangeCode{}, &pojo.ExchangeCodeRedeem{}); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_user device fingerprint schema...\n")
	if err = ensureTgUserDeviceFingerprintSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_user invite valid schema...\n")
	if err = ensureTgUserInviteValidSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_user rebate withdraw schema...\n")
	if err = ensureTgUserRebateWithdrawSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_user sport balance schema...\n")
	if err = ensureTgUserSportBalanceSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure sys_tenant withdraw auto review schema...\n")
	if err = ensureSysTenantWithdrawAutoReviewSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_user withdraw flow outbox schema...\n")
	if err = ensureTgUserWithdrawFlowOutboxSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_user invite reward log schema...\n")
	if err = ensureTgUserInviteRewardLogSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_task_activity_config schema...\n")
	if err = ensureTgTaskActivityConfigSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure tg_task_activity_record schema...\n")
	if err = ensureTgTaskActivityRecordSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure withdraw_order_br performance schema...\n")
	if err = ensureWithdrawOrderBrPerformanceSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure usdt recharge schema...\n")
	if err = ensureUsdtRechargeSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure recharge_order wallet schema...\n")
	if err = ensureRechargeOrderWalletSchema(db); err != nil {
		panic(err)
	}
	log.Print("init tables: ensure trial lucky item pick index...\n")
	if err = ensureTrialLuckyMoneyItemPickIndex(db); err != nil {
		panic(err)
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

func ensureTgUserDeviceFingerprintSchema(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasColumn(&pojo.TgUser{}, "DeviceFingerprint") {
		if err := migrator.AddColumn(&pojo.TgUser{}, "DeviceFingerprint"); err != nil {
			return err
		}
	}
	if !migrator.HasIndex(&pojo.TgUser{}, "idx_tg_user_device_fingerprint") {
		if err := migrator.CreateIndex(&pojo.TgUser{}, "idx_tg_user_device_fingerprint"); err != nil {
			return err
		}
	}
	return nil
}

func ensureTgUserInviteValidSchema(db *gorm.DB) error {
	migrator := db.Migrator()
	for _, column := range []string{
		"InviteValidFlag",
		"InviteValidAt",
		"InviteValidRechargeAmount",
		"InviteValidBetAmount",
	} {
		if !migrator.HasColumn(&pojo.TgUser{}, column) {
			if err := migrator.AddColumn(&pojo.TgUser{}, column); err != nil {
				return err
			}
		}
	}
	if !migrator.HasIndex(&pojo.TgUser{}, "idx_tg_user_invite_valid") {
		if err := migrator.CreateIndex(&pojo.TgUser{}, "idx_tg_user_invite_valid"); err != nil {
			return err
		}
	}
	return nil
}

func ensureTgUserRebateWithdrawSchema(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasColumn(&pojo.TgUser{}, "RebateWithdrawDisabled") {
		if err := migrator.AddColumn(&pojo.TgUser{}, "RebateWithdrawDisabled"); err != nil {
			return err
		}
	}
	return nil
}

func ensureTgUserSportBalanceSchema(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasColumn(&pojo.TgUser{}, "SportBalance") {
		if err := migrator.AddColumn(&pojo.TgUser{}, "SportBalance"); err != nil {
			return err
		}
	}
	return nil
}

func ensureSysTenantWithdrawAutoReviewSchema(db *gorm.DB) error {
	migrator := db.Migrator()
	for _, column := range []string{
		"WithdrawAutoReviewEnabled",
		"WithdrawAutoReviewMaxAmount",
		"WithdrawAutoReviewDailyLimit",
	} {
		if !migrator.HasColumn(&pojo.SysTenant{}, column) {
			if err := migrator.AddColumn(&pojo.SysTenant{}, column); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureTgUserWithdrawFlowOutboxSchema(db *gorm.DB) error {
	return db.AutoMigrate(&pojo.TgUserWithdrawFlowOutbox{})
}

func ensureTgUserInviteRewardLogSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(&pojo.TgUserInviteRewardLog{}); err != nil {
		return err
	}
	migrator := db.Migrator()
	if migrator.HasIndex(&pojo.TgUserInviteRewardLog{}, "idx_invite_reward_user_sub") {
		if err := migrator.DropIndex(&pojo.TgUserInviteRewardLog{}, "idx_invite_reward_user_sub"); err != nil {
			return err
		}
	}
	if !migrator.HasIndex(&pojo.TgUserInviteRewardLog{}, "idx_invite_reward_user_sub_stage") {
		if err := migrator.CreateIndex(&pojo.TgUserInviteRewardLog{}, "idx_invite_reward_user_sub_stage"); err != nil {
			return err
		}
	}
	return nil
}

func ensureTgTaskActivityConfigSchema(db *gorm.DB) error {
	return db.AutoMigrate(&pojo.TgTaskActivityConfig{})
}

func ensureTgTaskActivityRecordSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(&pojo.TgTaskActivityRecord{}); err != nil {
		return err
	}
	migrator := db.Migrator()
	if migrator.HasIndex(&pojo.TgTaskActivityRecord{}, "uk_task_activity_user_config") {
		if err := migrator.DropIndex(&pojo.TgTaskActivityRecord{}, "uk_task_activity_user_config"); err != nil {
			return err
		}
	}
	if !migrator.HasIndex(&pojo.TgTaskActivityRecord{}, "idx_task_activity_user_config") {
		if err := migrator.CreateIndex(&pojo.TgTaskActivityRecord{}, "idx_task_activity_user_config"); err != nil {
			return err
		}
	}
	return nil
}

func ensureTrialLuckyMoneyItemPickIndex(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasTable(&pojo.TrialLuckyMoneyItem{}) {
		return nil
	}
	if !migrator.HasIndex(&pojo.TrialLuckyMoneyItem{}, "idx_trial_lucky_item_pick") {
		if err := migrator.CreateIndex(&pojo.TrialLuckyMoneyItem{}, "idx_trial_lucky_item_pick"); err != nil {
			return err
		}
	}
	return nil
}

func ensureWithdrawOrderBrPerformanceSchema(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasTable(&pojo.WithdrawOrderBr{}) {
		return nil
	}
	if !migrator.HasColumn(&pojo.WithdrawOrderBr{}, "WithdrawSource") {
		if err := migrator.AddColumn(&pojo.WithdrawOrderBr{}, "WithdrawSource"); err != nil {
			return err
		}
	}
	if !migrator.HasColumn(&pojo.WithdrawOrderBr{}, "AutoReviewed") {
		if err := migrator.AddColumn(&pojo.WithdrawOrderBr{}, "AutoReviewed"); err != nil {
			return err
		}
	}
	if err := backfillWithdrawOrderBrSource(db); err != nil {
		return err
	}
	if !migrator.HasIndex(&pojo.WithdrawOrderBr{}, "idx_withdraw_order_user_created") {
		if err := db.Exec("CREATE INDEX `idx_withdraw_order_user_created` ON `withdraw_order_br` (`user_id`, `created_at`)").Error; err != nil {
			return err
		}
	}
	if !migrator.HasIndex(&pojo.WithdrawOrderBr{}, "idx_withdraw_order_auto_reviewed") {
		if err := db.Exec("CREATE INDEX `idx_withdraw_order_auto_reviewed` ON `withdraw_order_br` (`auto_reviewed`)").Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureUsdtRechargeSchema(db *gorm.DB) error {
	return db.Clauses(dbresolver.Write).Connection(func(connection *gorm.DB) error {
		var databaseName string
		if err := connection.Raw("SELECT DATABASE()").Scan(&databaseName).Error; err != nil {
			return err
		}
		lockName := "bgu_crypto_schema_" + databaseName
		if len(lockName) > 64 {
			lockName = lockName[:64]
		}
		var acquired int
		if err := connection.Raw("SELECT GET_LOCK(?, 60)", lockName).Scan(&acquired).Error; err != nil {
			return err
		}
		if acquired != 1 {
			return fmt.Errorf("crypto schema migration lock unavailable for database %q", databaseName)
		}
		defer func() {
			var released int
			_ = connection.Raw("SELECT RELEASE_LOCK(?)", lockName).Scan(&released).Error
		}()
		return ensureUsdtRechargeSchemaLocked(connection)
	})
}

func ensureUsdtRechargeSchemaLocked(db *gorm.DB) error {
	if cryptoRechargeMigrationApplied(db) && verifyCryptoRechargeSchema(db) == nil {
		return nil
	}
	if err := db.AutoMigrate(
		&pojo.UsdtRechargeOrder{},
		&pojo.UsdtRechargeTx{},
		&pojo.CryptoRechargeScanCursor{},
		&pojo.CryptoRechargeAmountReservation{},
		&pojo.CryptoRechargeManualSupplement{},
		&pojo.CryptoRechargeManualProof{},
		&pojo.CryptoRechargeSchemaVersion{},
	); err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE usdt_recharge_tx
		SET review_status = ?, exception_code = ?
		WHERE matched_order_no IS NULL
			AND review_status = ?
			AND COALESCE(amount_atomic, '') = ''
			AND amount_micro > 0
	`, pojo.CryptoRechargeTxReviewManualRequired, pojo.CryptoRechargeExceptionLegacy, pojo.CryptoRechargeTxReviewNew).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE usdt_recharge_order
		SET amount_decimals = 6,
			base_amount_atomic = CAST(base_amount_micro AS CHAR),
			expected_amount_atomic = CAST(expected_amount_micro AS CHAR)
		WHERE network = 'TRC20'
			AND token = 'USDT'
			AND COALESCE(expected_amount_atomic, '') = ''
			AND expected_amount_micro > 0
	`).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE usdt_recharge_tx
		SET amount_decimals = 6,
			amount_atomic = CAST(amount_micro AS CHAR)
		WHERE network = 'TRC20'
			AND token = 'USDT'
			AND COALESCE(amount_atomic, '') = ''
			AND amount_micro > 0
	`).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		INSERT IGNORE INTO crypto_recharge_amount_reservation
			(reservation_key, network, token, receive_address, expected_amount_atomic, order_no, created_at, updated_at)
		SELECT
			SHA2(CONCAT_WS('|', network, token, receive_address, expected_amount_atomic), 256),
			network, token, receive_address, expected_amount_atomic, order_no, NOW(3), NOW(3)
		FROM usdt_recharge_order
		WHERE COALESCE(expected_amount_atomic, '') <> ''
	`).Error; err != nil {
		return err
	}
	marker := pojo.CryptoRechargeSchemaVersion{
		MigrationKey: cryptoRechargeSchemaMigrationKey,
		Version:      cryptoRechargeSchemaVersion,
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "migration_key"}},
		DoUpdates: clause.Assignments(map[string]any{"version": cryptoRechargeSchemaVersion}),
	}).Create(&marker).Error; err != nil {
		return err
	}
	return verifyCryptoRechargeSchema(db)
}

func cryptoRechargeMigrationApplied(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	db = db.Clauses(dbresolver.Write)
	if !db.Migrator().HasTable(&pojo.CryptoRechargeSchemaVersion{}) {
		return false
	}
	var marker pojo.CryptoRechargeSchemaVersion
	if err := db.Where("migration_key = ?", cryptoRechargeSchemaMigrationKey).First(&marker).Error; err != nil {
		return false
	}
	return marker.Version >= cryptoRechargeSchemaVersion
}

func verifyCryptoRechargeSchema(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("crypto recharge schema database is unavailable")
	}
	db = db.Clauses(dbresolver.Write)
	requiredColumns := []struct {
		model   any
		columns []string
	}{
		{&pojo.UsdtRechargeOrder{}, []string{"network", "token", "receive_address", "amount_decimals", "base_amount_atomic", "expected_amount_atomic", "paid_amount_atomic"}},
		{&pojo.UsdtRechargeTx{}, []string{"tx_id", "amount_decimals", "amount_atomic", "block_height", "block_hash", "confirmations", "review_status", "exception_code"}},
		{&pojo.CryptoRechargeScanCursor{}, []string{"cursor_key", "last_block_height", "last_block_hash", "last_block_timestamp", "pagination_cursor", "target_block_height", "target_block_timestamp", "contract_address"}},
		{&pojo.CryptoRechargeAmountReservation{}, []string{"reservation_key", "expected_amount_atomic", "order_no"}},
		{&pojo.CryptoRechargeManualSupplement{}, []string{"order_no", "recharge_order_id", "operator_id", "reason", "status"}},
		{&pojo.CryptoRechargeManualProof{}, []string{"supplement_id", "source_network", "source_token", "tx_id", "amount_decimals", "amount_atomic"}},
		{&pojo.CryptoRechargeSchemaVersion{}, []string{"migration_key", "version"}},
	}
	for _, required := range requiredColumns {
		if !db.Migrator().HasTable(required.model) {
			return fmt.Errorf("crypto recharge schema missing table for %T; apply the production migration before starting with BGU_SKIP_AUTO_MIGRATE=1", required.model)
		}
		for _, column := range required.columns {
			if !db.Migrator().HasColumn(required.model, column) {
				return fmt.Errorf("crypto recharge schema missing column %s for %T; apply the production migration before starting with BGU_SKIP_AUTO_MIGRATE=1", column, required.model)
			}
		}
	}
	uniqueColumns := []struct {
		table  string
		column string
	}{
		{pojo.UsdtRechargeOrderTableName, "order_no"},
		{pojo.UsdtRechargeTxTableName, "tx_id"},
		{pojo.CryptoRechargeScanCursorTableName, "cursor_key"},
		{pojo.CryptoRechargeAmountReservationTableName, "reservation_key"},
		{pojo.CryptoRechargeAmountReservationTableName, "order_no"},
		{pojo.CryptoRechargeManualSupplementTableName, "order_no"},
		{pojo.CryptoRechargeManualProofTableName, "tx_id"},
		{pojo.CryptoRechargeSchemaVersionTableName, "migration_key"},
	}
	for _, required := range uniqueColumns {
		var count int64
		if err := db.Raw(`
			SELECT COUNT(*)
			FROM (
				SELECT index_name
				FROM information_schema.statistics
				WHERE table_schema = DATABASE()
					AND table_name = ?
					AND non_unique = 0
				GROUP BY index_name
				HAVING COUNT(*) = 1
					AND MAX(column_name) = ?
			) AS single_column_unique_indexes
		`, required.table, required.column).Scan(&count).Error; err != nil {
			return err
		}
		if count <= 0 {
			return fmt.Errorf("crypto recharge schema missing unique index on %s.%s; apply the production migration before starting with BGU_SKIP_AUTO_MIGRATE=1", required.table, required.column)
		}
	}
	if !cryptoRechargeMigrationApplied(db) {
		return fmt.Errorf("crypto recharge schema migration version %d is not applied; run the controlled production migration before starting with BGU_SKIP_AUTO_MIGRATE=1", cryptoRechargeSchemaVersion)
	}
	return nil
}

func ensureRechargeOrderWalletSchema(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasColumn(&pojo.RechargeOrder{}, "WalletType") {
		if err := migrator.AddColumn(&pojo.RechargeOrder{}, "WalletType"); err != nil {
			return err
		}
	}
	return db.Model(&pojo.RechargeOrder{}).
		Where("COALESCE(wallet_type, '') = ''").
		Update("wallet_type", pojo.RechargeWalletTypeBalance).Error
}

func backfillWithdrawOrderBrSource(db *gorm.DB) error {
	const (
		balanceSource = "balance"
		rebateSource  = "rebate"
	)
	if err := db.Exec(`
UPDATE withdraw_order_br
SET withdraw_source = ?
WHERE COALESCE(withdraw_source, '') = ''
  AND (
    COALESCE(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.source')), '') = ?
    OR COALESCE(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.balanceSource')), '') = ?
    OR COALESCE(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.withdrawSource')), '') = ?
  )`,
		rebateSource,
		rebateSource,
		rebateSource,
		rebateSource,
	).Error; err != nil {
		return err
	}
	return db.Exec(`
UPDATE withdraw_order_br
SET withdraw_source = ?
WHERE COALESCE(withdraw_source, '') = ''`,
		balanceSource,
	).Error
}
