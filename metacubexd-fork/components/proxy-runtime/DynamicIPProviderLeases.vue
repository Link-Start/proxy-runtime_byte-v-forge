<script setup lang="ts">
import { IconTrash } from '@tabler/icons-vue'
import type { ProxyRuntimeDynamicLeasesState } from '~/composables/useProxyRuntimeDynamicLeases'
import type { ProxyDynamicLease } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { ProxyDynamicLeaseStatus } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

defineProps<{ runtime: ProxyRuntimeDynamicLeasesState }>()

function leaseTitle(lease: ProxyDynamicLease) {
  return (
    lease.session?.labels?.display_name ||
    lease.account_id ||
    lease.purpose ||
    lease.lease_id
  )
}

function endpointText(lease: ProxyDynamicLease) {
  if (!lease.egress) return '-'
  return `${lease.egress.host}:${lease.egress.port}`
}

function updatedText(lease: ProxyDynamicLease) {
  if (!lease.acquired_at) return '-'
  return new Date(lease.acquired_at).toLocaleString()
}

function expiresText(lease: ProxyDynamicLease) {
  if (!lease.expires_at) return ''
  return `过期 ${new Date(lease.expires_at).toLocaleString()}`
}

function statusText(status: ProxyDynamicLeaseStatus) {
  switch (status) {
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE:
      return '可用'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_FAILED:
      return '不可用'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_RELEASED:
      return '停用'
    case ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_EXPIRED:
      return '过期'
    default:
      return '未知'
  }
}

function statusClass(status: ProxyDynamicLeaseStatus) {
  if (status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE) {
    return 'badge-primary'
  }
  if (status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_FAILED) {
    return 'badge-error'
  }
  return 'badge-ghost'
}

function dynamicProviderText(lease: ProxyDynamicLease) {
  return (
    lease.session?.labels?.dynamic_provider_display_name ||
    lease.session?.labels?.dynamic_provider_id ||
    lease.selection_plan?.selected_endpoint?.dynamic_provider_id ||
    ''
  )
}

function routeText(lease: ProxyDynamicLease) {
  return (
    lease.session?.labels?.route_text ||
    lease.egress?.labels?.route_text ||
    lease.selection_plan?.selection_reasons?.[0] ||
    ''
  )
}

function exitIPText(lease: ProxyDynamicLease) {
  return lease.session?.labels?.exit_ip || lease.egress?.labels?.exit_ip || ''
}
</script>

<template>
  <section
    class="flex min-h-0 flex-col gap-4 rounded-xl border border-base-content/8 bg-base-200/60 p-4"
  >
    <div class="flex items-center justify-between gap-3">
      <h2 class="text-base font-semibold">动态租约</h2>
      <span class="badge badge-primary badge-sm">
        {{ runtime.activeLeases.value.length }}
      </span>
    </div>

    <div
      v-if="runtime.activeLeases.value.length === 0"
      class="rounded-lg border border-dashed border-base-content/15 p-6 text-center text-sm opacity-60"
    >
      暂无活跃租约
    </div>
    <div v-else class="flex min-h-0 flex-col gap-2 overflow-y-auto">
      <div
        v-for="lease in runtime.activeLeases.value"
        :key="lease.lease_id"
        class="rounded-lg bg-base-100/70 p-3 text-sm"
      >
        <div class="min-w-0">
          <div class="truncate font-medium">{{ leaseTitle(lease) }}</div>
          <div class="truncate text-xs opacity-60">
            {{ endpointText(lease) }}
          </div>
          <div
            v-if="routeText(lease)"
            class="mt-1 truncate text-xs opacity-70"
          >
            {{ routeText(lease) }}
          </div>
          <div
            v-if="exitIPText(lease)"
            class="mt-1 truncate text-xs opacity-70"
          >
            出口 IP {{ exitIPText(lease) }}
          </div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-sm" :class="statusClass(lease.status)">
              {{ statusText(lease.status) }}
            </span>
            <span class="badge badge-ghost badge-sm">
              更新 {{ updatedText(lease) }}
            </span>
            <span v-if="expiresText(lease)" class="badge badge-ghost badge-sm">
              {{ expiresText(lease) }}
            </span>
            <span
              v-if="lease.session?.provider_id"
              class="badge badge-ghost badge-sm"
            >
              {{ runtime.providerName(lease.session.provider_id) }}
            </span>
            <span
              v-if="dynamicProviderText(lease)"
              class="badge badge-ghost badge-sm"
            >
              {{ dynamicProviderText(lease) }}
            </span>
            <span v-if="exitIPText(lease)" class="badge badge-ghost badge-sm">
              IP {{ exitIPText(lease) }}
            </span>
          </div>
          <div class="mt-2 flex items-center justify-between gap-2">
            <div v-if="lease.error_message" class="truncate text-xs text-error">
              {{ lease.error_message }}
            </div>
            <div v-else></div>
            <Button
              class="btn-ghost btn-xs btn-square"
              :disabled="runtime.busyLeaseID.value === lease.lease_id"
              title="释放租约"
              @click="runtime.release(lease)"
            >
              <IconTrash :size="14" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
