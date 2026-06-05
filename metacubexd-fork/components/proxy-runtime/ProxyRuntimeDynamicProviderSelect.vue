<script setup lang="ts">
import type { ProxyProviderDescriptor } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

defineProps<{
  modelValue: string
  providers: ProxyProviderDescriptor[]
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function providerLabel(provider: ProxyProviderDescriptor) {
  return provider.display_name || provider.provider_id
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
      :key="provider.provider_id"
      :value="provider.provider_id"
    >
      {{ providerLabel(provider) }}
    </option>
  </select>
</template>
