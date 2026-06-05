<script setup lang="ts">
import type { ProxyIngressRuleSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { ProxyRuntimeIngressRulesState } from '~/composables/useProxyRuntimeIngressRules'
import { profileLabelByID } from '~/composables/proxyRuntimeIngressRuleHelpers'
import { IconKey, IconPencil, IconPlus, IconTrash } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeIngressRulesState }>()
const modal = ref<{ open: () => void; close: () => void }>()

function addRule() {
  props.runtime.resetForm()
  modal.value?.open()
}

function editRule(rule: ProxyIngressRuleSettings) {
  props.runtime.editRule(rule)
  modal.value?.open()
}

function targetText(rule: ProxyIngressRuleSettings) {
  return profileLabelByID(props.runtime.profiles.value, rule.profile_id)
}
</script>

<template>
  <section class="flex min-h-0 flex-col gap-3">
    <div class="flex items-center justify-between gap-3 border-b border-base-content/8 pb-2">
      <h2 class="flex items-center gap-2 text-base font-semibold">
        <IconKey :size="18" />
        入口规则
      </h2>
      <button
        aria-label="添加入口规则"
        class="btn btn-primary btn-sm btn-square"
        title="添加入口规则"
        type="button"
        @click="addRule"
      >
        <IconPlus :size="16" />
      </button>
    </div>

    <div v-if="runtime.error.value" class="alert alert-error py-2 text-sm">
      {{ runtime.error.value }}
    </div>

    <div v-if="runtime.rules.value.length === 0" class="py-5 text-sm opacity-55">
      暂无入口规则
    </div>
    <div v-else class="divide-y divide-base-content/8">
      <div
        v-for="rule in runtime.rules.value"
        :key="rule.rule_id"
        class="grid gap-2 py-3 md:grid-cols-[1fr_auto]"
      >
        <div class="min-w-0 text-sm">
          <div class="truncate font-medium">
            {{ rule.display_name || rule.username }}
          </div>
          <div class="truncate text-xs opacity-60" :title="targetText(rule)">
            {{ rule.username }} -> {{ targetText(rule) }}
          </div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-sm">{{ rule.enabled ? '启用' : '停用' }}</span>
            <span class="badge badge-ghost badge-sm">
              <IconKey :size="12" />
              {{ rule.username }}
            </span>
          </div>
        </div>
        <div class="flex items-center gap-1">
          <button
            aria-label="编辑"
            class="btn btn-ghost btn-sm btn-square"
            title="编辑"
            type="button"
            @click="editRule(rule)"
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
    </div>

    <ProxyRuntimeIngressRuleModal ref="modal" :runtime="runtime" />
  </section>
</template>
