package pojo

const TaskTypeRechargeFirstGiftInstallment = "recharge:first_gift_installment"

type RechargeFirstGiftInstallmentPayload struct {
	TablePrefix      string  `json:"tablePrefix"`
	OrderNo          string  `json:"orderNo"`
	UserId           int64   `json:"userId"`
	InstallmentIndex int     `json:"installmentIndex"`
	GiftAmount       float64 `json:"giftAmount"`
	TotalRate        float64 `json:"totalRate"`
	Ratio            int     `json:"ratio"`
	RatioBase        int     `json:"ratioBase"`
	TenantId         int64   `json:"tenantId"`
	SourceChannelId  *int64  `json:"sourceChannelId"`
}
