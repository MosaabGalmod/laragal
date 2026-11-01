<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="quit-confirm-modal-title"
    tabindex="-1"
    @click.self="closeModal"
  >
    <div
      class="w-full max-w-sm rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-4 animate-scale-in"
    >
      <div class="flex items-center gap-3 text-rose-600 dark:text-rose-400">
        <Power class="w-6 h-6 flex-shrink-0" />
        <h3
          id="quit-confirm-modal-title"
          class="font-bold text-base text-slate-900 dark:text-white"
        >
          {{ t('confirm_quit_title') }}
        </h3>
      </div>
      <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
        {{ t('confirm_quit_msg') }}
      </p>
      <div
        class="pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-end gap-2"
      >
        <button
          type="button"
          class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          @click="closeModal"
        >
          {{ t('cancel_btn') }}
        </button>
        <button
          type="button"
          class="px-4 py-1.5 text-xs font-bold rounded-lg bg-rose-600 text-white hover:bg-rose-700 shadow-sm transition-colors cursor-pointer"
          @click="handleConfirm"
        >
          {{ t('confirm_btn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { Power } from 'lucide-vue-next'
import { useI18n } from '../../composables/useI18n'

interface Props {
  modelValue: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
}>()

const { t } = useI18n()

function closeModal() {
  emit('update:modelValue', false)
}

function handleConfirm() {
  closeModal()
  emit('confirm')
  const w = window as any
  if (w.go?.main?.App?.QuitApp) {
    w.go.main.App.QuitApp()
  } else if (w.runtime?.Quit) {
    w.runtime.Quit()
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
