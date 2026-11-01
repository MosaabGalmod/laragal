<template>
  <div :dir="currentLang === 'ar' ? 'rtl' : 'ltr'" class="h-screen w-screen flex flex-col overflow-hidden bg-slate-50 dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150 select-none">
    <!-- ================= UNIFIED DESKTOP HEADER (Single Bar - Zero Duplication) ================= -->
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
          class="inline-flex items-center justify-center gap-1.5 h-8 px-3 rounded-lg text-xs font-bold bg-blue-600 hover:bg-blue-700 text-white shadow-xs transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0"
          :aria-label="t('new_project_btn')"
        >
          <Plus class="w-3.5 h-3.5 shrink-0" />
          <span>{{ t('new_project_btn') }}</span>
        </button>

        <button
          @click="syncHostsNow"
          :disabled="isSyncingHosts"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 disabled:opacity-60 disabled:cursor-not-allowed shrink-0"
          :title="t('sync_hosts_hint')"
          :aria-label="t('sync_hosts_btn')"
        >
          <RefreshCw class="w-3.5 h-3.5 shrink-0 text-blue-600 dark:text-blue-400" :class="{ 'animate-spin': isSyncingHosts }" />
          <span class="hidden xl:inline">{{ t('sync_hosts_btn') }}</span>
        </button>

        <button
          @click="openAdminer()"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0"
          :title="t('open_adminer_title')"
          :aria-label="t('open_adminer')"
        >
          <Database class="w-3.5 h-3.5 shrink-0 text-amber-600 dark:text-amber-400" />
          <span class="hidden xl:inline">{{ t('open_adminer') }}</span>
        </button>

        <button
          @click="openPhpMyAdmin()"
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0"
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
          class="inline-flex items-center justify-center gap-1.5 h-8 w-8 xl:w-auto px-0 xl:px-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap text-xs font-semibold focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0"
          :title="t('ssl_install_hint')"
          :aria-label="t('install_ssl_btn')"
        >
          <ShieldCheck class="w-3.5 h-3.5 shrink-0 text-emerald-600 dark:text-emerald-400" />
          <span class="hidden xl:inline">{{ t('install_ssl_btn') }}</span>
        </button>

        <button
          @click="toggleLang"
          class="inline-flex items-center justify-center h-8 px-2.5 rounded-lg text-xs font-bold border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0"
          :title="currentLang === 'ar' ? 'English' : 'العربية'"
          :aria-label="currentLang === 'ar' ? 'English' : 'العربية'"
        >
          <span v-if="currentLang === 'ar'" lang="en">English</span>
          <span v-else lang="ar">العربية</span>
        </button>

        <button
          @click="toggleTheme"
          class="inline-flex items-center justify-center w-8 h-8 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50 shrink-0"
          :title="isDark ? t('light_mode') : t('dark_mode')"
          :aria-label="isDark ? t('light_mode') : t('dark_mode')"
        >
          <Sun v-if="isDark" class="w-3.5 h-3.5 text-amber-400" />
          <Moon v-else class="w-3.5 h-3.5 text-slate-600 dark:text-slate-300" />
        </button>

        <div class="h-5 w-px bg-slate-200 dark:bg-slate-800 mx-1 shrink-0"></div>

        <!-- Window Controls: minimize, maximize, close -->
        <div class="flex items-center gap-1 shrink-0">
          <button
            @click.stop="minimizeWindow"
            class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :title="t('minimize')"
            :aria-label="t('minimize')"
          >
            <Minus class="w-3.5 h-3.5" />
          </button>
          <button
            @click.stop="toggleMaximize"
            class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :title="t('maximize_restore')"
            :aria-label="t('maximize_restore')"
          >
            <Square class="w-3 h-3" />
          </button>
          <button
            @click.stop="closeWindow"
            class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-rose-600 hover:text-white transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
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
              class="absolute end-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500"
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
            type="button"
            :aria-pressed="activeFilter === 'web'"
            @click="activeFilter = 'web'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'web' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <span>{{ t('filter_web') }}</span>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'web' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ webSitesCount }}
            </bdi>
          </button>

          <button
            v-if="laravelSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'laravel'"
            @click="activeFilter = 'laravel'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'laravel' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>Laravel</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'laravel' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ laravelSitesCount }}
            </bdi>
          </button>

          <button
            v-if="phpSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'php'"
            @click="activeFilter = 'php'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'php' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>PHP</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'php' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ phpSitesCount }}
            </bdi>
          </button>

          <button
            v-if="goSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'go'"
            @click="activeFilter = 'go'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'go' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>Go</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'go' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ goSitesCount }}
            </bdi>
          </button>

          <button
            v-if="rustSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'rust'"
            @click="activeFilter = 'rust'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'rust' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>Rust</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'rust' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ rustSitesCount }}
            </bdi>
          </button>

          <button
            v-if="dotnetSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'dotnet'"
            @click="activeFilter = 'dotnet'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'dotnet' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>.NET</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'dotnet' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ dotnetSitesCount }}
            </bdi>
          </button>

          <button
            v-if="pythonSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'python'"
            @click="activeFilter = 'python'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'python' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>Python</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'python' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ pythonSitesCount }}
            </bdi>
          </button>

          <button
            v-if="nodeSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'node'"
            @click="activeFilter = 'node'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'node' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>Node.js</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'node' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ nodeSitesCount }}
            </bdi>
          </button>

          <button
            v-if="flutterSitesCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'flutter'"
            @click="activeFilter = 'flutter'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'flutter' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <bdi>Flutter</bdi>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'flutter' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ flutterSitesCount }}
            </bdi>
          </button>

          <button
            v-if="nonWebCount > 0"
            type="button"
            :aria-pressed="activeFilter === 'non_web'"
            @click="activeFilter = 'non_web'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'non_web' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <span>{{ t('filter_non_web') }}</span>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'non_web' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ nonWebCount }}
            </bdi>
          </button>

          <button
            type="button"
            :aria-pressed="activeFilter === 'all'"
            @click="activeFilter = 'all'"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-all shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            :class="activeFilter === 'all' ? 'bg-white dark:bg-slate-700 text-blue-700 dark:text-blue-300 shadow-xs border border-slate-300 dark:border-slate-600 font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200 bg-slate-100/80 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/60'"
          >
            <span>{{ t('filter_all') }}</span>
            <bdi class="px-1.5 py-0.2 rounded-full text-[10px] font-mono leading-tight" :class="activeFilter === 'all' ? 'bg-blue-100 dark:bg-blue-900/60 text-blue-800 dark:text-blue-200' : 'bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300'">
              {{ sites.length }}
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
              class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
            >
              <X class="w-3.5 h-3.5" />
              <span>{{ t('clear_search_btn') }}</span>
            </button>
            <template v-else>
              <button
                v-if="activeFilter === 'web' && sites.length > 0"
                type="button"
                @click="activeFilter = 'all'"
                class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
              >
                <Folder class="w-3.5 h-3.5" />
                <span>{{ t('show_all_folders') }} (<bdi>{{ sites.length }}</bdi>)</span>
              </button>
              <button
                type="button"
                @click="openNewModal"
                class="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold bg-blue-600 text-white hover:bg-blue-700 shadow-xs transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
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
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">{{ t('footer_tld') }}</span>
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">{{ t('footer_ssl') }}</span>
          <button
            @click="promptQuitApp"
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded font-sans font-bold text-slate-500 hover:text-rose-600 dark:text-slate-400 dark:hover:text-rose-400 transition-colors"
            :title="t('quit_app_hint')"
          >
            <Power class="w-3 h-3 text-rose-500" />
            <span>{{ t('quit_app_btn') }}</span>
          </button>
          <span class="text-slate-400 dark:text-slate-500 font-bold">{{ t('app_name') }} {{ appVersion }}</span>
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
              :placeholder="t('project_name_placeholder')"
              class="w-full px-3.5 py-2.5 rounded-lg border text-sm text-slate-900 dark:text-white transition-colors focus:outline-none"
              :class="projectError ? 'border-rose-500 bg-rose-50/40 dark:bg-rose-950/20 focus:ring-2 focus:ring-rose-500' : 'border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-blue-500'"
            />
            <!-- Custom accessible validation message -->
            <p v-if="projectError" id="projectNameError" role="alert" aria-live="polite" class="text-xs text-rose-500 font-semibold mt-1.5 flex items-center gap-1">
              <AlertCircle class="w-3.5 h-3.5 flex-shrink-0" />
              <span>{{ projectError }}</span>
            </p>
            <p v-else class="text-[11px] text-slate-400 mt-1 font-mono">
              <span>{{ t('auto_url_hint') }}</span> <bdi dir="ltr">https://{{ newProject.name ? newProject.name.toLowerCase() : 'name' }}.test</bdi>
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
              <option
                v-for="opt in phpSelectOptions"
                :key="opt.version"
                :value="opt.version"
              >
                {{ opt.label }}
              </option>
            </select>
            <div v-if="selectedNewProjectPhpMeta?.eol" class="mt-1.5 flex items-center gap-1.5">
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
            <div v-else-if="selectedNewProjectPhpMeta?.eol_date" class="mt-1.5 flex items-center gap-1.5">
              <span
                class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-700 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-200 dark:border-amber-900"
              >
                {{ t('runtime_security_only', { date: selectedNewProjectPhpMeta.eol_date }) }}
              </span>
            </div>
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
          <div class="flex items-center gap-2">
            <Download class="w-4 h-4 text-blue-600" />
            <h3 class="font-bold text-base text-slate-900 dark:text-white">{{ t('modal_download_title') }}</h3>
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="checkUpdatesNow"
              :disabled="isCheckingUpdates"
              class="px-2.5 py-1 text-xs font-semibold rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 transition-colors flex items-center gap-1.5 cursor-pointer"
              :title="t('check_updates_title')"
            >
              <RefreshCw class="w-3 h-3 text-blue-500" :class="{ 'animate-spin': isCheckingUpdates }" />
              <span>{{ isCheckingUpdates ? t('checking_updates') : t('check_updates_btn') }}</span>
            </button>
            <button @click="showDownloadModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 font-bold text-lg leading-none cursor-pointer">&times;</button>
          </div>
        </div>

        <div class="space-y-3 max-h-[60vh] overflow-y-auto px-0.5">
          <div
            v-for="rt in availableRuntimes"
            :key="rt.name"
            class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-850 space-y-2 transition-all"
            :class="{ 'border-rose-300 dark:border-rose-900/60 bg-rose-50/20 dark:bg-rose-950/20': (runtimeUpdates[rt.name]?.has_update && runtimeUpdates[rt.name]?.is_security_fix) || getCatalogItem(rt.name)?.eol }"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="space-y-0.5">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <h4 class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ rt.label }}</h4>
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
                    <ShieldAlert class="w-2.5 h-2.5" /> {{ t('runtime_security_upgrade') }} <bdi dir="ltr">v{{ runtimeUpdates[rt.name].latest_version }}</bdi>
                  </span>
                  <span
                    v-else-if="runtimeUpdates[rt.name]?.has_update"
                    class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-700 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-200 dark:border-amber-900"
                  >
                    <ArrowUpCircle class="w-2.5 h-2.5" /> {{ t('runtime_update_available') }} <bdi dir="ltr">v{{ runtimeUpdates[rt.name].latest_version }}</bdi>
                  </span>
                </div>
                <p class="text-[11px] text-slate-400 font-mono">{{ t(rt.desc_key) }}</p>
              </div>
              <button
                @click="triggerDownloadRuntime(rt.name)"
                :disabled="downloadingRuntime === rt.name"
                class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors flex items-center gap-1.5 cursor-pointer flex-shrink-0"
                :class="runtimeUpdates[rt.name]?.has_update ? 'bg-emerald-600 hover:bg-emerald-700 text-white shadow-xs' : 'bg-blue-600 hover:bg-blue-700 text-white'"
              >
                <Download class="w-3.5 h-3.5" :class="downloadingRuntime === rt.name ? 'animate-bounce' : ''" />
                <span>
                  {{ downloadingRuntime === rt.name ? `${downloadProgress[rt.name] || 0}%` : (runtimeUpdates[rt.name]?.has_update ? t('upgrade_one_click') : t('download_now')) }}
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
                :placeholder="t('dir_path_placeholder')"
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
              <option value="">{{ t('project_type_auto') }}</option>
              <option value="laravel">{{ t('project_type_laravel') }}</option>
              <option value="wordpress">{{ t('project_type_wordpress') }}</option>
              <option value="php">{{ t('project_type_php') }}</option>
              <option value="node">{{ t('project_type_node') }}</option>
              <option value="go">{{ t('project_type_go') }}</option>
              <option value="rust">{{ t('project_type_rust') }}</option>
              <option value="dotnet">{{ t('project_type_dotnet') }}</option>
              <option value="python">{{ t('project_type_python') }}</option>
              <option value="flutter">{{ t('project_type_flutter') }}</option>
              <option value="static">{{ t('project_type_static') }}</option>
              <option value="non_web">{{ t('project_type_non_web') }}</option>
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
              :placeholder="t('setting_proxy_port_placeholder')"
            />
          </div>

          <!-- PHP Version Selector (if PHP type) -->
          <div v-if="settingsForm.type === 'laravel' || settingsForm.type === 'wordpress' || settingsForm.type === 'php' || (!settingsForm.type && isPhpProject(selectedSiteForConfig))">
            <label class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1">
              {{ t('custom_php_version') }}
            </label>
            <select
              v-model="settingsForm.php_version"
              dir="ltr"
              class="w-full px-3 py-2 rounded-xl text-xs border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">{{ t('php_default_global') }}</option>
              <option
                v-for="opt in phpSelectOptionsPrefixed"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </option>
            </select>
            <div v-if="selectedSettingsPhpMeta?.eol" class="mt-1.5 flex items-center gap-1.5">
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
            <div v-else-if="selectedSettingsPhpMeta?.eol_date" class="mt-1.5 flex items-center gap-1.5">
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
                type="checkbox"
                v-model="settingsForm.is_ignored"
                class="w-4 h-4 rounded text-blue-600 focus:ring-blue-500"
              />
              <span class="text-xs text-slate-600 dark:text-slate-400">{{ t('hide_folder_from_list') }}</span>
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
            :title="t('recheck_btn')"
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
              <span dir="ltr" class="font-mono text-[11px]" :class="diagResult.has_index_file ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'">
                {{ diagResult.index_file || t('missing_label') }}
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
                <span v-if="diagResult.has_env_file" class="text-emerald-600 dark:text-emerald-400 font-bold">{{ t('present_label') }}</span>
                <button
                  v-else
                  type="button"
                  @click="triggerFixEnv"
                  :disabled="isFixingEnv"
                  class="px-2 py-1 rounded bg-amber-100 hover:bg-amber-200 text-amber-800 dark:bg-amber-950 dark:hover:bg-amber-900 dark:text-amber-200 font-bold text-[11px] transition-colors cursor-pointer"
                >
                  <span v-if="isFixingEnv">{{ t('creating_env') }}</span>
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
                {{ (diagResult.has_vendor_dir ? t('vendor_installed') : (diagResult.has_node_modules ? t('node_modules_installed') : t('not_installed_deps'))) }}
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
                ({{ diagResult.is_port_active ? t('port_connected') : t('port_disconnected') }})
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
                {{ discoveredDatabases.length }} {{ t('discovered_badge') }}
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
                      {{ dbItem.type === 'schema_file' ? t('db_schema_file') : t('db_legacy_dir') }}
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
              <li v-for="(issue, idx) in diagResult.issues" :key="idx"><bdi>{{ renderUIMessage(issue) }}</bdi></li>
            </ul>
          </div>

          <div v-if="diagResult.suggestions && diagResult.suggestions.length > 0" class="p-3.5 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/50 space-y-1.5">
            <span class="text-xs font-bold text-blue-700 dark:text-blue-300 block">{{ t('suggestions_title') }}</span>
            <ul class="text-xs text-blue-600 dark:text-blue-400 space-y-1 list-disc list-inside">
              <li v-for="(sug, idx) in diagResult.suggestions" :key="idx"><bdi>{{ renderUIMessage(sug) }}</bdi></li>
            </ul>
          </div>
        </div>

        <div class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
          <button
            type="button"
            @click="showDiagModal = false; openSettingsModal(selectedSiteForDiag)"
            class="px-3.5 py-2 text-xs font-semibold rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors flex items-center gap-1.5 cursor-pointer"
          >
            <Settings class="w-3.5 h-3.5" />
            <span>{{ t('edit_project_settings_btn') }}</span>
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

    <!-- ================= 10. DESIGN TOOLS MODAL (Responsive, Contrast, Resources) ================= -->
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
import { ref, computed, watch, onMounted } from 'vue'
import {
  Server, Globe, Plus, ShieldCheck, Sun, Moon,
  Lock, CheckCircle2, Info, Download,
  Folder, Check, Minus, Square, X,
  AlertCircle, AlertTriangle, FolderOpen, Power, RefreshCw,
  Settings, Wrench, Database, ShieldAlert, ArrowUpCircle,
  Search, Layout
} from 'lucide-vue-next'

import ServiceCard from './components/ServiceCard.vue'
import ProjectRow from './components/ProjectRow.vue'
import DesignToolsModal from './components/DesignToolsModal.vue'
import { getProjectTypeLabel } from './utils/projectType'

import arLocale from './locales/ar.json'
import enLocale from './locales/en.json'

const appVersion = 'v1.1.0'

// Dark theme preference persistence
const isDark = ref(localStorage.getItem('laragal_theme') !== 'light')
const currentLang = ref<'ar' | 'en'>('ar')
const showNewModal = ref(false)
const showDownloadModal = ref(false)
const showDesignToolsModal = ref(false)
const designToolsSiteDomain = ref('')
const downloadingRuntime = ref('')
const notification = ref('')
const copiedItem = ref('')
const projectError = ref('')
const confirmService = ref<ServiceItem | null>(null)
const togglingServices = ref<Set<string>>(new Set())
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
    notification.value = t('notif_app_stopped', { name: pName })
  } else {
    notification.value = t('notif_app_starting', { name: pName })
    if ((window as any).go?.main?.App?.StartProjectApp) {
      try {
        const ok = await (window as any).go.main.App.StartProjectApp(site.path || site.root_path, site.proxy_port || 8080)
        if (ok) {
          runningProjects.value[site.domain] = true
          notification.value = t('notif_app_started', { name: pName, port: site.proxy_port || 8080 })
          return
        }
      } catch (err: any) {
        notification.value = t('notif_app_start_failed', { name: pName, error: err?.message || err })
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
  if (sites.value && sites.value.length > 0) {
    sites.value = sites.value.map(s => ({
      ...s,
      type_label: getProjectTypeLabel(s.project_type, s.type_label, t),
    }))
  }
}, { immediate: true })

function toggleTheme() {
  isDark.value = !isDark.value
}

// Window control functions connected to Wails backend
function closeWindow() {
  const w = window as any
  if (w.go?.main?.App?.CloseWindow) {
    w.go.main.App.CloseWindow()
  } else if (w.runtime?.WindowHide) {
    w.runtime.WindowHide()
  } else if (w.runtime?.Quit) {
    w.runtime.Quit()
  }
}

function toggleMaximize() {
  const w = window as any
  if (w.go?.main?.App?.ToggleMaximizeWindow) {
    w.go.main.App.ToggleMaximizeWindow()
  } else if (w.runtime?.WindowToggleMaximise) {
    w.runtime.WindowToggleMaximise()
  }
}

function minimizeWindow() {
  const w = window as any
  if (w.go?.main?.App?.MinimizeWindow) {
    w.go.main.App.MinimizeWindow()
  } else if (w.runtime?.WindowMinimise) {
    w.runtime.WindowMinimise()
  }
}

function promptQuitApp() {
  showQuitConfirm.value = true
}

function executeQuitApp() {
  showQuitConfirm.value = false
  const w = window as any
  if (w.go?.main?.App?.QuitApp) {
    w.go.main.App.QuitApp()
  } else if (w.runtime?.Quit) {
    w.runtime.Quit()
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
  const typeLabel = getProjectTypeLabel(projectType, item.type_label || item.TypeLabel, t)
  const rawPhp = item.php_version || item.PHPVersion || ''
  const isPhpType = projectType === 'laravel' || projectType === 'wordpress' || projectType === 'php'
  const phpVersion = rawPhp ? (rawPhp.startsWith('PHP') ? rawPhp : `PHP ${rawPhp}`) : (isWeb && isPhpType ? 'PHP 8.3' : '')

  return {
    name: item.name || item.Name || item.domain || item.Domain || '',
    domain: item.domain || item.Domain || '',
    path: item.path || item.Path || '',
    root_path: item.root_path || item.RootPath || '',
    php_version: phpVersion,
    php_port: item.php_port !== undefined ? item.php_port : (item.PHPPort !== undefined ? item.PHPPort : (isWeb && isPhpType ? getPHPPort(item.php_version) : 0)),
    proxy_port: item.proxy_port || item.ProxyPort || (projectType === 'go' ? 8080 : 0),
    project_type: projectType,
    type_label: typeLabel,
    is_web: isWeb,
  }
}

const sites = ref<SiteItem[]>([])
const activeFilter = ref<'web' | 'all' | 'laravel' | 'php' | 'go' | 'rust' | 'dotnet' | 'python' | 'node' | 'flutter' | 'non_web'>('web')
const searchQuery = ref('')
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
  let list = sites.value

  if (activeFilter.value === 'web') {
    list = list.filter(s => s.is_web)
  } else if (activeFilter.value === 'laravel') {
    list = list.filter(s => s.project_type === 'laravel')
  } else if (activeFilter.value === 'php') {
    list = list.filter(s => s.project_type === 'php')
  } else if (activeFilter.value === 'go') {
    list = list.filter(s => s.project_type === 'go' || s.project_type === 'proxy')
  } else if (activeFilter.value === 'rust') {
    list = list.filter(s => s.project_type === 'rust')
  } else if (activeFilter.value === 'dotnet') {
    list = list.filter(s => s.project_type === 'dotnet')
  } else if (activeFilter.value === 'python') {
    list = list.filter(s => s.project_type === 'python')
  } else if (activeFilter.value === 'node') {
    list = list.filter(s => s.project_type === 'node')
  } else if (activeFilter.value === 'flutter') {
    list = list.filter(s => s.project_type === 'flutter')
  } else if (activeFilter.value === 'non_web') {
    list = list.filter(s => !s.is_web)
  }

  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter(s => {
      const name = (s.name || '').toLowerCase()
      const domain = (s.domain || '').toLowerCase()
      const path = (s.path || s.root_path || '').toLowerCase()
      return name.includes(q) || domain.includes(q) || path.includes(q)
    })
  }

  return list
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
  if (clean.startsWith('8.5')) return '8.5'
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
      site.php_port = getPHPPort(newVer)
      notification.value = `${t('php_version_changed')} ${site.domain} ➔ PHP ${newVer}`
      if (Array.isArray(updated)) {
        sites.value = updated.map(mapSiteItem)
      }
    } catch (e: any) {
      notification.value = `${t('error_changing_php')}: ${e?.message || e}`
    }
  } else {
    site.php_version = `PHP ${newVer}`
    site.php_port = getPHPPort(newVer)
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
      notification.value = t('error_saving_settings', { error: e?.message || e })
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
      php_port: site.php_port || getPHPPort(site.php_version),
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
      notification.value = t('error_fixing_env', { error: e?.message || e })
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
    notification.value = t('error_importing_db', { error: err?.message || err })
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
      let s = await (window as any).go.main.App.GetSites()
      if (Array.isArray(s) && s.length === 0 && (window as any).go?.main?.App?.ScanProjects) {
        s = await (window as any).go.main.App.ScanProjects()
      }
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
        const map: Record<string, RuntimeUpdateInfo> = {}
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
      console.warn("Could not load projects directory:", e)
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

interface RuntimeUpdateInfo {
  name: string
  current_version: string
  latest_version: string
  has_update: boolean
  is_security_fix: boolean
  download_url: string
}

const runtimeUpdates = ref<Record<string, RuntimeUpdateInfo>>({})
const isCheckingUpdates = ref(false)
const downloadProgress = ref<Record<string, number>>({})

const availableRuntimes = ref([
  { name: 'caddy', label: 'Caddy Server v2.8', desc_key: 'service_caddy_desc' },
  { name: 'php-8.5', label: 'PHP 8.5 (NTS x64)', desc_key: 'service_php85_desc' },
  { name: 'php-8.4', label: 'PHP 8.4 (NTS x64)', desc_key: 'service_php84_desc' },
  { name: 'php-8.3', label: 'PHP 8.3 (NTS x64)', desc_key: 'service_php83_desc' },
  { name: 'php-8.2', label: 'PHP 8.2 (NTS x64)', desc_key: 'service_php82_desc' },
  { name: 'php-8.1', label: 'PHP 8.1 (NTS x64)', desc_key: 'service_php81_desc' },
  { name: 'mysql', label: 'MariaDB / MySQL Server', desc_key: 'service_mysql_desc' },
  { name: 'redis', label: 'Redis Server (Windows)', desc_key: 'service_redis_desc' },
  { name: 'mailpit', label: 'Mailpit Server v1.21', desc_key: 'service_mailpit_desc' },
])

const locales: Record<string, Record<string, string>> = {
  ar: arLocale,
  en: enLocale
}

function renderUIMessage(msg: any): string {
  if (!msg) return ''
  if (typeof msg === 'string') return msg
  if (msg.key) {
    return t(msg.key, msg.params)
  }
  return ''
}

function t(key: string, params?: Record<string, string | number | undefined>): string {
  let str = locales[currentLang.value]?.[key] || key
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      str = str.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v ?? ''))
    }
  }
  return str
}

function toggleLang() {
  currentLang.value = currentLang.value === 'ar' ? 'en' : 'ar'
}

interface RuntimeArtifact {
  name: string
  version?: string
  url?: string
  sha256?: string
  eol?: boolean
  eol_date?: string
  php_port?: number
}

const runtimeCatalog = ref<RuntimeArtifact[]>([])

async function loadRuntimeCatalog() {
  const w = window as any
  if (w.go?.main?.App?.GetRuntimeCatalog) {
    try {
      const list = await w.go.main.App.GetRuntimeCatalog()
      if (Array.isArray(list) && list.length > 0) {
        runtimeCatalog.value = list
      }
    } catch (e) {
      console.warn("Could not load runtime catalog:", e)
    }
  }
}

function getCatalogItem(name: string): RuntimeArtifact | undefined {
  return runtimeCatalog.value.find(r => r.name === name || r.name === `php-${name}` || (name.startsWith('php-') && r.name === name))
}

function compareVersions(v1: string, v2: string): number {
  const p1 = v1.split('.').map(n => parseInt(n, 10) || 0)
  const p2 = v2.split('.').map(n => parseInt(n, 10) || 0)
  for (let i = 0; i < Math.max(p1.length, p2.length); i++) {
    const a = p1[i] || 0
    const b = p2[i] || 0
    if (a > b) return 1
    if (a < b) return -1
  }
  return 0
}

const phpCatalog = computed(() => {
  return runtimeCatalog.value.filter(r => r.name.startsWith('php-') && (r.php_port || 0) > 0)
})

const latestPhpVersion = computed(() => {
  let highest = ''
  for (const item of phpCatalog.value) {
    const ver = item.name.replace(/^php-/, '')
    if (!highest || compareVersions(ver, highest) > 0) {
      highest = ver
    }
  }
  return highest || '8.5'
})

function getPHPPort(version?: string): number {
  if (!version) return 9083
  const clean = extractPhpVersionNumber(version)
  const exact = runtimeCatalog.value.find(r => r.name === `php-${clean}` || r.name === clean)
  if (exact && exact.php_port) return exact.php_port
  const parts = clean.split('.')
  if (parts.length === 2) {
    const maj = parseInt(parts[0], 10)
    const min = parseInt(parts[1], 10)
    if (!isNaN(maj) && !isNaN(min)) {
      return 9000 + (maj * 10) + min
    }
  }
  return 9083
}

const phpSelectOptions = computed(() => {
  return phpCatalog.value.map(item => {
    const ver = item.name.replace(/^php-/, '')
    let suffix = ''
    if (ver === latestPhpVersion.value) {
      suffix = ` (${t('php_latest_label')})`
    } else if (ver === '8.3') {
      suffix = ` (${t('php_recommended_badge')})`
    } else if (item.eol) {
      suffix = ` (${t('runtime_eol_badge')})`
    } else if (item.eol_date) {
      suffix = ` (${t('runtime_security_only', { date: item.eol_date })})`
    }
    return {
      version: ver,
      label: `PHP ${ver}${suffix}`,
      eol: Boolean(item.eol),
      eol_date: item.eol_date,
      php_port: item.php_port,
    }
  })
})

const phpSelectOptionsPrefixed = computed(() => {
  return phpCatalog.value.map(item => {
    const ver = item.name.replace(/^php-/, '')
    let suffix = ''
    if (ver === latestPhpVersion.value) {
      suffix = ` (${t('php_latest_label')})`
    } else if (ver === '8.3') {
      suffix = ` (${t('php_recommended_badge')})`
    } else if (item.eol) {
      suffix = ` (${t('runtime_eol_badge')})`
    } else if (item.eol_date) {
      suffix = ` (${t('runtime_security_only', { date: item.eol_date })})`
    }
    return {
      value: `PHP ${ver}`,
      version: ver,
      label: `PHP ${ver}${suffix}`,
      eol: Boolean(item.eol),
      eol_date: item.eol_date,
      php_port: item.php_port,
    }
  })
})

const selectedNewProjectPhpMeta = computed(() => {
  const ver = newProject.value.php
  return phpCatalog.value.find(item => item.name === `php-${ver}` || item.name === ver)
})

const selectedSettingsPhpMeta = computed(() => {
  const ver = extractPhpVersionNumber(settingsForm.value.php_version)
  return phpCatalog.value.find(item => item.name === `php-${ver}` || item.name === ver)
})

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
    id: 'php-8.5',
    name_key: 'service_php85',
    desc_key: 'service_php85_desc',
    port: 9085,
    running: false,
    credentials: {
      host: '127.0.0.1',
      info: 'FastCGI / 8 Workers'
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
      notification.value = renderUIMessage(msg) || t('hosts_synced_success')
    } else {
      notification.value = t('hosts_synced_success')
    }
  } catch (err: any) {
    notification.value = t('error_syncing_hosts', { error: err?.message || err })
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
  const app = (window as any).go?.main?.App
  if (app?.OpenFolder) {
    app.OpenFolder(path)
  }
}

function openAdminer(dbName?: string) {
  const app = (window as any).go?.main?.App
  if (app?.OpenAdminer) {
    app.OpenAdminer(dbName || '').catch((err: any) => {
      notification.value = t('operation_failed', { error: err?.message || String(err) })
    })
  } else {
    const dbParam = dbName ? `&db=${encodeURIComponent(dbName)}` : ''
    openBrowser(`http://localhost:3307/?server=127.0.0.1&username=root${dbParam}`)
  }
}

function openPhpMyAdmin(dbName?: string) {
  const app = (window as any).go?.main?.App
  if (app?.OpenPhpMyAdmin) {
    app.OpenPhpMyAdmin(dbName || '').catch((err: any) => {
      notification.value = t('operation_failed', { error: err?.message || String(err) })
    })
  } else {
    const dbParam = dbName ? `index.php?route=/database/structure&db=${encodeURIComponent(dbName)}` : ''
    openBrowser(`http://localhost:3308/${dbParam}`)
  }
}

function openBrowser(url: string) {
  const app = (window as any).go?.main?.App
  if (app?.OpenBrowser) {
    app.OpenBrowser(url)
  } else {
    window.open(url, '_blank')
  }
}

function openVSCode(path: string) {
  const app = (window as any).go?.main?.App
  if (app?.OpenVSCode) {
    app.OpenVSCode(path)
  }
}

function openNewModal() {
  projectError.value = ''
  newProject.value.name = ''
  showNewModal.value = true
}

function openDesignToolsModal(domain?: string) {
  designToolsSiteDomain.value = domain || ''
  showDesignToolsModal.value = true
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
      const rendered = renderUIMessage(msg)
      if (rendered) {
        notification.value = rendered
      }
    } catch (e) {
      console.warn("CheckPortConflict err:", e)
    }
  }
}

// Custom service toggle: immediate toggle without blocking modal
async function promptToggleService(service: ServiceItem) {
  if (togglingServices.value.has(service.id)) return
  togglingServices.value.add(service.id)
  togglingServices.value = new Set(togglingServices.value)

  try {
    if ((window as any).go?.main?.App?.ToggleService) {
      const originalState = service.running
      try {
        const running = await (window as any).go.main.App.ToggleService(service.id)
        service.running = Boolean(running)
        if (running) {
          notification.value = t('service_started_msg', { service: t(service.name_key) })
        } else {
          notification.value = t('service_stopped_msg', { service: t(service.name_key) })
        }
        await loadServices()
      } catch (err: any) {
        service.running = originalState
        notification.value = t('service_toggle_error', { service: t(service.name_key), error: err?.message || err })
        await loadServices()
      }
    } else {
      service.running = !service.running
      if (service.running) {
        notification.value = t('service_started_msg', { service: t(service.name_key) })
      } else {
        notification.value = t('service_stopped_msg', { service: t(service.name_key) })
      }
    }
  } finally {
    togglingServices.value.delete(service.id)
    togglingServices.value = new Set(togglingServices.value)
  }
}

async function executeStopService() {
  if (!confirmService.value) return
  const s = confirmService.value
  if ((window as any).go?.main?.App?.ToggleService) {
    try {
      const running = await (window as any).go.main.App.ToggleService(s.id)
      s.running = Boolean(running)
      if (!running) {
        notification.value = t('service_stopped_msg', { service: t(s.name_key) })
      } else {
        notification.value = t('service_stop_failed', { service: t(s.name_key) })
      }
      await loadServices()
    } catch (err: any) {
      notification.value = t('service_stop_failed_error', { service: t(s.name_key), error: err?.message || err })
      await loadServices()
    }
  } else {
    s.running = false
    notification.value = t('service_stopped_msg', { service: t(s.name_key) })
  }
  confirmService.value = null
}

async function installSSL() {
  if ((window as any).go?.main?.App?.InstallSSLTrust) {
    try {
      const res = await (window as any).go.main.App.InstallSSLTrust()
      notification.value = renderUIMessage(res) || t('success_ssl')
      return
    } catch (err: any) {
      notification.value = t('error_installing_ssl', { error: err })
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
  const port = getPHPPort(newProject.value.php)
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

  notification.value = t('project_created_success', { name: projectName, domain: domain })
  showNewModal.value = false
  newProject.value.name = ''
  projectError.value = ''
}

async function checkUpdatesNow() {
  isCheckingUpdates.value = true
  if ((window as any).go?.main?.App?.CheckRuntimeUpdates) {
    try {
      const list = await (window as any).go.main.App.CheckRuntimeUpdates()
      if (Array.isArray(list)) {
        const map: Record<string, RuntimeUpdateInfo> = {}
        let count = 0
        for (const item of list) {
          if (item && item.name) {
            map[item.name] = item
            if (item.has_update) count++
          }
        }
        runtimeUpdates.value = map
        if (count > 0) {
          notification.value = t('updates_found_count', { count })
        } else {
          notification.value = t('all_runtimes_up_to_date')
        }
      }
    } catch (err: any) {
      console.warn("Check updates error:", err)
    }
  }
  isCheckingUpdates.value = false
}

async function triggerDownloadRuntime(name: string) {
  downloadingRuntime.value = name
  downloadProgress.value[name] = 0
  notification.value = t('runtime_download_starting', { name })

  if ((window as any).go?.main?.App?.UpgradeRuntime) {
    try {
      const msg = await (window as any).go.main.App.UpgradeRuntime(name)
      downloadingRuntime.value = ''
      downloadProgress.value[name] = 100
      if (runtimeUpdates.value[name]) {
        runtimeUpdates.value[name].has_update = false
        runtimeUpdates.value[name].current_version = runtimeUpdates.value[name].latest_version
      }
      notification.value = renderUIMessage(msg) || t('runtime_upgrade_success', { name })
      return
    } catch (err: any) {
      console.warn("Upgrade err:", err)
      downloadingRuntime.value = ''
      notification.value = t('runtime_upgrade_failed', { name, error: err?.message || err })
      return
    }
  }

  if ((window as any).go?.main?.App?.DownloadBinary) {
    try {
      const msg = await (window as any).go.main.App.DownloadBinary(name)
      downloadingRuntime.value = ''
      downloadProgress.value[name] = 100
      notification.value = renderUIMessage(msg) || t('runtime_install_success', { name })
      return
    } catch (err: any) {
      console.warn("Download err:", err)
      downloadingRuntime.value = ''
      notification.value = t('runtime_download_failed', { name, error: err?.message || err })
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
    notification.value = t('runtime_ready_in_dir', { name })
  }, 1500)
}
</script>
