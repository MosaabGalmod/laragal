<template>
  <div :dir="currentLang === 'ar' ? 'rtl' : 'ltr'" class="h-screen w-screen flex flex-col overflow-hidden bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150 select-none">
    <!-- ================= UNIFIED DESKTOP HEADER (Single Bar - Zero Duplication) ================= -->
    <header
      dir="ltr"
      @dblclick="toggleMaximize"
      class="wails-drag flex-shrink-0 h-14 border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-slate-900/95 backdrop-blur-md z-40 px-4 flex items-center justify-between select-none cursor-default"
    >
      <!-- PHYSICAL LEFT (Window Controls + Actions in Arabic; Brand in English) -->
      <div v-if="currentLang === 'ar'" class="flex items-center gap-2.5 wails-no-drag">
        <!-- Window Controls: Always on physical LEFT in Arabic -->
        <div class="flex items-center gap-1">
          <button
            @click.stop="closeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-rose-600 hover:text-white transition-colors"
            :title="t('close_to_background')"
          >
            <X class="w-3.5 h-3.5" />
          </button>
          <button
            @click.stop="toggleMaximize"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            :title="t('maximize_restore')"
          >
            <Square class="w-3 h-3" />
          </button>
          <button
            @click.stop="minimizeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            :title="t('minimize')"
          >
            <Minus class="w-3.5 h-3.5" />
          </button>
        </div>

        <div class="h-5 w-px bg-slate-200 dark:bg-slate-800 mx-1"></div>

        <!-- Quick Actions Toolbar -->
        <button
          @click="openNewModal"
          class="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-bold bg-blue-600 hover:bg-blue-700 text-white shadow-xs transition-colors"
        >
          <Plus class="w-3.5 h-3.5" />
          <span>{{ t('new_project_btn') }}</span>
        </button>

        <button
          @click="syncHostsNow"
          :disabled="isSyncingHosts"
          class="hidden xl:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-blue-50 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-800 hover:bg-blue-100 dark:hover:bg-blue-900/80 transition-colors"
          :title="t('sync_hosts_hint')"
        >
          <RefreshCw class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" :class="{ 'animate-spin': isSyncingHosts }" />
          <span>{{ t('sync_hosts_btn') }}</span>
        </button>

        <button
          @click="installSSL"
          class="hidden xl:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-emerald-50 text-emerald-700 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-100 dark:hover:bg-emerald-900/80 transition-colors"
          :title="t('ssl_install_hint')"
        >
          <ShieldCheck class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
          <span>{{ t('install_ssl_btn') }}</span>
        </button>

        <button
          @click="toggleLang"
          class="px-2.5 py-1.5 rounded-lg text-xs font-bold border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
        >
          English
        </button>

        <button
          @click="toggleTheme"
          class="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
          :title="isDark ? t('light_mode') : t('dark_mode')"
        >
          <Sun v-if="isDark" class="w-3.5 h-3.5 text-amber-400" />
          <Moon v-else class="w-3.5 h-3.5 text-slate-600" />
        </button>
      </div>

      <!-- In English: Single Brand on Physical Left -->
      <div v-else class="flex items-center gap-3 wails-no-drag">
        <img src="/logo.png" alt="LaraGal" class="w-8 h-8 rounded-lg shadow-xs ring-1 ring-blue-500/20 object-contain flex-shrink-0" />
        <div class="flex items-center gap-2">
          <span class="text-lg font-black tracking-tight text-slate-900 dark:text-white">LaraGal</span>
          <span class="text-[10px] px-1.5 py-0.5 rounded-md bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 font-mono font-bold border border-blue-200 dark:border-blue-900">
            v1.0.0
          </span>
        </div>
        <div class="hidden md:block h-5 w-px bg-slate-200 dark:bg-slate-800"></div>
        <p class="hidden md:block text-xs text-slate-500 dark:text-slate-400 font-medium">
          Multi-Runtime Local Web Development &amp; Hosting Suite (PHP, Go, Rust, .NET, Python, Node.js)
        </p>
      </div>

      <!-- CENTER: Draggable Window Space -->
      <div class="flex-1 h-full flex items-center justify-center cursor-default"></div>

      <!-- PHYSICAL RIGHT (Single Brand in Arabic; Actions + Controls in English) -->
      <div v-if="currentLang === 'ar'" class="flex items-center gap-3 wails-no-drag" dir="rtl">
        <img src="/logo.png" alt="LaraGal" class="w-8 h-8 rounded-lg shadow-xs ring-1 ring-blue-500/20 object-contain flex-shrink-0" />
        <div class="flex items-center gap-2">
          <span class="text-lg font-black tracking-tight text-slate-900 dark:text-white">LaraGal</span>
          <span class="text-[10px] px-1.5 py-0.5 rounded-md bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 font-mono font-bold border border-blue-200 dark:border-blue-900">
            v1.0.0
          </span>
        </div>
        <div class="hidden md:block h-5 w-px bg-slate-200 dark:bg-slate-800"></div>
        <p class="hidden md:block text-xs text-slate-500 dark:text-slate-400 font-medium" dir="rtl">
          منصة الاستضافة والتطوير الشاملة لتطبيقات الويب (<bdi>PHP</bdi>، <bdi>Go</bdi>، <bdi>Rust</bdi>، <bdi>.NET</bdi>، <bdi>Python</bdi>، <bdi>Node.js</bdi>)
        </p>
      </div>

      <!-- In English: Actions + Window Controls on Physical Right -->
      <div v-else class="flex items-center gap-2.5 wails-no-drag">
        <button
          @click="openNewModal"
          class="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-bold bg-blue-600 hover:bg-blue-700 text-white shadow-xs transition-colors"
        >
          <Plus class="w-3.5 h-3.5" />
          <span>{{ t('new_project_btn') }}</span>
        </button>

        <button
          @click="syncHostsNow"
          :disabled="isSyncingHosts"
          class="hidden xl:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-blue-50 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-800 hover:bg-blue-100 dark:hover:bg-blue-900/80 transition-colors"
          :title="t('sync_hosts_hint')"
        >
          <RefreshCw class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" :class="{ 'animate-spin': isSyncingHosts }" />
          <span>{{ t('sync_hosts_btn') }}</span>
        </button>

        <button
          @click="installSSL"
          class="hidden xl:inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-emerald-50 text-emerald-700 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-100 dark:hover:bg-emerald-900/80 transition-colors"
          :title="t('ssl_install_hint')"
        >
          <ShieldCheck class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
          <span>{{ t('install_ssl_btn') }}</span>
        </button>

        <button
          @click="toggleLang"
          class="px-2.5 py-1.5 rounded-lg text-xs font-bold border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
        >
          العربية
        </button>

        <button
          @click="toggleTheme"
          class="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
          :title="isDark ? t('light_mode') : t('dark_mode')"
        >
          <Sun v-if="isDark" class="w-3.5 h-3.5 text-amber-400" />
          <Moon v-else class="w-3.5 h-3.5 text-slate-600" />
        </button>

        <div class="h-5 w-px bg-slate-200 dark:bg-slate-800 mx-1"></div>

        <!-- Window Controls: Far Right in English -->
        <div class="flex items-center gap-1">
          <button
            @click.stop="minimizeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            :title="t('minimize')"
          >
            <Minus class="w-3.5 h-3.5" />
          </button>
          <button
            @click.stop="toggleMaximize"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            :title="t('maximize_restore')"
          >
            <Square class="w-3 h-3" />
          </button>
          <button
            @click.stop="closeWindow"
            class="w-7 h-7 rounded-md flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-rose-600 hover:text-white transition-colors"
            :title="t('close_to_background')"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </header>

    <!-- ================= FLOATING TOAST NOTIFICATION ================= -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div
        v-if="notification"
        class="fixed top-16 inset-x-0 mx-auto w-fit max-w-lg z-50 flex items-center gap-2.5 px-4 py-2.5 rounded-xl bg-slate-900/95 dark:bg-slate-100/95 text-white dark:text-slate-900 shadow-xl border border-slate-700/80 dark:border-slate-300 text-xs font-medium backdrop-blur-md"
      >
        <Info class="w-4 h-4 text-blue-400 dark:text-blue-600 flex-shrink-0" />
        <span class="truncate"><bdi>{{ notification }}</bdi></span>
        <button @click="notification = ''" class="ms-2 opacity-70 hover:opacity-100 text-sm font-bold leading-none">&times;</button>
      </div>
    </transition>

    <!-- ================= MAIN SCROLLABLE CONTENT ================= -->
    <main class="flex-1 overflow-y-auto px-6 lg:px-8 py-5 space-y-6">
      <!-- Projects Directory Strip -->
      <div class="flex flex-wrap items-center justify-between gap-3 p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-white/70 dark:bg-slate-900/70 backdrop-blur-xs text-xs">
        <div class="flex items-center gap-2 text-slate-600 dark:text-slate-400">
          <Folder class="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0" />
          <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('root_directory') }}:</span>
          <bdi dir="ltr" class="font-mono text-[11px] font-bold px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-200 select-all">{{ projectsDirectory }}</bdi>
          <button
            @click="changeProjectsDirectory"
            class="text-[11px] font-bold text-blue-600 hover:text-blue-700 dark:text-blue-400 hover:underline ms-1"
          >
            {{ t('change_btn') }}
          </button>
        </div>
        <div class="flex items-center gap-2 text-[11px] text-slate-500 dark:text-slate-400 font-mono">
          <span class="px-2 py-0.5 rounded-md bg-emerald-50 text-emerald-700 dark:bg-emerald-950/70 dark:text-emerald-300 font-semibold border border-emerald-200 dark:border-emerald-800">
            <bdi>{{ webSitesCount }}</bdi> {{ t('web_count') }}
          </span>
          <span v-if="nonWebCount > 0" class="px-2 py-0.5 rounded-md bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400 font-semibold border border-slate-200 dark:border-slate-700">
            <bdi>{{ nonWebCount }}</bdi> {{ t('non_web_count') }}
          </span>
        </div>
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
            :key="service.id"
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
                <span dir="ltr" class="text-[11px] font-mono font-bold px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">:{{ service.port }}</span>
              </div>

              <h3 class="font-bold text-sm text-slate-900 dark:text-slate-100 leading-snug">{{ t(service.name_key) }}</h3>
              <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-1 line-clamp-1" :title="t(service.desc_key)">
                {{ t(service.desc_key) }}
              </p>

              <!-- Credentials & Connection Settings Box -->
              <div v-if="service.credentials" class="mt-2.5 p-2 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-100 dark:border-slate-800/80 text-[11px] space-y-1 select-text">
                <!-- MySQL Credentials -->
                <template v-if="service.id === 'mysql'">
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('user_label') }}:</span>
                    <div class="flex items-center gap-1 font-mono font-semibold">
                      <span>root</span>
                      <button @click.stop="copyText('root')" class="p-0.5 hover:text-blue-600 transition-colors" :title="t('copy_url')">
                        <Check v-if="copiedItem === 'root'" class="w-3 h-3 text-emerald-500" />
                        <Copy v-else class="w-3 h-3 text-slate-400 hover:text-slate-600" />
                      </button>
                    </div>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('password_label') }}:</span>
                    <span class="text-slate-400 italic font-mono text-[10px]">{{ t('empty_password') }}</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('host_label') }}:</span>
                    <span dir="ltr" class="font-mono font-semibold text-[10px]">127.0.0.1:3306</span>
                  </div>
                </template>

                <!-- Mailpit Settings -->
                <template v-else-if="service.id === 'mailpit'">
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('smtp_label') }}:</span>
                    <div class="flex items-center gap-1 font-mono font-semibold">
                      <span dir="ltr">127.0.0.1:1025</span>
                      <button @click.stop="copyText('127.0.0.1:1025')" class="p-0.5 hover:text-blue-600 transition-colors" :title="t('copy_url')">
                        <Check v-if="copiedItem === '127.0.0.1:1025'" class="w-3 h-3 text-emerald-500" />
                        <Copy v-else class="w-3 h-3 text-slate-400 hover:text-slate-600" />
                      </button>
                    </div>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('web_ui_label') }}:</span>
                    <span dir="ltr" class="font-mono font-semibold text-[10px]">127.0.0.1:8025</span>
                  </div>
                </template>

                <!-- Redis Settings -->
                <template v-else-if="service.id === 'redis'">
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('host_label') }}:</span>
                    <span dir="ltr" class="font-mono font-semibold text-[10px]">127.0.0.1:6379</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('password_label') }}:</span>
                    <span class="text-slate-400 italic text-[10px]">{{ t('no_auth') }}</span>
                  </div>
                </template>

                <!-- PHP FastCGI Settings -->
                <template v-else-if="service.id.startsWith('php-')">
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('workers_label') }}:</span>
                    <span class="font-mono font-semibold text-emerald-600 dark:text-emerald-400 text-[10px]">8 Workers</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">{{ t('opcache_enabled') }}:</span>
                    <span class="text-emerald-600 dark:text-emerald-400 font-semibold text-[10px]">✓</span>
                  </div>
                </template>

                <!-- Caddy Settings -->
                <template v-else-if="service.id === 'caddy'">
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">API:</span>
                    <span dir="ltr" class="font-mono font-semibold text-[10px]">localhost:2019</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400 dark:text-slate-500">SSL:</span>
                    <span class="font-semibold text-emerald-600 dark:text-emerald-400 text-[10px]">Auto *.test</span>
                  </div>
                </template>
              </div>
            </div>

            <div class="mt-3.5 pt-2.5 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between gap-2">
              <div>
                <button
                  v-if="service.running && service.url"
                  @click="openBrowser(service.url)"
                  class="inline-flex items-center gap-1 px-2 py-1 rounded text-[11px] font-bold text-blue-600 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 dark:bg-blue-950/60 dark:text-blue-300 dark:hover:bg-blue-900/60 transition-colors"
                  :title="service.url"
                >
                  <span>{{ t('open_dashboard') }}</span>
                  <ExternalLink class="w-3 h-3" />
                </button>
              </div>

              <button
                @click="promptToggleService(service)"
                class="text-xs font-bold px-3 py-1 rounded-md transition-colors border"
                :class="service.running ? 'text-rose-600 border-rose-200 hover:bg-rose-50 dark:border-rose-900/60 dark:hover:bg-rose-950/40' : 'text-blue-600 border-blue-200 hover:bg-blue-50 dark:border-blue-900/60 dark:hover:bg-blue-950/40'"
              >
                {{ service.running ? t('stop_btn') : t('start_btn') }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- Sites & Projects Section -->
      <section>
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 mb-4">
          <div class="flex items-center gap-3">
            <h2 class="text-base font-bold flex items-center gap-2 text-slate-900 dark:text-white">
              <Globe class="w-4 h-4 text-blue-600" />
              {{ t('sites_title') }}
            </h2>
            <div class="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400 font-mono">
              <span class="px-2 py-0.5 rounded-md bg-blue-50 dark:bg-blue-950/60 text-blue-700 dark:text-blue-300 font-bold border border-blue-200 dark:border-blue-900">
                <bdi>{{ webSitesCount }}</bdi> {{ t('web_count') }}
              </span>
              <span v-if="nonWebCount > 0" class="px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
                <bdi>{{ nonWebCount }}</bdi> {{ t('non_web_count') }}
              </span>
            </div>
          </div>

          <!-- Filter tabs -->
          <div class="flex items-center gap-1 p-1 rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700/80 text-xs font-semibold self-start md:self-auto overflow-x-auto">
            <button
              @click="activeFilter = 'web'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'web' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              {{ t('filter_web') }} (<bdi>{{ webSitesCount }}</bdi>)
            </button>
            <button
              v-if="laravelSitesCount > 0"
              @click="activeFilter = 'laravel'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'laravel' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              Laravel (<bdi>{{ laravelSitesCount }}</bdi>)
            </button>
            <button
              v-if="phpSitesCount > 0"
              @click="activeFilter = 'php'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'php' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              PHP (<bdi>{{ phpSitesCount }}</bdi>)
            </button>
            <button
              v-if="goSitesCount > 0"
              @click="activeFilter = 'go'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'go' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              Go (<bdi>{{ goSitesCount }}</bdi>)
            </button>
            <button
              v-if="rustSitesCount > 0"
              @click="activeFilter = 'rust'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'rust' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              Rust (<bdi>{{ rustSitesCount }}</bdi>)
            </button>
            <button
              v-if="dotnetSitesCount > 0"
              @click="activeFilter = 'dotnet'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'dotnet' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              .NET (<bdi>{{ dotnetSitesCount }}</bdi>)
            </button>
            <button
              v-if="pythonSitesCount > 0"
              @click="activeFilter = 'python'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'python' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              Python (<bdi>{{ pythonSitesCount }}</bdi>)
            </button>
            <button
              v-if="nodeSitesCount > 0"
              @click="activeFilter = 'node'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'node' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              Node.js (<bdi>{{ nodeSitesCount }}</bdi>)
            </button>
            <button
              v-if="flutterSitesCount > 0"
              @click="activeFilter = 'flutter'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'flutter' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              Flutter (<bdi>{{ flutterSitesCount }}</bdi>)
            </button>
            <button
              v-if="nonWebCount > 0"
              @click="activeFilter = 'non_web'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'non_web' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              {{ t('filter_non_web') }} (<bdi>{{ nonWebCount }}</bdi>)
            </button>
            <button
              @click="activeFilter = 'all'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'all' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              {{ t('filter_all') }} (<bdi>{{ sites.length }}</bdi>)
            </button>
          </div>
        </div>

        <div v-if="filteredSites.length === 0" class="p-12 text-center rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 bg-white dark:bg-slate-900/60">
          <Globe class="w-10 h-10 text-slate-400 mx-auto mb-3 opacity-60" />
          <p class="text-sm text-slate-500 dark:text-slate-400 max-w-sm mx-auto">
            <span v-if="activeFilter === 'web' && sites.length > 0">
              {{ t('no_web_sites_in_dir') }}
            </span>
            <span v-else>{{ t('empty_sites') }}</span>
          </p>
          <div class="mt-4 flex items-center justify-center gap-2">
            <button
              v-if="activeFilter === 'web' && sites.length > 0"
              @click="activeFilter = 'all'"
              class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors"
            >
              <Folder class="w-3.5 h-3.5" />
              <span>{{ t('show_all_folders') }} (<bdi>{{ sites.length }}</bdi>)</span>
            </button>
            <button
              @click="openNewModal"
              class="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
            >
              <Plus class="w-4 h-4" />
              {{ t('new_project_btn') }}
            </button>
          </div>
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          <div
            v-for="site in filteredSites"
            :key="site.domain || site.root_path"
            class="p-5 rounded-2xl border bg-white dark:bg-slate-900 shadow-xs transition-all flex flex-col justify-between"
            :class="site.is_web ? 'border-slate-200 dark:border-slate-800 hover:border-blue-500/50 dark:hover:border-blue-600' : 'border-dashed border-slate-300 dark:border-slate-800 opacity-90'"
          >
            <div>
              <div class="flex items-center justify-between mb-3">
                <span v-if="site.is_web" class="inline-flex items-center gap-1 text-[11px] px-2.5 py-1 rounded-md bg-emerald-50 text-emerald-700 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 font-mono font-bold">
                  <Lock class="w-3 h-3 text-emerald-600" />
                  HTTPS
                </span>
                <span v-else class="inline-flex items-center gap-1 text-[11px] px-2.5 py-1 rounded-md bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400 border border-slate-200 dark:border-slate-700 font-sans font-semibold">
                  <Folder class="w-3 h-3 text-slate-400" />
                  {{ t('non_web_badge') }}
                </span>

                <!-- Interactive PHP Version Selector for PHP Projects -->
                <div v-if="isPhpProject(site)" dir="ltr" class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md bg-blue-50/90 dark:bg-blue-950/70 border border-blue-200 dark:border-blue-900/80 text-xs font-mono">
                  <span v-if="site.project_type === 'laravel'" class="font-bold text-blue-700 dark:text-blue-300">Laravel •</span>
                  <span v-else-if="site.project_type === 'wordpress'" class="font-bold text-blue-700 dark:text-blue-300">WordPress •</span>
                  <select
                    :value="extractPhpVersionNumber(site.php_version)"
                    @change="updateSitePHPVersion(site, $event)"
                    class="bg-transparent text-blue-800 dark:text-blue-200 font-bold focus:outline-none cursor-pointer text-xs"
                    :title="t('change_php_hint')"
                  >
                    <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.4">PHP 8.4</option>
                    <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.3">PHP 8.3</option>
                    <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.2">PHP 8.2</option>
                    <option class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white" value="8.1">PHP 8.1</option>
                  </select>
                </div>

                <span v-else-if="site.is_web" dir="ltr" class="text-xs font-mono px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-bold border border-slate-200 dark:border-slate-700">
                  <span v-if="site.project_type === 'go'">Go • Proxy :{{ site.proxy_port || 8080 }}</span>
                  <span v-else-if="site.project_type === 'rust'">Rust • Proxy :{{ site.proxy_port || 8080 }}</span>
                  <span v-else-if="site.project_type === 'dotnet'">.NET • Proxy :{{ site.proxy_port || 5000 }}</span>
                  <span v-else-if="site.project_type === 'python'">Python • Proxy :{{ site.proxy_port || 8000 }}</span>
                  <span v-else-if="site.project_type === 'ruby'">Ruby • Proxy :{{ site.proxy_port || 3000 }}</span>
                  <span v-else-if="site.project_type === 'java'">Java • Proxy :{{ site.proxy_port || 8080 }}</span>
                  <span v-else-if="site.project_type === 'flutter'">Flutter Web</span>
                  <span v-else-if="site.project_type === 'static'">HTML / ثابت</span>
                  <span v-else-if="site.project_type === 'node'">Node.js</span>
                  <span v-else>{{ site.type_label || site.php_version || 'Web' }}</span>
                </span>
                <span v-else class="text-xs font-sans px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
                  {{ t('local_folder_badge') }}
                </span>
              </div>

              <h3 class="text-base font-bold text-slate-900 dark:text-white truncate">
                {{ site.is_web ? site.domain : (site.name || site.domain) }}
              </h3>

              <!-- Monospace Main Project Path Box with Actions -->
              <div dir="ltr" class="mt-3 flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700/80 text-[11px] font-mono text-slate-600 dark:text-slate-300">
                <span class="truncate flex-1 text-left select-all" :title="site.path || site.root_path">{{ site.path || site.root_path }}</span>
                <div class="flex items-center gap-1 flex-shrink-0 ms-2">
                  <button
                    @click="openDiagnosticsModal(site)"
                    :title="t('diagnose_project_btn')"
                    class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-amber-500 transition-colors"
                  >
                    <Wrench class="w-3 h-3" />
                  </button>
                  <button
                    @click="openSettingsModal(site)"
                    :title="t('project_settings_btn')"
                    class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-blue-500 transition-colors"
                  >
                    <Settings class="w-3 h-3" />
                  </button>
                  <button
                    @click="copyText(site.path || site.root_path)"
                    :title="t('copy_url')"
                    class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                  >
                    <Check v-if="copiedItem === (site.path || site.root_path)" class="w-3 h-3 text-emerald-500" />
                    <Copy v-else class="w-3 h-3 text-slate-400" />
                  </button>
                  <button
                    @click="openExplorer(site.path || site.root_path)"
                    :title="t('open_folder')"
                    class="p-1 rounded hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                  >
                    <Folder class="w-3 h-3 text-slate-400" />
                  </button>
                </div>
              </div>

              <!-- DocRoot Subfolder indicator if distinct from main project directory -->
              <div v-if="getDocRootRelative(site)" class="mt-1.5 flex items-center justify-between text-[10px] text-slate-500 dark:text-slate-400 font-mono">
                <div class="flex items-center gap-1.5">
                  <span class="font-sans font-semibold text-slate-400">{{ t('doc_root') }}:</span>
                  <span class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-bold border border-slate-200 dark:border-slate-700">
                    <bdi>{{ getDocRootRelative(site) }}</bdi>
                  </span>
                </div>
                <span v-if="site.project_type === 'go'" class="text-amber-600 dark:text-amber-400 font-sans font-semibold">
                  <bdi>Reverse Proxy &rarr; :{{ site.proxy_port || 8080 }}</bdi>
                </span>
              </div>
            </div>

            <!-- Quick action bar -->
            <div class="mt-5 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span v-if="site.is_web" class="text-[11px] text-emerald-600 dark:text-emerald-400 font-bold flex items-center gap-1">
                  <CheckCircle2 class="w-3.5 h-3.5" />
                  {{ t('ssl_trusted') }}
                </span>
                <span v-else class="text-[11px] text-slate-500 dark:text-slate-400 font-semibold flex items-center gap-1">
                  <Folder class="w-3.5 h-3.5" />
                  {{ t('local_folder_badge') }}
                </span>
                <button
                  @click="openVSCode(site.path || site.root_path)"
                  class="hidden sm:inline-flex items-center gap-1 text-[11px] font-bold text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors ms-1"
                  :title="t('open_vscode')"
                >
                  <Code class="w-3.5 h-3.5" />
                  <span>VS Code</span>
                </button>
              </div>

              <div class="flex items-center gap-1.5">
                <button
                  v-if="site.is_web && ((site.proxy_port && site.proxy_port > 0) || site.project_type === 'go' || site.project_type === 'rust' || site.project_type === 'dotnet' || site.project_type === 'python' || site.project_type === 'ruby' || site.project_type === 'java' || site.project_type === 'proxy')"
                  @click="toggleProjectApp(site)"
                  :title="runningProjects[site.domain] ? t('stop_app_btn') : t('start_app_btn')"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-semibold transition-colors"
                  :class="runningProjects[site.domain] ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-200 dark:border-amber-800 hover:bg-amber-100' : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/80 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800 hover:bg-emerald-100'"
                >
                  <Power class="w-3.5 h-3.5" :class="runningProjects[site.domain] ? 'text-amber-600' : 'text-emerald-600'" />
                  <span>{{ runningProjects[site.domain] ? t('stop_btn') : t('start_app_btn') }}</span>
                </button>

                <a
                  v-if="site.is_web"
                  :href="'https://' + site.domain"
                  target="_blank"
                  @click.prevent="openBrowser('https://' + site.domain)"
                  class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold bg-blue-50 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-800 hover:bg-blue-100 dark:hover:bg-blue-900/80 transition-colors"
                >
                  <span>{{ t('open_browser') }}</span>
                  <ExternalLink class="w-3.5 h-3.5" />
                </a>
                <button
                  v-else
                  @click="openExplorer(site.path || site.root_path)"
                  class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
                >
                  <span>{{ t('open_folder_btn') }}</span>
                  <FolderOpen class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- ================= 4. FIXED STATUS BAR ================= -->
    <footer class="flex-shrink-0 border-t border-slate-200 dark:border-slate-800 bg-white/90 dark:bg-slate-900/90 backdrop-blur py-2.5 px-6 lg:px-8 text-xs text-slate-500 dark:text-slate-400">
      <div class="w-full flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-1.5">
            <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span class="font-bold text-slate-700 dark:text-slate-300">{{ t('all_services_healthy') }}</span>
          </div>
          <span class="text-slate-300 dark:text-slate-700">|</span>
          <div class="flex items-center gap-2 font-mono">
            <span>{{ t('root_directory') }}:</span>
            <button @click="openExplorer(normalizedProjectsDir)" class="text-blue-600 dark:text-blue-400 font-bold hover:underline truncate max-w-[220px]" :title="normalizedProjectsDir">
              {{ normalizedProjectsDir }}
            </button>
            <button
              @click="changeProjectsDirectory"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-sans font-bold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors"
              :title="t('change_dir_hint')"
            >
              <FolderOpen class="w-3 h-3 text-amber-500" />
              <span>{{ t('change_btn') }}</span>
            </button>
            <select
              v-if="normalizedRecentDirs.length > 1"
              @change="switchRecentDirectory($event)"
              :value="normalizedProjectsDir"
              dir="ltr"
              class="text-[11px] font-sans font-semibold px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 focus:outline-none cursor-pointer max-w-[160px] truncate"
              :title="t('recent_directories')"
            >
              <option v-for="dir in normalizedRecentDirs" :key="dir" :value="dir">{{ dir }}</option>
            </select>
          </div>
        </div>

        <div class="flex items-center gap-3 font-mono text-[11px]">
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">TLD: *.test</span>
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">SSL: Auto-TLS</span>
          <button
            @click="promptQuitApp"
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded font-sans font-bold text-slate-500 hover:text-rose-600 dark:text-slate-400 dark:hover:text-rose-400 transition-colors"
            :title="t('quit_app_hint')"
          >
            <Power class="w-3 h-3 text-rose-500" />
            <span>{{ t('quit_app_btn') }}</span>
          </button>
          <span class="text-slate-400 dark:text-slate-500 font-bold">LaraGal v1.0.0</span>
        </div>
      </div>
    </footer>

    <!-- ================= 5. NEW PROJECT MODAL (Strict novalidate & Custom Validation) ================= -->
    <div v-if="showNewModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white">{{ t('modal_new_title') }}</h3>
          <button @click="showNewModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 font-bold">&times;</button>
        </div>

        <!-- Custom validated form with novalidate (Zero native browser tooltips) -->
        <form novalidate @submit.prevent="handleCreateProject" class="space-y-4">
          <div>
            <label for="projectNameInput" class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
              <span>{{ t('project_name') }}</span>
              <span class="text-rose-500 ms-1">*</span>
            </label>
            <input
              id="projectNameInput"
              v-model="newProject.name"
              type="text"
              autocomplete="off"
              @input="validateProjectName"
              @blur="validateProjectName"
              :aria-invalid="!!projectError"
              aria-describedby="projectNameError"
              placeholder="my-laravel-app"
              class="w-full px-3.5 py-2.5 rounded-lg border text-sm text-slate-900 dark:text-white transition-colors focus:outline-none"
              :class="projectError ? 'border-rose-500 bg-rose-50/40 dark:bg-rose-950/20 focus:ring-2 focus:ring-rose-500' : 'border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-blue-500'"
            />
            <!-- Custom accessible validation message -->
            <p v-if="projectError" id="projectNameError" role="alert" aria-live="polite" class="text-xs text-rose-500 font-semibold mt-1.5 flex items-center gap-1">
              <AlertCircle class="w-3.5 h-3.5 flex-shrink-0" />
              <span>{{ projectError }}</span>
            </p>
            <p v-else class="text-[11px] text-slate-400 mt-1 font-mono">
              الرابط التلقائي: https://{{ newProject.name ? newProject.name.toLowerCase() : 'name' }}.test
            </p>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">{{ t('target_directory') }}</label>
            <div class="flex items-center gap-2">
              <input
                type="text"
                readonly
                :value="projectsDirectory + (newProject.name ? '/' + newProject.name.trim().toLowerCase() : '')"
                dir="ltr"
                class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-xs font-mono text-slate-600 dark:text-slate-300 focus:outline-none"
              />
              <button
                type="button"
                @click="changeProjectsDirectory"
                class="flex-shrink-0 px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-xs font-bold text-slate-700 dark:text-slate-200 flex items-center gap-1 transition-colors"
                :title="t('change_dir_hint')"
              >
                <FolderOpen class="w-3.5 h-3.5 text-amber-500" />
                <span>{{ t('change_btn') }}</span>
              </button>
            </div>
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
              class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors"
            >
              {{ t('cancel_btn') }}
            </button>
            <button
              type="submit"
              class="px-4 py-2 text-xs font-bold rounded-lg bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors"
            >
              {{ t('create_btn') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- ================= 6. CUSTOM CONFIRMATION MODAL (Zero Native window.confirm) ================= -->
    <div v-if="confirmService" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-sm rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-4 animate-scale-in">
        <div class="flex items-center gap-3 text-amber-600 dark:text-amber-400">
          <AlertTriangle class="w-6 h-6 flex-shrink-0" />
          <h3 class="font-bold text-base text-slate-900 dark:text-white">{{ t('confirm_stop_title') }}</h3>
        </div>
        <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
          {{ t('confirm_stop_msg') }} ({{ t(confirmService.name_key) }})
        </p>
        <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-2">
          <button
            @click="confirmService = null"
            class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors"
          >
            {{ t('cancel_btn') }}
          </button>
          <button
            @click="executeStopService"
            class="px-4 py-1.5 text-xs font-bold rounded-lg bg-rose-600 text-white hover:bg-rose-700 shadow-sm transition-colors"
          >
            {{ t('confirm_btn') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ================= 7. QUIT APP CONFIRMATION MODAL (Zero Native window.confirm) ================= -->
    <div v-if="showQuitConfirm" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-sm rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-4 animate-scale-in">
        <div class="flex items-center gap-3 text-rose-600 dark:text-rose-400">
          <Power class="w-6 h-6 flex-shrink-0" />
          <h3 class="font-bold text-base text-slate-900 dark:text-white">{{ t('confirm_quit_title') }}</h3>
        </div>
        <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
          {{ t('confirm_quit_msg') }}
        </p>
        <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-2">
          <button
            @click="showQuitConfirm = false"
            class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors"
          >
            {{ t('cancel_btn') }}
          </button>
          <button
            @click="executeQuitApp"
            class="px-4 py-1.5 text-xs font-bold rounded-lg bg-rose-600 text-white hover:bg-rose-700 shadow-sm transition-colors"
          >
            {{ t('confirm_btn') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ================= 8. DOWNLOAD RUNTIMES MODAL ================= -->
    <div v-if="showDownloadModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2">
            <Download class="w-4 h-4 text-blue-600" />
            {{ t('modal_download_title') }}
          </h3>
          <button @click="showDownloadModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 font-bold">&times;</button>
        </div>

        <div class="space-y-3 max-h-[60vh] overflow-y-auto px-0.5">
          <div
            v-for="rt in availableRuntimes"
            :key="rt.name"
            class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-850 space-y-2"
          >
            <div class="flex items-center justify-between">
              <div>
                <h4 class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ rt.label }}</h4>
                <p class="text-[11px] text-slate-400 font-mono">{{ rt.desc }}</p>
              </div>
              <button
                @click="triggerDownloadRuntime(rt.name)"
                :disabled="downloadingRuntime === rt.name"
                class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 transition-colors flex items-center gap-1.5 cursor-pointer flex-shrink-0"
              >
                <Download class="w-3.5 h-3.5" :class="downloadingRuntime === rt.name ? 'animate-bounce' : ''" />
                <span>
                  {{ downloadingRuntime === rt.name ? `${downloadProgress[rt.name] || 0}%` : t('download_now') }}
                </span>
              </button>
            </div>
            <div v-if="downloadingRuntime === rt.name" class="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 overflow-hidden">
              <div
                class="bg-blue-600 h-1.5 rounded-full transition-all duration-200"
                :style="{ width: `${downloadProgress[rt.name] || 0}%` }"
              ></div>
            </div>
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

    <!-- ================= 9. DIRECTORY SELECTION MODAL (100% Arabic & Bilingual) ================= -->
    <div v-if="showDirModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2">
            <FolderOpen class="w-4 h-4 text-amber-500" />
            {{ t('modal_dir_title') }}
          </h3>
          <button @click="showDirModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 font-bold">&times;</button>
        </div>

        <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
          {{ t('modal_dir_desc') }}
        </p>

        <form novalidate @submit.prevent="applyCustomDirectory" class="space-y-4">
          <div>
            <label for="customDirInput" class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
              {{ t('dir_path_label') }}
            </label>
            <div class="flex items-center gap-2">
              <input
                id="customDirInput"
                v-model="customDirInput"
                type="text"
                dir="ltr"
                class="flex-1 px-3.5 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="D:\projects"
              />
              <button
                type="button"
                @click="triggerWindowsPicker"
                class="px-3 py-2 rounded-xl text-xs font-semibold bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors flex items-center gap-1.5 flex-shrink-0 cursor-pointer"
                :title="t('browse_windows_btn')"
              >
                <FolderOpen class="w-3.5 h-3.5 text-amber-500" />
                <span class="hidden sm:inline">{{ t('browse_windows_btn') }}</span>
              </button>
            </div>
          </div>

          <div v-if="suggestedPaths.length > 0">
            <span class="block text-[11px] font-bold text-slate-400 mb-1.5">{{ t('quick_paths') }}:</span>
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="p in suggestedPaths"
                :key="p"
                type="button"
                dir="ltr"
                @click="customDirInput = p"
                class="px-2.5 py-1 rounded-lg text-[11px] font-mono border transition-all cursor-pointer"
                :class="customDirInput === p ? 'bg-blue-50 border-blue-300 text-blue-700 dark:bg-blue-950 dark:border-blue-700 dark:text-blue-300 font-bold' : 'bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-750'"
              >
                {{ p }}
              </button>
            </div>
          </div>

          <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-2">
            <button
              type="button"
              @click="showDirModal = false"
              class="px-4 py-2 text-xs font-semibold rounded-xl text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            >
              {{ t('cancel_btn') }}
            </button>
            <button
              type="submit"
              class="px-5 py-2 text-xs font-bold rounded-xl bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors cursor-pointer"
            >
              {{ t('save_dir_btn') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- ================= 10. PROJECT SETTINGS & OVERRIDES MODAL ================= -->
    <div v-if="showSettingsModal && selectedSiteForConfig" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2">
            <Settings class="w-4 h-4 text-blue-500" />
            {{ t('modal_settings_title') }}
          </h3>
          <button @click="showSettingsModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-lg leading-none cursor-pointer">&times;</button>
        </div>

        <div class="p-2.5 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700/60 space-y-1">
          <div class="flex items-center justify-between text-xs">
            <span class="font-bold text-slate-700 dark:text-slate-200">{{ selectedSiteForConfig.name || selectedSiteForConfig.domain }}</span>
            <span dir="ltr" class="font-mono text-[11px] text-blue-600 dark:text-blue-400">{{ selectedSiteForConfig.domain }}</span>
          </div>
          <p dir="ltr" class="text-[11px] font-mono text-slate-500 dark:text-slate-400 truncate">
            {{ selectedSiteForConfig.path || selectedSiteForConfig.root_path }}
          </p>
        </div>

        <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
          {{ t('modal_settings_desc') }}
        </p>

        <form novalidate @submit.prevent="applyProjectSettings" class="space-y-4">
          <!-- Project Type Selector -->
          <div>
            <label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
              {{ t('setting_project_type') }}
            </label>
            <select
              v-model="settingsForm.type"
              class="w-full px-3 py-2 rounded-xl text-xs border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">كشف تلقائي (Auto-Detect)</option>
              <option value="laravel">Laravel (PHP FastCGI)</option>
              <option value="wordpress">WordPress (PHP FastCGI)</option>
              <option value="php">PHP عادي / Custom</option>
              <option value="node">Node.js (Next, Nuxt, Vite, Express)</option>
              <option value="go">Go Application (Proxy)</option>
              <option value="rust">Rust Application (Proxy)</option>
              <option value="dotnet">.NET / C# Application (Proxy)</option>
              <option value="python">Python (FastAPI, Flask, Django)</option>
              <option value="flutter">Flutter Web</option>
              <option value="static">صفحات ثابتة (HTML / CSS / JS)</option>
              <option value="non_web">غير ويب / مجلد محلي فقط</option>
            </select>
          </div>

          <!-- Document Root Subfolder -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="text-xs font-bold text-slate-700 dark:text-slate-300">
                {{ t('setting_docroot') }}
              </label>
              <div class="flex items-center gap-1">
                <span class="text-[10px] text-slate-400">سريع:</span>
                <button
                  type="button"
                  @click="settingsForm.doc_root = 'public'"
                  class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 hover:bg-blue-100 dark:hover:bg-blue-900/40 text-slate-600 dark:text-slate-300 cursor-pointer"
                >
                  public
                </button>
                <button
                  type="button"
                  @click="settingsForm.doc_root = 'dist'"
                  class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 hover:bg-blue-100 dark:hover:bg-blue-900/40 text-slate-600 dark:text-slate-300 cursor-pointer"
                >
                  dist
                </button>
                <button
                  type="button"
                  @click="settingsForm.doc_root = ''"
                  class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 hover:bg-blue-100 dark:hover:bg-blue-900/40 text-slate-600 dark:text-slate-300 cursor-pointer"
                >
                  الجذر (.)
                </button>
              </div>
            </div>
            <input
              v-model="settingsForm.doc_root"
              type="text"
              dir="ltr"
              class="w-full px-3 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="مثال: public أو dist أو اتركه فارغاً"
            />
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
              placeholder="0 (افتراضي للتطبيقات الخلفية، مثل: 8080 أو 3000)"
            />
          </div>

          <!-- PHP Version Selector (if PHP type) -->
          <div v-if="settingsForm.type === 'laravel' || settingsForm.type === 'wordpress' || settingsForm.type === 'php' || (!settingsForm.type && isPhpProject(selectedSiteForConfig))">
            <label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
              إصدار PHP المخصص
            </label>
            <select
              v-model="settingsForm.php_version"
              dir="ltr"
              class="w-full px-3 py-2 rounded-xl text-xs border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">الافتراضي العام للبرنامج</option>
              <option value="PHP 8.4">PHP 8.4</option>
              <option value="PHP 8.3">PHP 8.3</option>
              <option value="PHP 8.2">PHP 8.2</option>
              <option value="PHP 8.1">PHP 8.1</option>
            </select>
          </div>

          <!-- Ignore toggle -->
          <div class="pt-2">
            <label class="inline-flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                v-model="settingsForm.is_ignored"
                class="w-4 h-4 rounded text-blue-600 focus:ring-blue-500"
              />
              <span class="text-xs text-slate-600 dark:text-slate-400">إخفاء هذا المجلد من قائمة المشاريع كلياً</span>
            </label>
          </div>

          <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-2">
            <button
              type="button"
              @click="showSettingsModal = false"
              class="px-4 py-2 text-xs font-semibold rounded-xl text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors cursor-pointer"
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

    <!-- ================= 11. PROJECT DIAGNOSTICS & TROUBLESHOOTING MODAL ================= -->
    <div v-if="showDiagModal && selectedSiteForDiag" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2">
            <Wrench class="w-4 h-4 text-amber-500" />
            {{ t('modal_diag_title') }}
          </h3>
          <button @click="showDiagModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-lg leading-none cursor-pointer">&times;</button>
        </div>

        <!-- Project Badge & Retest button -->
        <div class="flex items-center justify-between p-3 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700/60">
          <div>
            <div class="font-bold text-xs text-slate-900 dark:text-white flex items-center gap-2">
              <span>{{ selectedSiteForDiag.name || selectedSiteForDiag.domain }}</span>
              <span dir="ltr" class="text-[11px] font-mono text-blue-600 dark:text-blue-400">({{ selectedSiteForDiag.domain }})</span>
            </div>
            <p dir="ltr" class="text-[10px] font-mono text-slate-500 dark:text-slate-400 truncate max-w-xs mt-0.5">
              {{ selectedSiteForDiag.path || selectedSiteForDiag.root_path }}
            </p>
          </div>
          <button
            type="button"
            @click="runDiagnostics(selectedSiteForDiag)"
            :disabled="isDiagnosing"
            class="px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-50 flex items-center gap-1.5 cursor-pointer"
            :title="'إعادة الفحص'"
          >
            <RefreshCw class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" :class="{ 'animate-spin': isDiagnosing }" />
            <span>إعادة الفحص</span>
          </button>
        </div>

        <div v-if="isDiagnosing" class="py-12 text-center">
          <RefreshCw class="w-8 h-8 text-blue-600 animate-spin mx-auto mb-2" />
          <p class="text-xs text-slate-500">جاري فحص ملفات ومنافذ المشروع...</p>
        </div>

        <div v-else-if="diagResult" class="space-y-4">
          <!-- Overall Status Banner -->
          <div
            class="p-3.5 rounded-xl border flex items-center gap-3"
            :class="diagResult.status === 'healthy'
              ? 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-200 dark:border-emerald-800 text-emerald-800 dark:text-emerald-200'
              : (diagResult.status === 'warning'
                ? 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-200'
                : 'bg-rose-50 dark:bg-rose-950/40 border-rose-200 dark:border-rose-800 text-rose-800 dark:text-rose-200')"
          >
            <CheckCircle2 v-if="diagResult.status === 'healthy'" class="w-5 h-5 flex-shrink-0 text-emerald-600 dark:text-emerald-400" />
            <AlertTriangle v-else-if="diagResult.status === 'warning'" class="w-5 h-5 flex-shrink-0 text-amber-600 dark:text-amber-400" />
            <AlertCircle v-else class="w-5 h-5 flex-shrink-0 text-rose-600 dark:text-rose-400" />
            <div class="text-xs">
              <span class="font-bold block">
                {{ diagResult.status === 'healthy' ? t('diag_status_healthy') : (diagResult.status === 'warning' ? t('diag_status_warning') : t('diag_status_error')) }}
              </span>
              <span class="text-[11px] opacity-80">
                النوع: {{ diagResult.project_type }} • {{ diagResult.php_version || 'FastCGI/Proxy' }}
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
              <span dir="ltr" class="font-mono text-[11px]" :class="diagResult.has_index_file ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'">
                {{ diagResult.index_file || 'مفقود' }}
              </span>
            </div>

            <!-- 3. .env file check (PHP / Laravel) -->
            <div v-if="diagResult.project_type === 'laravel' || diagResult.project_type === 'php' || diagResult.project_type === 'wordpress'" class="p-3 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Check v-if="diagResult.has_env_file" class="w-4 h-4 text-emerald-500" />
                <AlertCircle v-else class="w-4 h-4 text-amber-500" />
                <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_env_check') }}</span>
              </div>
              <div class="flex items-center gap-2">
                <span v-if="diagResult.has_env_file" class="text-emerald-600 dark:text-emerald-400 font-bold">موجود</span>
                <button
                  v-else
                  type="button"
                  @click="triggerFixEnv"
                  :disabled="isFixingEnv"
                  class="px-2 py-1 rounded bg-amber-100 hover:bg-amber-200 text-amber-800 dark:bg-amber-950 dark:hover:bg-amber-900 dark:text-amber-200 font-bold text-[11px] transition-colors cursor-pointer"
                >
                  <span v-if="isFixingEnv">جاري الإنشاء...</span>
                  <span v-else>{{ t('fix_env_btn') }}</span>
                </button>
              </div>
            </div>

            <!-- 4. Dependencies check -->
            <div v-if="diagResult.project_type === 'laravel' || diagResult.project_type === 'php' || diagResult.project_type === 'node'" class="p-3 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Check v-if="diagResult.has_vendor_dir || diagResult.has_node_modules" class="w-4 h-4 text-emerald-500" />
                <AlertTriangle v-else class="w-4 h-4 text-amber-500" />
                <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_deps_check') }}</span>
              </div>
              <span class="font-mono text-[11px]" :class="(diagResult.has_vendor_dir || diagResult.has_node_modules) ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'">
                {{ (diagResult.has_vendor_dir ? 'vendor/ مثبت' : (diagResult.has_node_modules ? 'node_modules مثبت' : 'غير مثبتة')) }}
              </span>
            </div>

            <!-- 5. Port Connection check -->
            <div class="p-3 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Check v-if="diagResult.is_port_active" class="w-4 h-4 text-emerald-500" />
                <X v-else class="w-4 h-4 text-rose-500" />
                <span class="font-semibold text-slate-800 dark:text-slate-200">{{ t('diag_port_check') }}</span>
              </div>
              <span dir="ltr" class="font-mono text-[11px]" :class="diagResult.is_port_active ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">
                {{ diagResult.php_port > 0 ? `PHP FastCGI :${diagResult.php_port}` : `Proxy :${diagResult.proxy_port}` }}
                ({{ diagResult.is_port_active ? 'متصل ومتاح' : 'غير متصل' }})
              </span>
            </div>
          </div>

          <!-- Discovered Databases & Schemas -->
          <div v-if="discoveredDatabases && discoveredDatabases.length > 0" class="p-3.5 rounded-xl bg-indigo-50/70 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800/60 space-y-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Database class="w-4 h-4 text-indigo-600 dark:text-indigo-400 flex-shrink-0" />
                <span class="text-xs font-bold text-indigo-900 dark:text-indigo-200">{{ t('db_discovery_title') }}</span>
              </div>
              <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-indigo-100 dark:bg-indigo-900/60 text-indigo-700 dark:text-indigo-300">
                {{ discoveredDatabases.length }} مُكتشف
              </span>
            </div>

            <p class="text-xs text-indigo-700 dark:text-indigo-300 leading-relaxed">
              {{ t('db_discovery_prompt') }}
            </p>

            <div v-if="dbImportSuccess" class="p-2.5 rounded-lg bg-emerald-100 dark:bg-emerald-950/60 border border-emerald-300 dark:border-emerald-800 text-xs text-emerald-800 dark:text-emerald-200 flex items-center gap-2 font-medium">
              <CheckCircle2 class="w-4 h-4 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
              <span>{{ t('db_import_success') }}</span>
            </div>

            <div v-for="(dbItem, idx) in discoveredDatabases" :key="idx" class="p-3 rounded-xl bg-white dark:bg-slate-900 border border-indigo-100 dark:border-indigo-900/50 shadow-xs space-y-2">
              <div class="flex items-start justify-between gap-2">
                <div>
                  <div class="flex items-center gap-1.5">
                    <span class="font-mono text-xs font-bold text-slate-800 dark:text-slate-100">{{ dbItem.file_name || dbItem.target_db }}</span>
                    <span class="text-[10px] px-1.5 py-0.5 rounded font-mono uppercase" :class="dbItem.type === 'schema_file' ? 'bg-amber-100 dark:bg-amber-950 text-amber-700 dark:text-amber-300' : 'bg-blue-100 dark:bg-blue-950 text-blue-700 dark:text-blue-300'">
                      {{ dbItem.type === 'schema_file' ? 'مخطط SQL' : 'مجلد بيانات سابق' }}
                    </span>
                  </div>
                  <span class="text-[11px] text-slate-500 block mt-0.5">{{ dbItem.description }}</span>
                </div>
                <button
                  type="button"
                  @click="importDiscoveredDatabase(dbItem)"
                  :disabled="isImportingDB"
                  class="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs transition-colors shadow-xs flex items-center gap-1.5 flex-shrink-0 cursor-pointer disabled:opacity-50"
                >
                  <RefreshCw v-if="isImportingDB" class="w-3.5 h-3.5 animate-spin" />
                  <Database v-else class="w-3.5 h-3.5" />
                  <span>{{ isImportingDB ? 'جاري الاستيراد...' : t('db_import_btn') }}</span>
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
                    +{{ dbItem.found_tables.length - 6 }} أخرى
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Issues & Suggestions -->
          <div v-if="diagResult.issues && diagResult.issues.length > 0" class="p-3.5 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 space-y-1.5">
            <span class="text-xs font-bold text-rose-700 dark:text-rose-300 block">{{ t('issues_found_title') }}</span>
            <ul class="text-xs text-rose-600 dark:text-rose-400 space-y-1 list-disc list-inside">
              <li v-for="(issue, idx) in diagResult.issues" :key="idx">{{ issue }}</li>
            </ul>
          </div>

          <div v-if="diagResult.suggestions && diagResult.suggestions.length > 0" class="p-3.5 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/50 space-y-1.5">
            <span class="text-xs font-bold text-blue-700 dark:text-blue-300 block">{{ t('suggestions_title') }}</span>
            <ul class="text-xs text-blue-600 dark:text-blue-400 space-y-1 list-disc list-inside">
              <li v-for="(sug, idx) in diagResult.suggestions" :key="idx">{{ sug }}</li>
            </ul>
          </div>
        </div>

        <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
          <button
            type="button"
            @click="openSettingsModal(selectedSiteForDiag)"
            class="px-3.5 py-2 text-xs font-semibold rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors flex items-center gap-1.5 cursor-pointer"
          >
            <Settings class="w-3.5 h-3.5" />
            <span>تعديل إعدادات المشروع</span>
          </button>
          <button
            type="button"
            @click="showDiagModal = false"
            class="px-5 py-2 text-xs font-bold rounded-xl bg-blue-600 text-white hover:bg-blue-700 transition-colors cursor-pointer"
          >
            {{ t('close_btn') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import {
  Server, Globe, Plus, ShieldCheck, Sun, Moon,
  Lock, ExternalLink, CheckCircle2, Info, Download,
  Folder, Code, Copy, Check, Minus, Square, X,
  AlertCircle, AlertTriangle, FolderOpen, Power, RefreshCw,
  Settings, Wrench, Database
} from 'lucide-vue-next'

import arLocale from './locales/ar.json'
import enLocale from './locales/en.json'

// Dark theme preference persistence
const isDark = ref(localStorage.getItem('laragal_theme') !== 'light')
const currentLang = ref<'ar' | 'en'>('ar')
const showNewModal = ref(false)
const showDownloadModal = ref(false)
const downloadingRuntime = ref('')
const notification = ref('')
const copiedItem = ref('')
const projectError = ref('')
const confirmService = ref<ServiceItem | null>(null)
const showQuitConfirm = ref(false)
const projectsDirectory = ref('C:/laragal/sites')
const recentDirectories = ref<string[]>([])
const runningProjects = ref<Record<string, boolean>>({})
const showDirModal = ref(false)
const customDirInput = ref('')

const showSettingsModal = ref(false)
const selectedSiteForConfig = ref<SiteItem | null>(null)
const settingsForm = ref({
  type: '',
  doc_root: '',
  proxy_port: 0,
  php_version: '',
  is_ignored: false,
})

const showDiagModal = ref(false)
const selectedSiteForDiag = ref<SiteItem | null>(null)
const diagResult = ref<any | null>(null)
const isDiagnosing = ref(false)
const isFixingEnv = ref(false)
const discoveredDatabases = ref<any[]>([])
const isImportingDB = ref(false)
const dbImportSuccess = ref(false)

const normalizedProjectsDir = computed(() => {
  return (projectsDirectory.value || '').replace(/\\/g, '/')
})

const normalizedRecentDirs = computed(() => {
  const current = normalizedProjectsDir.value
  const list = recentDirectories.value.map(d => (d || '').replace(/\\/g, '/')).filter(Boolean)
  if (current && !list.includes(current)) {
    list.unshift(current)
  }
  return Array.from(new Set(list))
})

const suggestedPaths = computed(() => {
  const defaults = ['D:/projects', 'C:/laragal/sites', 'C:/projects']
  const set = new Set<string>()
  if (normalizedProjectsDir.value) set.add(normalizedProjectsDir.value)
  for (const p of normalizedRecentDirs.value) {
    if (p) set.add(p)
  }
  for (const d of defaults) {
    set.add(d)
  }
  return Array.from(set).slice(0, 5)
})

async function toggleProjectApp(site: SiteItem) {
  const isRunning = runningProjects.value[site.domain] || false
  const pName = site.name || site.domain
  if (isRunning) {
    if ((window as any).go?.main?.App?.StopProjectApp) {
      await (window as any).go.main.App.StopProjectApp(pName)
    }
    runningProjects.value[site.domain] = false
    notification.value = `تم إيقاف تطبيق ${pName}.`
  } else {
    notification.value = `جاري تشغيل تطبيق ${pName}...`
    if ((window as any).go?.main?.App?.StartProjectApp) {
      try {
        const ok = await (window as any).go.main.App.StartProjectApp(site.path || site.root_path, site.proxy_port || 8080)
        if (ok) {
          runningProjects.value[site.domain] = true
          notification.value = `تم تشغيل خادم تطبيق ${pName} بنجاح على المنفذ :${site.proxy_port || 8080}.`
          return
        }
      } catch (err: any) {
        notification.value = `تعذر تشغيل تطبيق ${pName}: ${err?.message || err}`
        return
      }
    }
    runningProjects.value[site.domain] = true
  }
}

// Synchronize dark class to html document element
watch(isDark, (val) => {
  if (val) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('laragal_theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('laragal_theme', 'light')
  }
}, { immediate: true })

// Synchronize language and text direction to html document element
watch(currentLang, (lang) => {
  document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr'
  document.documentElement.lang = lang
}, { immediate: true })

function toggleTheme() {
  isDark.value = !isDark.value
}

// Window control functions connected to Wails backend
function closeWindow() {
  if ((window as any).go?.main?.App?.CloseWindow) {
    (window as any).go.main.App.CloseWindow()
  } else if ((window as any).runtime?.WindowHide) {
    (window as any).runtime.WindowHide()
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

function promptQuitApp() {
  showQuitConfirm.value = true
}

function executeQuitApp() {
  showQuitConfirm.value = false
  if ((window as any).go?.main?.App?.QuitApp) {
    (window as any).go.main.App.QuitApp()
  } else if ((window as any).runtime?.Quit) {
    (window as any).runtime.Quit()
  }
}

interface SiteItem {
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

function mapSiteItem(item: any): SiteItem {
  const isWeb = item.is_web !== undefined ? Boolean(item.is_web) : (item.IsWeb !== undefined ? Boolean(item.IsWeb) : true)
  const projectType = item.project_type || item.ProjectType || (isWeb ? 'php' : 'non_web')
  let typeLabel = item.type_label || item.TypeLabel || ''
  if (!typeLabel) {
    if (projectType === 'laravel') typeLabel = 'Laravel'
    else if (projectType === 'wordpress') typeLabel = 'WordPress'
    else if (projectType === 'php') typeLabel = 'PHP'
    else if (projectType === 'go') typeLabel = 'Go'
    else if (projectType === 'rust') typeLabel = 'Rust'
    else if (projectType === 'dotnet') typeLabel = '.NET / C#'
    else if (projectType === 'python') typeLabel = 'Python'
    else if (projectType === 'ruby') typeLabel = 'Ruby'
    else if (projectType === 'java') typeLabel = 'Java'
    else if (projectType === 'flutter') typeLabel = 'Flutter Web'
    else if (projectType === 'static') typeLabel = 'HTML / ثابت'
    else if (projectType === 'node') typeLabel = 'Node.js'
    else typeLabel = 'غير ويب / ملفات'
  }
  const rawPhp = item.php_version || item.PHPVersion || ''
  const isPhpType = projectType === 'laravel' || projectType === 'wordpress' || projectType === 'php'
  const phpVersion = rawPhp ? (rawPhp.startsWith('PHP') ? rawPhp : `PHP ${rawPhp}`) : (isWeb && isPhpType ? 'PHP 8.3' : '')

  return {
    name: item.name || item.Name || item.domain || item.Domain || '',
    domain: item.domain || item.Domain || '',
    path: item.path || item.Path || '',
    root_path: item.root_path || item.RootPath || '',
    php_version: phpVersion,
    php_port: item.php_port !== undefined ? item.php_port : (item.PHPPort !== undefined ? item.PHPPort : (isWeb && isPhpType ? 9083 : 0)),
    proxy_port: item.proxy_port || item.ProxyPort || (projectType === 'go' ? 8080 : 0),
    project_type: projectType,
    type_label: typeLabel,
    is_web: isWeb,
  }
}

const sites = ref<SiteItem[]>([])
const activeFilter = ref<'web' | 'all' | 'laravel' | 'php' | 'go' | 'rust' | 'dotnet' | 'python' | 'node' | 'flutter' | 'non_web'>('web')
const isSyncingHosts = ref(false)

const webSites = computed(() => sites.value.filter(s => s.is_web))
const webSitesCount = computed(() => webSites.value.length)
const laravelSitesCount = computed(() => sites.value.filter(s => s.project_type === 'laravel').length)
const phpSitesCount = computed(() => sites.value.filter(s => s.project_type === 'php').length)
const goSitesCount = computed(() => sites.value.filter(s => s.project_type === 'go' || s.project_type === 'proxy').length)
const rustSitesCount = computed(() => sites.value.filter(s => s.project_type === 'rust').length)
const dotnetSitesCount = computed(() => sites.value.filter(s => s.project_type === 'dotnet').length)
const pythonSitesCount = computed(() => sites.value.filter(s => s.project_type === 'python').length)
const nodeSitesCount = computed(() => sites.value.filter(s => s.project_type === 'node').length)
const flutterSitesCount = computed(() => sites.value.filter(s => s.project_type === 'flutter').length)
const nonWebCount = computed(() => sites.value.filter(s => !s.is_web).length)

const filteredSites = computed(() => {
  if (activeFilter.value === 'web') return sites.value.filter(s => s.is_web)
  if (activeFilter.value === 'laravel') return sites.value.filter(s => s.project_type === 'laravel')
  if (activeFilter.value === 'php') return sites.value.filter(s => s.project_type === 'php')
  if (activeFilter.value === 'go') return sites.value.filter(s => s.project_type === 'go' || s.project_type === 'proxy')
  if (activeFilter.value === 'rust') return sites.value.filter(s => s.project_type === 'rust')
  if (activeFilter.value === 'dotnet') return sites.value.filter(s => s.project_type === 'dotnet')
  if (activeFilter.value === 'python') return sites.value.filter(s => s.project_type === 'python')
  if (activeFilter.value === 'node') return sites.value.filter(s => s.project_type === 'node')
  if (activeFilter.value === 'flutter') return sites.value.filter(s => s.project_type === 'flutter')
  if (activeFilter.value === 'non_web') return sites.value.filter(s => !s.is_web)
  return sites.value
})

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
  return Boolean(site.is_web && (site.project_type === 'laravel' || site.project_type === 'wordpress' || site.project_type === 'php' || (site.php_port || 0) > 0))
}

function extractPhpVersionNumber(ver?: string): string {
  if (!ver) return '8.3'
  const clean = ver.replace(/PHP\s*/i, '').trim()
  if (clean.startsWith('8.4')) return '8.4'
  if (clean.startsWith('8.3')) return '8.3'
  if (clean.startsWith('8.2')) return '8.2'
  if (clean.startsWith('8.1')) return '8.1'
  return clean || '8.3'
}

async function updateSitePHPVersion(site: SiteItem, event: Event) {
  const target = event.target as HTMLSelectElement
  const newVer = target.value
  const targetPath = site.path || site.root_path || site.domain
  if (!targetPath || !newVer) return

  if ((window as any).go?.main?.App?.SetProjectPHPVersion) {
    try {
      const updated = await (window as any).go.main.App.SetProjectPHPVersion(targetPath, newVer)
      site.php_version = `PHP ${newVer}`
      site.php_port = newVer === '8.4' ? 9084 : (newVer === '8.3' ? 9083 : (newVer === '8.2' ? 9082 : 9081))
      notification.value = `${t('php_version_changed')} ${site.domain} ➔ PHP ${newVer}`
      if (Array.isArray(updated)) {
        sites.value = updated.map(mapSiteItem)
      }
    } catch (e: any) {
      notification.value = `${t('error_changing_php')}: ${e?.message || e}`
    }
  } else {
    site.php_version = `PHP ${newVer}`
    notification.value = `${t('php_version_changed')} ${site.domain} ➔ PHP ${newVer}`
  }
}

function openSettingsModal(site: SiteItem) {
  selectedSiteForConfig.value = site
  const relDoc = getDocRootRelative(site).replace(/^\//, '')
  settingsForm.value = {
    type: site.project_type || '',
    doc_root: relDoc,
    proxy_port: site.proxy_port || 0,
    php_version: site.php_version ? `PHP ${extractPhpVersionNumber(site.php_version)}` : '',
    is_ignored: false,
  }
  showSettingsModal.value = true
}

async function applyProjectSettings() {
  if (!selectedSiteForConfig.value) return
  const site = selectedSiteForConfig.value
  const targetPath = site.path || site.root_path || site.domain

  const payload = {
    type: settingsForm.value.type,
    doc_root: settingsForm.value.doc_root.trim(),
    proxy_port: Number(settingsForm.value.proxy_port) || 0,
    php_version: settingsForm.value.php_version,
    is_ignored: settingsForm.value.is_ignored,
  }

  if ((window as any).go?.main?.App?.SaveProjectSettings) {
    try {
      const updated = await (window as any).go.main.App.SaveProjectSettings(targetPath, payload)
      if (Array.isArray(updated)) {
        sites.value = updated.map(mapSiteItem)
      }
      notification.value = t('settings_saved')
    } catch (e: any) {
      notification.value = `خطأ أثناء حفظ الإعدادات: ${e?.message || e}`
    }
  } else {
    notification.value = t('settings_saved')
  }

  showSettingsModal.value = false
}

async function runDiagnostics(site: SiteItem) {
  const targetPath = site.path || site.root_path || site.domain
  isDiagnosing.value = true
  discoveredDatabases.value = []
  dbImportSuccess.value = false
  if ((window as any).go?.main?.App?.DiagnoseProject) {
    try {
      const res = await (window as any).go.main.App.DiagnoseProject(targetPath)
      diagResult.value = res
      // Scan for legacy schemas / databases
      if ((window as any).go?.main?.App?.ScanProjectDatabases) {
        const dbs = await (window as any).go.main.App.ScanProjectDatabases(targetPath)
        discoveredDatabases.value = dbs || []
      }
    } catch (e: any) {
      console.warn("Diagnostics error:", e)
    }
  } else {
    diagResult.value = {
      project_path: targetPath,
      domain: site.domain,
      project_type: site.project_type || 'php',
      php_version: site.php_version || 'PHP 8.3',
      php_port: site.php_port || 9083,
      proxy_port: site.proxy_port || 0,
      doc_root: site.root_path,
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

async function openDiagnosticsModal(site: SiteItem) {
  selectedSiteForDiag.value = site
  showDiagModal.value = true
  await runDiagnostics(site)
}

async function triggerFixEnv() {
  if (!selectedSiteForDiag.value) return
  const targetPath = selectedSiteForDiag.value.path || selectedSiteForDiag.value.root_path || selectedSiteForDiag.value.domain
  isFixingEnv.value = true
  if ((window as any).go?.main?.App?.FixProjectEnv) {
    try {
      const ok = await (window as any).go.main.App.FixProjectEnv(targetPath)
      if (ok) {
        notification.value = t('env_fixed_success')
        await runDiagnostics(selectedSiteForDiag.value)
      } else {
        notification.value = t('env_fix_failed')
      }
    } catch (e: any) {
      notification.value = `تعذر إنشاء .env: ${e?.message || e}`
    }
  }
  isFixingEnv.value = false
}

async function importDiscoveredDatabase(candidate: any) {
  if (!selectedSiteForDiag.value) return
  const targetPath = selectedSiteForDiag.value.path || selectedSiteForDiag.value.root_path || selectedSiteForDiag.value.domain
  isImportingDB.value = true
  try {
    if ((window as any).go?.main?.App?.ImportProjectDatabase) {
      const ok = await (window as any).go.main.App.ImportProjectDatabase(targetPath, candidate)
      if (ok) {
        dbImportSuccess.value = true
        notification.value = t('db_import_success')
        await runDiagnostics(selectedSiteForDiag.value)
      }
    }
  } catch (err: any) {
    console.error("Import database error:", err)
    notification.value = `تعذر استيراد قاعدة البيانات: ${err?.message || err}`
  } finally {
    isImportingDB.value = false
  }
}

let toastTimer: any = null
watch(notification, (val) => {
  if (toastTimer) clearTimeout(toastTimer)
  if (val) {
    toastTimer = setTimeout(() => {
      notification.value = ''
    }, 4000)
  }
})

async function loadSites() {
  if ((window as any).go?.main?.App?.GetSites) {
    try {
      const s = await (window as any).go.main.App.GetSites()
      if (Array.isArray(s)) {
        sites.value = s.map(mapSiteItem)
        if (webSitesCount.value === 0 && sites.value.length > 0 && activeFilter.value === 'web') {
          activeFilter.value = 'all'
        } else if (webSitesCount.value > 0 && activeFilter.value === 'all') {
          activeFilter.value = 'web'
        }
      }
    } catch (e) {
      console.warn("Could not load sites:", e)
    }
  }
}

async function loadRecentDirectories() {
  if ((window as any).go?.main?.App?.GetRecentDirectories) {
    try {
      const rec = await (window as any).go.main.App.GetRecentDirectories()
      if (Array.isArray(rec)) {
        recentDirectories.value = rec.map((p: string) => (p || '').replace(/\\/g, '/'))
      }
    } catch (e) {
      console.warn("Could not load recent directories:", e)
    }
  }
}

async function switchRecentDirectory(event: Event) {
  const target = event.target as HTMLSelectElement
  const chosen = (target.value || '').replace(/\\/g, '/')
  if (!chosen || chosen === normalizedProjectsDir.value) return
  projectsDirectory.value = chosen
  if ((window as any).go?.main?.App?.SetProjectsDirectory) {
    try {
      const s = await (window as any).go.main.App.SetProjectsDirectory(chosen)
      notification.value = `${t('projects_dir_updated')} ${chosen}`
      if (Array.isArray(s)) {
        sites.value = s.map(mapSiteItem)
        if (webSitesCount.value === 0 && sites.value.length > 0) {
          activeFilter.value = 'all'
        } else {
          activeFilter.value = 'web'
        }
      }
      await loadRecentDirectories()
    } catch (e: any) {
      notification.value = `${t('error_changing_dir')}: ${e?.message || e}`
    }
  }
}

function changeProjectsDirectory() {
  customDirInput.value = normalizedProjectsDir.value
  showDirModal.value = true
}

async function applyCustomDirectory() {
  const p = customDirInput.value.trim()
  if (!p) return
  showDirModal.value = false
  const norm = p.replace(/\\/g, '/')
  projectsDirectory.value = norm
  if ((window as any).go?.main?.App?.SetProjectsDirectory) {
    try {
      const sitesList = await (window as any).go.main.App.SetProjectsDirectory(norm)
      notification.value = `${t('projects_dir_updated')} ${norm}`
      if (Array.isArray(sitesList)) {
        sites.value = sitesList.map(mapSiteItem)
        if (webSitesCount.value === 0 && sites.value.length > 0 && activeFilter.value === 'web') {
          activeFilter.value = 'all'
        } else if (webSitesCount.value > 0 && activeFilter.value === 'all') {
          activeFilter.value = 'web'
        }
      }
      await loadRecentDirectories()
    } catch (e: any) {
      notification.value = `${t('error_changing_dir')}: ${e?.message || e}`
    }
  } else {
    notification.value = `${t('projects_dir_updated')} ${norm}`
  }
}

async function triggerWindowsPicker() {
  if ((window as any).go?.main?.App?.SelectProjectsDirectory) {
    try {
      const newPath = await (window as any).go.main.App.SelectProjectsDirectory(currentLang.value)
      if (newPath) {
        const norm = newPath.replace(/\\/g, '/')
        customDirInput.value = norm
        projectsDirectory.value = norm
        showDirModal.value = false
        notification.value = `${t('projects_dir_updated')} ${norm}`
        await loadSites()
        await loadRecentDirectories()
      }
    } catch (e: any) {
      console.warn("Picker error:", e)
    }
  }
}

onMounted(async () => {
  // Listen for backend asynchronous updates so the UI reflects live status without freezing
  if ((window as any).runtime?.EventsOn) {
    (window as any).runtime.EventsOn('sites-updated', (s: any) => {
      if (Array.isArray(s)) {
        sites.value = s.map(mapSiteItem)
        if (webSitesCount.value === 0 && sites.value.length > 0 && activeFilter.value === 'web') {
          activeFilter.value = 'all'
        } else if (webSitesCount.value > 0 && activeFilter.value === 'all') {
          activeFilter.value = 'web'
        }
      }
    })
    (window as any).runtime.EventsOn('services-updated', (live: any) => {
      if (Array.isArray(live)) {
        for (const ls of live) {
          const match = services.value.find(s => s.id === ls.name)
          if (match) match.running = Boolean(ls.running)
        }
      }
    })
    (window as any).runtime.EventsOn('download-progress', (data: any) => {
      if (data && data.name) {
        downloadProgress.value[data.name] = Number(data.percent) || 0
      }
    })
  }

  if ((window as any).go?.main?.App?.GetProjectsDirectory) {
    try {
      const p = await (window as any).go.main.App.GetProjectsDirectory()
      if (p) projectsDirectory.value = p.replace(/\\/g, '/')
    } catch (e) {
      console.warn("Could not load projects directory:", e)
    }
  }
  Promise.all([
    loadRecentDirectories(),
    loadSites(),
    loadServices(),
    checkPortConflicts()
  ])
})

const downloadProgress = ref<Record<string, number>>({})

const availableRuntimes = ref([
  { name: 'caddy', label: 'Caddy Server v2.8', desc: 'خادم الويب الرئيسي وإدارة شهادات SSL التلقائية' },
  { name: 'php-8.4', label: 'PHP 8.4 (NTS x64)', desc: 'أحدث إصدار رسمي لمحرك PHP' },
  { name: 'php-8.3', label: 'PHP 8.3 (NTS x64)', desc: 'إصدار مستقر موصى به لـ Laravel 11/12' },
  { name: 'php-8.2', label: 'PHP 8.2 (NTS x64)', desc: 'إصدار مستقر متوافق مع المشاريع السابقة' },
  { name: 'php-8.1', label: 'PHP 8.1 (NTS x64)', desc: 'إصدار متوافق مع مشاريع PHP 8.1 القديمة' },
  { name: 'mysql', label: 'MariaDB / MySQL Server', desc: 'خادم قواعد البيانات المحمول مع التهيئة التلقائية' },
  { name: 'redis', label: 'Redis Server (Windows)', desc: 'خادم الكاش وتخزين الجلسات محلياً' },
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
  id: string
  name_key: string
  desc_key: string
  port: number
  running: boolean
  url?: string
  credentials?: {
    host?: string
    user?: string
    password?: string
    database?: string
    smtp_host?: string
    smtp_port?: number
    web_port?: number
    info?: string
  }
}

const services = ref<ServiceItem[]>([
  {
    id: 'caddy',
    name_key: 'service_caddy',
    desc_key: 'service_caddy_desc',
    port: 443,
    running: true,
    url: 'http://localhost:2019',
    credentials: {
      host: 'localhost',
      info: 'Admin: :2019 / SSL Root'
    }
  },
  {
    id: 'php-8.4',
    name_key: 'service_php84',
    desc_key: 'service_php84_desc',
    port: 9084,
    running: true,
    credentials: {
      host: '127.0.0.1',
      info: 'FastCGI / 8 Workers'
    }
  },
  {
    id: 'php-8.3',
    name_key: 'service_php83',
    desc_key: 'service_php83_desc',
    port: 9083,
    running: true,
    credentials: {
      host: '127.0.0.1',
      info: 'FastCGI / 8 Workers'
    }
  },
  {
    id: 'mysql',
    name_key: 'service_mysql',
    desc_key: 'service_mysql_desc',
    port: 3306,
    running: true,
    credentials: {
      host: '127.0.0.1',
      user: 'root',
      password: '',
      database: 'laravel'
    }
  },
  {
    id: 'redis',
    name_key: 'service_redis',
    desc_key: 'service_redis_desc',
    port: 6379,
    running: false,
    credentials: {
      host: '127.0.0.1',
      password: ''
    }
  },
  {
    id: 'mailpit',
    name_key: 'service_mailpit',
    desc_key: 'service_mailpit_desc',
    port: 1025,
    running: true,
    url: 'http://127.0.0.1:8025',
    credentials: {
      smtp_host: '127.0.0.1',
      smtp_port: 1025,
      web_port: 8025
    }
  },
])


async function syncHostsNow() {
  isSyncingHosts.value = true
  try {
    if ((window as any).go?.main?.App?.SyncHosts) {
      const msg = await (window as any).go.main.App.SyncHosts()
      notification.value = msg
    } else {
      notification.value = 'تمت مزامنة النطاقات مع ملف hosts وتحديث DNS بنجاح.'
    }
  } catch (err: any) {
    notification.value = `حدث خطأ أثناء مزامنة النطاقات: ${err?.message || err}`
  } finally {
    isSyncingHosts.value = false
  }
}

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

function openNewModal() {
  projectError.value = ''
  newProject.value.name = ''
  showNewModal.value = true
}

// Instant custom field validation
function validateProjectName(): boolean {
  const val = newProject.value.name.trim()
  if (!val) {
    projectError.value = t('val_name_required')
    return false
  }
  if (!/^[a-z0-9-]+$/.test(val)) {
    projectError.value = t('val_name_invalid')
    return false
  }
  const domain = `${val.toLowerCase()}.test`
  if (sites.value.some(s => s.domain.toLowerCase() === domain)) {
    projectError.value = t('val_name_exists')
    return false
  }
  projectError.value = ''
  return true
}

async function loadServices() {
  if ((window as any).go?.main?.App?.GetServices) {
    try {
      const live = await (window as any).go.main.App.GetServices()
      if (Array.isArray(live)) {
        for (const ls of live) {
          const match = services.value.find(s => s.id === ls.name)
          if (match) {
            match.running = Boolean(ls.running)
          }
        }
      }
    } catch (e) {
      console.warn("Could not load services:", e)
    }
  }
}

async function checkPortConflicts() {
  if ((window as any).go?.main?.App?.CheckPortConflict) {
    try {
      const msg = await (window as any).go.main.App.CheckPortConflict(443)
      if (msg) {
        notification.value = msg
      }
    } catch (e) {
      console.warn("CheckPortConflict err:", e)
    }
  }
}

// Custom service toggle with custom confirmation modal (No native confirm())
async function promptToggleService(service: ServiceItem) {
  if (service.running) {
    // Open custom accessible confirmation modal
    confirmService.value = service
  } else {
    if ((window as any).go?.main?.App?.ToggleService) {
      try {
        const running = await (window as any).go.main.App.ToggleService(service.id)
        service.running = running
        notification.value = `تم تشغيل ${t(service.name_key)} بنجاح.`
      } catch (err: any) {
        notification.value = `تعذر تشغيل ${t(service.name_key)}: ${err?.message || err}`
      }
    } else {
      service.running = true
      notification.value = `تم تشغيل ${t(service.name_key)} بنجاح.`
    }
  }
}

async function executeStopService() {
  if (!confirmService.value) return
  const s = confirmService.value
  if ((window as any).go?.main?.App?.ToggleService) {
    try {
      const running = await (window as any).go.main.App.ToggleService(s.id)
      s.running = running
      notification.value = `تم إيقاف ${t(s.name_key)} بنجاح.`
    } catch (err: any) {
      notification.value = `تعذر إيقاف ${t(s.name_key)}: ${err?.message || err}`
    }
  } else {
    s.running = false
    notification.value = `تم إيقاف ${t(s.name_key)} بنجاح.`
  }
  confirmService.value = null
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
  // Validate custom rules before proceeding
  if (!validateProjectName()) return

  const projectName = newProject.value.name.trim().toLowerCase()
  const domain = `${projectName}.test`
  const port = newProject.value.php === '8.4' ? 9084 : (newProject.value.php === '8.3' ? 9083 : 9082)
  const rootPath = `${projectsDirectory.value.replace(/\\/g, '/')}/${projectName}/public`

  if ((window as any).go?.main?.App?.CreateProject) {
    try {
      await (window as any).go.main.App.CreateProject(projectName, newProject.value.type, newProject.value.php)
      await loadSites()
    } catch (e: any) {
      console.warn("Go backend create project:", e)
    }
  } else {
    sites.value.push(mapSiteItem({
      name: projectName,
      domain,
      root_path: rootPath,
      php_version: `PHP ${newProject.value.php}`,
      php_port: port,
      project_type: newProject.value.type,
      type_label: newProject.value.type === 'laravel' ? 'Laravel' : 'PHP',
      is_web: true,
    }))
  }

  notification.value = `تم إنشاء المشروع «${projectName}» بنجاح وضبط رابط https://${domain} الموثوق بـ SSL.`
  showNewModal.value = false
  newProject.value.name = ''
  projectError.value = ''
}

async function triggerDownloadRuntime(name: string) {
  downloadingRuntime.value = name
  downloadProgress.value[name] = 0
  notification.value = `جاري بدء تنزيل وتثبيت حزمة ${name}...`

  if ((window as any).go?.main?.App?.DownloadBinary) {
    try {
      const msg = await (window as any).go.main.App.DownloadBinary(name)
      downloadingRuntime.value = ''
      downloadProgress.value[name] = 100
      notification.value = msg || `تم تجهيز وتثبيت حزمة ${name} بنجاح.`
      return
    } catch (err: any) {
      console.warn("Download err:", err)
      downloadingRuntime.value = ''
      notification.value = `خطأ أثناء تنزيل حزمة ${name}: ${err?.message || err}`
      return
    }
  }

  const interval = setInterval(() => {
    if ((downloadProgress.value[name] || 0) < 95) {
      downloadProgress.value[name] = (downloadProgress.value[name] || 0) + 15
    }
  }, 200)

  setTimeout(() => {
    clearInterval(interval)
    downloadProgress.value[name] = 100
    downloadingRuntime.value = ''
    notification.value = `تم تجهيز وتثبيت حزمة ${name} بنجاح في مجلد bin/${name}.`
  }, 1500)
}
</script>
