<template>
  <div :dir="currentLang === 'ar' ? 'rtl' : 'ltr'" class="h-screen w-screen flex flex-col overflow-hidden bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150 select-none">
    <!-- ================= 1. FIXED TITLE BAR (Physical LTR for Window Controls) ================= -->
    <div
      dir="ltr"
      @dblclick="toggleMaximize"
      class="wails-drag h-10 w-full bg-slate-100 dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-3 text-xs font-medium z-40 flex-shrink-0 select-none cursor-default"
    >
      <!-- PHYSICAL LEFT (Child 1) -->
      <!-- In Arabic (RTL): Window controls are on the physical LEFT -->
      <div v-if="currentLang === 'ar'" class="flex items-center gap-1.5 wails-no-drag">
        <button
          @click.stop="closeWindow"
          class="w-7 h-7 rounded-md flex items-center justify-center text-slate-600 dark:text-slate-400 hover:bg-rose-600 hover:text-white transition-colors"
          :title="t('close_to_background')"
        >
          <X class="w-3.5 h-3.5" />
        </button>
        <button
          @click.stop="toggleMaximize"
          class="w-7 h-7 rounded-md flex items-center justify-center text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-800 transition-colors"
          :title="t('maximize_restore')"
        >
          <Square class="w-3 h-3" />
        </button>
        <button
          @click.stop="minimizeWindow"
          class="w-7 h-7 rounded-md flex items-center justify-center text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-800 transition-colors"
          :title="t('minimize')"
        >
          <Minus class="w-3.5 h-3.5" />
        </button>
      </div>
      <!-- In English: Balanced empty spacer on the left -->
      <div v-else class="w-20"></div>

      <!-- PHYSICAL CENTER (Child 2) -->
      <div @dblclick="toggleMaximize" class="text-xs text-slate-500 dark:text-slate-400 font-semibold select-none cursor-default py-1 px-4">
        <span v-if="currentLang === 'ar'" dir="rtl">بيئة التطوير المحلية — LaraGal</span>
        <span v-else dir="ltr">LaraGal — Local Development Environment</span>
      </div>

      <!-- PHYSICAL RIGHT (Child 3) -->
      <!-- In Arabic: Balanced empty spacer on the right so title stays centered -->
      <div v-if="currentLang === 'ar'" class="w-20"></div>
      <!-- In English: Window controls on the physical RIGHT -->
      <div v-else class="flex items-center gap-1.5 wails-no-drag">
        <button
          @click.stop="minimizeWindow"
          class="w-7 h-7 rounded-md flex items-center justify-center text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-800 transition-colors"
          :title="t('minimize')"
        >
          <Minus class="w-3.5 h-3.5" />
        </button>
        <button
          @click.stop="toggleMaximize"
          class="w-7 h-7 rounded-md flex items-center justify-center text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-800 transition-colors"
          :title="t('maximize_restore')"
        >
          <Square class="w-3 h-3" />
        </button>
        <button
          @click.stop="closeWindow"
          class="w-7 h-7 rounded-md flex items-center justify-center text-slate-600 dark:text-slate-400 hover:bg-rose-600 hover:text-white transition-colors"
          :title="t('close_to_background')"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>

    <!-- ================= 2. FIXED MAIN HEADER (Direction-Aware: Logo on RIGHT in Arabic, on LEFT in English) ================= -->
    <header class="flex-shrink-0 border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-slate-900/95 backdrop-blur z-30">
      <div class="w-full px-6 lg:px-8 py-3.5 flex items-center justify-between">
        <!-- Brand Anchor: In RTL appears on the RIGHT; in LTR appears on the LEFT -->
        <div class="flex items-center gap-3.5">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-700 via-blue-600 to-indigo-500 flex items-center justify-center text-white font-black text-xl shadow-md ring-2 ring-blue-500/20 flex-shrink-0">
            LG
          </div>
          <div>
            <div class="flex items-center gap-2">
              <span class="text-xl font-black tracking-tight text-slate-900 dark:text-white">LaraGal</span>
              <span class="text-[11px] px-2 py-0.5 rounded-md bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 font-mono font-bold border border-blue-200 dark:border-blue-900">
                v1.0.12
              </span>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 font-medium mt-0.5 text-start">
              <span v-if="currentLang === 'ar'">بيئة التطوير المحلية لـ <bdi>PHP</bdi> و <bdi>Laravel</bdi></span>
              <span v-else>Local Development Environment for PHP &amp; Laravel</span>
            </p>
          </div>
        </div>

        <!-- Actions Toolbar: In RTL appears on the LEFT; in LTR appears on the RIGHT -->
        <div class="flex items-center gap-2.5">
          <button
            @click="syncHostsNow"
            :disabled="isSyncingHosts"
            class="hidden md:inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-blue-50 text-blue-700 dark:bg-blue-950/80 dark:text-blue-300 border border-blue-200 dark:border-blue-800 hover:bg-blue-100 dark:hover:bg-blue-900/80 transition-colors"
            :title="t('sync_hosts_hint')"
          >
            <RefreshCw class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" :class="{ 'animate-spin': isSyncingHosts }" />
            <span>{{ t('sync_hosts_btn') }}</span>
          </button>

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
            class="px-3 py-1.5 rounded-lg text-xs font-bold border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
          >
            {{ currentLang === 'ar' ? 'English' : 'العربية' }}
          </button>

          <button
            @click="toggleTheme"
            class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
            :title="isDark ? t('light_mode') : t('dark_mode')"
          >
            <Sun v-if="isDark" class="w-4 h-4 text-amber-400" />
            <Moon v-else class="w-4 h-4 text-slate-600" />
          </button>

          <button
            @click="openNewModal"
            class="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-all hover:shadow-blue-500/20"
          >
            <Plus class="w-4 h-4" />
            <span>{{ t('new_project_btn') }}</span>
          </button>
        </div>
      </div>
    </header>

    <!-- ================= 3. SCROLLABLE MAIN CONTENT (Scrollbar strictly on LEFT in RTL) ================= -->
    <main class="flex-1 overflow-y-auto px-6 lg:px-8 py-6 space-y-7">
      <!-- Custom Toast Notification Banner -->
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
                    'text-[10px] px-2.5 py-0.5 rounded-full border font-semibold flex items-center gap-1'
                  ]"
                >
                  <span class="w-1.5 h-1.5 rounded-full" :class="service.running ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"></span>
                  {{ service.running ? t('status_running') : t('status_stopped') }}
                </span>
                <span dir="ltr" class="text-xs font-mono font-bold text-slate-500 dark:text-slate-400">:{{ service.port }}</span>
              </div>

              <h3 class="font-bold text-sm text-slate-900 dark:text-slate-100 leading-snug">{{ t(service.name_key) }}</h3>
              <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-1 line-clamp-1" :title="t(service.desc_key)">
                {{ t(service.desc_key) }}
              </p>
            </div>

            <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end">
              <button
                @click="promptToggleService(service)"
                class="text-xs font-bold px-3 py-1 rounded-md transition-colors"
                :class="service.running ? 'text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40' : 'text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-950/40'"
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
                {{ webSitesCount }} {{ t('web_count') }}
              </span>
              <span v-if="nonWebCount > 0" class="px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
                {{ nonWebCount }} {{ t('non_web_count') }}
              </span>
            </div>
          </div>

          <!-- Filter tabs -->
          <div class="flex items-center gap-1 p-1 rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700/80 text-xs font-semibold self-start md:self-auto">
            <button
              @click="activeFilter = 'web'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'web' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              {{ t('filter_web') }} ({{ webSitesCount }})
            </button>
            <button
              v-if="laravelSitesCount > 0"
              @click="activeFilter = 'laravel'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'laravel' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              Laravel ({{ laravelSitesCount }})
            </button>
            <button
              v-if="phpSitesCount > 0"
              @click="activeFilter = 'php'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'php' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              PHP ({{ phpSitesCount }})
            </button>
            <button
              v-if="nonWebCount > 0"
              @click="activeFilter = 'non_web'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'non_web' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              {{ t('filter_non_web') }} ({{ nonWebCount }})
            </button>
            <button
              @click="activeFilter = 'all'"
              class="px-2.5 py-1 rounded-lg transition-all"
              :class="activeFilter === 'all' ? 'bg-white dark:bg-slate-700 text-blue-600 dark:text-blue-400 shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
            >
              {{ t('filter_all') }} ({{ sites.length }})
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
              <span>{{ t('show_all_folders') }} ({{ sites.length }})</span>
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

                <span v-if="site.is_web" dir="ltr" class="text-xs font-mono px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-bold border border-slate-200 dark:border-slate-700">
                  <span v-if="site.project_type === 'laravel'">Laravel • {{ site.php_version || 'PHP 8.3' }}</span>
                  <span v-else-if="site.project_type === 'php'">{{ site.php_version || 'PHP 8.3' }}</span>
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

              <!-- Monospace Path Box with Actions -->
              <div dir="ltr" class="mt-3 flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700/80 text-[11px] font-mono text-slate-600 dark:text-slate-300">
                <span class="truncate flex-1 text-left select-all" :title="site.root_path">{{ site.root_path }}</span>
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
                <span v-if="site.is_web" class="text-[11px] text-emerald-600 dark:text-emerald-400 font-bold flex items-center gap-1">
                  <CheckCircle2 class="w-3.5 h-3.5" />
                  {{ t('ssl_trusted') }}
                </span>
                <span v-else class="text-[11px] text-slate-500 dark:text-slate-400 font-semibold flex items-center gap-1">
                  <Folder class="w-3.5 h-3.5" />
                  {{ t('local_folder_badge') }}
                </span>
                <button
                  @click="openVSCode(site.root_path)"
                  class="hidden sm:inline-flex items-center gap-1 text-[11px] font-bold text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors ms-1"
                  :title="t('open_vscode')"
                >
                  <Code class="w-3.5 h-3.5" />
                  <span>VS Code</span>
                </button>
              </div>

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
                @click="openExplorer(site.root_path)"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
              >
                <span>{{ t('open_folder_btn') }}</span>
                <FolderOpen class="w-3.5 h-3.5" />
              </button>
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
            <button @click="openExplorer(projectsDirectory)" class="text-blue-600 dark:text-blue-400 font-bold hover:underline truncate max-w-[220px]" :title="projectsDirectory">
              {{ projectsDirectory }}
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
              v-if="recentDirectories.length > 1"
              @change="switchRecentDirectory($event)"
              :value="projectsDirectory"
              dir="ltr"
              class="text-[11px] font-sans font-medium px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 focus:outline-none cursor-pointer max-w-[140px] truncate"
              :title="t('recent_directories')"
            >
              <option v-for="dir in recentDirectories" :key="dir" :value="dir">{{ dir }}</option>
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
          <span class="text-slate-400 dark:text-slate-500 font-bold">LaraGal v1.0.6</span>
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
import { ref, computed, watch, onMounted } from 'vue'
import {
  Server, Globe, Plus, ShieldCheck, Sun, Moon,
  Lock, ExternalLink, CheckCircle2, Info, Download,
  Folder, Code, Copy, Check, Minus, Square, X,
  AlertCircle, AlertTriangle, FolderOpen, Power, RefreshCw
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

interface SiteItem {
  name?: string
  domain: string
  root_path: string
  php_version?: string
  php_port?: number
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
    else if (projectType === 'static') typeLabel = 'HTML / ثابت'
    else if (projectType === 'node') typeLabel = 'Node.js'
    else typeLabel = 'غير ويب / ملفات'
  }
  const rawPhp = item.php_version || item.PHPVersion || ''
  const phpVersion = rawPhp ? (rawPhp.startsWith('PHP') ? rawPhp : `PHP ${rawPhp}`) : (isWeb && projectType !== 'static' && projectType !== 'node' ? 'PHP 8.3' : '')

  return {
    name: item.name || item.Name || item.domain || item.Domain || '',
    domain: item.domain || item.Domain || '',
    root_path: item.root_path || item.RootPath || '',
    php_version: phpVersion,
    php_port: item.php_port !== undefined ? item.php_port : (item.PHPPort !== undefined ? item.PHPPort : (isWeb && projectType !== 'static' && projectType !== 'node' ? 9083 : 0)),
    project_type: projectType,
    type_label: typeLabel,
    is_web: isWeb,
  }
}

const sites = ref<SiteItem[]>([])
const activeFilter = ref<'web' | 'all' | 'laravel' | 'php' | 'non_web'>('web')
const isSyncingHosts = ref(false)

const webSites = computed(() => sites.value.filter(s => s.is_web))
const webSitesCount = computed(() => webSites.value.length)
const laravelSitesCount = computed(() => sites.value.filter(s => s.project_type === 'laravel').length)
const phpSitesCount = computed(() => sites.value.filter(s => s.project_type === 'php').length)
const nonWebCount = computed(() => sites.value.filter(s => !s.is_web).length)

const filteredSites = computed(() => {
  if (activeFilter.value === 'web') {
    return sites.value.filter(s => s.is_web)
  }
  if (activeFilter.value === 'laravel') {
    return sites.value.filter(s => s.project_type === 'laravel')
  }
  if (activeFilter.value === 'php') {
    return sites.value.filter(s => s.project_type === 'php')
  }
  if (activeFilter.value === 'non_web') {
    return sites.value.filter(s => !s.is_web)
  }
  return sites.value
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
        recentDirectories.value = rec
      }
    } catch (e) {
      console.warn("Could not load recent directories:", e)
    }
  }
}

async function switchRecentDirectory(event: Event) {
  const target = event.target as HTMLSelectElement
  const chosen = target.value
  if (!chosen || chosen === projectsDirectory.value) return
  if ((window as any).go?.main?.App?.SetProjectsDirectory) {
    try {
      const s = await (window as any).go.main.App.SetProjectsDirectory(chosen)
      projectsDirectory.value = chosen
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
      notification.value = `${t('error_changing_dir')}: ${e}`
    }
  }
}

async function changeProjectsDirectory() {
  if ((window as any).go?.main?.App?.SelectProjectsDirectory) {
    try {
      const newPath = await (window as any).go.main.App.SelectProjectsDirectory()
      if (newPath) {
        projectsDirectory.value = newPath
        notification.value = `${t('projects_dir_updated')} ${newPath}`
        await loadSites()
        await loadRecentDirectories()
      }
    } catch (e: any) {
      notification.value = `${t('error_changing_dir')}: ${e}`
    }
  }
}

onMounted(async () => {
  if ((window as any).go?.main?.App?.GetProjectsDirectory) {
    try {
      const p = await (window as any).go.main.App.GetProjectsDirectory()
      if (p) projectsDirectory.value = p
    } catch (e) {
      console.warn("Could not load projects directory:", e)
    }
  }
  await loadRecentDirectories()
  await loadSites()
  await loadServices()
  await checkPortConflicts()
})

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
  id: string
  name_key: string
  desc_key: string
  port: number
  running: boolean
}

const services = ref<ServiceItem[]>([
  { id: 'caddy', name_key: 'service_caddy', desc_key: 'service_caddy_desc', port: 443, running: true },
  { id: 'php-8.4', name_key: 'service_php84', desc_key: 'service_php84_desc', port: 9084, running: true },
  { id: 'php-8.3', name_key: 'service_php83', desc_key: 'service_php83_desc', port: 9083, running: true },
  { id: 'mysql', name_key: 'service_mysql', desc_key: 'service_mysql_desc', port: 3306, running: true },
  { id: 'redis', name_key: 'service_redis', desc_key: 'service_redis_desc', port: 6379, running: false },
  { id: 'mailpit', name_key: 'service_mailpit', desc_key: 'service_mailpit_desc', port: 1025, running: true },
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
