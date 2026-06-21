<script setup lang="ts">
import type { ProxyGatewayDynamicIPProvidersState } from '~/composables/useProxyGatewayDynamicIPProviders'

defineProps<{ runtime: ProxyGatewayDynamicIPProvidersState }>()
const catalog = ref<{ openCreate: () => void }>()

function openCreate() {
  catalog.value?.openCreate()
}

defineExpose({ openCreate })
</script>

<template>
  <div class="flex min-h-0 flex-col gap-4">
    <div v-if="runtime.error.value" class="alert alert-error text-sm">
      {{ runtime.error.value }}
    </div>

    <DynamicIPProviderCatalog ref="catalog" :runtime="runtime" />
  </div>
</template>
