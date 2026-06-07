<script setup lang="ts">
import type { ProxyRuntimePluginsState } from '~/composables/useProxyRuntimePlugins'
import { IconMapPin, IconPlus, IconSearch } from '@tabler/icons-vue'

defineProps<{ runtime: ProxyRuntimePluginsState }>()
const modal = ref<{ open: () => void }>()
const manualExpanded = ref(false)
</script>

<template>
  <section class="flex min-h-0 flex-col gap-3 p-2">
    <div class="flex items-center justify-between gap-3">
      <h2 class="flex items-center gap-2 text-base font-semibold">
        <IconMapPin :size="18" />
        IP Geo 检测
      </h2>
      <div class="flex items-center gap-2">
        <span class="badge badge-primary badge-sm">{{ runtime.geoRows.value.length }}</span>
        <button
          aria-label="添加 Geo Provider"
          class="btn btn-primary btn-sm btn-square"
          :disabled="runtime.availableGeoProviderOptions.value.length === 0"
          title="添加 Geo Provider"
          type="button"
          @click="modal?.open()"
        >
          <IconPlus :size="16" />
        </button>
      </div>
    </div>

    <div
      v-if="runtime.geoRows.value.length === 0"
      class="grid gap-3 rounded-lg border border-dashed border-base-content/15 p-6 text-center text-sm"
    >
      <div class="opacity-60">暂无已添加 IP Geo Provider</div>
      <div class="flex justify-center">
        <button
          aria-label="添加 Geo Provider"
          class="btn btn-primary btn-sm btn-square"
          :disabled="runtime.availableGeoProviderOptions.value.length === 0"
          title="添加 Geo Provider"
          type="button"
          @click="modal?.open()"
        >
          <IconPlus :size="16" />
        </button>
      </div>
    </div>

    <div v-else class="grid gap-3">
      <ProxyRuntimeIPGeoProviderCard
        v-for="(row, index) in runtime.geoRows.value"
        :key="`${row.provider_id}-${index}`"
        :index="index"
        :row="row"
        :runtime="runtime"
      />
    </div>

    <Collapse :is-open="manualExpanded" @collapse="manualExpanded = $event">
      <template #title>
        <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
          <div class="min-w-0">
            <h3 class="truncate text-base font-semibold">手动检测</h3>
            <div class="mt-2 flex flex-wrap gap-1">
              <span class="badge badge-ghost badge-sm">
                <IconMapPin :size="12" />
                Geo
              </span>
            </div>
          </div>
        </div>
      </template>

      <div class="col-span-full grid gap-3">
        <div class="grid gap-2 md:grid-cols-[1fr_auto]">
          <input
            v-model.trim="runtime.geoIP.value"
            class="input input-bordered input-sm"
            placeholder="IP，例如 8.8.8.8"
          />
          <button
            aria-label="检测 IP Geo"
            class="btn btn-sm btn-square"
            :disabled="runtime.checking.value"
            title="检测 IP Geo"
            type="button"
            @click="runtime.checkGeo"
          >
            <IconSearch :size="16" />
          </button>
        </div>

        <div v-if="runtime.geoResult.value" class="grid gap-2 text-sm sm:grid-cols-4">
          <div class="rounded-lg bg-base-100 p-3">
            <div class="text-xs opacity-50">IP</div>
            <div class="truncate font-medium">{{ runtime.geoResult.value.ip }}</div>
          </div>
          <div class="rounded-lg bg-base-100 p-3">
            <div class="text-xs opacity-50">国家</div>
            <div class="font-medium">{{ runtime.geoResult.value.country_code || '-' }}</div>
          </div>
          <div class="rounded-lg bg-base-100 p-3">
            <div class="text-xs opacity-50">地区</div>
            <div class="font-medium">{{ runtime.geoResult.value.region || '-' }}</div>
          </div>
          <div class="rounded-lg bg-base-100 p-3">
            <div class="text-xs opacity-50">城市</div>
            <div class="font-medium">{{ runtime.geoResult.value.city || '-' }}</div>
          </div>
        </div>
      </div>
    </Collapse>

    <ProxyRuntimeIPGeoProviderModal ref="modal" :runtime="runtime" />
  </section>
</template>
