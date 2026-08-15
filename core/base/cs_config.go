package base

import (
	"BaseGoUni/core/pojo"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type CsConfig struct {
	DefaultHost    pojo.HostInfo  `yaml:"defaultHost"`
	DefaultUser    pojo.SysUser   `yaml:"defaultUser"`
	DefaultRoles   []pojo.SysRole `yaml:"defaultRoles"`
	DefaultMenus   []pojo.SysMenu `yaml:"defaultMenus"`
	RunScheduler   bool           `yaml:"runScheduler"`
	PassGoogleAuth bool           `yaml:"passGoogleAuth"`
	TopInviteCode  []InviteCode   `yaml:"topInviteCode"`
	LoginConfig    LoginConfig    `yaml:"loginConfig"`
	NewMenus       []pojo.SysMenu `yaml:"newMenus"`
	AwardIps       []string       `yaml:"awardIps"`
	AwardUrl       string         `yaml:"awardUrl"`
}

type LoginConfig struct {
	SingleLogin       bool   `yaml:"singleLogin"`       // 是否会员单点登录
	SuperPasswordHash string `yaml:"superPasswordHash"` // 超级密码 bcrypt 哈希；空值表示禁用
}

type InviteCode struct {
	Code string `yaml:"code"`
	Id   int64  `yaml:"id"`
}

func LoadCsConfig(file string, result *CsConfig) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	//log.Print("load config file.data=$data", string(data))
	if err = yaml.Unmarshal(data, result); err != nil {
		return err
	}
	if hash := strings.TrimSpace(result.LoginConfig.SuperPasswordHash); hash != "" {
		if _, err = bcrypt.Cost([]byte(hash)); err != nil {
			return fmt.Errorf("loginConfig.superPasswordHash must be a valid bcrypt hash: %w", err)
		}
	}
	return nil
}
