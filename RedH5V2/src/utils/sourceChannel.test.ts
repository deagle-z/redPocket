import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  captureInviteCode,
  captureSourceChannelCode,
  getSourceChannelCode,
  getStoredInviteCode,
} from './sourceChannel'

function installStorage() {
  const data = new Map<string, string>()
  const storage = {
    get length() {
      return data.size
    },
    clear: vi.fn(() => data.clear()),
    getItem: vi.fn((key: string) => data.get(key) ?? null),
    key: vi.fn((index: number) => Array.from(data.keys())[index] ?? null),
    removeItem: vi.fn((key: string) => data.delete(key)),
    setItem: vi.fn((key: string, value: string) => data.set(key, value)),
  }

  vi.stubGlobal('localStorage', storage)

  return storage
}

describe('invite and source channel utilities', () => {
  beforeEach(() => {
    vi.unstubAllGlobals()
    installStorage()
  })

  it('captures invite code from supported query aliases', () => {
    expect(captureInviteCode({ c: 'ABC123' })).toBe('ABC123')
    expect(getStoredInviteCode()).toBe('ABC123')
    expect(captureInviteCode({ inviteCode: 'INVITE02' })).toBe('INVITE02')
    expect(getStoredInviteCode()).toBe('INVITE02')
  })

  it('captures source channel from supported query aliases', () => {
    expect(captureSourceChannelCode({ sc: 'fb01' })).toBe('FB01')
    expect(getSourceChannelCode()).toBe('FB01')
  })
})
