import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Tailwind color shade verification', () => {
  const VALID_SHADES = new Set([
    '50',
    '100',
    '200',
    '300',
    '400',
    '500',
    '600',
    '700',
    '800',
    '900',
    '950',
  ])

  const TAILWIND_COLORS = [
    'slate',
    'gray',
    'zinc',
    'neutral',
    'stone',
    'red',
    'orange',
    'amber',
    'yellow',
    'lime',
    'green',
    'emerald',
    'teal',
    'cyan',
    'sky',
    'blue',
    'indigo',
    'violet',
    'purple',
    'fuchsia',
    'pink',
    'rose',
  ]

  const colorRegex = new RegExp(
    `\\b(?:[a-z0-9:-]+-)?(?:${TAILWIND_COLORS.join('|')})-(\\d+)\\b`,
    'g',
  )

  function collectVueFilesRecursively(dir: string): string[] {
    const results: string[] = []
    if (!fs.existsSync(dir)) return results
    const entries = fs.readdirSync(dir, { withFileTypes: true })
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name)
      if (entry.isDirectory()) {
        results.push(...collectVueFilesRecursively(fullPath))
      } else if (entry.isFile() && entry.name.endsWith('.vue')) {
        results.push(fullPath)
      }
    }
    return results
  }

  it('ensures no .vue files contain slate-850 or invalid non-default Tailwind shades', () => {
    const srcDir = path.resolve(__dirname)
    const vueFiles = collectVueFilesRecursively(srcDir)
    expect(vueFiles.length).toBeGreaterThan(0)

    const invalidFindings: string[] = []

    for (const file of vueFiles) {
      const relativePath = path.relative(srcDir, file).replace(/\\/g, '/')
      const content = fs.readFileSync(file, 'utf-8')
      const lines = content.split('\n')

      lines.forEach((line, index) => {
        let match: RegExpExecArray | null
        const regex = new RegExp(colorRegex)
        while ((match = regex.exec(line)) !== null) {
          const shade = match[1]
          if (!VALID_SHADES.has(shade)) {
            invalidFindings.push(
              `${relativePath}:${index + 1} uses invalid class "${match[0]}" (shade "${shade}" does not exist in Tailwind)`,
            )
          }
        }
      })
    }

    expect(
      invalidFindings,
      `Found invalid Tailwind shades:\n${invalidFindings.join('\n')}`,
    ).toEqual([])
  })
})
