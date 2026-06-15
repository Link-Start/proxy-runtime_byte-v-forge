import type Highcharts from 'highcharts'
import type { ProxyRuntimeMetricRow } from '~/composables/proxyRuntimeMetricsRows'
import {
  formatProxyRuntimeMetricCount,
  formatProxyRuntimeMetricPercent,
  formatProxyRuntimeMetricSeconds,
  proxyRuntimeMetricOperationLabel,
  proxyRuntimeMetricStatusLabel,
} from '~/composables/proxyRuntimeMetricsRows'

interface ProxyRuntimeChartTheme {
  backgroundColor: string
  gridLineColor: string
  lineColor: string
  seriesColors: string[]
  textColor: string
  textColorHover: string
}

export function proxyRuntimeStatusChartOptions(
  rows: ProxyRuntimeMetricRow[],
  theme: ProxyRuntimeChartTheme,
): Highcharts.Options {
  return {
    chart: chartBase('pie', theme),
    credits: { enabled: false },
    accessibility: { enabled: false },
    title: chartTitle('状态分布', theme),
    tooltip: {
      pointFormatter() {
        const percent =
          (this as Highcharts.Point & { percentage?: number }).percentage || 0
        return `${this.name}: <b>${formatProxyRuntimeMetricCount(this.y || 0)}</b> (${percent.toFixed(1)}%)`
      },
    },
    plotOptions: { pie: { dataLabels: { enabled: false }, showInLegend: true } },
    legend: chartLegend(theme),
    series: [{ type: 'pie', name: '状态', data: statusData(rows, theme) }],
  }
}

export function proxyRuntimeCountChartOptions(
  rows: ProxyRuntimeMetricRow[],
  theme: ProxyRuntimeChartTheme,
): Highcharts.Options {
  const ranked = rankRows(rows, (row) => row.count)
  return barChartOptions({
    data: ranked.map((row) => row.count),
    rows: ranked,
    theme,
    title: '操作量排行',
    tooltipValue: (value) => `次数: ${formatProxyRuntimeMetricCount(value)}`,
    valueLabel: formatProxyRuntimeMetricCount,
  })
}

export function proxyRuntimeAverageChartOptions(
  rows: ProxyRuntimeMetricRow[],
  theme: ProxyRuntimeChartTheme,
): Highcharts.Options {
  const ranked = rankRows(rows, (row) => row.averageSeconds)
  return barChartOptions({
    data: ranked.map((row) => row.averageSeconds),
    rows: ranked,
    theme,
    title: '平均耗时排行',
    tooltipValue: (value) => `平均: ${formatProxyRuntimeMetricSeconds(value)}`,
    valueLabel: formatProxyRuntimeMetricSeconds,
  })
}

export function proxyRuntimeSlowRatioChartOptions(
  rows: ProxyRuntimeMetricRow[],
  theme: ProxyRuntimeChartTheme,
): Highcharts.Options {
  const ranked = rankRows(rows, slowRatio)
  return barChartOptions({
    data: ranked.map(slowRatio),
    rows: ranked,
    theme,
    title: '慢路径占比排行',
    tooltipValue: (value) => `慢路径: ${formatProxyRuntimeMetricPercent(value)}`,
    valueLabel: formatProxyRuntimeMetricPercent,
  })
}

function barChartOptions(opts: {
  data: number[]
  rows: ProxyRuntimeMetricRow[]
  theme: ProxyRuntimeChartTheme
  title: string
  tooltipValue: (value: number) => string
  valueLabel: (value: number) => string
}): Highcharts.Options {
  return {
    chart: chartBase('bar', opts.theme),
    credits: { enabled: false },
    accessibility: { enabled: false },
    title: chartTitle(opts.title, opts.theme),
    xAxis: {
      categories: opts.rows.map((row) => proxyRuntimeMetricOperationLabel(row.operation)),
      labels: { style: { color: opts.theme.textColor } },
      lineColor: opts.theme.lineColor,
    },
    yAxis: {
      min: 0,
      title: { text: undefined },
      labels: {
        style: { color: opts.theme.textColor },
        formatter() {
          return opts.valueLabel(this.value as number)
        },
      },
      gridLineColor: opts.theme.gridLineColor,
    },
    tooltip: {
      formatter() {
        const row = opts.rows[this.x as number]
        const status = row ? proxyRuntimeMetricStatusLabel(row.status) : '-'
        return `<b>${this.key}</b><br/>状态: ${status}<br/>${opts.tooltipValue(this.y || 0)}`
      },
    },
    legend: { enabled: false },
    plotOptions: { bar: { dataLabels: { enabled: false }, animation: false } },
    series: [{
      type: 'bar',
      name: opts.title,
      data: opts.data,
      color: opts.theme.seriesColors[0],
    }],
  }
}

function statusData(rows: ProxyRuntimeMetricRow[], theme: ProxyRuntimeChartTheme) {
  const totals = new Map<string, number>()
  rows.forEach((row) => {
    totals.set(row.status, (totals.get(row.status) || 0) + row.count)
  })
  return [...totals.entries()].map(([status, count], index) => ({
    name: proxyRuntimeMetricStatusLabel(status),
    y: count,
    color: theme.seriesColors[index % theme.seriesColors.length],
  }))
}

function rankRows(rows: ProxyRuntimeMetricRow[], value: (row: ProxyRuntimeMetricRow) => number) {
  return [...rows].sort((a, b) => value(b) - value(a)).slice(0, 8)
}

function slowRatio(row: ProxyRuntimeMetricRow) {
  return row.count > 0 ? row.slowCount / row.count : 0
}

function chartBase(type: 'bar' | 'pie', theme: ProxyRuntimeChartTheme): Highcharts.ChartOptions {
  return { type, backgroundColor: theme.backgroundColor, animation: false }
}

function chartTitle(text: string, theme: ProxyRuntimeChartTheme) {
  return { text, style: { color: theme.textColor } }
}

function chartLegend(theme: ProxyRuntimeChartTheme) {
  return {
    itemStyle: { color: theme.textColor },
    itemHoverStyle: { color: theme.textColorHover },
  }
}
