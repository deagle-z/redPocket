package pojo

import (
	"encoding/json"
	"strconv"
	"strings"
)

type GSCFlexibleString string

func (v *GSCFlexibleString) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || strings.EqualFold(raw, "null") {
		*v = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*v = GSCFlexibleString(strings.TrimSpace(text))
		return nil
	}
	*v = GSCFlexibleString(raw)
	return nil
}

func (v GSCFlexibleString) String() string {
	return strings.TrimSpace(string(v))
}

type GSCFlexibleInt int

func (v *GSCFlexibleInt) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	raw = strings.Trim(raw, `"`)
	if raw == "" || strings.EqualFold(raw, "null") {
		*v = 0
		return nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}
	*v = GSCFlexibleInt(n)
	return nil
}

func (v GSCFlexibleInt) Int() int {
	return int(v)
}

type GSCFlexibleInt64 int64

func (v *GSCFlexibleInt64) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	raw = strings.Trim(raw, `"`)
	if raw == "" || strings.EqualFold(raw, "null") {
		*v = 0
		return nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*v = GSCFlexibleInt64(n)
	return nil
}

func (v GSCFlexibleInt64) Int64() int64 {
	return int64(v)
}

type GSCFlexibleFloat float64

func (v *GSCFlexibleFloat) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	raw = strings.Trim(raw, `"`)
	if raw == "" || strings.EqualFold(raw, "null") {
		*v = 0
		return nil
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return err
	}
	*v = GSCFlexibleFloat(n)
	return nil
}

func (v GSCFlexibleFloat) Float64() float64 {
	return float64(v)
}

type GSCOpenBalanceReq struct {
	BatchRequests []GSCOpenBalanceBatchRequest `json:"batch_requests"`
	OperatorCode  string                       `json:"operator_code"`
	Currency      string                       `json:"currency"`
	Sign          string                       `json:"sign"`
	RequestTime   GSCFlexibleString            `json:"request_time"`
}

type GSCOpenBalanceBatchRequest struct {
	MemberAccount string         `json:"member_account"`
	ProductCode   GSCFlexibleInt `json:"product_code"`
}

type GSCOpenTransferReq struct {
	BatchRequests []GSCOpenTransferBatchRequest `json:"batch_requests"`
	OperatorCode  string                        `json:"operator_code"`
	GameType      string                        `json:"game_type"`
	Currency      string                        `json:"currency"`
	Sign          string                        `json:"sign"`
	RequestTime   GSCFlexibleString             `json:"request_time"`
}

type GSCOpenTransferBatchRequest struct {
	MemberAccount string                     `json:"member_account"`
	ProductCode   GSCFlexibleInt             `json:"product_code"`
	GameType      string                     `json:"game_type"`
	Transactions  []GSCOpenWalletTransaction `json:"transactions"`
}

type GSCOpenWalletTransaction struct {
	ID             string           `json:"id"`
	Action         string           `json:"action"`
	WagerCode      string           `json:"wager_code"`
	WagerStatus    string           `json:"wager_status"`
	Amount         GSCFlexibleFloat `json:"amount"`
	BetAmount      GSCFlexibleFloat `json:"bet_amount"`
	ValidBetAmount GSCFlexibleFloat `json:"valid_bet_amount"`
	PrizeAmount    GSCFlexibleFloat `json:"prize_amount"`
	TipAmount      GSCFlexibleFloat `json:"tip_amount"`
	SettledAt      GSCFlexibleInt64 `json:"settled_at"`
	GameCode       string           `json:"game_code"`
	RoundID        string           `json:"round_id"`
	ChannelCode    string           `json:"channel_code"`
	WagerType      string           `json:"wager_type"`
}

type GSCOpenPushBetDataReq struct {
	OperatorCode string            `json:"operator_code"`
	Wagers       []GSCOpenWager    `json:"wagers"`
	Sign         string            `json:"sign"`
	RequestTime  GSCFlexibleString `json:"request_time"`
}

type GSCOpenWager struct {
	MemberAccount  string            `json:"member_account"`
	BetAmount      GSCFlexibleFloat  `json:"bet_amount"`
	ValidBetAmount GSCFlexibleFloat  `json:"valid_bet_amount"`
	PrizeAmount    GSCFlexibleFloat  `json:"prize_amount"`
	TipAmount      GSCFlexibleFloat  `json:"tip_amount"`
	WagerType      string            `json:"wager_type"`
	WagerCode      string            `json:"wager_code"`
	WagerStatus    string            `json:"wager_status"`
	RoundID        string            `json:"round_id"`
	ChannelCode    string            `json:"channel_code"`
	GameType       string            `json:"game_type"`
	SettledAt      GSCFlexibleInt64  `json:"settled_at"`
	CreatedAt      GSCFlexibleInt64  `json:"created_at"`
	Payload        json.RawMessage   `json:"payload"`
	ProductCode    GSCFlexibleString `json:"product_code"`
	GameCode       string            `json:"game_code"`
	Currency       string            `json:"currency"`
}

type GSCOpenBalanceResp struct {
	Data []GSCOpenBalanceData `json:"data"`
}

type GSCOpenBalanceData struct {
	MemberAccount string  `json:"member_account"`
	ProductCode   int     `json:"product_code"`
	Balance       float64 `json:"balance"`
	Code          int     `json:"code"`
	Message       string  `json:"message"`
}

type GSCOpenTransferResp struct {
	Data []GSCOpenTransferData `json:"data"`
}

type GSCOpenTransferData struct {
	MemberAccount string  `json:"member_account"`
	ProductCode   int     `json:"product_code"`
	BeforeBalance float64 `json:"before_balance"`
	Balance       float64 `json:"balance"`
	Code          int     `json:"code"`
	Message       string  `json:"message"`
}

type GSCOpenBaseResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
