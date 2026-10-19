<template>
  <div :dir="currentLang === 'ar' ? 'rtl' : 'ltr'" :class="[isDark ? 'dark bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-800', 'min-h-screen transition-colors duration-200']">
    <!-- Header -->
    <header class="border-b border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-900/80 backdrop-blur sticky top-0 z-30">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
        <div class="flex items-center space-x-3 rtl:space-x-reverse">
          <div class="w-10 h-10 rounded-xl bg-blue-600 flex items-center justify-center text-white font-bold text-xl shadow-sm">
            LG
          </div>
          <div>
            <h1 class="text-xl font-bold tracking-tight text-slate-900 dark:text-white flex items-center gap-2">
              {{ t('app_name') }}
              <span class="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300 font-mono">v1.0</span>
            </h1>
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ t('app_subtitle') }}</p>
          </div>
        </div>

        <!-- Controls: Language, Theme & Actions -->
        <div class="flex items-center gap-3">
          <button
            @click="installSSL"
            class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-emerald-50 text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-100 transition-colors"
            :title="t('ssl_install_hint')"
          >
            <ShieldCheck class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
            <span>{{ t('install_ssl_btn') }}</span>
          </button>

          <button
            @click="toggleLang"
            class="px-2.5 py-1.5 rounded-lg text-xs font-semibold border border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            {{ currentLang === 'ar' ? 'English' : 'العربية' }}
          </button>

          <button
            @click="isDark = !isDark"
            class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            :title="isDark ? t('light_mode') : t('dark_mode')"
          >
            <Sun v-if="isDark" class="w-4 h-4 text-amber-400" />
            <Moon v-else class="w-4 h-4 text-slate-600" />
          </button>

          <button
            @click="showNewModal = true"
            class="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-semibold bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
          >
            <Plus class="w-4 h-4" />
            <span>{{ t('new_project_btn') }}</span>
          </button>
        </div>
      </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-8">
      <!-- Notification Banner -->
      <div v-if="notification" class="p-4 rounded-xl border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/40 text-blue-800 dark:text-blue-200 flex items-center justify-between text-sm animate-fade-in">
        <div class="flex items-center gap-2">
          <Info class="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0" />
          <span>{{ notification }}</span>
        </div>
        <button @click="notification = ''" class="text-blue-500 hover:text-blue-700 font-bold">&times;</button>
      </div>

      <!-- Services Section -->
      <section>
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-bold flex items-center gap-2 text-slate-900 dark:text-white">
            <Server class="w-5 h-5 text-blue-600" />
            {{ t('services_title') }}
          </h2>
          <div class="flex items-center gap-3">
            <button
              @click="showDownloadModal = true"
              class="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400 font-semibold flex items-center gap-1 hover:underline"
            >
              <Download class="w-3.5 h-3.5" />
              <span>{{ t('download_runtimes_btn') }}</span>
            </button>
            <span class="text-xs text-slate-500">{{ services.filter(s => s.running).length }} / {{ services.length }} {{ t('status_running') }}</span>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
          <div
            v-for="service in services"
            :key="service.name"
            class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 shadow-xs flex flex-col justify-between"
          >
            <div class="flex items-start justify-between">
              <div>
                <h3 class="font-bold text-sm text-slate-800 dark:text-slate-200">{{ service.label }}</h3>
                <p class="text-xs text-slate-400 font-mono mt-0.5">:{{ service.port }}</p>
              </div>
              <span
                :class="[
                  service.running
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-400 border-emerald-200 dark:border-emerald-800'
                    : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400 border-slate-200 dark:border-slate-700',
                  'text-[10px] px-2 py-0.5 rounded-full border font-medium flex items-center gap-1'
                ]"
              >
                <span class="w-1.5 h-1.5 rounded-full" :class="service.running ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"></span>
                {{ service.running ? t('status_running') : t('status_stopped') }}
              </span>
            </div>

            <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-between">
              <span class="text-[11px] text-slate-400">{{ service.role }}</span>
              <button
                @click="toggleService(service)"
                class="text-xs font-semibold px-2 py-1 rounded-md transition-colors"
                :class="service.running ? 'text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/30' : 'text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-950/30'"
              >
                {{ service.running ? 'إيقاف' : 'تشغيل' }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- Sites & Projects Section -->
      <section>
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-bold flex items-center gap-2 text-slate-900 dark:text-white">
            <Globe class="w-5 h-5 text-blue-600" />
            {{ t('sites_title') }}
          </h2>
          <span class="text-xs text-slate-500">{{ sites.length }} متاح</span>
        </div>

        <div v-if="sites.length === 0" class="p-12 text-center rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 bg-white dark:bg-slate-900/30">
          <Globe class="w-10 h-10 text-slate-400 mx-auto mb-3 opacity-60" />
          <p class="text-sm text-slate-500 max-w-sm mx-auto">{{ t('empty_sites') }}</p>
          <button
            @click="showNewModal = true"
            class="mt-4 inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-semibold bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
          >
            <Plus class="w-4 h-4" />
            {{ t('new_project_btn') }}
          </button>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="site in sites"
            :key="site.domain"
            class="p-5 rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs hover:border-blue-400 dark:hover:border-blue-600 transition-all flex flex-col justify-between"
          >
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-md bg-emerald-50 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 font-mono">
                  <Lock class="w-3 h-3 text-emerald-600" />
                  HTTPS
                </span>
                <span class="text-xs font-mono px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 font-medium">
                  PHP :{{ site.php_port }}
                </span>
              </div>

              <h3 class="text-base font-bold text-slate-900 dark:text-white truncate">
                {{ site.domain }}
              </h3>
              <p class="text-xs text-slate-400 truncate mt-1 font-mono" :title="site.root_path">
                {{ site.root_path }}
              </p>
            </div>

            <div class="mt-5 pt-3 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-between">
              <span class="text-[11px] text-emerald-600 dark:text-emerald-400 font-medium flex items-center gap-1">
                <CheckCircle2 class="w-3.5 h-3.5" />
                {{ t('ssl_trusted') }}
              </span>
              <a
                :href="'https://' + site.domain"
                target="_blank"
                class="inline-flex items-center gap-1 text-xs font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
              >
                <span>{{ t('open_browser') }}</span>
                <ExternalLink class="w-3.5 h-3.5" />
              </a>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- New Project Modal -->
    <div v-if="showNewModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-xs">
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-xl p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white">{{ t('modal_new_title') }}</h3>
          <button @click="showNewModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">&times;</button>
        </div>

        <form @submit.prevent="handleCreateProject" class="space-y-4">
          <div>
            <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('project_name') }}</label>
            <input
              v-model="newProject.name"
              type="text"
              required
              placeholder="my-laravel-app"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
            />
            <p class="text-[11px] text-slate-400 mt-1">الرابط التلقائي: https://{{ newProject.name || 'name' }}.test</p>
          </div>

          <div>
            <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('project_type') }}</label>
            <select
              v-model="newProject.type"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
            >
              <option value="laravel">{{ t('type_laravel') }}</option>
              <option value="wordpress">{{ t('type_wordpress') }}</option>
              <option value="generic">{{ t('type_generic') }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">{{ t('php_version') }}</label>
            <select
              v-model="newProject.php"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
            >
              <option value="8.4">PHP 8.4 (الأحدث)</option>
              <option value="8.3">PHP 8.3 (الموصى به لـ Laravel 11/12)</option>
              <option value="8.2">PHP 8.2</option>
              <option value="8.1">PHP 8.1</option>
            </select>
          </div>

          <div class="pt-4 flex items-center justify-end gap-2 border-t border-slate-100 dark:border-slate-800">
            <button
              type="button"
              @click="showNewModal = false"
              class="px-3 py-1.5 text-xs font-medium rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
            >
              {{ t('cancel_btn') }}
            </button>
            <button
              type="submit"
              class="px-4 py-2 text-xs font-bold rounded-lg bg-blue-600 text-white hover:bg-blue-700 shadow-sm"
            >
              {{ t('create_btn') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Download Runtimes Modal -->
    <div v-if="showDownloadModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-xs">
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-xl p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2">
            <Download class="w-4 h-4 text-blue-600" />
            {{ t('modal_download_title') }}
          </h3>
          <button @click="showDownloadModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">&times;</button>
        </div>

        <div class="space-y-3">
          <div
            v-for="rt in availableRuntimes"
            :key="rt.name"
            class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50 dark:bg-slate-800/40"
          >
            <div>
              <h4 class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ rt.label }}</h4>
              <p class="text-[11px] text-slate-400 font-mono">{{ rt.desc }}</p>
            </div>
            <button
              @click="triggerDownloadRuntime(rt.name)"
              :disabled="downloadingRuntime === rt.name"
              class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 transition-colors flex items-center gap-1.5"
            >
              <Download class="w-3.5 h-3.5" :class="downloadingRuntime === rt.name ? 'animate-bounce' : ''" />
              <span>{{ downloadingRuntime === rt.name ? t('downloading') : t('download_now') }}</span>
            </button>
          </div>
        </div>

        <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex justify-end">
          <button
            @click="showDownloadModal = false"
            class="px-3 py-1.5 text-xs font-medium rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {{ t('cancel_btn') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  Server, Globe, Plus, ShieldCheck, Sun, Moon,
  Lock, ExternalLink, CheckCircle2, Info, Download
} from 'lucide-vue-next'

import arLocale from './locales/ar.json'
import enLocale from './locales/en.json'

const isDark = ref(true)
const currentLang = ref<'ar' | 'en'>('ar')
const showNewModal = ref(false)
const showDownloadModal = ref(false)
const downloadingRuntime = ref('')
const notification = ref('')

const availableRuntimes = ref([
  { name: 'caddy', label: 'Caddy Server v2.8', desc: 'خادم الويب الرئيسي وإدارة شهادات SSL التلقائية' },
  { name: 'php-8.4', label: 'PHP 8.4 (NTS x64)', desc: 'أحدث إصدار رسمي لمحرك PHP' },
  { name: 'php-8.3', label: 'PHP 8.3 (NTS x64)', desc: 'إصدار مستقر موصى به لـ Laravel 11/12' },
  { name: 'mailpit', label: 'Mailpit Server v1.21', desc: 'صيد واختبار رسائل البريد الإلكتروني محلياً' },
])

const locales: Record<string, Record<string, string>> = {
  ar: arLocale,
  en: enLocale
}

function t(key: string): string {
  return locales[currentLang.value]?.[key] || key
}

function toggleLang() {
  currentLang.value = currentLang.value === 'ar' ? 'en' : 'ar'
}

interface ServiceItem {
  name: string
  label: string
  port: number
  running: boolean
  role: string
}

const services = ref<ServiceItem[]>([
  { name: 'caddy', label: 'Caddy Web Server', port: 443, running: true, role: 'Auto-TLS Reverse Proxy' },
  { name: 'php-8.4', label: 'PHP 8.4 FastCGI', port: 9084, running: true, role: 'FastCGI Daemon' },
  { name: 'php-8.3', label: 'PHP 8.3 FastCGI', port: 9083, running: true, role: 'FastCGI Daemon' },
  { name: 'mysql', label: 'MySQL Server', port: 3306, running: true, role: 'Portable Relational DB' },
  { name: 'redis', label: 'Redis Cache', port: 6379, running: false, role: 'In-Memory Cache' },
  { name: 'mailpit', label: 'Mailpit Mail Catcher', port: 1025, running: true, role: 'Local SMTP & Web UI (:8025)' },
])

interface SiteItem {
  domain: string
  root_path: string
  php_port: number
}

const sites = ref<SiteItem[]>([
  { domain: 'laravel-starter.test', root_path: 'C:/laragal/sites/laravel-starter/public', php_port: 9083 },
  { domain: 'blog.test', root_path: 'C:/laragal/sites/blog/public', php_port: 9084 }
])

const newProject = ref({
  name: '',
  type: 'laravel',
  php: '8.3'
})

function toggleService(service: ServiceItem) {
  service.running = !service.running
  notification.value = `تم ${service.running ? 'تشغيل' : 'إيقاف'} خدمة ${service.label} بنجاح.`
}

function installSSL() {
  notification.value = t('success_ssl')
}

function handleCreateProject() {
  if (!newProject.value.name) return

  const domain = `${newProject.value.name}.test`
  const port = newProject.value.php === '8.4' ? 9084 : (newProject.value.php === '8.3' ? 9083 : 9082)

  sites.value.push({
    domain,
    root_path: `C:/laragal/sites/${newProject.value.name}/public`,
    php_port: port
  })

  notification.value = `تم إنشاء المشروع بنجاح وضبط رابط https://${domain} الموثوق بـ SSL.`
  showNewModal.value = false
  newProject.value.name = ''
}

async function triggerDownloadRuntime(name: string) {
  downloadingRuntime.value = name
  notification.value = `جاري بدء تنزيل وتثبيت حزمة ${name}...`

  // Simulate or call Wails bridge
  setTimeout(() => {
    downloadingRuntime.value = ''
    notification.value = `تم تجهيز وتثبيت حزمة ${name} بنجاح في مجلد bin/${name}.`
  }, 2000)
}
</script>
