package pojo

type TgAuthLoginReq struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

type TgAuthLoginBack struct {
	AccessToken string     `json:"accessToken"`
	UserType    int        `json:"userType"`
	ExpiresIn   int64      `json:"expiresIn"`
	TgUser      TgUserBack `json:"tgUser"`
}
