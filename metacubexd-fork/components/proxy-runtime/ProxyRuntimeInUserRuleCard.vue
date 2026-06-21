<script setup lang="ts">
import type { ProxyGatewayInUserRulesState } from '~/composables/useProxyGatewayInUserRules'
import type {
  EgressProfileSettings,
  ProxyIngressRuleSettings,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import { IconKey, IconPencil, IconRoute, IconTrash } from '@tabler/icons-vue'

defineProps<{
  expanded: boolean
  index: number
  profile?: EgressProfileSettings
  rule: ProxyIngressRuleSettings
  runtime: ProxyGatewayInUserRulesState
}>()

defineEmits<{
  collapse: [value: boolean]
  edit: [rule: ProxyIngressRuleSettings]
}>()
</script>

<template>
  <Collapse
    class="animate-fade-slide-in w-full"
    :is-open="expanded"
    :style="{ animationDelay: `${index * 40}ms` }"
    @collapse="$emit('collapse', $event)"
  >
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">
            {{ rule.display_name || rule.username }}
          </h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">
              <IconKey :size="12" />
              {{ rule.username }}
            </span>
            <span class="badge badge-ghost badge-sm">
              <IconRoute :size="12" />
              {{ profile ? runtime.exitText(profile) : '出口' }}
            </span>
          </div>
        </div>

        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <button
            aria-label="编辑"
            class="btn btn-ghost btn-sm btn-square"
            title="编辑"
            type="button"
            @click="$emit('edit', rule)"
          >
            <IconPencil :size="16" />
          </button>
          <button
            aria-label="删除"
            class="btn btn-ghost btn-sm btn-square text-error"
            :disabled="runtime.saving.value"
            title="删除"
            type="button"
            @click="runtime.deleteRule(rule)"
          >
            <IconTrash :size="16" />
          </button>
        </div>
      </div>
    </template>

    <div class="col-span-full grid gap-3 text-sm sm:grid-cols-2">
      <div class="min-w-0">
        <div class="mb-1 text-xs opacity-60">线路</div>
        <div
          class="truncate font-medium"
          :title="profile ? runtime.lineText(profile) : ''"
        >
          {{ profile ? runtime.lineText(profile) : '未配置' }}
        </div>
      </div>
      <div class="min-w-0">
        <div class="mb-1 text-xs opacity-60">出口</div>
        <div
          class="truncate font-medium"
          :title="profile ? runtime.exitText(profile) : ''"
        >
          {{ profile ? runtime.exitText(profile) : '未配置' }}
        </div>
      </div>
    </div>
  </Collapse>
</template>
