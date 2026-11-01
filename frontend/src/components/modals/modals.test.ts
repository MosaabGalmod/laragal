import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import NewProjectModal from './NewProjectModal.vue'
import QuitConfirmModal from './QuitConfirmModal.vue'
import DownloadModal from './DownloadModal.vue'
import DirectoryModal from './DirectoryModal.vue'
import ProjectSettingsModal from './ProjectSettingsModal.vue'
import DiagnosticsModal from './DiagnosticsModal.vue'
import DesignToolsModal from '../DesignToolsModal.vue'
import { useI18n } from '../../composables/useI18n'
import type { SiteItem } from '../../types'

describe('Modal smoke tests', () => {
  const { t } = useI18n()

  const mockSite: SiteItem = {
    name: 'demo',
    domain: 'demo.test',
    path: 'C:/laragal/sites/demo',
    root_path: 'C:/laragal/sites/demo/public',
    is_web: true,
    project_type: 'laravel',
    php_version: 'PHP 8.3',
  }

  beforeEach(() => {
    ;(window as any).go = {
      main: {
        App: {
          GetRuntimeCatalog: async () => [],
          CheckRuntimeUpdates: async () => [],
          GetRecentDirectories: async () => [],
          GetSites: async () => [],
          GetServices: async () => [],
          DiagnoseProject: async () => null,
          ScanProjectDatabases: async () => [],
        },
      },
    }
    ;(window as any).runtime = {
      EventsOn: () => {},
      BrowserOpenURL: () => {},
    }
  })

  it('NewProjectModal mounts and renders title', () => {
    const wrapper = mount(NewProjectModal, {
      props: {
        modelValue: true,
        projectsDirectory: 'C:/laragal/sites',
        sites: [mockSite],
      },
    })
    const title = t('modal_new_title')
    expect(wrapper.text()).toContain(title)
  })

  it('QuitConfirmModal mounts and renders title', () => {
    const wrapper = mount(QuitConfirmModal, {
      props: {
        modelValue: true,
      },
    })
    const title = t('confirm_quit_title')
    expect(wrapper.text()).toContain(title)
  })

  it('DownloadModal mounts and renders title', () => {
    const wrapper = mount(DownloadModal, {
      props: {
        modelValue: true,
      },
    })
    const title = t('modal_download_title')
    expect(wrapper.text()).toContain(title)
  })

  it('DirectoryModal mounts and renders title', () => {
    const wrapper = mount(DirectoryModal, {
      props: {
        modelValue: true,
        currentDirectory: 'C:/laragal/sites',
        suggestedPaths: ['D:/projects', 'C:/laragal/sites'],
      },
    })
    const title = t('modal_dir_title')
    expect(wrapper.text()).toContain(title)
  })

  it('ProjectSettingsModal mounts and renders title', () => {
    const wrapper = mount(ProjectSettingsModal, {
      props: {
        modelValue: true,
        site: mockSite,
      },
    })
    const title = t('modal_settings_title')
    expect(wrapper.text()).toContain(title)
  })

  it('DiagnosticsModal mounts and renders title', () => {
    const wrapper = mount(DiagnosticsModal, {
      props: {
        modelValue: true,
        site: mockSite,
      },
    })
    const title = t('modal_diag_title')
    expect(wrapper.text()).toContain(title)
  })

  it('DesignToolsModal mounts and renders title', () => {
    const wrapper = mount(DesignToolsModal, {
      props: {
        modelValue: true,
        sites: [mockSite],
        t,
      },
    })
    const title = t('design_tools_title')
    expect(wrapper.text()).toContain(title)
  })

  it('DownloadModal displays installed state and reinstall action for up-to-date runtime', async () => {
    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'php-8.5',
        current_version: '8.5.12',
        latest_version: '8.5.12',
        has_update: false,
        is_security_fix: false,
        download_url: '',
      },
      {
        name: 'caddy',
        current_version: '',
        latest_version: '2.8.4',
        has_update: false,
        is_security_fix: false,
        download_url: '',
      },
    ]

    const wrapper = mount(DownloadModal, {
      props: {
        modelValue: true,
      },
    })

    // Trigger update check
    const { useRuntimeCatalog } = await import('../../composables/useRuntimeCatalog')
    const { checkUpdatesNow } = useRuntimeCatalog()
    await checkUpdatesNow({ silent: true })
    await wrapper.vm.$nextTick()

    // Up to date runtime should display installed badge and reinstall button
    expect(wrapper.text()).toContain(t('runtime_installed'))
    expect(wrapper.text()).toContain('v8.5.12')
    expect(wrapper.text()).toContain(t('runtime_reinstall'))

    // Not-installed runtime should show download now button
    expect(wrapper.text()).toContain(t('download_now'))
  })

  it('DownloadModal displays feed-only runtime with badge runtime_new_from_feed', async () => {
    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'php-8.6',
        current_version: '',
        latest_version: '8.6.0',
        has_update: false,
        is_security_fix: false,
        download_url: 'https://example.com/php-8.6.zip',
        is_new: true,
      },
    ]

    const wrapper = mount(DownloadModal, {
      props: {
        modelValue: true,
      },
    })

    const { useRuntimeCatalog } = await import('../../composables/useRuntimeCatalog')
    const { checkUpdatesNow } = useRuntimeCatalog()
    await checkUpdatesNow({ silent: true })
    await wrapper.vm.$nextTick()

    // Feed-only runtime should display the badge t('runtime_new_from_feed')
    expect(wrapper.text()).toContain(t('runtime_new_from_feed'))
    expect(wrapper.text()).toContain('PHP 8.6')
  })

  it('major-change entry shows amber badge, opens confirm modal and calls UpgradeRuntimeConfirmed only after confirm', async () => {
    let confirmedCalledWith = ''
    let unconfirmedCalledWith = ''
    ;(window as any).go.main.App.UpgradeRuntime = async (name: string) => {
      unconfirmedCalledWith = name
      return { key: 'ok' }
    }
    ;(window as any).go.main.App.UpgradeRuntimeConfirmed = async (name: string) => {
      confirmedCalledWith = name
      return { key: 'ok' }
    }

    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'mysql',
        current_version: '10.11.11',
        latest_version: '11.4.5',
        has_update: true,
        is_security_fix: false,
        major_change: true,
        download_url: 'https://downloads.mariadb.org/rest-api/mariadb/11.4.5/mariadb-11.4.5-winx64.zip',
        compat_note: '',
      },
    ]

    const wrapper = mount(DownloadModal, {
      props: {
        modelValue: true,
      },
    })

    const { useRuntimeCatalog } = await import('../../composables/useRuntimeCatalog')
    const { checkUpdatesNow } = useRuntimeCatalog()
    await checkUpdatesNow({ silent: true })
    await wrapper.vm.$nextTick()

    // 1. Must show amber badge t('runtime_major_change')
    expect(wrapper.text()).toContain(t('runtime_major_change'))

    // 2. Click the upgrade button
    const upgradeBtn = wrapper.findAll('button').find((b) => b.text().includes(t('upgrade_one_click')))
    expect(upgradeBtn).toBeDefined()
    await upgradeBtn!.trigger('click')
    await wrapper.vm.$nextTick()

    // 3. Confirmation modal must open
    expect(wrapper.text()).toContain(t('runtime_major_change_confirm_title'))
    expect(wrapper.text()).toContain(t('runtime_major_change_warning'))
    expect(wrapper.text()).toContain(t('runtime_major_change_mysql_extra'))

    // Neither confirmed nor unconfirmed upgrade should have been called yet
    expect(confirmedCalledWith).toBe('')
    expect(unconfirmedCalledWith).toBe('')

    // 4. Click confirm button inside the modal
    const confirmBtn = wrapper.findAll('button').find((b) => b.text().trim() === t('confirm_btn'))
    expect(confirmBtn).toBeDefined()
    await confirmBtn!.trigger('click')
    await wrapper.vm.$nextTick()

    // 5. UpgradeRuntimeConfirmed must have been called with 'mysql'
    expect(confirmedCalledWith).toBe('mysql')
    expect(unconfirmedCalledWith).toBe('')
  })

  it('DownloadModal renders catalog_fallback note and tooltip reason when source_status is catalog_fallback', async () => {
    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'redis',
        current_version: '5.0.14.1',
        latest_version: '5.0.14.1',
        has_update: false,
        is_security_fix: false,
        download_url: 'https://example.com/redis.zip',
        source_status: 'catalog_fallback',
        source_error: 'github rate limit, resets 23:40',
      },
    ]

    const wrapper = mount(DownloadModal, {
      props: {
        modelValue: true,
      },
    })

    const { useRuntimeCatalog } = await import('../../composables/useRuntimeCatalog')
    const { checkUpdatesNow } = useRuntimeCatalog()
    await checkUpdatesNow({ silent: true })
    await wrapper.vm.$nextTick()

    // Must render the fallback note
    expect(wrapper.text()).toContain(t('runtime_source_unavailable'))

    // Must have the error in title attribute tooltip
    const fallbackEl = wrapper.find('[title="github rate limit, resets 23:40"]')
    expect(fallbackEl.exists()).toBe(true)
    expect(fallbackEl.text()).toContain(t('runtime_source_unavailable'))

    // When source_status is live, fallback note must NOT render
    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'redis',
        current_version: '5.0.14.1',
        latest_version: '8.10.3',
        has_update: true,
        is_security_fix: false,
        download_url: 'https://example.com/redis.zip',
        source_status: 'live',
      },
    ]
    await checkUpdatesNow({ silent: true })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).not.toContain(t('runtime_source_unavailable'))
  })
})

