import type {
  ProxyEdgeAccessCheck,
  ProxyExitGeo,
  ProxyIPFraudCheck,
  ProxyIPFraudProviderDescriptor,
  ProxyIPGeoProviderDescriptor,
  ProxyRuntimeSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  type ProxyRuntimeIPFraudProviderRow,
  proxyRuntimeIPFraudRowFromDescriptor,
  proxyRuntimeIPFraudRowFromSettings,
  proxyRuntimeUniqueIPFraudProviderID,
} from '~/composables/proxyRuntimeIPFraudProviderRows'
import {
  type ProxyRuntimeIPGeoProviderRow,
  proxyRuntimeIPGeoRowFromDescriptor,
  proxyRuntimeIPGeoRowFromSettings,
  proxyRuntimeUniqueIPGeoProviderID,
} from '~/composables/proxyRuntimeIPGeoProviderRows'
import {
  proxyRuntimeIPFraudSettingsFromRows,
  proxyRuntimeIPGeoSettingsFromRows,
} from '~/composables/proxyRuntimePluginSettingsPayload'
import {
  applyIPFraudDescriptor,
  applyIPGeoDescriptor,
} from '~/composables/proxyRuntimePluginRowActions'

export function useProxyRuntimePlugins() {
  const api = useProxyRuntimeApi()
  const settings = ref<ProxyRuntimeSettings>()
  const descriptors = ref<ProxyIPFraudProviderDescriptor[]>([])
  const geoDescriptors = ref<ProxyIPGeoProviderDescriptor[]>([])
  const fraudRows = ref<ProxyRuntimeIPFraudProviderRow[]>([])
  const geoRows = ref<ProxyRuntimeIPGeoProviderRow[]>([])
  const edgeForm = reactive({ enabled: false, url: '', token_value: '', clear_token: false })
  const fraudIP = ref('')
  const geoIP = ref('')
  const geoResult = ref<ProxyExitGeo>()
  const fraudResult = ref<ProxyIPFraudCheck>()
  const edgeIP = ref('')
  const edgeCountry = ref('')
  const edgeResult = ref<ProxyEdgeAccessCheck>()
  const loading = ref(false)
  const saving = ref(false)
  const checking = ref(false)
  const error = ref('')
  const availableFraudProviderOptions = computed(() => descriptors.value)
  const availableGeoProviderOptions = computed(() => geoDescriptors.value)

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [settingsRes, fraudRes, geoRes] = await Promise.all([
        api.getSettings(),
        api.listIPFraudProviders(),
        api.listIPGeoProviders(),
      ])
      settings.value = settingsRes.settings
      descriptors.value = fraudRes.providers || []
      geoDescriptors.value = geoRes.providers || []
      fraudRows.value = (settings.value?.ip_fraud_providers || []).map(proxyRuntimeIPFraudRowFromSettings)
      geoRows.value = (settings.value?.ip_geo_providers || []).map(proxyRuntimeIPGeoRowFromSettings)
      edgeForm.enabled = !!settings.value?.edge_canary?.enabled
      edgeForm.url = settings.value?.edge_canary?.url || ''
      edgeForm.token_value = ''
      edgeForm.clear_token = false
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  async function save() {
    if (!settings.value) return false
    return withSave(async () => {
      const response = await api.updateRuntimeSettings({
        edge_canary: {
          enabled: edgeForm.enabled,
          url: edgeForm.url,
          token_value: edgeForm.token_value,
          clear_token: edgeForm.clear_token,
          token_secret_ref: undefined,
        },
        ip_fraud_providers: proxyRuntimeIPFraudSettingsFromRows(fraudRows.value),
        ip_geo_providers: proxyRuntimeIPGeoSettingsFromRows(geoRows.value),
        dynamic_ip_providers: settings.value?.dynamic_ip_providers || [],
        check_settings: settings.value?.check_settings,
        egress_profiles: settings.value?.egress_profiles || [],
        ingress_rules: settings.value?.ingress_rules || [],
      })
      settings.value = response.settings
      await load()
    })
  }

  async function saveFraudProvider(row: ProxyRuntimeIPFraudProviderRow) {
    if (!fraudRows.value.includes(row)) return false
    return save()
  }

  async function saveGeoProvider(row: ProxyRuntimeIPGeoProviderRow) {
    if (!geoRows.value.includes(row)) return false
    return save()
  }

  function addFraudProvider(providerID: string) {
    const descriptor = descriptors.value.find((item) => item.provider_id === providerID)
    if (!descriptor) return
    const row = proxyRuntimeIPFraudRowFromDescriptor(descriptor)
    row.provider_id = proxyRuntimeUniqueIPFraudProviderID(descriptor.provider_id, fraudRows.value)
    fraudRows.value.push(row)
  }

  function addGeoProvider(providerID: string) {
    const descriptor = geoDescriptors.value.find((item) => item.provider_id === providerID)
    if (!descriptor) return
    const row = proxyRuntimeIPGeoRowFromDescriptor(descriptor)
    row.provider_id = proxyRuntimeUniqueIPGeoProviderID(descriptor.provider_id, geoRows.value)
    geoRows.value.push(row)
  }

  async function deleteFraudProvider(index: number) {
    return deleteProviderRow(fraudRows.value, index)
  }

  async function deleteGeoProvider(index: number) {
    return deleteProviderRow(geoRows.value, index)
  }

  async function deleteProviderRow<T>(rows: T[], index: number) {
    const removed = rows[index]
    if (!removed) return false
    rows.splice(index, 1)
    const saved = await save()
    if (!saved) rows.splice(index, 0, removed)
    return saved
  }

  function applyDescriptor(row: ProxyRuntimeIPFraudProviderRow) {
    const descriptor = descriptors.value.find((item) => item.kind === row.kind)
    if (descriptor) applyIPFraudDescriptor(row, descriptor, fraudRows.value)
  }

  function applyGeoDescriptor(row: ProxyRuntimeIPGeoProviderRow) {
    const descriptor = geoDescriptors.value.find((item) => item.kind === row.kind)
    if (descriptor) applyIPGeoDescriptor(row, descriptor, geoRows.value)
  }

  async function checkGeo() {
    await withCheck(async () => { geoResult.value = (await api.checkProxyExitGeo({ ip: geoIP.value })).proxy_exit_geo })
  }

  async function checkFraud() {
    await withCheck(async () => { fraudResult.value = (await api.checkProxyIPFraud({ ip: fraudIP.value })).check })
  }

  async function checkEdge() {
    await withCheck(async () => {
      edgeResult.value = (await api.checkProxyEdgeAccess({ ip: edgeIP.value, expected_country_code: edgeCountry.value, listener_id: '' })).check
    })
  }

  return {
    addFraudProvider, addGeoProvider, applyDescriptor, applyGeoDescriptor,
    availableFraudProviderOptions, availableGeoProviderOptions, checkEdge,
    checkFraud, checkGeo, checking, deleteFraudProvider, deleteGeoProvider,
    descriptors, edgeCountry, edgeForm, edgeIP, edgeResult, error, fraudIP,
    fraudResult, fraudRows, geoDescriptors, geoIP, geoResult, geoRows, load,
    loading, save, saveFraudProvider, saveGeoProvider, saving,
  }

  async function withSave(action: () => Promise<void>) {
    saving.value = true
    error.value = ''
    try { await action(); return true } catch (err) { error.value = err instanceof Error ? err.message : String(err); return false }
    finally { saving.value = false }
  }

  async function withCheck(action: () => Promise<void>) {
    checking.value = true
    error.value = ''
    try { await action() } catch (err) { error.value = err instanceof Error ? err.message : String(err) }
    finally { checking.value = false }
  }
}

export type ProxyRuntimePluginsState = ReturnType<typeof useProxyRuntimePlugins>
