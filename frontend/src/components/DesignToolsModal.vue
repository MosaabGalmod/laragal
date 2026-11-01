<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/75 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="design-tools-modal-title"
    tabindex="-1"
    @click.self="closeModal"
  >
    <div
      class="w-full max-w-4xl max-h-[92vh] rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl flex flex-col overflow-hidden animate-scale-in"
    >
      <!-- Header -->
      <div
        class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between flex-shrink-0"
      >
        <div class="flex items-center gap-2.5">
          <div
            class="w-8 h-8 rounded-lg bg-indigo-50 dark:bg-indigo-950/80 border border-indigo-200 dark:border-indigo-800 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0"
          >
            <Layout class="w-4 h-4" />
          </div>
          <div>
            <h3
              id="design-tools-modal-title"
              class="font-bold text-base text-slate-900 dark:text-white"
            >
              {{ t('design_tools_title') }}
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              {{ t('design_tools_desc') }}
            </p>
          </div>
        </div>

        <button
          type="button"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1.5 rounded-lg transition-colors cursor-pointer"
          :title="t('close_btn')"
          :aria-label="t('close_btn')"
          @click="closeModal"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Tabs Navigation -->
      <div
        class="px-6 pt-3 border-b border-slate-100 dark:border-slate-800 flex items-center gap-2 flex-shrink-0 bg-slate-50/50 dark:bg-slate-900/50"
      >
        <button
          type="button"
          class="px-4 py-2 rounded-t-xl text-xs font-bold border-b-2 transition-all cursor-pointer flex items-center gap-2"
          :class="
            activeTab === 'responsive'
              ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
          "
          @click="activeTab = 'responsive'"
        >
          <Smartphone class="w-3.5 h-3.5" />
          <span>{{ t('tab_responsive') }}</span>
        </button>

        <button
          type="button"
          class="px-4 py-2 rounded-t-xl text-xs font-bold border-b-2 transition-all cursor-pointer flex items-center gap-2"
          :class="
            activeTab === 'contrast'
              ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
          "
          @click="activeTab = 'contrast'"
        >
          <Palette class="w-3.5 h-3.5" />
          <span>{{ t('tab_contrast') }}</span>
        </button>

        <button
          type="button"
          class="px-4 py-2 rounded-t-xl text-xs font-bold border-b-2 transition-all cursor-pointer flex items-center gap-2"
          :class="
            activeTab === 'cloud'
              ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
          "
          @click="activeTab = 'cloud'"
        >
          <ExternalLink class="w-3.5 h-3.5" />
          <span>{{ t('tab_cloud_tools') }}</span>
        </button>
      </div>

      <!-- Body Container -->
      <div class="p-6 overflow-y-auto flex-1 space-y-4">
        <ResponsivePreview
          v-if="activeTab === 'responsive'"
          :sites="sites"
          :initial-domain="initialDomain"
          :t="t"
        />
        <ContrastChecker
          v-else-if="activeTab === 'contrast'"
          :t="t"
        />
        <DesignResources
          v-else-if="activeTab === 'cloud'"
          :t="t"
        />
      </div>

      <!-- Footer -->
      <div
        class="px-6 py-3 border-t border-slate-100 dark:border-slate-800 flex justify-end flex-shrink-0 bg-slate-50/50 dark:bg-slate-900/50"
      >
        <button
          type="button"
          class="px-4 py-2 text-xs font-semibold rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-600 transition-colors cursor-pointer"
          @click="closeModal"
        >
          {{ t('close_btn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import {
  Layout,
  Smartphone,
  Palette,
  ExternalLink,
  X,
} from 'lucide-vue-next'
import type { SiteItem } from '../types'
import ResponsivePreview from './design-tools/ResponsivePreview.vue'
import ContrastChecker from './design-tools/ContrastChecker.vue'
import DesignResources from './design-tools/DesignResources.vue'

interface Props {
  modelValue: boolean
  sites?: SiteItem[]
  initialDomain?: string
  t: (key: string, params?: Record<string, string | number | undefined>) => string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  sites: () => [],
  initialDomain: '',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
}>()

const activeTab = ref<'responsive' | 'contrast' | 'cloud'>('responsive')

function closeModal() {
  emit('update:modelValue', false)
  emit('close')
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.modelValue) {
    closeModal()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>
