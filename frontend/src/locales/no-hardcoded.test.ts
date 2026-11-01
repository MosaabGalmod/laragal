import { describe, it, expect } from 'vitest'
import fs from 'node:fs'
import path from 'node:path'

describe('no hardcoded Arabic UI text in frontend components', () => {
  const arabicRegex = /[\u0600-\u06FF]/

  function isAllowedLanguageSwitcherException(file: string, line: string): boolean {
    if (!file.endsWith('App.vue')) {
      return false
    }

    // The only allowed exception is the language-switcher label showing the other language's own name ("English" / "العربية")
    const trimmed = line.trim()
    if (trimmed === '<span v-else lang="ar">العربية</span>') {
      return true
    }
    if (
      trimmed.includes("currentLang === 'ar'") &&
      trimmed.includes("'العربية'") &&
      (trimmed.includes(':title=') || trimmed.includes(':aria-label='))
    ) {
      return true
    }

    return false
  }

  it('App.vue and components/*.vue contain zero hardcoded Arabic outside the language-switcher exception', () => {
    const srcDir = path.resolve(__dirname, '..')
    const appVue = path.join(srcDir, 'App.vue')
    const componentsDir = path.join(srcDir, 'components')

    const vueFiles: string[] = []
    if (fs.existsSync(appVue)) {
      vueFiles.push(appVue)
    }

    if (fs.existsSync(componentsDir)) {
      const entries = fs.readdirSync(componentsDir)
      for (const entry of entries) {
        if (entry.endsWith('.vue')) {
          vueFiles.push(path.join(componentsDir, entry))
        }
      }
    }

    expect(vueFiles.length).toBeGreaterThan(0)

    const violations: { file: string; line: number; text: string }[] = []

    for (const file of vueFiles) {
      const content = fs.readFileSync(file, 'utf8')
      const lines = content.split('\n')

      lines.forEach((lineText, idx) => {
        if (arabicRegex.test(lineText)) {
          if (!isAllowedLanguageSwitcherException(file, lineText)) {
            violations.push({
              file: path.relative(srcDir, file),
              line: idx + 1,
              text: lineText.trim()
            })
          }
        }
      })
    }

    expect(violations, `Hardcoded Arabic strings found outside allowed language-switcher exception: ${JSON.stringify(violations, null, 2)}`).toEqual([])
  })
})
