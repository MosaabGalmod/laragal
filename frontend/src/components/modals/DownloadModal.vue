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
        <template
          v-for="rt in displayedRuntimes"
          :key="rt.name"
        >
          <div
            class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-800 space-y-2 transition-all"
            :class="{
              'border-rose-300 dark:border-rose-900/60 bg-rose-50/20 dark:bg-rose-950/20':
                (runtimeUpdates[rt.name]?.has_update && runtimeUpdates[rt.name]?.is_security_fix) ||
                shouldShowEol(rt.name),
            }"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="space-y-0.5">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <h4 class="text-sm font-bold text-slate-800 dark:text-slate-200">
                    {{ rt.label }}
                  </h4>
                  <span
                    v-if="rt.is_feed_only || runtimeUpdates[rt.name]?.is_new"
                    class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-blue-100 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-900"
                  >
                    <Globe class="w-2.5 h-2.5" /> {{ t('runtime_new_from_feed') }}
                  </span>
                  <span
                    v-else-if="shouldShowEol(rt.name)"
                    class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-rose-100 text-rose-700 dark:bg-rose-950/80 dark:text-rose-300 border border-rose-200 dark:border-rose-900"
                    :title="t('runtime_eol_tooltip')"
                  >
                    <AlertTriangle class="w-2.5 h-2.5" /> {{ t('runtime_eol_badge') }}
                  </span>
                  <span
                    v-else-if="shouldShowSecurityOnly(rt.name)"
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
                  <span
                    v-if="runtimeUpdates[rt.name]?.major_change && runtimeUpdates[rt.name]?.has_update"
                    class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-200 border border-amber-300 dark:border-amber-800"
                  >
                    <AlertTriangle class="w-2.5 h-2.5" /> {{ t('runtime_major_change') }}
                  </span>
                </div>
                <p class="text-[11px] text-slate-400 font-mono">
                  {{ getRuntimeDescription(rt) }}
                </p>
                <p
                  v-if="rt.name === 'mysql' && getMysqlEolDate(rt.name)"
                  class="text-[11px] text-slate-500 dark:text-slate-400 flex items-center gap-1 mt-0.5"
                >
                  <span v-if="currentLang === 'ar'">
                    <template v-if="runtimeUpdates[rt.name]?.current_version">
                      {{ t('runtime_installed') }} <bdi dir="ltr">v{{ runtimeUpdates[rt.name].current_version }}</bdi> — 
                    </template>
                    {{ getSupportedUntilPrefix() }} <bdi dir="ltr">{{ formatEolDate(getMysqlEolDate(rt.name), 'ar') }}</bdi>
                  </span>
                  <span v-else>
                    <template v-if="runtimeUpdates[rt.name]?.current_version">
                      {{ t('runtime_installed') }} v{{ runtimeUpdates[rt.name].current_version }} — 
                    </template>
                    {{ t('runtime_supported_until', { date: formatEolDate(getMysqlEolDate(rt.name), 'en') }) }}
                  </span>
                </p>
                <p
                  v-if="runtimeUpdates[rt.name]?.compat_note"
                  class="text-[11px] text-amber-700 dark:text-amber-400 font-medium flex items-center gap-1 mt-1"
                >
                  <AlertTriangle class="w-3 h-3 flex-shrink-0" />
                  <span>{{ runtimeUpdates[rt.name].compat_note }}</span>
                </p>
                <p
                  v-if="runtimeUpdates[rt.name]?.source_status === 'catalog_fallback'"
                  class="text-[11px] text-slate-500 dark:text-slate-400 flex items-center gap-1 mt-1"
                  :title="runtimeUpdates[rt.name]?.source_error || undefined"
                >
                  <Info class="w-3 h-3 flex-shrink-0 text-slate-400 dark:text-slate-500" />
                  <span>{{ t('runtime_source_unavailable') }}</span>
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
                    class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 cursor-pointer text-white shadow-xs"
                    :class="
                      runtimeUpdates[rt.name]?.major_change
                        ? 'bg-amber-600 hover:bg-amber-700'
                        : 'bg-emerald-600 hover:bg-emerald-700'
                    "
                    @click="handleRuntimeAction(rt)"
                  >
                    <Download class="w-3.5 h-3.5" />
                    <span>
                      {{
                        runtimeUpdates[rt.name]?.major_change
                          ? t('runtime_major_upgrade_btn')
                          : t('upgrade_one_click')
                      }}
                    </span>
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
                    @click="handleRuntimeAction(rt)"
                  >
                    {{ t('runtime_reinstall') }}
                  </button>
                </template>
                <template v-else>
                  <button
                    type="button"
                    class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 cursor-pointer bg-blue-600 hover:bg-blue-700 text-white"
                    @click="handleRuntimeAction(rt)"
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

          <!-- Merged MariaDB newer series card rendered directly after mysql card -->
          <div
            v-if="rt.name === 'mysql' && mariadbSeriesUpdates.length > 0"
            class="p-3 rounded-xl border border-amber-200 dark:border-amber-900/60 bg-amber-50/20 dark:bg-amber-950/20 space-y-2 transition-all"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="space-y-1.5 flex-1 min-w-0">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <h4 class="text-sm font-bold text-slate-800 dark:text-slate-200">
                    {{ t('mariadb_newer_series') }}
                  </h4>
                  <span
                    class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-200 border border-amber-300 dark:border-amber-800"
                  >
                    <AlertTriangle class="w-2.5 h-2.5" /> {{ t('runtime_major_change') }}
                  </span>
                  <span
                    class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-blue-100 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-900"
                  >
                    <Globe class="w-2.5 h-2.5" /> {{ t('runtime_new_from_feed') }}
                  </span>
                </div>
                <p class="text-[11px] text-slate-400 font-mono">
                  {{ t('service_mysql_desc') }}
                </p>
                <div class="pt-0.5">
                  <select
                    v-model="selectedMariaDBSeries"
                    class="w-full text-xs rounded-lg px-2.5 py-1.5 bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 text-slate-800 dark:text-slate-200 focus:outline-hidden focus:ring-1 focus:ring-amber-500 font-sans cursor-pointer rtl:text-right ltr:text-left"
                    :aria-label="t('mariadb_newer_series')"
                  >
                    <option
                      v-for="opt in mariaDBSeriesOptions"
                      :key="opt.value"
                      :value="opt.value"
                    >
                      {{ opt.label }}
                    </option>
                  </select>
                </div>
              </div>
              <div class="flex items-center gap-2 flex-shrink-0 self-end sm:self-center">
                <template v-if="downloadingRuntime === selectedMariaDBSeries">
                  <button
                    type="button"
                    disabled
                    class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 opacity-50 bg-amber-600 text-white cursor-not-allowed"
                  >
                    <Download class="w-3.5 h-3.5 animate-bounce" />
                    <span>{{ downloadProgress[selectedMariaDBSeries] || 0 }}%</span>
                  </button>
                </template>
                <template v-else>
                  <button
                    type="button"
                    class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 cursor-pointer text-white shadow-xs bg-amber-600 hover:bg-amber-700"
                    @click="handleMariaDBSeriesUpgrade"
                  >
                    <Download class="w-3.5 h-3.5" />
                    <span>{{ t('runtime_major_upgrade_btn') }}</span>
                  </button>
                </template>
              </div>
            </div>
            <div
              v-if="downloadingRuntime === selectedMariaDBSeries"
              class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 overflow-hidden"
            >
              <div
                class="bg-amber-600 h-1.5 rounded-full transition-all duration-200"
                :style="{ width: `${downloadProgress[selectedMariaDBSeries] || 0}%` }"
              ></div>
            </div>
          </div>
        </template>

        <!-- Fallback if mysql was not in displayed runtimes -->
        <div
          v-if="!hasMysqlInDisplayed && mariadbSeriesUpdates.length > 0"
          class="p-3 rounded-xl border border-amber-200 dark:border-amber-900/60 bg-amber-50/20 dark:bg-amber-950/20 space-y-2 transition-all"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="space-y-1.5 flex-1 min-w-0">
              <div class="flex items-center gap-1.5 flex-wrap">
                <h4 class="text-sm font-bold text-slate-800 dark:text-slate-200">
                  {{ t('mariadb_newer_series') }}
                </h4>
                <span
                  class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-200 border border-amber-300 dark:border-amber-800"
                >
                  <AlertTriangle class="w-2.5 h-2.5" /> {{ t('runtime_major_change') }}
                </span>
                <span
                  class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-blue-100 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-900"
                >
                  <Globe class="w-2.5 h-2.5" /> {{ t('runtime_new_from_feed') }}
                </span>
              </div>
              <p class="text-[11px] text-slate-400 font-mono">
                {{ t('service_mysql_desc') }}
              </p>
              <div class="pt-0.5">
                <select
                  v-model="selectedMariaDBSeries"
                  class="w-full text-xs rounded-lg px-2.5 py-1.5 bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 text-slate-800 dark:text-slate-200 focus:outline-hidden focus:ring-1 focus:ring-amber-500 font-sans cursor-pointer rtl:text-right ltr:text-left"
                  :aria-label="t('mariadb_newer_series')"
                >
                  <option
                    v-for="opt in mariaDBSeriesOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </option>
                </select>
              </div>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0 self-end sm:self-center">
              <template v-if="downloadingRuntime === selectedMariaDBSeries">
                <button
                  type="button"
                  disabled
                  class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 opacity-50 bg-amber-600 text-white cursor-not-allowed"
                >
                  <Download class="w-3.5 h-3.5 animate-bounce" />
                  <span>{{ downloadProgress[selectedMariaDBSeries] || 0 }}%</span>
                </button>
              </template>
              <template v-else>
                <button
                  type="button"
                  class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 cursor-pointer text-white shadow-xs bg-amber-600 hover:bg-amber-700"
                  @click="handleMariaDBSeriesUpgrade"
                >
                  <Download class="w-3.5 h-3.5" />
                  <span>{{ t('runtime_major_upgrade_btn') }}</span>
                </button>
              </template>
            </div>
          </div>
          <div
            v-if="downloadingRuntime === selectedMariaDBSeries"
            class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 overflow-hidden"
          >
            <div
              class="bg-amber-600 h-1.5 rounded-full transition-all duration-200"
              :style="{ width: `${downloadProgress[selectedMariaDBSeries] || 0}%` }"
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

    <!-- ================= MAJOR CHANGE CONFIRMATION MODAL ================= -->
    <div
      v-if="confirmMajorChange"
      class="fixed inset-0 z-60 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs"
      role="dialog"
      aria-modal="true"
      aria-labelledby="major-change-modal-title"
      tabindex="-1"
    >
      <div
        class="w-full max-w-sm rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-5 space-y-4 animate-scale-in"
      >
        <div class="flex items-center gap-2">
          <AlertTriangle class="w-5 h-5 text-amber-500 flex-shrink-0" />
          <h3 id="major-change-modal-title" class="font-bold text-base text-slate-900 dark:text-white">
            {{ t('runtime_major_change_confirm_title') }}
          </h3>
        </div>
        <p class="text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
          {{ t('runtime_major_change_warning') }}
        </p>
        <p
          v-if="isMySQLFamily(confirmMajorChange.name)"
          class="text-xs text-amber-800 dark:text-amber-200 bg-amber-50 dark:bg-amber-950/40 p-2.5 rounded-lg border border-amber-200 dark:border-amber-900/50 leading-relaxed"
        >
          {{ t('runtime_major_change_mysql_extra') }}
        </p>
        <p
          v-if="confirmMajorChange.name === 'adminer'"
          class="text-xs text-amber-800 dark:text-amber-200 bg-amber-50 dark:bg-amber-950/40 p-2.5 rounded-lg border border-amber-200 dark:border-amber-900/50 leading-relaxed"
        >
          {{ t('runtime_major_change_adminer_extra') }}
        </p>
        <p
          v-if="runtimeUpdates[confirmMajorChange.name]?.compat_note"
          class="text-xs text-amber-700 dark:text-amber-300 font-mono"
        >
          {{ runtimeUpdates[confirmMajorChange.name].compat_note }}
        </p>
        <div class="flex justify-end gap-2 pt-2">
          <button
            type="button"
            class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 cursor-pointer"
            @click="confirmMajorChange = null"
          >
            {{ t('cancel_btn') }}
          </button>
          <button
            type="button"
            class="px-3.5 py-1.5 text-xs font-bold rounded-lg bg-amber-600 hover:bg-amber-700 text-white cursor-pointer"
            @click="executeMajorChangeUpgrade"
          >
            {{ t('confirm_btn') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import {
  Download,
  RefreshCw,
  AlertTriangle,
  ShieldAlert,
  ArrowUpCircle,
  Check,
  Globe,
  Info,
} from 'lucide-vue-next'
import { useI18n } from '../../composables/useI18n'
import {
  useRuntimeCatalog,
  getMajorMinor,
  formatEolDate,
  getRecommendedMariaDBSeries,
} from '../../composables/useRuntimeCatalog'

interface Props {
  modelValue: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const { t, currentLang } = useI18n()
const {
  availableRuntimes,
  runtimeUpdates,
  mariadbSeriesUpdates,
  isCheckingUpdates,
  downloadProgress,
  downloadingRuntime,
  getCatalogItem,
  checkUpdatesNow,
  triggerDownloadRuntime,
} = useRuntimeCatalog()

const confirmMajorChange = ref<{ name: string; label: string } | null>(null)
const selectedMariaDBSeries = ref<string>('')

const displayedRuntimes = computed(() => {
  return availableRuntimes.value.filter(
    (r) => !r.name.startsWith('mariadb-') && (r.name === 'mysql' || !r.name.startsWith('mysql-')),
  )
})

const hasMysqlInDisplayed = computed(() => displayedRuntimes.value.some((r) => r.name === 'mysql'))

watch(
  mariadbSeriesUpdates,
  (updates) => {
    if (updates.length > 0) {
      const rec = getRecommendedMariaDBSeries(updates)
      if (rec) {
        selectedMariaDBSeries.value = rec.name
      }
    } else {
      selectedMariaDBSeries.value = ''
    }
  },
  { immediate: true },
)

const mariaDBSeriesOptions = computed(() => {
  return mariadbSeriesUpdates.value.map((u) => {
    const series = u.name.replace(/^(mariadb|mysql)-/, '')
    const formattedDate = formatEolDate(u.eol_date, currentLang.value)
    const label = formattedDate ? `${series} (LTS) — ${formattedDate}` : `${series} (LTS)`
    return {
      value: u.name,
      label,
      series,
      eol_date: u.eol_date,
    }
  })
})

function handleMariaDBSeriesUpgrade() {
  const targetName = selectedMariaDBSeries.value
  if (!targetName) return
  const series = targetName.replace(/^(mariadb|mysql)-/, '')
  confirmMajorChange.value = {
    name: targetName,
    label: `MariaDB ${series} (LTS)`,
  }
}

function getMysqlEolDate(name: string): string {
  if (name !== 'mysql') return ''
  return runtimeUpdates.value['mysql']?.eol_date || getCatalogItem('mysql')?.eol_date || ''
}

function getSupportedUntilPrefix(): string {
  return t('runtime_supported_until', { date: '' }).replace('{date}', '').trim()
}

function shouldShowEol(name: string): boolean {
  const item = getCatalogItem(name)
  if (!item?.eol) return false
  const update = runtimeUpdates.value[name]
  const curVer = update?.current_version
  if (!curVer) return true
  return getMajorMinor(curVer) === getMajorMinor(item.version)
}

function shouldShowSecurityOnly(name: string): boolean {
  const item = getCatalogItem(name)
  if (!item?.eol_date || item?.eol) return false
  const update = runtimeUpdates.value[name]
  const curVer = update?.current_version
  if (!curVer) return true
  return getMajorMinor(curVer) === getMajorMinor(item.version)
}

function getRuntimeDescription(rt: { name: string; desc_key: string }): string {
  if (!rt.desc_key || rt.desc_key === 'runtime_new_from_feed') {
    if (rt.name.startsWith('mariadb-') || rt.name.startsWith('mysql-')) {
      return t('service_mysql_desc')
    }
    if (rt.name.startsWith('php-')) {
      return t('service_php85_desc')
    }
    return t('service_mysql_desc')
  }
  return t(rt.desc_key)
}

function isMySQLFamily(name: string): boolean {
  const n = name.toLowerCase()
  return n === 'mysql' || n.startsWith('mariadb') || n.startsWith('mysql')
}

function handleRuntimeAction(rt: { name: string; label: string }) {
  const update = runtimeUpdates.value[rt.name]
  if (update?.major_change) {
    confirmMajorChange.value = {
      name: rt.name,
      label: rt.label,
    }
    return
  }
  triggerDownloadRuntime(rt.name)
}

async function executeMajorChangeUpgrade() {
  if (!confirmMajorChange.value) return
  const target = confirmMajorChange.value.name
  confirmMajorChange.value = null
  await triggerDownloadRuntime(target, true)
}

function closeModal() {
  emit('update:modelValue', false)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (confirmMajorChange.value) {
      confirmMajorChange.value = null
      return
    }
    if (props.modelValue) {
      closeModal()
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>
