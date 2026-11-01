<template>
  <div
    v-if="modelValue && site"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="diagnostics-modal-title"
    tabindex="-1"
    @click.self="closeModal"
  >
    <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in max-h-[90vh] overflow-y-auto">
      <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
        <h3 id="diagnostics-modal-title" class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2">
          <Wrench class="w-4 h-4 text-amber-500" />
          {{ t('modal_diag_title') }}
        </h3>
        <button
          type="button"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-lg leading-none cursor-pointer"
          :title="t('close_btn')"
          :aria-label="t('close_btn')"
          @click="closeModal"
        >
          &times;
        </button>
      </div>

      <!-- Project Badge & Retest button -->
      <div class="flex items-center justify-between p-3 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700/60">
        <div>
          <div class="font-bold text-xs text-slate-900 dark:text-white flex items-center gap-2">
            <span>{{ site.name || site.domain }}</span>
            <span dir="ltr" class="text-[11px] font-mono text-blue-600 dark:text-blue-400">({{ site.domain }})</span>
          </div>
          <p dir="ltr" class="text-[10px] font-mono text-slate-500 dark:text-slate-400 truncate max-w-xs mt-0.5">
            {{ site.path || site.root_path }}
          </p>
        </div>
        <button
          type="button"
          :disabled="isDiagnosing"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-50 flex items-center gap-1.5 cursor-pointer disabled:opacity-50"
          :title="t('recheck_btn')"
          @click="runDiagnostics(site)"
        >
          <RefreshCw class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" :class="{ 'animate-spin': isDiagnosing }" />
          <span>{{ t('recheck_btn') }}</span>
        </button>
      </div>

      <div v-if="isDiagnosing" class="py-12 text-center">
        <RefreshCw class="w-8 h-8 text-blue-600 animate-spin mx-auto mb-2" />
        <p class="text-xs text-slate-500">{{ t('diag_scanning_msg') }}</p>
      </div>

      <div v-else-if="diagResult" class="space-y-4">
        <!-- Overall Status Banner -->
        <div class="p-3.5 rounded-xl border flex items-center gap-3" :class="statusBannerClass">
          <CheckCircle2 v-if="diagResult.status === 'healthy'" class="w-5 h-5 flex-shrink-0 text-emerald-600 dark:text-emerald-400" />
          <AlertTriangle v-else-if="diagResult.status === 'warning'" class="w-5 h-5 flex-shrink-0 text-amber-600 dark:text-amber-400" />
          <AlertCircle v-else class="w-5 h-5 flex-shrink-0 text-rose-600 dark:text-rose-400" />
          <div class="text-xs">
            <span class="font-bold block">{{ statusLabel }}</span>
            <span class="text-[11px] opacity-80">
              <span>{{ t('type_prefix') }}</span> {{ diagResult.project_type }} • {{ diagResult.php_version || 'FastCGI/Proxy' }}
            </span>
          </div>
        </div>

        <!-- Checks Checklist -->
        <div class="rounded-xl border border-slate-200 dark:border-slate-800 divide-y divide-slate-100 dark:divide-slate-800 text-xs">
          <!-- 1. DocRoot check -->
          <div class="p-3 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Check v-if="diagResult.doc_root_exists" class="w-4 h-4 text-emerald-500" />
              <X v-else class="w-4 h-4 text-rose-500" />
              <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_docroot_check') }}</span>
            </div>
            <span dir="ltr" class="font-mono text-[11px] text-slate-500 max-w-[200px] truncate" :title="diagResult.doc_root">
              {{ diagResult.doc_root }}
            </span>
          </div>

          <!-- 2. Index file check -->
          <div class="p-3 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Check v-if="diagResult.has_index_file" class="w-4 h-4 text-emerald-500" />
              <AlertTriangle v-else class="w-4 h-4 text-amber-500" />
              <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_index_check') }}</span>
            </div>
            <span
              dir="ltr"
              class="font-mono text-[11px]"
              :class="diagResult.has_index_file ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'"
            >
              {{ diagResult.index_file || t('missing_label') }}
            </span>
          </div>

          <!-- 3. .env file check (PHP / Laravel) -->
          <div
            v-if="diagResult.project_type === 'laravel' || diagResult.project_type === 'php' || diagResult.project_type === 'wordpress'"
            class="p-3 flex items-center justify-between"
          >
            <div class="flex items-center gap-2">
              <Check v-if="diagResult.has_env_file" class="w-4 h-4 text-emerald-500" />
              <AlertCircle v-else class="w-4 h-4 text-amber-500" />
              <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_env_check') }}</span>
            </div>
            <div class="flex items-center gap-2">
              <span v-if="diagResult.has_env_file" class="text-emerald-600 dark:text-emerald-400 font-bold">
                {{ t('present_label') }}
              </span>
              <button
                v-else
                type="button"
                :disabled="isFixingEnv"
                class="px-2 py-1 rounded bg-amber-100 hover:bg-amber-200 text-amber-800 dark:bg-amber-950 dark:hover:bg-amber-900 dark:text-amber-200 font-bold text-[11px] transition-colors cursor-pointer disabled:opacity-50"
                @click="triggerFixEnv"
              >
                <span v-if="isFixingEnv">{{ t('creating_env') }}</span>
                <span v-else>{{ t('fix_env_btn') }}</span>
              </button>
            </div>
          </div>

          <!-- 4. Dependencies check -->
          <div
            v-if="diagResult.project_type === 'laravel' || diagResult.project_type === 'php' || diagResult.project_type === 'node'"
            class="p-3 flex items-center justify-between"
          >
            <div class="flex items-center gap-2">
              <Check v-if="diagResult.has_vendor_dir || diagResult.has_node_modules" class="w-4 h-4 text-emerald-500" />
              <AlertTriangle v-else class="w-4 h-4 text-amber-500" />
              <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_deps_check') }}</span>
            </div>
            <span
              class="font-mono text-[11px]"
              :class="diagResult.has_vendor_dir || diagResult.has_node_modules ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'"
            >
              {{ diagResult.has_vendor_dir ? t('vendor_installed') : diagResult.has_node_modules ? t('node_modules_installed') : t('not_installed_deps') }}
            </span>
          </div>

          <!-- 5. Port Connection check -->
          <div class="p-3 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Check v-if="diagResult.is_port_active" class="w-4 h-4 text-emerald-500" />
              <X v-else class="w-4 h-4 text-rose-500" />
              <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_port_check') }}</span>
            </div>
            <span
              dir="ltr"
              class="font-mono text-[11px]"
              :class="diagResult.is_port_active ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
            >
              {{ diagResult.php_port > 0 ? `PHP FastCGI :${diagResult.php_port}` : `Proxy :${diagResult.proxy_port}` }}
              ({{ diagResult.is_port_active ? t('port_connected') : t('port_disconnected') }})
            </span>
          </div>
        </div>

        <!-- Discovered Databases & Schemas -->
        <div
          v-if="discoveredDatabases && discoveredDatabases.length > 0"
          class="p-3.5 rounded-xl bg-indigo-50/70 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800/60 space-y-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Database class="w-4 h-4 text-indigo-600 dark:text-indigo-400 flex-shrink-0" />
              <span class="text-xs font-bold text-indigo-900 dark:text-indigo-200">{{ t('db_discovery_title') }}</span>
            </div>
            <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-indigo-100 dark:bg-indigo-900/60 text-indigo-700 dark:text-indigo-300">
              {{ discoveredDatabases.length }} {{ t('discovered_badge') }}
            </span>
          </div>

          <p class="text-xs text-indigo-700 dark:text-indigo-300 leading-relaxed">{{ t('db_discovery_prompt') }}</p>

          <div
            v-if="dbImportSuccess"
            class="p-2.5 rounded-lg bg-emerald-100 dark:bg-emerald-950/60 border border-emerald-300 dark:border-emerald-800 text-xs text-emerald-800 dark:text-emerald-200 flex items-center gap-2 font-medium"
          >
            <CheckCircle2 class="w-4 h-4 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
            <span>{{ t('db_import_success') }}</span>
          </div>

          <div
            v-for="(dbItem, idx) in discoveredDatabases"
            :key="idx"
            class="p-3 rounded-xl bg-white dark:bg-slate-900 border border-indigo-100 dark:border-indigo-900/50 shadow-xs space-y-2"
          >
            <div class="flex items-start justify-between gap-2">
              <div>
                <div class="flex items-center gap-1.5">
                  <span class="font-mono text-xs font-bold text-slate-800 dark:text-slate-100">{{ dbItem.file_name || dbItem.target_db }}</span>
                  <span
                    class="text-[10px] px-1.5 py-0.5 rounded font-mono uppercase"
                    :class="dbItem.type === 'schema_file' ? 'bg-amber-100 dark:bg-amber-950 text-amber-700 dark:text-amber-300' : 'bg-blue-100 dark:bg-blue-950 text-blue-700 dark:text-blue-300'"
                  >
                    {{ dbItem.type === 'schema_file' ? t('db_schema_file') : t('db_legacy_dir') }}
                  </span>
                </div>
                <span class="text-[11px] text-slate-500 block mt-0.5">{{ dbItem.description }}</span>
              </div>
              <button
                type="button"
                :disabled="isImportingDB"
                class="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs transition-colors shadow-xs flex items-center gap-1.5 flex-shrink-0 cursor-pointer disabled:opacity-50"
                @click="importDiscoveredDatabase(dbItem)"
              >
                <RefreshCw v-if="isImportingDB" class="w-3.5 h-3.5 animate-spin" />
                <Database v-else class="w-3.5 h-3.5" />
                <span>{{ isImportingDB ? t('db_importing') : t('db_import_btn') }}</span>
              </button>
            </div>

            <!-- Detected tables chips -->
            <div v-if="dbItem.found_tables && dbItem.found_tables.length > 0" class="pt-1.5 border-t border-slate-100 dark:border-slate-800/80">
              <span class="text-[10px] text-slate-400 font-semibold block mb-1">{{ t('db_found_tables') }} ({{ dbItem.found_tables.length }}):</span>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="tbl in dbItem.found_tables.slice(0, 6)"
                  :key="tbl"
                  class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300"
                >
                  {{ tbl }}
                </span>
                <span v-if="dbItem.found_tables.length > 6" class="text-[10px] text-slate-400 self-center">
                  +{{ dbItem.found_tables.length - 6 }} {{ t('other_count') }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Issues & Suggestions -->
        <div v-if="diagResult.issues && diagResult.issues.length > 0" class="p-3.5 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 space-y-1.5">
          <span class="text-xs font-bold text-rose-700 dark:text-rose-300 block">{{ t('issues_found_title') }}</span>
          <ul class="text-xs text-rose-600 dark:text-rose-400 space-y-1 list-disc list-inside">
            <li v-for="(issue, idx) in diagResult.issues" :key="idx">
              <bdi>{{ renderUIMessage(issue) }}</bdi>
            </li>
          </ul>
        </div>

        <div v-if="diagResult.suggestions && diagResult.suggestions.length > 0" class="p-3.5 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/50 space-y-1.5">
          <span class="text-xs font-bold text-blue-700 dark:text-blue-300 block">{{ t('suggestions_title') }}</span>
          <ul class="text-xs text-blue-600 dark:text-blue-400 space-y-1 list-disc list-inside">
            <li v-for="(sug, idx) in diagResult.suggestions" :key="idx">
              <bdi>{{ renderUIMessage(sug) }}</bdi>
            </li>
          </ul>
        </div>
      </div>

      <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
        <button
          type="button"
          class="px-3.5 py-2 text-xs font-semibold rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors flex items-center gap-1.5 cursor-pointer"
          @click="openSettings"
        >
          <Settings class="w-3.5 h-3.5" />
          <span>{{ t('edit_project_settings_btn') }}</span>
        </button>
        <button
          type="button"
          class="px-5 py-2 text-xs font-bold rounded-xl bg-blue-600 text-white hover:bg-blue-700 transition-colors cursor-pointer"
          @click="closeModal"
        >
          {{ t('close_btn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import {
  Wrench,
  RefreshCw,
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  Check,
  X,
  Database,
  Settings,
} from 'lucide-vue-next'
import type { SiteItem } from '../../types'
import { useI18n } from '../../composables/useI18n'
import { useToast } from '../../composables/useToast'
import { useRuntimeCatalog } from '../../composables/useRuntimeCatalog'

interface Props {
  modelValue: boolean
  site: SiteItem | null
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  site: null,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'openSettings', site: SiteItem): void
}>()

const { t, renderUIMessage } = useI18n()
const { showToast } = useToast()
const { getPHPPort } = useRuntimeCatalog()

const diagResult = ref<any | null>(null)
const isDiagnosing = ref(false)
const isFixingEnv = ref(false)
const discoveredDatabases = ref<any[]>([])
const isImportingDB = ref(false)
const dbImportSuccess = ref(false)

const statusBannerClass = computed(() => {
  if (diagResult.value?.status === 'healthy') {
    return 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800 text-emerald-800 dark:text-emerald-200'
  }
  if (diagResult.value?.status === 'warning') {
    return 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-200'
  }
  return 'bg-rose-50 dark:bg-rose-950/40 border-rose-200 dark:border-rose-800 text-rose-800 dark:text-rose-200'
})

const statusLabel = computed(() => {
  if (diagResult.value?.status === 'healthy') return t('diag_status_healthy')
  if (diagResult.value?.status === 'warning') return t('diag_status_warning')
  return t('diag_status_error')
})

function closeModal() {
  emit('update:modelValue', false)
}

function openSettings() {
  if (props.site) {
    emit('openSettings', props.site)
  }
  closeModal()
}

async function runDiagnostics(siteItem: SiteItem) {
  const targetPath = siteItem.path || siteItem.root_path || siteItem.domain
  isDiagnosing.value = true
  discoveredDatabases.value = []
  dbImportSuccess.value = false

  const w = window as any
  if (w.go?.main?.App?.DiagnoseProject) {
    try {
      const res = await w.go.main.App.DiagnoseProject(targetPath)
      diagResult.value = res
      if (w.go?.main?.App?.ScanProjectDatabases) {
        const dbs = await w.go.main.App.ScanProjectDatabases(targetPath)
        discoveredDatabases.value = dbs || []
      }
    } catch (e: any) {
      console.warn('Diagnostics error:', e)
    }
  } else {
    diagResult.value = {
      project_path: targetPath,
      domain: siteItem.domain,
      project_type: siteItem.project_type || 'php',
      php_version: siteItem.php_version || 'PHP 8.3',
      php_port: siteItem.php_port || getPHPPort(siteItem.php_version),
      proxy_port: siteItem.proxy_port || 0,
      doc_root: siteItem.root_path,
      doc_root_exists: true,
      has_index_file: true,
      index_file: 'index.php',
      has_env_file: true,
      has_vendor_dir: true,
      is_port_active: true,
      status: 'healthy',
      issues: [],
      suggestions: [],
    }
  }
  isDiagnosing.value = false
}

async function triggerFixEnv() {
  if (!props.site) return
  const targetPath = props.site.path || props.site.root_path || props.site.domain
  isFixingEnv.value = true

  const w = window as any
  if (w.go?.main?.App?.FixProjectEnv) {
    try {
      const ok = await w.go.main.App.FixProjectEnv(targetPath)
      if (ok) {
        showToast(t('env_fixed_success'))
        await runDiagnostics(props.site)
      } else {
        showToast(t('env_fix_failed'))
      }
    } catch (e: any) {
      showToast(t('error_fixing_env', { error: e?.message || e }))
    }
  }
  isFixingEnv.value = false
}

async function importDiscoveredDatabase(candidate: any) {
  if (!props.site) return
  const targetPath = props.site.path || props.site.root_path || props.site.domain
  isImportingDB.value = true

  const w = window as any
  try {
    if (w.go?.main?.App?.ImportProjectDatabase) {
      const ok = await w.go.main.App.ImportProjectDatabase(targetPath, candidate)
      if (ok) {
        dbImportSuccess.value = true
        showToast(t('db_import_success'))
        await runDiagnostics(props.site)
      }
    }
  } catch (err: any) {
    console.error('Import database error:', err)
    showToast(t('error_importing_db', { error: err?.message || err }))
  } finally {
    isImportingDB.value = false
  }
}

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen && props.site) {
      runDiagnostics(props.site)
    }
  },
)

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
