<template>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 py-2">
    <!-- Figma Web -->
    <a
      href="https://www.figma.com/files"
      target="_blank"
      rel="noopener noreferrer"
      class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
      :title="t('open_external_resource')"
      @click.prevent="openUrl('https://www.figma.com/files')"
    >
      <div
        class="w-10 h-10 rounded-xl bg-purple-100 dark:bg-purple-950 text-purple-600 dark:text-purple-400 flex items-center justify-center font-bold flex-shrink-0"
      >
        <Layout class="w-5 h-5" />
      </div>
      <div class="min-w-0 flex-1">
        <h4
          class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
        >
          <span>{{ t('open_figma_web') }}</span>
          <ExternalLink class="w-3 h-3 shrink-0" />
        </h4>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          {{ t('resource_figma_desc') }}
        </p>
      </div>
    </a>

    <!-- Coolors Palette -->
    <a
      href="https://coolors.co/generate"
      target="_blank"
      rel="noopener noreferrer"
      class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
      :title="t('open_external_resource')"
      @click.prevent="openUrl('https://coolors.co/generate')"
    >
      <div
        class="w-10 h-10 rounded-xl bg-blue-100 dark:bg-blue-950 text-blue-600 dark:text-blue-400 flex items-center justify-center font-bold flex-shrink-0"
      >
        <Palette class="w-5 h-5" />
      </div>
      <div class="min-w-0 flex-1">
        <h4
          class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
        >
          <span>{{ t('open_coolors') }}</span>
          <ExternalLink class="w-3 h-3 shrink-0" />
        </h4>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          {{ t('resource_coolors_desc') }}
        </p>
      </div>
    </a>

    <!-- Google Fonts -->
    <a
      href="https://fonts.google.com"
      target="_blank"
      rel="noopener noreferrer"
      class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
      :title="t('open_external_resource')"
      @click.prevent="openUrl('https://fonts.google.com')"
    >
      <div
        class="w-10 h-10 rounded-xl bg-amber-100 dark:bg-amber-950 text-amber-600 dark:text-amber-400 flex items-center justify-center font-bold flex-shrink-0"
      >
        <Type class="w-5 h-5" />
      </div>
      <div class="min-w-0 flex-1">
        <h4
          class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
        >
          <span>{{ t('open_google_fonts') }}</span>
          <ExternalLink class="w-3 h-3 shrink-0" />
        </h4>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          {{ t('resource_google_fonts_desc') }}
        </p>
      </div>
    </a>

    <!-- Lucide Icons -->
    <a
      href="https://lucide.dev/icons"
      target="_blank"
      rel="noopener noreferrer"
      class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
      :title="t('open_external_resource')"
      @click.prevent="openUrl('https://lucide.dev/icons')"
    >
      <div
        class="w-10 h-10 rounded-xl bg-emerald-100 dark:bg-emerald-950 text-emerald-600 dark:text-emerald-400 flex items-center justify-center font-bold flex-shrink-0"
      >
        <Layout class="w-5 h-5" />
      </div>
      <div class="min-w-0 flex-1">
        <h4
          class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
        >
          <span>{{ t('open_lucide_icons') }}</span>
          <ExternalLink class="w-3 h-3 shrink-0" />
        </h4>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          {{ t('resource_lucide_desc') }}
        </p>
      </div>
    </a>
  </div>
</template>

<script setup lang="ts">
import {
  Layout,
  Palette,
  Type,
  ExternalLink,
} from 'lucide-vue-next'

interface Props {
  t: (key: string, params?: Record<string, string | number | undefined>) => string
}

defineProps<Props>()

function openUrl(url: string) {
  const w = window as any
  if (typeof w.runtime?.BrowserOpenURL === 'function') {
    w.runtime.BrowserOpenURL(url)
  } else {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
}
</script>
