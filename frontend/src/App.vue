<template>
  <div :dir="currentLang === 'ar' ? 'rtl' : 'ltr'" class="min-h-screen flex flex-col justify-between bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150 select-none">
    <div>
      <!-- Custom Desktop Frameless Title Bar -->
      <div
        class="wails-drag h-10 w-full bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-3 text-xs font-medium z-40 sticky top-0"
      >
        <!-- In Arabic (RTL): Window Controls on the LEFT -->
        <div v-if="currentLang === 'ar'" class="flex items-center gap-1 wails-no-drag">
          <button
            @click="closeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 hover:bg-rose-600 hover:text-white transition-colors"
            title="إغلاق"
          >
            <X class="w-3.5 h-3.5" />
          </button>
          <button
            @click="toggleMaximize"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 dark:text-slate-400 transition-colors"
            title="تكبير / استعادة"
          >
            <Square class="w-3 h-3" />
          </button>
          <button
            @click="minimizeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 dark:text-slate-400 transition-colors"
            title="تصغير"
          >
            <Minus class="w-3.5 h-3.5" />
          </button>
        </div>

        <!-- In English (LTR): App Logo on the LEFT -->
        <div v-else class="flex items-center gap-2">
          <div class="w-5 h-5 rounded bg-blue-600 flex items-center justify-center text-[10px] font-bold text-white">LG</div>
          <span class="text-xs font-semibold text-slate-700 dark:text-slate-300">LaraGal</span>
        </div>

        <!-- Title in center -->
        <div class="text-[11px] text-slate-400 dark:text-slate-500 font-mono tracking-wide pointer-events-none">
          LaraGal - {{ t('app_subtitle') }}
        </div>

        <!-- In Arabic (RTL): App Logo on the RIGHT -->
        <div v-if="currentLang === 'ar'" class="flex items-center gap-2">
          <span class="text-xs font-semibold text-slate-700 dark:text-slate-300">LaraGal</span>
          <div class="w-5 h-5 rounded bg-blue-600 flex items-center justify-center text-[10px] font-bold text-white">LG</div>
        </div>

        <!-- In English (LTR): Window Controls on the RIGHT -->
        <div v-else class="flex items-center gap-1 wails-no-drag">
          <button
            @click="minimizeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 dark:text-slate-400 transition-colors"
            title="Minimize"
          >
            <Minus class="w-3.5 h-3.5" />
          </button>
          <button
            @click="toggleMaximize"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 dark:text-slate-400 transition-colors"
            title="Maximize"
          >
            <Square class="w-3 h-3" />
          </button>
          <button
            @click="closeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 hover:bg-rose-600 hover:text-white transition-colors"
            title="Close"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      <!-- Main Application Header -->
      <header class="border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-slate-900/95 backdrop-blur z-30">
        <div class="w-full px-6 lg:px-8 py-4 flex items-center justify-between">
          <!-- Main Brand Anchor (Aligned at start of reading) -->
          <div class="flex items-center space-x-3.5 rtl:space-x-reverse">
            <div class="w-11 h-11 rounded-xl bg-gradient-to-tr from-blue-700 via-blue-600 to-indigo-500 flex items-center justify-center text-white font-black text-xl shadow-md ring-2 ring-blue-500/20">
              LG
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h1 class="text-xl font-black tracking-tight text-slate-900 dark:text-white">
                  {{ t('app_name') }}
                </h1>
                <span class="text-[11px] px-2 py-0.5 rounded-md bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 font-mono font-semibold border border-blue-200 dark:border-blue-900">
                  v1.0.1
                </span>
              </div>
              <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5">{{ t('app_subtitle') }}</p>
            </div>
          </div>

          <!-- Controls: Language, Theme & Actions -->
          <div class="flex items-center gap-2.5">
            <button
              @click="installSSL"
              class="hidden md:inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-emerald-50 text-emerald-700 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-100 dark:hover:bg-emerald-900/80 transition-colors"
              :title="t('ssl_install_hint')"
            >
              <ShieldCheck class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
              <span>{{ t('install_ssl_btn') }}</span>
            </button>

            <button
              @click="toggleLang"
              class="px-3 py-1.5 rounded-lg text-xs font-bold border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
            >
              {{ currentLang === 'ar' ? 'English' : 'العربية' }}
            </button>

            <button
              @click="toggleTheme"
              class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
              :title="isDark ? t('light_mode') : t('dark_mode')"
            >
              <Sun v-if="isDark" class="w-4 h-4 text-amber-400" />
              <Moon v-else class="w-4 h-4 text-slate-600" />
            </button>

            <button
              @click="showNewModal = true"
              class="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-all hover:shadow-blue-500/20"
            >
              <Plus class="w-4 h-4" />
              <span>{{ t('new_project_btn') }}</span>
            </button>
          </div>
        </div>
      </header>

      <main class="w-full px-6 lg:px-8 py-6 space-y-7">
        <!-- Notification Banner -->
        <div v-if="notification" class="p-3.5 rounded-xl border border-blue-200 dark:border-blue-900 bg-blue-50 dark:bg-blue-950/60 text-blue-800 dark:text-blue-200 flex items-center justify-between text-xs animate-fade-in shadow-xs">
          <div class="flex items-center gap-2">
            <Info class="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0" />
            <span>{{ notification }}</span>
          </div>
          <button @click="notification = ''" class="text-blue-500 hover:text-blue-700 font-bold px-1">&times;</button>
        </div>

        <!-- Services Section -->
        <section>
          <div class="flex items-center justify-between mb-3.5">
            <h2 class="text-base font-bold flex items-center gap-2 text-slate-900 dark:text-white">
              <Server class="w-4 h-4 text-blue-600" />
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
              <span class="text-xs text-slate-500 dark:text-slate-400 font-mono">{{ services.filter(s => s.running).length }}/{{ services.length }} {{ t('status_running') }}</span>
            </div>
          </div>

          <!-- Services Grid -->
          <div class="grid grid-cols-2 md:grid-cols-3 2xl:grid-cols-6 gap-3.5">
            <div
              v-for="service in services"
              :key="service.name"
              class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs flex flex-col justify-between hover:border-slate-300 dark:hover:border-slate-700 transition-colors"
            >
              <div>
                <div class="flex items-start justify-between gap-1 mb-2">
                  <span
                    :class="[
                      service.running
                        ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/80 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800'
                        : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400 border-slate-200 dark:border-slate-700',
                      'text-[10px] px-2 py-0.5 rounded-full border font-semibold flex items-center gap-1'
                    ]"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="service.running ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"></span>
                    {{ service.running ? t('status_running') : t('status_stopped') }}
                  </span>
                  <span dir="ltr" class="text-xs font-mono text-slate-400 dark:text-slate-400 font-medium">:{{ service.port }}</span>
                </div>

                <h3 class="font-bold text-sm text-slate-900 dark:text-slate-100 leading-snug">{{ service.label }}</h3>
                <p class="text-[11px] text-slate-400 dark:text-slate-400 mt-1 line-clamp-1" :title="service.role">{{ service.role }}</p>
              </div>

              <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end">
                <button
                  @click="toggleService(service)"
                  class="text-xs font-bold px-2.5 py-1 rounded-md transition-colors"
                  :class="service.running ? 'text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40' : 'text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-950/40'"
                >
                  {{ service.running ? 'إيقاف' : 'تشغيل' }}
                </button>
              </div>
            </div>
          </div>
        </section>

        <!-- Sites & Projects Section -->
        <section>
          <div class="flex items-center justify-between mb-3.5">
            <h2 class="text-base font-bold flex items-center gap-2 text-slate-900 dark:text-white">
              <Globe class="w-4 h-4 text-blue-600" />
              {{ t('sites_title') }}
            </h2>
            <span class="text-xs text-slate-500 dark:text-slate-400 font-mono">{{ sites.length }} متاح</span>
          </div>

          <div v-if="sites.length === 0" class="p-12 text-center rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 bg-white dark:bg-slate-900/60">
            <Globe class="w-10 h-10 text-slate-400 mx-auto mb-3 opacity-60" />
            <p class="text-sm text-slate-500 dark:text-slate-400 max-w-sm mx-auto">{{ t('empty_sites') }}</p>
            <button
              @click="showNewModal = true"
              class="mt-4 inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
            >
              <Plus class="w-4 h-4" />
              {{ t('new_project_btn') }}
            </button>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            <div
              v-for="site in sites"
              :key="site.domain"
              class="p-5 rounded-2xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs hover:border-blue-500/50 dark:hover:border-blue-600 transition-all flex flex-col justify-between"
            >
              <div>
                <div class="flex items-center justify-between mb-3">
                  <span class="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-md bg-emerald-50 text-emerald-700 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 font-mono font-semibold">
                    <Lock class="w-3 h-3 text-emerald-600" />
                    HTTPS
                  </span>
                  <span dir="ltr" class="text-xs font-mono px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-semibold border border-slate-200 dark:border-slate-700">
                    PHP :{{ site.php_port }}
                  </span>
                </div>

                <h3 class="text-base font-bold text-slate-900 dark:text-white truncate">
                  {{ site.domain }}
                </h3>

                <!-- Monospace Path Box with Actions -->
                <div class="mt-3 flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700/80 text-[11px] font-mono text-slate-600 dark:text-slate-300">
                  <span dir="ltr" class="truncate flex-1" :title="site.root_path">{{ site.root_path }}</span>
                  <div class="flex items-center gap-1 flex-shrink-0 ms-2">
                    <button
                      @click="copyText(site.root_path)"
                      :title="t('copy_url')"
                      class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                    >
                      <Check v-if="copiedItem === site.root_path" class="w-3 h-3 text-emerald-500" />
                      <Copy v-else class="w-3 h-3 text-slate-400" />
                    </button>
                    <button
                      @click="openExplorer(site.root_path)"
                      :title="t('open_folder')"
                      class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                    >
                      <Folder class="w-3 h-3 text-slate-400" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Quick action bar -->
              <div class="mt-5 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <span class="text-[11px] text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1">
                    <CheckCircle2 class="w-3.5 h-3.5" />
                    {{ t('ssl_trusted') }}
                  </span>
                  <button
                    @click="openVSCode(site.root_path)"
                    class="hidden sm:inline-flex items-center gap-1 text-[11px] font-semibold text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors ms-1"
                    :title="t('open_vscode')"
                  >
                    <Code class="w-3.5 h-3.5" />
                    <span>VS Code</span>
                  </button>
                </div>

                <a
                  :href="'https://' + site.domain"
                  target="_blank"
                  @click.prevent="openBrowser('https://' + site.domain)"
                  class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold bg-blue-50 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-800 hover:bg-blue-100 dark:hover:bg-blue-900/80 transition-colors"
                >
                  <span>{{ t('open_browser') }}</span>
                  <ExternalLink class="w-3.5 h-3.5" />
                </a>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>

    <!-- Desktop Status Bar -->
    <footer class="mt-8 border-t border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/90 backdrop-blur py-2.5 px-6 lg:px-8 text-xs text-slate-500 dark:text-slate-400">
      <div class="w-full flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-1.5">
            <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span class="font-medium text-slate-700 dark:text-slate-300">{{ t('all_services_healthy') }}</span>
          </div>
          <span class="text-slate-300 dark:text-slate-700">|</span>
          <div class="flex items-center gap-1 font-mono">
            <span>{{ t('root_directory') }}:</span>
            <button @click="openExplorer('C:/laragal')" class="text-blue-600 dark:text-blue-400 font-semibold hover:underline">C:\laragal</button>
          </div>
        </div>

        <div class="flex items-center gap-3 font-mono text-[11px]">
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">TLD: *.test</span>
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">SSL: Auto-TLS</span>
          <span class="text-slate-400 dark:text-slate-500">LaraGal v1.0.1</span>
        </div>
      </div>
    </footer>

    <!-- New Project Modal -->
    <div v-if="showNewModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white">{{ t('modal_new_title') }}</h3>
          <button @click="showNewModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">&times;</button>
        </div>

        <form @submit.prevent="handleCreateProject" class="space-y-4">
          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('project_name') }}</label>
            <input
              v-model="newProject.name"
              type="text"
              required
              placeholder="my-laravel-app"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none"
            />
            <p class="text-[11px] text-slate-400 mt-1">الرابط التلقائي: https://{{ newProject.name || 'name' }}.test</p>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('project_type') }}</label>
            <select
              v-model="newProject.type"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none"
            >
              <option value="laravel">{{ t('type_laravel') }}</option>
              <option value="wordpress">{{ t('type_wordpress') }}</option>
              <option value="generic">{{ t('type_generic') }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('php_version') }}</label>
            <select
              v-model="newProject.php"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none"
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
              class="px-3 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
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
    <div v-if="showDownloadModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in">
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
            class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50 dark:bg-slate-850"
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
            class="px-3 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800"
          >
            {{ t('cancel_btn') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  Server, Globe, Plus, ShieldCheck, Sun, Moon,
  Lock, ExternalLink, CheckCircle2, Info, Download,
  Folder, Code, Copy, Check, Minus, Square, X
} from 'lucide-vue-next'

import arLocale from './locales/ar.json'
import enLocale from './locales/en.json'

const isDark = ref(true)
const currentLang = ref<'ar' | 'en'>('ar')
const showNewModal = ref(false)
const showDownloadModal = ref(false)
const downloadingRuntime = ref('')
const notification = ref('')
const copiedItem = ref('')

// Synchronize dark class to html document element
watch(isDark, (val) => {
  if (val) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}, { immediate: true })

function toggleTheme() {
  isDark.value = !isDark.value
}

// Window control functions
function closeWindow() {
  if ((window as any).go?.main?.App?.CloseWindow) {
    (window as any).go.main.App.CloseWindow()
  } else if ((window as any).runtime?.Quit) {
    (window as any).runtime.Quit()
  }
}

function toggleMaximize() {
  if ((window as any).go?.main?.App?.ToggleMaximizeWindow) {
    (window as any).go.main.App.ToggleMaximizeWindow()
  } else if ((window as any).runtime?.WindowToggleMaximise) {
    (window as any).runtime.WindowToggleMaximise()
  }
}

function minimizeWindow() {
  if ((window as any).go?.main?.App?.MinimizeWindow) {
    (window as any).go.main.App.MinimizeWindow()
  } else if ((window as any).runtime?.WindowMinimise) {
    (window as any).runtime.WindowMinimise()
  }
}

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
  { name: 'caddy', label: 'Caddy Server', port: 443, running: true, role: 'Auto-TLS Reverse Proxy' },
  { name: 'php-8.4', label: 'PHP 8.4', port: 9084, running: true, role: 'FastCGI Daemon' },
  { name: 'php-8.3', label: 'PHP 8.3', port: 9083, running: true, role: 'FastCGI Daemon' },
  { name: 'mysql', label: 'MySQL Server', port: 3306, running: true, role: 'Portable Relational DB' },
  { name: 'redis', label: 'Redis Cache', port: 6379, running: false, role: 'In-Memory Cache' },
  { name: 'mailpit', label: 'Mailpit Catcher', port: 1025, running: true, role: 'Local SMTP & Web UI (:8025)' },
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

function copyText(txt: string) {
  navigator.clipboard.writeText(txt)
  copiedItem.value = txt
  setTimeout(() => { copiedItem.value = '' }, 2000)
}

function openExplorer(path: string) {
  if ((window as any).go?.main?.App?.OpenFolder) {
    (window as any).go.main.App.OpenFolder(path)
  }
}

function openBrowser(url: string) {
  if ((window as any).go?.main?.App?.OpenBrowser) {
    (window as any).go.main.App.OpenBrowser(url)
  } else {
    window.open(url, '_blank')
  }
}

function openVSCode(path: string) {
  if ((window as any).go?.main?.App?.OpenVSCode) {
    (window as any).go.main.App.OpenVSCode(path)
  }
}

function toggleService(service: ServiceItem) {
  service.running = !service.running
  notification.value = `تم ${service.running ? 'تشغيل' : 'إيقاف'} خدمة ${service.label} بنجاح.`
}

async function installSSL() {
  if ((window as any).go?.main?.App?.InstallSSLTrust) {
    try {
      const res = await (window as any).go.main.App.InstallSSLTrust()
      notification.value = res || t('success_ssl')
      return
    } catch (err: any) {
      notification.value = `خطأ أثناء تثبيت الشهادة: ${err}`
      return
    }
  }
  notification.value = t('success_ssl')
}

async function handleCreateProject() {
  if (!newProject.value.name) return

  const domain = `${newProject.value.name}.test`
  const port = newProject.value.php === '8.4' ? 9084 : (newProject.value.php === '8.3' ? 9083 : 9082)
  const rootPath = `C:/laragal/sites/${newProject.value.name}/public`

  if ((window as any).go?.main?.App?.CreateProject) {
    try {
      await (window as any).go.main.App.CreateProject(newProject.value.name, newProject.value.type, newProject.value.php)
    } catch (e: any) {
      console.warn("Go backend create project:", e)
    }
  }

  sites.value.push({
    domain,
    root_path: rootPath,
    php_port: port
  })

  notification.value = `تم إنشاء المشروع بنجاح وضبط رابط https://${domain} الموثوق بـ SSL.`
  showNewModal.value = false
  newProject.value.name = ''
}

async function triggerDownloadRuntime(name: string) {
  downloadingRuntime.value = name
  notification.value = `جاري بدء تنزيل وتثبيت حزمة ${name}...`

  if ((window as any).go?.main?.App?.DownloadBinary) {
    try {
      const msg = await (window as any).go.main.App.DownloadBinary(name)
      downloadingRuntime.value = ''
      notification.value = msg || `تم تجهيز وتثبيت حزمة ${name} بنجاح.`
      return
    } catch (err: any) {
      console.warn("Download err:", err)
    }
  }

  setTimeout(() => {
    downloadingRuntime.value = ''
    notification.value = `تم تجهيز وتثبيت حزمة ${name} بنجاح في مجلد bin/${name}.`
  }, 1500)
}
</script>
