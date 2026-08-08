<script setup lang="ts">
import '@/assets/styles/pp-mx-home.css'
import type { CreateWithdrawOrderReq, TgWithdrawFlowBatchSummary, WithdrawAccountItem } from '@/api/user'
import {
  createWithdrawOrderV2,
  getCurrentTgWithdrawFlowBatchSummary,
  getWithdrawAccounts,
  redeemExchangeCode,
  transferTgWalletBalance,
} from '@/api/user'
import {
  APP_COUNTRY_CODE,
  APP_CURRENCY,
  APP_CURRENCY_SYMBOL,
} from '@/config/market'
import { formatMoney, fromCent, toDisplayCents } from '@/utils/money'
import { showToast } from 'vant'
import PpmxActionCard from '@/components/PpmxActionCard.vue'
import PpmxDialog from '@/components/PpmxDialog.vue'
import PpmxGuestPanel from '@/components/PpmxGuestPanel.vue'
import PpmxMoneyText from '@/components/PpmxMoneyText.vue'
import PpmxRechargeModal from '@/components/PpmxRechargeModal.vue'
import PpmxSelectInput from '@/components/PpmxSelectInput.vue'
import PpmxSkeleton from '@/components/PpmxSkeleton.vue'
import PpmxSocialModal from '@/components/PpmxSocialModal.vue'
import PpmxWithdrawModal from '@/components/PpmxWithdrawModal.vue'
import {
  ppmxAccountExtraItems,
  ppmxAccountHubItems,
} from '../ppmx-home/pageData'
import { usePpmxLocale } from '../ppmx-home/composables/usePpmxLocale'
import type { LocalizedText, PpmxAccountHubItem, PpmxAuthMode } from '../ppmx-home/types'
import PpmxPageChrome from '../ppmx-home/components/PpmxPageChrome.vue'
import PpmxThemeToggle from '@/components/PpmxThemeToggle.vue'

definePage({
  name: 'profile',
  meta: {
    titleKey: 'profile.title',
    shell: 'ppmx',
    tabbar: false,
  },
})

const userStore = useUserStore()
const route = useRoute()
const router = useRouter()
const { pickText } = usePpmxLocale()
const pageChromeRef = ref<InstanceType<typeof PpmxPageChrome> | null>(null)
const FREE_WITHDRAW_COUNT = 1
const WITHDRAW_FEE_RATE = 0.05
const profileText = {
  pageTitle: { es: 'Mi cuenta', en: 'My account', zh: '我的账户' },
  eyebrow: { es: 'Cuenta PP.PE', en: 'PP.PE account', zh: 'PP.PE 账户' },
  guestTitle: { es: 'Necesitas iniciar sesión', en: 'Sign in required', zh: '需要先登录' },
  guestDescription: {
    es: 'Entra para ver tu saldo, depositar, retirar e historial.',
    en: 'Sign in to view your balance, deposit, withdraw, and history.',
    zh: '登录后可查看余额、存款、提现和历史记录。',
  },
  loadingTitle: { es: 'Cargando cuenta', en: 'Loading account', zh: '正在加载账户' },
  loadingMessage: {
    es: 'Estamos obteniendo tu saldo e información de perfil.',
    en: 'We are loading your balance and profile information.',
    zh: '正在获取你的余额和资料信息。',
  },
  loadErrorTitle: { es: 'No se pudo cargar la cuenta', en: 'Could not load account', zh: '账户加载失败' },
  loadErrorMessage: {
    es: 'Revisa tu conexión e inténtalo nuevamente.',
    en: 'Check your connection and try again.',
    zh: '请检查网络后重试。',
  },
  emptyTitle: { es: 'Cuenta no disponible', en: 'Account unavailable', zh: '账户信息不可用' },
  emptyMessage: {
    es: 'No recibimos datos de tu perfil. Intenta actualizar.',
    en: 'No profile data was returned. Try refreshing.',
    zh: '暂未获取到账户资料，请刷新重试。',
  },
  retry: { es: 'Reintentar', en: 'Retry', zh: '重试' },
  login: { es: 'Iniciar sesión', en: 'Sign in', zh: '登录' },
  createAccount: { es: 'Crear cuenta', en: 'Create account', zh: '创建账户' },
  availableBalance: { es: 'Saldo disponible', en: 'Available balance', zh: '可用余额' },
  walletBalance: { es: 'Cartera electrónica', en: 'E-wallet', zh: '电子钱包' },
  walletSport: { es: 'Cartera deportiva', en: 'Sports wallet', zh: '体育钱包' },
  walletTransferTitle: { es: 'Transferir saldo', en: 'Balance transfer', zh: '余额转移' },
  walletTransferSub: {
    es: 'Mueve el saldo disponible a la cartera seleccionada.',
    en: 'Move the available balance to the selected wallet.',
    zh: '将可用余额转移至所选钱包。',
  },
  walletTransferClose: { es: 'Cerrar transferencia', en: 'Close transfer dialog', zh: '关闭余额转移弹窗' },
  walletTransferFrom: { es: 'Cartera deportiva', en: 'Sports wallet', zh: '体育钱包' },
  walletTransferTo: { es: 'A cartera', en: 'To wallet', zh: '转入钱包' },
  walletTransferAmount: { es: 'Monto a transferir', en: 'Transfer amount', zh: '转移金额' },
  walletTransferAll: { es: 'Transferir todo', en: 'Transfer all', zh: '全部转入' },
  walletTransferConfirm: { es: 'Confirmar transferencia', en: 'Confirm transfer', zh: '确认转移' },
  walletTransferSubmitting: { es: 'Transfiriendo...', en: 'Transferring...', zh: '转移中...' },
  walletTransferInvalid: { es: 'Ingresa un monto válido.', en: 'Enter a valid transfer amount.', zh: '请输入有效的转移金额。' },
  walletTransferExceed: { es: 'El monto supera el saldo deportivo.', en: 'Amount exceeds sports wallet balance.', zh: '转移金额不能超过体育钱包余额。' },
  walletTransferSuccess: { es: 'Transferencia completada', en: 'Transfer completed', zh: '余额转移成功' },
  walletTransferFailed: { es: 'No se pudo transferir el saldo.', en: 'Could not transfer balance.', zh: '余额转移失败，请稍后再试。' },
  deposit: { es: 'Depositar', en: 'Deposit', zh: '存款' },
  withdraw: { es: 'Retirar', en: 'Withdraw', zh: '提现' },
  withdrawalAccount: { es: 'Cuenta de retiro', en: 'Withdrawal account', zh: '提现账户' },
  depositSoon: { es: 'Depósitos próximamente', en: 'Deposits coming soon', zh: '存款功能即将开放' },
  withdrawSoon: { es: 'Retiros próximamente', en: 'Withdrawals coming soon', zh: '提现功能即将开放' },
  bankAccountSoon: { es: 'Cuenta bancaria próximamente', en: 'Bank account coming soon', zh: '银行卡账户即将开放' },
  bindBankSoon: { es: 'Vinculación próximamente', en: 'Linking coming soon', zh: '绑定功能即将开放' },
  comingSoon: { es: 'Próximamente', en: 'Coming soon', zh: '即将开放' },
  promoCodeTitle: { es: 'Canjear código', en: 'Redeem code', zh: '兑换优惠码' },
  promoCodeSub: {
    es: 'Ingresa tu código de 6 dígitos para recibir tu recompensa.',
    en: 'Enter your 6-digit code to receive your reward.',
    zh: '输入 6 位数字优惠码以领取奖励。',
  },
  promoCodePlaceholder: { es: '000000', en: '000000', zh: '000000' },
  promoCodeRedeem: { es: 'Canjear', en: 'Redeem', zh: '立即兑换' },
  promoCodeCancel: { es: 'Cerrar', en: 'Close', zh: '关闭' },
  promoCodeHint: {
    es: 'Cada código se puede activar una sola vez por cuenta.',
    en: 'Each code can be activated only once per account.',
    zh: '每个优惠码每个账户仅可激活一次。',
  },
  promoCodeInvalid: {
    es: 'Ingresa un código de 6 dígitos.',
    en: 'Enter a 6-digit code.',
    zh: '请输入 6 位数字优惠码。',
  },
  promoCodeSubmitting: { es: 'Canjeando', en: 'Redeeming', zh: '兑换中' },
  promoCodeSuccess: {
    es: 'Código canjeado: +{amount}',
    en: 'Code redeemed: +{amount}',
    zh: '兑换成功：+{amount}',
  },
  promoCodeFailed: {
    es: 'No se pudo canjear el código.',
    en: 'Could not redeem the code.',
    zh: '兑换失败，请检查兑换码。',
  },
  bankTitle: {
    es: 'Vincula una cuenta bancaria para retirar',
    en: 'Link a bank account to withdraw',
    zh: '绑定银行账户后即可提现',
  },
  bankDescription: {
    es: 'Agrega tu número de cuenta para poder solicitar retiros bancarios.',
    en: 'Add your bank account number to request withdrawals.',
    zh: '添加银行账号后即可发起提现。',
  },
  bind: { es: 'Vincular', en: 'Link', zh: '绑定' },
  moreEyebrow: { es: 'Más en PP.PE', en: 'More in PP.PE', zh: '更多 PP.PE' },
  downloadGuide: { es: 'Descarga y guía', en: 'Download and guide', zh: '下载与指南' },
  accountCenter: { es: 'Centro de cuenta', en: 'Account center', zh: '账户中心' },
  copyId: { es: 'Copiar ID', en: 'Copy ID', zh: '复制 ID' },
  copyIdSuccess: { es: 'ID copiado', en: 'ID copied', zh: 'ID 已复制' },
  copyIdFailure: { es: 'No se pudo copiar el ID', en: 'Could not copy ID', zh: 'ID 复制失败' },
  withdrawTitle: { es: 'Retirar saldo', en: 'Withdraw balance', zh: '提现' },
  withdrawClose: { es: 'Cerrar retiro', en: 'Close withdrawal', zh: '关闭提现弹窗' },
  withdrawAvailable: { es: 'Saldo disponible', en: 'Available balance', zh: '可用余额' },
  withdrawWagerNeed: { es: 'Apuesta requerida de', en: 'Required wager', zh: '提现所需流水' },
  withdrawWagerSuffix: { es: 'para retirar', en: 'to withdraw', zh: '可提现' },
  withdrawWagerCurrent: { es: 'Apuesta actual', en: 'Current wager', zh: '当前流水' },
  withdrawWagerMultiple: { es: 'Múltiplo de apuesta', en: 'Wager multiple', zh: '流水倍数' },
  withdrawMultDeposit: { es: 'Recarga', en: 'Deposit', zh: '充值' },
  withdrawMultBonus: { es: 'Bono', en: 'Bonus', zh: '赠送' },
  withdrawAmountLabel: { es: 'Monto a retirar', en: 'Withdrawal amount', zh: '提现金额' },
  withdrawAmountPlaceholder: { es: '$', en: '$', zh: '请输入金额' },
  withdrawFee: { es: 'Comisión de retiro', en: 'Withdrawal fee', zh: '提现手续费' },
  withdrawReceive: { es: 'Recibes', en: 'You receive', zh: '预计到账' },
  withdrawDeduct: { es: 'Total descontado', en: 'Total deducted', zh: '合计扣除' },
  withdrawAccount: { es: 'Cuenta destino', en: 'Destination account', zh: '到账账户' },
  withdrawAccountPlaceholder: { es: 'Selecciona una cuenta', en: 'Select an account', zh: '请选择提现账户' },
  withdrawAccountsLoading: { es: 'Cargando cuentas...', en: 'Loading accounts...', zh: '正在加载提现账户...' },
  withdrawAccountsEmpty: {
    es: 'Vincula una cuenta de retiro antes de solicitar el pago.',
    en: 'Link a withdrawal account before requesting a payout.',
    zh: '请先绑定提现账户后再发起提现。',
  },
  withdrawAccountsError: {
    es: 'No se pudieron cargar tus cuentas de retiro.',
    en: 'Could not load your withdrawal accounts.',
    zh: '提现账户加载失败，请重试。',
  },
  withdrawSummaryLoading: {
    es: 'Cargando saldo retirable...',
    en: 'Loading withdrawable balance...',
    zh: '正在加载可提现余额...',
  },
  withdrawSummaryError: {
    es: 'No se pudo cargar el saldo retirable.',
    en: 'Could not load withdrawable balance.',
    zh: '可提现余额加载失败，请重试。',
  },
  withdrawRetry: { es: 'Reintentar', en: 'Retry', zh: '重试' },
  withdrawGoBank: { es: 'Vincular cuenta', en: 'Link account', zh: '去绑定账户' },
  withdrawConfirm: { es: 'Confirmar retiro', en: 'Confirm withdrawal', zh: '确认提现' },
  withdrawSubmitting: { es: 'Enviando...', en: 'Submitting...', zh: '提交中...' },
  withdrawSent: { es: 'Solicitud de retiro enviada', en: 'Withdrawal request submitted', zh: '提现申请已提交' },
  withdrawOrderFailed: {
    es: 'No se pudo enviar el retiro. Intenta nuevamente.',
    en: 'Could not submit the withdrawal. Try again.',
    zh: '提现申请提交失败，请重试。',
  },
  withdrawDisabled: {
    es: 'El retiro está desactivado para tu cuenta.',
    en: 'Withdrawal is disabled for your account.',
    zh: '管理员已关闭你的提现功能。',
  },
  withdrawDisabledToast: {
    es: 'Esta cuenta no puede solicitar retiros por ahora. Contacta a soporte.',
    en: 'This account cannot request withdrawals right now. Contact support.',
    zh: '当前账户暂不能发起提现，请联系客服。',
  },
  withdrawMinToast: { es: 'Monto mínimo de retiro: S/50.00 PEN', en: 'Minimum withdrawal: S/50.00 PEN', zh: '最低提现金额为 S/50.00 PEN' },
  withdrawBalanceToast: {
    es: 'El monto supera tu saldo disponible.',
    en: 'The amount exceeds your available balance.',
    zh: '提现金额超过可用余额。',
  },
  withdrawFlowToast: {
    es: 'Completa la apuesta requerida antes de retirar.',
    en: 'Complete the required wager before withdrawing.',
    zh: '请先完成提现流水要求。',
  },
  withdrawAccountRequired: {
    es: 'Selecciona una cuenta de retiro.',
    en: 'Select a withdrawal account.',
    zh: '请选择提现账户。',
  },
  withdrawRulesTitle: { es: 'Reglas de retiro', en: 'Withdrawal rules', zh: '提现规则' },
  withdrawRule1: { es: 'Monto mínimo de retiro: 50.', en: 'Minimum withdrawal amount: 50.', zh: '最低提现金额：50。' },
  withdrawRule2: {
    es: 'Para retirar debes cumplir el requisito de apuesta correspondiente.',
    en: 'You must meet the corresponding wager requirement before withdrawing.',
    zh: '发起提现前需要满足对应流水要求。',
  },
  withdrawRule3: {
    es: 'El primer retiro diario es gratis; después puede aplicarse comisión.',
    en: 'The first daily withdrawal is free; fees may apply afterwards.',
    zh: '每日首笔提现免费，之后可能收取手续费。',
  },
  withdrawRule4: {
    es: 'Tiempo normal de acreditación: 1-3 días hábiles.',
    en: 'Normal settlement time: 1-3 business days.',
    zh: '通常到账时间：1-3 个工作日。',
  },
  withdrawRule5: {
    es: 'Asegúrate de que los datos de tu cuenta bancaria sean correctos.',
    en: 'Make sure your bank account information is correct.',
    zh: '请确认提现账户信息准确无误。',
  },
  logout: { es: 'Cerrar sesión', en: 'Sign out', zh: '退出登录' },
  logoutConfirmTitle: { es: 'Cerrar sesión', en: 'Sign out', zh: '确认退出登录' },
  logoutConfirmMessage: {
    es: 'Tendrás que iniciar sesión nuevamente para jugar y ver tu cuenta.',
    en: 'You will need to sign in again to play and view your account.',
    zh: '退出后需要重新登录才能继续游戏和查看账户。',
  },
  logoutConfirmAction: { es: 'Cerrar sesión', en: 'Sign out', zh: '退出登录' },
  logoutConfirmCancel: { es: 'Cancelar', en: 'Cancel', zh: '取消' },
  logoutSuccess: { es: 'Sesión cerrada', en: 'Signed out', zh: '已退出登录' },
} satisfies Record<string, LocalizedText>

const profileLoading = ref(false)
const profileLoadError = ref(false)
const promoModalOpen = ref(false)
const promoCode = ref('')
const socialModalOpen = ref(false)
const logoutDialogOpen = ref(false)
const rechargeModalOpen = ref(false)
const withdrawModalOpen = ref(false)
const walletTransferModalOpen = ref(false)
const walletTransferAmount = ref('')
const walletTransferTarget = ref('balance')
const walletTransferSubmitting = ref(false)
const withdrawAmount = ref('')
const withdrawAccountId = ref('')
const withdrawAccounts = ref<WithdrawAccountItem[]>([])
const withdrawAccountsLoading = ref(false)
const withdrawAccountsLoaded = ref(false)
const withdrawAccountsError = ref(false)
const withdrawSummary = ref<TgWithdrawFlowBatchSummary | null>(null)
const withdrawSummaryLoading = ref(false)
const withdrawSummaryLoaded = ref(false)
const withdrawSummaryError = ref(false)
const withdrawMinCents = 5_000
const withdrawSubmitLock = useSubmitLock(submitWithdrawOrder, { minLockMs: 600 })
const isWithdrawSubmitting = computed(() => withdrawSubmitLock.locked.value)
const promoSubmitLock = useSubmitLock(submitPromoCodeRequest, { minLockMs: 600 })
const isPromoSubmitting = computed(() => promoSubmitLock.locked.value)
const profileReady = computed(() => userStore.isLogin && Boolean(userStore.userInfo))
const accountHandle = computed(() => (
  userStore.userInfo?.nickname
  || userStore.userInfo?.username
  || '--'
))
const accountIdValue = computed(() => userStore.userInfo?.id ? String(userStore.userInfo.id) : '')
const accountId = computed(() => accountIdValue.value ? `ID ${accountIdValue.value}` : '--')
const accountBalanceCents = computed(() => toDisplayCents(userStore.userInfo?.balance))
const accountSportBalanceCents = computed(() => toDisplayCents(userStore.userInfo?.sportBalance))
const accountTotalBalanceCents = computed(() => accountBalanceCents.value + accountSportBalanceCents.value)
const accountWalletItems = computed(() => [
  {
    id: 'balance',
    label: pickText(profileText.walletBalance),
    cents: accountBalanceCents.value,
  },
  {
    id: 'sport',
    label: pickText(profileText.walletSport),
    cents: accountSportBalanceCents.value,
  },
])
const walletTransferSourceBalanceCents = computed(() => accountSportBalanceCents.value)
const walletTransferTargetOptions = computed(() => [
  {
    label: pickText(profileText.walletBalance),
    value: 'balance',
  },
])
const walletTransferAmountValue = computed(() => {
  const normalized = Number(walletTransferAmount.value)

  return Number.isFinite(normalized) && normalized > 0 ? normalized : 0
})
const walletTransferSubmitDisabled = computed(() => (
  walletTransferSubmitting.value
  ||
  walletTransferAmountValue.value <= 0
  || walletTransferAmountValue.value > fromCent(walletTransferSourceBalanceCents.value)
))
const showBankPrompt = computed(() => userStore.userInfo?.hasWithdrawAccount !== true)
const withdrawSummaryBalanceCents = computed(() =>
  withdrawSummary.value ? toDisplayCents(withdrawSummary.value.balance) : accountBalanceCents.value,
)
const withdrawHasUnfinishedBatch = computed(() =>
  withdrawSummary.value?.hasUnfinishedBatch === true,
)
const withdrawRequiredFlowCents = computed(() =>
  withdrawSummary.value ? toDisplayCents(withdrawSummary.value.requiredFlow) : 0,
)
const withdrawCompletedFlowCents = computed(() =>
  withdrawSummary.value ? toDisplayCents(withdrawSummary.value.completedFlow) : 0,
)
const withdrawRemainingFlowCents = computed(() =>
  withdrawSummary.value ? toDisplayCents(withdrawSummary.value.remainingFlow) : 0,
)
const withdrawableCents = computed(() =>
  withdrawHasUnfinishedBatch.value ? 0 : withdrawSummaryBalanceCents.value,
)
const withdrawWagerRequiredText = computed(() =>
  formatMarketMoney(withdrawRequiredFlowCents.value),
)
const withdrawWagerCurrentText = computed(() =>
  formatMarketMoney(withdrawCompletedFlowCents.value),
)
const withdrawMultiplierValue = computed(() => {
  const n = Number(withdrawSummary.value?.withdrawMultiplier)
  return Number.isFinite(n) && n > 0 ? Number(n.toFixed(2)) : 0
})
const withdrawGiftMultiplierValue = computed(() => {
  const n = Number(withdrawSummary.value?.giftMultiplier)
  return Number.isFinite(n) && n > 0 ? Number(n.toFixed(2)) : 0
})
const withdrawWagerComplete = computed(() =>
  !withdrawHasUnfinishedBatch.value,
)
const withdrawTodayWithdrawCount = computed(() => 0)
const withdrawMaxAmount = computed(() => Math.floor(withdrawableCents.value / 100))
const withdrawAmountCents = computed(() => {
  const normalized = Number(withdrawAmount.value)

  return Number.isFinite(normalized) && normalized > 0 ? Math.round(normalized * 100) : 0
})
const withdrawFeeCents = computed(() => {
  if (withdrawAmountCents.value <= 0 || withdrawTodayWithdrawCount.value < FREE_WITHDRAW_COUNT) return 0

  return Math.round(withdrawAmountCents.value * WITHDRAW_FEE_RATE)
})
const withdrawReceiveCents = computed(() =>
  Math.max(0, withdrawAmountCents.value - withdrawFeeCents.value),
)
const withdrawDisabledByAdmin = computed(() => Number(userStore.userInfo?.rebateWithdrawDisabled ?? 0) === 1)
const selectedWithdrawAccount = computed(() =>
  withdrawAccounts.value.find(account => String(account.id) === withdrawAccountId.value) ?? null,
)
const withdrawAccountOptions = computed(() =>
  withdrawAccounts.value
    .filter(account => account.countryCode === APP_COUNTRY_CODE)
    .map(account => ({
      label: formatWithdrawAccountLabel(account),
      value: String(account.id),
    })),
)
const withdrawSubmitDisabled = computed(() =>
  withdrawAccountsLoading.value
  || withdrawSummaryLoading.value
  || !withdrawSummaryLoaded.value
  || isWithdrawSubmitting.value
  || withdrawDisabledByAdmin.value
  || withdrawAmountCents.value < withdrawMinCents
  || withdrawAmountCents.value > withdrawableCents.value
  || !withdrawAccountId.value,
)
const accountAvatarText = computed(() => {
  const label = accountHandle.value.trim()

  return label && label !== '--' ? label.slice(0, 2).toUpperCase() : 'PP'
})

async function loadProfileData() {
  if (!userStore.isLogin) {
    profileLoading.value = false
    profileLoadError.value = false
    return
  }

  profileLoading.value = true
  profileLoadError.value = false

  try {
    await userStore.loadUserInfo()
  } catch {
    profileLoadError.value = true
  } finally {
    profileLoading.value = false
  }
}

function formatMarketMoney(cents: number) {
  return `${formatMoney(cents, { currency: APP_CURRENCY_SYMBOL })} ${APP_CURRENCY}`
}

function parseWithdrawAccountData(raw: string) {
  try {
    const parsed = JSON.parse(raw) as unknown
    if (!parsed || typeof parsed !== 'object') return {}

    return Object.fromEntries(
      Object.entries(parsed as Record<string, unknown>).map(([key, value]) => [key, String(value ?? '')]),
    )
  } catch {
    return {}
  }
}

function maskAccountValue(value: string) {
  const cleanValue = value.trim()
  if (!cleanValue) return ''

  const suffix = cleanValue.slice(-4)

  return suffix ? `**** ${suffix}` : cleanValue
}

function formatWithdrawAccountLabel(account: WithdrawAccountItem) {
  const data = parseWithdrawAccountData(account.accountData)
  const bankName = data.bank || data.bankName || data.Bank || data.bankCode || 'PE'
  const accountValue = data.accountNumber
    || data.routingNumber
    || data.accountNo
    || data.cardNo
    || Object.values(data).find(value => /\d{4,}/.test(value))
    || String(account.id)
  const masked = maskAccountValue(accountValue)

  return masked ? `${bankName} · ${masked}` : bankName
}

function syncDefaultWithdrawAccount() {
  if (withdrawAccountId.value && withdrawAccountOptions.value.some(option => option.value === withdrawAccountId.value)) {
    return
  }

  const defaultAccount = withdrawAccounts.value.find(account => account.countryCode === APP_COUNTRY_CODE && account.isDefault === 1)
    ?? withdrawAccounts.value.find(account => account.countryCode === APP_COUNTRY_CODE)

  withdrawAccountId.value = defaultAccount ? String(defaultAccount.id) : ''
}

async function loadWithdrawAccounts() {
  withdrawAccountsLoading.value = true
  withdrawAccountsError.value = false

  try {
    withdrawAccounts.value = await getWithdrawAccounts()
    withdrawAccountsLoaded.value = true
    syncDefaultWithdrawAccount()
  } catch {
    withdrawAccountsError.value = true
    withdrawAccounts.value = []
    withdrawAccountId.value = ''
  } finally {
    withdrawAccountsLoading.value = false
  }
}

async function loadWithdrawSummary() {
  withdrawSummaryLoading.value = true
  withdrawSummaryError.value = false

  try {
    withdrawSummary.value = await getCurrentTgWithdrawFlowBatchSummary()
    withdrawSummaryLoaded.value = true
  } catch {
    withdrawSummary.value = null
    withdrawSummaryLoaded.value = false
    withdrawSummaryError.value = true
  } finally {
    withdrawSummaryLoading.value = false
  }
}

function loadWithdrawModalData() {
  if (!withdrawSummaryLoading.value) {
    void loadWithdrawSummary()
  }

  if (!withdrawAccountsLoaded.value && !withdrawAccountsLoading.value) {
    void loadWithdrawAccounts()
  }
}

function openAuth(mode: PpmxAuthMode) {
  pageChromeRef.value?.openAuth(mode)
}

function requireRegisterAuth() {
  if (userStore.isLogin) return true

  pageChromeRef.value?.openAuth('register')
  return false
}

function showUpcoming(message: LocalizedText = profileText.comingSoon) {
  if (!requireRegisterAuth()) return

  showToast(pickText(message))
}

function openRecharge() {
  if (!requireRegisterAuth()) return false

  rechargeModalOpen.value = true
  return true
}

function openWalletTransfer(walletId: string) {
  if (walletId === 'sport') {
    walletTransferAmount.value = ''
    walletTransferTarget.value = 'balance'
    walletTransferModalOpen.value = true
  }
}

function closeWalletTransferModal() {
  walletTransferModalOpen.value = false
}

function transferAllSportBalance() {
  const available = fromCent(walletTransferSourceBalanceCents.value)
  walletTransferAmount.value = available > 0 ? available.toFixed(2) : ''
}

async function confirmWalletTransfer() {
  if (walletTransferAmountValue.value <= 0) {
    showToast(pickText(profileText.walletTransferInvalid))
    return
  }

  if (walletTransferAmountValue.value > fromCent(walletTransferSourceBalanceCents.value)) {
    showToast(pickText(profileText.walletTransferExceed))
    return
  }

  if (walletTransferSubmitting.value) return

  walletTransferSubmitting.value = true

  try {
    const resp = await transferTgWalletBalance({
      amount: Number(walletTransferAmountValue.value.toFixed(2)),
      fromWallet: 'sport',
      toWallet: 'balance',
    })

    if (userStore.userInfo) {
      userStore.userInfo.balance = resp.balance
      userStore.userInfo.sportBalance = resp.sportBalance
    }

    showToast(pickText(profileText.walletTransferSuccess))
    closeWalletTransferModal()
  } catch (error) {
    const message = error instanceof Error && error.message.trim()
      ? error.message
      : pickText(profileText.walletTransferFailed)

    showToast(message)
  } finally {
    walletTransferSubmitting.value = false
  }
}

function closeRechargeModal() {
  rechargeModalOpen.value = false

  if (route.query.recharge === 'open') {
    const query = { ...route.query }
    delete query.recharge
    void router.replace({ query })
  }
}

function openWithdrawModal() {
  if (!requireRegisterAuth()) return
  if (withdrawDisabledByAdmin.value) {
    showToast(pickText(profileText.withdrawDisabledToast))
    return
  }

  withdrawModalOpen.value = true
  loadWithdrawModalData()
}

function closeWithdrawModal() {
  withdrawModalOpen.value = false
}

function openBankFromWithdraw() {
  closeWithdrawModal()
  openBank()
}

function validateWithdrawSubmit() {
  if (withdrawDisabledByAdmin.value) {
    return pickText(profileText.withdrawDisabledToast)
  }

  if (withdrawAmountCents.value < withdrawMinCents) {
    return pickText(profileText.withdrawMinToast)
  }

  if (!withdrawSummaryLoaded.value || withdrawSummaryError.value) {
    return pickText(profileText.withdrawSummaryError)
  }

  if (withdrawHasUnfinishedBatch.value || withdrawRemainingFlowCents.value > 0) {
    return pickText(profileText.withdrawFlowToast)
  }

  if (withdrawAmountCents.value > withdrawableCents.value) {
    return pickText(profileText.withdrawBalanceToast)
  }

  if (!withdrawAccountId.value) {
    return pickText(profileText.withdrawAccountRequired)
  }

  return ''
}

function getWithdrawSubmitErrorMessage(error: unknown) {
  const message = error instanceof Error ? error.message.trim() : ''

  if (!message) return pickText(profileText.withdrawOrderFailed)
  if (message === 'user_balance_insufficient') return pickText(profileText.withdrawBalanceToast)
  if (message === 'withdraw_disabled') return pickText(profileText.withdrawDisabledToast)
  if (message === 'invalid_withdraw_amount') return pickText(profileText.withdrawMinToast)
  if (message === 'withdraw_flow_insufficient') return pickText(profileText.withdrawFlowToast)
  if (message === 'account_not_found' || message === 'account_country_mismatch') {
    return pickText(profileText.withdrawAccountsError)
  }

  return message
}

function buildWithdrawFieldValues() {
  const account = selectedWithdrawAccount.value
  if (!account) return undefined

  const entries = Object.entries(parseWithdrawAccountData(account.accountData))
    .map(([key, value]) => [key, value.trim()] as const)
    .filter(([, value]) => value)

  return entries.length ? Object.fromEntries(entries) : undefined
}

function buildWithdrawPayload(): CreateWithdrawOrderReq {
  return {
    amount: Number((withdrawAmountCents.value / 100).toFixed(2)),
    countryCode: APP_COUNTRY_CODE,
    accountId: Number(withdrawAccountId.value),
    fieldValues: buildWithdrawFieldValues(),
  }
}

async function submitWithdrawOrder() {
  const message = validateWithdrawSubmit()
  if (message) {
    showToast(message)
    return
  }

  try {
    const order = await createWithdrawOrderV2(buildWithdrawPayload())
    closeWithdrawModal()
    showToast(order.orderNo ? `${pickText(profileText.withdrawSent)} ${order.orderNo}` : pickText(profileText.withdrawSent))

    await Promise.allSettled([
      userStore.loadUserInfo(),
      loadWithdrawSummary(),
    ])
  } catch (error) {
    showToast(getWithdrawSubmitErrorMessage(error))
    void loadWithdrawSummary()
  }
}

function submitWithdraw() {
  void withdrawSubmitLock.run().catch(() => {})
}

function openBank() {
  void router.push('/bank')
}

function openPromoModal() {
  if (!requireRegisterAuth()) return

  promoModalOpen.value = true
}

function normalizePromoCode(value: string) {
  return value.replace(/\D/g, '').slice(0, 6)
}

function updatePromoCode(event: Event) {
  const input = event.target as HTMLInputElement | null
  promoCode.value = normalizePromoCode(input?.value ?? '')
}

function submitPromoCode() {
  void promoSubmitLock.run().catch(() => {})
}

async function submitPromoCodeRequest() {
  const normalizedCode = normalizePromoCode(promoCode.value)
  promoCode.value = normalizedCode

  if (normalizedCode.length !== 6) {
    showToast(pickText(profileText.promoCodeInvalid))
    return
  }

  try {
    const result = await redeemExchangeCode(normalizedCode)
    promoModalOpen.value = false
    promoCode.value = ''
    await userStore.loadUserInfo()
    const amount = formatMarketMoney(toDisplayCents(result.amount))
    showToast(pickText(profileText.promoCodeSuccess).replace('{amount}', amount))
  } catch (error) {
    const message = error instanceof Error ? error.message : pickText(profileText.promoCodeFailed)
    showToast(message || pickText(profileText.promoCodeFailed))
  }
}

function openRecords() {
  void router.push('/records')
}

function openSecurity() {
  void router.push('/security')
}

function openVip() {
  void router.push('/vip')
}

function openUserProfile() {
  void router.push('/user-profile')
}

function openDownload() {
  void router.push('/download')
}

function openGuide() {
  void router.push('/guide')
}

function openHubItem(item: PpmxAccountHubItem) {
  if (item.id === 'records') {
    openRecords()
    return
  }

  if (item.id === 'profile') {
    openUserProfile()
    return
  }

  if (item.id === 'security') {
    openSecurity()
    return
  }

  if (item.id === 'vip') {
    openVip()
    return
  }

  if (item.id === 'support') {
    pageChromeRef.value?.openSupport()
    return
  }

  if (item.id === 'promo-code') {
    openPromoModal()
    return
  }

  if (item.id === 'social') {
    socialModalOpen.value = true
    return
  }

  showUpcoming()
}

function openExtraItem(item: PpmxAccountHubItem) {
  if (item.id === 'download') {
    openDownload()
    return
  }

  if (item.id === 'guide') {
    openGuide()
    return
  }

  showUpcoming()
}

async function copyAccountId() {
  if (!accountIdValue.value) {
    showToast(pickText(profileText.copyIdFailure))
    return
  }

  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(accountIdValue.value)
    } else {
      const input = document.createElement('input')
      input.value = accountIdValue.value
      input.setAttribute('readonly', '')
      input.style.position = 'fixed'
      input.style.opacity = '0'
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
    }

    showToast(pickText(profileText.copyIdSuccess))
  } catch {
    showToast(pickText(profileText.copyIdFailure))
  }
}

function openLogoutDialog() {
  logoutDialogOpen.value = true
}

function closeLogoutDialog() {
  logoutDialogOpen.value = false
}

function logoutAccount() {
  closeLogoutDialog()
  userStore.logout()
  profileLoading.value = false
  profileLoadError.value = false
  showToast(pickText(profileText.logoutSuccess))
  void router.push('/')
}

onMounted(() => {
  if (userStore.isLogin && !userStore.userInfo) {
    void loadProfileData()
  }
})

watch(
  () => userStore.userInfo,
  (userInfo) => {
    if (userInfo) profileLoadError.value = false
  },
)

watch(
  () => route.query.auth,
  (auth) => {
    if (auth === 'login') openAuth('login')
    if (auth === 'register') openAuth('register')
  },
  { immediate: true },
)

watch(
  () => route.query.recharge,
  (recharge) => {
    if (recharge !== 'open') return

    void nextTick(() => {
      if (!openRecharge()) return
    })
  },
  { immediate: true, flush: 'post' },
)

watch(withdrawModalOpen, (isOpen) => {
  if (!isOpen) withdrawAmount.value = ''
})

watch(walletTransferModalOpen, (isOpen) => {
  if (!isOpen) walletTransferAmount.value = ''
})

watch(promoModalOpen, (isOpen) => {
  if (!isOpen) promoCode.value = ''
})
</script>

<template>
  <main class="ppmx-page ppmx-subpage ppmx-account-page">
    <PpmxPageChrome ref="pageChromeRef">
      <section class="ppmx-subpage__compact ppmx-account-stack">
        <header class="ppmx-account-head">
          <div>
            <h1 class="ppmx-display">{{ pickText(profileText.pageTitle) }}</h1>
          </div>
          <PpmxThemeToggle class="ppmx-account-theme" show-toggle-label />
        </header>

        <PpmxGuestPanel
          v-if="!userStore.isLogin"
          id="accGuest"
          icon="fa-user-lock"
          :title="pickText(profileText.guestTitle)"
          :message="pickText(profileText.guestDescription)"
          :secondary-text="pickText(profileText.login)"
          :primary-text="pickText(profileText.createAccount)"
          @secondary="openAuth('login')"
          @primary="openAuth('register')"
        />

        <AppState
          v-else-if="profileLoading"
          class="ppmx-account-state"
          status="loading"
          :title="pickText(profileText.loadingTitle)"
          :message="pickText(profileText.loadingMessage)"
          skeleton-variant="profile-page"
        />

        <AppState
          v-else-if="profileLoadError"
          class="ppmx-account-state"
          status="error"
          :title="pickText(profileText.loadErrorTitle)"
          :message="pickText(profileText.loadErrorMessage)"
          :action-text="pickText(profileText.retry)"
          @retry="loadProfileData"
        />

        <AppState
          v-else-if="!profileReady"
          class="ppmx-account-state"
          status="empty"
          :title="pickText(profileText.emptyTitle)"
          :message="pickText(profileText.emptyMessage)"
          :action-text="pickText(profileText.retry)"
          @retry="loadProfileData"
        />

        <section v-else id="accView" class="ppmx-account-view">
          <div
            class="ppmx-account-idcard ppmx-glass reveal"
            role="button"
            tabindex="0"
            @click="openUserProfile"
            @keydown.enter.prevent="openUserProfile"
            @keydown.space.prevent="openUserProfile"
          >
            <span class="ppmx-account-idcard__av">{{ accountAvatarText }}</span>
            <span class="ppmx-account-idcard__body">
              <span class="ppmx-account-idcard__row">
                <span class="ppmx-account-idcard__handle">
                  <i class="fa-solid fa-at" />
                  <span>{{ accountHandle }}</span>
                  <i class="fa-solid fa-circle-check" />
                </span>
                <button class="ppmx-account-vip-badge" type="button" @click.stop="openVip">
                  <i class="fa-solid fa-gem" />
                  <span>VIP</span>
                </button>
              </span>
              <button
                class="ppmx-account-uid-btn"
                type="button"
                :aria-label="pickText(profileText.copyId)"
                :disabled="!accountIdValue"
                @click.stop="copyAccountId"
              >
                <i class="fa-solid fa-id-badge" />
                <span>{{ accountId }}</span>
                <i class="fa-regular fa-copy" />
              </button>
            </span>
            <i class="fa-solid fa-chevron-right ppmx-account-idcard__chevron" />
          </div>

          <section class="ppmx-account-balance reveal">
            <div class="ppmx-account-balance__glow" />
            <div class="ppmx-account-balance__content">
              <span id="accAvatar" class="ppmx-account-avatar">{{ accountAvatarText }}</span>
              <p class="ppmx-eyebrow">{{ pickText(profileText.availableBalance) }}</p>
              <PpmxMoneyText id="balMain" :cents="accountTotalBalanceCents" :suffix="APP_CURRENCY" />
              <small>
                <i class="fa-solid fa-at" />
                <span id="accHandle">{{ accountHandle }}</span>
                <button
                  id="accIdCopy"
                  class="ppmx-account-id-copy"
                  type="button"
                  :aria-label="pickText(profileText.copyId)"
                  :disabled="!accountIdValue"
                  @click="copyAccountId"
                >
                  <span>{{ accountId }}</span>
                  <i class="fa-solid fa-copy" />
                </button>
              </small>
            </div>
            <div class="ppmx-account-wallets" :aria-label="pickText(profileText.availableBalance)">
              <div
                v-for="wallet in accountWalletItems"
                :key="wallet.id"
                class="ppmx-account-wallet-row"
              >
                <span class="ppmx-account-wallet-row__icon">
                  <i class="fa-solid fa-wallet" />
                </span>
                <span class="ppmx-account-wallet-row__label">{{ wallet.label }}</span>
                <PpmxMoneyText
                  class="ppmx-account-wallet-row__amount"
                  :cents="wallet.cents"
                  :suffix="APP_CURRENCY"
                />
                <button
                  v-if="wallet.id === 'sport'"
                  class="ppmx-account-wallet-transfer-btn"
                  type="button"
                  :aria-label="pickText(profileText.walletTransferTitle)"
                  @click="openWalletTransfer(wallet.id)"
                >
                  <i class="fa-solid fa-right-left" />
                </button>
              </div>
            </div>
            <div class="ppmx-account-actions">
              <button type="button" @click="openRecharge">
                <i class="fa-solid fa-plus" />
                <span>{{ pickText(profileText.deposit) }}</span>
              </button>
              <button
                type="button"
                :disabled="withdrawDisabledByAdmin"
                @click="openWithdrawModal"
              >
                <i class="fa-solid fa-money-bill-transfer" />
                <span>{{ pickText(profileText.withdraw) }}</span>
              </button>
              <button type="button" @click="openBank">
                <i class="fa-solid fa-building-columns" />
                <span>{{ pickText(profileText.withdrawalAccount) }}</span>
              </button>
            </div>
          </section>

          <section v-if="showBankPrompt" id="bankPrompt" class="ppmx-account-bank-prompt reveal">
            <span class="ppmx-account-bank-prompt__icon">
              <i class="fa-solid fa-building-columns" />
            </span>
            <div class="ppmx-account-bank-prompt__content">
              <strong>{{ pickText(profileText.bankTitle) }}</strong>
              <p>{{ pickText(profileText.bankDescription) }}</p>
            </div>
            <button class="ppmx-btn-gold ppmx-account-bank-prompt__cta" type="button" @click="openBank">
              {{ pickText(profileText.bind) }}
            </button>
          </section>

          <section class="ppmx-account-hub" :aria-label="pickText(profileText.accountCenter)">
            <PpmxActionCard
              v-for="item in ppmxAccountHubItems"
              :key="item.id"
              base-class="ppmx-account-hub-card reveal"
              :card-class="{ 'ppmx-account-hub-card--promo': item.id === 'promo-code' }"
              :icon="item.icon"
              :tone="item.tone"
              :title="pickText(item.label)"
              :description="pickText(item.description)"
              @click="openHubItem(item)"
            />
          </section>

          <section class="ppmx-account-extra reveal">
            <div class="ppmx-account-extra__head">
              <span class="ppmx-eyebrow">{{ pickText(profileText.moreEyebrow) }}</span>
              <h2 class="ppmx-display">{{ pickText(profileText.downloadGuide) }}</h2>
            </div>
            <div class="ppmx-account-extra__grid">
              <PpmxActionCard
                v-for="item in ppmxAccountExtraItems"
                :key="item.id"
                base-class="ppmx-account-extra-card"
                :icon="item.icon"
                :tone="item.tone"
                :title="pickText(item.label)"
                :description="pickText(item.description)"
                @click="openExtraItem(item)"
              />
            </div>
          </section>

          <button class="ppmx-account-logout ppmx-glass reveal" type="button" @click="openLogoutDialog">
            <i class="fa-solid fa-arrow-right-from-bracket" />
            <span>{{ pickText(profileText.logout) }}</span>
          </button>
        </section>
      </section>
    </PpmxPageChrome>

    <PpmxDialog
      v-model="promoModalOpen"
      variant="gold"
      icon="fa-ticket"
      :title="pickText(profileText.promoCodeTitle)"
      :confirm-text="isPromoSubmitting ? pickText(profileText.promoCodeSubmitting) : pickText(profileText.promoCodeRedeem)"
      :cancel-text="pickText(profileText.promoCodeCancel)"
      :confirm-loading="isPromoSubmitting"
      :aria-label="pickText(profileText.promoCodeTitle)"
      @confirm="submitPromoCode"
    >
      <div id="promoModal" class="ppmx-promo-code-modal">
        <p>{{ pickText(profileText.promoCodeSub) }}</p>
        <input
          id="promoInput"
          :value="promoCode"
          class="ppmx-promo-code-input"
          type="text"
          inputmode="numeric"
          maxlength="6"
          autocomplete="off"
          :placeholder="pickText(profileText.promoCodePlaceholder)"
          aria-label="Promo code"
          @input="updatePromoCode"
          @keydown.enter.prevent="submitPromoCode"
        >
        <small>{{ pickText(profileText.promoCodeHint) }}</small>
      </div>
    </PpmxDialog>

    <PpmxDialog
      v-model="logoutDialogOpen"
      variant="danger"
      icon="fa-arrow-right-from-bracket"
      :title="pickText(profileText.logoutConfirmTitle)"
      :message="pickText(profileText.logoutConfirmMessage)"
      :confirm-text="pickText(profileText.logoutConfirmAction)"
      :cancel-text="pickText(profileText.logoutConfirmCancel)"
      :aria-label="pickText(profileText.logoutConfirmTitle)"
      @confirm="logoutAccount"
      @cancel="closeLogoutDialog"
      @close="closeLogoutDialog"
    />

    <PpmxRechargeModal
      v-model="rechargeModalOpen"
      @close="closeRechargeModal"
    />

     <PpmxSocialModal v-model="socialModalOpen" />

     <PpmxWithdrawModal
       id="walletTransferModal"
       v-model="walletTransferModalOpen"
       title-id="walletTransferModalTitle"
       :title="pickText(profileText.walletTransferTitle)"
       :close-aria-label="pickText(profileText.walletTransferClose)"
       @close="closeWalletTransferModal"
     >
       <form class="ppmx-account-wallet-transfer" @submit.prevent="confirmWalletTransfer">
         <p class="ppmx-account-wallet-transfer__sub">{{ pickText(profileText.walletTransferSub) }}</p>

         <section class="ppmx-account-wallet-transfer__source">
           <span>{{ pickText(profileText.walletTransferFrom) }}</span>
           <PpmxMoneyText :cents="walletTransferSourceBalanceCents" :suffix="APP_CURRENCY" />
         </section>

         <div class="ppmx-account-wallet-transfer__arrow">
           <i class="fa-solid fa-arrow-down" />
         </div>

         <label class="ppmx-withdraw-field">
           <span>{{ pickText(profileText.walletTransferTo) }}</span>
           <PpmxSelectInput
             v-model="walletTransferTarget"
             :options="walletTransferTargetOptions"
             :title="pickText(profileText.walletTransferTo)"
           />
         </label>

         <label class="ppmx-withdraw-field">
           <span>{{ pickText(profileText.walletTransferAmount) }}</span>
           <div class="ppmx-account-wallet-transfer__amount-row">
             <input
               v-model="walletTransferAmount"
               type="number"
               inputmode="decimal"
               min="0.01"
               :max="fromCent(walletTransferSourceBalanceCents)"
               step="0.01"
               placeholder="S/0.00"
             >
             <button type="button" class="ppmx-account-wallet-transfer__all" @click="transferAllSportBalance">
               {{ pickText(profileText.walletTransferAll) }}
             </button>
           </div>
         </label>

         <button
           class="ppmx-btn-gold ppmx-account-wallet-transfer__submit"
           type="submit"
           :disabled="walletTransferSubmitDisabled"
         >
           {{ walletTransferSubmitting
             ? pickText(profileText.walletTransferSubmitting)
             : pickText(profileText.walletTransferConfirm) }}
         </button>
       </form>
     </PpmxWithdrawModal>

     <PpmxWithdrawModal
       id="withdrawModal"
      v-model="withdrawModalOpen"
      :eyebrow="APP_CURRENCY"
      title-id="withdrawModalTitle"
      :title="pickText(profileText.withdrawTitle)"
      :close-aria-label="pickText(profileText.withdrawClose)"
      @close="closeWithdrawModal"
    >
      <form class="ppmx-withdraw-form" @submit.prevent="submitWithdraw">
        <div class="ppmx-withdraw-balance">
          <PpmxSkeleton
            v-if="withdrawSummaryLoading"
            variant="withdraw-modal"
            :rows="2"
            compact
            :aria-label="pickText(profileText.withdrawSummaryLoading)"
          />
          <template v-else>
            <span>{{ pickText(profileText.withdrawAvailable) }}</span>
            <strong>{{ formatMarketMoney(withdrawableCents) }}</strong>
          </template>
        </div>

        <div
          v-if="withdrawSummaryError"
          class="ppmx-withdraw-account-note is-error"
        >
          <i class="fa-solid fa-triangle-exclamation" />
          <span>{{ pickText(profileText.withdrawSummaryError) }}</span>
          <button type="button" @click="loadWithdrawSummary">
            {{ pickText(profileText.withdrawRetry) }}
          </button>
        </div>

        <div
          v-if="withdrawDisabledByAdmin"
          class="ppmx-withdraw-account-note is-error"
        >
          <i class="fa-solid fa-circle-exclamation" />
          <span>{{ pickText(profileText.withdrawDisabled) }}</span>
        </div>

        <div class="wd-wager">
          <div>
            <p class="wd-wager-t">
              {{ pickText(profileText.withdrawWagerNeed) }}
              <b class="wd-hl">{{ withdrawWagerRequiredText }}</b>
              {{ pickText(profileText.withdrawWagerSuffix) }}
            </p>
            <p class="wd-wager-cur">
              {{ pickText(profileText.withdrawWagerCurrent) }}
              <b :class="withdrawWagerComplete ? 'wd-hl-ok' : 'wd-hl'">{{ withdrawWagerCurrentText }}</b>
            </p>
            <p v-if="withdrawMultiplierValue > 0 || withdrawGiftMultiplierValue > 0" class="wd-wager-cur">
              {{ pickText(profileText.withdrawWagerMultiple) }}：
              <template v-if="withdrawMultiplierValue > 0">
                {{ pickText(profileText.withdrawMultDeposit) }} <b class="wd-hl">×{{ withdrawMultiplierValue }}</b>
              </template>
              <template v-if="withdrawGiftMultiplierValue > 0">
                · {{ pickText(profileText.withdrawMultBonus) }} <b class="wd-hl">×{{ withdrawGiftMultiplierValue }}</b>
              </template>
            </p>
          </div>
          <span class="wd-badge" :class="withdrawWagerComplete ? 'wd-badge--ok' : 'wd-badge--bad'">
            <i class="fa-solid" :class="withdrawWagerComplete ? 'fa-check' : 'fa-triangle-exclamation'" />
          </span>
        </div>

        <label class="ppmx-withdraw-field">
          <span>{{ pickText(profileText.withdrawAmountLabel) }}</span>
          <input
            v-model="withdrawAmount"
            type="number"
            inputmode="decimal"
            min="50"
            :max="withdrawMaxAmount"
            step="0.01"
            required
            :placeholder="pickText(profileText.withdrawAmountPlaceholder)"
          >
        </label>

        <div class="wd-fees">
          <div class="wd-fee-row">
            <span class="wd-fee-k">{{ pickText(profileText.withdrawFee) }}</span>
            <span class="wd-fee-v">{{ formatMarketMoney(withdrawFeeCents) }}</span>
          </div>
          <div class="wd-fee-row">
            <span class="wd-fee-k">{{ pickText(profileText.withdrawReceive) }}</span>
            <span class="wd-fee-v wd-fee-v--ok">{{ formatMarketMoney(withdrawReceiveCents) }}</span>
          </div>
          <div class="wd-fee-div" />
          <div class="wd-fee-row">
            <span class="wd-fee-k is-strong">{{ pickText(profileText.withdrawDeduct) }}</span>
            <span class="wd-fee-v--gold">{{ formatMarketMoney(withdrawAmountCents) }}</span>
          </div>
        </div>

        <label class="ppmx-withdraw-field">
          <span>{{ pickText(profileText.withdrawAccount) }}</span>
          <PpmxSkeleton
            v-if="withdrawAccountsLoading && !withdrawAccountOptions.length"
            variant="withdraw-modal"
            :rows="1"
            compact
            :aria-label="pickText(profileText.withdrawAccountsLoading)"
          />
          <PpmxSelectInput
            v-else
            id="wdAccount"
            v-model="withdrawAccountId"
            :options="withdrawAccountOptions"
            :placeholder="pickText(profileText.withdrawAccountPlaceholder)"
            :title="pickText(profileText.withdrawAccount)"
            :disabled="withdrawAccountsLoading || !withdrawAccountOptions.length"
          />
        </label>

        <div
          v-if="withdrawAccountsError || (!withdrawAccountsLoading && !withdrawAccountOptions.length)"
          class="ppmx-withdraw-account-note"
          :class="{ 'is-error': withdrawAccountsError }"
        >
          <i class="fa-solid" :class="withdrawAccountsError ? 'fa-triangle-exclamation' : 'fa-link'" />
          <span>
            {{ pickText(withdrawAccountsError ? profileText.withdrawAccountsError : profileText.withdrawAccountsEmpty) }}
          </span>
          <button type="button" @click="openBankFromWithdraw">
            {{ pickText(profileText.withdrawGoBank) }}
          </button>
        </div>

        <button
          class="ppmx-btn-gold ppmx-withdraw-submit"
          type="submit"
          :disabled="withdrawSubmitDisabled"
        >
          {{ isWithdrawSubmitting
            ? pickText(profileText.withdrawSubmitting)
            : pickText(profileText.withdrawConfirm) }}
        </button>
      </form>

      <div class="wd-rules">
        <h3 class="wd-rules-t">
          <i class="fa-solid fa-circle-info" />
          <span>{{ pickText(profileText.withdrawRulesTitle) }}</span>
        </h3>
        <ul class="wd-rules-list">
          <li>{{ pickText(profileText.withdrawRule1) }}</li>
          <li>{{ pickText(profileText.withdrawRule2) }}</li>
          <li>{{ pickText(profileText.withdrawRule3) }}</li>
          <li>{{ pickText(profileText.withdrawRule4) }}</li>
          <li>{{ pickText(profileText.withdrawRule5) }}</li>
        </ul>
      </div>
    </PpmxWithdrawModal>
  </main>
</template>
