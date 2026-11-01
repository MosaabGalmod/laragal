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
})

