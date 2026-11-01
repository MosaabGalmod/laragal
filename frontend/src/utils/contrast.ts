/**
 * WCAG 2.x Relative Luminance and Contrast Ratio Utilities
 *
 * Implements the official W3C Web Content Accessibility Guidelines (WCAG) 2.1
 * relative luminance formula and contrast ratio calculation.
 * https://www.w3.org/WAI/GL/wiki/Relative_luminance
 * https://www.w3.org/TR/WCAG21/#contrast-minimum
 */

export interface RgbColor {
  r: number
  g: number
  b: number
}

export interface WcagCompliance {
  ratio: number
  formattedRatio: string
  aaNormal: boolean
  aaLarge: boolean
  aaaNormal: boolean
  aaaLarge: boolean
}

/**
 * Normalizes and parses a hex color string (3 or 6 hex digits, optional leading #).
 * Returns null if the string is not a valid hex color.
 */
export function parseHexColor(input: string): RgbColor | null {
  if (!input || typeof input !== 'string') {
    return null
  }

  const clean = input.trim().replace(/^#/, '')

  if (/^[0-9a-fA-F]{3}$/.test(clean)) {
    const r = parseInt(clean[0] + clean[0], 16)
    const g = parseInt(clean[1] + clean[1], 16)
    const b = parseInt(clean[2] + clean[2], 16)
    return { r, g, b }
  }

  if (/^[0-9a-fA-F]{6}$/.test(clean)) {
    const r = parseInt(clean.substring(0, 2), 16)
    const g = parseInt(clean.substring(2, 4), 16)
    const b = parseInt(clean.substring(4, 6), 16)
    return { r, g, b }
  }

  return null
}

/**
 * Computes the relative luminance of a color in sRGB space according to WCAG 2.x.
 * Returns a value between 0.0 (darkest black) and 1.0 (lightest white).
 */
export function getRelativeLuminance(color: string | RgbColor): number {
  const rgb = typeof color === 'string' ? parseHexColor(color) : color

  if (!rgb) {
    return 0
  }

  const channels = [rgb.r, rgb.g, rgb.b].map((c) => {
    const srgb = Math.max(0, Math.min(255, c)) / 255
    return srgb <= 0.03928 ? srgb / 12.92 : Math.pow((srgb + 0.055) / 1.055, 2.4)
  })

  return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2]
}

/**
 * Calculates the contrast ratio between two colors according to WCAG 2.x:
 * (L1 + 0.05) / (L2 + 0.05), where L1 is the lighter and L2 is the darker color.
 * Returns a ratio between 1.0 and 21.0.
 */
export function getContrastRatio(color1: string | RgbColor, color2: string | RgbColor): number {
  const l1 = getRelativeLuminance(color1)
  const l2 = getRelativeLuminance(color2)

  const lighter = Math.max(l1, l2)
  const darker = Math.min(l1, l2)

  return (lighter + 0.05) / (darker + 0.05)
}

/**
 * Evaluates WCAG 2.x compliance thresholds for normal and large text at AA and AAA levels:
 * - AA Normal text: contrast ratio >= 4.5:1
 * - AA Large text (>= 18pt or 14pt bold): contrast ratio >= 3.0:1
 * - AAA Normal text: contrast ratio >= 7.0:1
 * - AAA Large text (>= 18pt or 14pt bold): contrast ratio >= 4.5:1
 */
export function checkWcagCompliance(ratio: number): Omit<WcagCompliance, 'ratio' | 'formattedRatio'> {
  return {
    aaNormal: ratio >= 4.5,
    aaLarge: ratio >= 3.0,
    aaaNormal: ratio >= 7.0,
    aaaLarge: ratio >= 4.5,
  }
}

/**
 * Convenience helper that evaluates ratio and all compliance flags for two colors.
 */
export function getWcagDetails(foreground: string, background: string): WcagCompliance {
  const ratio = getContrastRatio(foreground, background)
  const compliance = checkWcagCompliance(ratio)

  return {
    ratio,
    formattedRatio: `${ratio.toFixed(2)}:1`,
    ...compliance,
  }
}
