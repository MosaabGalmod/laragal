import { ref } from 'vue'
import type { ServiceItem } from '../types'
import { defaultServices } from '../utils/defaultServices'
import { useI18n } from './useI18n'
import { useToast } from './useToast'

export function useServices() {
  const { t, renderUIMessage } = useI18n()
  const { showToast } = useToast()

  const services = ref<ServiceItem[]>(defaultServices.map((s) => ({ ...s })))
  const togglingServices = ref<Set<string>>(new Set())
  const confirmService = ref<ServiceItem | null>(null)

  function openBrowser(url: string) {
    const app = (window as any).go?.main?.App
    if (app?.OpenBrowser) {
      app.OpenBrowser(url)
    } else {
      window.open(url, '_blank')
    }
  }

  function openAdminer(dbName?: string) {
    const app = (window as any).go?.main?.App
    if (app?.OpenAdminer) {
      app.OpenAdminer(dbName || '').catch((err: any) => {
        showToast(t('operation_failed', { error: err?.message || String(err) }))
      })
    } else {
      const dbParam = dbName ? `&db=${encodeURIComponent(dbName)}` : ''
      openBrowser(`http://localhost:3307/?server=127.0.0.1&username=root${dbParam}`)
    }
  }

  function openPhpMyAdmin(dbName?: string) {
    const app = (window as any).go?.main?.App
    if (app?.OpenPhpMyAdmin) {
      app.OpenPhpMyAdmin(dbName || '').catch((err: any) => {
        showToast(t('operation_failed', { error: err?.message || String(err) }))
      })
    } else {
      const dbParam = dbName ? `index.php?route=/database/structure&db=${encodeURIComponent(dbName)}` : ''
      openBrowser(`http://localhost:3308/${dbParam}`)
    }
  }

  async function loadServices() {
    const w = window as any
    if (w.go?.main?.App?.GetServices) {
      try {
        const live = await w.go.main.App.GetServices()
        if (Array.isArray(live)) {
          for (const ls of live) {
            const match = services.value.find((s) => s.id === ls.name)
            if (match) {
              match.running = Boolean(ls.running)
            }
          }
        }
      } catch (e) {
        console.warn('Could not load services:', e)
      }
    }
  }

  async function checkPortConflicts() {
    const w = window as any
    if (w.go?.main?.App?.CheckPortConflict) {
      try {
        const msg = await w.go.main.App.CheckPortConflict(443)
        const rendered = renderUIMessage(msg)
        if (rendered) {
          showToast(rendered)
        }
      } catch (e) {
        console.warn('CheckPortConflict err:', e)
      }
    }
  }

  async function promptToggleService(service: ServiceItem) {
    if (togglingServices.value.has(service.id)) return
    togglingServices.value.add(service.id)
    togglingServices.value = new Set(togglingServices.value)

    const w = window as any
    try {
      if (w.go?.main?.App?.ToggleService) {
        const originalState = service.running
        try {
          const running = await w.go.main.App.ToggleService(service.id)
          service.running = Boolean(running)
          if (running) {
            showToast(t('service_started_msg', { service: t(service.name_key) }))
          } else {
            showToast(t('service_stopped_msg', { service: t(service.name_key) }))
          }
          await loadServices()
        } catch (err: any) {
          service.running = originalState
          showToast(
            t('service_toggle_error', {
              service: t(service.name_key),
              error: err?.message || err,
            }),
          )
          await loadServices()
        }
      } else {
        service.running = !service.running
        if (service.running) {
          showToast(t('service_started_msg', { service: t(service.name_key) }))
        } else {
          showToast(t('service_stopped_msg', { service: t(service.name_key) }))
        }
      }
    } finally {
      togglingServices.value.delete(service.id)
      togglingServices.value = new Set(togglingServices.value)
    }
  }

  async function executeStopService() {
    if (!confirmService.value) return
    const s = confirmService.value
    const w = window as any
    if (w.go?.main?.App?.ToggleService) {
      try {
        const running = await w.go.main.App.ToggleService(s.id)
        s.running = Boolean(running)
        if (!running) {
          showToast(t('service_stopped_msg', { service: t(s.name_key) }))
        } else {
          showToast(t('service_stop_failed', { service: t(s.name_key) }))
        }
        await loadServices()
      } catch (err: any) {
        showToast(
          t('service_stop_failed_error', {
            service: t(s.name_key),
            error: err?.message || err,
          }),
        )
        await loadServices()
      }
    } else {
      s.running = false
      showToast(t('service_stopped_msg', { service: t(s.name_key) }))
    }
    confirmService.value = null
  }

  async function installSSL() {
    const w = window as any
    if (w.go?.main?.App?.InstallSSLTrust) {
      try {
        const res = await w.go.main.App.InstallSSLTrust()
        showToast(renderUIMessage(res) || t('success_ssl'))
        return
      } catch (err: any) {
        showToast(t('error_installing_ssl', { error: err }))
        return
      }
    }
    showToast(t('success_ssl'))
  }

  return {
    services,
    togglingServices,
    confirmService,
    loadServices,
    checkPortConflicts,
    promptToggleService,
    executeStopService,
    installSSL,
    openBrowser,
    openAdminer,
    openPhpMyAdmin,
  }
}
