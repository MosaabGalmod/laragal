<template>
  <div
    class="flex flex-wrap md:flex-nowrap items-center justify-between gap-3 px-4 py-2.5 min-h-[56px] transition-colors motion-reduce:transition-none hover:bg-slate-50/80 dark:hover:bg-slate-800/40"
    :class="site.is_web ? 'text-slate-800 dark:text-slate-200' : 'text-slate-500 dark:text-slate-400 opacity-80'"
  >
    <!-- Start: Icon & Identity (Domain / Name + Path & DocRoot) -->
    <div class="flex items-center gap-3 min-w-0 flex-1 basis-64">
      <!-- Type icon in a small neutral square -->
      <div
        class="w-9 h-9 rounded-lg bg-slate-100 dark:bg-slate-800 flex items-center justify-center shrink-0"
        :class="site.is_web ? 'text-slate-600 dark:text-slate-300' : 'text-slate-400 dark:text-slate-500'"
      >
        <component :is="projectIcon" class="w-4 h-4" />
      </div>

      <!-- Text: Primary & Secondary lines -->
      <div class="min-w-0 flex-1">
        <!-- Primary line -->
        <div class="flex items-center gap-2">
          <a
            v-if="site.is_web"
            :href="'https://' + site.domain"
            target="_blank"
            dir="ltr"
            @click.prevent="emit('open-browser')"
            class="font-medium text-sm text-slate-900 dark:text-slate-100 hover:text-blue-600 dark:hover:text-blue-400 truncate transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 rounded"
            :title="site.domain"
          >
            {{ site.domain }}
          </a>
          <span
            v-else
            class="font-medium text-sm text-slate-700 dark:text-slate-300 truncate"
            :title="site.name || site.domain"
          >
            {{ site.name || site.domain }}
          </span>
        </div>

        <!-- Secondary line: path + doc root chip -->
        <div class="flex items-center gap-1.5 mt-0.5 min-w-0 text-[11px]">
          <span
            dir="ltr"
            class="font-mono text-slate-500 dark:text-slate-400 truncate"
            :title="site.path || site.root_path"
          >
            {{ site.path || site.root_path }}
          </span>
          <span
            v-if="docRootRelative"
            class="inline-flex items-center px-1.5 py-0.2 rounded text-[10px] font-mono font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700 shrink-0 select-none"
            :title="t('doc_root') + ': ' + docRootRelative"
          >
            <bdi dir="ltr">{{ docRootRelative }}</bdi>
          </span>
        </div>
      </div>
    </div>

    <!-- Middle: Type / Runtime Cell -->
    <div class="flex items-center gap-2 shrink-0">
      <!-- PHP Projects: Framework label + Version select -->
      <template v-if="isPhp">
        <div class="inline-flex items-center gap-1.5 text-xs">
          <span
            v-if="site.project_type === 'laravel'"
            class="font-semibold text-slate-700 dark:text-slate-300 text-xs"
          >
            <bdi>Laravel</bdi>
          </span>
          <span
            v-else-if="site.project_type === 'wordpress'"
            class="font-semibold text-slate-700 dark:text-slate-300 text-xs"
          >
            <bdi>WordPress</bdi>
          </span>
          <div dir="ltr" class="inline-flex items-center rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 px-2 py-0.5 text-xs font-mono shadow-2xs">
            <select
              :value="phpVersionNumber"
              @change="emit('change-php', $event)"
              class="bg-transparent text-blue-700 dark:text-blue-300 font-bold focus:outline-none cursor-pointer text-xs"
              :aria-label="t('change_php_hint')"
              :title="t('change_php_hint')"
            >
              <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.5">PHP 8.5</option>
              <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.4">PHP 8.4</option>
              <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.3">PHP 8.3</option>
              <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.2">PHP 8.2</option>
              <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.1">PHP 8.1</option>
            </select>
          </div>
        </div>
      </template>

      <!-- Proxy Projects: e.g. "Go · :8080" -->
      <template v-else-if="proxyRuntimeInfo">
        <span class="inline-flex items-center gap-1 text-xs font-mono px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
          <bdi>{{ proxyRuntimeInfo.name }}</bdi>
          <span class="text-slate-400">·</span>
          <bdi dir="ltr">:{{ proxyRuntimeInfo.port }}</bdi>
        </span>
      </template>

      <!-- Other Web Projects -->
      <template v-else-if="site.is_web">
        <span v-if="site.project_type === 'flutter'" class="text-xs font-mono px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
          <bdi>Flutter Web</bdi>
        </span>
        <span v-else-if="site.project_type === 'static'" class="text-xs font-medium px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
          {{ t('type_static') }}
        </span>
        <span v-else-if="site.project_type === 'node'" class="text-xs font-mono px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
          <bdi>Node.js</bdi>
        </span>
        <span v-else class="text-xs font-mono px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
          <bdi>{{ site.type_label || site.php_version || t('type_web') }}</bdi>
        </span>
      </template>

      <!-- Non-Web Projects: Local Folder Badge -->
      <template v-else>
        <span class="text-xs text-slate-500 dark:text-slate-400 px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700">
          {{ t('local_folder_badge') }}
        </span>
      </template>
    </div>

    <!-- End: Action Buttons -->
    <div class="flex items-center gap-1.5 shrink-0">
      <!-- Primary Action: Open in Browser (Web) or Open Folder (Non-Web) -->
      <a
        v-if="site.is_web"
        :href="'https://' + site.domain"
        target="_blank"
        @click.prevent="emit('open-browser')"
        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium border border-blue-200 dark:border-blue-800 bg-blue-50/80 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/60 transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('open_browser')"
        :aria-label="t('open_browser')"
      >
        <span>{{ t('open_browser') }}</span>
        <ExternalLink class="w-3.5 h-3.5" />
      </a>
      <button
        v-else
        type="button"
        @click="emit('open-folder')"
        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700/60 transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('open_folder_btn')"
        :aria-label="t('open_folder_btn')"
      >
        <span>{{ t('open_folder_btn') }}</span>
        <FolderOpen class="w-3.5 h-3.5 text-amber-500" />
      </button>

      <!-- Start/Stop App Button for Proxy Projects -->
      <button
        v-if="isProxyApp"
        type="button"
        @click="emit('toggle-app')"
        class="inline-flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium border transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :class="isRunning
          ? 'border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/60 text-amber-700 dark:text-amber-300 hover:bg-amber-100 dark:hover:bg-amber-900/60'
          : 'border-emerald-200 dark:border-emerald-800 bg-emerald-50 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-100 dark:hover:bg-emerald-900/60'"
        :title="isRunning ? t('stop_app_btn') : t('start_app_btn')"
        :aria-label="isRunning ? t('stop_app_btn') : t('start_app_btn')"
      >
        <Power class="w-3.5 h-3.5" :class="isRunning ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'" />
        <span>{{ isRunning ? t('stop_btn') : t('start_app_btn') }}</span>
      </button>

      <!-- Icon-only Buttons (28-32px neutral, colored icon on hover only) -->
      <!-- DB (conditional) -->
      <button
        v-if="hasDb"
        type="button"
        @click="emit('open-adminer')"
        class="w-7 h-7 rounded-md flex items-center justify-center text-slate-400 hover:text-amber-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('open_project_db_hint')"
        :aria-label="t('open_project_db_hint')"
      >
        <Database class="w-3.5 h-3.5" />
      </button>

      <!-- VS Code -->
      <button
        type="button"
        @click="emit('open-vscode')"
        class="w-7 h-7 rounded-md flex items-center justify-center text-slate-400 hover:text-blue-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('open_vscode')"
        :aria-label="t('open_vscode')"
      >
        <Code class="w-3.5 h-3.5" />
      </button>

      <!-- Open Folder (for web projects) -->
      <button
        v-if="site.is_web"
        type="button"
        @click="emit('open-folder')"
        class="w-7 h-7 rounded-md flex items-center justify-center text-slate-400 hover:text-amber-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('open_folder')"
        :aria-label="t('open_folder')"
      >
        <Folder class="w-3.5 h-3.5" />
      </button>

      <!-- Copy Path -->
      <button
        type="button"
        @click="emit('copy', site.path || site.root_path)"
        class="w-7 h-7 rounded-md flex items-center justify-center hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :class="copiedItem === (site.path || site.root_path) ? 'text-emerald-500' : 'text-slate-400 hover:text-slate-600 dark:hover:text-slate-200'"
        :title="t('copy_path')"
        :aria-label="t('copy_path')"
      >
        <Check v-if="copiedItem === (site.path || site.root_path)" class="w-3.5 h-3.5 text-emerald-500" />
        <Copy v-else class="w-3.5 h-3.5" />
      </button>

      <!-- Settings -->
      <button
        type="button"
        @click="emit('settings')"
        class="w-7 h-7 rounded-md flex items-center justify-center text-slate-400 hover:text-blue-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('project_settings_btn')"
        :aria-label="t('project_settings_btn')"
      >
        <Settings class="w-3.5 h-3.5" />
      </button>

      <!-- Diagnostics -->
      <button
        type="button"
        @click="emit('diagnostics')"
        class="w-7 h-7 rounded-md flex items-center justify-center text-slate-400 hover:text-amber-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('diagnose_project_btn')"
        :aria-label="t('diagnose_project_btn')"
      >
        <Wrench class="w-3.5 h-3.5" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Code2,
  Globe,
  Server,
  Hexagon,
  FileText,
  Folder,
  ExternalLink,
  FolderOpen,
  Power,
  Database,
  Code,
  Copy,
  Check,
  Settings,
  Wrench
} from 'lucide-vue-next'

export interface SiteItem {
  name?: string
  domain: string
  path?: string
  root_path: string
  php_version?: string
  php_port?: number
  proxy_port?: number
  project_type?: string
  type_label?: string
  is_web?: boolean
}

interface Props {
  site: SiteItem
  t: (key: string, params?: Record<string, string | number>) => string
  copiedItem?: string
  isRunning?: boolean
  docRootRelative?: string
  isPhp?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  copiedItem: '',
  isRunning: false,
  docRootRelative: '',
  isPhp: false
})

const emit = defineEmits<{
  (e: 'open-browser'): void
  (e: 'open-folder'): void
  (e: 'open-vscode'): void
  (e: 'open-adminer'): void
  (e: 'copy', text: string): void
  (e: 'settings'): void
  (e: 'diagnostics'): void
  (e: 'toggle-app'): void
  (e: 'change-php', event: Event): void
}>()

const projectIcon = computed(() => {
  if (!props.site.is_web) return Folder
  const type = props.site.project_type
  if (type === 'laravel' || type === 'php') return Code2
  if (type === 'wordpress') return Globe
  if (type === 'node') return Hexagon
  if (type === 'static') return FileText
  if (['go', 'rust', 'dotnet', 'python', 'ruby', 'java', 'proxy'].includes(type || '')) return Server
  return Globe
})

function extractPhpVersionNumber(ver?: string): string {
  if (!ver) return '8.3'
  const clean = ver.replace(/PHP\s*/i, '').trim()
  if (clean.startsWith('8.5')) return '8.5'
  if (clean.startsWith('8.4')) return '8.4'
  if (clean.startsWith('8.3')) return '8.3'
  if (clean.startsWith('8.2')) return '8.2'
  if (clean.startsWith('8.1')) return '8.1'
  return clean || '8.3'
}

const phpVersionNumber = computed(() => extractPhpVersionNumber(props.site.php_version))

const isProxyApp = computed(() => {
  return Boolean(
    props.site.is_web && (
      (props.site.proxy_port && props.site.proxy_port > 0) ||
      ['go', 'rust', 'dotnet', 'python', 'ruby', 'java', 'proxy'].includes(props.site.project_type || '')
    )
  )
})

const proxyRuntimeInfo = computed(() => {
  if (!props.site.is_web) return null
  const type = props.site.project_type
  let name = ''
  let defaultPort = 8080

  if (type === 'go') { name = 'Go'; defaultPort = 8080 }
  else if (type === 'rust') { name = 'Rust'; defaultPort = 8080 }
  else if (type === 'dotnet') { name = '.NET'; defaultPort = 5000 }
  else if (type === 'python') { name = 'Python'; defaultPort = 8000 }
  else if (type === 'ruby') { name = 'Ruby'; defaultPort = 3000 }
  else if (type === 'java') { name = 'Java'; defaultPort = 8080 }
  else if (type === 'proxy') { name = 'Proxy'; defaultPort = 8080 }
  else if (props.site.proxy_port && props.site.proxy_port > 0) {
    name = props.site.type_label || 'Proxy'
    defaultPort = props.site.proxy_port
  }

  if (!name) return null
  const port = props.site.proxy_port || defaultPort
  return { name, port }
})

const hasDb = computed(() => {
  return Boolean(
    props.site.is_web && (
      props.site.project_type === 'laravel' ||
      props.site.project_type === 'wordpress' ||
      ((props.site.php_port || 0) > 0)
    )
  )
})
</script>
