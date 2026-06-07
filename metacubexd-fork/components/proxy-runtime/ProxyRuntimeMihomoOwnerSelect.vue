<script setup lang="ts">
import {
  mihomoOwnerLabel,
  type MihomoEgressOwner,
} from '~/composables/proxyRuntimeMihomoOwnerHelpers'

defineProps<{
  modelValue: string
  placeholder: string
  owners: MihomoEgressOwner[]
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

function selectOwner(event: Event) {
  emit('update:modelValue', (event.target as HTMLSelectElement).value)
}
</script>

<template>
  <select
    class="select-bordered select w-full"
    required
    :value="modelValue"
    @change="selectOwner"
  >
    <option value="">{{ placeholder }}</option>
    <option v-for="owner in owners" :key="owner.owner_id" :value="owner.owner_id">
      {{ mihomoOwnerLabel(owner) }}
    </option>
  </select>
</template>
