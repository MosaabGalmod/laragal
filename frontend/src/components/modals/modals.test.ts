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

    // 2. Click the upgrade button (should use distinct major upgrade label and amber style)
    const upgradeBtn = wrapper.findAll('button').find((b) => b.text().includes(t('runtime_major_upgrade_btn')))
    expect(upgradeBtn).toBeDefined()
    expect(upgradeBtn!.classes()).toContain('bg-amber-600')
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

  it('installed newer redis (v8.10.3) shows no EOL badge and no major-change badge', async () => {
    ;(window as any).go.main.App.GetRuntimeCatalog = async () => [
      {
        name: 'redis',
        version: '5.0.14.1',
        eol: true,
      },
    ]

    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'redis',
        current_version: '8.10.3',
        latest_version: '8.10.3',
        has_update: false,
        is_security_fix: false,
        major_change: true,
        download_url: '',
      },
    ]

    const wrapper = mount(DownloadModal, {
      props: {
        modelValue: true,
      },
    })

    const { useRuntimeCatalog } = await import('../../composables/useRuntimeCatalog')
    const { loadRuntimeCatalog, checkUpdatesNow } = useRuntimeCatalog()
    await loadRuntimeCatalog()
    await checkUpdatesNow({ silent: true })
    await wrapper.vm.$nextTick()

    // 1. Must show installed version
    expect(wrapper.text()).toContain(t('runtime_installed'))
    expect(wrapper.text()).toContain('v8.10.3')

    // 2. Must NOT show EOL badge because installed is 8.10.3, not catalog 5.0.x
    expect(wrapper.text()).not.toContain(t('runtime_eol_badge'))

    // 3. Must NOT show major-change badge because has_update is false
    expect(wrapper.text()).not.toContain(t('runtime_major_change'))

    // 4. If Redis was NOT installed or installed version matched catalog 5.0.x, EOL badge SHOULD show
    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'redis',
        current_version: '',
        latest_version: '5.0.14.1',
        has_update: false,
        is_security_fix: false,
        download_url: '',
      },
    ]
    await checkUpdatesNow({ silent: true })
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain(t('runtime_eol_badge'))
  })

  it('feed mariadb entry renders major upgrade button with amber styling and single new indicator', async () => {
    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'mariadb-11.4',
        current_version: '',
        latest_version: '11.4.5',
        has_update: true,
        is_new: true,
        major_change: true,
        is_security_fix: false,
        download_url: 'https://downloads.mariadb.org/11.4.5.zip',
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

    // 1. Major upgrade button label and amber class
    const upgradeBtn = wrapper.findAll('button').find((b) => b.text().includes(t('runtime_major_upgrade_btn')))
    expect(upgradeBtn).toBeDefined()
    expect(upgradeBtn!.classes()).toContain('bg-amber-600')

    // 2. Single "new" indicator: badge is present
    const mariaCard = wrapper.findAll('.p-3').find((c) => c.text().includes(t('mariadb_newer_series')))
    expect(mariaCard).toBeDefined()
    expect(mariaCard!.text()).toContain(t('runtime_new_from_feed'))

    // Count occurrences of t('runtime_new_from_feed') on the card - must only appear once as the badge
    const badgeText = t('runtime_new_from_feed')
    const matches = mariaCard!.text().split(badgeText).length - 1
    expect(matches).toBe(1)

    // 3. Description line shows family description (mysql -> service_mysql_desc)
    expect(mariaCard!.text()).toContain(t('service_mysql_desc'))
  })

  it('merges 3 MariaDB newer series entries into ONE card with series picker, defaults to latest EOL (11.8), and confirms upgrade with selected name', async () => {
    let confirmedCalledWith = ''
    ;(window as any).go.main.App.UpgradeRuntimeConfirmed = async (name: string) => {
      confirmedCalledWith = name
      return { key: 'ok' }
    }

    ;(window as any).go.main.App.CheckRuntimeUpdates = async () => [
      {
        name: 'mysql',
        current_version: '10.11.19',
        latest_version: '10.11.19',
        has_update: false,
        eol_date: '2028-02-16',
      },
      {
        name: 'mariadb-11.4',
        current_version: '',
        latest_version: '11.4.5',
        has_update: true,
        is_new: true,
        major_change: true,
        eol_date: '2029-05-29',
        download_url: 'https://downloads.mariadb.org/11.4.5.zip',
      },
      {
        name: 'mariadb-11.8',
        current_version: '',
        latest_version: '11.8.10',
        has_update: true,
        is_new: true,
        major_change: true,
        eol_date: '2030-02-13',
        download_url: 'https://downloads.mariadb.org/11.8.10.zip',
      },
      {
        name: 'mariadb-12.3',
        current_version: '',
        latest_version: '12.3.1',
        has_update: true,
        is_new: true,
        major_change: true,
        eol_date: '2029-06-12',
        download_url: 'https://downloads.mariadb.org/12.3.1.zip',
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

    // 1. Verify 3 entries -> 1 card titled t('mariadb_newer_series')
    const newerSeriesCards = wrapper.findAll('.p-3').filter((c) => c.text().includes(t('mariadb_newer_series')))
    expect(newerSeriesCards.length).toBe(1)
    const mariaCard = newerSeriesCards[0]

    // 2. Main mysql card shows installed version and supported until date
    const mysqlCard = wrapper.findAll('.p-3').find((c) => c.text().includes('MariaDB / MySQL Server'))
    expect(mysqlCard).toBeDefined()
    expect(mysqlCard!.text()).toContain('v10.11.19')
    expect(mysqlCard!.text()).toContain(t('runtime_supported_until', { date: '' }).replace('{date}', '').trim())

    // 3. 3 options in select, default is 11.8 (latest EOL 2030-02-13 over 12.3 2029-06-12)
    const select = mariaCard.find('select')
    expect(select.exists()).toBe(true)
    const options = select.findAll('option')
    expect(options.length).toBe(3)
    expect((select.element as HTMLSelectElement).value).toBe('mariadb-11.8')

    // 4. Click major upgrade button on the card
    const upgradeBtn = mariaCard.find('button.bg-amber-600')
    expect(upgradeBtn.exists()).toBe(true)
    await upgradeBtn.trigger('click')
    await wrapper.vm.$nextTick()

    // 5. Confirm modal opened with mysql warning
    expect(wrapper.text()).toContain(t('runtime_major_change_confirm_title'))
    expect(wrapper.text()).toContain(t('runtime_major_change_mysql_extra'))

    // 6. Confirm calls UpgradeRuntimeConfirmed with selected name ('mariadb-11.8')
    const confirmBtn = wrapper.findAll('button').find((b) => b.text().trim() === t('confirm_btn'))
    expect(confirmBtn).toBeDefined()
    await confirmBtn!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(confirmedCalledWith).toBe('mariadb-11.8')
  })
})

