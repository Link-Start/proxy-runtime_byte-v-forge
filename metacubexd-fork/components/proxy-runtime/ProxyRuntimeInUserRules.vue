<script setup lang="ts">
import type { ProxyGatewayInUserRulesState } from '~/composables/useProxyGatewayInUserRules'
import type { ProxyIngressRuleSettings } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import { IconKey, IconPlus } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyGatewayInUserRulesState }>()
const modal = ref<{ open: () => void; close: () => void }>()
const expandedRules = reactive<Record<string, boolean>>({})
const evenRows = computed(() =>
  props.runtime.rows.value.filter((_, index) => index % 2 === 0),
)
const oddRows = computed(() =>
  props.runtime.rows.value.filter((_, index) => index % 2 === 1),
)

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
  <section class="flex min-h-0 w-full flex-col gap-3">
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
    <ProxiesRenderWrapper v-else>
      <template #even>
        <ProxyGatewayInUserRuleCard
          v-for="(row, index) in evenRows"
          :key="row.rule.rule_id"
          :expanded="expandedRules[row.rule.rule_id] || false"
          :index="index"
          :profile="row.profile"
          :rule="row.rule"
          :runtime="runtime"
          @collapse="expandedRules[row.rule.rule_id] = $event"
          @edit="editRule"
        />
      </template>
      <template #odd>
        <ProxyGatewayInUserRuleCard
          v-for="(row, index) in oddRows"
          :key="row.rule.rule_id"
          :expanded="expandedRules[row.rule.rule_id] || false"
          :index="index"
          :profile="row.profile"
          :rule="row.rule"
          :runtime="runtime"
          @collapse="expandedRules[row.rule.rule_id] = $event"
          @edit="editRule"
        />
      </template>
      <template #default>
        <ProxyGatewayInUserRuleCard
          v-for="(row, index) in runtime.rows.value"
          :key="row.rule.rule_id"
          :expanded="expandedRules[row.rule.rule_id] || false"
          :index="index"
          :profile="row.profile"
          :rule="row.rule"
          :runtime="runtime"
          @collapse="expandedRules[row.rule.rule_id] = $event"
          @edit="editRule"
        />
      </template>
    </ProxiesRenderWrapper>

    <ProxyGatewayInUserRuleModal ref="modal" :runtime="runtime" />
  </section>
</template>
