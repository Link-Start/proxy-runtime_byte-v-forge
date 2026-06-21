<script setup lang="ts">
import type { ProxyGatewayPluginsState } from '~/composables/useProxyGatewayPlugins'
import { IconCloudCheck, IconDeviceFloppy, IconPlayerPlay } from '@tabler/icons-vue'

defineProps<{ runtime: ProxyGatewayPluginsState }>()
</script>

<template>
  <section class="rounded-2xl border border-base-content/10 bg-base-200/45 p-4">
    <div class="mb-3 flex items-center justify-between gap-2">
      <div>
        <h2 class="flex items-center gap-2 text-base font-semibold">
          <IconCloudCheck :size="18" /> CF Canary
        </h2>
        <p class="mt-1 text-xs opacity-60">通过出口访问 canary，判断边缘访问风险</p>
      </div>
      <button
        aria-label="保存CF Canary"
        class="btn btn-primary btn-sm btn-square"
        :disabled="runtime.saving.value"
        title="保存CF Canary"
        type="button"
        @click="runtime.save"
      >
        <IconDeviceFloppy :size="16" />
      </button>
    </div>

    <div class="grid gap-2 md:grid-cols-[auto_1fr_14rem_auto]">
      <label class="flex items-center gap-2 text-sm">
        <input v-model="runtime.edgeForm.enabled" class="toggle toggle-primary toggle-sm" type="checkbox" />
        启用
      </label>
      <input v-model.trim="runtime.edgeForm.url" class="input input-bordered input-sm" placeholder="Canary URL" />
      <input v-model="runtime.edgeForm.token_value" class="input input-bordered input-sm" placeholder="Token / 空则保留" type="password" />
      <label class="flex items-center gap-1 text-xs">
        <input v-model="runtime.edgeForm.clear_token" class="checkbox checkbox-xs" type="checkbox" />
        清除Token
      </label>
    </div>

    <div class="mt-3 grid gap-2 md:grid-cols-[1fr_10rem_auto]">
      <input v-model.trim="runtime.edgeIP.value" class="input input-bordered input-sm" placeholder="可选：指定IP；空则探测默认出口" />
      <input v-model.trim="runtime.edgeCountry.value" class="input input-bordered input-sm" placeholder="期望国家" />
      <button class="btn btn-sm" :disabled="runtime.checking.value" type="button" @click="runtime.checkEdge">
        <IconPlayerPlay :size="15" /> 执行
      </button>
    </div>

    <div v-if="runtime.edgeResult.value" class="mt-3 flex flex-wrap gap-2 text-xs">
      <span class="badge" :class="riskBadgeClass(edgeRiskLabel(runtime.edgeResult.value.risk_level))">
        {{ edgeRiskLabel(runtime.edgeResult.value.risk_level) }}
      </span>
      <span class="badge badge-ghost">{{ runtime.edgeResult.value.ip }}</span>
      <span class="badge badge-ghost">score {{ runtime.edgeResult.value.risk_score }}</span>
      <span v-if="runtime.edgeResult.value.error_message" class="badge badge-error">
        {{ runtime.edgeResult.value.error_message }}
      </span>
    </div>
  </section>
</template>
