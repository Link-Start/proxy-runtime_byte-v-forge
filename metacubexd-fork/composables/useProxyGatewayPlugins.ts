import {
  isProxyGatewayCancellation,
  proxyGatewayUserMessage,
} from '~/composables/proxyGatewayFetch'
import type {
  ProxyEdgeAccessCheck,
  ProxyExitGeo,
  ProxyIPFraudCheck,
  ProxyIPFraudProviderDescriptor,
  ProxyIPGeoProviderDescriptor,
  ProxyGatewaySettings,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  type ProxyGatewayIPFraudProviderRow,
  proxyGatewayIPFraudRowFromDescriptor,
  proxyGatewayIPFraudRowFromSettings,
  proxyGatewayUniqueIPFraudProviderID,
} from '~/composables/proxyGatewayIPFraudProviderRows'
import {
  type ProxyGatewayIPGeoProviderRow,
  proxyGatewayIPGeoRowFromDescriptor,
  proxyGatewayIPGeoRowFromSettings,
  proxyGatewayUniqueIPGeoProviderID,
} from '~/composables/proxyGatewayIPGeoProviderRows'
import {
  proxyGatewayIPFraudSettingsFromRows,
  proxyGatewayIPGeoSettingsFromRows,
} from '~/composables/proxyGatewayPluginSettingsPayload'
import {
  applyIPFraudDescriptor,
  applyIPGeoDescriptor,
} from '~/composables/proxyGatewayPluginRowActions'

export function useProxyGatewayPlugins() {
  const api = useProxyGatewayApi()
  const settings = ref<ProxyGatewaySettings>()
  const descriptors = ref<ProxyIPFraudProviderDescriptor[]>([])
  const geoDescriptors = ref<ProxyIPGeoProviderDescriptor[]>([])
  const fraudRows = ref<ProxyGatewayIPFraudProviderRow[]>([])
  const geoRows = ref<ProxyGatewayIPGeoProviderRow[]>([])
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
  let loadController: AbortController | undefined
  let loadSequence = 0
  const availableFraudProviderOptions = computed(() => descriptors.value)
  const availableGeoProviderOptions = computed(() => geoDescriptors.value)

  async function load() {
    const sequence = nextLoadSequence()
    const controller = new AbortController()
    loadController = controller
    loading.value = true
    error.value = ''
    try {
      const requestOptions = { signal: controller.signal }
      const [settingsRes, fraudRes, geoRes] = await Promise.all([
        api.getSettings(requestOptions),
        api.listIPFraudProviders(requestOptions),
        api.listIPGeoProviders(requestOptions),
      ])
      if (!currentLoad(sequence)) return
      settings.value = settingsRes.settings
      descriptors.value = fraudRes.providers || []
      geoDescriptors.value = geoRes.providers || []
      fraudRows.value = (settings.value?.ip_fraud_providers || []).map(proxyGatewayIPFraudRowFromSettings)
      geoRows.value = (settings.value?.ip_geo_providers || []).map(proxyGatewayIPGeoRowFromSettings)
      edgeForm.enabled = !!settings.value?.edge_canary?.enabled
      edgeForm.url = settings.value?.edge_canary?.url || ''
      edgeForm.token_value = ''
      edgeForm.clear_token = false
    } catch (err) {
      if (!currentLoad(sequence) || isProxyGatewayCancellation(err)) return
      error.value = proxyGatewayUserMessage(err)
    } finally {
      if (currentLoad(sequence)) {
        loading.value = false
        loadController = undefined
      }
    }
  }

  function abortLoad() {
    loadController?.abort()
    loadController = undefined
  }

  function nextLoadSequence() {
    abortLoad()
    loadSequence += 1
    return loadSequence
  }

  function currentLoad(sequence: number) {
    return sequence === loadSequence
  }

  onBeforeUnmount(abortLoad)

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
        ip_fraud_providers: proxyGatewayIPFraudSettingsFromRows(fraudRows.value),
        ip_geo_providers: proxyGatewayIPGeoSettingsFromRows(geoRows.value),
        dynamic_ip_providers: settings.value?.dynamic_ip_providers || [],
        check_settings: settings.value?.check_settings,
        egress_profiles: settings.value?.egress_profiles || [],
        ingress_rules: settings.value?.ingress_rules || [],
      })
      settings.value = response.settings
      await load()
    })
  }

  async function saveFraudProvider(row: ProxyGatewayIPFraudProviderRow) {
    if (!fraudRows.value.includes(row)) return false
    return save()
  }

  async function saveGeoProvider(row: ProxyGatewayIPGeoProviderRow) {
    if (!geoRows.value.includes(row)) return false
    return save()
  }

  function addFraudProvider(providerID: string) {
    const descriptor = descriptors.value.find((item) => item.provider_id === providerID)
    if (!descriptor) return
    const row = proxyGatewayIPFraudRowFromDescriptor(descriptor)
    row.provider_id = proxyGatewayUniqueIPFraudProviderID(descriptor.provider_id, fraudRows.value)
    fraudRows.value.push(row)
  }

  function addGeoProvider(providerID: string) {
    const descriptor = geoDescriptors.value.find((item) => item.provider_id === providerID)
    if (!descriptor) return
    const row = proxyGatewayIPGeoRowFromDescriptor(descriptor)
    row.provider_id = proxyGatewayUniqueIPGeoProviderID(descriptor.provider_id, geoRows.value)
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

  function applyDescriptor(row: ProxyGatewayIPFraudProviderRow) {
    const descriptor = descriptors.value.find((item) => item.kind === row.kind)
    if (descriptor) applyIPFraudDescriptor(row, descriptor, fraudRows.value)
  }

  function applyGeoDescriptor(row: ProxyGatewayIPGeoProviderRow) {
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
    try { await action(); return true } catch (err) { error.value = proxyGatewayUserMessage(err); return false }
    finally { saving.value = false }
  }

  async function withCheck(action: () => Promise<void>) {
    checking.value = true
    error.value = ''
    try { await action() } catch (err) { error.value = proxyGatewayUserMessage(err) }
    finally { checking.value = false }
  }
}

export type ProxyGatewayPluginsState = ReturnType<typeof useProxyGatewayPlugins>
