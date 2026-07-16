import { describe, expect, it } from 'vitest'

const { existsSync, readFileSync } = await import('node:' + 'fs')
const bankSource = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')
const userApiSource = readFileSync(new URL('../../api/user.ts', import.meta.url), 'utf8')
const selectInputUrl = new URL('../../components/PpmxSelectInput.vue', import.meta.url)
const selectInputSource = existsSync(selectInputUrl)
  ? readFileSync(selectInputUrl, 'utf8')
  : ''

describe('bank withdrawal account API integration', () => {
  it('uses the backend country withdrawal-field contract for Mexico', () => {
    expect(bankSource).toContain("import {")
    expect(bankSource).toContain('APP_COUNTRY_CODE')
    expect(bankSource).toContain('getAppCountries()')
    expect(bankSource).toContain('getCountryWithdrawFields(APP_COUNTRY_CODE)')
    expect(bankSource).toContain('getWithdrawAccounts()')
    expect(bankSource).toContain('addWithdrawAccount({')
    expect(bankSource).toContain('updateWithdrawAccount(editingId.value')
    expect(bankSource).toContain('deleteWithdrawAccount(account.id)')
    expect(bankSource).toContain('setDefaultWithdrawAccount(account.id)')
    expect(bankSource).toContain('countryCode: activeCountryCode.value')
    expect(bankSource).toContain('accountData')
    expect(bankSource).toContain('withdrawFields.value.map(field => [')
    expect(bankSource).not.toContain('country-pill')
  })

  it('does not show the error copy while the bank page is still loading', () => {
    expect(bankSource).toContain('const pageStateMessage = computed')
    expect(bankSource).toContain("pageStatus.value === 'loading'")
    expect(bankSource).toContain(':message="pageStateMessage"')
    expect(bankSource).not.toContain(':message="t(\'bank.errorMessage\')"')
  })

  it('delegates select input behavior to the shared PP.MX select component', () => {
    expect(selectInputSource).toContain('defineModel<string>')
    expect(bankSource).toContain('PpmxSelectInput')
    expect(bankSource).toContain(':options="fieldOptions(field)"')
    expect(bankSource).toContain('v-model="fieldValues[field.fieldKey]"')
    expect(bankSource).not.toContain('selectingField')
    expect(bankSource).not.toContain('showFieldPicker')
    expect(bankSource).not.toContain('isSelectOpen(field)')
    expect(bankSource).not.toContain('<select')
  })

  it('keeps responsive popup and outside-click behavior inside the shared select component', () => {
    expect(selectInputSource).toContain('useResponsiveLayout()')
    expect(selectInputSource).toContain('closeSelectOnOutsidePointerDown')
    expect(selectInputSource).toContain("document.addEventListener('pointerdown', closeSelectOnOutsidePointerDown)")
    expect(selectInputSource).toContain("document.removeEventListener('pointerdown', closeSelectOnOutsidePointerDown)")
    expect(selectInputSource).toContain("target.closest('.ppmx-select-input')")
    expect(selectInputSource).toContain('ppmx-select-input__menu')
    expect(selectInputSource).toContain('ppmx-select-input__popup')
    expect(selectInputSource).toContain('ppmx-select-input__popup-list')
  })

  it('keeps the mobile popup header separate from the scrollable option list', () => {
    expect(selectInputSource).toContain('ppmx-select-input__popup-list')
    expect(selectInputSource).toContain('overflow-y: auto')
    expect(selectInputSource).not.toContain('.ppmx-select-input__popup-body h3 {\n  position: sticky;')
  })

  it('adds typed API wrappers without the legacy /api prefix', () => {
    expect(userApiSource).toContain("get<WithdrawAccountItem[]>('/v1/app/withdrawAccount/list')")
    expect(userApiSource).toContain("post<WithdrawAccountItem>('/v1/app/withdrawAccount', data)")
    expect(userApiSource).toContain('post<WithdrawAccountItem>(`/v1/app/withdrawAccount/${id}/update`, data)')
    expect(userApiSource).toContain('del<string>(`/v1/app/withdrawAccount/${id}`)')
    expect(userApiSource).toContain('post<string>(`/v1/app/withdrawAccount/${id}/setDefault`, {})')
    expect(userApiSource).toContain('`/v1/app/country/${code}/withdrawFields`')
    expect(userApiSource).not.toContain("'/api/v1/app/withdrawAccount")
  })
})
