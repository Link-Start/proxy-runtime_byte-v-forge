<script setup lang="ts">
import { IconRefresh } from '@tabler/icons-vue'

const runtimeStatus = useProxyRuntimeStatus()
const {
  canSave,
  checks,
  copied,
  credentials,
  copyText,
  curlCommand,
  dynamicExit,
  gatewayHost,
  gatewayPort,
  leases,
  proxyAuthority,
  regeneratePassword,
  refresh,
  refreshing,
  runtime,
  save,
} = useProxyRuntimePlayground()
</script>

<template>
  <main class="flex h-full min-h-0 flex-col gap-3">
    <div class="animate-fade-slide-in flex shrink-0 items-center justify-end gap-2">
      <ProxyRuntimeStatusBadge :state="runtimeStatus" />
      <Button class="flex h-9 w-9 items-center justify-center rounded-[0.625rem] border border-base-content/10 bg-base-200/80 transition-all duration-200 hover:border-primary/30 hover:bg-primary/15 hover:text-primary disabled:cursor-not-allowed disabled:opacity-60" :disabled="refreshing" title="刷新" @click="refresh">
        <IconRefresh :size="18" :class="{ 'animate-spin': refreshing }" />
      </Button>
    </div>

    <p v-if="runtime.error.value" class="alert alert-error py-2 text-sm">{{ runtime.error.value }}</p>

    <div class="min-h-0 flex-1 overflow-y-auto">
      <ProxiesRenderWrapper>
        <template #even>
          <ProxyRuntimePlaygroundUsage
            :copied="copied"
            :copy-text="copyText"
            :credentials="credentials"
            :curl-command="curlCommand"
            :gateway-host="gatewayHost"
            :gateway-port="gatewayPort"
            :proxy-authority="proxyAuthority"
            :username="runtime.form.username"
          />
          <ProxyRuntimePlaygroundChecks
            :can-run="canSave"
            :copied="copied"
            :copy-text="copyText"
            :state="checks"
          />
          <ProxyRuntimePlaygroundLeases :dynamic-exit="dynamicExit" :proxy-authority="proxyAuthority" :state="leases" />
        </template>
        <template #odd>
          <ProxyRuntimePlaygroundConfig
            :can-save="canSave"
            :regenerate-password="regeneratePassword"
            :runtime="runtime"
            :save="save"
          />
        </template>
        <template #default>
          <ProxyRuntimePlaygroundUsage
            :copied="copied"
            :copy-text="copyText"
            :credentials="credentials"
            :curl-command="curlCommand"
            :gateway-host="gatewayHost"
            :gateway-port="gatewayPort"
            :proxy-authority="proxyAuthority"
            :username="runtime.form.username"
          />
          <ProxyRuntimePlaygroundConfig
            :can-save="canSave"
            :regenerate-password="regeneratePassword"
            :runtime="runtime"
            :save="save"
          />
          <ProxyRuntimePlaygroundChecks
            :can-run="canSave"
            :copied="copied"
            :copy-text="copyText"
            :state="checks"
          />
          <ProxyRuntimePlaygroundLeases :dynamic-exit="dynamicExit" :proxy-authority="proxyAuthority" :state="leases" />
        </template>
      </ProxiesRenderWrapper>
    </div>
  </main>
</template>
