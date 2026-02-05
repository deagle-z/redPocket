package base

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Salt   string `yaml:"salt"`
	PriKey string `yaml:"priKey"`
	Mysql  struct {
		DataSource string `yaml:"dataSource"`
		Master     string `yaml:"master"`
		Slave      string `yaml:"slave"`
	} `yaml:"mysql"`
	Redis    Redis         `yaml:"redis"`
	RabbitMq RabbitMq      `yaml:"rabbitMq"`
	AliOss   AliOssConfig  `yaml:"aliOss"`
	Telegram TelegramConfig `yaml:"telegram"`
	R2       R2Config      `yaml:"r2"`
}

type TelegramConfig struct {
	BotToken      string `yaml:"botToken"`      // Telegram Bot Token
	WebhookURL    string `yaml:"webhookUrl"`    // Webhook URL (可选)
	SafeMode      bool   `yaml:"safeMode"`      // 是否验证IP
	Enabled       bool   `yaml:"enabled"`       // 是否启用
	InviteBaseURL string `yaml:"inviteBaseUrl"` // 邀请链接前缀（如 https://t.me/xxxBot/?start=）
}

type AliOssConfig struct {
	BaseDownloadFolder string `yaml:"baseDownloadFolder"`
	DownloadLink       string `yaml:"downloadLink"`
	Bucket             string `yaml:"bucket"`
	Language           string `yaml:"language"`
	Endpoint           string `yaml:"endpoint"`
	AccessKeyId        string `yaml:"accessKeyId"`
	AccessKeySecret    string `yaml:"accessKeySecret"`
}

type R2Config struct {
	Endpoint       string `yaml:"endpoint"`       // R2 S3 endpoint
	Bucket         string `yaml:"bucket"`         // Bucket name
	AccessKeyId    string `yaml:"accessKeyId"`    // Access key ID
	AccessKeySecret string `yaml:"accessKeySecret"` // Access key secret
	PublicBaseURL  string `yaml:"publicBaseUrl"`  // Public base URL (optional)
}

type Redis struct {
	Host string `yaml:"host"`
	Pass string `yaml:"pass"`
	Db   int    `yaml:"db"`
}

type RabbitMq struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Exchange string `yaml:"exchange"`
	Routing  string `yaml:"routing"`
	Queue    string `yaml:"queue"`
}

func InitGlobalConfig(file string, GlobalConfig *Config) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	//log.Print("load config file.data=$data", string(data))
	return yaml.Unmarshal(data, &GlobalConfig)
}
