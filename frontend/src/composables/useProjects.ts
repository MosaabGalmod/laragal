import { ref, computed } from 'vue'
import type { SiteItem } from '../types'
import { getProjectTypeLabel } from '../utils/projectType'
import { useI18n } from './useI18n'
import { useToast } from './useToast'
import { useRuntimeCatalog } from './useRuntimeCatalog'

export function useProjects() {
  const { t, renderUIMessage } = useI18n()
  const { showToast } = useToast()
  const { getPHPPort } = useRuntimeCatalog()

  const sites = ref<SiteItem[]>([])
  const activeFilter = ref<
    'web' | 'all' | 'laravel' | 'php' | 'go' | 'rust' | 'dotnet' | 'python' | 'node' | 'flutter' | 'non_web'
  >('web')
  const searchQuery = ref('')
  const isSyncingHosts = ref(false)

  const projectsDirectory = ref('C:/laragal/sites')
  const recentDirectories = ref<string[]>([])
  const runningProjects = ref<Record<string, boolean>>({})

  const normalizedProjectsDir = computed(() => {
    return (projectsDirectory.value || '').replace(/\\/g, '/')
  })

  const normalizedRecentDirs = computed(() => {
    const current = normalizedProjectsDir.value
    const list = recentDirectories.value
      .map((d) => (d || '').replace(/\\/g, '/'))
      .filter(Boolean)
    if (current && !list.includes(current)) {
      list.unshift(current)
    }
    return Array.from(new Set(list))
  })

  const suggestedPaths = computed(() => {
    const defaults = ['D:/projects', 'C:/laragal/sites', 'C:/projects']
    const set = new Set<string>()
    if (normalizedProjectsDir.value) set.add(normalizedProjectsDir.value)
    for (const p of normalizedRecentDirs.value) {
      if (p) set.add(p)
    }
    for (const d of defaults) {
      set.add(d)
    }
    return Array.from(set).slice(0, 5)
  })

  function mapSiteItem(item: any): SiteItem {
    const isWeb =
      item.is_web !== undefined
        ? Boolean(item.is_web)
        : item.IsWeb !== undefined
          ? Boolean(item.IsWeb)
          : true
    const projectType =
      item.project_type || item.ProjectType || (isWeb ? 'php' : 'non_web')
    const typeLabel = getProjectTypeLabel(
      projectType,
      item.type_label || item.TypeLabel,
      t,
    )
    const rawPhp = item.php_version || item.PHPVersion || ''
    const isPhpType =
      projectType === 'laravel' ||
      projectType === 'wordpress' ||
      projectType === 'php'
    const phpVersion = rawPhp
      ? rawPhp.startsWith('PHP')
        ? rawPhp
        : `PHP ${rawPhp}`
      : isWeb && isPhpType
        ? 'PHP 8.3'
        : ''

    return {
      name: item.name || item.Name || item.domain || item.Domain || '',
      domain: item.domain || item.Domain || '',
      path: item.path || item.Path || '',
      root_path: item.root_path || item.RootPath || '',
      php_version: phpVersion,
      php_port:
        item.php_port !== undefined
          ? item.php_port
          : item.PHPPort !== undefined
            ? item.PHPPort
            : isWeb && isPhpType
              ? getPHPPort(item.php_version)
              : 0,
      proxy_port:
        item.proxy_port || item.ProxyPort || (projectType === 'go' ? 8080 : 0),
      project_type: projectType,
      type_label: typeLabel,
      is_web: isWeb,
    }
  }

  const webSites = computed(() => sites.value.filter((s) => s.is_web))
  const webSitesCount = computed(() => webSites.value.length)
  const laravelSitesCount = computed(
    () => sites.value.filter((s) => s.project_type === 'laravel').length,
  )
  const phpSitesCount = computed(
    () => sites.value.filter((s) => s.project_type === 'php').length,
  )
  const goSitesCount = computed(
    () =>
      sites.value.filter(
        (s) => s.project_type === 'go' || s.project_type === 'proxy',
      ).length,
  )
  const rustSitesCount = computed(
    () => sites.value.filter((s) => s.project_type === 'rust').length,
  )
  const dotnetSitesCount = computed(
    () => sites.value.filter((s) => s.project_type === 'dotnet').length,
  )
  const pythonSitesCount = computed(
    () => sites.value.filter((s) => s.project_type === 'python').length,
  )
  const nodeSitesCount = computed(
    () => sites.value.filter((s) => s.project_type === 'node').length,
  )
  const flutterSitesCount = computed(
    () => sites.value.filter((s) => s.project_type === 'flutter').length,
  )
  const nonWebCount = computed(
    () => sites.value.filter((s) => !s.is_web).length,
  )

  const filterOptions = computed(() => {
    const items: {
      id: typeof activeFilter.value
      label: string
      count: number
    }[] = [{ id: 'web', label: t('filter_web'), count: webSitesCount.value }]
    if (laravelSitesCount.value > 0)
      items.push({
        id: 'laravel',
        label: 'Laravel',
        count: laravelSitesCount.value,
      })
    if (phpSitesCount.value > 0)
      items.push({ id: 'php', label: 'PHP', count: phpSitesCount.value })
    if (goSitesCount.value > 0)
      items.push({ id: 'go', label: 'Go', count: goSitesCount.value })
    if (rustSitesCount.value > 0)
      items.push({ id: 'rust', label: 'Rust', count: rustSitesCount.value })
    if (dotnetSitesCount.value > 0)
      items.push({ id: 'dotnet', label: '.NET', count: dotnetSitesCount.value })
    if (pythonSitesCount.value > 0)
      items.push({ id: 'python', label: 'Python', count: pythonSitesCount.value })
    if (nodeSitesCount.value > 0)
      items.push({ id: 'node', label: 'Node.js', count: nodeSitesCount.value })
    if (flutterSitesCount.value > 0)
      items.push({
        id: 'flutter',
        label: 'Flutter',
        count: flutterSitesCount.value,
      })
    if (nonWebCount.value > 0)
      items.push({
        id: 'non_web',
        label: t('filter_non_web'),
        count: nonWebCount.value,
      })
    items.push({ id: 'all', label: t('filter_all'), count: sites.value.length })
    return items
  })

  const filteredSites = computed(() => {
    let list = sites.value

    if (activeFilter.value === 'web') {
      list = list.filter((s) => s.is_web)
    } else if (activeFilter.value === 'laravel') {
      list = list.filter((s) => s.project_type === 'laravel')
    } else if (activeFilter.value === 'php') {
      list = list.filter((s) => s.project_type === 'php')
    } else if (activeFilter.value === 'go') {
      list = list.filter(
        (s) => s.project_type === 'go' || s.project_type === 'proxy',
      )
    } else if (activeFilter.value === 'rust') {
      list = list.filter((s) => s.project_type === 'rust')
    } else if (activeFilter.value === 'dotnet') {
      list = list.filter((s) => s.project_type === 'dotnet')
    } else if (activeFilter.value === 'python') {
      list = list.filter((s) => s.project_type === 'python')
    } else if (activeFilter.value === 'node') {
      list = list.filter((s) => s.project_type === 'node')
    } else if (activeFilter.value === 'flutter') {
      list = list.filter((s) => s.project_type === 'flutter')
    } else if (activeFilter.value === 'non_web') {
      list = list.filter((s) => !s.is_web)
    }

    const q = searchQuery.value.trim().toLowerCase()
    if (q) {
      list = list.filter((s) => {
        const name = (s.name || '').toLowerCase()
        const domain = (s.domain || '').toLowerCase()
        const path = (s.path || s.root_path || '').toLowerCase()
        return name.includes(q) || domain.includes(q) || path.includes(q)
      })
    }

    return list
  })

  function getDocRootRelative(site: SiteItem): string {
    if (!site.path || !site.root_path || site.path === site.root_path) {
      return ''
    }
    const normPath = site.path.replace(/\\/g, '/').replace(/\/$/, '')
    const normRoot = site.root_path.replace(/\\/g, '/').replace(/\/$/, '')
    if (normRoot.startsWith(normPath)) {
      const sub = normRoot.slice(normPath.length).replace(/^\//, '')
      return sub ? `/${sub}` : ''
    }
    return ''
  }

  function isPhpProject(site: SiteItem): boolean {
    return Boolean(
      site.is_web &&
        (site.project_type === 'laravel' ||
          site.project_type === 'wordpress' ||
          site.project_type === 'php' ||
          (site.php_port || 0) > 0),
    )
  }

  async function loadSites() {
    const w = window as any
    if (w.go?.main?.App?.GetSites) {
      try {
        let s = await w.go.main.App.GetSites()
        if (Array.isArray(s) && s.length === 0 && w.go?.main?.App?.ScanProjects) {
          s = await w.go.main.App.ScanProjects()
        }
        if (Array.isArray(s)) {
          sites.value = s.map(mapSiteItem)
          if (
            webSitesCount.value === 0 &&
            sites.value.length > 0 &&
            activeFilter.value === 'web'
          ) {
            activeFilter.value = 'all'
          } else if (webSitesCount.value > 0 && activeFilter.value === 'all') {
            activeFilter.value = 'web'
          }
        }
      } catch (e) {
        console.warn('Could not load sites:', e)
      }
    }
  }

  async function loadRecentDirectories() {
    const w = window as any
    if (w.go?.main?.App?.GetRecentDirectories) {
      try {
        const rec = await w.go.main.App.GetRecentDirectories()
        if (Array.isArray(rec)) {
          recentDirectories.value = rec.map((p: string) =>
            (p || '').replace(/\\/g, '/'),
          )
        }
      } catch (e) {
        console.warn('Could not load recent directories:', e)
      }
    }
  }

  async function switchRecentDirectory(event: Event) {
    const target = event.target as HTMLSelectElement
    const chosen = (target.value || '').replace(/\\/g, '/')
    if (!chosen || chosen === normalizedProjectsDir.value) return
    projectsDirectory.value = chosen
    const w = window as any
    if (w.go?.main?.App?.SetProjectsDirectory) {
      try {
        const s = await w.go.main.App.SetProjectsDirectory(chosen)
        showToast(`${t('projects_dir_updated')} ${chosen}`)
        if (Array.isArray(s)) {
          sites.value = s.map(mapSiteItem)
          if (webSitesCount.value === 0 && sites.value.length > 0) {
            activeFilter.value = 'all'
          } else {
            activeFilter.value = 'web'
          }
        }
        await loadRecentDirectories()
      } catch (e: any) {
        showToast(`${t('error_changing_dir')}: ${e?.message || e}`)
      }
    }
  }

  async function applyCustomDirectory(chosenPath: string) {
    const p = (chosenPath || '').trim()
    if (!p) return
    const norm = p.replace(/\\/g, '/')
    projectsDirectory.value = norm
    const w = window as any
    if (w.go?.main?.App?.SetProjectsDirectory) {
      try {
        const sitesList = await w.go.main.App.SetProjectsDirectory(norm)
        showToast(`${t('projects_dir_updated')} ${norm}`)
        if (Array.isArray(sitesList)) {
          sites.value = sitesList.map(mapSiteItem)
          if (
            webSitesCount.value === 0 &&
            sites.value.length > 0 &&
            activeFilter.value === 'web'
          ) {
            activeFilter.value = 'all'
          } else if (webSitesCount.value > 0 && activeFilter.value === 'all') {
            activeFilter.value = 'web'
          }
        }
        await loadRecentDirectories()
      } catch (e: any) {
        showToast(`${t('error_changing_dir')}: ${e?.message || e}`)
      }
    } else {
      showToast(`${t('projects_dir_updated')} ${norm}`)
    }
  }

  async function updateSitePHPVersion(site: SiteItem, event: Event) {
    const target = event.target as HTMLSelectElement
    const newVer = target.value
    const targetPath = site.path || site.root_path || site.domain
    if (!targetPath || !newVer) return

    const w = window as any
    if (w.go?.main?.App?.SetProjectPHPVersion) {
      try {
        const updated = await w.go.main.App.SetProjectPHPVersion(targetPath, newVer)
        site.php_version = `PHP ${newVer}`
        site.php_port = getPHPPort(newVer)
        showToast(`${t('php_version_changed')} ${site.domain} ➔ PHP ${newVer}`)
        if (Array.isArray(updated)) {
          sites.value = updated.map(mapSiteItem)
        }
      } catch (e: any) {
        showToast(`${t('error_changing_php')}: ${e?.message || e}`)
      }
    } else {
      site.php_version = `PHP ${newVer}`
      site.php_port = getPHPPort(newVer)
      showToast(`${t('php_version_changed')} ${site.domain} ➔ PHP ${newVer}`)
    }
  }

  async function toggleProjectApp(site: SiteItem) {
    const isRunning = runningProjects.value[site.domain] || false
    const pName = site.name || site.domain
    const w = window as any
    if (isRunning) {
      if (w.go?.main?.App?.StopProjectApp) {
        await w.go.main.App.StopProjectApp(pName)
      }
      runningProjects.value[site.domain] = false
      showToast(t('notif_app_stopped', { name: pName }))
    } else {
      showToast(t('notif_app_starting', { name: pName }))
      if (w.go?.main?.App?.StartProjectApp) {
        try {
          const ok = await w.go.main.App.StartProjectApp(
            site.path || site.root_path,
            site.proxy_port || 8080,
          )
          if (ok) {
            runningProjects.value[site.domain] = true
            showToast(
              t('notif_app_started', {
                name: pName,
                port: site.proxy_port || 8080,
              }),
            )
            return
          }
        } catch (err: any) {
          showToast(
            t('notif_app_start_failed', {
              name: pName,
              error: err?.message || err,
            }),
          )
          return
        }
      }
      runningProjects.value[site.domain] = true
    }
  }

  async function syncHostsNow() {
    isSyncingHosts.value = true
    const w = window as any
    try {
      if (w.go?.main?.App?.SyncHosts) {
        const msg = await w.go.main.App.SyncHosts()
        showToast(renderUIMessage(msg) || t('hosts_synced_success'))
      } else {
        showToast(t('hosts_synced_success'))
      }
    } catch (err: any) {
      showToast(t('error_syncing_hosts', { error: err?.message || err }))
    } finally {
      isSyncingHosts.value = false
    }
  }

  function handleProjectCreated(project: {
    name: string
    type: string
    php: string
  }) {
    const projectName = project.name.trim().toLowerCase()
    const domain = `${projectName}.test`
    const port = getPHPPort(project.php)
    const rootPath = `${projectsDirectory.value.replace(/\\/g, '/')}/${projectName}/public`

    const w = window as any
    if (w.go?.main?.App?.CreateProject) {
      loadSites()
    } else {
      sites.value.push(
        mapSiteItem({
          name: projectName,
          domain,
          root_path: rootPath,
          php_version: `PHP ${project.php}`,
          php_port: port,
          project_type: project.type,
          type_label: project.type === 'laravel' ? 'Laravel' : 'PHP',
          is_web: true,
        }),
      )
    }
  }

  return {
    sites,
    activeFilter,
    searchQuery,
    isSyncingHosts,
    projectsDirectory,
    recentDirectories,
    runningProjects,
    normalizedProjectsDir,
    normalizedRecentDirs,
    suggestedPaths,
    webSites,
    webSitesCount,
    nonWebCount,
    filterOptions,
    filteredSites,
    mapSiteItem,
    getDocRootRelative,
    isPhpProject,
    loadSites,
    loadRecentDirectories,
    switchRecentDirectory,
    applyCustomDirectory,
    updateSitePHPVersion,
    toggleProjectApp,
    syncHostsNow,
    handleProjectCreated,
  }
}
