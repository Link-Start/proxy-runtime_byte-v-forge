<script setup lang="ts">
import type { ProxyRuntimeSourcesState } from '~/composables/useProxyRuntimeSources'
import type { ProxySourceDescriptor } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconLink, IconPlug, IconPlus } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeSourcesState }>()

const egressProfiles = useProxyRuntimeEgressProfiles()
const subscriptionModal = ref<{ open: () => void; close: () => void }>()
const fixedModal = ref<{ open: () => void; close: () => void }>()
const deleteModal = ref<{ open: () => void; close: () => void }>()
const deleteTarget = ref<ProxySourceDescriptor | null>(null)

function addSubscription() {
  props.runtime.resetSubscriptionForm()
  subscriptionModal.value?.open()
}

function editSubscription(source: ProxySourceDescriptor) {
  props.runtime.editSubscription(source)
  subscriptionModal.value?.open()
}

function addFixed() {
  props.runtime.resetFixedForm()
  fixedModal.value?.open()
}

function editFixed(source: ProxySourceDescriptor) {
  props.runtime.editFixed(source)
  fixedModal.value?.open()
}

function requestDelete(source: ProxySourceDescriptor) {
  deleteTarget.value = source
  deleteModal.value?.open()
}

onMounted(() => {
  egressProfiles.load()
})
</script>

<template>
  <div class="mx-auto flex w-full max-w-5xl min-h-0 flex-col gap-6 px-2 py-2">
    <div v-if="runtime.error.value" class="alert alert-error text-sm">
      {{ runtime.error.value }}
    </div>
    <div v-if="egressProfiles.error.value" class="alert alert-error text-sm">
      {{ egressProfiles.error.value }}
    </div>

    <ProxyRuntimeEgressProfiles :runtime="egressProfiles" />

    <section class="flex min-h-0 flex-col gap-3">
      <div class="flex items-center justify-between gap-3 border-b border-base-content/8 pb-2">
        <h2 class="flex items-center gap-2 text-base font-semibold">
          <IconLink :size="18" />
          订阅源
        </h2>
        <button
          aria-label="添加订阅"
          class="btn btn-primary btn-sm btn-square"
          title="添加订阅"
          type="button"
          @click="addSubscription"
        >
          <IconPlus :size="16" />
        </button>
      </div>
      <ProxyRuntimeSourceList
        empty-text="暂无订阅源"
        :sources="runtime.subscriptions.value"
        @delete-source="requestDelete"
        @edit-source="editSubscription"
      />
    </section>

    <section class="flex min-h-0 flex-col gap-3">
      <div class="flex items-center justify-between gap-3 border-b border-base-content/8 pb-2">
        <h2 class="flex items-center gap-2 text-base font-semibold">
          <IconPlug :size="18" />
          固定代理
        </h2>
        <button
          aria-label="添加固定代理"
          class="btn btn-primary btn-sm btn-square"
          title="添加固定代理"
          type="button"
          @click="addFixed"
        >
          <IconPlus :size="16" />
        </button>
      </div>
      <ProxyRuntimeSourceList
        empty-text="暂无固定代理"
        :sources="runtime.fixedSources.value"
        @delete-source="requestDelete"
        @edit-source="editFixed"
      />
    </section>

    <ProxyRuntimeSubscriptionModal
      ref="subscriptionModal"
      :runtime="runtime"
    />
    <ProxyRuntimeFixedSourceModal ref="fixedModal" :runtime="runtime" />
    <ProxyRuntimeDeleteSourceModal
      ref="deleteModal"
      :runtime="runtime"
      :source="deleteTarget"
    />
  </div>
</template>
