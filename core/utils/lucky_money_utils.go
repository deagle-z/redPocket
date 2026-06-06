package utils

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

// LuckyEnvelopeAmountBounds returns per-packet amount bounds based on the total amount.
func LuckyEnvelopeAmountBounds(totalAmount float64, totalCount int) (float64, float64) {
	minUnits, maxUnits := luckyEnvelopeAmountBoundsUnits(int64(ToMoney(totalAmount)), totalCount)
	return float64(minUnits) / 100, float64(maxUnits) / 100
}

func luckyEnvelopeAmountBoundsUnits(totalUnits int64, totalCount int) (int64, int64) {
	if totalUnits <= 0 {
		return 1, 1
	}
	if totalCount <= 0 {
		return 1, totalUnits
	}

	countUnits := int64(totalCount)
	minUnits := ceilDivInt64(totalUnits, 10) // ceil(total * 10%)
	maxUnits := (totalUnits * 2) / 5         // floor(total * 40%)

	if totalCount > 10 {
		minUnits = totalUnits / countUnits
	}
	if totalCount < 3 {
		maxUnits = ceilDivInt64(totalUnits, countUnits)
	}

	if minUnits < 1 {
		minUnits = 1
	}
	if maxUnits < 1 {
		maxUnits = 1
	}

	// Cent precision can make strict percentage bounds impossible, for example
	// 10 packets with a total not divisible into exact 10% cents.
	if minUnits*countUnits > totalUnits {
		minUnits = totalUnits / countUnits
		if minUnits < 1 {
			minUnits = 1
		}
	}
	if maxUnits*countUnits < totalUnits {
		maxUnits = ceilDivInt64(totalUnits, countUnits)
	}
	if maxUnits < minUnits {
		maxUnits = minUnits
	}

	return minUnits, maxUnits
}

func ceilDivInt64(a int64, b int64) int64 {
	if b <= 0 {
		return 0
	}
	return (a + b - 1) / b
}

// RedEnvelope 红包金额分配算法
// totalAmount: 红包总金额
// totalCount: 红包总个数
// minAmount: 每个红包最小金额
// maxAmount: 每个红包最大金额
// 返回: 红包金额数组
func RedEnvelope(totalAmount float64, totalCount int, minAmount float64, maxAmount float64) []float64 {
	result := make([]float64, 0, totalCount)
	leftUnits := int64(ToMoney(totalAmount)) // 剩余金额，按 0.01 计
	leftCount := totalCount                  // 剩余个数
	minUnits := int64(ToMoney(minAmount))
	maxUnits := int64(ToMoney(maxAmount))
	if minUnits < 1 {
		minUnits = 1
	}

	for i := 1; i <= totalCount; i++ {
		if leftCount == 1 {
			// 最后一个红包，剩余金额全部放入红包
			lastUnits := leftUnits
			// 确保不是整数（末两位不为00）
			if lastUnits%100 == 0 && len(result) > 0 {
				for idx := 0; idx < len(result); idx++ {
					prevUnits := int64(ToMoney(result[idx]))
					// 尝试从前一个红包挪0.01给最后一个
					if prevUnits-1 >= minUnits && (prevUnits-1)%100 != 0 && lastUnits+1 <= maxUnits && (lastUnits+1)%100 != 0 {
						prevUnits -= 1
						lastUnits += 1
						result[idx] = float64(prevUnits) / 100
						break
					}
					// 或者从最后一个挪0.01给前一个
					if prevUnits+1 <= maxUnits && (prevUnits+1)%100 != 0 && lastUnits-1 >= minUnits && (lastUnits-1)%100 != 0 {
						prevUnits += 1
						lastUnits -= 1
						result[idx] = float64(prevUnits) / 100
						break
					}
				}
			}
			result = append(result, float64(lastUnits)/100)
		} else {
			// 计算随机金额（0.01）
			max := minInt64(maxUnits, leftUnits-int64(leftCount-1)*minUnits) // 红包最大金额不能超过剩余金额和最大金额的较小值
			min := maxInt64(minUnits, leftUnits-int64(leftCount-1)*maxUnits) // 红包最小金额不能低于剩余金额和最小金额的较大值
			amountUnits := randomUnitsNotInteger(min, max)
			// 如果只剩最后一个红包且会是整数，重试
			if leftCount == 2 {
				for tries := 0; tries < 20 && (leftUnits-amountUnits)%100 == 0; tries++ {
					amountUnits = randomUnitsNotInteger(min, max)
				}
			}
			result = append(result, float64(amountUnits)/100)
			leftUnits -= amountUnits
			leftCount--
		}
	}

	return result
}

// RandomFloat 生成0-1之间的随机浮点数
func RandomFloat() float64 {
	return rand.Float64()
}

func randomUnitsNotInteger(minUnits int64, maxUnits int64) int64 {
	if minUnits > maxUnits {
		return minUnits
	}
	for i := 0; i < 20; i++ {
		amount := minUnits + rand.Int64N(maxUnits-minUnits+1)
		if amount%100 != 0 {
			return amount
		}
	}
	// fallback: 调整0.01避免整数
	amount := minUnits
	if amount%100 == 0 {
		if amount+1 <= maxUnits {
			amount++
		} else if amount-1 >= minUnits {
			amount--
		}
	}
	return amount
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ParseLuckyNumConfig 解析 lucky_num 配置项
// 支持格式：单个数字 "3" 或范围 "3|9"
// 返回: [min, max] 数组，如果是单个数字则 min=max
func ParseLuckyNumConfig(configValue string) (int, int) {
	if strings.Contains(configValue, "|") {
		// 范围格式：3|9
		parts := strings.Split(configValue, "|")
		min, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		max := min
		if len(parts) > 1 {
			max, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		}
		return min, max
	}
	// 单个数字格式：3
	num, _ := strconv.Atoi(configValue)
	return num, num
}

// GetLuckyNumMin 获取 lucky_num 配置的最小值
func GetLuckyNumMin(configValue string) int {
	min, _ := ParseLuckyNumConfig(configValue)
	return min
}

// ValidateLuckyCount 验证红包数量是否符合配置规则
// luckyCount: 用户指定的红包数量
// configValue: lucky_num 配置值
// 返回: [valid, message]
func ValidateLuckyCount(luckyCount int, configValue string) (bool, string) {
	min, max := ParseLuckyNumConfig(configValue)

	if luckyCount < min || luckyCount > max {
		return false, fmt.Sprintf("红包数量不符合规则！\n规则：红包数量必须在 %d-%d 之间\n您输入的数量：%d", min, max, luckyCount)
	}

	return true, ""
}

// FormatName 格式化用户名（截断）
func FormatName(str string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 8
	}
	runes := []rune(str)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + ".."
	}
	return str
}
