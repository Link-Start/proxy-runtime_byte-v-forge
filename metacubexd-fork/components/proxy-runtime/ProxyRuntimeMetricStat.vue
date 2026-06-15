<script setup lang="ts">
defineProps<{
  label: string
  note: string
  tone?: 'primary' | 'success' | 'warning' | 'error'
  value: string
}>()
</script>

<template>
  <div
    class="animate-fade-slide-in rounded-2xl border border-base-content/10 bg-base-200/70 p-4 backdrop-blur-sm"
  >
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <div class="text-xs opacity-60">{{ label }}</div>
        <div class="mt-2 truncate text-2xl font-semibold">{{ value }}</div>
      </div>
      <span
        class="h-2.5 w-2.5 rounded-full"
        :class="{
          'bg-primary': tone === 'primary' || !tone,
          'bg-success': tone === 'success',
          'bg-warning': tone === 'warning',
          'bg-error': tone === 'error',
        }"
      />
    </div>
    <div class="mt-3 truncate text-xs opacity-60">{{ note }}</div>
  </div>
</template>
