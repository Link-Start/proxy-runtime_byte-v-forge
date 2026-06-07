<script setup lang="ts">
import type { ProxyIngressRuleSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { ProxyRuntimeInUserRulesState } from '~/composables/useProxyRuntimeInUserRules'
import { IconKey, IconPencil, IconPlus, IconRoute, IconTrash } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeInUserRulesState }>()
const modal = ref<{ open: () => void; close: () => void }>()
const expandedRules = reactive<Record<string, boolean>>({})

function addRule() {
  props.runtime.resetForm()
  modal.value?.open()
}

function editRule(rule: ProxyIngressRuleSettings) {
  props.runtime.editRule(rule)
  modal.value?.open()
}
</script>

<template>
  <section class="mx-auto flex w-full max-w-5xl min-h-0 flex-col gap-4 px-2 py-2">
    <div v-if="runtime.error.value" class="alert alert-error text-sm">
      {{ runtime.error.value }}
    </div>

    <div class="flex items-center justify-between gap-3">
      <h2 class="flex items-center gap-2 text-base font-semibold">
        <IconKey :size="18" />
        入口用户
      </h2>
      <div class="flex items-center gap-2">
        <span class="badge badge-primary badge-sm">{{ runtime.ruleCount.value }}</span>
        <button
          aria-label="添加入口用户"
          class="btn btn-primary btn-sm btn-square"
          title="添加入口用户"
          type="button"
          @click="addRule"
        >
          <IconPlus :size="16" />
        </button>
      </div>
    </div>

    <div
      v-if="runtime.rows.value.length === 0"
      class="rounded-lg border border-dashed border-base-content/15 p-6 text-center text-sm opacity-60"
    >
      暂无入口用户
    </div>
    <div v-else class="grid gap-3">
      <Collapse
        v-for="{ rule, profile } in runtime.rows.value"
        :key="rule.rule_id"
        :is-open="!!expandedRules[rule.rule_id]"
        @collapse="expandedRules[rule.rule_id] = $event"
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
        </template>

        <div class="col-span-full grid gap-3 text-sm sm:grid-cols-2">
          <div class="min-w-0">
            <div class="mb-1 text-xs opacity-60">线路</div>
            <div class="truncate font-medium" :title="profile ? runtime.lineText(profile) : ''">
              {{ profile ? runtime.lineText(profile) : '未配置' }}
            </div>
          </div>
          <div class="min-w-0">
            <div class="mb-1 text-xs opacity-60">出口</div>
            <div class="truncate font-medium" :title="profile ? runtime.exitText(profile) : ''">
              {{ profile ? runtime.exitText(profile) : '未配置' }}
            </div>
          </div>
        </div>
      </Collapse>
    </div>

    <ProxyRuntimeInUserRuleModal ref="modal" :runtime="runtime" />
  </section>
</template>
