<script setup lang="ts">
import type Highcharts from 'highcharts'
import type { ProxyRuntimeMetricRow } from '~/composables/proxyRuntimeMetricsRows'
import {
  formatProxyRuntimeMetricCount,
  formatProxyRuntimeMetricSeconds,
  proxyRuntimeMetricOperationLabel,
  proxyRuntimeMetricStatusLabel,
} from '~/composables/proxyRuntimeMetricsRows'
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

const statusChartOptions = computed<Highcharts.Options>(() => {
  const data = statusData()
  return {
    chart: chartBase('pie'),
    credits: { enabled: false },
    accessibility: { enabled: false },
    title: chartTitle('状态分布'),
    tooltip: {
      pointFormatter() {
        const percent =
          (this as Highcharts.Point & { percentage?: number }).percentage || 0
        return `${this.name}: <b>${formatProxyRuntimeMetricCount(this.y || 0)}</b> (${percent.toFixed(1)}%)`
      },
    },
    plotOptions: { pie: { dataLabels: { enabled: false }, showInLegend: true } },
    legend: chartLegend(),
    series: [{ type: 'pie', name: '状态', data }],
  }
})

const operationChartOptions = computed<Highcharts.Options>(() => {
  const rows = [...props.rows].sort((a, b) => b.count - a.count).slice(0, 8)
  const categories = rows.map((row) =>
    proxyRuntimeMetricOperationLabel(row.operation),
  )
  return {
    chart: chartBase('bar'),
    credits: { enabled: false },
    accessibility: { enabled: false },
    title: chartTitle('操作量排行'),
    xAxis: {
      categories,
      labels: { style: { color: themeColors.value.textColor } },
      lineColor: themeColors.value.lineColor,
    },
    yAxis: {
      min: 0,
      title: { text: undefined },
      labels: {
        style: { color: themeColors.value.textColor },
        formatter() {
          return formatProxyRuntimeMetricCount(this.value as number)
        },
      },
      gridLineColor: themeColors.value.gridLineColor,
    },
    tooltip: {
      formatter() {
        const row = rows[this.x as number]
        const avg = row ? formatProxyRuntimeMetricSeconds(row.averageSeconds) : '-'
        return `<b>${this.key}</b><br/>次数: ${formatProxyRuntimeMetricCount(this.y || 0)}<br/>平均: ${avg}`
      },
    },
    legend: { enabled: false },
    plotOptions: { bar: { dataLabels: { enabled: false }, animation: false } },
    series: [
      {
        type: 'bar',
        name: '次数',
        data: rows.map((row) => row.count),
        color: themeColors.value.seriesColors[0],
      },
    ],
  }
})

function statusData() {
  const totals = new Map<string, number>()
  props.rows.forEach((row) => {
    totals.set(row.status, (totals.get(row.status) || 0) + row.count)
  })
  return [...totals.entries()].map(([status, count], index) => ({
    name: proxyRuntimeMetricStatusLabel(status),
    y: count,
    color: themeColors.value.seriesColors[index % themeColors.value.seriesColors.length],
  }))
}

function chartBase(type: 'bar' | 'pie'): Highcharts.ChartOptions {
  return {
    type,
    backgroundColor: themeColors.value.backgroundColor,
    animation: false,
  }
}

function chartTitle(text: string) {
  return { text, style: { color: themeColors.value.textColor } }
}

function chartLegend() {
  return {
    itemStyle: { color: themeColors.value.textColor },
    itemHoverStyle: { color: themeColors.value.textColorHover },
  }
}
</script>

<template>
  <div class="grid min-w-0 grid-cols-1 gap-4 lg:grid-cols-2">
    <div
      class="animate-fade-slide-in h-72 min-w-0 overflow-hidden rounded-xl border border-base-content/10 bg-base-200 p-2 lg:h-80"
    >
      <HighchartsAutoSize
        :is-loading="loading && rows.length === 0"
        :options="statusChartOptions"
      />
    </div>
    <div
      class="animate-fade-slide-in h-72 min-w-0 overflow-hidden rounded-xl border border-base-content/10 bg-base-200 p-2 lg:h-80"
    >
      <HighchartsAutoSize
        :is-loading="loading && rows.length === 0"
        :options="operationChartOptions"
      />
    </div>
  </div>
</template>
