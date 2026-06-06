<script setup lang="ts">
import type { ProxyDynamicIPProviderSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { endpointIDFromURL } from '~/composables/proxyRuntimeDynamicProfilePolicyHelpers'

const props = defineProps<{
  dynamicProviderId: string
  modelValue: string
  providers: ProxyDynamicIPProviderSettings[]
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const endpointOptions = computed(() => {
  if (!props.dynamicProviderId) return []
  const seen = new Set<string>()
  return props.providers
    .filter(
      (provider) => provider.dynamic_provider_id === props.dynamicProviderId,
    )
    .flatMap((provider) =>
      (provider.endpoints || []).map((endpoint) => {
        const endpointURL = endpoint.endpoint_url || ''
        const endpointID = endpointIDFromURL(endpointURL)
        const providerName =
          provider.display_name || provider.dynamic_provider_id || provider.provider_id
        return {
          id: endpointID,
          label: `${providerName} / ${endpointURL}`,
        }
      }),
    )
    .filter((option) => {
      if (!option.id || seen.has(option.id)) return false
      seen.add(option.id)
      return true
    })
})

function selectEndpoint(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value)
}

const placeholder = computed(() => {
  if (!props.dynamicProviderId) return '先选择动态提供商'
  if (endpointOptions.value.length === 0) return '该提供商暂无端点'
  return '任意端点'
})
</script>

<template>
  <select
    class="select-bordered select w-full"
    :disabled="!dynamicProviderId || endpointOptions.length === 0"
    :value="modelValue"
    @change="selectEndpoint"
  >
    <option value="">{{ placeholder }}</option>
    <option v-for="endpoint in endpointOptions" :key="endpoint.id" :value="endpoint.id">
      {{ endpoint.label }}
    </option>
  </select>
</template>
