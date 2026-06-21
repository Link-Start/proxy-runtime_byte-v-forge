<script setup lang="ts">
import { IconAlertTriangle, IconCircleCheck, IconLoader2 } from '@tabler/icons-vue'
import type { ProxyGatewayStatusState } from '~/composables/useProxyGatewayStatus'

const props = defineProps<{ state: ProxyGatewayStatusState }>()

const status = computed(() => props.state.status.value)
const applying = computed(
  () =>
    status.value?.reconcile_pending ||
    status.value?.reconcile_running ||
    status.value?.lease_restore_running,
)
const healthy = computed(() => Boolean(status.value?.ready))
const label = computed(() => {
  if (props.state.error.value) return props.state.error.value
  if (applying.value) return '应用中'
  if (healthy.value) return '运行中'
  if (status.value?.status) return status.value.status
  return props.state.loading.value ? '加载中' : '未知'
})
const badgeClass = computed(() => {
  if (props.state.error.value) return 'badge-error'
  if (applying.value) return 'badge-warning'
  if (healthy.value) return 'badge-success'
  return 'badge-ghost'
})
</script>

<template>
  <span
    class="badge badge-sm gap-1"
    :class="badgeClass"
    :title="label"
  >
    <IconLoader2 v-if="applying || state.loading.value" :size="13" class="animate-spin" />
    <IconCircleCheck v-else-if="healthy" :size="13" />
    <IconAlertTriangle v-else :size="13" />
    {{ label }}
  </span>
</template>
