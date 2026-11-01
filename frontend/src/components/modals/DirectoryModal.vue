<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="dir-modal-title"
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
          id="dir-modal-title"
          class="font-bold text-base text-slate-900 dark:text-white flex items-center gap-2"
        >
          <FolderOpen class="w-4 h-4 text-amber-500" />
          {{ t('modal_dir_title') }}
        </h3>
        <button
          type="button"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 font-bold cursor-pointer"
          :title="t('close_btn')"
          :aria-label="t('close_btn')"
          @click="closeModal"
        >
          &times;
        </button>
      </div>

      <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
        {{ t('modal_dir_desc') }}
      </p>

      <form
        novalidate
        class="space-y-4"
        @submit.prevent="applyCustomDirectory"
      >
        <div>
          <label
            for="customDirInput"
            class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1"
          >
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
            >
            <button
              type="button"
              class="px-3 py-2 rounded-xl text-xs font-semibold bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors flex items-center gap-1.5 flex-shrink-0 cursor-pointer"
              :title="t('browse_windows_btn')"
              @click="triggerWindowsPicker"
            >
              <FolderOpen class="w-3.5 h-3.5 text-amber-500" />
              <span class="hidden sm:inline">{{ t('browse_windows_btn') }}</span>
            </button>
          </div>
        </div>

        <div v-if="suggestedPaths && suggestedPaths.length > 0">
          <span class="block text-[11px] font-bold text-slate-400 mb-1.5">{{ t('quick_paths') }}:</span>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="p in suggestedPaths"
              :key="p"
              type="button"
              dir="ltr"
              class="px-2.5 py-1 rounded-lg text-[11px] font-mono border transition-all cursor-pointer"
              :class="
                customDirInput === p
                  ? 'bg-blue-50 border-blue-300 text-blue-700 dark:bg-blue-950 dark:border-blue-700 dark:text-blue-300 font-bold'
                  : 'bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700'
              "
              @click="customDirInput = p"
            >
              {{ p }}
            </button>
          </div>
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
            {{ t('save_dir_btn') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { FolderOpen } from 'lucide-vue-next'
import { useI18n } from '../../composables/useI18n'

interface Props {
  modelValue: boolean
  currentDirectory: string
  suggestedPaths?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  currentDirectory: '',
  suggestedPaths: () => [],
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'save', path: string): void
}>()

const { t, currentLang } = useI18n()
const customDirInput = ref('')

watch(
  () => props.currentDirectory,
  (val) => {
    customDirInput.value = (val || '').replace(/\\/g, '/')
  },
  { immediate: true },
)

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      customDirInput.value = (props.currentDirectory || '').replace(/\\/g, '/')
    }
  },
)

function closeModal() {
  emit('update:modelValue', false)
}

function applyCustomDirectory() {
  const p = customDirInput.value.trim()
  if (!p) return
  const norm = p.replace(/\\/g, '/')
  emit('save', norm)
  closeModal()
}

async function triggerWindowsPicker() {
  const w = window as any
  if (w.go?.main?.App?.SelectProjectsDirectory) {
    try {
      const newPath = await w.go.main.App.SelectProjectsDirectory(currentLang.value)
      if (newPath) {
        const norm = newPath.replace(/\\/g, '/')
        customDirInput.value = norm
        emit('save', norm)
        closeModal()
      }
    } catch (e: any) {
      console.warn('Picker error:', e)
    }
  }
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
