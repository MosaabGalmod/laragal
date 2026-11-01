import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import ResponsivePreview from './ResponsivePreview.vue'
import { useI18n } from '../../composables/useI18n'
import type { SiteItem } from '../../types'

describe('ResponsivePreview component', () => {
  const { t } = useI18n()

  const mockSites: SiteItem[] = [
    {
      name: 'blocked-site',
      domain: 'blocked.test',
      root_path: '/path/to/blocked',
      is_web: true,
    },
    {
      name: 'allowed-site',
      domain: 'allowed.test',
      root_path: '/path/to/allowed',
      is_web: true,
    },
  ]

  let browserOpenUrlMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.useRealTimers()
    browserOpenUrlMock = vi.fn()
    ;(window as any).runtime = {
      BrowserOpenURL: browserOpenUrlMock,
    }
    ;(window as any).go = {
      main: {
        App: {
          CheckFrameable: vi.fn(),
        },
      },
    }
  })

  it('renders blocked panel and no iframe when X-Frame-Options blocks framing', async () => {
    ;(window as any).go.main.App.CheckFrameable.mockResolvedValue({
      frameable: false,
      reason: 'x_frame_options',
      status: 200,
    })

    const wrapper = mount(ResponsivePreview, {
      props: {
        sites: mockSites,
        initialDomain: 'blocked.test',
        t,
      },
    })

    await flushPromises()

    expect((window as any).go.main.App.CheckFrameable).toHaveBeenCalledWith('blocked.test')
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.text()).toContain(t('preview_blocked_title'))

    // Click open in browser button
    const openBtn = wrapper.findAll('button').find((b) => b.text().includes(t('preview_open_in_browser')))
    expect(openBtn).toBeDefined()
    await openBtn?.trigger('click')
    expect(browserOpenUrlMock).toHaveBeenCalledWith('https://blocked.test')
  })

  it('renders CSP explanation when reason is csp_frame_ancestors', async () => {
    ;(window as any).go.main.App.CheckFrameable.mockResolvedValue({
      frameable: false,
      reason: 'csp_frame_ancestors',
      status: 200,
    })

    const wrapper = mount(ResponsivePreview, {
      props: {
        sites: mockSites,
        initialDomain: 'blocked.test',
        t,
      },
    })

    await flushPromises()

    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.html()).toContain('frame-ancestors')
  })

  it('renders unreachable explanation when reason is unreachable', async () => {
    ;(window as any).go.main.App.CheckFrameable.mockResolvedValue({
      frameable: false,
      reason: 'unreachable',
      status: 0,
    })

    const wrapper = mount(ResponsivePreview, {
      props: {
        sites: mockSites,
        initialDomain: 'blocked.test',
        t,
      },
    })

    await flushPromises()

    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.text()).toContain(t('preview_unreachable'))
  })

  it('renders iframe when site is frameable and clears loading on load', async () => {
    ;(window as any).go.main.App.CheckFrameable.mockResolvedValue({
      frameable: true,
      reason: '',
      status: 200,
    })

    const wrapper = mount(ResponsivePreview, {
      props: {
        sites: mockSites,
        initialDomain: 'allowed.test',
        t,
      },
    })

    await flushPromises()

    const iframe = wrapper.find('iframe')
    expect(iframe.exists()).toBe(true)
    expect(iframe.attributes('src')).toBe('https://allowed.test')

    // Trigger iframe load event
    await iframe.trigger('load')
    expect(wrapper.find('.animate-spin').exists()).toBe(false)
  })

  it('shows 10s fallback message when iframe load times out', async () => {
    vi.useFakeTimers()
    ;(window as any).go.main.App.CheckFrameable.mockResolvedValue({
      frameable: true,
      reason: '',
      status: 200,
    })

    const wrapper = mount(ResponsivePreview, {
      props: {
        sites: mockSites,
        initialDomain: 'allowed.test',
        t,
      },
    })

    await flushPromises()

    expect(wrapper.find('iframe').exists()).toBe(true)
    expect(wrapper.text()).not.toContain(t('preview_load_timeout'))

    // Fast-forward 10 seconds
    vi.advanceTimersByTime(10000)
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain(t('preview_load_timeout'))

    // Clicking open in browser from timeout banner
    const openTimeoutBtn = wrapper.findAll('button').find((b) => b.text().includes(t('preview_open_in_browser')))
    expect(openTimeoutBtn).toBeDefined()
    await openTimeoutBtn?.trigger('click')
    expect(browserOpenUrlMock).toHaveBeenCalledWith('https://allowed.test')

    vi.useRealTimers()
  })
})
