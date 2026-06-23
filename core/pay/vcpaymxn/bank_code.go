package vcpaymxn

import "strings"

var payoutBankCodeMap = map[string]string{
	"ABCCAPITAL":      "MXA001",
	"ABC_CAPITAL":     "MXA001",
	"ACTINVER":        "MXA002",
	"AFIRME":          "MXA003",
	"ALTERNATIVOS":    "MXA004",
	"ARCUS":           "MXA005",
	"ARCUS_FI":        "MXA005",
	"ASP_INTEGRA_OPC": "MXA006",
	"AUTOFIN":         "MXA007",
	"AZTECA":          "MXA008",
	"ALBO":            "MXA009",
	"BABIEN":          "MXB001",
	"BAJIO":           "MXB002",
	"BANCO_DEL_BAJIO": "MXB002",
	"BDB":             "MXB002",
	"BANAMEX":         "MXB003",
	"BANCO_COVALTO":   "MXB004",
	"BANCO_S3":        "MXB005",
	"BANCOMEXT":       "MXB006",
	"BCT":             "MXB006",
	"BANCOPPEL":       "MXB007",
	"BANCREA":         "MXB008",
	"BANJERCITO":      "MXB009",
	"BANK_OF_AMERICA": "MXB010",
	"BANK_OF_CHINA":   "MXB011",
	"BANKAOOL":        "MXB012",
	"BANOBRAS":        "MXB013",
	"BANORTE":         "MXB014",
	"BANREGIO":        "MXB015",
	"BANSI":           "MXB016",
	"BARCLAYS":        "MXB017",
	"BBASE":           "MXB018",
	"BBVA_MEXICO":     "MXB019",
	"BBVABANCOMER":    "MXB019",
	"BMONEX":          "MXB020",
	"CAJA_POP_MEXICA": "MXC001",
	"CAJA_TELEFONIST": "MXC002",
	"CB_INTERCAM":     "MXC003",
	"CI_BOLSA":        "MXC004",
	"CIBANCO":         "MXC005",
	"CITI_MEXICO":     "MXC006",
	"CLS":             "MXC007",
	"COMPARTAMOS":     "MXC008",
	"CONSUBANCO":      "MXC009",
	"CREDICAPITAL":    "MXC010",
	"CREDICLUB":       "MXC011",
	"CRISTOBAL_COLON": "MXC012",
	"CUENCA":          "MXC013",
	"CREDIT_SUISSE":   "MXC014",
	"DONDE":           "MXD001",
	"FINAMEX":         "MXF001",
	"FINCOMUN":        "MXF002",
	"FOMPED":          "MXF003",
	"FONDEADORA":      "MXF004",
	"FONDO_FIRA":      "MXF005",
	"GBM":             "MXG001",
	"HIPOTECARIA_FED": "MXH001",
	"HSBC":            "MXH002",
	"ICBC":            "MXI001",
	"IBA":             "MXI002",
	"INBURSA":         "MXI002",
	"INDEVAL":         "MXI003",
	"INMOBILIARIO":    "MXI004",
	"INTERCAM_BANCO":  "MXI005",
	"INVERCAP":        "MXI006",
	"INVEX":           "MXI007",
	"JP_MORGAN":       "MXJ001",
	"KUSPIT":          "MXK001",
	"LIBERTAD":        "MXL001",
	"MASARI":          "MXM001",
	"MERCADO_PAGO_W":  "MXM002",
	"MIFEL":           "MXM003",
	"MIZUHO_BANK":     "MXM004",
	"MONEXCB":         "MXM005",
	"MUFG":            "MXM006",
	"MULTIVA_BANCO":   "MXM007",
	"MEXPAGO":         "MXM008",
	"NAFIN":           "MXN001",
	"NU_MEXICO":       "MXN002",
	"NVIO":            "MXN003",
	"PAGATODO":        "MXP001",
	"PEIBO":           "MXP002",
	"PROFUTURO":       "MXP003",
	"SABADELL":        "MXS001",
	"SANTANDER":       "MXS002",
	"SCOTIABANK":      "MXS003",
	"SHINHAN":         "MXS004",
	"SPIN_BY_OXXO":    "MXS005",
	"SPEI":            "MXS006",
	"STP":             "MXS006",
	"TESORED":         "MXT001",
	"TRANSFER":        "MXT002",
	"UNAGRA":          "MXU001",
	"VALMEX":          "MXV001",
	"VALUE":           "MXV002",
	"VE_POR_MAS":      "MXV003",
	"VECTOR":          "MXV004",
	"VOLKSWAGEN":      "MXV005",
}

func resolvePayoutBankCode(bankCode string) string {
	code := strings.ToUpper(strings.TrimSpace(bankCode))
	if code == "" {
		return ""
	}
	if isVcpayPayoutBankCode(code) {
		return code
	}
	if strings.HasPrefix(code, "MXN") {
		code = strings.TrimPrefix(code, "MXN")
	}
	if mapped, ok := payoutBankCodeMap[code]; ok {
		return mapped
	}
	return strings.ToUpper(strings.TrimSpace(bankCode))
}

func isVcpayPayoutBankCode(code string) bool {
	if len(code) != 6 || !strings.HasPrefix(code, "MX") {
		return false
	}
	return code[2] >= 'A' && code[2] <= 'Z' &&
		code[3] >= '0' && code[3] <= '9' &&
		code[4] >= '0' && code[4] <= '9' &&
		code[5] >= '0' && code[5] <= '9'
}
