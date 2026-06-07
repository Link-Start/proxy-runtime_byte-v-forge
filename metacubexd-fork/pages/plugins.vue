<script setup lang="ts">
const runtime = useProxyRuntimePlugins()

useHead({ title: '插件' })
onMounted(runtime.load)
</script>

<template>
  <ProxyRuntimePlugins :runtime="runtime" />
</template>
