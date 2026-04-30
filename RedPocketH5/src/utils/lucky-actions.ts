import { truncate2 } from './currency'

type TranslateFn = (key: string, params?: Record<string, unknown>) => string

export interface LuckyPacketAction {
  seqNo: number
  isGrabbed?: boolean
  isGrabMine?: boolean
  amount?: number
  thunder?: number | boolean
  displayLoading?: boolean
}

export function formatLuckyActionAmount(amount: number) {
  return truncate2(amount).toFixed(2)
}

export function formatLuckyActionLabel(t: TranslateFn, action: LuckyPacketAction, isOngoing: boolean) {
  const isGrabbed = Boolean(action.isGrabbed)
  const amount = Number(action.amount || 0)
  const seqNo = Number(action.seqNo || 0)

  if (action.displayLoading)
    return t('homeLucky.loadingLabel')
  if (!isGrabbed && isOngoing)
    return t('homeLucky.grabAction', { seq: seqNo })
  if (isGrabbed && isOngoing && amount <= 0)
    return t('homeLucky.loadingLabel')
  if (!isGrabbed && !isOngoing)
    return amount > 0 ? formatLuckyActionAmount(amount) : '—'
  return formatLuckyActionAmount(amount)
}

export function isLuckyActionThunder(action: LuckyPacketAction) {
  return Number(action.thunder || 0) === 1
}
