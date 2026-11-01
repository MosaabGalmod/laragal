<template>
  <div class="space-y-4">
    <div
      class="flex flex-wrap items-center justify-between gap-3 p-3 rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700"
    >
      <div class="flex items-center gap-2">
        <label
          for="design-site-select"
          class="text-xs font-bold text-slate-600 dark:text-slate-300"
        >
          {{ t('select_site_to_preview') }}:
        </label>
        <select
          id="design-site-select"
          v-model="selectedDomain"
          class="px-3 py-1.5 rounded-lg text-xs font-mono font-bold bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 text-slate-900 dark:text-white"
        >
          <option
            v-for="s in webSitesList"
            :key="s.domain"
            :value="s.domain"
          >
            {{ s.name || s.domain }} (https://{{ s.domain }})
          </option>
        </select>
      </div>

      <div class="flex items-center gap-1.5">
        <!-- Mobile (375px) -->
        <button
          type="button"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors cursor-pointer"
          :class="
            viewportMode === 'mobile'
              ? 'bg-indigo-600 text-white shadow-xs'
              : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700'
          "
          :title="t('preview_mobile')"
          @click="viewportMode = 'mobile'"
        >
          <Smartphone class="w-3.5 h-3.5" />
          <span>{{ t('preview_mobile') }}</span>
        </button>

        <!-- Tablet (768px) -->
        <button
          type="button"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors cursor-pointer"
          :class="
            viewportMode === 'tablet'
              ? 'bg-indigo-600 text-white shadow-xs'
              : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700'
          "
          :title="t('preview_tablet')"
          @click="viewportMode = 'tablet'"
        >
          <Tablet class="w-3.5 h-3.5" />
          <span>{{ t('preview_tablet') }}</span>
        </button>

        <!-- Desktop (100%) -->
        <button
          type="button"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors cursor-pointer"
          :class="
            viewportMode === 'desktop'
              ? 'bg-indigo-600 text-white shadow-xs'
              : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700'
          "
          :title="t('preview_desktop')"
          @click="viewportMode = 'desktop'"
        >
          <Monitor class="w-3.5 h-3.5" />
          <span>{{ t('preview_desktop') }}</span>
        </button>

        <!-- Rotate -->
        <button
          type="button"
          class="p-1.5 rounded-lg border transition-colors cursor-pointer"
          :class="
            isRotated
              ? 'bg-indigo-100 text-indigo-700 border-indigo-300 dark:bg-indigo-950 dark:border-indigo-700 dark:text-indigo-300'
              : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700'
          "
          :title="t('preview_rotate')"
          :aria-label="t('preview_rotate')"
          @click="isRotated = !isRotated"
        >
          <RotateCw class="w-3.5 h-3.5" />
        </button>

        <!-- Refresh -->
        <button
          type="button"
          class="p-1.5 rounded-lg border bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          :title="t('refresh_preview')"
          :aria-label="t('refresh_preview')"
          @click="reloadPreviewFrame"
        >
          <RefreshCw
            class="w-3.5 h-3.5"
            :class="{ 'animate-spin': isRefreshing }"
          />
        </button>
      </div>
    </div>

    <!-- Preview Frame Wrapper -->
    <div
      class="w-full min-h-[420px] bg-slate-200/60 dark:bg-slate-950/60 rounded-2xl border border-slate-200 dark:border-slate-800 p-4 flex flex-col items-center justify-center overflow-hidden"
    >
      <div
        v-if="selectedDomain"
        class="w-full flex flex-col items-center gap-3"
      >
        <!-- 1. Checking Frameable / Initial Loading State -->
        <div
          v-if="isCheckingFrameable"
          class="text-center py-16 space-y-3 flex flex-col items-center justify-center"
        >
          <Loader2 class="w-8 h-8 text-indigo-600 dark:text-indigo-400 animate-spin" />
          <p class="text-xs font-semibold text-slate-600 dark:text-slate-300">
            {{ t('preview_loading') }}
          </p>
        </div>

        <!-- 2. Blocked / Unreachable Panel (Do NOT render iframe) -->
        <div
          v-else-if="frameCheckResult && !frameCheckResult.frameable"
          class="w-full max-w-lg mx-auto p-6 rounded-2xl bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-800 shadow-sm flex flex-col items-center text-center space-y-4 my-6"
        >
          <div
            class="w-12 h-12 rounded-full bg-amber-100 dark:bg-amber-950/60 border border-amber-300 dark:border-amber-800 flex items-center justify-center text-amber-600 dark:text-amber-400"
          >
            <ShieldAlert class="w-6 h-6" />
          </div>

          <div class="space-y-2">
            <h3 class="text-sm font-bold text-slate-900 dark:text-slate-100">
              {{ t('preview_blocked_title') }}
            </h3>
            <p
              class="text-xs leading-relaxed text-slate-600 dark:text-slate-300"
              v-html="blockedExplanation"
            ></p>
          </div>

          <button
            type="button"
            class="px-4 py-2 rounded-xl text-xs font-bold bg-indigo-600 hover:bg-indigo-700 active:bg-indigo-800 text-white shadow-xs flex items-center gap-2 transition-colors cursor-pointer"
            @click="openInBrowser"
          >
            <ExternalLink class="w-3.5 h-3.5" />
            <span>{{ t('preview_open_in_browser') }}</span>
          </button>
        </div>

        <!-- 3. Frameable: Render Iframe Container -->
        <div
          v-else-if="frameCheckResult && frameCheckResult.frameable"
          class="w-full flex flex-col items-center gap-2"
        >
          <div
            class="text-[11px] font-mono font-bold text-slate-500 dark:text-slate-400 bg-white/80 dark:bg-slate-900/80 px-2.5 py-0.5 rounded-md border border-slate-200 dark:border-slate-800 flex items-center gap-1.5"
          >
            <span>{{ t('preview_dimensions') }}:</span>
            <bdi dir="ltr">{{ currentWidth }} &times; {{ currentHeight }}px</bdi>
          </div>

          <!-- 10s Timeout Fallback Message -->
          <div
            v-if="iframeTimedOut"
            class="w-full max-w-lg p-3 rounded-xl bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-800 flex items-center justify-between gap-3 text-xs text-amber-800 dark:text-amber-200"
          >
            <div class="flex items-center gap-2">
              <AlertCircle class="w-4 h-4 shrink-0 text-amber-600 dark:text-amber-400" />
              <span class="leading-relaxed">{{ t('preview_load_timeout') }}</span>
            </div>
            <button
              type="button"
              class="shrink-0 px-2.5 py-1 rounded-lg text-xs font-bold bg-amber-600 hover:bg-amber-700 text-white flex items-center gap-1.5 transition-colors cursor-pointer"
              @click="openInBrowser"
            >
              <ExternalLink class="w-3 h-3" />
              <span>{{ t('preview_open_in_browser') }}</span>
            </button>
          </div>

          <div
            class="relative transition-all duration-300 bg-white rounded-xl shadow-lg border border-slate-300 dark:border-slate-700 overflow-hidden flex justify-center max-w-full"
            :style="previewFrameContainerStyle"
          >
            <!-- Loading Spinner Overlay until iframe load event fires -->
            <div
              v-if="isFrameLoading"
              class="absolute inset-0 z-10 bg-white/85 dark:bg-slate-900/85 backdrop-blur-xs flex flex-col items-center justify-center gap-2 transition-opacity"
            >
              <Loader2 class="w-6 h-6 text-indigo-600 dark:text-indigo-400 animate-spin" />
              <span class="text-xs font-semibold text-slate-600 dark:text-slate-300">
                {{ t('preview_loading') }}
              </span>
            </div>

            <iframe
              :key="frameKey"
              :src="'https://' + selectedDomain"
              :style="previewIframeStyle"
              class="border-0 bg-white"
              loading="lazy"
              title="LaraGal Web Responsive Preview"
              sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
              @load="onIframeLoad"
            ></iframe>
          </div>
        </div>
      </div>

      <div
        v-else
        class="text-center py-12 space-y-2"
      >
        <Layout class="w-8 h-8 text-slate-400 mx-auto" />
        <p class="text-xs font-semibold text-slate-500 dark:text-slate-400">
          {{ t('preview_no_site_selected') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import {
  Layout,
  Smartphone,
  Tablet,
  Monitor,
  RotateCw,
  RefreshCw,
  Loader2,
  ShieldAlert,
  ExternalLink,
  AlertCircle,
} from 'lucide-vue-next'
import type { SiteItem, FrameCheckResult } from '../../types'

interface Props {
  sites?: SiteItem[]
  initialDomain?: string
  t: (key: string, params?: Record<string, string | number | undefined>) => string
}

const props = withDefaults(defineProps<Props>(), {
  sites: () => [],
  initialDomain: '',
})

const selectedDomain = ref('')
const viewportMode = ref<'mobile' | 'tablet' | 'desktop'>('mobile')
const isRotated = ref(false)
const frameKey = ref(1)
const isRefreshing = ref(false)

const isCheckingFrameable = ref(false)
const isFrameLoading = ref(false)
const frameCheckResult = ref<FrameCheckResult | null>(null)
const iframeTimedOut = ref(false)
let iframeTimeoutTimer: ReturnType<typeof setTimeout> | null = null

const webSitesList = computed(() => {
  return props.sites.filter((s) => s.is_web && s.domain)
})

function clearIframeTimeout() {
  if (iframeTimeoutTimer) {
    clearTimeout(iframeTimeoutTimer)
    iframeTimeoutTimer = null
  }
}

function startIframeTimeout() {
  clearIframeTimeout()
  iframeTimedOut.value = false
  iframeTimeoutTimer = setTimeout(() => {
    if (isFrameLoading.value) {
      iframeTimedOut.value = true
    }
  }, 10000)
}

function onIframeLoad() {
  isFrameLoading.value = false
  iframeTimedOut.value = false
  clearIframeTimeout()
}

onUnmounted(() => {
  clearIframeTimeout()
})

async function checkCurrentDomain(domain: string) {
  if (!domain) {
    frameCheckResult.value = null
    isCheckingFrameable.value = false
    isFrameLoading.value = false
    clearIframeTimeout()
    return
  }

  isCheckingFrameable.value = true
  isFrameLoading.value = false
  iframeTimedOut.value = false
  clearIframeTimeout()
  frameCheckResult.value = null

  const w = window as any
  if (typeof w.go?.main?.App?.CheckFrameable === 'function') {
    try {
      const res = await w.go.main.App.CheckFrameable(domain)
      frameCheckResult.value = res
    } catch {
      frameCheckResult.value = {
        frameable: false,
        reason: 'unreachable',
        status: 0,
      }
    }
  } else {
    // Fallback if Wails bindings not present in dev/test environment
    frameCheckResult.value = {
      frameable: true,
      reason: '',
      status: 200,
    }
  }

  isCheckingFrameable.value = false

  if (frameCheckResult.value?.frameable) {
    isFrameLoading.value = true
    startIframeTimeout()
  }
}

watch(
  () => props.initialDomain,
  (newDomain) => {
    if (newDomain) {
      selectedDomain.value = newDomain
    } else if (!selectedDomain.value && webSitesList.value.length > 0) {
      selectedDomain.value = webSitesList.value[0].domain
    }
  },
  { immediate: true },
)

watch(
  webSitesList,
  (list) => {
    if (!selectedDomain.value && list.length > 0) {
      selectedDomain.value = list[0].domain
    }
  },
  { immediate: true },
)

watch(
  selectedDomain,
  (newDomain) => {
    if (newDomain) {
      checkCurrentDomain(newDomain)
    }
  },
  { immediate: true },
)

const blockedExplanation = computed(() => {
  const reason = frameCheckResult.value?.reason
  if (reason === 'x_frame_options') {
    return props.t('preview_blocked_xfo')
  }
  if (reason === 'csp_frame_ancestors') {
    return props.t('preview_blocked_csp')
  }
  if (reason === 'unreachable') {
    return props.t('preview_unreachable')
  }
  return props.t('preview_blocked_xfo')
})

function openInBrowser() {
  if (!selectedDomain.value) return
  const url = 'https://' + selectedDomain.value
  const w = window as any
  if (typeof w.runtime?.BrowserOpenURL === 'function') {
    w.runtime.BrowserOpenURL(url)
  } else {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
}

const currentWidth = computed(() => {
  if (viewportMode.value === 'desktop') return 1024
  if (viewportMode.value === 'tablet') {
    return isRotated.value ? 1024 : 768
  }
  return isRotated.value ? 667 : 375
})

const currentHeight = computed(() => {
  if (viewportMode.value === 'desktop') return 640
  if (viewportMode.value === 'tablet') {
    return isRotated.value ? 768 : 1024
  }
  return isRotated.value ? 375 : 667
})

const previewFrameContainerStyle = computed(() => {
  if (viewportMode.value === 'desktop') {
    return { width: '100%', maxWidth: '100%' }
  }
  return {
    width: `${currentWidth.value}px`,
    maxWidth: '100%',
  }
})

const previewIframeStyle = computed(() => {
  if (viewportMode.value === 'desktop') {
    return { width: '100%', height: '520px' }
  }
  return {
    width: '100%',
    height: `${Math.min(currentHeight.value, 520)}px`,
  }
})

async function reloadPreviewFrame() {
  if (isRefreshing.value || !selectedDomain.value) return
  isRefreshing.value = true
  frameKey.value++
  await checkCurrentDomain(selectedDomain.value)
  setTimeout(() => {
    isRefreshing.value = false
  }, 400)
}
</script>
