import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { usePrizePoolStore } from './prizePool'

describe('prize pool realtime store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('stores valid prize pool balances by pool code', () => {
    const store = usePrizePoolStore()

    store.updatePrizePoolBalance({ poolCode: 'lucky', balance: 1234.56 })

    expect(store.getBalance('lucky')).toBe(1234.56)
    expect(store.luckyBalance).toBe(1234.56)
  })

  it('ignores missing pool codes and invalid balances', () => {
    const store = usePrizePoolStore()

    store.updatePrizePoolBalance({ poolCode: '', balance: 100 })
    store.updatePrizePoolBalance({ poolCode: 'lucky', balance: Number.NaN })

    expect(store.getBalance('lucky')).toBeNull()
  })
})
