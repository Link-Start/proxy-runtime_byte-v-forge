<script setup lang="ts">
import type { ProxyRuntimeInUserRulesState } from '~/composables/useProxyRuntimeInUserRules'
import { IconDeviceFloppy, IconKey, IconReload, IconRoute } from '@tabler/icons-vue'

defineProps<{
  allowDynamicExit?: boolean
  canSave: boolean
  regeneratePassword: () => void
  runtime: ProxyRuntimeInUserRulesState
  save: () => Promise<void>
}>()

const expanded = ref(true)
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">可保存配置</h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">
              <IconKey :size="12" />
              {{ runtime.form.username }}
            </span>
            <span class="badge badge-ghost badge-sm">
              <IconRoute :size="12" />
              保存后继续使用
            </span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <Button
            class="btn-primary btn-sm btn-square"
            :disabled="runtime.saving.value || !canSave"
            title="保存"
            @click="save"
          >
            <IconDeviceFloppy :size="16" />
          </Button>
        </div>
      </div>
    </template>

    <div class="col-span-full flex flex-col gap-2">
      <div class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <div class="flex items-center gap-2 text-sm">
          <IconKey :size="16" />
          <span>密码</span>
        </div>
        <div class="flex min-w-0 flex-1 items-center justify-end gap-2">
          <input
            v-model="runtime.form.password_value"
            class="input-bordered input input-sm max-w-md flex-1 font-mono"
          />
          <Button
            class="btn-ghost btn-sm btn-square"
            title="重新生成密码"
            @click="regeneratePassword"
          >
            <IconReload :size="16" />
          </Button>
        </div>
      </div>

      <ProxyRuntimeInUserRouteFields
        :allow-dynamic-exit="allowDynamicExit"
        :runtime="runtime"
        dense
      />
    </div>
  </Collapse>
</template>
