<template>
  <div class="space-y-5 max-w-2xl mx-auto py-2">
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
            class="flex-1 px-3 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-850 text-slate-900 dark:text-white uppercase focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
            class="flex-1 px-3 py-2 rounded-xl text-xs font-mono border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-850 text-slate-900 dark:text-white uppercase focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
        borderColor: contrastTextColor + '33',
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
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  ArrowUpDown,
  CircleCheck,
  CircleX,
} from 'lucide-vue-next'
import { getWcagDetails } from '../../utils/contrast'

interface Props {
  t: (key: string, params?: Record<string, string | number | undefined>) => string
}

defineProps<Props>()

const contrastTextColor = ref('#0F172A')
const contrastBgColor = ref('#FFFFFF')

const wcagInfo = computed(() => {
  return getWcagDetails(contrastTextColor.value, contrastBgColor.value)
})

function swapContrastColors() {
  const temp = contrastTextColor.value
  contrastTextColor.value = contrastBgColor.value
  contrastBgColor.value = temp
}
</script>
