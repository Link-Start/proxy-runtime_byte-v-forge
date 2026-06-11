<script setup lang="ts">
import { IconReload } from '@tabler/icons-vue'

const nativeRuntime = useProxyRuntimeMihomoNativeConfig()
const dynamicRuntime = useProxyRuntimeDynamicIPProviders()
const activeTab = ref<'mihomo' | 'dynamic-ip'>('mihomo')

const tabs = computed(() => [
  {
    id: 'mihomo' as const,
    label: 'Mihomo 资源',
    count: nativeRuntime.itemCount.value,
  },
  {
    id: 'dynamic-ip' as const,
    label: '动态 IP 提供商',
    count: dynamicRuntime.dynamicProviderCount.value,
  },
])
const activeLoading = computed(() =>
  activeTab.value === 'mihomo'
    ? nativeRuntime.loading.value
    : dynamicRuntime.loading.value,
)

function refreshActive() {
  void (
    activeTab.value === 'mihomo'
      ? nativeRuntime.load()
      : dynamicRuntime.load()
  )
}

useHead({ title: '出口' })
onMounted(() => {
  void Promise.all([nativeRuntime.load(), dynamicRuntime.load()])
})
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
          <span class="badge badge-sm">{{ tab.count }}</span>
        </button>
      </div>

      <button
        aria-label="刷新出口资源"
        class="flex h-9 w-9 items-center justify-center rounded-[0.625rem] border border-base-content/10 bg-base-200/80 transition-all duration-200 hover:border-primary/30 hover:bg-primary/15 hover:text-primary"
        :disabled="activeLoading"
        title="刷新出口资源"
        type="button"
        @click="refreshActive"
      >
        <IconReload :size="18" :class="{ 'animate-spin': activeLoading }" />
      </button>
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto">
      <ProxyRuntimeMihomoNative
        v-if="activeTab === 'mihomo'"
        :runtime="nativeRuntime"
      />
      <ProxyRuntimeDynamicIPProviders v-else :runtime="dynamicRuntime" />
    </div>
  </main>
</template>
