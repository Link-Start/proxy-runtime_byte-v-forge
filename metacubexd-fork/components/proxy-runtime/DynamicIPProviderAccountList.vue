<script setup lang="ts">
import type { ProxyProviderAccount } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import { ProxyProviderAccountStatus } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import { IconPencil, IconTrash } from '@tabler/icons-vue'

defineProps<{ accounts: ProxyProviderAccount[] }>()

defineEmits<{
  deleteAccount: [accountID: string]
  editAccount: [account: ProxyProviderAccount]
}>()

function statusText(status: ProxyProviderAccountStatus) {
  return status ===
    ProxyProviderAccountStatus.PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED
    ? '启用'
    : '停用'
}
</script>

<template>
  <section class="min-w-0">
    <div class="divider my-0 text-xs uppercase opacity-40">账号</div>
    <div v-if="accounts.length === 0" class="py-3 text-sm opacity-55">
      暂无账号
    </div>
    <div v-else class="grid gap-2">
      <div
        v-for="account in accounts"
        :key="account.account_id"
        class="grid gap-2 py-2 md:grid-cols-[1fr_auto]"
      >
        <div class="min-w-0 text-sm">
          <div class="truncate font-medium">
            {{ account.display_name || account.account_id }}
          </div>
          <div class="truncate text-xs opacity-60">
            {{ account.username }}
          </div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-sm">{{ statusText(account.status) }}</span>
            <span
              v-if="account.credential_configured"
              class="badge badge-ghost badge-sm"
            >
              已配置凭据
            </span>
          </div>
        </div>
        <div class="flex items-center gap-1">
          <button
            aria-label="编辑账号"
            class="btn btn-ghost btn-sm btn-square"
            title="编辑账号"
            type="button"
            @click="$emit('editAccount', account)"
          >
            <IconPencil :size="16" />
          </button>
          <button
            aria-label="删除账号"
            class="btn btn-ghost btn-sm btn-square text-error"
            title="删除账号"
            type="button"
            @click="$emit('deleteAccount', account.account_id)"
          >
            <IconTrash :size="16" />
          </button>
        </div>
      </div>
    </div>
  </section>
</template>
