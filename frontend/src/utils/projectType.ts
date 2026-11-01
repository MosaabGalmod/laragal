/**
 * Known proper names that must remain as-is and never translated.
 */
export const PROPER_NAMES: string[] = [
  'Laravel',
  'WordPress',
  'Go',
  'Rust',
  'Node.js',
  'Node',
  'Python',
  '.NET',
  '.NET / C#',
  'Vue',
  'React',
  'Ruby',
  'Ruby on Rails',
  'Java',
  'Java / Spring',
  'Flutter',
  'Flutter Web',
  'Next.js',
  'Next',
  'Nuxt',
  'Vite',
  'FastAPI',
  'Flask',
  'Django',
  'Express',
  'PHP',
  'Streamlit',
]

/**
 * Normalizes input string by trimming and converting to lowercase snake_case.
 */
export function normalizeTypeKey(val: string): string {
  return val
    .trim()
    .toLowerCase()
    .replace(/[\s\-_/]+/g, '_')
    .replace(/^_+|_+$/g, '')
}

/**
 * Checks whether the given string represents a known proper name.
 */
export function isProperName(val: string): boolean {
  const trimmed = val.trim()
  if (!trimmed) return false

  const lower = trimmed.toLowerCase()

  // Exact match against known proper names (case-insensitive)
  for (const name of PROPER_NAMES) {
    if (lower === name.toLowerCase()) {
      return true
    }
  }

  // Check if string starts with a proper name followed by details, e.g. "Next.js (frontend)" or ".NET / C#"
  for (const name of PROPER_NAMES) {
    const nameLower = name.toLowerCase()
    if (
      lower.startsWith(nameLower + ' ') ||
      lower.startsWith(nameLower + '(') ||
      lower.startsWith(nameLower + '/')
    ) {
      return true
    }
  }

  return false
}

/**
 * Resolves a translated word label using project_type_label_* or project_type_*.
 */
function resolveWordLabel(key: string, t: (k: string) => string): string {
  const labelKey = `project_type_label_${key}`
  const labelVal = t(labelKey)
  if (labelVal && labelVal !== labelKey) {
    return labelVal
  }
  const typeKey = `project_type_${key}`
  const typeVal = t(typeKey)
  if (typeVal && typeVal !== typeKey) {
    return typeVal
  }
  return ''
}

/**
 * Maps a project type and/or label to a localized display string.
 * Proper names (Laravel, WordPress, Go, Rust, Node.js, Python, .NET, Vue, React...) stay as-is.
 * Word labels / descriptors (non-web, static, etc.) are rendered via locale keys (t()).
 */
export function getProjectTypeLabel(
  projectType: string | undefined,
  typeLabel: string | undefined,
  t: (key: string) => string
): string {
  const label = (typeLabel || '').trim()
  const pType = (projectType || '').trim()

  const normLabel = normalizeTypeKey(label)
  const normType = normalizeTypeKey(pType)

  // Subservice detection with word prefixes, e.g. "Static (docs)" or "Non-Web (data)"
  const subMatch = label.match(/^(Non[-_ ]?Web|Static)\s*\((.+)\)$/i)
  if (subMatch) {
    const isStatic = subMatch[1].toLowerCase().includes('static')
    const base = isStatic
      ? (resolveWordLabel('static', t) || t('type_static_badge') || 'Static')
      : (resolveWordLabel('non_web', t) || t('non_web_badge') || 'Non-Web')
    return `${base} (${subMatch[2]})`
  }

  // 1. Check for non-web projects
  if (
    normLabel === 'non_web' ||
    normLabel === 'nonweb' ||
    normLabel === 'files' ||
    normType === 'non_web'
  ) {
    return resolveWordLabel('non_web', t) || t('non_web_badge') || 'Non-Web'
  }

  // 2. Check for static HTML projects
  if (
    normLabel === 'static' ||
    normLabel === 'html_static' ||
    normLabel === 'static_html' ||
    normType === 'static'
  ) {
    return resolveWordLabel('static', t) || t('type_static_badge') || 'Static'
  }

  // 3. If typeLabel is provided and matches a proper name, keep it as-is
  if (label && isProperName(label)) {
    return label
  }

  // 4. If projectType is provided and matches a proper name, use canonical proper name
  if (pType && isProperName(pType)) {
    for (const name of PROPER_NAMES) {
      if (pType.toLowerCase() === name.toLowerCase()) {
        return name
      }
    }
    return pType
  }

  // 5. If label is a descriptive word (not proper name), try translating
  if (normLabel) {
    const translated = resolveWordLabel(normLabel, t)
    if (translated) {
      return translated
    }
  }

  // 6. Try translating projectType via word label
  if (normType) {
    const translated = resolveWordLabel(normType, t)
    if (translated) {
      return translated
    }
  }

  // 7. Fallback to label or projectType
  return label || pType || ''
}
