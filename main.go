package main

import (
	appservices "BaseGoUni/app/services"
	utils2 "BaseGoUni/app/utils"
	"BaseGoUni/core/base"
	"BaseGoUni/core/common"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"flag"
	"github.com/jinzhu/copier"
	"log"
	_ "net/http/pprof"
)

var coreFile = flag.String("cc", "core.yaml", "the config file")
var csFile = flag.String("cs", "cs.yaml", "the config file")
var syncFile = flag.String("sc", "sc.yaml", "the config file")

func main() {
	log.Print("Start run main...\n")
	flag.Parse()
	err := base.InitGlobalConfig(*coreFile, &utils.GlobalConfig)
	if err != nil {
		log.Printf("Init file read error.err=%v\n", err)
		return
	}
	log.Print("init global config success\n")
	err = base.LoadCsConfig(*csFile, &utils.CsConfig)
	if err != nil {
		log.Printf("Init cs file read error.err=%v\n", err)
		return
	}
	log.Print("init cs config success\n")
	err = base.LoadCsConfig(*syncFile, &utils.CsConfig)
	if err != nil {
		log.Printf("Init sc file read error.err=%v\n", err)
		return
	}
	log.Print("init sc config success\n")
	err = utils.InitRD()
	if err != nil {
		log.Printf("Init redis error.err=%v\n", err)
		return
	}
	log.Print("init redis success\n")
	firstInit, err := utils2.InitDb()
	if err != nil {
		log.Printf("Init mysql error.err=%v\n", err)
		return
	}
	log.Print("init tables success\n")
	var dbHostInfo pojo.HostInfo
	utils.Db.Where("host_name = ?", utils.CsConfig.DefaultHost.HostName).First(&dbHostInfo)
	if firstInit {
		if dbHostInfo.ID == 0 {
			_ = copier.Copy(&dbHostInfo, &utils.CsConfig.DefaultHost)
			dbHostInfo.Enabled = true
			dbHostInfo.ShowAdmin = true
			utils.Db.Create(&dbHostInfo)
			hostInfoStr, _ := json.Marshal(dbHostInfo)
			log.Printf("create host info = %s", string(hostInfoStr))
		}
		err = utils.InitConfig(dbHostInfo)
		if err != nil {
			log.Printf("Init config error.err=%v\n", err)
			return
		}
	}
	err = utils.InitMenus(dbHostInfo)
	if err != nil {
		log.Printf("Init menu error.err=%v\n", err)
		return
	}
	//go utils2.InitRabbitMq()
	log.Print("init mq success\n")
	if utils.CsConfig.RunScheduler {
		common.InitScheduler()
		log.Print("init scheduler success\n")
	}
	go utils.Test()
	utils.InitI18n()
	
	// 初始化 Telegram Bot（如果配置了）
	if utils.GlobalConfig.Telegram.Enabled && utils.GlobalConfig.Telegram.BotToken != "" {
		db := utils.NewPrefixDb(dbHostInfo.TablePrefix)
		err = appservices.InitTelegramBot(db, dbHostInfo.TablePrefix, utils.GlobalConfig.Telegram.BotToken)
		if err != nil {
			log.Printf("Init telegram bot error.err=%v\n", err)
		} else {
			log.Print("init telegram bot success\n")
		}
	}
	
	common.InitGin()
}
