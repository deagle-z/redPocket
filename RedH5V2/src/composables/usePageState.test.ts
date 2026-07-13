import { describe, expect, it } from 'vitest'
import { usePageState } from './usePageState'

describe('usePageState', () => {
  it('tracks loading, empty, and error state transitions', () => {
    const state = usePageState()

    state.startLoading()
    expect(state.status.value).toBe('loading')

    state.setEmpty('暂无记录')
    expect(state.status.value).toBe('empty')
    expect(state.message.value).toBe('暂无记录')

    state.setError(new Error('网络异常'), 'network')
    expect(state.status.value).toBe('network')
    expect(state.message.value).toBe('网络异常')

    state.setReady()
    expect(state.status.value).toBe('ready')
  })
})
