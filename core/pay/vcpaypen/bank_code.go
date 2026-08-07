package vcpaypen

import "strings"

// payoutBankCodeMap 把前端/后台使用的银行名称映射为 VcPay 秘鲁原生银行编码。
// 原生编码见 readme.md「代付 BankCode」。电子钱包 Yape / Plin 也走同一张表。
var payoutBankCodeMap = map[string]string{
	// 电子钱包
	"YAPE": "PEW001",
	"PLIN": "PEW002",

	// 银行 / 储蓄合作社
	"AZTECA":                    "PEBA01",
	"BANCO_AZTECA":              "PEBA01",
	"BANCO_AZTECA_PERU":         "PEBA01",
	"BANBIF":                    "PEBB01",
	"BANCO_DE_LA_NACION":        "PEBB02",
	"BN":                        "PEBB02",
	"BBVA":                      "PEBB03",
	"BBVA_PERU":                 "PEBB03",
	"BBVA_CONTINENTAL":          "PEBB03",
	"BCP":                       "PEBB04",
	"BANCO_DE_CREDITO":          "PEBB04",
	"BANCO_DE_CREDITO_DEL_PERU": "PEBB04",
	"BANCO_DE_COMERCIO":         "PEBB05",
	"CAJA_AREQUIPA":             "PEBC01",
	"CAJA_CUSCO":                "PEBC02",
	"CAJA_HUANCAYO":             "PEBC03",
	"CAJA_MAYNAS":               "PEBC04",
	"CAJA_METROPOLITANA":        "PEBC05",
	"CAJA_MUNICIPAL_ICA":        "PEBC06",
	"CAJA_ICA":                  "PEBC06",
	"CAJA_PIURA":                "PEBC07",
	"CAJA_SULLANA":              "PEBC08",
	"CAJA_TACNA":                "PEBC09",
	"CAJA_TRUJILLO":             "PEBC10",
	"CENCOSUD":                  "PEBC11",
	"BANCO_CENCOSUD":            "PEBC11",
	"CITIBANK":                  "PEBC12",
	"CITIBANK_PERU":             "PEBC12",
	"FALABELLA":                 "PEBF01",
	"BANCO_FALABELLA":           "PEBF01",
	"GNB":                       "PEBG01",
	"BANCO_GNB":                 "PEBG01",
	"BANCO_GNB_PERU":            "PEBG01",
	"INTERBANK":                 "PEBI01",
	"ICBC":                      "PEBI02",
	"ICBC_PERU":                 "PEBI02",
	"MIBANCO":                   "PEBM01",
	"PICHINCHA":                 "PEBP01",
	"BANCO_PICHINCHA":           "PEBP01",
	"RIPLEY":                    "PEBR01",
	"BANCO_RIPLEY":              "PEBR01",
	"SANTANDER":                 "PEBS01",
	"SANTANDER_PERU":            "PEBS01",
	"SCOTIABANK":                "PEBS02",
	"SCOTIABANK_PERU":           "PEBS02",
}

// walletBankCodes 电子钱包银行编码。代付时 trade_type 与 bank_code 同值，
// 且 bank_account 填开户手机号（9 开头 9 位）而非 20 位 CCI 账号。
var walletBankCodes = map[string]struct{}{
	"PEW001": {},
	"PEW002": {},
}

func resolvePayoutBankCode(bankCode string) string {
	code := normalizeBankCodeKey(bankCode)
	if code == "" {
		return ""
	}
	if isVcpayPayoutBankCode(code) {
		return code
	}
	// 前端可能带国家前缀，如 PENBCP / PEBCP
	for _, prefix := range []string{"PEN", "PE"} {
		if trimmed := strings.TrimPrefix(code, prefix); trimmed != code && trimmed != "" {
			if mapped, ok := payoutBankCodeMap[trimmed]; ok {
				return mapped
			}
		}
	}
	if mapped, ok := payoutBankCodeMap[code]; ok {
		return mapped
	}
	return strings.ToUpper(strings.TrimSpace(bankCode))
}

func isWalletBankCode(code string) bool {
	_, ok := walletBankCodes[strings.ToUpper(strings.TrimSpace(code))]
	return ok
}

// normalizeBankCodeKey 转大写，并把空格/连字符/点归一成下划线，便于匹配 payoutBankCodeMap 的 key。
func normalizeBankCodeKey(bankCode string) string {
	var builder strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(bankCode)) {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_':
			builder.WriteRune(r)
		case r == ' ' || r == '-' || r == '.':
			builder.WriteRune('_')
		}
	}
	return builder.String()
}

// isVcpayPayoutBankCode 判断是否已经是 VcPay 秘鲁原生编码：PE + B/W + 3 位字母数字。
// 如 PEW001（Yape）、PEBA01（Banco Azteca）、PEBC10（Caja Trujillo）。
func isVcpayPayoutBankCode(code string) bool {
	if len(code) != 6 || !strings.HasPrefix(code, "PE") {
		return false
	}
	if code[2] != 'B' && code[2] != 'W' {
		return false
	}
	for i := 3; i < 6; i++ {
		if !((code[i] >= '0' && code[i] <= '9') || (code[i] >= 'A' && code[i] <= 'Z')) {
			return false
		}
	}
	return true
}
