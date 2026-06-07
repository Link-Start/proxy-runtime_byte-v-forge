<script setup lang="ts">
import type { ProxyRuntimePluginsState } from '~/composables/useProxyRuntimePlugins'
import { IconShieldCheck } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimePluginsState }>()
const expanded = ref(false)

function descriptorName(providerID: string) {
  const descriptor = props.runtime.descriptors.value.find(
    (item) => item.provider_id === providerID,
  )
  return descriptor?.display_name || providerID
}
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">手动检测</h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">
              <IconShieldCheck :size="12" />
              Fraud
            </span>
          </div>
        </div>
      </div>
    </template>

    <div class="col-span-full grid gap-3">
      <div class="grid gap-2 md:grid-cols-[1fr_auto]">
        <input
          v-model.trim="runtime.fraudIP.value"
          class="input input-bordered input-sm"
          placeholder="IP，例如 8.8.8.8"
        />
        <button
          aria-label="检测 IP Fraud"
          class="btn btn-sm btn-square"
          :disabled="runtime.checking.value"
          title="检测 IP Fraud"
          type="button"
          @click="runtime.checkFraud"
        >
          <IconShieldCheck :size="16" />
        </button>
      </div>

      <div v-if="runtime.fraudResult.value" class="flex flex-wrap gap-2 text-xs">
        <span
          class="badge"
          :class="riskBadgeClass(fraudRiskLabel(runtime.fraudResult.value.risk_level))"
        >
          {{ fraudRiskLabel(runtime.fraudResult.value.risk_level) }}
        </span>
        <span class="badge badge-ghost">
          {{ descriptorName(runtime.fraudResult.value.provider_id) }}
        </span>
        <span class="badge badge-ghost">
          score {{ runtime.fraudResult.value.risk_score }}
        </span>
        <span class="badge badge-ghost">
          {{ runtime.fraudResult.value.country_code }}
        </span>
        <span
          v-for="signal in runtime.fraudResult.value.risk_signals"
          :key="signal"
          class="badge badge-outline"
        >
          {{ signal.replace('PROXY_IP_FRAUD_SIGNAL_', '') }}
        </span>
      </div>
    </div>
  </Collapse>
</template>
