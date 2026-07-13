import { get, post } from '@/utils/http'

export interface VipLevelItem {
  id: number
  level: number
  levelName: string
  totalRechargeCount: number | null
  totalRechargeAmount: number | null
  totalWithdrawCount: number | null
  totalValidBet: number | null
  upgradeBonusAmount: number | null
  upgradeType: number
  status: number
  weeklySalary: number | null
  maxWithdrawAmount: number | null
  exclusiveService: number | boolean | null
  monthRechargeAmount: number | null
  monthValidBet: number | null
  keepLevelCondition: number | null
}

export interface VipProgressLevelRef {
  level: number
  levelName: string
  upgradeBonusAmount: number
}

export interface VipProgressData {
  currentLevel: VipProgressLevelRef | null
  prevLevel: VipProgressLevelRef | null
  nextLevel: VipProgressLevelRef | null
  levels?: VipProgressLevelRef[]
  /** Percentage (0-100) of progress toward the next level, computed by the backend. */
  progress: number
  currentValue: number
  targetValue: number
  nextBonusAmount: number
}

export function getVipLevels(): Promise<VipLevelItem[]> {
  return get<VipLevelItem[]>('/v1/app/vip/levels')
}

export function getVipProgress(): Promise<VipProgressData> {
  return get<VipProgressData>('/v1/app/vip/progress')
}

export interface VipWeeklySalaryStatus {
  level: number
  levelName: string
  weeklySalary: number
  claimable: boolean
  claimed: boolean
  nextResetAt: string | null
}

export function getVipWeeklySalary(): Promise<VipWeeklySalaryStatus> {
  return get<VipWeeklySalaryStatus>('/v1/app/vip/weeklySalary')
}

export function claimVipWeeklySalary(): Promise<string> {
  return post<string>('/v1/app/vip/weeklySalary/claim', undefined, {
    meta: { showError: false },
  })
}
