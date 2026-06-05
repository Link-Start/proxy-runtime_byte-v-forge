<script setup lang="ts">
import type { ProxyRuntimeDynamicIPProvidersState } from '~/composables/useProxyRuntimeDynamicIPProviders'

defineProps<{ runtime: ProxyRuntimeDynamicIPProvidersState }>()
</script>

<template>
  <div class="flex min-h-0 flex-col gap-4">
    <div v-if="runtime.error.value" class="alert alert-error text-sm">
      {{ runtime.error.value }}
    </div>

    <DynamicIPProviderCatalog :runtime="runtime" />
  </div>
</template>
