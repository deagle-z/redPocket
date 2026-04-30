package utils

import "math"

type Money int64

const moneyScale = 100

// Truncate2 keeps two decimal places by discarding the remaining fractional part.
func Truncate2(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Trunc(adjustScaledMoney(value)) / moneyScale
}

// ToMoney converts a float64 amount in dollars to Money with 2 decimal places.
func ToMoney(dollars float64) Money {
	if math.IsNaN(dollars) || math.IsInf(dollars, 0) {
		return 0
	}
	return Money(math.Trunc(adjustScaledMoney(dollars)))
}

func adjustScaledMoney(value float64) float64 {
	scaled := value * moneyScale
	if scaled > 0 {
		return scaled + 1e-9
	}
	if scaled < 0 {
		return scaled - 1e-9
	}
	return scaled
}

// ToDollars converts Money to float64 dollars.
func (m Money) ToDollars() float64 {
	return float64(m) / moneyScale
}

// Add adds two Money values
func (m Money) Add(other Money) Money {
	return m + other
}

// Subtract subtracts another Money value from the current Money value
func (m Money) Subtract(other Money) Money {
	return m - other
}

// Multiply multiplies Money by a factor, keeping two decimal places by truncation.
func (m Money) Multiply(factor float64) Money {
	result := float64(m) * factor
	return ToMoney(result / moneyScale)
}

// Divide divides Money by a divisor, keeping two decimal places by truncation.
func (m Money) Divide(divisor float64) Money {
	result := float64(m) / divisor
	return ToMoney(result / moneyScale)
}
