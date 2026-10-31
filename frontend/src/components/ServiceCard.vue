<template>
  <div
    class="h-full flex flex-col rounded-xl border bg-white dark:bg-slate-900 shadow-xs transition-colors p-3.5"
    :class="[
      service.running
        ? 'border-slate-200 dark:border-slate-800 border-s-2 border-s-emerald-500 hover:border-slate-300 dark:hover:border-slate-700'
        : 'border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700'
    ]"
  >
    <!-- Header: Icon, Name + Status, and Switch -->
    <div class="flex items-start justify-between gap-2.5">
      <div class="flex items-center gap-2.5 min-w-0 flex-1">
        <!-- Service Icon -->
        <div
          class="w-8 h-8 rounded-lg bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-slate-600 dark:text-slate-300 shrink-0"
        >
          <component :is="serviceIcon" class="w-4 h-4" />
        </div>

        <!-- Name & Status -->
        <div class="min-w-0 flex-1">
          <h3 class="font-semibold text-sm text-slate-900 dark:text-slate-100 truncate leading-tight">
            {{ t(service.name_key) }}
          </h3>
          <div class="flex items-center gap-1.5 mt-0.5">
            <span
              class="w-1.5 h-1.5 rounded-full shrink-0"
              :class="service.running ? 'bg-emerald-500' : 'bg-slate-400 dark:bg-slate-500'"
            ></span>
            <span
              class="text-[11px] font-medium leading-none"
              :class="service.running ? 'text-emerald-700 dark:text-emerald-400' : 'text-slate-600 dark:text-slate-400'"
            >
              {{ service.running ? t('status_running') : t('status_stopped') }}
            </span>
          </div>
        </div>
      </div>

      <!-- Accessible Switch -->
      <button
        type="button"
        role="switch"
        :aria-checked="service.running"
        :aria-label="t(service.running ? 'stop_btn' : 'start_btn') + ' ' + t(service.name_key)"
        :disabled="busy"
        :aria-busy="busy"
        @click="emit('toggle')"
        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full p-0.5 transition-colors motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-slate-900 disabled:opacity-60 disabled:cursor-wait"
        :class="service.running ? 'bg-emerald-600 dark:bg-emerald-500' : 'bg-slate-300 dark:bg-slate-700'"
      >
        <span
          class="flex h-4 w-4 items-center justify-center rounded-full bg-white dark:bg-slate-100 shadow-xs transition-transform motion-reduce:transition-none"
          :class="service.running ? 'translate-x-4 rtl:-translate-x-4' : 'translate-x-0'"
        >
          <Loader2
            v-if="busy"
            class="h-2.5 w-2.5 animate-spin text-slate-700 motion-reduce:animate-none"
          />
        </span>
      </button>
    </div>

    <!-- Description -->
    <p
      class="mt-2 text-xs text-slate-500 dark:text-slate-400 line-clamp-2 leading-relaxed"
      :title="t(service.desc_key)"
    >
      {{ t(service.desc_key) }}
    </p>

    <!-- Details Grid -->
    <div class="mt-3 p-2.5 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800/70 select-text">
      <dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-[11px] items-center">
        <!-- Port (Always included) -->
        <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('port_label') }}</dt>
        <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200">
          {{ service.port }}
        </dd>

        <!-- MySQL Rows -->
        <template v-if="service.id === 'mysql'">
          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('user_label') }}</dt>
          <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200 flex items-center gap-1">
            <span>root</span>
            <button
              type="button"
              @click.stop="emit('copy', 'root')"
              class="p-0.5 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors inline-flex items-center justify-center focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500"
              :aria-label="t('copy_value')"
              :title="t('copy_value')"
            >
              <Check v-if="copiedItem === 'root'" class="w-3 h-3 text-emerald-500" />
              <Copy v-else class="w-3 h-3" />
            </button>
          </dd>

          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('password_label') }}</dt>
          <dd class="justify-self-end text-end text-slate-500 dark:text-slate-400 italic text-[10px]">
            {{ t('empty_password') }}
          </dd>

          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('host_label') }}</dt>
          <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200">
            127.0.0.1:3306
          </dd>
        </template>

        <!-- Mailpit Rows -->
        <template v-else-if="service.id === 'mailpit'">
          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('smtp_label') }}</dt>
          <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200 flex items-center gap-1">
            <span>127.0.0.1:1025</span>
            <button
              type="button"
              @click.stop="emit('copy', '127.0.0.1:1025')"
              class="p-0.5 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors inline-flex items-center justify-center focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-blue-500"
              :aria-label="t('copy_value')"
              :title="t('copy_value')"
            >
              <Check v-if="copiedItem === '127.0.0.1:1025'" class="w-3 h-3 text-emerald-500" />
              <Copy v-else class="w-3 h-3" />
            </button>
          </dd>

          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('web_ui_label') }}</dt>
          <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200">
            127.0.0.1:8025
          </dd>
        </template>

        <!-- Redis Rows -->
        <template v-else-if="service.id === 'redis'">
          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('host_label') }}</dt>
          <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200">
            127.0.0.1:6379
          </dd>

          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('password_label') }}</dt>
          <dd class="justify-self-end text-end text-slate-500 dark:text-slate-400 italic text-[10px]">
            {{ t('no_auth') }}
          </dd>
        </template>

        <!-- PHP-* Rows -->
        <template v-else-if="service.id.startsWith('php-')">
          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('workers_label') }}</dt>
          <dd class="justify-self-end text-end text-slate-700 dark:text-slate-200">
            <bdi dir="ltr" class="font-mono">8</bdi> {{ t('workers_unit') }}
          </dd>

          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('opcache_label') }}</dt>
          <dd class="justify-self-end text-end text-emerald-600 dark:text-emerald-400 font-medium">
            {{ t('enabled_value') }}
          </dd>
        </template>

        <!-- Caddy Rows -->
        <template v-else-if="service.id === 'caddy'">
          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('admin_api_label') }}</dt>
          <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200">
            localhost:2019
          </dd>

          <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('tls_label') }}</dt>
          <dd class="justify-self-end text-end text-emerald-600 dark:text-emerald-400 font-medium">
            <span>{{ t('tls_auto_value') }}</span> <bdi dir="ltr" class="font-mono">*.test</bdi>
          </dd>
        </template>

        <!-- Generic Fallback -->
        <template v-else-if="service.credentials">
          <template v-if="service.credentials.host">
            <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('host_label') }}</dt>
            <dd dir="ltr" class="justify-self-end text-end font-mono text-slate-700 dark:text-slate-200">{{ service.credentials.host }}</dd>
          </template>
          <template v-if="service.credentials.info">
            <dt class="text-slate-500 dark:text-slate-400 font-medium truncate">{{ t('info_label') }}</dt>
            <dd class="justify-self-end text-end text-slate-700 dark:text-slate-200">{{ service.credentials.info }}</dd>
          </template>
        </template>
      </dl>
    </div>

    <!-- Footer Actions (only rendered when there is at least one action) -->
    <div
      v-if="hasActions"
      class="mt-auto pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center gap-2 flex-wrap"
    >
      <button
        v-if="service.id === 'mysql' && service.running"
        type="button"
        @click="emit('open-adminer')"
        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700/60 transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="t('open_adminer_title')"
      >
        <Database class="w-3.5 h-3.5 text-amber-600 dark:text-amber-400 shrink-0" />
        <span>{{ t('open_adminer') }}</span>
      </button>

      <button
        v-if="service.running && service.url"
        type="button"
        @click="emit('open-url', service.url)"
        class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-700/60 transition-colors whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/50"
        :title="service.url"
      >
        <ExternalLink class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400 shrink-0" />
        <span>{{ t('open_dashboard') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Server,
  Code2,
  Database,
  Layers,
  Mail,
  Loader2,
  Check,
  Copy,
  ExternalLink
} from 'lucide-vue-next'

export interface ServiceItem {
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

interface Props {
  service: ServiceItem
  busy?: boolean
  copiedItem?: string
  t: (key: string, params?: Record<string, string | number>) => string
}

const props = withDefaults(defineProps<Props>(), {
  busy: false,
  copiedItem: ''
})

const emit = defineEmits<{
  (e: 'toggle'): void
  (e: 'open-adminer'): void
  (e: 'open-url', url: string): void
  (e: 'copy', text: string): void
}>()

const serviceIcon = computed(() => {
  if (props.service.id === 'caddy') return Server
  if (props.service.id.startsWith('php-')) return Code2
  if (props.service.id === 'mysql') return Database
  if (props.service.id === 'redis') return Layers
  if (props.service.id === 'mailpit') return Mail
  return Server
})

const hasActions = computed(() => {
  return (
    (props.service.id === 'mysql' && props.service.running) ||
    (props.service.running && Boolean(props.service.url))
  )
})
</script>
