<script setup lang="ts">
import type { EgressProfileSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { ProxyRuntimeEgressProfilesState } from '~/composables/useProxyRuntimeEgressProfiles'
import { exitText, lineText } from '~/composables/proxyRuntimeEgressProfileHelpers'
import { IconPencil, IconPlus, IconRoute, IconTrash } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeEgressProfilesState }>()
const modal = ref<{ open: () => void; close: () => void }>()

function addProfile() {
  props.runtime.resetForm()
  modal.value?.open()
}

function editProfile(profile: EgressProfileSettings) {
  props.runtime.editProfile(profile)
  modal.value?.open()
}

function profileText(profile: EgressProfileSettings) {
  return `路线: ${lineText(profile)} / 出口: ${exitText(profile)}`
}
</script>

<template>
  <section class="flex min-h-0 flex-col gap-3">
    <div class="flex items-center justify-between gap-3 border-b border-base-content/8 pb-2">
      <h2 class="flex items-center gap-2 text-base font-semibold">
        <IconRoute :size="18" />
        出口 Profile
      </h2>
      <button
        aria-label="添加出口 Profile"
        class="btn btn-primary btn-sm btn-square"
        title="添加出口 Profile"
        type="button"
        @click="addProfile"
      >
        <IconPlus :size="16" />
      </button>
    </div>

    <div v-if="runtime.profiles.value.length === 0" class="py-5 text-sm opacity-55">
      暂无出口 Profile
    </div>
    <div v-else class="divide-y divide-base-content/8">
      <div
        v-for="profile in runtime.profiles.value"
        :key="profile.profile_id"
        class="grid gap-2 py-3 md:grid-cols-[1fr_auto]"
      >
        <div class="min-w-0 text-sm">
          <div class="truncate font-medium">
            {{ profile.display_name || profile.profile_id }}
          </div>
          <div class="truncate text-xs opacity-60" :title="profileText(profile)">
            {{ profileText(profile) }}
          </div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-sm">{{ profile.enabled ? '启用' : '停用' }}</span>
            <span class="badge badge-ghost badge-sm">
              <IconRoute :size="12" />
              {{ exitText(profile) }}
            </span>
          </div>
        </div>
        <div class="flex items-center gap-1">
          <button
            aria-label="编辑"
            class="btn btn-ghost btn-sm btn-square"
            title="编辑"
            type="button"
            @click="editProfile(profile)"
          >
            <IconPencil :size="16" />
          </button>
          <button
            aria-label="删除"
            class="btn btn-ghost btn-sm btn-square text-error"
            :disabled="runtime.saving.value"
            title="删除"
            type="button"
            @click="runtime.deleteProfile(profile)"
          >
            <IconTrash :size="16" />
          </button>
        </div>
      </div>
    </div>

    <ProxyRuntimeEgressProfileModal ref="modal" :runtime="runtime" />
  </section>
</template>
