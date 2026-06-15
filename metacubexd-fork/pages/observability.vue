<script setup lang="ts">
const runtime = useProxyRuntimeMetrics()

useHead({ title: '观测' })
</script>

<template>
  <ProxyRuntimeMetrics :runtime="runtime" />
</template>
