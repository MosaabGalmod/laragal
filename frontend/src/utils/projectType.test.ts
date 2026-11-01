import { describe, it, expect } from 'vitest'
import { getProjectTypeLabel, isProperName, normalizeTypeKey } from './projectType'

describe('projectType utils', () => {
  const arLocales: Record<string, string> = {
    project_type_label_non_web: 'غير ويب / ملفات',
    project_type_label_static: 'ثابت',
    project_type_static: 'صفحات ثابتة (HTML / CSS / JS)',
    project_type_non_web: 'غير ويب / مجلد محلي فقط',
  }
  const tAr = (key: string) => arLocales[key] || key

  const enLocales: Record<string, string> = {
    project_type_label_non_web: 'Non-Web',
    project_type_label_static: 'Static',
    project_type_static: 'Static Pages (HTML / CSS / JS)',
    project_type_non_web: 'Non-Web / Local Folder Only',
  }
  const tEn = (key: string) => enLocales[key] || key

  describe('normalizeTypeKey', () => {
    it('normalizes spaces, dashes, slashes, and cases', () => {
      expect(normalizeTypeKey(' Non-Web ')).toBe('non_web')
      expect(normalizeTypeKey('static/html')).toBe('static_html')
      expect(normalizeTypeKey('HTML / Static')).toBe('html_static')
    })
  })

  describe('isProperName', () => {
    it('recognizes proper framework and language names', () => {
      expect(isProperName('Laravel')).toBe(true)
      expect(isProperName('laravel')).toBe(true)
      expect(isProperName('WordPress')).toBe(true)
      expect(isProperName('Go')).toBe(true)
      expect(isProperName('Rust')).toBe(true)
      expect(isProperName('Node.js')).toBe(true)
      expect(isProperName('.NET / C#')).toBe(true)
      expect(isProperName('Python')).toBe(true)
      expect(isProperName('Flutter Web')).toBe(true)
    })

    it('rejects general word descriptors', () => {
      expect(isProperName('Non-Web')).toBe(false)
      expect(isProperName('Static')).toBe(false)
      expect(isProperName('files')).toBe(false)
      expect(isProperName('')).toBe(false)
    })
  })

  describe('getProjectTypeLabel', () => {
    describe('word-based labels (translated via locale keys)', () => {
      it('translates Non-Web detector label in Arabic and English', () => {
        expect(getProjectTypeLabel('non_web', 'Non-Web', tAr)).toBe('غير ويب / ملفات')
        expect(getProjectTypeLabel('non_web', 'Non-Web', tEn)).toBe('Non-Web')
      })

      it('translates Static detector label in Arabic and English', () => {
        expect(getProjectTypeLabel('static', 'Static', tAr)).toBe('ثابت')
        expect(getProjectTypeLabel('static', 'Static', tEn)).toBe('Static')
      })

      it('translates when only projectType is passed with empty typeLabel', () => {
        expect(getProjectTypeLabel('non_web', '', tAr)).toBe('غير ويب / ملفات')
        expect(getProjectTypeLabel('static', '', tAr)).toBe('ثابت')
      })

      it('translates case-variations of non-web and static', () => {
        expect(getProjectTypeLabel(undefined, 'non-web', tAr)).toBe('غير ويب / ملفات')
        expect(getProjectTypeLabel(undefined, 'non_web', tAr)).toBe('غير ويب / ملفات')
        expect(getProjectTypeLabel(undefined, 'static', tAr)).toBe('ثابت')
        expect(getProjectTypeLabel(undefined, 'HTML / Static', tAr)).toBe('ثابت')
      })

      it('translates subservice labels with word prefixes like Static (docs)', () => {
        expect(getProjectTypeLabel('static', 'Static (docs)', tAr)).toBe('ثابت (docs)')
        expect(getProjectTypeLabel('static', 'Static (docs)', tEn)).toBe('Static (docs)')
        expect(getProjectTypeLabel('non_web', 'Non-Web (data)', tAr)).toBe('غير ويب / ملفات (data)')
      })

      it('works with fallback if only project_type_non_web is present in t', () => {
        const customT = (k: string) => {
          if (k === 'project_type_non_web') return 'غير ويب / ملفات'
          if (k === 'project_type_static') return 'ثابت'
          return k
        }
        expect(getProjectTypeLabel('non_web', 'Non-Web', customT)).toBe('غير ويب / ملفات')
        expect(getProjectTypeLabel('static', 'Static', customT)).toBe('ثابت')
      })
    })

    describe('proper names (kept as-is)', () => {
      it('preserves Laravel in both Arabic and English', () => {
        expect(getProjectTypeLabel('laravel', 'Laravel', tAr)).toBe('Laravel')
        expect(getProjectTypeLabel('laravel', 'Laravel', tEn)).toBe('Laravel')
      })

      it('preserves WordPress', () => {
        expect(getProjectTypeLabel('wordpress', 'WordPress', tAr)).toBe('WordPress')
      })

      it('preserves Go and Rust', () => {
        expect(getProjectTypeLabel('go', 'Go', tAr)).toBe('Go')
        expect(getProjectTypeLabel('rust', 'Rust', tAr)).toBe('Rust')
      })

      it('preserves Node.js, Python, and .NET / C#', () => {
        expect(getProjectTypeLabel('node', 'Node.js', tAr)).toBe('Node.js')
        expect(getProjectTypeLabel('python', 'Python', tAr)).toBe('Python')
        expect(getProjectTypeLabel('dotnet', '.NET / C#', tAr)).toBe('.NET / C#')
      })

      it('preserves Vue, React, and Flutter Web', () => {
        expect(getProjectTypeLabel('vue', 'Vue', tAr)).toBe('Vue')
        expect(getProjectTypeLabel('react', 'React', tAr)).toBe('React')
        expect(getProjectTypeLabel('flutter', 'Flutter Web', tAr)).toBe('Flutter Web')
      })

      it('preserves subservice proper names like Next.js (frontend)', () => {
        expect(getProjectTypeLabel('node', 'Next.js (frontend)', tAr)).toBe('Next.js (frontend)')
      })
    })
  })
})
