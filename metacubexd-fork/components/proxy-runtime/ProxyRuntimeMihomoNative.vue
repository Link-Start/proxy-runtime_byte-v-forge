<script setup lang="ts">
import type {
  ProxyRuntimeNativeFixedProxy,
  ProxyRuntimeNativeSubscription,
} from '~/composables/proxyRuntimeNativeConfigTypes'
import type { ProxyRuntimeNativeConfigState } from '~/composables/useProxyRuntimeNativeConfig'
import {
  IconLink,
  IconPencil,
  IconPlus,
  IconServer,
  IconTrash,
} from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeNativeConfigState }>()
const fixedModal = ref<{ open: () => void; close: () => void }>()
const subscriptionModal = ref<{ open: () => void; close: () => void }>()

function addFixedProxy() {
  props.runtime.resetFixedForm()
  fixedModal.value?.open()
}

function editFixedProxy(proxy: ProxyRuntimeNativeFixedProxy) {
  props.runtime.editFixedProxy(proxy)
  fixedModal.value?.open()
}

function addSubscription() {
  props.runtime.resetSubscriptionForm()
  subscriptionModal.value?.open()
}

function editSubscription(subscription: ProxyRuntimeNativeSubscription) {
  props.runtime.editSubscription(subscription)
  subscriptionModal.value?.open()
}
</script>

<template>
  <section class="flex min-h-0 flex-col gap-4 p-2">
    <div class="grid gap-4 xl:grid-cols-2">
      <section class="min-w-0">
        <div class="mb-2 flex items-center justify-between gap-3">
          <h2 class="flex items-center gap-2 text-base font-semibold">
            <IconServer :size="18" />
            固定代理
          </h2>
          <button
            aria-label="添加固定代理"
            class="btn btn-primary btn-sm btn-square"
            title="添加固定代理"
            type="button"
            @click="addFixedProxy"
          >
            <IconPlus :size="16" />
          </button>
        </div>
        <div v-if="runtime.fixedProxies.value.length === 0" class="py-5 text-sm opacity-55">
          暂无固定代理
        </div>
        <div v-else class="divide-y divide-base-content/8">
          <div
            v-for="proxy in runtime.fixedProxies.value"
            :key="proxy.name"
            class="grid gap-2 py-3 md:grid-cols-[1fr_auto]"
          >
            <div class="min-w-0 text-sm">
              <div class="truncate font-medium">{{ proxy.name }}</div>
              <div class="truncate text-xs opacity-60" :title="proxy.uri">
                {{ proxy.type || 'proxy' }}
              </div>
            </div>
            <div class="flex items-center gap-1">
              <button
                aria-label="编辑"
                class="btn btn-ghost btn-sm btn-square"
                title="编辑"
                type="button"
                @click="editFixedProxy(proxy)"
              >
                <IconPencil :size="16" />
              </button>
              <button
                aria-label="删除"
                class="btn btn-ghost btn-sm btn-square text-error"
                :disabled="runtime.saving.value"
                title="删除"
                type="button"
                @click="runtime.deleteFixedProxy(proxy)"
              >
                <IconTrash :size="16" />
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="min-w-0">
        <div class="mb-2 flex items-center justify-between gap-3">
          <h2 class="flex items-center gap-2 text-base font-semibold">
            <IconLink :size="18" />
            订阅
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
        <div v-if="runtime.subscriptions.value.length === 0" class="py-5 text-sm opacity-55">
          暂无订阅
        </div>
        <div v-else class="divide-y divide-base-content/8">
          <div
            v-for="subscription in runtime.subscriptions.value"
            :key="subscription.name"
            class="grid gap-2 py-3 md:grid-cols-[1fr_auto]"
          >
            <div class="min-w-0 text-sm">
              <div class="truncate font-medium">{{ subscription.name }}</div>
              <div class="truncate text-xs opacity-60" :title="subscription.url">
                proxy-provider
              </div>
            </div>
            <div class="flex items-center gap-1">
              <button
                aria-label="编辑"
                class="btn btn-ghost btn-sm btn-square"
                title="编辑"
                type="button"
                @click="editSubscription(subscription)"
              >
                <IconPencil :size="16" />
              </button>
              <button
                aria-label="删除"
                class="btn btn-ghost btn-sm btn-square text-error"
                :disabled="runtime.saving.value"
                title="删除"
                type="button"
                @click="runtime.deleteSubscription(subscription)"
              >
                <IconTrash :size="16" />
              </button>
            </div>
          </div>
        </div>
      </section>
    </div>

    <p v-if="runtime.error.value" class="text-sm text-error">
      {{ runtime.error.value }}
    </p>

    <ProxyRuntimeNativeFixedProxyModal ref="fixedModal" :runtime="runtime" />
    <ProxyRuntimeNativeSubscriptionModal ref="subscriptionModal" :runtime="runtime" />
  </section>
</template>
