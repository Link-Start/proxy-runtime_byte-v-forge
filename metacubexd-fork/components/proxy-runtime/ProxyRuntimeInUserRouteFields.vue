<script setup lang="ts">
import type { ProxyGatewayInUserRulesState } from '~/composables/useProxyGatewayInUserRules'
import { EgressProfileExitKind, EgressProfileLineKind, ProxySessionMode } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

const props = defineProps<{ dense?: boolean, runtime: ProxyGatewayInUserRulesState }>()

const lineKinds = [
  [EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_DIRECT, '直连'],
  [EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE, '指定节点'],
] as const
const exitKinds = [
  [EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT, '使用线路出口'],
  [EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP, '静态 IP'],
  [EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP, '动态 IP'],
] as const
const dynamicSessionModes = [
  [ProxySessionMode.PROXY_SESSION_MODE_ROTATING, '轮转 IP'],
  [ProxySessionMode.PROXY_SESSION_MODE_STICKY, '粘性 IP'],
] as const

const lineUsesMihomoNode = computed(
  () =>
    props.runtime.form.line_kind ===
    EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE,
)
const exitUsesMihomoNode = computed(
  () =>
    props.runtime.form.exit_kind ===
    EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP,
)
const exitUsesDynamicProvider = computed(
  () =>
    props.runtime.form.exit_kind ===
    EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP,
)
const exitUsesStickyDynamicIP = computed(
  () =>
    exitUsesDynamicProvider.value &&
    props.runtime.form.exit_dynamic_session_mode ===
      ProxySessionMode.PROXY_SESSION_MODE_STICKY,
)
const availableExitKinds = computed(() =>
  exitKinds.filter(
    ([value]) =>
      !lineUsesMihomoNode.value ||
      value !== EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP,
  ),
)
const rootClass = computed(() => [
  'grid',
  props.dense ? 'gap-2 [&_.input]:input-sm [&_.select]:select-sm' : 'gap-4',
])
const sectionClass = computed(() =>
  props.dense
    ? 'flex flex-col gap-2'
    : 'rounded-lg border border-base-content/10 p-3',
)
const titleClass = computed(() =>
  props.dense
    ? 'divider my-0 text-xs uppercase opacity-40'
    : 'mb-2 text-sm font-medium',
)
const lineGridClass = computed(() =>
  props.dense
    ? 'grid gap-2 sm:grid-cols-3 xl:grid-cols-4'
    : 'grid gap-2 sm:grid-cols-2',
)
const exitGridClass = computed(() =>
  props.dense
    ? 'grid gap-2 sm:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-6'
    : 'grid gap-2 sm:grid-cols-2',
)

watch(lineUsesMihomoNode, (usesMihomoNode) => {
  if (
    usesMihomoNode &&
    props.runtime.form.exit_kind ===
      EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP
  ) {
    props.runtime.form.exit_kind =
      EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT
    props.runtime.form.exit_resource_id = ''
    props.runtime.form.exit_node_id = ''
  }
})

watch(
  () => props.runtime.form.exit_dynamic_provider_id,
  () => {
    props.runtime.form.exit_dynamic_endpoint_id = ''
  },
)
</script>

<template>
  <div :class="rootClass">
    <section :class="sectionClass">
      <div :class="titleClass">线路</div>
      <div :class="lineGridClass">
        <select v-model="runtime.form.line_kind" class="select-bordered select w-full">
          <option v-for="[value, label] in lineKinds" :key="value" :value="value">
            {{ label }}
          </option>
        </select>
        <ProxyGatewayMihomoOwnerSelect
          v-if="lineUsesMihomoNode"
          v-model="runtime.form.line_resource_id"
          placeholder="线路节点组"
          :owners="runtime.lineSources.value"
          @update:model-value="runtime.form.line_node_id = ''"
        />
        <ProxyGatewayMihomoNodeSelect
          v-if="lineUsesMihomoNode"
          v-model="runtime.form.line_node_id"
          placeholder="线路节点"
          :runtime="runtime"
          :owner-id="runtime.form.line_resource_id"
        />
      </div>
    </section>

    <section :class="sectionClass">
      <div :class="titleClass">出口</div>
      <div :class="exitGridClass">
        <select v-model="runtime.form.exit_kind" class="select-bordered select w-full">
          <option v-for="[value, label] in availableExitKinds" :key="value" :value="value">
            {{ label }}
          </option>
        </select>
        <ProxyGatewayMihomoOwnerSelect
          v-if="exitUsesMihomoNode"
          v-model="runtime.form.exit_resource_id"
          placeholder="静态出口节点组"
          :owners="runtime.staticExitSources.value"
          @update:model-value="runtime.form.exit_node_id = ''"
        />
        <ProxyGatewayMihomoNodeSelect
          v-if="exitUsesMihomoNode"
          v-model="runtime.form.exit_node_id"
          placeholder="静态出口节点"
          :runtime="runtime"
          :owner-id="runtime.form.exit_resource_id"
        />
        <ProxyGatewayDynamicProviderSelect
          v-if="exitUsesDynamicProvider"
          v-model="runtime.form.exit_dynamic_provider_id"
          :providers="runtime.dynamicProviderOptions.value"
        />
        <select
          v-if="exitUsesDynamicProvider"
          v-model="runtime.form.exit_dynamic_session_mode"
          class="select-bordered select w-full"
        >
          <option
            v-for="[value, label] in dynamicSessionModes"
            :key="value"
            :value="value"
          >
            {{ label }}
          </option>
        </select>
        <ProxyGatewayDynamicIPEndpointSelect
          v-if="exitUsesDynamicProvider"
          v-model="runtime.form.exit_dynamic_endpoint_id"
          :dynamic-provider-id="runtime.form.exit_dynamic_provider_id"
          :providers="runtime.dynamicProviderOptions.value"
        />
        <ProxyGatewayDynamicIPGeoSelects
          v-if="exitUsesDynamicProvider"
          :runtime="runtime"
        />
        <input
          v-if="exitUsesStickyDynamicIP"
          v-model.number="runtime.form.exit_dynamic_sticky_minutes"
          class="input-bordered input w-full"
          max="120"
          min="1"
          placeholder="粘性分钟"
          type="number"
        />
      </div>
    </section>
  </div>
</template>
