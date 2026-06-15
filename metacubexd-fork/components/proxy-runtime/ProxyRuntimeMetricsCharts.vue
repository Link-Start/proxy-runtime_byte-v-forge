<script setup lang="ts">
import type { ProxyRuntimeMetricRow } from '~/composables/proxyRuntimeMetricsRows'
import {
  proxyRuntimeAverageChartOptions,
  proxyRuntimeCountChartOptions,
  proxyRuntimeSlowRatioChartOptions,
  proxyRuntimeStatusChartOptions,
} from '~/composables/proxyRuntimeMetricsCharts'
import { getChartThemeColors } from '~/utils'

const props = defineProps<{
  loading: boolean
  rows: ProxyRuntimeMetricRow[]
}>()

const configStore = useConfigStore()
const themeColors = computed(() => {
  void configStore.curTheme
  return getChartThemeColors()
})

const charts = computed(() => [
  proxyRuntimeStatusChartOptions(props.rows, themeColors.value),
  proxyRuntimeCountChartOptions(props.rows, themeColors.value),
  proxyRuntimeAverageChartOptions(props.rows, themeColors.value),
  proxyRuntimeSlowRatioChartOptions(props.rows, themeColors.value),
])
</script>

<template>
  <div class="grid min-w-0 grid-cols-1 gap-4 lg:grid-cols-2">
    <div
      v-for="(chart, index) in charts"
      :key="index"
      class="animate-fade-slide-in h-72 min-w-0 overflow-hidden rounded-xl border border-base-content/10 bg-base-200 p-2 lg:h-80"
      :style="{ animationDelay: `${index * 45}ms` }"
    >
      <HighchartsAutoSize
        :is-loading="loading && rows.length === 0"
        :options="chart"
      />
    </div>
  </div>
</template>
