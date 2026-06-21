<script setup lang="ts">
import type { ProxyGatewayInUserRulesState } from '~/composables/useProxyGatewayInUserRules'
import {
  countryOptions,
  currentOnlyOptions,
  dynamicIPAsnOptions,
  stateOptions,
} from '~/composables/proxyGatewayDynamicIPRegionOptions'

const props = defineProps<{ runtime: ProxyGatewayInUserRulesState }>()

const countries = computed(() =>
  countryOptions(props.runtime.form.exit_dynamic_region),
)
const states = computed(() =>
  stateOptions(
    props.runtime.form.exit_dynamic_region,
    props.runtime.form.exit_dynamic_state,
  ),
)
const cities = computed(() =>
  currentOnlyOptions('不指定城市', props.runtime.form.exit_dynamic_city),
)
const asns = computed(() =>
  dynamicIPAsnOptions(props.runtime.form.exit_dynamic_asn),
)

watch(
  () => props.runtime.form.exit_dynamic_region,
  () => {
    props.runtime.form.exit_dynamic_state = ''
    props.runtime.form.exit_dynamic_city = ''
  },
)

watch(
  () => props.runtime.form.exit_dynamic_state,
  () => {
    props.runtime.form.exit_dynamic_city = ''
  },
)
</script>

<template>
  <select v-model="runtime.form.exit_dynamic_region" class="select-bordered select w-full">
    <option v-for="country in countries" :key="country.value" :value="country.value">
      {{ country.label }}
    </option>
  </select>
  <select v-model="runtime.form.exit_dynamic_state" class="select-bordered select w-full">
    <option v-for="state in states" :key="state.value" :value="state.value">
      {{ state.label }}
    </option>
  </select>
  <select v-model="runtime.form.exit_dynamic_city" class="select-bordered select w-full">
    <option v-for="city in cities" :key="city.value" :value="city.value">
      {{ city.label }}
    </option>
  </select>
  <select v-model="runtime.form.exit_dynamic_asn" class="select-bordered select w-full">
    <option v-for="asn in asns" :key="asn.value" :value="asn.value">
      {{ asn.label }}
    </option>
  </select>
</template>
