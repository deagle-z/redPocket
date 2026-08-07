import type { RechargeField, RechargeFieldOption } from '@/api/user'

/**
 * 秘鲁（PE）VCPAYPEN 通道的充值 / 提现字段，前端写死，不再依赖后端
 * `/v1/app/country/PE/rechargeFields` 与 `/v1/app/country/PE/withdrawFields` 的配置。
 *
 * 字段定义来源：core/pay/vcpaypen/readme.md
 * - 代收 POST /pay/save：identity_type(Y)、identity(Y)、identity_name(N, string[4,64])
 * - 代付 POST /wd/save：bank_code(Y)、bank_owner(Y, [2,64])、bank_account(Y)、
 *   bank_account_type(Y, SA/CA)、identity_type(Y)、identity(Y, [8,64])
 *
 * fieldKey 必须与后端取值别名对得上：
 * - 充值：core/pay/vcpaypen/vcpaypen.go resolvePayinIdentity / buildPayinParams
 * - 提现：core/repository/withdraw_order_br.go submitWithdrawPayout（accName / accNo /
 *   bankCode / accountType / identityType / identityNo / email / phone）
 */

type Translate = (key: string, named?: Record<string, unknown>) => string

export interface PayFieldDef {
  fieldKey: string
  labelKey: string
  placeholderKey?: string
  errorTipsKey?: string
  fieldType: RechargeField['fieldType']
  dataType: string
  isRequired: 0 | 1
  defaultValue?: string
  minLength?: number
  maxLength?: number
  regexRule?: string
  options?: RechargeFieldOption[]
}

/** 证件类型，见 readme.md「身份类型（identity_type）取值」 */
const IDENTITY_TYPE_OPTIONS: RechargeFieldOption[] = [
  { label: 'DNI', value: 'DNI' },
  { label: 'CE', value: 'CE' },
  { label: 'PAS', value: 'PAS' },
  { label: 'RUC', value: 'RUC' },
]

/** 证件号：DNI 8 位、RUC 11 位、PAS 7-12 位、CE 8-12 位字母数字 */
const IDENTITY_REGEX = '^[A-Za-z0-9]{7,20}$'

/** 电子钱包 bank_code，账号填开户手机号；其余为银行，账号填 20 位 CCI */
export const WALLET_BANK_CODES = ['PEW001', 'PEW002'] as const

/** 代付 BankCode，见 readme.md「代付 BankCode」 */
export const WITHDRAW_BANK_OPTIONS: RechargeFieldOption[] = [
  { label: 'Yape', value: 'PEW001' },
  { label: 'Plin', value: 'PEW002' },
  { label: 'BCP (Banco de Crédito del Perú)', value: 'PEBB04' },
  { label: 'BBVA Perú', value: 'PEBB03' },
  { label: 'Interbank', value: 'PEBI01' },
  { label: 'Scotiabank Perú', value: 'PEBS02' },
  { label: 'BanBif', value: 'PEBB01' },
  { label: 'Banco de la Nación', value: 'PEBB02' },
  { label: 'Banco de Comercio', value: 'PEBB05' },
  { label: 'Banco Azteca Perú', value: 'PEBA01' },
  { label: 'Banco Cencosud', value: 'PEBC11' },
  { label: 'Banco Falabella', value: 'PEBF01' },
  { label: 'Banco GNB Perú', value: 'PEBG01' },
  { label: 'Banco Pichincha', value: 'PEBP01' },
  { label: 'Banco Ripley', value: 'PEBR01' },
  { label: 'Citibank Perú', value: 'PEBC12' },
  { label: 'ICBC Perú', value: 'PEBI02' },
  { label: 'MiBanco', value: 'PEBM01' },
  { label: 'Santander Perú', value: 'PEBS01' },
  { label: 'Caja Arequipa', value: 'PEBC01' },
  { label: 'Caja Cusco', value: 'PEBC02' },
  { label: 'Caja Huancayo', value: 'PEBC03' },
  { label: 'Caja Maynas', value: 'PEBC04' },
  { label: 'Caja Metropolitana', value: 'PEBC05' },
  { label: 'Caja Municipal Ica', value: 'PEBC06' },
  { label: 'Caja Piura', value: 'PEBC07' },
  { label: 'Caja Sullana', value: 'PEBC08' },
  { label: 'Caja Tacna', value: 'PEBC09' },
  { label: 'Caja Trujillo', value: 'PEBC10' },
]

/** 受益人账号类型，储蓄 SA / 活期 CA */
const ACCOUNT_TYPE_OPTIONS: RechargeFieldOption[] = [
  { label: 'SA', value: 'SA' },
  { label: 'CA', value: 'CA' },
]

/** 充值（代收）字段 */
export const RECHARGE_FIELD_DEFS: PayFieldDef[] = [
  {
    fieldKey: 'identityType',
    labelKey: 'payFields.identityType',
    placeholderKey: 'payFields.identityTypePlaceholder',
    fieldType: 'select',
    dataType: 'string',
    isRequired: 1,
    defaultValue: 'DNI',
    options: IDENTITY_TYPE_OPTIONS,
  },
  {
    fieldKey: 'identity',
    labelKey: 'payFields.identity',
    placeholderKey: 'payFields.identityPlaceholder',
    errorTipsKey: 'payFields.identityError',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 1,
    minLength: 7,
    maxLength: 20,
    regexRule: IDENTITY_REGEX,
  },
  {
    fieldKey: 'identityName',
    labelKey: 'payFields.identityName',
    placeholderKey: 'payFields.identityNamePlaceholder',
    errorTipsKey: 'payFields.identityNameError',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 0,
    minLength: 4,
    maxLength: 64,
  },
]

/** 提现（代付）字段 */
export const WITHDRAW_FIELD_DEFS: PayFieldDef[] = [
  {
    fieldKey: 'bankCode',
    labelKey: 'payFields.bankCode',
    placeholderKey: 'payFields.bankCodePlaceholder',
    fieldType: 'select',
    dataType: 'string',
    isRequired: 1,
    defaultValue: 'PEW001',
    options: WITHDRAW_BANK_OPTIONS,
  },
  {
    fieldKey: 'accName',
    labelKey: 'payFields.accName',
    placeholderKey: 'payFields.accNamePlaceholder',
    errorTipsKey: 'payFields.accNameError',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 1,
    minLength: 2,
    maxLength: 64,
  },
  {
    // accNo 的长度/格式随 bankCode 变化，见 resolveWithdrawAccNoDef
    fieldKey: 'accNo',
    labelKey: 'payFields.accNoWallet',
    placeholderKey: 'payFields.accNoWalletPlaceholder',
    errorTipsKey: 'payFields.accNoWalletError',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 1,
    minLength: 9,
    maxLength: 9,
    regexRule: '^9[0-9]{8}$',
  },
  {
    fieldKey: 'accountType',
    labelKey: 'payFields.accountType',
    placeholderKey: 'payFields.accountTypePlaceholder',
    fieldType: 'select',
    dataType: 'string',
    isRequired: 1,
    defaultValue: 'SA',
    options: ACCOUNT_TYPE_OPTIONS,
  },
  {
    fieldKey: 'identityType',
    labelKey: 'payFields.identityType',
    placeholderKey: 'payFields.identityTypePlaceholder',
    fieldType: 'select',
    dataType: 'string',
    isRequired: 1,
    defaultValue: 'DNI',
    options: IDENTITY_TYPE_OPTIONS,
  },
  {
    fieldKey: 'identityNo',
    labelKey: 'payFields.identity',
    placeholderKey: 'payFields.identityPlaceholder',
    errorTipsKey: 'payFields.identityError',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 1,
    minLength: 7,
    maxLength: 20,
    regexRule: IDENTITY_REGEX,
  },
  {
    // submitWithdrawPayout 要求 phone 非空（withdraw_payout_phone_required）
    fieldKey: 'phone',
    labelKey: 'payFields.phone',
    placeholderKey: 'payFields.phonePlaceholder',
    errorTipsKey: 'payFields.phoneError',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 1,
    minLength: 9,
    maxLength: 9,
    regexRule: '^9[0-9]{8}$',
  },
  {
    // submitWithdrawPayout 要求 email 非空（withdraw_payout_email_required）
    fieldKey: 'email',
    labelKey: 'payFields.email',
    placeholderKey: 'payFields.emailPlaceholder',
    errorTipsKey: 'payFields.emailError',
    fieldType: 'input',
    dataType: 'string',
    isRequired: 1,
    maxLength: 128,
    regexRule: '^[^\\s@]+@[^\\s@]+\\.[^\\s@]{2,}$',
  },
]

export function isWalletBankCode(bankCode: string | undefined | null) {
  const code = String(bankCode ?? '').trim().toUpperCase()

  return (WALLET_BANK_CODES as readonly string[]).includes(code)
}

/**
 * 电子钱包填 9 开头的 9 位手机号，银行填 20 位 CCI 账号。
 * 见 core/pay/vcpaypen/vcpaypen.go validatePayoutBankAccount。
 */
function resolveWithdrawAccNoDef(def: PayFieldDef, bankCode: string | undefined | null): PayFieldDef {
  if (isWalletBankCode(bankCode)) return def

  return {
    ...def,
    labelKey: 'payFields.accNoBank',
    placeholderKey: 'payFields.accNoBankPlaceholder',
    errorTipsKey: 'payFields.accNoBankError',
    minLength: 20,
    maxLength: 20,
    regexRule: '^[0-9]{20}$',
  }
}

function toRechargeField(def: PayFieldDef, t: Translate): RechargeField {
  return {
    fieldKey: def.fieldKey,
    fieldLabel: t(def.labelKey),
    fieldPlaceholder: def.placeholderKey ? t(def.placeholderKey) : '',
    fieldType: def.fieldType,
    dataType: def.dataType,
    isRequired: def.isRequired,
    defaultValue: def.defaultValue ?? '',
    minLength: def.minLength ?? null,
    maxLength: def.maxLength ?? null,
    regexRule: def.regexRule ?? null,
    errorTips: def.errorTipsKey ? t(def.errorTipsKey) : null,
    optionsJson: def.options ? JSON.stringify(def.options) : null,
  }
}

/** 充值页面写死的字段列表 */
export function buildRechargeFields(t: Translate): RechargeField[] {
  return RECHARGE_FIELD_DEFS.map(def => toRechargeField(def, t))
}

/** 提现账户写死的字段列表，accNo 规则随已选 bankCode 变化 */
export function buildWithdrawFields(t: Translate, bankCode?: string | null): RechargeField[] {
  return WITHDRAW_FIELD_DEFS
    .map(def => (def.fieldKey === 'accNo' ? resolveWithdrawAccNoDef(def, bankCode) : def))
    .map(def => toRechargeField(def, t))
}
