<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/75 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="design-tools-modal-title"
    tabindex="-1"
    @click.self="closeModal"
  >
    <div
      class="w-full max-w-4xl max-h-[92vh] rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl flex flex-col overflow-hidden animate-scale-in"
    >
      <!-- Header -->
      <div
        class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between flex-shrink-0"
      >
        <div class="flex items-center gap-2.5">
          <div
            class="w-8 h-8 rounded-lg bg-indigo-50 dark:bg-indigo-950/80 border border-indigo-200 dark:border-indigo-800 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0"
          >
            <Layout class="w-4 h-4" />
          </div>
          <div>
            <h3
              id="design-tools-modal-title"
              class="font-bold text-base text-slate-900 dark:text-white"
            >
              {{ t('design_tools_title') }}
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              {{ t('design_tools_desc') }}
            </p>
          </div>
        </div>

        <button
          type="button"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1.5 rounded-lg transition-colors cursor-pointer"
          :title="t('close_btn')"
          :aria-label="t('close_btn')"
          @click="closeModal"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Tabs Navigation -->
      <div
        class="px-6 pt-3 border-b border-slate-100 dark:border-slate-800 flex items-center gap-2 flex-shrink-0 bg-slate-50/50 dark:bg-slate-900/50"
      >
        <button
          type="button"
          class="px-4 py-2 rounded-t-xl text-xs font-bold border-b-2 transition-all cursor-pointer flex items-center gap-2"
          :class="
            activeTab === 'responsive'
              ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
          "
          @click="activeTab = 'responsive'"
        >
          <Smartphone class="w-3.5 h-3.5" />
          <span>{{ t('tab_responsive') }}</span>
        </button>

        <button
          type="button"
          class="px-4 py-2 rounded-t-xl text-xs font-bold border-b-2 transition-all cursor-pointer flex items-center gap-2"
          :class="
            activeTab === 'contrast'
              ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
          "
          @click="activeTab = 'contrast'"
        >
          <Palette class="w-3.5 h-3.5" />
          <span>{{ t('tab_contrast') }}</span>
        </button>

        <button
          type="button"
          class="px-4 py-2 rounded-t-xl text-xs font-bold border-b-2 transition-all cursor-pointer flex items-center gap-2"
          :class="
            activeTab === 'cloud'
              ? 'border-indigo-600 text-indigo-600 dark:text-indigo-400 bg-white dark:bg-slate-900'
              : 'border-transparent text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
          "
          @click="activeTab = 'cloud'"
        >
          <ExternalLink class="w-3.5 h-3.5" />
          <span>{{ t('tab_cloud_tools') }}</span>
        </button>
      </div>

      <!-- Body Container -->
      <div class="p-6 overflow-y-auto flex-1 space-y-4">
        <!-- TAB 1: RESPONSIVE VIEWPORT PREVIEWER -->
        <div
          v-if="activeTab === 'responsive'"
          class="space-y-4"
        >
          <div
            class="flex flex-wrap items-center justify-between gap-3 p-3 rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700"
          >
            <div class="flex items-center gap-2">
              <label
                for="design-site-select"
                class="text-xs font-bold text-slate-600 dark:text-slate-300"
              >
                {{ t('select_site_to_preview') }}:
              </label>
              <select
                id="design-site-select"
                v-model="selectedDomain"
                class="px-3 py-1.5 rounded-lg text-xs font-mono font-bold bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 text-slate-900 dark:text-white"
              >
                <option
                  v-for="s in webSitesList"
                  :key="s.domain"
                  :value="s.domain"
                >
                  {{ s.name || s.domain }} (https://{{ s.domain }})
                </option>
              </select>
            </div>

            <div class="flex items-center gap-1.5">
              <!-- Mobile (375px) -->
              <button
                type="button"
                class="px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors cursor-pointer"
                :class="
                  viewportMode === 'mobile'
                    ? 'bg-indigo-600 text-white shadow-xs'
                    : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700'
                "
                :title="t('preview_mobile')"
                @click="viewportMode = 'mobile'"
              >
                <Smartphone class="w-3.5 h-3.5" />
                <span>{{ t('preview_mobile') }}</span>
              </button>

              <!-- Tablet (768px) -->
              <button
                type="button"
                class="px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors cursor-pointer"
                :class="
                  viewportMode === 'tablet'
                    ? 'bg-indigo-600 text-white shadow-xs'
                    : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700'
                "
                :title="t('preview_tablet')"
                @click="viewportMode = 'tablet'"
              >
                <Tablet class="w-3.5 h-3.5" />
                <span>{{ t('preview_tablet') }}</span>
              </button>

              <!-- Desktop (100%) -->
              <button
                type="button"
                class="px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors cursor-pointer"
                :class="
                  viewportMode === 'desktop'
                    ? 'bg-indigo-600 text-white shadow-xs'
                    : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700'
                "
                :title="t('preview_desktop')"
                @click="viewportMode = 'desktop'"
              >
                <Monitor class="w-3.5 h-3.5" />
                <span>{{ t('preview_desktop') }}</span>
              </button>

              <!-- Rotate -->
              <button
                type="button"
                class="p-1.5 rounded-lg border transition-colors cursor-pointer"
                :class="
                  isRotated
                    ? 'bg-indigo-100 text-indigo-700 border-indigo-300 dark:bg-indigo-950 dark:border-indigo-700 dark:text-indigo-300'
                    : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700'
                "
                :title="t('preview_rotate')"
                :aria-label="t('preview_rotate')"
                @click="isRotated = !isRotated"
              >
                <RotateCw class="w-3.5 h-3.5" />
              </button>

              <!-- Refresh -->
              <button
                type="button"
                class="p-1.5 rounded-lg border bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
                :title="t('refresh_preview')"
                :aria-label="t('refresh_preview')"
                @click="reloadPreviewFrame"
              >
                <RefreshCw
                  class="w-3.5 h-3.5"
                  :class="{ 'animate-spin': isRefreshing }"
                />
              </button>
            </div>
          </div>

          <!-- Preview Frame Wrapper -->
          <div
            class="w-full min-h-[420px] bg-slate-200/60 dark:bg-slate-950/60 rounded-2xl border border-slate-200 dark:border-slate-800 p-4 flex flex-col items-center justify-center overflow-hidden"
          >
            <div
              v-if="selectedDomain"
              class="w-full flex flex-col items-center gap-2"
            >
              <div
                class="text-[11px] font-mono font-bold text-slate-500 dark:text-slate-400 bg-white/80 dark:bg-slate-900/80 px-2.5 py-0.5 rounded-md border border-slate-200 dark:border-slate-800 flex items-center gap-1.5"
              >
                <span>{{ t('preview_dimensions') }}:</span>
                <bdi dir="ltr">{{ currentWidth }} &times; {{ currentHeight }}px</bdi>
              </div>

              <div
                class="transition-all duration-300 bg-white rounded-xl shadow-lg border border-slate-300 dark:border-slate-700 overflow-hidden flex justify-center max-w-full"
                :style="previewFrameContainerStyle"
              >
                <iframe
                  :key="frameKey"
                  :src="'https://' + selectedDomain"
                  :style="previewIframeStyle"
                  class="border-0 bg-white"
                  loading="lazy"
                  title="LaraGal Web Responsive Preview"
                  sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
                ></iframe>
              </div>
            </div>

            <div
              v-else
              class="text-center py-12 space-y-2"
            >
              <Layout class="w-8 h-8 text-slate-400 mx-auto" />
              <p class="text-xs font-semibold text-slate-500 dark:text-slate-400">
                {{ t('preview_no_site_selected') }}
              </p>
            </div>
          </div>
        </div>

        <!-- TAB 2: CONTRAST CHECKER (WCAG 2.x AA & AAA) -->
        <div
          v-else-if="activeTab === 'contrast'"
          class="space-y-5 max-w-2xl mx-auto py-2"
        >
          <!-- Controls Grid -->
          <div class="grid grid-cols-1 md:grid-cols-[1fr,auto,1fr] gap-3 items-end">
            <!-- Text Color (Foreground) -->
            <div class="space-y-1">
              <label
                for="contrast-fg-text"
                class="block text-xs font-bold text-slate-700 dark:text-slate-300"
              >
                {{ t('text_color_label') }}
              </label>
              <div class="flex items-center gap-2">
                <input
                  id="contrast-fg-color"
                  v-model="contrastTextColor"
                  type="color"
                  class="w-10 h-10 rounded-xl cursor-pointer border border-slate-200 dark:border-slate-700 bg-transparent p-0.5"
                  :aria-label="t('text_color_label')"
                >
                <input
                  id="contrast-fg-text"
                  v-model="contrastTextColor"
                  type="text"
                  dir="ltr"
                  class="flex-1 px-3 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white uppercase focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  maxlength="7"
                >
              </div>
            </div>

            <!-- Swap Button -->
            <div class="flex justify-center pb-0.5">
              <button
                type="button"
                class="p-2.5 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors cursor-pointer"
                :title="t('contrast_swap_colors')"
                :aria-label="t('contrast_swap_colors')"
                @click="swapContrastColors"
              >
                <ArrowUpDown class="w-4 h-4" />
              </button>
            </div>

            <!-- Background Color -->
            <div class="space-y-1">
              <label
                for="contrast-bg-text"
                class="block text-xs font-bold text-slate-700 dark:text-slate-300"
              >
                {{ t('bg_color_label') }}
              </label>
              <div class="flex items-center gap-2">
                <input
                  id="contrast-bg-color"
                  v-model="contrastBgColor"
                  type="color"
                  class="w-10 h-10 rounded-xl cursor-pointer border border-slate-200 dark:border-slate-700 bg-transparent p-0.5"
                  :aria-label="t('bg_color_label')"
                >
                <input
                  id="contrast-bg-text"
                  v-model="contrastBgColor"
                  type="text"
                  dir="ltr"
                  class="flex-1 px-3 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-white uppercase focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  maxlength="7"
                >
              </div>
            </div>
          </div>

          <!-- Sample Live Preview Card -->
          <div
            class="p-6 rounded-2xl border transition-all shadow-sm"
            :style="{
              backgroundColor: contrastBgColor,
              color: contrastTextColor,
              borderColor: contrastTextColor + '33'
            }"
          >
            <div class="flex items-center justify-between gap-3 mb-2">
              <span class="text-xs font-mono font-bold opacity-75">
                Aa Bb Cc 123
              </span>
              <span
                class="text-[11px] font-mono px-2 py-0.5 rounded font-bold border"
                :style="{ borderColor: contrastTextColor + '44' }"
              >
                <bdi dir="ltr">{{ wcagInfo.formattedRatio }}</bdi>
              </span>
            </div>
            <h4 class="text-lg font-bold mb-1">
              {{ t('contrast_sample_heading') }}
            </h4>
            <p class="text-xs leading-relaxed opacity-90">
              {{ t('contrast_sample_body') }}
            </p>
          </div>

          <!-- Results & Badges -->
          <div
            class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-850 space-y-4"
          >
            <!-- Contrast Ratio Big Banner -->
            <div class="flex items-center justify-between flex-wrap gap-2">
              <div class="flex items-center gap-2">
                <span class="text-xs font-bold text-slate-600 dark:text-slate-300">
                  {{ t('contrast_ratio') }}:
                </span>
                <span class="text-2xl font-black font-mono text-slate-900 dark:text-white">
                  <bdi dir="ltr">{{ wcagInfo.formattedRatio }}</bdi>
                </span>
              </div>

              <!-- Overall AA verdict badge -->
              <span
                class="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-bold border"
                :class="
                  wcagInfo.aaNormal
                    ? 'bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950/80 dark:text-emerald-300 dark:border-emerald-800'
                    : 'bg-rose-50 text-rose-700 border-rose-300 dark:bg-rose-950/80 dark:text-rose-300 dark:border-rose-800'
                "
              >
                <CircleCheck
                  v-if="wcagInfo.aaNormal"
                  class="w-3.5 h-3.5"
                />
                <CircleX
                  v-else
                  class="w-3.5 h-3.5"
                />
                <span>{{ wcagInfo.aaNormal ? t('wcag_pass_aa') : t('wcag_fail_aa') }}</span>
              </span>
            </div>

            <!-- Detailed Thresholds Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2 border-t border-slate-200 dark:border-slate-800">
              <!-- AA Normal Text (>= 4.5:1) -->
              <div
                class="p-3 rounded-lg border flex items-center justify-between gap-2"
                :class="
                  wcagInfo.aaNormal
                    ? 'bg-emerald-50/60 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-900/50'
                    : 'bg-rose-50/60 border-rose-200 dark:bg-rose-950/30 dark:border-rose-900/50'
                "
              >
                <div>
                  <div class="text-xs font-bold text-slate-800 dark:text-slate-200">
                    {{ t('wcag_aa_level') }}
                  </div>
                  <div class="text-[11px] text-slate-500 dark:text-slate-400">
                    {{ t('wcag_normal_text') }} (4.5:1)
                  </div>
                </div>
                <span
                  class="inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md"
                  :class="
                    wcagInfo.aaNormal
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
                      : 'bg-rose-100 text-rose-800 dark:bg-rose-900 dark:text-rose-200'
                  "
                >
                  <CircleCheck
                    v-if="wcagInfo.aaNormal"
                    class="w-3 h-3"
                  />
                  <CircleX
                    v-else
                    class="w-3 h-3"
                  />
                  <span>{{ wcagInfo.aaNormal ? t('wcag_pass') : t('wcag_fail') }}</span>
                </span>
              </div>

              <!-- AA Large Text (>= 3.0:1) -->
              <div
                class="p-3 rounded-lg border flex items-center justify-between gap-2"
                :class="
                  wcagInfo.aaLarge
                    ? 'bg-emerald-50/60 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-900/50'
                    : 'bg-rose-50/60 border-rose-200 dark:bg-rose-950/30 dark:border-rose-900/50'
                "
              >
                <div>
                  <div class="text-xs font-bold text-slate-800 dark:text-slate-200">
                    {{ t('wcag_aa_level') }}
                  </div>
                  <div class="text-[11px] text-slate-500 dark:text-slate-400">
                    {{ t('wcag_large_text') }} (3.0:1)
                  </div>
                </div>
                <span
                  class="inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md"
                  :class="
                    wcagInfo.aaLarge
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
                      : 'bg-rose-100 text-rose-800 dark:bg-rose-900 dark:text-rose-200'
                  "
                >
                  <CircleCheck
                    v-if="wcagInfo.aaLarge"
                    class="w-3 h-3"
                  />
                  <CircleX
                    v-else
                    class="w-3 h-3"
                  />
                  <span>{{ wcagInfo.aaLarge ? t('wcag_pass') : t('wcag_fail') }}</span>
                </span>
              </div>

              <!-- AAA Normal Text (>= 7.0:1) -->
              <div
                class="p-3 rounded-lg border flex items-center justify-between gap-2"
                :class="
                  wcagInfo.aaaNormal
                    ? 'bg-emerald-50/60 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-900/50'
                    : 'bg-rose-50/60 border-rose-200 dark:bg-rose-950/30 dark:border-rose-900/50'
                "
              >
                <div>
                  <div class="text-xs font-bold text-slate-800 dark:text-slate-200">
                    {{ t('wcag_aaa_level') }}
                  </div>
                  <div class="text-[11px] text-slate-500 dark:text-slate-400">
                    {{ t('wcag_normal_text') }} (7.0:1)
                  </div>
                </div>
                <span
                  class="inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md"
                  :class="
                    wcagInfo.aaaNormal
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
                      : 'bg-rose-100 text-rose-800 dark:bg-rose-900 dark:text-rose-200'
                  "
                >
                  <CircleCheck
                    v-if="wcagInfo.aaaNormal"
                    class="w-3 h-3"
                  />
                  <CircleX
                    v-else
                    class="w-3 h-3"
                  />
                  <span>{{ wcagInfo.aaaNormal ? t('wcag_pass') : t('wcag_fail') }}</span>
                </span>
              </div>

              <!-- AAA Large Text (>= 4.5:1) -->
              <div
                class="p-3 rounded-lg border flex items-center justify-between gap-2"
                :class="
                  wcagInfo.aaaLarge
                    ? 'bg-emerald-50/60 border-emerald-200 dark:bg-emerald-950/30 dark:border-emerald-900/50'
                    : 'bg-rose-50/60 border-rose-200 dark:bg-rose-950/30 dark:border-rose-900/50'
                "
              >
                <div>
                  <div class="text-xs font-bold text-slate-800 dark:text-slate-200">
                    {{ t('wcag_aaa_level') }}
                  </div>
                  <div class="text-[11px] text-slate-500 dark:text-slate-400">
                    {{ t('wcag_large_text') }} (4.5:1)
                  </div>
                </div>
                <span
                  class="inline-flex items-center gap-1 text-[11px] font-bold px-2 py-0.5 rounded-md"
                  :class="
                    wcagInfo.aaaLarge
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
                      : 'bg-rose-100 text-rose-800 dark:bg-rose-900 dark:text-rose-200'
                  "
                >
                  <CircleCheck
                    v-if="wcagInfo.aaaLarge"
                    class="w-3 h-3"
                  />
                  <CircleX
                    v-else
                    class="w-3 h-3"
                  />
                  <span>{{ wcagInfo.aaaLarge ? t('wcag_pass') : t('wcag_fail') }}</span>
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- TAB 3: DESIGN CLOUD TOOLS & RESOURCES -->
        <div
          v-else-if="activeTab === 'cloud'"
          class="grid grid-cols-1 md:grid-cols-2 gap-4 py-2"
        >
          <!-- Figma Web -->
          <a
            href="https://www.figma.com/files"
            target="_blank"
            rel="noopener noreferrer"
            class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-850 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
            :title="t('open_external_resource')"
            @click.prevent="openUrl('https://www.figma.com/files')"
          >
            <div
              class="w-10 h-10 rounded-xl bg-purple-100 dark:bg-purple-950 text-purple-600 dark:text-purple-400 flex items-center justify-center font-bold flex-shrink-0"
            >
              <Layout class="w-5 h-5" />
            </div>
            <div class="min-w-0 flex-1">
              <h4
                class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
              >
                <span>{{ t('open_figma_web') }}</span>
                <ExternalLink class="w-3 h-3 shrink-0" />
              </h4>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {{ t('resource_figma_desc') }}
              </p>
            </div>
          </a>

          <!-- Coolors Palette -->
          <a
            href="https://coolors.co/generate"
            target="_blank"
            rel="noopener noreferrer"
            class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-850 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
            :title="t('open_external_resource')"
            @click.prevent="openUrl('https://coolors.co/generate')"
          >
            <div
              class="w-10 h-10 rounded-xl bg-blue-100 dark:bg-blue-950 text-blue-600 dark:text-blue-400 flex items-center justify-center font-bold flex-shrink-0"
            >
              <Palette class="w-5 h-5" />
            </div>
            <div class="min-w-0 flex-1">
              <h4
                class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
              >
                <span>{{ t('open_coolors') }}</span>
                <ExternalLink class="w-3 h-3 shrink-0" />
              </h4>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {{ t('resource_coolors_desc') }}
              </p>
            </div>
          </a>

          <!-- Google Fonts -->
          <a
            href="https://fonts.google.com"
            target="_blank"
            rel="noopener noreferrer"
            class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-850 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
            :title="t('open_external_resource')"
            @click.prevent="openUrl('https://fonts.google.com')"
          >
            <div
              class="w-10 h-10 rounded-xl bg-amber-100 dark:bg-amber-950 text-amber-600 dark:text-amber-400 flex items-center justify-center font-bold flex-shrink-0"
            >
              <Type class="w-5 h-5" />
            </div>
            <div class="min-w-0 flex-1">
              <h4
                class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
              >
                <span>{{ t('open_google_fonts') }}</span>
                <ExternalLink class="w-3 h-3 shrink-0" />
              </h4>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {{ t('resource_google_fonts_desc') }}
              </p>
            </div>
          </a>

          <!-- Lucide Icons -->
          <a
            href="https://lucide.dev/icons"
            target="_blank"
            rel="noopener noreferrer"
            class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-850 hover:border-indigo-400 dark:hover:border-indigo-600 transition-all flex items-start gap-3.5 group cursor-pointer"
            :title="t('open_external_resource')"
            @click.prevent="openUrl('https://lucide.dev/icons')"
          >
            <div
              class="w-10 h-10 rounded-xl bg-emerald-100 dark:bg-emerald-950 text-emerald-600 dark:text-emerald-400 flex items-center justify-center font-bold flex-shrink-0"
            >
              <Layout class="w-5 h-5" />
            </div>
            <div class="min-w-0 flex-1">
              <h4
                class="text-sm font-bold text-slate-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 flex items-center gap-1"
              >
                <span>{{ t('open_lucide_icons') }}</span>
                <ExternalLink class="w-3 h-3 shrink-0" />
              </h4>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
                {{ t('resource_lucide_desc') }}
              </p>
            </div>
          </a>
        </div>
      </div>

      <!-- Footer -->
      <div
        class="px-6 py-3 border-t border-slate-100 dark:border-slate-800 flex justify-end flex-shrink-0 bg-slate-50/50 dark:bg-slate-900/50"
      >
        <button
          type="button"
          class="px-4 py-2 text-xs font-semibold rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-200 hover:bg-slate-300 dark:hover:bg-slate-600 transition-colors cursor-pointer"
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
  Layout,
  Smartphone,
  Tablet,
  Monitor,
  RotateCw,
  RefreshCw,
  Palette,
  Type,
  ExternalLink,
  CircleCheck,
  CircleX,
  ArrowUpDown,
  X,
} from 'lucide-vue-next'
import { getWcagDetails } from '../utils/contrast'

export interface SiteItem {
  name?: string
  domain: string
  is_web?: boolean
  path?: string
  root_path?: string
}

interface Props {
  modelValue: boolean
  sites: SiteItem[]
  initialDomain?: string
  t: (key: string, params?: Record<string, string | number>) => string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  sites: () => [],
  initialDomain: '',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
}>()

const activeTab = ref<'responsive' | 'contrast' | 'cloud'>('responsive')

// Viewport state
const selectedDomain = ref('')
const viewportMode = ref<'mobile' | 'tablet' | 'desktop'>('mobile')
const isRotated = ref(false)
const frameKey = ref(1)
const isRefreshing = ref(false)

// Contrast state
const contrastTextColor = ref('#0F172A')
const contrastBgColor = ref('#FFFFFF')

const webSitesList = computed(() => {
  return props.sites.filter((s) => s.is_web && s.domain)
})

watch(
  () => props.initialDomain,
  (newDomain) => {
    if (newDomain) {
      selectedDomain.value = newDomain
    }
  },
  { immediate: true },
)

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen && !selectedDomain.value && webSitesList.value.length > 0) {
      selectedDomain.value = props.initialDomain || webSitesList.value[0].domain
    }
  },
)

const currentWidth = computed(() => {
  if (viewportMode.value === 'desktop') return 1024
  if (viewportMode.value === 'tablet') {
    return isRotated.value ? 1024 : 768
  }
  return isRotated.value ? 667 : 375
})

const currentHeight = computed(() => {
  if (viewportMode.value === 'desktop') return 640
  if (viewportMode.value === 'tablet') {
    return isRotated.value ? 768 : 1024
  }
  return isRotated.value ? 375 : 667
})

const previewFrameContainerStyle = computed(() => {
  if (viewportMode.value === 'desktop') {
    return { width: '100%', maxWidth: '100%' }
  }
  return {
    width: `${currentWidth.value}px`,
    maxWidth: '100%',
  }
})

const previewIframeStyle = computed(() => {
  if (viewportMode.value === 'desktop') {
    return { width: '100%', height: '520px' }
  }
  return {
    width: '100%',
    height: `${Math.min(currentHeight.value, 520)}px`,
  }
})

const wcagInfo = computed(() => {
  return getWcagDetails(contrastTextColor.value, contrastBgColor.value)
})

function swapContrastColors() {
  const temp = contrastTextColor.value
  contrastTextColor.value = contrastBgColor.value
  contrastBgColor.value = temp
}

function reloadPreviewFrame() {
  isRefreshing.value = true
  frameKey.value++
  setTimeout(() => {
    isRefreshing.value = false
  }, 400)
}

function openUrl(url: string) {
  if (typeof (window as any).runtime?.BrowserOpenURL === 'function') {
    ;(window as any).runtime.BrowserOpenURL(url)
  } else {
    window.open(url, '_blank', 'noopener,noreferrer')
  }
}

function closeModal() {
  emit('update:modelValue', false)
  emit('close')
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
