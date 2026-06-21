<script setup lang="ts">
import type { ProxyDynamicIPProviderSettings } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

defineProps<{
  modelValue: string
  providers: ProxyDynamicIPProviderSettings[]
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function providerLabel(provider: ProxyDynamicIPProviderSettings) {
  return provider.display_name || provider.dynamic_provider_id
}

function selectProvider(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value)
}
</script>

<template>
  <select class="select-bordered select w-full" :value="modelValue" @change="selectProvider">
    <option value="">不指定</option>
    <option
      v-for="provider in providers"
      :key="provider.dynamic_provider_id"
      :value="provider.dynamic_provider_id"
    >
      {{ providerLabel(provider) }}
    </option>
  </select>
</template>
