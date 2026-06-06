package utils

import "testing"

func TestRedEnvelopeUsesTotalAmountPercentBoundsForNormalCounts(t *testing.T) {
	totalAmount := 100.0
	totalCount := 5
	minAmount, maxAmount := LuckyEnvelopeAmountBounds(totalAmount, totalCount)

	assertBounds(t, minAmount, maxAmount, 10, 40)

	for i := 0; i < 100; i++ {
		got := RedEnvelope(totalAmount, totalCount, minAmount, maxAmount)
		assertEnvelopeAmountsWithinBounds(t, got, totalAmount, minAmount, maxAmount)
	}
}

func TestRedEnvelopeRelaxesMinimumToAverageFloorWhenCountExceedsTen(t *testing.T) {
	totalAmount := 100.0
	totalCount := 12
	minAmount, maxAmount := LuckyEnvelopeAmountBounds(totalAmount, totalCount)

	assertBounds(t, minAmount, maxAmount, 8.33, 40)

	for i := 0; i < 100; i++ {
		got := RedEnvelope(totalAmount, totalCount, minAmount, maxAmount)
		assertEnvelopeAmountsWithinBounds(t, got, totalAmount, minAmount, maxAmount)
	}
}

func TestRedEnvelopeRelaxesMaximumToAverageCeilWhenCountIsBelowThree(t *testing.T) {
	totalAmount := 100.0
	totalCount := 2
	minAmount, maxAmount := LuckyEnvelopeAmountBounds(totalAmount, totalCount)

	assertBounds(t, minAmount, maxAmount, 10, 50)

	for i := 0; i < 100; i++ {
		got := RedEnvelope(totalAmount, totalCount, minAmount, maxAmount)
		assertEnvelopeAmountsWithinBounds(t, got, totalAmount, minAmount, maxAmount)
	}
}

func TestLuckyEnvelopeAmountBoundsUsesCentsForFractionalTotals(t *testing.T) {
	totalAmount := 10.05
	totalCount := 3
	minAmount, maxAmount := LuckyEnvelopeAmountBounds(totalAmount, totalCount)

	assertBounds(t, minAmount, maxAmount, 1.01, 4.02)

	for i := 0; i < 100; i++ {
		got := RedEnvelope(totalAmount, totalCount, minAmount, maxAmount)
		assertEnvelopeAmountsWithinBounds(t, got, totalAmount, minAmount, maxAmount)
	}
}

func TestLuckyEnvelopeAmountBoundsKeepsCentPrecisionBoundsFeasible(t *testing.T) {
	totalAmount := 10.05
	totalCount := 10
	minAmount, maxAmount := LuckyEnvelopeAmountBounds(totalAmount, totalCount)

	assertBounds(t, minAmount, maxAmount, 1.00, 4.02)

	for i := 0; i < 100; i++ {
		got := RedEnvelope(totalAmount, totalCount, minAmount, maxAmount)
		assertEnvelopeAmountsWithinBounds(t, got, totalAmount, minAmount, maxAmount)
	}
}

func assertBounds(t *testing.T, minAmount float64, maxAmount float64, wantMin float64, wantMax float64) {
	t.Helper()

	if ToMoney(minAmount) != ToMoney(wantMin) || ToMoney(maxAmount) != ToMoney(wantMax) {
		t.Fatalf("LuckyEnvelopeAmountBounds() = (%.2f, %.2f), want (%.2f, %.2f)", minAmount, maxAmount, wantMin, wantMax)
	}
}

func assertEnvelopeAmountsWithinBounds(t *testing.T, amounts []float64, totalAmount float64, minAmount float64, maxAmount float64) {
	t.Helper()

	totalUnits := int64(ToMoney(totalAmount))
	minUnits := int64(ToMoney(minAmount))
	maxUnits := int64(ToMoney(maxAmount))
	sumUnits := int64(0)

	if len(amounts) == 0 {
		t.Fatalf("RedEnvelope() returned no amounts")
	}

	for idx, amount := range amounts {
		amountUnits := int64(ToMoney(amount))
		if amountUnits < minUnits || amountUnits > maxUnits {
			t.Fatalf("amount[%d] = %.2f outside [%.2f, %.2f]", idx, amount, minAmount, maxAmount)
		}
		sumUnits += amountUnits
	}

	if sumUnits != totalUnits {
		t.Fatalf("sum = %.2f, want %.2f", float64(sumUnits)/100, float64(totalUnits)/100)
	}
}
