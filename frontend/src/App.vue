<!--
  App.vue: Central orchestrator for desktop layout, global Wails window controls,
  and reactive service/project stores. Modals and sub-panels are modularized under
  frontend/src/components/modals/ and frontend/src/components/design-tools/.
  This file remains above 500 lines due to coordinating top-level global state and
  orchestration (desktop window lifecycle, multi-service stores, and live sites directory watcher)
  to prevent state fragmentation across independent panels.
-->
<template>
  <div :dir="currentLang === 'ar' ? 'rtl' : 'ltr'" class="h-screen w-screen flex flex-col overflow-hidden bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150 select-none">
    <!-- ================= UNIFIED DESKTOP HEADER ================= -->
    <header
      @dblclick="toggleMaximize"
      class="wails-drag flex-shrink-0 h-14 border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-slate-900/95 backdrop-blur-md z-40 px-4 flex items-center justify-between select-none cursor-default"
    >
      <!-- Brand Block -->
      <div class="flex items-center gap-3 wails-no-drag shrink-0">
        <img src="/logo.png" :alt="t('app_name')" class="w-8 h-8 rounded-lg shadow-xs ring-1 ring-blue-500/20 object-contain shrink-0" />
        <div class="flex items-center gap-2">
          <span class="text-lg font-black tracking-tight text-slate-900 dark:text-white">{{ t('app_name') }}</span>
          <span dir="ltr" class="text-[10px] px-1.5 py-0.5 rounded-md bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300 font-mono font-bold border border-blue-200 dark:border-blue-900">
            {{ appVersion }}
          </span>
        </div>
        <div class="hidden lg:block h-5 w-px bg-slate-200 dark:bg-slate-800"></div>
        <p
          class="hidden lg:block text-xs text-slate-500 dark:text-slate-400 font-medium truncate max-w-md"
          :title="t('app_subtitle')"
        >
          {{ t('app_subtitle') }}
        </p>
      </div>

      <!-- CENTER: Draggable Window Space -->
      <div class="flex-1 h-full cursor-default"></div>

      <!-- Actions Toolbar & Window Controls -->
      <div class="flex items-center gap-2.5 wails-no-drag shrink-0">
        <button
          @click="openNewModal"
          class="inline-flex items-center justify-center gap-1.5 h-8 px-3 rounded-lg text-xs font-bold bg-blue-600 hover:bg-blue-700 text-white shadow-xs transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0 cursor-pointer"
          :aria-label="t('new_project_btn')"
        >
          <Plus class="w-3.5 h-3.5 shrink-0" />
          <span>{{ t('new_project_btn') }}</span>
        </button>

        <button
          @click="syncHostsNow"
          :disabled="isSyncingHosts"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 disabled:opacity-60 disabled:cursor-not-allowed shrink-0 cursor-pointer"
          :title="t('sync_hosts_hint')"
          :aria-label="t('sync_hosts_btn')"
        >
          <RefreshCw class="w-3.5 h-3.5 shrink-0 text-blue-600 dark:text-blue-400" :class="{ 'animate-spin': isSyncingHosts }" />
          <span class="hidden xl:inline">{{ t('sync_hosts_btn') }}</span>
        </button>

        <button
          @click="openAdminer()"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0 cursor-pointer"
          :title="t('open_adminer_title')"
          :aria-label="t('open_adminer')"
        >
          <Database class="w-3.5 h-3.5 shrink-0 text-amber-600 dark:text-amber-400" />
          <span class="hidden xl:inline">{{ t('open_adminer') }}</span>
        </button>

        <button
          @click="openPhpMyAdmin()"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0 cursor-pointer"
          :title="t('open_phpmyadmin_title')"
          :aria-label="t('open_phpmyadmin')"
        >
          <Database class="w-3.5 h-3.5 shrink-0 text-orange-600 dark:text-orange-400" />
          <span class="hidden xl:inline">{{ t('open_phpmyadmin') }}</span>
        </button>

        <button
          @click="openDesignToolsModal()"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-indigo-200 dark:border-indigo-800 bg-indigo-50 dark:bg-indigo-950/80 text-indigo-700 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-900/80 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/50 shrink-0 cursor-pointer"
          :title="t('design_tools_desc')"
          :aria-label="t('design_tools_btn')"
        >
          <Layout class="w-3.5 h-3.5 shrink-0 text-indigo-600 dark:text-indigo-400" />
          <span class="hidden xl:inline">{{ t('design_tools_btn') }}</span>
        </button>

        <button
          @click="installSSL"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0 cursor-pointer"
          :title="t('ssl_install_hint')"
          :aria-label="t('install_ssl_btn')"
        >
          <ShieldCheck class="w-3.5 h-3.5 shrink-0 text-emerald-600 dark:text-emerald-400" />
          <span class="hidden xl:inline">{{ t('install_ssl_btn') }}</span>
        </button>

        <button
          @click="toggleLang"
          class="inline-flex items-center justify-center h-8 px-2.5 rounded-lg text-xs font-bold border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0 cursor-pointer"
          :title="currentLang === 'ar' ? 'English' : 'العربية'"
          :aria-label="currentLang === 'ar' ? 'English' : 'العربية'"
        >
          <span v-if="currentLang === 'ar'" lang="en">English</span>
          <span v-else lang="ar">العربية</span>
        </button>

        <button
          @click="toggleTheme"
          class="inline-flex items-center justify-center w-8 h-8 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0 cursor-pointer"
          :title="isDark ? t('light_mode') : t('dark_mode')"
          :aria-label="isDark ? t('light_mode') : t('dark_mode')"
        >
          <Sun v-if="isDark" class="w-3.5 h-3.5 text-amber-400" />
          <Moon v-else class="w-3.5 h-3.5 text-slate-600 dark:text-slate-300" />
        </button>

        <div class="h-5 w-px bg-slate-200 dark:bg-slate-800 mx-1 shrink-0"></div>

        <!-- Window Controls -->
        <div class="flex items-center gap-1 shrink-0">
          <button
            @click.stop="minimizeWindow"
            class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 cursor-pointer"
            :title="t('minimize')"
            :aria-label="t('minimize')"
          >
            <Minus class="w-3.5 h-3.5" />
          </button>
          <button
            @click.stop="toggleMaximize"
            class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 cursor-pointer"
            :title="t('maximize_restore')"
            :aria-label="t('maximize_restore')"
          >
            <Square class="w-3 h-3" />
          </button>
          <button
            @click.stop="closeWindow"
            class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-rose-600 hover:text-white transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 cursor-pointer"
            :title="t('close_to_background')"
            :aria-label="t('close_to_background')"
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
        <button @click="notification = ''" class="ms-2 opacity-70 hover:opacity-100 text-sm font-bold leading-none cursor-pointer">&times;</button>
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
            class="text-[11px] font-bold text-blue-600 hover:text-blue-700 dark:text-blue-400 hover:underline ms-1 cursor-pointer"
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
              class="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400 font-semibold flex items-center gap-1 hover:underline cursor-pointer"
            >
              <Download class="w-3.5 h-3.5" />
              <span>{{ t('download_runtimes_btn') }}</span>
            </button>
            <span class="text-xs text-slate-500 dark:text-slate-400 font-mono"><bdi>{{ services.filter(s => s.running).length }}/{{ services.length }}</bdi> {{ t('status_running') }}</span>
          </div>
        </div>

        <!-- Services Grid -->
        <div class="grid gap-3 grid-cols-[repeat(auto-fill,minmax(15rem,1fr))]">
          <ServiceCard
            v-for="service in services"
            :key="service.id"
            :service="service"
            :busy="togglingServices.has(service.id)"
            :copied-item="copiedItem"
            :t="t"
            @toggle="promptToggleService(service)"
            @open-adminer="openAdminer()"
            @open-phpmyadmin="openPhpMyAdmin()"
            @open-url="openBrowser"
            @copy="copyText"
          />
        </div>
      </section>

      <!-- Sites & Projects Section -->
      <section>
        <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 mb-3">
          <div class="flex flex-wrap items-center gap-2.5">
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

            <!-- Single SSL trusted note in section header -->
            <span class="inline-flex items-center gap-1 text-[11px] font-medium text-emerald-700 dark:text-emerald-400 px-2 py-0.5 rounded-md bg-emerald-50/80 dark:bg-emerald-950/40 border border-emerald-200/60 dark:border-emerald-800/60">
              <Lock class="w-3 h-3 text-emerald-600 dark:text-emerald-400 shrink-0" />
              <span>{{ t('ssl_trusted') }}</span>
            </span>
          </div>

          <!-- Search Input -->
          <div class="relative w-full lg:w-72">
            <Search class="w-4 h-4 text-slate-400 absolute start-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
            <input
              v-model="searchQuery"
              type="search"
              :placeholder="t('search_placeholder')"
              :aria-label="t('search_aria_label')"
              class="w-full ps-8 pe-8 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800/90 text-xs text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-colors"
            />
            <button
              v-if="searchQuery"
              type="button"
              @click="searchQuery = ''"
              class="absolute end-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500 cursor-pointer"
              :title="t('clear_search_btn')"
              :aria-label="t('clear_search_btn')"
            >
              <X class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- Filter Chips Row -->
        <div class="flex items-center gap-1.5 pb-2 mb-3 overflow-x-auto text-xs font-semibold">
          <button
            v-for="f in filterOptions"
            :key="f.id"
            type="button"
            :aria-pressed="activeFilter === f.id"
            @click="activeFilter = f.id"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 cursor-pointer"
            :class="activeFilter === f.id ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>{{ f.label }}</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === f.id ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ f.count }}
            </bdi>
          </button>
        </div>

        <!-- Empty State -->
        <div v-if="filteredSites.length === 0" class="p-12 text-center rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 bg-white dark:bg-slate-900/60">
          <Globe class="w-10 h-10 text-slate-400 mx-auto mb-3 opacity-60" />
          <p class="text-sm text-slate-500 dark:text-slate-400 max-w-sm mx-auto">
            <span v-if="searchQuery.trim()">
              {{ t('no_search_results', { query: searchQuery.trim() }) }}
            </span>
            <span v-else-if="activeFilter === 'web' && sites.length > 0">
              {{ t('no_web_sites_in_dir') }}
            </span>
            <span v-else>{{ t('empty_sites') }}</span>
          </p>
          <div class="mt-4 flex items-center justify-center gap-2">
            <button
              v-if="searchQuery.trim()"
              type="button"
              @click="searchQuery = ''"
              class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 cursor-pointer"
            >
              <X class="w-3.5 h-3.5" />
              <span>{{ t('clear_search_btn') }}</span>
            </button>
            <template v-else>
              <button
                v-if="activeFilter === 'web' && sites.length > 0"
                type="button"
                @click="activeFilter = 'all'"
                class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 cursor-pointer"
              >
                <Folder class="w-3.5 h-3.5" />
                <span>{{ t('show_all_folders') }} (<bdi>{{ sites.length }}</bdi>)</span>
              </button>
              <button
                type="button"
                @click="openNewModal"
                class="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold bg-blue-600 text-white hover:bg-blue-700 shadow-xs transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 cursor-pointer"
              >
                <Plus class="w-4 h-4" />
                {{ t('new_project_btn') }}
              </button>
            </template>
          </div>
        </div>

        <!-- Project Row List Container -->
        <div v-else class="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 divide-y divide-slate-100 dark:divide-slate-800/80 shadow-2xs overflow-hidden">
          <ProjectRow
            v-for="site in filteredSites"
            :key="site.domain || site.root_path"
            :site="site"
            :t="t"
            :copied-item="copiedItem"
            :is-running="Boolean(runningProjects[site.domain])"
            :doc-root-relative="getDocRootRelative(site)"
            :is-php="isPhpProject(site)"
            @open-browser="openBrowser('https://' + site.domain)"
            @open-folder="openExplorer(site.path || site.root_path)"
            @open-vscode="openVSCode(site.path || site.root_path)"
            @open-adminer="openAdminer(site.name ? site.name.replace(/-/g, '_') : '')"
            @copy="copyText"
            @settings="openSettingsModal(site)"
            @diagnostics="openDiagnosticsModal(site)"
            @toggle-app="toggleProjectApp(site)"
            @change-php="updateSitePHPVersion(site, $event)"
          />
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
            <button @click="openExplorer(normalizedProjectsDir)" class="text-blue-600 dark:text-blue-400 font-bold hover:underline truncate max-w-[220px] cursor-pointer" :title="normalizedProjectsDir">
              {{ normalizedProjectsDir }}
            </button>
            <button
              @click="changeProjectsDirectory"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-sans font-bold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors cursor-pointer"
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
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">{{ t('footer_tld') }}</span>
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">{{ t('footer_ssl') }}</span>
          <button
            @click="promptQuitApp"
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded font-sans font-bold text-slate-500 hover:text-rose-600 dark:text-slate-400 dark:hover:text-rose-400 transition-colors cursor-pointer"
            :title="t('quit_app_hint')"
          >
            <Power class="w-3 h-3 text-rose-500" />
            <span>{{ t('quit_app_btn') }}</span>
          </button>
          <span class="text-slate-400 dark:text-slate-500 font-bold">{{ t('app_name') }} {{ appVersion }}</span>
        </div>
      </div>
    </footer>

    <!-- ================= 5. NEW PROJECT MODAL ================= -->
    <NewProjectModal
      v-model="showNewModal"
      :projects-directory="projectsDirectory"
      :sites="sites"
      @change-directory="changeProjectsDirectory"
      @project-created="handleProjectCreated"
    />

    <!-- ================= 6. CUSTOM CONFIRMATION MODAL ================= -->
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
            class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          >
            {{ t('cancel_btn') }}
          </button>
          <button
            @click="executeStopService"
            class="px-4 py-1.5 text-xs font-bold rounded-lg bg-rose-600 text-white hover:bg-rose-700 shadow-sm transition-colors cursor-pointer"
          >
            {{ t('confirm_btn') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ================= 7. QUIT APP CONFIRMATION MODAL ================= -->
    <QuitConfirmModal v-model="showQuitConfirm" />

    <!-- ================= 8. DOWNLOAD RUNTIMES MODAL ================= -->
    <DownloadModal v-model="showDownloadModal" />

    <!-- ================= 9. DIRECTORY SELECTION MODAL ================= -->
    <DirectoryModal
      v-model="showDirModal"
      :current-directory="projectsDirectory"
      :suggested-paths="suggestedPaths"
      @save="applyCustomDirectory"
    />

    <!-- ================= 10. PROJECT SETTINGS MODAL ================= -->
    <ProjectSettingsModal
      v-model="showSettingsModal"
      :site="selectedSiteForConfig"
      @saved="handleSettingsSaved"
    />

    <!-- ================= 11. PROJECT DIAGNOSTICS MODAL ================= -->
    <DiagnosticsModal
      v-model="showDiagModal"
      :site="selectedSiteForDiag"
      @open-settings="openSettingsModal"
    />

    <!-- ================= 12. DESIGN TOOLS MODAL ================= -->
    <DesignToolsModal
      v-model="showDesignToolsModal"
      :sites="sites"
      :initial-domain="designToolsSiteDomain"
      :t="t"
      @close="showDesignToolsModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import {
  Server, Globe, Plus, ShieldCheck, Sun, Moon,
  Lock, Info, Download, Folder, Minus, Square, X,
  AlertTriangle, FolderOpen, Power, RefreshCw,
  Database, Search, Layout
} from 'lucide-vue-next'

import type { SiteItem } from './types'
import ServiceCard from './components/ServiceCard.vue'
import ProjectRow from './components/ProjectRow.vue'
import DesignToolsModal from './components/DesignToolsModal.vue'
import NewProjectModal from './components/modals/NewProjectModal.vue'
import QuitConfirmModal from './components/modals/QuitConfirmModal.vue'
import DownloadModal from './components/modals/DownloadModal.vue'
import DirectoryModal from './components/modals/DirectoryModal.vue'
import ProjectSettingsModal from './components/modals/ProjectSettingsModal.vue'
import DiagnosticsModal from './components/modals/DiagnosticsModal.vue'

import { useI18n } from './composables/useI18n'
import { useToast } from './composables/useToast'
import { useWindow } from './composables/useWindow'
import { useRuntimeCatalog } from './composables/useRuntimeCatalog'
import { useServices } from './composables/useServices'
import { useProjects } from './composables/useProjects'

const appVersion = 'v1.1.0'

// Dark theme preference persistence
const isDark = ref(localStorage.getItem('laragal_theme') !== 'light')

const { currentLang, t, toggleLang } = useI18n()
const { notification } = useToast()
const { closeWindow, toggleMaximize, minimizeWindow } = useWindow()
const {
  loadRuntimeCatalog,
  downloadProgress,
  runtimeUpdates,
} = useRuntimeCatalog()

const {
  services,
  togglingServices,
  confirmService,
  loadServices,
  checkPortConflicts,
  promptToggleService,
  executeStopService,
  installSSL,
  openBrowser,
  openAdminer,
  openPhpMyAdmin,
} = useServices()

const {
  sites,
  activeFilter,
  searchQuery,
  runningProjects,
  isSyncingHosts,
  projectsDirectory,
  normalizedProjectsDir,
  normalizedRecentDirs,
  suggestedPaths,
  webSitesCount,
  nonWebCount,
  filterOptions,
  filteredSites,
  mapSiteItem,
  getDocRootRelative,
  isPhpProject,
  loadSites,
  loadRecentDirectories,
  switchRecentDirectory,
  applyCustomDirectory,
  updateSitePHPVersion,
  toggleProjectApp,
  syncHostsNow,
  handleProjectCreated,
} = useProjects()

// Modals visibility state
const showNewModal = ref(false)
const showDownloadModal = ref(false)
const showDesignToolsModal = ref(false)
const designToolsSiteDomain = ref('')
const showQuitConfirm = ref(false)
const showDirModal = ref(false)
const showSettingsModal = ref(false)
const selectedSiteForConfig = ref<SiteItem | null>(null)
const showDiagModal = ref(false)
const selectedSiteForDiag = ref<SiteItem | null>(null)

const copiedItem = ref('')

function copyText(txt: string) {
  navigator.clipboard.writeText(txt)
  copiedItem.value = txt
  setTimeout(() => { copiedItem.value = '' }, 2000)
}

function openExplorer(path: string) {
  const app = (window as any).go?.main?.App
  if (app?.OpenFolder) {
    app.OpenFolder(path)
  }
}

function openVSCode(path: string) {
  const app = (window as any).go?.main?.App
  if (app?.OpenVSCode) {
    app.OpenVSCode(path)
  }
}

function openNewModal() {
  showNewModal.value = true
}

function openDesignToolsModal(domain?: string) {
  designToolsSiteDomain.value = domain || ''
  showDesignToolsModal.value = true
}

function promptQuitApp() {
  showQuitConfirm.value = true
}

function changeProjectsDirectory() {
  showDirModal.value = true
}

function openSettingsModal(site: SiteItem) {
  selectedSiteForConfig.value = site
  showSettingsModal.value = true
}

function handleSettingsSaved(updatedSites?: any[]) {
  if (Array.isArray(updatedSites)) {
    sites.value = updatedSites.map(mapSiteItem)
  } else {
    loadSites()
  }
}

function openDiagnosticsModal(site: SiteItem) {
  selectedSiteForDiag.value = site
  showDiagModal.value = true
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

function toggleTheme() {
  isDark.value = !isDark.value
}

onMounted(async () => {
  const rt = (window as any).runtime
  if (rt?.EventsOn) {
    rt.EventsOn('sites-updated', (s: any) => {
      if (Array.isArray(s)) {
        sites.value = s.map(mapSiteItem)
        if (webSitesCount.value === 0 && sites.value.length > 0 && activeFilter.value === 'web') {
          activeFilter.value = 'all'
        } else if (webSitesCount.value > 0 && activeFilter.value === 'all') {
          activeFilter.value = 'web'
        }
      }
    })
    rt.EventsOn('services-updated', (live: any) => {
      if (Array.isArray(live)) {
        for (const ls of live) {
          const match = services.value.find(s => s.id === ls.name)
          if (match) match.running = Boolean(ls.running)
        }
      }
    })
    rt.EventsOn('download-progress', (data: any) => {
      if (data && data.name) {
        downloadProgress.value[data.name] = Number(data.percent) || 0
      }
    })
    rt.EventsOn('runtime-updates', (list: any) => {
      if (Array.isArray(list)) {
        const map: Record<string, any> = {}
        for (const item of list) {
          if (item && item.name) {
            map[item.name] = item
          }
        }
        runtimeUpdates.value = map
      }
    })
  }

  if ((window as any).go?.main?.App?.GetProjectsDirectory) {
    try {
      const p = await (window as any).go.main.App.GetProjectsDirectory()
      if (p) projectsDirectory.value = p.replace(/\\/g, '/')
    } catch (e) {
      console.warn('Could not load projects directory:', e)
    }
  }
  Promise.all([
    loadRuntimeCatalog(),
    loadRecentDirectories(),
    loadSites(),
    loadServices(),
    checkPortConflicts()
  ])
})
</script>
