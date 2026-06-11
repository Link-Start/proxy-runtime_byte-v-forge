<script setup lang="ts">
const runtime = useProxyRuntimeInUserRules()

useHead({ title: '接入' })
onMounted(() => {
  void runtime.load()
})
</script>

<template>
  <ProxyRuntimeInUserRules :runtime="runtime" />
</template>
