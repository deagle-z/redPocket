package pojo

type GameCashGetReq struct {
	UserID string `json:"userid"`
}

type GameCashTransferInOutReq struct {
	UserID  string  `json:"userid"`
	TID     string  `json:"tid"`
	Amount  float64 `json:"amount"`
	RoundID string  `json:"roundid"`
	GameID  string  `json:"gameid"`
	ReqTime int64   `json:"req_time"`
	Reason  string  `json:"reason"`
	IsEnd   bool    `json:"is_end"`
	IsBuy   bool    `json:"is_buy"`
	Bet     float64 `json:"bet"`
}

type GameCashGetData struct {
	Balance float64 `json:"balance"`
}
