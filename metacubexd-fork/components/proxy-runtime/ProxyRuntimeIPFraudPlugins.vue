<script setup lang="ts">
import type { ProxyRuntimePluginsState } from '~/composables/useProxyRuntimePlugins'
import {
  IconPlus,
  IconShieldCheck,
} from '@tabler/icons-vue'

defineProps<{ runtime: ProxyRuntimePluginsState }>()
const modal = ref<{ open: () => void }>()
</script>

<template>
  <section class="flex min-h-0 flex-col gap-3 p-2">
    <div class="flex items-center justify-between gap-3">
      <h2 class="flex items-center gap-2 text-base font-semibold">
        <IconShieldCheck :size="18" />
        IP Fraud 检测
      </h2>
      <div class="flex items-center gap-2">
        <span class="badge badge-primary badge-sm">{{ runtime.fraudRows.value.length }}</span>
        <button
          aria-label="添加检测提供商"
          class="btn btn-primary btn-sm btn-square"
          :disabled="runtime.availableFraudProviderOptions.value.length === 0"
          title="添加检测提供商"
          type="button"
          @click="modal?.open()"
        >
          <IconPlus :size="16" />
        </button>
      </div>
    </div>

    <div
      v-if="runtime.fraudRows.value.length === 0"
      class="grid gap-3 rounded-lg border border-dashed border-base-content/15 p-6 text-center text-sm"
    >
      <div class="opacity-60">暂无已添加 IP Fraud Provider</div>
      <div class="flex justify-center">
        <button
          aria-label="添加检测提供商"
          class="btn btn-primary btn-sm btn-square"
          :disabled="runtime.availableFraudProviderOptions.value.length === 0"
          title="添加检测提供商"
          type="button"
          @click="modal?.open()"
        >
          <IconPlus :size="16" />
        </button>
      </div>
    </div>

    <div v-else class="grid gap-3">
      <ProxyRuntimeIPFraudProviderCard
        v-for="(row, index) in runtime.fraudRows.value"
        :key="`${row.provider_id}-${index}`"
        :index="index"
        :row="row"
        :runtime="runtime"
      />
    </div>

    <ProxyRuntimeIPFraudCheckPanel :runtime="runtime" />
    <ProxyRuntimeIPFraudProviderModal ref="modal" :runtime="runtime" />
  </section>
</template>
