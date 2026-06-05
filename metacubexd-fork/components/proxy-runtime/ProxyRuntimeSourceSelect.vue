<script setup lang="ts">
import type { ProxySourceDescriptor } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { sourceLabel } from '~/composables/proxyRuntimeEgressProfileHelpers'

defineProps<{
  modelValue: string
  placeholder: string
  sources: ProxySourceDescriptor[]
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function selectSource(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value)
}
</script>

<template>
  <select
    class="select-bordered select w-full"
    required
    :value="modelValue"
    @change="selectSource"
  >
    <option value="">{{ placeholder }}</option>
    <option v-for="source in sources" :key="source.source_id" :value="source.source_id">
      {{ sourceLabel(source) }}
    </option>
  </select>
</template>
