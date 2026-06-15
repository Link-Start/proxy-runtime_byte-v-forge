<script setup lang="ts">
import { IconRefresh, IconTrash, IconWorldBolt } from '@tabler/icons-vue'
import type { ProxyRuntimePlaygroundLeasesState } from '~/composables/useProxyRuntimePlaygroundLeases'
import { ProxyDynamicLeaseStatus, type ProxyDynamicLease } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

defineProps<{ dynamicExit: boolean, proxyAuthority: string, state: ProxyRuntimePlaygroundLeasesState }>()

const expanded = ref(true)

function statusText(status: ProxyDynamicLeaseStatus) {
  switch (status) {
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE:
      return '可用'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_RELEASED:
      return '已释放'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_EXPIRED:
      return '已过期'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_FAILED:
      return '失败'
    default:
      return '未知'
  }
}

function statusClass(status: ProxyDynamicLeaseStatus) {
  if (status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE) return 'badge-success'
  if (status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_FAILED) return 'badge-error'
  return 'badge-ghost'
}

function endpointText(lease: ProxyDynamicLease) {
  return lease.selection_plan?.selected_endpoint?.endpoint_url || lease.egress?.labels?.dynamic_ip_endpoint_id || '-'
}

function egressText(lease: ProxyDynamicLease) {
  const host = lease.egress?.host || ''
  const port = lease.egress?.port || 0
  return host && port ? `${host}:${port}` : '-'
}


function shortID(value: string | undefined) {
  const normalized = value || ''
  if (normalized.length <= 12) return normalized || '-'
  return `${normalized.slice(0, 6)}…${normalized.slice(-4)}`
}

function timeText(value: string | undefined) {
  return value ? new Date(value).toLocaleString() : '-'
}

function active(lease?: ProxyDynamicLease) {
  return lease?.status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE
}
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">租约</h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">{{ state.profileID.value }}</span>
            <span v-if="active(state.currentLease.value)" class="badge badge-primary badge-sm">1 active</span>
            <span v-else-if="dynamicExit" class="badge badge-ghost badge-sm">未申请</span>
            <span v-else class="badge badge-ghost badge-sm">非动态出口</span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <Button class="btn-ghost btn-xs btn-square" :disabled="state.loading.value" title="刷新租约" @click="state.load()">
            <IconRefresh :size="15" :class="{ 'animate-spin': state.loading.value }" />
          </Button>
          <Button class="btn-primary btn-xs btn-square" :disabled="state.busy.value || !state.canAcquire.value" title="申请租约" @click="state.acquire">
            <IconWorldBolt :size="15" />
          </Button>
        </div>
      </div>
    </template>

    <div class="col-span-full flex flex-col gap-2">
      <p v-if="state.error.value" class="alert alert-error py-2 text-xs">{{ state.error.value }}</p>
      <p v-else-if="dynamicExit && !state.canAcquire.value && !active(state.currentLease.value)" class="alert alert-warning py-2 text-xs">
        {{ state.acquireDisabledReason.value }}
      </p>
      <div v-if="!dynamicExit" class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <span class="text-sm">当前出口不是动态 IP</span>
        <span class="truncate font-mono text-xs">{{ proxyAuthority }}</span>
      </div>
      <div v-else-if="!state.currentLease.value" class="px-2 py-3 text-sm opacity-55">暂无 PlayGround 租约</div>
      <div v-else class="flex flex-col gap-2 rounded-lg bg-base-content/5 px-2 py-2 text-sm">
        <div class="flex items-center justify-between gap-4">
          <span>状态</span>
          <div class="flex items-center gap-2">
            <span class="badge badge-xs" :class="statusClass(state.currentLease.value.status)">{{ statusText(state.currentLease.value.status) }}</span>
            <Button v-if="active(state.currentLease.value)" class="btn-ghost btn-xs btn-square" :disabled="state.busy.value" title="释放租约" @click="state.release(state.currentLease.value)">
              <IconTrash :size="14" />
            </Button>
          </div>
        </div>
        <div class="flex items-center justify-between gap-4"><span>租约</span><span class="font-mono text-xs">{{ shortID(state.currentLease.value.lease_id) }}</span></div>
        <div class="flex items-center justify-between gap-4"><span>端点</span><span class="max-w-72 truncate text-xs">{{ endpointText(state.currentLease.value) }}</span></div>
        <div class="flex items-center justify-between gap-4"><span>出口</span><span class="font-mono text-xs">{{ egressText(state.currentLease.value) }}</span></div>
        <div class="flex items-center justify-between gap-4"><span>过期</span><span class="text-xs">{{ timeText(state.currentLease.value.expires_at) }}</span></div>
      </div>
    </div>
  </Collapse>
</template>
