import type { RechargeField } from '@/api/user'
import { getTtlStorage, setTtlStorage } from './storage'

const RECHARGE_FIELD_CACHE_PREFIX = 'ppmx_recharge_fields'
const RECHARGE_FIELD_CACHE_TTL_MS = 30 * 24 * 60 * 60 * 1000

type RechargeFieldValues = Record<string, string>

function normalizeCountryCode(countryCode: string) {
  return countryCode.trim().toUpperCase() || 'UNKNOWN'
}

function normalizeUserId(userId?: number | string | null) {
  return String(userId ?? 'anonymous').trim() || 'anonymous'
}

function normalizeValue(value: unknown) {
  if (value == null) return ''

  return String(value)
}

function pickCurrentFieldValues(fields: RechargeField[], values: Record<string, unknown>) {
  return fields.reduce<RechargeFieldValues>((picked, field) => {
    if (!field.fieldKey) return picked
    if (Object.prototype.hasOwnProperty.call(values, field.fieldKey)) {
      picked[field.fieldKey] = normalizeValue(values[field.fieldKey])
    }

    return picked
  }, {})
}

export function buildRechargeFieldCacheKey(countryCode: string, userId?: number | string | null) {
  return `${RECHARGE_FIELD_CACHE_PREFIX}:${normalizeCountryCode(countryCode)}:${normalizeUserId(userId)}`
}

export function getCachedRechargeFieldValues(countryCode: string, userId?: number | string | null) {
  const cached = getTtlStorage<unknown>(buildRechargeFieldCacheKey(countryCode, userId))
  if (!cached || typeof cached !== 'object' || Array.isArray(cached)) return null

  return Object.entries(cached as Record<string, unknown>).reduce<RechargeFieldValues>((values, [key, value]) => {
    if (key) values[key] = normalizeValue(value)

    return values
  }, {})
}

export function getRechargeFieldInitialValues(
  countryCode: string,
  userId: number | string | null | undefined,
  fields: RechargeField[],
) {
  const cached = getCachedRechargeFieldValues(countryCode, userId) ?? {}

  return fields.reduce<RechargeFieldValues>((values, field) => {
    values[field.fieldKey] = Object.prototype.hasOwnProperty.call(cached, field.fieldKey)
      ? cached[field.fieldKey]
      : normalizeValue(field.defaultValue)

    return values
  }, {})
}

export function setCachedRechargeFieldValues(
  countryCode: string,
  userId: number | string | null | undefined,
  fields: RechargeField[],
  values: Record<string, unknown>,
) {
  setTtlStorage(
    buildRechargeFieldCacheKey(countryCode, userId),
    pickCurrentFieldValues(fields, values),
    RECHARGE_FIELD_CACHE_TTL_MS,
  )
}
