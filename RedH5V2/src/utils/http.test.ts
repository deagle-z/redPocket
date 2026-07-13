import { describe, expect, it } from 'vitest'

import { getErrorPayloadMessage } from './http'

describe('http error payload message', () => {
  it('uses backend message from failed business envelope even when code is 500', () => {
    expect(getErrorPayloadMessage({
      message: 'phone number is already registered',
      data: null,
      code: 500,
      success: false,
    })).toBe('phone number is already registered')
  })

  it('supports common backend error message fields', () => {
    expect(getErrorPayloadMessage({ msg: '业务错误' })).toBe('业务错误')
    expect(getErrorPayloadMessage({ error: 'invalid request' })).toBe('invalid request')
  })

  it('parses stringified backend error envelope', () => {
    expect(getErrorPayloadMessage(JSON.stringify({
      message: 'phone number is already registered',
      data: null,
      code: 500,
      success: false,
    }))).toBe('phone number is already registered')
  })
})
