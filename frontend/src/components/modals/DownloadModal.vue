<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="download-modal-title"
    tabindex="-1"
    @click.self="closeModal"
  >
    <div
      class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in"
    >
      <div
        class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3"
      >
        <div class="flex items-center gap-2">
          <Download class="w-4 h-4 text-blue-600" />
          <h3
            id="download-modal-title"
            class="font-bold text-base text-slate-900 dark:text-white"
          >
            {{ t('modal_download_title') }}
          </h3>
        </div>
        <div class="flex items-center gap-2">
          <button
            type="button"
            :disabled="isCheckingUpdates"
            class="px-2.5 py-1 text-xs font-semibold rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 transition-colors flex items-center gap-1.5 cursor-pointer disabled:opacity-50"
            :title="t('check_updates_title')"
            @click="() => checkUpdatesNow()"
          >
            <RefreshCw
              class="w-3 h-3 text-blue-500"
              :class="{ 'animate-spin': isCheckingUpdates }"
            />
            <span>{{ isCheckingUpdates ? t('checking_updates') : t('check_updates_btn') }}</span>
          </button>
          <button
            type="button"
            class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 font-bold text-lg leading-none cursor-pointer"
            :title="t('close_btn')"
            :aria-label="t('close_btn')"
            @click="closeModal"
          >
            &times;
          </button>
        </div>
      </div>

      <div class="space-y-3 max-h-[60vh] overflow-y-auto px-0.5">
        <div
          v-for="rt in availableRuntimes"
          :key="rt.name"
          class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 space-y-2 transition-all"
          :class="{
            'border-rose-300 dark:border-rose-900/60 bg-rose-50/20 dark:bg-rose-950/20':
              (runtimeUpdates[rt.name]?.has_update && runtimeUpdates[rt.name]?.is_security_fix) ||
              getCatalogItem(rt.name)?.eol,
          }"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="space-y-0.5">
              <div class="flex items-center gap-1.5 flex-wrap">
                <h4 class="text-sm font-bold text-slate-800 dark:text-slate-200">
                  {{ rt.label }}
                </h4>
                <span
                  v-if="getCatalogItem(rt.name)?.eol"
                  class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-rose-100 text-rose-700 dark:bg-rose-950/80 dark:text-rose-300 border border-rose-200 dark:border-rose-900"
                  :title="t('runtime_eol_tooltip')"
                >
                  <AlertTriangle class="w-2.5 h-2.5" /> {{ t('runtime_eol_badge') }}
                </span>
                <span
                  v-else-if="getCatalogItem(rt.name)?.eol_date"
                  class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-700 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-200 dark:border-amber-900"
                >
                  {{ t('runtime_security_only', { date: getCatalogItem(rt.name)?.eol_date }) }}
                </span>
                <span
                  v-if="runtimeUpdates[rt.name]?.has_update && runtimeUpdates[rt.name]?.is_security_fix"
                  class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-rose-100 text-rose-700 dark:bg-rose-950/80 dark:text-rose-300 border border-rose-200 dark:border-rose-900"
                >
                  <ShieldAlert class="w-2.5 h-2.5" /> {{ t('runtime_security_upgrade') }}
                  <bdi dir="ltr">v{{ runtimeUpdates[rt.name].latest_version }}</bdi>
                </span>
                <span
                  v-else-if="runtimeUpdates[rt.name]?.has_update"
                  class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-700 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-200 dark:border-amber-900"
                >
                  <ArrowUpCircle class="w-2.5 h-2.5" /> {{ t('runtime_update_available') }}
                  <bdi dir="ltr">v{{ runtimeUpdates[rt.name].latest_version }}</bdi>
                </span>
              </div>
              <p class="text-[11px] text-slate-400 font-mono">
                {{ t(rt.desc_key) }}
              </p>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <template v-if="downloadingRuntime === rt.name">
                <button
                  type="button"
                  disabled
                  class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 opacity-50 bg-blue-600 text-white cursor-not-allowed"
                >
                  <Download class="w-3.5 h-3.5 animate-bounce" />
                  <span>{{ downloadProgress[rt.name] || 0 }}%</span>
                </button>
              </template>
              <template v-else-if="runtimeUpdates[rt.name]?.has_update">
                <button
                  type="button"
                  class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 cursor-pointer bg-emerald-600 hover:bg-emerald-700 text-white shadow-xs"
                  @click="triggerDownloadRuntime(rt.name)"
                >
                  <Download class="w-3.5 h-3.5" />
                  <span>{{ t('upgrade_one_click') }}</span>
                </button>
              </template>
              <template v-else-if="runtimeUpdates[rt.name]?.current_version">
                <span
                  class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 select-none"
                >
                  <Check class="w-3.5 h-3.5 text-emerald-500" />
                  <span>{{ t('runtime_installed') }}</span>
                  <bdi dir="ltr">v{{ runtimeUpdates[rt.name].current_version }}</bdi>
                </span>
                <button
                  type="button"
                  class="text-[11px] font-medium text-slate-400 hover:text-slate-600 dark:text-slate-500 dark:hover:text-slate-300 underline underline-offset-2 cursor-pointer transition-colors"
                  @click="triggerDownloadRuntime(rt.name)"
                >
                  {{ t('runtime_reinstall') }}
                </button>
              </template>
              <template v-else>
                <button
                  type="button"
                  class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 cursor-pointer bg-blue-600 hover:bg-blue-700 text-white"
                  @click="triggerDownloadRuntime(rt.name)"
                >
                  <Download class="w-3.5 h-3.5" />
                  <span>{{ t('download_now') }}</span>
                </button>
              </template>
            </div>
          </div>
          <div
            v-if="downloadingRuntime === rt.name"
            class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 overflow-hidden"
          >
            <div
              class="bg-blue-600 h-1.5 rounded-full transition-all duration-200"
              :style="{ width: `${downloadProgress[rt.name] || 0}%` }"
            ></div>
          </div>
        </div>
      </div>

      <div
        class="pt-3 border-t border-slate-100 dark:border-slate-800 flex justify-end"
      >
        <button
          type="button"
          class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 cursor-pointer"
          @click="closeModal"
        >
          {{ t('cancel_btn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import {
  Download,
  RefreshCw,
  AlertTriangle,
  ShieldAlert,
  ArrowUpCircle,
  Check,
} from 'lucide-vue-next'
import { useI18n } from '../../composables/useI18n'
import { useRuntimeCatalog } from '../../composables/useRuntimeCatalog'

interface Props {
  modelValue: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const { t } = useI18n()
const {
  availableRuntimes,
  runtimeUpdates,
  isCheckingUpdates,
  downloadProgress,
  downloadingRuntime,
  getCatalogItem,
  checkUpdatesNow,
  triggerDownloadRuntime,
} = useRuntimeCatalog()

function closeModal() {
  emit('update:modelValue', false)
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
