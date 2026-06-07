<script setup lang="ts">
import type { ProxyRuntimePlaygroundChecksState } from '~/composables/useProxyRuntimePlaygroundChecks'
import type { ProxyIPFraudCheck } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconMapPin, IconSearch, IconShieldCheck } from '@tabler/icons-vue'

defineProps<{
  canRun: boolean
  state: ProxyRuntimePlaygroundChecksState
}>()

const expanded = ref(false)

function scoreText(value?: number) {
  return typeof value === 'number' ? value.toFixed(1) : '-'
}

function fraudScoreText(value?: ProxyIPFraudCheck) {
  if (!value) return '-'
  return scoreText(value.risk_score || 0)
}
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">出口检测</h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">{{ state.exitIP.value?.ip || '未检测' }}</span>
            <span v-if="state.fraud.value" class="badge badge-primary badge-sm">
              score {{ fraudScoreText(state.fraud.value) }}
            </span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <Button
            class="btn-primary btn-sm btn-square"
            :disabled="state.busy.value || !canRun"
            title="检测出口"
            @click="state.run"
          >
            <IconSearch :size="16" :class="{ 'animate-spin': state.busy.value }" />
          </Button>
        </div>
      </div>
    </template>

    <div class="col-span-full flex flex-col gap-2">
      <p v-if="state.error.value" class="alert alert-warning py-2 text-xs">
        {{ state.error.value }}
      </p>

      <div class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <div class="flex items-center gap-2 text-sm">
          <IconSearch :size="16" />
          <span>出口 IP</span>
        </div>
        <span class="truncate font-mono text-xs">{{ state.exitIP.value?.ip || '-' }}</span>
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <div class="flex items-center gap-2 text-sm">
          <IconMapPin :size="12" />
          <span>Geo</span>
        </div>
        <div class="flex flex-wrap gap-1 text-xs">
          <span class="badge badge-ghost badge-sm">{{ state.geo.value?.country_code || '-' }}</span>
          <span class="badge badge-ghost badge-sm">{{ state.geo.value?.region || '-' }}</span>
          <span class="badge badge-ghost badge-sm">{{ state.geo.value?.city || '-' }}</span>
        </div>
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <div class="flex items-center gap-2 text-sm">
          <IconShieldCheck :size="12" />
          <span>Fraud</span>
        </div>
        <div class="flex flex-wrap gap-1 text-xs">
          <span
            v-if="state.fraud.value"
            class="badge badge-sm"
            :class="riskBadgeClass(fraudRiskLabel(state.fraud.value?.risk_level))"
          >
            score {{ fraudScoreText(state.fraud.value) }}
          </span>
          <span v-else class="badge badge-ghost badge-sm">-</span>
          <span v-if="state.fraud.value?.provider_display_name" class="badge badge-ghost badge-sm">
            {{ state.fraud.value.provider_display_name }}
          </span>
          <span
            v-for="signal in state.fraud.value?.risk_signals || []"
            :key="signal"
            class="badge badge-outline badge-sm"
          >
            {{ signal.replace('PROXY_IP_FRAUD_SIGNAL_', '') }}
          </span>
        </div>
      </div>
    </div>
  </Collapse>
</template>
