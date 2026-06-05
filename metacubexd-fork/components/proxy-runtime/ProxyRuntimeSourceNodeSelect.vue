<script setup lang="ts">
import type { ProxyRuntimeEgressProfilesState } from '~/composables/useProxyRuntimeEgressProfiles'
import { nodeLabel } from '~/composables/proxyRuntimeEgressProfileHelpers'

const props = defineProps<{
  modelValue: string
  placeholder: string
  runtime: ProxyRuntimeEgressProfilesState
  sourceId: string
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

watch(
  () => props.sourceId,
  async (sourceID) => {
    if (sourceID) await props.runtime.loadSourceNodes(sourceID)
  },
  { immediate: true },
)

function selectNode(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value)
}
</script>

<template>
  <select
    class="select-bordered select w-full"
    :disabled="!sourceId || runtime.sourceNodesLoading(sourceId)"
    required
    :value="modelValue"
    @change="selectNode"
  >
    <option value="">
      {{ runtime.sourceNodesLoading(sourceId) ? '加载节点...' : placeholder }}
    </option>
    <option
      v-for="node in runtime.nodesForSource(sourceId)"
      :key="node.node_id"
      :value="node.node_id"
    >
      {{ nodeLabel(node) }}
    </option>
  </select>
</template>
