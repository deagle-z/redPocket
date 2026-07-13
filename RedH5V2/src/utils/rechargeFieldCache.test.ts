import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { RechargeField } from '@/api/user'
import {
  buildRechargeFieldCacheKey,
  getCachedRechargeFieldValues,
  getRechargeFieldInitialValues,
  setCachedRechargeFieldValues,
} from './rechargeFieldCache'

function installWindowStorage() {
  const data = new Map<string, string>()
  const localStorage = {
    get length() {
      return data.size
    },
    clear: vi.fn(() => data.clear()),
    getItem: vi.fn((key: string) => data.get(key) ?? null),
    key: vi.fn((index: number) => Array.from(data.keys())[index] ?? null),
    removeItem: vi.fn((key: string) => data.delete(key)),
    setItem: vi.fn((key: string, value: string) => data.set(key, value)),
  }

  vi.stubGlobal('window', { localStorage })

  return localStorage
}

const fields: RechargeField[] = [
  {
    fieldKey: 'payerName',
    fieldLabel: 'Payer name',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 1,
    defaultValue: 'Default User',
  },
  {
    fieldKey: 'payerEmail',
    fieldLabel: 'Payer email',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 0,
  },
]

describe('recharge field cache', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    installWindowStorage()
  })

  it('scopes cached recharge fields by country and user', () => {
    setCachedRechargeFieldValues('US', 7, fields, {
      payerName: 'Ana',
      payerEmail: 'ana@example.com',
    })
    setCachedRechargeFieldValues('BR', 7, fields, {
      payerName: 'Bea',
      payerEmail: 'bea@example.com',
    })

    expect(getCachedRechargeFieldValues('US', 7)).toEqual({
      payerName: 'Ana',
      payerEmail: 'ana@example.com',
    })
    expect(getCachedRechargeFieldValues('BR', 7)).toEqual({
      payerName: 'Bea',
      payerEmail: 'bea@example.com',
    })
    expect(getCachedRechargeFieldValues('US', 8)).toBeNull()
  })

  it('hydrates current recharge fields from cache without keeping removed fields', () => {
    window.localStorage.setItem(
      buildRechargeFieldCacheKey('US', 7),
      JSON.stringify({
        value: {
          payerName: 'Ana',
          payerEmail: 'ana@example.com',
          staleField: 'old value',
        },
        expiresAt: null,
      }),
    )

    expect(getRechargeFieldInitialValues('US', 7, fields)).toEqual({
      payerName: 'Ana',
      payerEmail: 'ana@example.com',
    })
  })

  it('falls back to field defaults when there is no matching cached value', () => {
    expect(getRechargeFieldInitialValues('US', 7, fields)).toEqual({
      payerName: 'Default User',
      payerEmail: '',
    })
  })
})
