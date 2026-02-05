package utils

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
)

// RedEnvelope 红包金额分配算法
// totalAmount: 红包总金额
// totalCount: 红包总个数
// minAmount: 每个红包最小金额
// maxAmount: 每个红包最大金额
// 返回: 红包金额数组
func RedEnvelope(totalAmount float64, totalCount int, minAmount float64, maxAmount float64) []float64 {
	result := make([]float64, 0, totalCount)
	leftCents := int64(math.Round(totalAmount * 100)) // 剩余金额(分)
	leftCount := totalCount                           // 剩余个数
	minCents := int64(math.Round(minAmount * 100))
	maxCents := int64(math.Round(maxAmount * 100))
	if minCents < 1 {
		minCents = 1
	}

	for i := 1; i <= totalCount; i++ {
		if leftCount == 1 {
			// 最后一个红包，剩余金额全部放入红包
			lastCents := leftCents
			// 确保不是整数（末两位不为00）
			if lastCents%100 == 0 && len(result) > 0 {
				for idx := 0; idx < len(result); idx++ {
					prevCents := int64(math.Round(result[idx] * 100))
					// 尝试从前一个红包挪1分给最后一个
					if prevCents-1 >= minCents && (prevCents-1)%100 != 0 && lastCents+1 <= maxCents && (lastCents+1)%100 != 0 {
						prevCents -= 1
						lastCents += 1
						result[idx] = float64(prevCents) / 100
						break
					}
					// 或者从最后一个挪1分给前一个
					if prevCents+1 <= maxCents && (prevCents+1)%100 != 0 && lastCents-1 >= minCents && (lastCents-1)%100 != 0 {
						prevCents += 1
						lastCents -= 1
						result[idx] = float64(prevCents) / 100
						break
					}
				}
			}
			result = append(result, float64(lastCents)/100)
		} else {
			// 计算随机金额（分）
			max := minInt64(maxCents, leftCents-int64(leftCount-1)*minCents) // 红包最大金额不能超过剩余金额和最大金额的较小值
			min := maxInt64(minCents, leftCents-int64(leftCount-1)*maxCents) // 红包最小金额不能低于剩余金额和最小金额的较大值
			amountCents := randomCentsNotInteger(min, max)
			// 如果只剩最后一个红包且会是整数，重试
			if leftCount == 2 {
				for tries := 0; tries < 20 && (leftCents-amountCents)%100 == 0; tries++ {
					amountCents = randomCentsNotInteger(min, max)
				}
			}
			result = append(result, float64(amountCents)/100)
			leftCents -= amountCents
			leftCount--
		}
	}

	return result
}

// RandomFloat 生成0-1之间的随机浮点数
func RandomFloat() float64 {
	return rand.Float64()
}

func randomCentsNotInteger(minCents int64, maxCents int64) int64 {
	if minCents > maxCents {
		return minCents
	}
	for i := 0; i < 20; i++ {
		amount := minCents + rand.Int64N(maxCents-minCents+1)
		if amount%100 != 0 {
			return amount
		}
	}
	// fallback: 调整1分避免整数
	amount := minCents
	if amount%100 == 0 {
		if amount+1 <= maxCents {
			amount++
		} else if amount-1 >= minCents {
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
