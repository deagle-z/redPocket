package pojo

type TgAuthLoginReq struct {
	ID                int64  `json:"id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Username          string `json:"username"`
	PhotoURL          string `json:"photo_url"`
	AuthDate          int64  `json:"auth_date"`
	Hash              string `json:"hash"`
	SourceChannelCode string `json:"sourceChannelCode"`
	ChannelCode       string `json:"channelCode"`
}

type TgAuthLoginBack struct {
	AccessToken string          `json:"accessToken"`
	UserType    int             `json:"userType"`
	ExpiresIn   int64           `json:"expiresIn"`
	TgUser      TgLoginUserBack `json:"tgUser"`
}

type TgLoginUserBack struct {
	ID        int64   `json:"id"`
	Uid       string  `json:"uid,omitempty"`
	TgID      int64   `json:"tgId,omitempty"`
	Username  *string `json:"username,omitempty"`
	TgName    *string `json:"tgName,omitempty"`
	FirstName *string `json:"firstName,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
}
