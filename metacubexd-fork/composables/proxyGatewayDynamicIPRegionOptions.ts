import { allCountries } from 'country-region-data'

export interface SelectOption {
  value: string
  label: string
}

export function countryOptions(current = '') {
  const options = allCountries.map(([name, code]) => ({
    value: code,
    label: `${name} / ${code}`,
  }))
  return withCurrentOption(
    [{ value: '', label: '不指定国家/地区' }, ...options],
    current,
  )
}

export function stateOptions(countryCode: string, current = '') {
  const country = allCountries.find(([, code]) => code === countryCode)
  const options =
    country?.[2]?.map(([name, code]) => ({ value: code, label: `${name} / ${code}` })) ||
    []
  return withCurrentOption([{ value: '', label: '不指定州/省' }, ...options], current)
}

export function dynamicIPAsnOptions(current = '') {
  return withCurrentOption([{ value: '', label: '不指定 ASN' }], current)
}

export function currentOnlyOptions(emptyLabel: string, current = '') {
  return withCurrentOption([{ value: '', label: emptyLabel }], current)
}

function withCurrentOption(options: SelectOption[], current: string) {
  const value = current.trim()
  if (!value || options.some((option) => option.value === value)) return options
  return [...options, { value, label: `当前：${value}` }]
}
