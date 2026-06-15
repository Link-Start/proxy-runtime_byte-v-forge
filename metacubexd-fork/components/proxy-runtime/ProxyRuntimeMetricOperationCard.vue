<script setup lang="ts">
import type { ProxyRuntimeMetricRow } from '~/composables/proxyRuntimeMetricsRows'
import { IconClock, IconGauge, IconRoute, IconTrendingUp } from '@tabler/icons-vue'
import {
  formatProxyRuntimeMetricCount,
  formatProxyRuntimeMetricPercent,
  formatProxyRuntimeMetricSeconds,
  proxyRuntimeMetricOperationLabel,
  proxyRuntimeMetricStatusClass,
  proxyRuntimeMetricStatusLabel,
} from '~/composables/proxyRuntimeMetricsRows'

const props = defineProps<{
  index: number
  row: ProxyRuntimeMetricRow
  slowThresholdLabel: string
}>()

const expanded = ref(false)
const slowRatio = computed(() =>
  props.row.count > 0 ? props.row.slowCount / props.row.count : 0,
)
const errorLike = computed(() => props.row.status !== 'success')
</script>

<template>
  <Collapse
    class="animate-fade-slide-in w-full"
    :style="{ animationDelay: `${index * 35}ms` }"
    :is-open="expanded"
    @collapse="expanded = $event"
  >
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <span class="badge badge-sm" :class="proxyRuntimeMetricStatusClass(row.status)">
              {{ proxyRuntimeMetricStatusLabel(row.status) }}
            </span>
            <h3 class="truncate text-base font-semibold">
              {{ proxyRuntimeMetricOperationLabel(row.operation) }}
            </h3>
          </div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">
              {{ formatProxyRuntimeMetricCount(row.count) }} 次
            </span>
            <span
              class="badge badge-sm"
              :class="errorLike ? 'badge-error badge-outline' : 'badge-ghost'"
            >
              avg {{ formatProxyRuntimeMetricSeconds(row.averageSeconds) }}
            </span>
            <span v-if="row.slowCount > 0" class="badge badge-warning badge-sm">
              slow {{ formatProxyRuntimeMetricCount(row.slowCount) }}
            </span>
          </div>
        </div>
      </div>
    </template>

    <div class="col-span-full flex flex-col gap-2">
      <div
        class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5"
      >
        <div class="flex items-center gap-2 text-sm">
          <IconRoute :size="16" />
          <span>Operation</span>
        </div>
        <span class="truncate font-mono text-xs">{{ row.operation }}</span>
      </div>

      <div
        class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5"
      >
        <div class="flex items-center gap-2 text-sm">
          <IconTrendingUp :size="16" />
          <span>总耗时</span>
        </div>
        <span class="font-mono text-xs">
          {{ formatProxyRuntimeMetricSeconds(row.durationSeconds) }}
        </span>
      </div>

      <div
        class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5"
      >
        <div class="flex items-center gap-2 text-sm">
          <IconClock :size="16" />
          <span>平均耗时</span>
        </div>
        <span class="font-mono text-xs">
          {{ formatProxyRuntimeMetricSeconds(row.averageSeconds) }}
        </span>
      </div>

      <div
        class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5"
      >
        <div class="flex items-center gap-2 text-sm">
          <IconGauge :size="16" />
          <span>慢路径</span>
        </div>
        <div class="flex flex-wrap justify-end gap-1">
          <span class="badge badge-ghost badge-sm">阈值 {{ slowThresholdLabel }}</span>
          <span class="badge badge-warning badge-sm">
            {{ formatProxyRuntimeMetricPercent(slowRatio) }}
          </span>
        </div>
      </div>
    </div>
  </Collapse>
</template>
