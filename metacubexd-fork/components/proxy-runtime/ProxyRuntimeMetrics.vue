<script setup lang="ts">
import type { ProxyGatewayMetricsState } from '~/composables/useProxyGatewayMetrics'
import { IconRefresh } from '@tabler/icons-vue'
import {
  formatProxyGatewayMetricCount,
  formatProxyGatewayMetricPercent,
  formatProxyGatewayMetricSeconds,
} from '~/composables/proxyGatewayMetricsRows'

const props = defineProps<{ runtime: ProxyGatewayMetricsState }>()
const runtimeStatus = useProxyGatewayStatus()

const stats = computed(() => {
  const overview = props.runtime.overview.value
  const total = overview.totalCount || 0
  const errorRatio = total > 0 ? overview.errorCount / total : 0
  const slowRatio = total > 0 ? overview.slowCount / total : 0
  return [
    {
      label: '运行操作',
      value: formatProxyGatewayMetricCount(total),
      note: `更新 ${props.runtime.updatedAtLabel.value}`,
      tone: 'primary' as const,
    },
    {
      label: '错误',
      value: formatProxyGatewayMetricCount(overview.errorCount),
      note: formatProxyGatewayMetricPercent(errorRatio),
      tone: overview.errorCount > 0 ? ('error' as const) : ('success' as const),
    },
    {
      label: '慢路径',
      value: formatProxyGatewayMetricCount(overview.slowCount),
      note: `阈值 ${props.runtime.slowThresholdLabel.value} · ${formatProxyGatewayMetricPercent(slowRatio)}`,
      tone: overview.slowCount > 0 ? ('warning' as const) : ('success' as const),
    },
    {
      label: '平均耗时',
      value: formatProxyGatewayMetricSeconds(overview.averageSeconds),
      note: `累计 ${formatProxyGatewayMetricSeconds(overview.durationSeconds)}`,
      tone: 'primary' as const,
    },
  ]
})
</script>

<template>
  <main class="flex min-w-0 flex-col gap-4">
    <div class="animate-fade-slide-in flex shrink-0 items-center justify-between gap-2">
      <div class="flex min-w-0 items-center gap-2">
        <ProxyGatewayStatusBadge :state="runtimeStatus" />
        <span class="truncate text-xs opacity-60">
          {{ runtime.updatedAtLabel.value }}
        </span>
      </div>

      <button
        aria-label="刷新观测指标"
        class="flex h-9 w-9 items-center justify-center rounded-[0.625rem] border border-base-content/10 bg-base-200/80 transition-all duration-200 hover:border-primary/30 hover:bg-primary/15 hover:text-primary disabled:cursor-not-allowed disabled:opacity-60"
        :disabled="runtime.loading.value"
        title="刷新观测指标"
        type="button"
        @click="runtime.load()"
      >
        <IconRefresh :size="18" :class="{ 'animate-spin': runtime.loading.value }" />
      </button>
    </div>

    <p v-if="runtime.error.value" class="alert alert-error py-2 text-sm">
      {{ runtime.error.value }}
    </p>

    <div class="grid shrink-0 gap-3 md:grid-cols-2 xl:grid-cols-4">
      <ProxyGatewayMetricStat
        v-for="stat in stats"
        :key="stat.label"
        :label="stat.label"
        :note="stat.note"
        :tone="stat.tone"
        :value="stat.value"
      />
    </div>

    <ProxyGatewayMetricsCharts
      :loading="runtime.loading.value"
      :rows="runtime.rows.value"
    />

    <div class="min-w-0">
      <div
        v-if="runtime.rows.value.length === 0"
        class="py-8 text-center text-sm opacity-60"
      >
        {{ runtime.loading.value ? '正在读取指标' : '暂无运行指标' }}
      </div>

      <ProxyGatewayMetricsTable
        v-else
        :rows="runtime.rows.value"
        :slow-threshold-label="runtime.slowThresholdLabel.value"
      />
    </div>
  </main>
</template>
