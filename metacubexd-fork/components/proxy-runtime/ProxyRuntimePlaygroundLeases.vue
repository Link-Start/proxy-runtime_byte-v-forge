<script setup lang="ts">
import {
  IconRefresh,
  IconTrash,
  IconWorldBolt,
} from '@tabler/icons-vue'
import type { ProxyRuntimePlaygroundLeasesState } from '~/composables/useProxyRuntimePlaygroundLeases'
import {
  ProxyDynamicLeaseStatus,
  type ProxyDynamicLease,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

defineProps<{
  dynamicExit: boolean
  proxyAuthority: string
  state: ProxyRuntimePlaygroundLeasesState
}>()

const expanded = ref(true)

function statusText(status: ProxyDynamicLeaseStatus) {
  switch (status) {
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE:
      return 'Active'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_RELEASED:
      return 'Released'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_EXPIRED:
      return 'Expired'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_FAILED:
      return 'Failed'
    default:
      return 'Unknown'
  }
}

function statusClass(status: ProxyDynamicLeaseStatus) {
  if (status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE) {
    return 'badge-success'
  }
  if (status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_FAILED) {
    return 'badge-error'
  }
  return 'badge-ghost'
}

function endpointText(lease: ProxyDynamicLease) {
  return (
    lease.selection_plan?.selected_endpoint?.endpoint_url ||
    lease.egress?.labels?.dynamic_ip_endpoint_id ||
    '-'
  )
}

function egressText(lease: ProxyDynamicLease) {
  const host = lease.egress?.host || ''
  const port = lease.egress?.port || 0
  return host && port ? `${host}:${port}` : '-'
}

function detailText(lease: ProxyDynamicLease) {
  return lease.error_message || lease.selection_plan?.selection_reasons?.[0] || '-'
}

function shortID(value: string) {
  if (value.length <= 12) return value || '-'
  return `${value.slice(0, 6)}…${value.slice(-4)}`
}

function timeText(value: string | undefined) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

function active(lease: ProxyDynamicLease) {
  return lease.status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE
}
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">动态租约</h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">{{ state.profileID.value }}</span>
            <span class="badge badge-primary badge-sm" v-if="dynamicExit">
              {{ state.activeRows.value.length }} active
            </span>
            <span v-else class="badge badge-ghost badge-sm">非动态出口</span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <Button
            class="btn-ghost btn-xs btn-square"
            title="刷新租约"
            @click="state.load"
          >
            <IconRefresh :size="15" :class="{ 'animate-spin': state.loading.value }" />
          </Button>
          <Button
            class="btn-primary btn-xs btn-square"
            :disabled="state.busy.value || !state.canAcquire.value || state.activeRows.value.length > 0"
            title="申请租约"
            @click="state.acquire"
          >
            <IconWorldBolt :size="15" />
          </Button>
        </div>
      </div>
    </template>

    <div class="col-span-full flex flex-col gap-2">
      <div
        v-if="!dynamicExit"
        class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5"
      >
        <span class="text-sm">当前出口不是动态 IP</span>
        <span class="truncate font-mono text-xs">{{ proxyAuthority }}</span>
      </div>
      <p v-else-if="state.error.value" class="alert alert-error py-2 text-xs">
        {{ state.error.value }}
      </p>
      <p v-else-if="!state.canAcquire.value" class="alert alert-warning py-2 text-xs">
        {{ state.acquireDisabledReason.value }}
      </p>

      <div
        v-if="dynamicExit && state.rows.value.length === 0"
        class="px-2 py-3 text-sm opacity-55"
      >
        暂无 PlayGround 租约
      </div>
      <div v-else-if="dynamicExit" class="overflow-x-auto">
        <table class="table table-xs">
          <thead>
            <tr>
              <th>状态</th>
              <th>租约</th>
              <th>端点</th>
              <th>出口</th>
              <th>过期</th>
              <th>说明</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="lease in state.rows.value" :key="lease.lease_id">
              <td>
                <span class="badge badge-xs" :class="statusClass(lease.status)">
                  {{ statusText(lease.status) }}
                </span>
              </td>
              <td class="font-mono">{{ shortID(lease.lease_id) }}</td>
              <td class="max-w-48 truncate">{{ endpointText(lease) }}</td>
              <td class="font-mono">{{ egressText(lease) }}</td>
              <td>{{ timeText(lease.expires_at) }}</td>
              <td class="max-w-48 truncate">{{ detailText(lease) }}</td>
              <td class="text-right">
                <Button
                  v-if="active(lease)"
                  class="btn-ghost btn-xs btn-square"
                  :disabled="state.busy.value"
                  title="释放租约"
                  @click="state.release(lease)"
                >
                  <IconTrash :size="14" />
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </Collapse>
</template>
