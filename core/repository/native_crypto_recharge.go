package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

const (
	BitcoinEnabledConfigKey         = "btc_enabled"
	BitcoinReceiveAddressConfigKey  = "btc_receive_address"
	BitcoinPlatformRateConfigKey    = "btc_platform_rate"
	EthereumEnabledConfigKey        = "eth_enabled"
	EthereumReceiveAddressConfigKey = "eth_receive_address"
	EthereumPlatformRateConfigKey   = "eth_platform_rate"

	cryptoDefaultExpireMinutes   = 30
	cryptoDefaultScanSeconds     = 10
	bitcoinDefaultConfirmations  = 3
	ethereumDefaultConfirmations = 12
	cryptoTailAllocationAttempts = int64(10000)
	cryptoPlatformMinorUnits     = int64(100)
)

var ethereumMainnetAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
var positiveDecimalPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?$`)

const bitcoinBase58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

type BitcoinRuntimeConfig struct {
	Enabled             bool
	EsploraAPIBaseURL   string
	ReceiveAddress      string
	PlatformRate        string
	OrderExpireMinutes  int
	ScanIntervalSeconds int
	MinConfirmations    int
}

type EthereumRuntimeConfig struct {
	Enabled             bool
	RPCURL              string
	ReceiveAddress      string
	PlatformRate        string
	OrderExpireMinutes  int
	ScanIntervalSeconds int
	MinConfirmations    int
}

type cryptoAssetRuntimeConfig struct {
	Network            string
	Token              string
	Channel            string
	PayMethod          string
	Provider           string
	ContractAddress    string
	ReceiveAddress     string
	PlatformRate       string
	Decimals           int
	OrderExpireMinutes int
}

func LoadBitcoinRuntimeConfig(tablePrefix string) (BitcoinRuntimeConfig, error) {
	global := utils.GlobalConfig.Pay.Bitcoin
	cfg := BitcoinRuntimeConfig{
		EsploraAPIBaseURL:   strings.TrimRight(strings.TrimSpace(global.EsploraAPIBaseURL), "/"),
		ReceiveAddress:      strings.TrimSpace(global.DefaultReceiveAddress),
		OrderExpireMinutes:  global.OrderExpireMinutes,
		ScanIntervalSeconds: global.ScanIntervalSeconds,
		MinConfirmations:    global.MinConfirmations,
	}
	if cfg.OrderExpireMinutes <= 0 {
		cfg.OrderExpireMinutes = cryptoDefaultExpireMinutes
	}
	if cfg.ScanIntervalSeconds < cryptoDefaultScanSeconds {
		cfg.ScanIntervalSeconds = cryptoDefaultScanSeconds
	}
	if cfg.MinConfirmations <= 0 {
		cfg.MinConfirmations = bitcoinDefaultConfirmations
	}
	if !global.Enabled {
		return cfg, nil
	}
	cfg.Enabled = parseUsdtBoolConfig(loadUsdtTenantConfig(tablePrefix, BitcoinEnabledConfigKey, "0"))
	cfg.ReceiveAddress = strings.TrimSpace(loadUsdtTenantConfig(tablePrefix, BitcoinReceiveAddressConfigKey, cfg.ReceiveAddress))
	if strings.HasPrefix(strings.ToLower(cfg.ReceiveAddress), "bc1") && cfg.ReceiveAddress == strings.ToUpper(cfg.ReceiveAddress) {
		cfg.ReceiveAddress = strings.ToLower(cfg.ReceiveAddress)
	}
	cfg.PlatformRate = strings.TrimSpace(loadUsdtTenantConfig(tablePrefix, BitcoinPlatformRateConfigKey, ""))
	return cfg, nil
}

func ValidateBitcoinRuntimeConfig(tablePrefix string) (BitcoinRuntimeConfig, error) {
	cfg, err := LoadBitcoinRuntimeConfig(tablePrefix)
	if err != nil {
		return cfg, err
	}
	if !cfg.Enabled {
		return cfg, errors.New("btc_disabled")
	}
	if cfg.EsploraAPIBaseURL == "" {
		return cfg, errors.New("btc_esplora_api_required")
	}
	if cfg.ReceiveAddress == "" {
		return cfg, errors.New("btc_receive_address_required")
	}
	if !isBitcoinMainnetAddress(cfg.ReceiveAddress) {
		return cfg, errors.New("btc_receive_address_invalid")
	}
	if err = validateCryptoReceiveAddressUnique(tablePrefix, pojo.CryptoRechargeNetworkBitcoin, cfg.ReceiveAddress); err != nil {
		return cfg, err
	}
	if _, err = parsePositiveRat(cfg.PlatformRate); err != nil {
		return cfg, errors.New("btc_platform_rate_invalid")
	}
	return cfg, nil
}

func isBitcoinMainnetAddress(address string) bool {
	address = strings.TrimSpace(address)
	if strings.HasPrefix(strings.ToLower(address), "bc1") {
		return isValidBitcoinBech32MainnetAddress(address)
	}
	if strings.HasPrefix(address, "1") || strings.HasPrefix(address, "3") {
		return isValidBitcoinBase58MainnetAddress(address)
	}
	return false
}

func isValidBitcoinBase58MainnetAddress(address string) bool {
	payload, ok := decodeBase58Check(address)
	return ok && len(payload) == 21 && (payload[0] == 0x00 || payload[0] == 0x05)
}

func decodeBase58Check(address string) ([]byte, bool) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, false
	}
	decoded := big.NewInt(0)
	base := big.NewInt(58)
	for _, char := range address {
		index := strings.IndexRune(bitcoinBase58Alphabet, char)
		if index < 0 {
			return nil, false
		}
		decoded.Mul(decoded, base)
		decoded.Add(decoded, big.NewInt(int64(index)))
	}
	bytes := decoded.Bytes()
	leadingZeroes := 0
	for leadingZeroes < len(address) && address[leadingZeroes] == '1' {
		leadingZeroes++
	}
	decodedWithChecksum := append(make([]byte, leadingZeroes), bytes...)
	if len(decodedWithChecksum) < 5 {
		return nil, false
	}
	payload := decodedWithChecksum[:len(decodedWithChecksum)-4]
	checksum := decodedWithChecksum[len(decodedWithChecksum)-4:]
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	if string(checksum) != string(second[:4]) {
		return nil, false
	}
	return payload, true
}

func isValidTronMainnetAddress(address string) bool {
	payload, ok := decodeBase58Check(address)
	return ok && len(payload) == 21 && payload[0] == 0x41
}

func isValidBitcoinBech32MainnetAddress(address string) bool {
	if len(address) < 14 || len(address) > 90 || (address != strings.ToLower(address) && address != strings.ToUpper(address)) {
		return false
	}
	address = strings.ToLower(address)
	separator := strings.LastIndexByte(address, '1')
	if separator != 2 || address[:separator] != "bc" || separator+7 > len(address) {
		return false
	}
	charset := "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	data := make([]byte, 0, len(address)-separator-1)
	for _, char := range address[separator+1:] {
		index := strings.IndexRune(charset, char)
		if index < 0 {
			return false
		}
		data = append(data, byte(index))
	}
	if len(data) < 7 || data[0] > 16 {
		return false
	}
	checksum := bitcoinBech32Polymod(append(bitcoinBech32HRPExpand("bc"), data...))
	if (data[0] == 0 && checksum != 1) || (data[0] != 0 && checksum != 0x2bc830a3) {
		return false
	}
	program, ok := bitcoinConvertBits(data[1:len(data)-6], 5, 8, false)
	if !ok || len(program) < 2 || len(program) > 40 {
		return false
	}
	return data[0] != 0 || len(program) == 20 || len(program) == 32
}

func bitcoinBech32HRPExpand(hrp string) []byte {
	result := make([]byte, 0, len(hrp)*2+1)
	for _, char := range hrp {
		result = append(result, byte(char>>5))
	}
	result = append(result, 0)
	for _, char := range hrp {
		result = append(result, byte(char&31))
	}
	return result
}

func bitcoinBech32Polymod(values []byte) uint32 {
	checksum := uint32(1)
	generators := [5]uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	for _, value := range values {
		top := checksum >> 25
		checksum = (checksum&0x1ffffff)<<5 ^ uint32(value)
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				checksum ^= generators[i]
			}
		}
	}
	return checksum
}

func bitcoinConvertBits(data []byte, fromBits int, toBits int, pad bool) ([]byte, bool) {
	accumulator := uint32(0)
	bits := 0
	maxValue := uint32(1<<toBits) - 1
	maxAccumulator := uint32(1<<(fromBits+toBits-1)) - 1
	result := make([]byte, 0, len(data)*fromBits/toBits)
	for _, value := range data {
		if value>>fromBits != 0 {
			return nil, false
		}
		accumulator = (accumulator<<fromBits | uint32(value)) & maxAccumulator
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, byte(accumulator>>bits&maxValue))
		}
	}
	if pad {
		if bits > 0 {
			result = append(result, byte(accumulator<<(toBits-bits)&maxValue))
		}
	} else if bits >= fromBits || accumulator<<(toBits-bits)&maxValue != 0 {
		return nil, false
	}
	return result, true
}

func LoadEthereumRuntimeConfig(tablePrefix string) (EthereumRuntimeConfig, error) {
	global := utils.GlobalConfig.Pay.Ethereum
	cfg := EthereumRuntimeConfig{
		RPCURL:              strings.TrimSpace(global.RPCURL),
		ReceiveAddress:      strings.TrimSpace(global.DefaultReceiveAddress),
		OrderExpireMinutes:  global.OrderExpireMinutes,
		ScanIntervalSeconds: global.ScanIntervalSeconds,
		MinConfirmations:    global.MinConfirmations,
	}
	if cfg.OrderExpireMinutes <= 0 {
		cfg.OrderExpireMinutes = cryptoDefaultExpireMinutes
	}
	if cfg.ScanIntervalSeconds < cryptoDefaultScanSeconds {
		cfg.ScanIntervalSeconds = cryptoDefaultScanSeconds
	}
	if cfg.MinConfirmations <= 0 {
		cfg.MinConfirmations = ethereumDefaultConfirmations
	}
	if !global.Enabled {
		return cfg, nil
	}
	cfg.Enabled = parseUsdtBoolConfig(loadUsdtTenantConfig(tablePrefix, EthereumEnabledConfigKey, "0"))
	cfg.ReceiveAddress = strings.TrimSpace(loadUsdtTenantConfig(tablePrefix, EthereumReceiveAddressConfigKey, cfg.ReceiveAddress))
	cfg.PlatformRate = strings.TrimSpace(loadUsdtTenantConfig(tablePrefix, EthereumPlatformRateConfigKey, ""))
	return cfg, nil
}

func ValidateEthereumRuntimeConfig(tablePrefix string) (EthereumRuntimeConfig, error) {
	cfg, err := LoadEthereumRuntimeConfig(tablePrefix)
	if err != nil {
		return cfg, err
	}
	if !cfg.Enabled {
		return cfg, errors.New("eth_disabled")
	}
	if cfg.RPCURL == "" {
		return cfg, errors.New("eth_rpc_url_required")
	}
	if !isValidEthereumMainnetAddress(cfg.ReceiveAddress) {
		return cfg, errors.New("eth_receive_address_invalid")
	}
	cfg.ReceiveAddress = strings.ToLower(cfg.ReceiveAddress)
	if err = validateCryptoReceiveAddressUnique(tablePrefix, pojo.CryptoRechargeNetworkEthereum, cfg.ReceiveAddress); err != nil {
		return cfg, err
	}
	if _, err = parsePositiveRat(cfg.PlatformRate); err != nil {
		return cfg, errors.New("eth_platform_rate_invalid")
	}
	return cfg, nil
}

func isValidEthereumMainnetAddress(address string) bool {
	if !ethereumMainnetAddressPattern.MatchString(address) {
		return false
	}
	hexAddress := address[2:]
	if strings.Trim(hexAddress, "0") == "" {
		return false
	}
	if hexAddress == strings.ToLower(hexAddress) || hexAddress == strings.ToUpper(hexAddress) {
		return true
	}
	lowerAddress := strings.ToLower(hexAddress)
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write([]byte(lowerAddress))
	checksum := hasher.Sum(nil)
	for index, char := range hexAddress {
		if char >= '0' && char <= '9' {
			continue
		}
		nibble := checksum[index/2]
		if index%2 == 0 {
			nibble >>= 4
		} else {
			nibble &= 0x0f
		}
		shouldUpper := nibble >= 8
		if shouldUpper != (char >= 'A' && char <= 'F') {
			return false
		}
	}
	return true
}

func validateCryptoReceiveAddressUnique(tablePrefix string, network string, receiveAddress string) error {
	tablePrefix = strings.TrimSpace(tablePrefix)
	receiveAddress = normalizeCryptoRechargeAddress(network, receiveAddress)
	for _, host := range utils.GetTempHostInfos() {
		otherPrefix := strings.TrimSpace(host.TablePrefix)
		if otherPrefix == "" || otherPrefix == tablePrefix {
			continue
		}
		var enabled bool
		var otherAddress string
		switch network {
		case pojo.CryptoRechargeNetworkBitcoin:
			cfg, err := LoadBitcoinRuntimeConfig(otherPrefix)
			if err != nil {
				return err
			}
			enabled = cfg.Enabled
			otherAddress = cfg.ReceiveAddress
		case pojo.CryptoRechargeNetworkEthereum:
			cfg, err := LoadEthereumRuntimeConfig(otherPrefix)
			if err != nil {
				return err
			}
			enabled = cfg.Enabled
			otherAddress = cfg.ReceiveAddress
		case pojo.CryptoRechargeNetworkTRC20:
			cfg, err := LoadUsdtTrc20RuntimeConfig(otherPrefix)
			if err != nil {
				return err
			}
			enabled = cfg.Enabled
			otherAddress = cfg.ReceiveAddress
		default:
			return errors.New("unsupported_crypto_network")
		}
		if enabled && normalizeCryptoRechargeAddress(network, otherAddress) == receiveAddress {
			return errors.New("crypto_receive_address_shared_across_tenants")
		}
		otherDB := utils.NewPrefixDb(otherPrefix)
		if otherDB == nil {
			return errors.New("crypto_receive_address_tenant_db_unavailable")
		}
		if otherDB.Migrator().HasTable(&pojo.UsdtRechargeOrder{}) {
			var historicalCount int64
			historicalQuery := otherDB.Clauses(dbresolver.Write).Model(&pojo.UsdtRechargeOrder{}).Where("network = ?", network)
			if network == pojo.CryptoRechargeNetworkEthereum || strings.HasPrefix(receiveAddress, "bc1") {
				historicalQuery = historicalQuery.Where("LOWER(receive_address) = ?", strings.ToLower(receiveAddress))
			} else {
				historicalQuery = historicalQuery.Where("receive_address = ?", receiveAddress)
			}
			if err := historicalQuery.Count(&historicalCount).Error; err != nil {
				return err
			}
			if historicalCount > 0 {
				return errors.New("crypto_receive_address_reserved_by_other_tenant")
			}
		}
	}
	return nil
}

func ValidateCryptoScanTargetOwnership(tablePrefix string, network string, receiveAddress string) error {
	return validateCryptoReceiveAddressUnique(tablePrefix, network, receiveAddress)
}

func listEnabledCryptoAssetConfigs(tablePrefix string) ([]cryptoAssetRuntimeConfig, error) {
	result := make([]cryptoAssetRuntimeConfig, 0, 3)

	usdt, err := LoadUsdtTrc20RuntimeConfig(tablePrefix)
	if err != nil {
		return nil, err
	}
	if usdt.Enabled {
		if usdt, err = validateUsdtTrc20RuntimeConfig(tablePrefix); err != nil {
			return nil, err
		}
		result = append(result, cryptoAssetRuntimeConfig{
			Network: pojo.CryptoRechargeNetworkTRC20, Token: pojo.CryptoRechargeTokenUSDT,
			Channel: pojo.CryptoRechargeChannelUSDTTRC20, PayMethod: pojo.CryptoRechargePayMethodTRC20,
			Provider: "native_usdt_trc20", ContractAddress: usdt.ContractAddress,
			ReceiveAddress: usdt.ReceiveAddress, PlatformRate: usdt.PlatformRate,
			Decimals: pojo.CryptoRechargeDecimalsUSDT, OrderExpireMinutes: usdt.OrderExpireMinutes,
		})
	}

	btc, err := LoadBitcoinRuntimeConfig(tablePrefix)
	if err != nil {
		return nil, err
	}
	if btc.Enabled {
		if btc, err = ValidateBitcoinRuntimeConfig(tablePrefix); err != nil {
			return nil, err
		}
		result = append(result, cryptoAssetRuntimeConfig{
			Network: pojo.CryptoRechargeNetworkBitcoin, Token: pojo.CryptoRechargeTokenBTC,
			Channel: pojo.CryptoRechargeChannelBTC, PayMethod: pojo.CryptoRechargePayMethodBTC,
			Provider: "native_btc", ReceiveAddress: btc.ReceiveAddress,
			PlatformRate: btc.PlatformRate, Decimals: pojo.CryptoRechargeDecimalsBTC,
			OrderExpireMinutes: btc.OrderExpireMinutes,
		})
	}

	eth, err := LoadEthereumRuntimeConfig(tablePrefix)
	if err != nil {
		return nil, err
	}
	if eth.Enabled {
		if eth, err = ValidateEthereumRuntimeConfig(tablePrefix); err != nil {
			return nil, err
		}
		result = append(result, cryptoAssetRuntimeConfig{
			Network: pojo.CryptoRechargeNetworkEthereum, Token: pojo.CryptoRechargeTokenETH,
			Channel: pojo.CryptoRechargeChannelETH, PayMethod: pojo.CryptoRechargePayMethodETH,
			Provider: "native_eth", ReceiveAddress: strings.ToLower(eth.ReceiveAddress),
			PlatformRate: eth.PlatformRate, Decimals: pojo.CryptoRechargeDecimalsETH,
			OrderExpireMinutes: eth.OrderExpireMinutes,
		})
	}

	return result, nil
}

func getCryptoAssetConfig(tablePrefix, network, token string) (cryptoAssetRuntimeConfig, error) {
	network = strings.ToUpper(strings.TrimSpace(network))
	token = strings.ToUpper(strings.TrimSpace(token))
	if network == "" && token == "" {
		network = pojo.CryptoRechargeNetworkTRC20
		token = pojo.CryptoRechargeTokenUSDT
	}
	switch {
	case network == pojo.CryptoRechargeNetworkTRC20 && token == pojo.CryptoRechargeTokenUSDT:
		cfg, err := validateUsdtTrc20RuntimeConfig(tablePrefix)
		if err != nil {
			return cryptoAssetRuntimeConfig{}, err
		}
		return cryptoAssetRuntimeConfig{
			Network: pojo.CryptoRechargeNetworkTRC20, Token: pojo.CryptoRechargeTokenUSDT,
			Channel: pojo.CryptoRechargeChannelUSDTTRC20, PayMethod: pojo.CryptoRechargePayMethodTRC20,
			Provider: "native_usdt_trc20", ContractAddress: cfg.ContractAddress,
			ReceiveAddress: cfg.ReceiveAddress, PlatformRate: cfg.PlatformRate,
			Decimals: pojo.CryptoRechargeDecimalsUSDT, OrderExpireMinutes: cfg.OrderExpireMinutes,
		}, nil
	case network == pojo.CryptoRechargeNetworkBitcoin && token == pojo.CryptoRechargeTokenBTC:
		cfg, err := ValidateBitcoinRuntimeConfig(tablePrefix)
		if err != nil {
			return cryptoAssetRuntimeConfig{}, err
		}
		return cryptoAssetRuntimeConfig{
			Network: pojo.CryptoRechargeNetworkBitcoin, Token: pojo.CryptoRechargeTokenBTC,
			Channel: pojo.CryptoRechargeChannelBTC, PayMethod: pojo.CryptoRechargePayMethodBTC,
			Provider: "native_btc", ReceiveAddress: cfg.ReceiveAddress,
			PlatformRate: cfg.PlatformRate, Decimals: pojo.CryptoRechargeDecimalsBTC,
			OrderExpireMinutes: cfg.OrderExpireMinutes,
		}, nil
	case network == pojo.CryptoRechargeNetworkEthereum && token == pojo.CryptoRechargeTokenETH:
		cfg, err := ValidateEthereumRuntimeConfig(tablePrefix)
		if err != nil {
			return cryptoAssetRuntimeConfig{}, err
		}
		return cryptoAssetRuntimeConfig{
			Network: pojo.CryptoRechargeNetworkEthereum, Token: pojo.CryptoRechargeTokenETH,
			Channel: pojo.CryptoRechargeChannelETH, PayMethod: pojo.CryptoRechargePayMethodETH,
			Provider: "native_eth", ReceiveAddress: strings.ToLower(cfg.ReceiveAddress),
			PlatformRate: cfg.PlatformRate, Decimals: pojo.CryptoRechargeDecimalsETH,
			OrderExpireMinutes: cfg.OrderExpireMinutes,
		}, nil
	}
	return cryptoAssetRuntimeConfig{}, errors.New("unsupported_crypto_asset")
}

func calculateCryptoBaseAmountAtomic(platformAmount float64, platformRate string, decimals int) (*big.Int, error) {
	if decimals < 0 || decimals > pojo.CryptoRechargeDecimalsETH {
		return nil, errors.New("crypto_decimals_invalid")
	}
	amountRat, err := parsePositiveRat(strconv.FormatFloat(platformAmount, 'f', 2, 64))
	if err != nil {
		return nil, err
	}
	rateRat, err := parsePositiveRat(platformRate)
	if err != nil {
		return nil, err
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	scaled := new(big.Rat).Mul(amountRat, new(big.Rat).SetInt(scale))
	scaled.Quo(scaled, rateRat)
	return ceilPositiveRatToBigInt(scaled)
}

func ceilPositiveRatToBigInt(rat *big.Rat) (*big.Int, error) {
	if rat == nil || rat.Sign() <= 0 {
		return nil, errors.New("invalid_decimal")
	}
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(new(big.Int).Set(rat.Num()), new(big.Int).Set(rat.Denom()), remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient, nil
}

func formatCryptoAtomicAmount(amount *big.Int, decimals int) string {
	if amount == nil {
		return ""
	}
	negative := amount.Sign() < 0
	value := new(big.Int).Abs(new(big.Int).Set(amount)).String()
	if decimals <= 0 {
		if negative {
			return "-" + value
		}
		return value
	}
	if len(value) <= decimals {
		value = strings.Repeat("0", decimals-len(value)+1) + value
	}
	whole := value[:len(value)-decimals]
	fraction := value[len(value)-decimals:]
	formatted := whole + "." + fraction
	if negative {
		return "-" + formatted
	}
	return formatted
}

func buildCryptoQRContent(cfg cryptoAssetRuntimeConfig, expectedAmountAtomic *big.Int, expectedAmount string) (string, error) {
	if expectedAmountAtomic == nil || expectedAmountAtomic.Sign() <= 0 {
		return "", errors.New("crypto_amount_invalid")
	}
	switch cfg.Network {
	case pojo.CryptoRechargeNetworkBitcoin:
		return "bitcoin:" + cfg.ReceiveAddress + "?amount=" + expectedAmount, nil
	case pojo.CryptoRechargeNetworkEthereum:
		return "ethereum:" + cfg.ReceiveAddress + "@1?value=" + expectedAmountAtomic.String(), nil
	case pojo.CryptoRechargeNetworkTRC20:
		return cfg.ReceiveAddress, nil
	default:
		return "", errors.New("unsupported_crypto_asset")
	}
}

func parseCryptoAtomicAmount(value string) (*big.Int, error) {
	parsed, ok := new(big.Int).SetString(strings.TrimSpace(value), 10)
	if !ok || parsed.Sign() <= 0 {
		return nil, errors.New("crypto_amount_invalid")
	}
	return parsed, nil
}

func allocateCryptoExpectedAmountAtomic(tx *gorm.DB, cfg cryptoAssetRuntimeConfig, baseAmount *big.Int, orderNo string) (*big.Int, error) {
	if baseAmount == nil || baseAmount.Sign() <= 0 {
		return nil, errors.New("crypto_amount_invalid")
	}
	if strings.TrimSpace(orderNo) == "" {
		return nil, errors.New("order_no_required")
	}
	maxTail, err := calculateCryptoMaxTailAtomic(cfg.PlatformRate, cfg.Decimals)
	if err != nil {
		return nil, err
	}
	candidateCount := new(big.Int).Add(new(big.Int).Set(maxTail), big.NewInt(1))
	seedHash := sha256.Sum256([]byte(strings.Join([]string{cfg.Network, cfg.Token, cfg.ReceiveAddress, orderNo}, "|")))
	start := new(big.Int).Mod(new(big.Int).SetBytes(seedHash[:]), candidateCount)
	attempts := cryptoTailAllocationAttempts
	if candidateCount.IsInt64() && candidateCount.Int64() < attempts {
		attempts = candidateCount.Int64()
	}
	for i := int64(0); i < attempts; i++ {
		tail := new(big.Int).Add(new(big.Int).Set(start), big.NewInt(i))
		tail.Mod(tail, candidateCount)
		expected := new(big.Int).Add(new(big.Int).Set(baseAmount), tail)
		keySource := strings.Join([]string{cfg.Network, cfg.Token, cfg.ReceiveAddress, expected.String()}, "|")
		reservationHash := sha256.Sum256([]byte(keySource))
		reservation := pojo.CryptoRechargeAmountReservation{
			ReservationKey:       fmt.Sprintf("%x", reservationHash[:]),
			Network:              cfg.Network,
			Token:                cfg.Token,
			ReceiveAddress:       cfg.ReceiveAddress,
			ExpectedAmountAtomic: expected.String(),
			OrderNo:              orderNo,
		}
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&reservation)
		if created.Error != nil {
			return nil, created.Error
		}
		if created.RowsAffected == 1 {
			return expected, nil
		}
	}
	return nil, errors.New("crypto_expected_amount_capacity_exhausted_rotate_receive_address")
}

func calculateCryptoMaxTailAtomic(platformRate string, decimals int) (*big.Int, error) {
	if decimals < 0 || decimals > pojo.CryptoRechargeDecimalsETH {
		return nil, errors.New("crypto_decimals_invalid")
	}
	rate, err := parsePositiveRat(platformRate)
	if err != nil {
		return nil, err
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	denominator := new(big.Rat).Mul(rate, new(big.Rat).SetInt64(cryptoPlatformMinorUnits))
	maximum := new(big.Rat).Quo(new(big.Rat).SetInt(scale), denominator)
	return new(big.Int).Quo(maximum.Num(), maximum.Denom()), nil
}
