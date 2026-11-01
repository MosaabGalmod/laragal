import { describe, it, expect } from 'vitest'
import ar from './ar.json'
import en from './en.json'

describe('locales bilingual completeness', () => {
  it('ar.json and en.json have exactly the same key set', () => {
    const arKeys = Object.keys(ar).sort()
    const enKeys = Object.keys(en).sort()

    expect(arKeys).toEqual(enKeys)
  })

  it('no values are empty in ar.json', () => {
    for (const [key, val] of Object.entries(ar)) {
      expect(typeof val === 'string' && val.trim().length > 0, `ar.json key "${key}" is empty`).toBe(true)
    }
  })

  it('no values are empty in en.json', () => {
    for (const [key, val] of Object.entries(en)) {
      expect(typeof val === 'string' && val.trim().length > 0, `en.json key "${key}" is empty`).toBe(true)
    }
  })
})
