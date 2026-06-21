<script setup lang="ts">
const runtime = useProxyGatewayInUserRules()

useHead({ title: '接入' })
onMounted(() => {
  void runtime.load()
})
</script>

<template>
  <ProxyGatewayInUserRules :runtime="runtime" />
</template>
