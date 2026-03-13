package pojo

import "time"

type SysCustomField struct {
	BaseModel
	FieldKey         string  `json:"fieldKey" gorm:"column:field_key;type:varchar(64);uniqueIndex;not null;comment:字段key，唯一，如 account_name / bank_name / upi_id"`
	FieldLabel       string  `json:"fieldLabel" gorm:"column:field_label;type:varchar(128);not null;comment:字段名称"`
	FieldPlaceholder *string `json:"fieldPlaceholder" gorm:"column:field_placeholder;type:varchar(255);comment:占位提示"`
	FieldType        string  `json:"fieldType" gorm:"column:field_type;type:varchar(32);not null;comment:字段类型 input/select/textarea/number/date/file/switch/radio/checkbox"`
	DataType         string  `json:"dataType" gorm:"column:data_type;type:varchar(32);not null;default:'string';comment:数据类型 string/int/decimal/bool/json/date"`
	DefaultValue     *string `json:"defaultValue" gorm:"column:default_value;type:varchar(255);comment:默认值"`
	IsRequired       int8    `json:"isRequired" gorm:"column:is_required;type:tinyint;not null;default:0;comment:是否必填 1是 0否"`
	IsSensitive      int8    `json:"isSensitive" gorm:"column:is_sensitive;type:tinyint;not null;default:0;comment:是否敏感字段 1是 0否"`
	MaxLength        *int    `json:"maxLength" gorm:"column:max_length;type:int;comment:最大长度"`
	MinLength        *int    `json:"minLength" gorm:"column:min_length;type:int;comment:最小长度"`
	RegexRule        *string `json:"regexRule" gorm:"column:regex_rule;type:varchar(255);comment:正则校验规则"`
	ErrorTips        *string `json:"errorTips" gorm:"column:error_tips;type:varchar(255);comment:校验失败提示"`
	OptionsJSON      *string `json:"optionsJson" gorm:"column:options_json;type:json;comment:下拉/单选/复选项JSON"`
	Status           int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index;comment:状态 1启用 0禁用"`
	Remark           *string `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
}

type SysCustomFieldSearch struct {
	PageInfo
	FieldKey   string `json:"fieldKey"`
	FieldLabel string `json:"fieldLabel"`
	FieldType  string `json:"fieldType"`
	DataType   string `json:"dataType"`
	Status     *int8  `json:"status"`
}

type SysCustomFieldSet struct {
	ID               int64   `json:"id"`
	FieldKey         string  `json:"fieldKey"`
	FieldLabel       string  `json:"fieldLabel"`
	FieldPlaceholder *string `json:"fieldPlaceholder"`
	FieldType        string  `json:"fieldType"`
	DataType         string  `json:"dataType"`
	DefaultValue     *string `json:"defaultValue"`
	IsRequired       int8    `json:"isRequired"`
	IsSensitive      int8    `json:"isSensitive"`
	MaxLength        *int    `json:"maxLength"`
	MinLength        *int    `json:"minLength"`
	RegexRule        *string `json:"regexRule"`
	ErrorTips        *string `json:"errorTips"`
	OptionsJSON      *string `json:"optionsJson"`
	Status           int8    `json:"status"`
	Remark           *string `json:"remark"`
}

type SysCustomFieldBack struct {
	ID               int64     `json:"id"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	FieldKey         string    `json:"fieldKey"`
	FieldLabel       string    `json:"fieldLabel"`
	FieldPlaceholder *string   `json:"fieldPlaceholder"`
	FieldType        string    `json:"fieldType"`
	DataType         string    `json:"dataType"`
	DefaultValue     *string   `json:"defaultValue"`
	IsRequired       int8      `json:"isRequired"`
	IsSensitive      int8      `json:"isSensitive"`
	MaxLength        *int      `json:"maxLength"`
	MinLength        *int      `json:"minLength"`
	RegexRule        *string   `json:"regexRule"`
	ErrorTips        *string   `json:"errorTips"`
	OptionsJSON      *string   `json:"optionsJson"`
	Status           int8      `json:"status"`
	Remark           *string   `json:"remark"`
}

type SysCustomFieldResp struct {
	BasePageResponse[SysCustomFieldBack]
}

var SysCustomFieldTableName = "sys_custom_field"

func (SysCustomField) TableName() string {
	return SysCustomFieldTableName
}
