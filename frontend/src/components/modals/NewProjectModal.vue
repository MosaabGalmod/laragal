<template>
  <div
    v-if="modelValue"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs"
    role="dialog"
    aria-modal="true"
    aria-labelledby="new-project-modal-title"
    tabindex="-1"
    @click.self="closeModal"
  >
    <div
      class="w-full max-w-md rounded-2xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-in"
    >
      <div
        class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3"
      >
        <h3
          id="new-project-modal-title"
          class="font-bold text-base text-slate-900 dark:text-white"
        >
          {{ t('modal_new_title') }}
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

      <!-- Custom validated form with novalidate (Zero native browser tooltips) -->
      <form
        novalidate
        class="space-y-4"
        @submit.prevent="handleCreateProject"
      >
        <div>
          <label
            for="projectNameInput"
            class="block text-xs font-bold text-slate-700 dark:text-slate-300 mb-1"
          >
            <span>{{ t('project_name') }}</span>
            <span class="text-rose-500 ms-1">*</span>
          </label>
          <input
            id="projectNameInput"
            v-model="newProject.name"
            type="text"
            autocomplete="off"
            :aria-invalid="!!projectError"
            aria-describedby="projectNameError"
            :placeholder="t('project_name_placeholder')"
            class="w-full px-3.5 py-2.5 rounded-lg border text-sm text-slate-900 dark:text-white transition-colors focus:outline-none"
            :class="
              projectError
                ? 'border-rose-500 bg-rose-50/40 dark:bg-rose-950/20 focus:ring-2 focus:ring-rose-500'
                : 'border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 focus:ring-2 focus:ring-blue-500'
            "
            @input="validateProjectName"
            @blur="validateProjectName"
          >
          <!-- Custom accessible validation message -->
          <p
            v-if="projectError"
            id="projectNameError"
            role="alert"
            aria-live="polite"
            class="text-xs text-rose-500 font-semibold mt-1.5 flex items-center gap-1"
          >
            <AlertCircle class="w-3.5 h-3.5 flex-shrink-0" />
            <span>{{ projectError }}</span>
          </p>
          <p
            v-else
            class="text-[11px] text-slate-400 mt-1 font-mono"
          >
            <span>{{ t('auto_url_hint') }}</span>
            <bdi dir="ltr">https://{{ newProject.name ? newProject.name.toLowerCase() : 'name' }}.test</bdi>
          </p>
        </div>

        <div>
          <label
            class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1"
          >{{ t('target_directory') }}</label>
          <div class="flex items-center gap-2">
            <input
              type="text"
              readonly
              :value="projectsDirectory + (newProject.name ? '/' + newProject.name.trim().toLowerCase() : '')"
              dir="ltr"
              class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-xs font-mono text-slate-600 dark:text-slate-300 focus:outline-none"
            >
            <button
              type="button"
              class="flex-shrink-0 px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-xs font-bold text-slate-700 dark:text-slate-200 flex items-center gap-1 transition-colors cursor-pointer"
              :title="t('change_dir_hint')"
              @click="$emit('changeDirectory')"
            >
              <FolderOpen class="w-3.5 h-3.5 text-amber-500" />
              <span>{{ t('change_btn') }}</span>
            </button>
          </div>
        </div>

        <div>
          <label
            class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1"
          >{{ t('project_type') }}</label>
          <select
            v-model="newProject.type"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm text-slate-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:outline-none"
          >
            <option value="laravel">
              {{ t('type_laravel') }}
            </option>
            <option value="wordpress">
              {{ t('type_wordpress') }}
            </option>
            <option value="generic">
              {{ t('type_generic') }}
            </option>
          </select>
        </div>

        <div>
          <label
            class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1"
          >{{ t('php_version') }}</label>
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
          <div
            v-if="selectedNewProjectPhpMeta?.eol"
            class="mt-1.5 flex items-center gap-1.5"
          >
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
          <div
            v-else-if="selectedNewProjectPhpMeta?.eol_date"
            class="mt-1.5 flex items-center gap-1.5"
          >
            <span
              class="inline-flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded font-bold bg-amber-100 text-amber-700 dark:bg-amber-950/80 dark:text-amber-300 border border-amber-200 dark:border-amber-900"
            >
              {{ t('runtime_security_only', { date: selectedNewProjectPhpMeta.eol_date }) }}
            </span>
          </div>
        </div>

        <div
          class="pt-4 flex items-center justify-end gap-2 border-t border-slate-100 dark:border-slate-800"
        >
          <button
            type="button"
            class="px-3.5 py-1.5 text-xs font-semibold rounded-lg text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            @click="closeModal"
          >
            {{ t('cancel_btn') }}
          </button>
          <button
            type="submit"
            class="px-4 py-2 text-xs font-bold rounded-lg bg-blue-600 text-white hover:bg-blue-700 shadow-sm transition-colors cursor-pointer"
          >
            {{ t('create_btn') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import {
  AlertCircle,
  FolderOpen,
  AlertTriangle,
} from 'lucide-vue-next'
import type { SiteItem } from '../../types'
import { useI18n } from '../../composables/useI18n'
import { useToast } from '../../composables/useToast'
import { useRuntimeCatalog } from '../../composables/useRuntimeCatalog'

interface Props {
  modelValue: boolean
  projectsDirectory: string
  sites?: SiteItem[]
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  projectsDirectory: 'C:/laragal/sites',
  sites: () => [],
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'changeDirectory'): void
  (e: 'projectCreated', project: { name: string; type: string; php: string }): void
}>()

const { t } = useI18n()
const { showToast } = useToast()
const { phpCatalog, phpSelectOptions } = useRuntimeCatalog()

const newProject = ref({
  name: '',
  type: 'laravel',
  php: '8.3',
})

const projectError = ref('')

const selectedNewProjectPhpMeta = computed(() => {
  const ver = newProject.value.php
  return phpCatalog.value.find(
    (item) => item.name === `php-${ver}` || item.name === ver,
  )
})

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      newProject.value.name = ''
      projectError.value = ''
    }
  },
)

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
  if (props.sites.some((s) => s.domain.toLowerCase() === domain)) {
    projectError.value = t('val_name_exists')
    return false
  }
  projectError.value = ''
  return true
}

async function handleCreateProject() {
  if (!validateProjectName()) return

  const projectName = newProject.value.name.trim().toLowerCase()
  const domain = `${projectName}.test`
  const w = window as any

  if (w.go?.main?.App?.CreateProject) {
    try {
      await w.go.main.App.CreateProject(
        projectName,
        newProject.value.type,
        newProject.value.php,
      )
    } catch (e: any) {
      console.warn('Go backend create project:', e)
    }
  }

  showToast(
    t('project_created_success', { name: projectName, domain }),
  )
  emit('projectCreated', {
    name: projectName,
    type: newProject.value.type,
    php: newProject.value.php,
  })
  closeModal()
}

function closeModal() {
  emit('update:modelValue', false)
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
