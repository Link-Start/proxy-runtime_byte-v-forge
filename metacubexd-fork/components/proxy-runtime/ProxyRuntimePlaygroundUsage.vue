<script setup lang="ts">
import { IconClipboard, IconKey, IconTerminal2, IconWorld } from '@tabler/icons-vue'

defineProps<{
  copied: string
  copyText: (key: string, value: string) => Promise<void>
  credentials: string
  curlCommand: string
  gatewayHost: string
  gatewayPort: string
  proxyAuthority: string
  username: string
}>()

const expanded = ref(true)
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">使用方式</h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-primary badge-sm">
              <IconWorld :size="12" />
              {{ gatewayHost || '-' }}:{{ gatewayPort }}
            </span>
            <span class="badge badge-ghost badge-sm">
              <IconKey :size="12" />
              {{ username }}
            </span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <Button
            class="btn-ghost btn-sm btn-square"
            title="复制 curl"
            @click.stop="copyText('curl', curlCommand)"
          >
            <IconClipboard :size="16" />
          </Button>
        </div>
      </div>
    </template>

    <div class="col-span-full flex flex-col gap-2">
      <div class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <div class="flex items-center gap-2 text-sm">
          <IconWorld :size="16" />
          <span>代理地址</span>
        </div>
        <div class="flex min-w-0 items-center gap-2">
          <span class="truncate font-mono text-xs">{{ proxyAuthority }}</span>
          <Button
            class="btn-ghost btn-xs btn-square"
            title="复制代理地址"
            @click.stop="copyText('proxy', proxyAuthority)"
          >
            <IconClipboard :size="14" />
          </Button>
        </div>
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <div class="flex items-center gap-2 text-sm">
          <IconKey :size="16" />
          <span>认证</span>
        </div>
        <div class="flex min-w-0 items-center gap-2">
          <span class="truncate font-mono text-xs">{{ credentials }}</span>
          <Button
            class="btn-ghost btn-xs btn-square"
            title="复制认证"
            @click.stop="copyText('auth', credentials)"
          >
            <IconClipboard :size="14" />
          </Button>
        </div>
      </div>

      <div class="flex items-center justify-between gap-4 rounded-lg px-2 py-1.5 transition-colors hover:bg-base-content/5">
        <div class="flex items-center gap-2 text-sm">
          <IconTerminal2 :size="16" />
          <span>curl</span>
        </div>
        <div class="flex min-w-0 items-center gap-2">
          <span class="truncate font-mono text-xs">{{ curlCommand }}</span>
          <Button
            class="btn-ghost btn-xs btn-square"
            title="复制 curl"
            @click.stop="copyText('curl', curlCommand)"
          >
            <IconClipboard :size="14" />
          </Button>
        </div>
      </div>

      <div v-if="copied" class="px-2 text-xs text-success">已复制</div>
    </div>
  </Collapse>
</template>
