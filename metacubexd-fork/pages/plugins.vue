<script setup lang="ts">
const runtime = useProxyGatewayPlugins()

useHead({ title: '插件' })
onMounted(runtime.load)
</script>

<template>
  <ProxyGatewayPlugins :runtime="runtime" />
</template>
