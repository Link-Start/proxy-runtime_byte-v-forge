<script setup lang="ts">
import type { ProxyGatewayMetricRow } from '~/composables/proxyGatewayMetricsRows'
import {
  formatProxyGatewayMetricCount,
  formatProxyGatewayMetricPercent,
  formatProxyGatewayMetricSeconds,
  proxyGatewayMetricOperationLabel,
  proxyGatewayMetricStatusClass,
  proxyGatewayMetricStatusLabel,
} from '~/composables/proxyGatewayMetricsRows'

defineProps<{
  rows: ProxyGatewayMetricRow[]
  slowThresholdLabel: string
}>()

function slowRatio(row: ProxyGatewayMetricRow) {
  return row.count > 0 ? row.slowCount / row.count : 0
}

function rowClass(row: ProxyGatewayMetricRow) {
  if (row.status !== 'success') return 'bg-error/8 hover:bg-error/12'
  if (row.slowCount > 0) return 'bg-warning/8 hover:bg-warning/12'
  return 'hover:bg-base-content/5'
}
</script>

<template>
  <section
    class="animate-fade-slide-in overflow-hidden rounded-xl border border-base-content/10 bg-base-200"
  >
    <div
      class="flex flex-wrap items-center justify-between gap-2 border-b border-base-content/10 bg-base-300/30 px-4 py-3"
    >
      <div>
        <h3 class="font-semibold tracking-tight text-base-content">操作明细</h3>
        <p class="mt-1 text-xs opacity-60">
          按 operation/status 汇总，慢路径阈值 {{ slowThresholdLabel }}
        </p>
      </div>
      <span class="badge badge-ghost badge-sm">
        {{ formatProxyGatewayMetricCount(rows.length) }} rows
      </span>
    </div>

    <div class="overflow-x-auto">
      <table class="table table-sm w-full border-collapse whitespace-nowrap">
        <thead>
          <tr class="sticky top-0 z-10 bg-base-200">
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase opacity-70">
              Operation
            </th>
            <th class="px-4 py-3 text-left text-xs font-semibold uppercase opacity-70">
              Status
            </th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase opacity-70">
              Count
            </th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase opacity-70">
              Slow
            </th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase opacity-70">
              Slow %
            </th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase opacity-70">
              Avg
            </th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase opacity-70">
              Total
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in rows"
            :key="`${row.operation}:${row.status}`"
            class="transition-colors duration-150"
            :class="rowClass(row)"
          >
            <td class="max-w-[20rem] px-4 py-3">
              <div class="truncate font-medium">
                {{ proxyGatewayMetricOperationLabel(row.operation) }}
              </div>
              <div class="truncate font-mono text-xs opacity-55">
                {{ row.operation }}
              </div>
            </td>
            <td class="px-4 py-3">
              <span class="badge badge-sm" :class="proxyGatewayMetricStatusClass(row.status)">
                {{ proxyGatewayMetricStatusLabel(row.status) }}
              </span>
            </td>
            <td class="px-4 py-3 text-right font-mono">
              {{ formatProxyGatewayMetricCount(row.count) }}
            </td>
            <td class="px-4 py-3 text-right font-mono">
              {{ formatProxyGatewayMetricCount(row.slowCount) }}
            </td>
            <td class="px-4 py-3 text-right font-mono">
              {{ formatProxyGatewayMetricPercent(slowRatio(row)) }}
            </td>
            <td class="px-4 py-3 text-right font-mono">
              {{ formatProxyGatewayMetricSeconds(row.averageSeconds) }}
            </td>
            <td class="px-4 py-3 text-right font-mono">
              {{ formatProxyGatewayMetricSeconds(row.durationSeconds) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
