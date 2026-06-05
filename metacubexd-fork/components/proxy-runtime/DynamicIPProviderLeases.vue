<script setup lang="ts">
import type { ProxyRuntimeDynamicLeasesState } from '~/composables/useProxyRuntimeDynamicLeases'
import type { ProxyDynamicLease } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconLogout } from '@tabler/icons-vue'

defineProps<{ runtime: ProxyRuntimeDynamicLeasesState }>()

function leaseTitle(lease: ProxyDynamicLease) {
  return lease.purpose || lease.lease_id
}

function endpointText(lease: ProxyDynamicLease) {
  if (!lease.egress) return '-'
  return `${lease.egress.host}:${lease.egress.port}`
}

function expiresText(lease: ProxyDynamicLease) {
  if (!lease.expires_at) return '-'
  return new Date(lease.expires_at).toLocaleString()
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
        class="grid gap-2 rounded-lg bg-base-100/70 p-3 text-sm sm:grid-cols-[1fr_auto]"
      >
        <div class="min-w-0">
          <div class="truncate font-medium">{{ leaseTitle(lease) }}</div>
          <div class="truncate text-xs opacity-60">
            {{ lease.account_id }} / {{ endpointText(lease) }}
          </div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-sm">到期 {{ expiresText(lease) }}</span>
            <span
              v-if="lease.session?.provider_id"
              class="badge badge-ghost badge-sm"
            >
              {{ runtime.providerName(lease.session.provider_id) }}
            </span>
          </div>
        </div>
        <button
          aria-label="释放"
          class="btn btn-ghost btn-sm btn-square text-error"
          :disabled="runtime.saving.value"
          title="释放"
          type="button"
          @click="runtime.releaseLease(lease)"
        >
          <IconLogout :size="16" />
        </button>
      </div>
    </div>
  </section>
</template>
