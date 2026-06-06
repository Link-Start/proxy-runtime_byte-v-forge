<script setup lang="ts">
import type { MihomoConfigNode } from '~/composables/proxyRuntimeMihomoController'
import { nodeLabel } from '~/composables/proxyRuntimeEgressProfileHelpers'

interface MihomoNodeRuntime {
  loadMihomoNodes(ownerID: string): Promise<void>
  mihomoNodesLoading(ownerID: string): boolean
  nodesForMihomoOwner(ownerID: string): MihomoConfigNode[]
}

const props = defineProps<{
  modelValue: string
  placeholder: string
  runtime: MihomoNodeRuntime
  ownerId: string
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

watch(
  () => props.ownerId,
  async (ownerID) => {
    if (ownerID) await props.runtime.loadMihomoNodes(ownerID)
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
    :disabled="!ownerId || runtime.mihomoNodesLoading(ownerId)"
    required
    :value="modelValue"
    @change="selectNode"
  >
    <option value="">
      {{ runtime.mihomoNodesLoading(ownerId) ? '加载节点...' : placeholder }}
    </option>
    <option
      v-for="node in runtime.nodesForMihomoOwner(ownerId)"
      :key="node.node_id"
      :value="node.node_id"
    >
      {{ nodeLabel(node) }}
    </option>
  </select>
</template>
