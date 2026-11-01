<template>
  <div
    v-if="modelValue && site"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="project-settings-modal-title"
    tabindex="-1"
    @click.self="closeModal"
  >
    <div
      class="w-full max-w-lg rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in"
    >
      <div
        class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3"
      >
        <h3
          id="project-settings-modal-title"
          class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2"
        >
          <Settings class="w-4 h-4 text-blue-500" />
          {{ t('modal_settings_title') }}
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

      <div
        class="p-2.5 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700/60 space-y-1"
      >
        <div class="flex items-center justify-between text-xs">
          <span class="font-bold text-slate-700 dark:text-slate-200">{{ site.name || site.domain }}</span>
          <span
            dir="ltr"
            class="font-mono text-[11px] text-blue-600 dark:text-blue-400"
          >{{ site.domain }}</span>
        </div>
        <p
          dir="ltr"
          class="text-[11px] font-mono text-slate-500 dark:text-slate-400 truncate"
        >
          {{ site.path || site.root_path }}
        </p>
      </div>

      <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
        {{ t('modal_settings_desc') }}
      </p>

      <form
        novalidate
        class="space-y-4"
        @submit.prevent="applyProjectSettings"
      >
        <!-- Project Type Selector -->
        <div>
          <label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
            {{ t('setting_project_type') }}
          </label>
          <select
            v-model="settingsForm.type"
            class="w-full px-3 py-2 rounded-xl text-xs border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">
              {{ t('project_type_auto') }}
            </option>
            <option value="laravel">
              {{ t('project_type_laravel') }}
            </option>
            <option value="wordpress">
              {{ t('project_type_wordpress') }}
            </option>
            <option value="php">
              {{ t('project_type_php') }}
            </option>
            <option value="node">
              {{ t('project_type_node') }}
            </option>
            <option value="go">
              {{ t('project_type_go') }}
            </option>
            <option value="rust">
              {{ t('project_type_rust') }}
            </option>
            <option value="dotnet">
              {{ t('project_type_dotnet') }}
            </option>
            <option value="python">
              {{ t('project_type_python') }}
            </option>
            <option value="flutter">
              {{ t('project_type_flutter') }}
            </option>
            <option value="static">
              {{ t('project_type_static') }}
            </option>
            <option value="non_web">
              {{ t('project_type_non_web') }}
            </option>
          </select>
        </div>

        <!-- Document Root Subfolder -->
        <div>
          <div class="flex items-center justify-between mb-1">
            <label class="text-xs font-bold text-slate-700 dark:text-slate-300">
              {{ t('setting_docroot') }}
            </label>
            <div class="flex items-center gap-1">
              <span class="text-[10px] text-slate-400">{{ t('quick_label') }}</span>
              <button
                type="button"
                class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 hover:bg-blue-100 dark:hover:bg-blue-900/40 text-slate-600 dark:text-slate-300 cursor-pointer"
                @click="settingsForm.doc_root = 'public'"
              >
                public
              </button>
              <button
                type="button"
                class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 hover:bg-blue-100 dark:hover:bg-blue-900/40 text-slate-600 dark:text-slate-300 cursor-pointer"
                @click="settingsForm.doc_root = 'dist'"
              >
                dist
              </button>
              <button
                type="button"
                class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 hover:bg-blue-100 dark:hover:bg-blue-900/40 text-slate-600 dark:text-slate-300 cursor-pointer"
                @click="settingsForm.doc_root = ''"
              >
                {{ t('root_dot') }}
              </button>
            </div>
          </div>
          <input
            v-model="settingsForm.doc_root"
            type="text"
            dir="ltr"
            class="w-full px-3 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            :placeholder="t('setting_docroot_placeholder')"
          >
          <p class="text-[10px] text-slate-400 mt-1">
            {{ t('setting_docroot_hint') }}
          </p>
        </div>

        <!-- Proxy Port (if applicable) -->
        <div v-if="settingsForm.type !== 'non_web' && settingsForm.type !== 'static'">
          <label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
            {{ t('setting_proxy_port') }}
          </label>
          <input
            v-model.number="settingsForm.proxy_port"
            type="number"
            dir="ltr"
            class="w-full px-3 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            :placeholder="t('setting_proxy_port_placeholder')"
          >
        </div>

        <!-- PHP Version Selector (if PHP type) -->
        <div
          v-if="
            settingsForm.type === 'laravel' ||
            settingsForm.type === 'wordpress' ||
            settingsForm.type === 'php' ||
            (!settingsForm.type && isPhpProject(site))
          "
        >
          <label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
            {{ t('custom_php_version') }}
          </label>
          <select
            v-model="settingsForm.php_version"
            dir="ltr"
            class="w-full px-3 py-2 rounded-xl text-xs border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">
              {{ t('php_default_global') }}
            </option>
            <option
              v-for="opt in availablePhpOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </option>
          </select>
          <div
            v-if="selectedSettingsPhpMeta?.eol"
            class="mt-1.5 flex items-center gap-1.5"
          >
            <span
              class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-rose-100 text-rose-700 dark:bg-rose-950/80 dark:text-rose-300 border border-rose-200 dark:border-rose-900"
              :title="t('runtime_eol_tooltip')"
            >
              <AlertTriangle class="w-2.5 h-2.5" /> {{ t('runtime_eol_badge') }}
            </span>
            <span class="text-[11px] text-rose-600 dark:text-rose-400">
              {{ t('runtime_eol_tooltip') }}
            </span>
          </div>
          <div
            v-else-if="selectedSettingsPhpMeta?.eol_date"
            class="mt-1.5 flex items-center gap-1.5"
          >
            <span
              class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-700 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-200 dark:border-amber-900"
            >
              {{ t('runtime_security_only', { date: selectedSettingsPhpMeta.eol_date }) }}
            </span>
          </div>
        </div>

        <!-- Ignore toggle -->
        <div class="pt-2">
          <label class="inline-flex items-center gap-2 cursor-pointer">
            <input
              v-model="settingsForm.is_ignored"
              type="checkbox"
              class="w-4 h-4 rounded text-blue-600 focus:ring-blue-500"
            >
            <span class="text-xs text-slate-600 dark:text-slate-400">{{ t('hide_folder_from_list') }}</span>
          </label>
        </div>

        <div
          class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-2"
        >
          <button
            type="button"
            class="px-4 py-2 text-xs font-semibold rounded-xl text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            @click="closeModal"
          >
            {{ t('cancel_btn') }}
          </button>
          <button
            type="submit"
            class="px-5 py-2 text-xs font-bold rounded-xl bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors cursor-pointer"
          >
            {{ t('save_settings_btn') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { Settings, AlertTriangle } from 'lucide-vue-next'
import type { SiteItem } from '../../types'
import { useI18n } from '../../composables/useI18n'
import { useToast } from '../../composables/useToast'
import {
  useRuntimeCatalog,
  extractPhpVersionNumber,
} from '../../composables/useRuntimeCatalog'

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
  (e: 'saved', updatedSites?: any[]): void
}>()

const { t } = useI18n()
const { showToast } = useToast()
const { phpCatalog, getPhpSelectOptionsPrefixed } = useRuntimeCatalog()

const settingsForm = ref({
  type: '',
  doc_root: '',
  proxy_port: 0,
  php_version: '',
  is_ignored: false,
})

const availablePhpOptions = computed(() =>
  getPhpSelectOptionsPrefixed(settingsForm.value.php_version || props.site?.php_version),
)

function getDocRootRelative(site: SiteItem): string {
  if (!site.path || !site.root_path || site.path === site.root_path) {
    return ''
  }
  const normPath = site.path.replace(/\\/g, '/').replace(/\/$/, '')
  const normRoot = site.root_path.replace(/\\/g, '/').replace(/\/$/, '')
  if (normRoot.startsWith(normPath)) {
    const sub = normRoot.slice(normPath.length).replace(/^\//, '')
    return sub ? `/${sub}` : ''
  }
  return ''
}

function isPhpProject(site: SiteItem): boolean {
  return Boolean(
    site.is_web &&
      (site.project_type === 'laravel' ||
        site.project_type === 'wordpress' ||
        site.project_type === 'php' ||
        (site.php_port || 0) > 0),
  )
}

const selectedSettingsPhpMeta = computed(() => {
  const ver = extractPhpVersionNumber(settingsForm.value.php_version)
  return phpCatalog.value.find(
    (item) => item.name === `php-${ver}` || item.name === ver,
  )
})

watch(
  () => props.site,
  (s) => {
    if (s) {
      const relDoc = getDocRootRelative(s).replace(/^\//, '')
      settingsForm.value = {
        type: s.project_type || '',
        doc_root: relDoc,
        proxy_port: s.proxy_port || 0,
        php_version: s.php_version
          ? `PHP ${extractPhpVersionNumber(s.php_version)}`
          : '',
        is_ignored: false,
      }
    }
  },
  { immediate: true },
)

function closeModal() {
  emit('update:modelValue', false)
}

async function applyProjectSettings() {
  if (!props.site) return
  const site = props.site
  const targetPath = site.path || site.root_path || site.domain

  const payload = {
    type: settingsForm.value.type,
    doc_root: settingsForm.value.doc_root.trim(),
    proxy_port: Number(settingsForm.value.proxy_port) || 0,
    php_version: settingsForm.value.php_version,
    is_ignored: settingsForm.value.is_ignored,
  }

  const w = window as any
  if (w.go?.main?.App?.SaveProjectSettings) {
    try {
      const updated = await w.go.main.App.SaveProjectSettings(targetPath, payload)
      showToast(t('settings_saved'))
      emit('saved', Array.isArray(updated) ? updated : undefined)
    } catch (e: any) {
      showToast(t('error_saving_settings', { error: e?.message || e }))
    }
  } else {
    showToast(t('settings_saved'))
    emit('saved')
  }

  closeModal()
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
