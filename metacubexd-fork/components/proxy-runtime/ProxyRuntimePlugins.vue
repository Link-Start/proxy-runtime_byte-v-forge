<script setup lang="ts">
import type { ProxyRuntimePluginsState } from '~/composables/useProxyRuntimePlugins'
import { IconReload } from '@tabler/icons-vue'

defineProps<{ runtime: ProxyRuntimePluginsState }>()

type PluginTab = 'ip-fraud' | 'cf-canary' | 'ip-geo'

const activeTab = ref<PluginTab>('ip-fraud')

const tabs = [
  { id: 'ip-fraud' as const, label: 'IP Fraud 检测' },
  { id: 'cf-canary' as const, label: 'CF Canary' },
  { id: 'ip-geo' as const, label: 'IP Geo 检测' },
]
</script>

<template>
  <main class="flex h-full min-h-0 flex-col gap-3">
    <div class="animate-fade-slide-in flex shrink-0 flex-wrap items-center gap-3">
      <div
        class="flex gap-1 rounded-xl border border-base-content/8 bg-base-200/60 p-1 backdrop-blur-sm"
      >
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm text-base-content/70 transition-all duration-200 hover:bg-base-content/5"
          :class="{
            'bg-primary text-primary-content shadow-md shadow-primary/30 hover:bg-primary':
              activeTab === tab.id,
          }"
          type="button"
          @click="activeTab = tab.id"
        >
          <span class="font-medium">{{ tab.label }}</span>
        </button>
      </div>

      <div class="flex items-center gap-2">
        <button
          aria-label="刷新插件"
          class="flex h-9 w-9 items-center justify-center rounded-[0.625rem] border border-base-content/10 bg-base-200/80 transition-all duration-200 hover:border-primary/30 hover:bg-primary/15 hover:text-primary"
          :disabled="runtime.loading.value"
          title="刷新插件"
          type="button"
          @click="runtime.load"
        >
          <IconReload :size="18" :class="{ 'animate-spin': runtime.loading.value }" />
        </button>
      </div>
    </div>

    <div v-if="runtime.error.value" class="alert alert-error py-2 text-sm">
      {{ runtime.error.value }}
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto">
      <ProxyRuntimeIPFraudPlugins v-if="activeTab === 'ip-fraud'" :runtime="runtime" />
      <ProxyRuntimeEdgeCanaryPlugin v-else-if="activeTab === 'cf-canary'" :runtime="runtime" />
      <ProxyRuntimeIPGeoPlugin v-else :runtime="runtime" />
    </div>
  </main>
</template>
