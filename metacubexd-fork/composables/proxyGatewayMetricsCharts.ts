import type Highcharts from 'highcharts'
import type { ProxyGatewayMetricRow } from '~/composables/proxyGatewayMetricsRows'
import {
  formatProxyGatewayMetricCount,
  formatProxyGatewayMetricPercent,
  formatProxyGatewayMetricSeconds,
  proxyGatewayMetricOperationLabel,
  proxyGatewayMetricStatusLabel,
} from '~/composables/proxyGatewayMetricsRows'

interface ProxyGatewayChartTheme {
  backgroundColor: string
  gridLineColor: string
  lineColor: string
  seriesColors: string[]
  textColor: string
  textColorHover: string
}

export function proxyGatewayStatusChartOptions(
  rows: ProxyGatewayMetricRow[],
  theme: ProxyGatewayChartTheme,
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
        return `${this.name}: <b>${formatProxyGatewayMetricCount(this.y || 0)}</b> (${percent.toFixed(1)}%)`
      },
    },
    plotOptions: { pie: { dataLabels: { enabled: false }, showInLegend: true } },
    legend: chartLegend(theme),
    series: [{ type: 'pie', name: '状态', data: statusData(rows, theme) }],
  }
}

export function proxyGatewayCountChartOptions(
  rows: ProxyGatewayMetricRow[],
  theme: ProxyGatewayChartTheme,
): Highcharts.Options {
  const ranked = rankRows(rows, (row) => row.count)
  return barChartOptions({
    data: ranked.map((row) => row.count),
    rows: ranked,
    theme,
    title: '操作量排行',
    tooltipValue: (value) => `次数: ${formatProxyGatewayMetricCount(value)}`,
    valueLabel: formatProxyGatewayMetricCount,
  })
}

export function proxyGatewayAverageChartOptions(
  rows: ProxyGatewayMetricRow[],
  theme: ProxyGatewayChartTheme,
): Highcharts.Options {
  const ranked = rankRows(rows, (row) => row.averageSeconds)
  return barChartOptions({
    data: ranked.map((row) => row.averageSeconds),
    rows: ranked,
    theme,
    title: '平均耗时排行',
    tooltipValue: (value) => `平均: ${formatProxyGatewayMetricSeconds(value)}`,
    valueLabel: formatProxyGatewayMetricSeconds,
  })
}

export function proxyGatewaySlowRatioChartOptions(
  rows: ProxyGatewayMetricRow[],
  theme: ProxyGatewayChartTheme,
): Highcharts.Options {
  const ranked = rankRows(rows, slowRatio)
  return barChartOptions({
    data: ranked.map(slowRatio),
    rows: ranked,
    theme,
    title: '慢路径占比排行',
    tooltipValue: (value) => `慢路径: ${formatProxyGatewayMetricPercent(value)}`,
    valueLabel: formatProxyGatewayMetricPercent,
  })
}

function barChartOptions(opts: {
  data: number[]
  rows: ProxyGatewayMetricRow[]
  theme: ProxyGatewayChartTheme
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
      categories: opts.rows.map((row) => proxyGatewayMetricOperationLabel(row.operation)),
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
        const status = row ? proxyGatewayMetricStatusLabel(row.status) : '-'
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

function statusData(rows: ProxyGatewayMetricRow[], theme: ProxyGatewayChartTheme) {
  const totals = new Map<string, number>()
  rows.forEach((row) => {
    totals.set(row.status, (totals.get(row.status) || 0) + row.count)
  })
  return [...totals.entries()].map(([status, count], index) => ({
    name: proxyGatewayMetricStatusLabel(status),
    y: count,
    color: theme.seriesColors[index % theme.seriesColors.length],
  }))
}

function rankRows(rows: ProxyGatewayMetricRow[], value: (row: ProxyGatewayMetricRow) => number) {
  return [...rows].sort((a, b) => value(b) - value(a)).slice(0, 8)
}

function slowRatio(row: ProxyGatewayMetricRow) {
  return row.count > 0 ? row.slowCount / row.count : 0
}

function chartBase(type: 'bar' | 'pie', theme: ProxyGatewayChartTheme): Highcharts.ChartOptions {
  return { type, backgroundColor: theme.backgroundColor, animation: false }
}

function chartTitle(text: string, theme: ProxyGatewayChartTheme) {
  return { text, style: { color: theme.textColor } }
}

function chartLegend(theme: ProxyGatewayChartTheme) {
  return {
    itemStyle: { color: theme.textColor },
    itemHoverStyle: { color: theme.textColorHover },
  }
}
