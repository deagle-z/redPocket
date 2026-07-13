export interface PrizePoolBalancePayload {
  poolCode: string
  balance: number
}

export const usePrizePoolStore = defineStore('prizePool', () => {
  const balances = ref<Record<string, number>>({})

  const luckyBalance = computed(() => getBalance('lucky'))

  function normalizePoolCode(poolCode: unknown) {
    return typeof poolCode === 'string' ? poolCode.trim() : ''
  }

  function getBalance(poolCode: string) {
    const normalizedPoolCode = normalizePoolCode(poolCode)
    if (!normalizedPoolCode) return null

    return balances.value[normalizedPoolCode] ?? null
  }

  function updatePrizePoolBalance(payload: PrizePoolBalancePayload) {
    const poolCode = normalizePoolCode(payload.poolCode)
    const balance = Number(payload.balance)

    if (!poolCode || !Number.isFinite(balance)) return

    balances.value = {
      ...balances.value,
      [poolCode]: balance,
    }
  }

  return {
    balances,
    luckyBalance,
    getBalance,
    updatePrizePoolBalance,
  }
})
