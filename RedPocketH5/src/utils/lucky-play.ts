export type LuckyPlayType = 'thunder' | 'parity'
export type ParityChoice = 'odd' | 'even'
export type OddEvenGuess = 0 | 1

const PLAY_TYPE_ALIASES = new Map<string, LuckyPlayType>([
  ['thunder', 'thunder'],
  ['mine', 'thunder'],
  ['mines', 'thunder'],
  ['lei', 'thunder'],
  ['thunder_number', 'thunder'],
  ['bomb', 'thunder'],
  ['1', 'thunder'],
  ['parity', 'parity'],
  ['odd_even', 'parity'],
  ['odd-even', 'parity'],
  ['oddeven', 'parity'],
  ['jiou', 'parity'],
  ['evenodd', 'parity'],
  ['2', 'parity'],
])

const PARITY_CHOICE_ALIASES = new Map<string, ParityChoice>([
  ['odd', 'odd'],
  ['odds', 'odd'],
  ['1', 'odd'],
  ['dan', 'odd'],
  ['oddnumber', 'odd'],
  ['even', 'even'],
  ['0', 'even'],
  ['2', 'even'],
  ['shuang', 'even'],
  ['evennumber', 'even'],
])

function normalizeToken(value: unknown) {
  return String(value ?? '')
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '_')
}

export function resolveLuckyPlayType(payload: Record<string, any> | null | undefined, fallback: LuckyPlayType = 'thunder'): LuckyPlayType {
  if (!payload)
    return fallback

  const numericMode = Number(payload.game_mode ?? payload.gameMode)
  if (numericMode === 0)
    return 'thunder'
  if (numericMode === 1)
    return 'parity'

  const candidates = [
    payload.playType,
    payload.gameType,
    payload.ruleType,
    payload.mode,
    payload.playMode,
    payload.packetType,
  ]

  for (const candidate of candidates) {
    const normalized = normalizeToken(candidate)
    const matched = PLAY_TYPE_ALIASES.get(normalized)
    if (matched)
      return matched
  }

  return fallback
}

export function resolveParityChoice(value: unknown): ParityChoice | null {
  const normalized = normalizeToken(value)
  if (!normalized)
    return null
  if (normalized === '奇')
    return 'odd'
  if (normalized === '偶')
    return 'even'
  return PARITY_CHOICE_ALIASES.get(normalized) || null
}

export function resolveOddEvenGuess(value: unknown): OddEvenGuess | null {
  const choice = resolveParityChoice(value)
  if (choice === 'odd')
    return 1
  if (choice === 'even')
    return 0
  return null
}

export function resolveGameMode(playType: LuckyPlayType): 0 | 1 {
  return playType === 'parity' ? 1 : 0
}

export function isParityPlayType(playType: LuckyPlayType) {
  return playType === 'parity'
}
