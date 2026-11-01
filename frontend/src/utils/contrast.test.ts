import { describe, it, expect } from 'vitest'
import {
  parseHexColor,
  getRelativeLuminance,
  getContrastRatio,
  checkWcagCompliance,
  getWcagDetails,
} from './contrast'

describe('contrast utils (WCAG 2.x)', () => {
  describe('parseHexColor', () => {
    it('parses 3-character hex colors with or without #', () => {
      expect(parseHexColor('#000')).toEqual({ r: 0, g: 0, b: 0 })
      expect(parseHexColor('fff')).toEqual({ r: 255, g: 255, b: 255 })
      expect(parseHexColor('#777')).toEqual({ r: 119, g: 119, b: 119 })
    })

    it('parses 6-character hex colors with or without #', () => {
      expect(parseHexColor('#000000')).toEqual({ r: 0, g: 0, b: 0 })
      expect(parseHexColor('ffffff')).toEqual({ r: 255, g: 255, b: 255 })
      expect(parseHexColor('#0f172a')).toEqual({ r: 15, g: 23, b: 42 })
    })

    it('returns null for invalid hex colors', () => {
      expect(parseHexColor('')).toBeNull()
      expect(parseHexColor('xyz')).toBeNull()
      expect(parseHexColor('#12345')).toBeNull()
      expect(parseHexColor('#1234567')).toBeNull()
    })
  })

  describe('getRelativeLuminance', () => {
    it('returns 0 for pure black', () => {
      expect(getRelativeLuminance('#000000')).toBe(0)
      expect(getRelativeLuminance('#000')).toBe(0)
    })

    it('returns 1 for pure white', () => {
      expect(getRelativeLuminance('#ffffff')).toBeCloseTo(1, 5)
      expect(getRelativeLuminance('#fff')).toBeCloseTo(1, 5)
    })
  })

  describe('getContrastRatio', () => {
    it('calculates 21:1 for black and white (#000/#fff)', () => {
      const ratio = getContrastRatio('#000', '#fff')
      expect(ratio).toBeCloseTo(21.0, 2)
    })

    it('calculates ~4.48 for #777 and #fff and fails AA normal', () => {
      const ratio = getContrastRatio('#777', '#fff')
      expect(ratio).toBeCloseTo(4.48, 2)
      expect(ratio).toBeLessThan(4.5)

      const compliance = checkWcagCompliance(ratio)
      expect(compliance.aaNormal).toBe(false)
      expect(compliance.aaLarge).toBe(true)
    })

    it('is order invariant (foreground vs background order does not matter)', () => {
      expect(getContrastRatio('#000', '#fff')).toBe(getContrastRatio('#fff', '#000'))
      expect(getContrastRatio('#777', '#fff')).toBe(getContrastRatio('#fff', '#777'))
      expect(getContrastRatio('#0f172a', '#e2e8f0')).toBe(getContrastRatio('#e2e8f0', '#0f172a'))
    })

    it('returns 1:1 for identical colors', () => {
      expect(getContrastRatio('#000', '#000')).toBeCloseTo(1.0, 4)
      expect(getContrastRatio('#fff', '#fff')).toBeCloseTo(1.0, 4)
      expect(getContrastRatio('#3b82f6', '#3b82f6')).toBeCloseTo(1.0, 4)
    })
  })

  describe('checkWcagCompliance and getWcagDetails', () => {
    it('correctly flags AA and AAA for #000 on #fff (21:1)', () => {
      const details = getWcagDetails('#000', '#fff')
      expect(details.ratio).toBeCloseTo(21.0, 2)
      expect(details.formattedRatio).toBe('21.00:1')
      expect(details.aaNormal).toBe(true)
      expect(details.aaLarge).toBe(true)
      expect(details.aaaNormal).toBe(true)
      expect(details.aaaLarge).toBe(true)
    })

    it('correctly flags thresholds for #777 on #fff (fails AA normal, passes AA large)', () => {
      const details = getWcagDetails('#777', '#fff')
      expect(details.ratio).toBeCloseTo(4.48, 2)
      expect(details.aaNormal).toBe(false)
      expect(details.aaLarge).toBe(true)
      expect(details.aaaNormal).toBe(false)
      expect(details.aaaLarge).toBe(false)
    })

    it('correctly handles AAA threshold (7:1 for normal, 4.5:1 for large)', () => {
      // 5:1 ratio: passes AA normal, AA large, AAA large; fails AAA normal
      const comp5 = checkWcagCompliance(5.0)
      expect(comp5.aaNormal).toBe(true)
      expect(comp5.aaLarge).toBe(true)
      expect(comp5.aaaLarge).toBe(true)
      expect(comp5.aaaNormal).toBe(false)

      // 7.5:1 ratio: passes all
      const comp75 = checkWcagCompliance(7.5)
      expect(comp75.aaNormal).toBe(true)
      expect(comp75.aaLarge).toBe(true)
      expect(comp75.aaaLarge).toBe(true)
      expect(comp75.aaaNormal).toBe(true)
    })
  })
})
