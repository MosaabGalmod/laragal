import { ref, computed } from 'vue'
import { useI18n } from './useI18n'
import { useToast } from './useToast'

export interface RuntimeArtifact {
  name: string
  version?: string
  url?: string
  sha256?: string
  eol?: boolean
  eol_date?: string
  php_port?: number
}

export interface RuntimeUpdateInfo {
  name: string
  current_version: string
  latest_version: string
  has_update: boolean
  is_security_fix: boolean
  download_url: string
  is_new?: boolean
  major_change?: boolean
  compat_note?: string
  source_status?: 'live' | 'catalog_fallback'
  source_error?: string
}

export interface RuntimeDefinition {
  name: string
  label: string
  desc_key: string
  is_feed_only?: boolean
}

const runtimeCatalog = ref<RuntimeArtifact[]>([])
const runtimeUpdates = ref<Record<string, RuntimeUpdateInfo>>({})
const isCheckingUpdates = ref(false)
const downloadProgress = ref<Record<string, number>>({})
const downloadingRuntime = ref('')

const availableRuntimes = ref<RuntimeDefinition[]>([
  { name: 'caddy', label: 'Caddy Server', desc_key: 'service_caddy_desc' },
  { name: 'php-8.5', label: 'PHP 8.5 (NTS x64)', desc_key: 'service_php85_desc' },
  { name: 'php-8.4', label: 'PHP 8.4 (NTS x64)', desc_key: 'service_php84_desc' },
  { name: 'php-8.3', label: 'PHP 8.3 (NTS x64)', desc_key: 'service_php83_desc' },
  { name: 'php-8.2', label: 'PHP 8.2 (NTS x64)', desc_key: 'service_php82_desc' },
  { name: 'php-8.1', label: 'PHP 8.1 (NTS x64)', desc_key: 'service_php81_desc' },
  { name: 'mysql', label: 'MariaDB / MySQL Server', desc_key: 'service_mysql_desc' },
  { name: 'redis', label: 'Redis Server (Windows)', desc_key: 'service_redis_desc' },
  { name: 'mailpit', label: 'Mailpit Server', desc_key: 'service_mailpit_desc' },
])

export function compareVersions(v1: string, v2: string): number {
  const p1 = v1.split('.').map((n) => parseInt(n, 10) || 0)
  const p2 = v2.split('.').map((n) => parseInt(n, 10) || 0)
  for (let i = 0; i < Math.max(p1.length, p2.length); i++) {
    const a = p1[i] || 0
    const b = p2[i] || 0
    if (a > b) return 1
    if (a < b) return -1
  }
  return 0
}

export function extractPhpVersionNumber(ver?: string): string {
  if (!ver) return '8.3'
  const clean = ver.replace(/PHP\s*/i, '').trim()
  const match = clean.match(/^(\d+\.\d+)/)
  if (match) return match[1]
  return clean || '8.3'
}

export function getMajorMinor(v?: string): string {
  if (!v) return ''
  const clean = v.replace(/^PHP\s*/i, '').replace(/^[vV]/, '').trim()
  const parts = clean.split('.')
  if (parts.length >= 2) {
    return `${parts[0]}.${parts[1]}`
  }
  return parts[0] || ''
}

export function useRuntimeCatalog() {
  const { t, renderUIMessage } = useI18n()
  const { showToast } = useToast()

  async function loadRuntimeCatalog() {
    const w = window as any
    if (w.go?.main?.App?.GetRuntimeCatalog) {
      try {
        const list = await w.go.main.App.GetRuntimeCatalog()
        if (Array.isArray(list) && list.length > 0) {
          runtimeCatalog.value = list
        }
      } catch (e) {
        console.warn('Could not load runtime catalog:', e)
      }
    }
  }

  function getCatalogItem(name: string): RuntimeArtifact | undefined {
    return runtimeCatalog.value.find(
      (r) =>
        r.name === name ||
        r.name === `php-${name}` ||
        (name.startsWith('php-') && r.name === name),
    )
  }

  function getPHPPort(version?: string): number {
    if (!version) return 9083
    const clean = extractPhpVersionNumber(version)
    const exact = runtimeCatalog.value.find(
      (r) => r.name === `php-${clean}` || r.name === clean,
    )
    if (exact && exact.php_port) return exact.php_port
    const parts = clean.split('.')
    if (parts.length === 2) {
      const maj = parseInt(parts[0], 10)
      const min = parseInt(parts[1], 10)
      if (!isNaN(maj) && !isNaN(min)) {
        return 9000 + maj * 10 + min
      }
    }
    return 9083
  }

  const phpCatalog = computed(() => {
    const list = runtimeCatalog.value.filter(
      (r) => r.name.startsWith('php-') && (r.php_port || 0) > 0,
    )
    return [...list].sort((a, b) => {
      const va = a.name.replace(/^php-/, '')
      const vb = b.name.replace(/^php-/, '')
      return compareVersions(vb, va)
    })
  })

  const latestPhpVersion = computed(() => {
    let highest = ''
    for (const item of phpCatalog.value) {
      const ver = item.name.replace(/^php-/, '')
      if (!highest || compareVersions(ver, highest) > 0) {
        highest = ver
      }
    }
    return highest || '8.5'
  })

  const phpSelectOptions = computed(() => {
    return phpCatalog.value.map((item) => {
      const ver = item.name.replace(/^php-/, '')
      let suffix = ''
      if (ver === latestPhpVersion.value) {
        suffix = ` (${t('php_latest_label')})`
      } else if (ver === '8.3') {
        suffix = ` (${t('php_recommended_badge')})`
      } else if (item.eol) {
        suffix = ` (${t('runtime_eol_badge')})`
      } else if (item.eol_date) {
        suffix = ` (${t('runtime_security_only', { date: item.eol_date })})`
      }
      return {
        version: ver,
        label: `PHP ${ver}${suffix}`,
        eol: Boolean(item.eol),
        eol_date: item.eol_date,
        php_port: item.php_port,
      }
    })
  })

  const phpSelectOptionsPrefixed = computed(() => {
    return phpCatalog.value.map((item) => {
      const ver = item.name.replace(/^php-/, '')
      let suffix = ''
      if (ver === latestPhpVersion.value) {
        suffix = ` (${t('php_latest_label')})`
      } else if (ver === '8.3') {
        suffix = ` (${t('php_recommended_badge')})`
      } else if (item.eol) {
        suffix = ` (${t('runtime_eol_badge')})`
      } else if (item.eol_date) {
        suffix = ` (${t('runtime_security_only', { date: item.eol_date })})`
      }
      return {
        value: `PHP ${ver}`,
        version: ver,
        label: `PHP ${ver}${suffix}`,
        eol: Boolean(item.eol),
        eol_date: item.eol_date,
        php_port: item.php_port,
      }
    })
  })

  async function checkUpdatesNow(options?: { silent?: boolean }) {
    const silent = Boolean(options?.silent)
    isCheckingUpdates.value = true
    const w = window as any
    if (w.go?.main?.App?.CheckRuntimeUpdates) {
      try {
        const list = await w.go.main.App.CheckRuntimeUpdates()
        if (Array.isArray(list)) {
          availableRuntimes.value = availableRuntimes.value.filter((r) => !r.is_feed_only)
          const map: Record<string, RuntimeUpdateInfo> = {}
          let count = 0
          for (const item of list) {
            if (item && item.name) {
              map[item.name] = item
              if (item.has_update) count++

              // Feed-only entries: add to available runtimes dynamically
              if (item.is_new && !availableRuntimes.value.some((r) => r.name === item.name)) {
                let label = item.name
                let desc_key = 'service_mysql_desc'
                if (item.name.startsWith('php-')) {
                  const ver = item.name.replace(/^php-/, '')
                  label = `PHP ${ver} (NTS x64)`
                  desc_key = 'service_php85_desc'
                } else if (item.name.startsWith('mariadb-') || item.name.startsWith('mysql-')) {
                  const ver = item.name.replace(/^(mariadb|mysql)-/, '')
                  label = `MariaDB ${ver} (LTS)`
                  desc_key = 'service_mysql_desc'
                }
                availableRuntimes.value.push({
                  name: item.name,
                  label,
                  desc_key,
                  is_feed_only: true,
                })
              }
            }
          }
          runtimeUpdates.value = map
          if (!silent) {
            if (count > 0) {
              showToast(t('updates_found_count', { count }))
            } else {
              showToast(t('all_runtimes_up_to_date'))
            }
          }
        }
      } catch (err: any) {
        console.warn('Check updates error:', err)
      }
    }
    isCheckingUpdates.value = false
  }

  async function triggerDownloadRuntime(name: string, confirmed: boolean = false) {
    downloadingRuntime.value = name
    downloadProgress.value[name] = 0
    showToast(t('runtime_download_starting', { name }))

    const w = window as any
    if (confirmed && w.go?.main?.App?.UpgradeRuntimeConfirmed) {
      try {
        const msg = await w.go.main.App.UpgradeRuntimeConfirmed(name)
        downloadingRuntime.value = ''
        downloadProgress.value[name] = 100
        await loadRuntimeCatalog()
        await checkUpdatesNow({ silent: true })
        showToast(
          renderUIMessage(msg) || t('runtime_upgrade_success', { name }),
        )
        return
      } catch (err: any) {
        console.warn('Upgrade confirmed err:', err)
        downloadingRuntime.value = ''
        showToast(
          t('runtime_upgrade_failed', { name, error: err?.message || err }),
        )
        return
      }
    }

    if (w.go?.main?.App?.UpgradeRuntime) {
      try {
        const msg = await w.go.main.App.UpgradeRuntime(name)
        downloadingRuntime.value = ''
        downloadProgress.value[name] = 100
        await loadRuntimeCatalog()
        await checkUpdatesNow({ silent: true })
        showToast(
          renderUIMessage(msg) || t('runtime_upgrade_success', { name }),
        )
        return
      } catch (err: any) {
        console.warn('Upgrade err:', err)
        downloadingRuntime.value = ''
        showToast(
          t('runtime_upgrade_failed', { name, error: err?.message || err }),
        )
        return
      }
    }

    if (w.go?.main?.App?.DownloadBinary) {
      try {
        const msg = await w.go.main.App.DownloadBinary(name)
        downloadingRuntime.value = ''
        downloadProgress.value[name] = 100
        await loadRuntimeCatalog()
        await checkUpdatesNow({ silent: true })
        showToast(
          renderUIMessage(msg) || t('runtime_install_success', { name }),
        )
        return
      } catch (err: any) {
        console.warn('Download err:', err)
        downloadingRuntime.value = ''
        showToast(
          t('runtime_download_failed', { name, error: err?.message || err }),
        )
        return
      }
    }

    const interval = setInterval(() => {
      if ((downloadProgress.value[name] || 0) < 95) {
        downloadProgress.value[name] = (downloadProgress.value[name] || 0) + 15
      }
    }, 200)

    setTimeout(() => {
      clearInterval(interval)
      downloadProgress.value[name] = 100
      downloadingRuntime.value = ''
      showToast(t('runtime_ready_in_dir', { name }))
    }, 1500)
  }

  return {
    runtimeCatalog,
    availableRuntimes,
    runtimeUpdates,
    isCheckingUpdates,
    downloadProgress,
    downloadingRuntime,
    loadRuntimeCatalog,
    getCatalogItem,
    getPHPPort,
    extractPhpVersionNumber,
    compareVersions,
    phpCatalog,
    latestPhpVersion,
    phpSelectOptions,
    phpSelectOptionsPrefixed,
    checkUpdatesNow,
    triggerDownloadRuntime,
  }
}
