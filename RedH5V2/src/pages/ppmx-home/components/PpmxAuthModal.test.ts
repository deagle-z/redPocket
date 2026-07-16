import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const authSource = readFileSync(new URL('./PpmxAuthModal.vue', import.meta.url), 'utf8') as string

describe('PpmxAuthModal business integration', () => {
  it('uses real phone auth APIs through the user store', () => {
    expect(authSource).toContain("import { checkRegisterPhone } from '@/api/user'")
    expect(authSource).toContain("import PpmxButton from '@/components/PpmxButton.vue'")
    expect(authSource).toContain('<PpmxButton')
    expect(authSource).toContain('submitError')
    expect(authSource).toContain('role="alert"')
    expect(authSource).toContain('userStore.loginByPhone')
    expect(authSource).toContain('userStore.registerByPhone')
    expect(authSource).toContain('await checkRegisterPhone({')
    expect(authSource).toContain('buildMxPhoneWithDialCode')
    expect(authSource).not.toContain("function submitLogin() {\n  showToast")
  })

  it('checks duplicate phones when continuing from registration step one', () => {
    expect(authSource).toContain('async function checkRegisterPhoneAvailable()')
    expect(authSource).toContain('async function goRegisterStep(delta: number)')
    expect(authSource).toContain('if (delta > 0 && registerStep.value === 1) {')
    expect(authSource).toContain('if (!await checkRegisterPhoneAvailable()) return')
    expect(authSource).toContain(':loading="submitting && registerStep === 1"')
    expect(authSource).toContain('phoneRegistered')
  })

  it('binds remember me to token persistence without storing passwords locally', () => {
    expect(authSource).toContain('LOGIN_REMEMBER_PHONE_KEY')
    expect(authSource).toContain('phone: getRememberedPhone()')
    expect(authSource).toContain('remember: true')
    expect(authSource).toContain('v-model="loginForm.remember"')
    expect(authSource).toContain('remember: loginForm.remember')
    expect(authSource).toContain('saveRememberedPhone(nationalPhone)')
    expect(authSource).not.toContain('localStorage.setItem(\'password')
    expect(authSource).not.toContain('localStorage.setItem("password')
  })

  it('shows the PP.MX logo on both login and registration modals', () => {
    const logoInstances = authSource.match(/<PpmxLogo \/>/g) ?? []

    expect(logoInstances).toHaveLength(2)
    expect(authSource).toContain('<template v-if="isLogin">\n      <PpmxLogo />')
    expect(authSource).toContain('<template v-else>\n      <PpmxLogo />')
  })

  it('supports invite/source capture, attribution, and Facebook CompleteRegistration', () => {
    expect(authSource).toContain('captureInviteCode')
    expect(authSource).toContain('captureSourceChannelCode')
    expect(authSource).toContain('trackAttributionEvent')
    expect(authSource).toContain('trackCompleteRegistration')
    expect(authSource).toContain('name="inviteCode"')
    expect(authSource).toContain(':disabled="submitting"')
  })

  it('sends a browser device fingerprint during phone registration', () => {
    expect(authSource).toContain("import { getDeviceFingerprint } from '@/utils/deviceFingerprint'")
    expect(authSource).toContain('const deviceFingerprint = await getDeviceFingerprint()')
    expect(authSource).toContain('deviceFingerprint,')
  })

  it('requires password and confirmation password during registration', () => {
    expect(authSource).toContain("password: ''")
    expect(authSource).toContain("confirmPassword: ''")
    expect(authSource).toContain('name="password"')
    expect(authSource).toContain('name="confirmPassword"')
    expect(authSource).toContain('requiredConfirmPassword')
    expect(authSource).toContain('passwordMismatch')
    expect(authSource).toContain('registerForm.password !== registerForm.confirmPassword')
  })

  it('redirects to the account page after successful login and registration', () => {
    expect(authSource).toContain("const AUTH_SUCCESS_REDIRECT = '/profile'")
    expect(authSource).toContain('await router.push(AUTH_SUCCESS_REDIRECT)')
    expect(authSource.match(/await router\.push\(AUTH_SUCCESS_REDIRECT\)/g)).toHaveLength(2)
  })

  it('keeps auth phone inputs as Mexico national numbers and adds the dial code only for submit', () => {
    expect(authSource).toContain('loginForm.phone = normalizeMxNationalPhone(loginForm.phone).slice(0, 10)')
    expect(authSource).toContain('registerForm.phone = normalizeMxNationalPhone(registerForm.phone).slice(0, 10)')
    expect(authSource).toContain('function validateRegisterForm()')
    expect(authSource).toContain('validateRegisterStep(1) && validateRegisterStep(2) && validateRegisterStep(3)')
    expect(authSource).toContain('if (submitting.value || !validateRegisterForm()) return')
    expect(authSource).toContain('const phone = buildMxPhoneWithDialCode(nationalPhone)')
    expect(authSource).toContain('phone,\n      password: loginForm.password')
    expect(authSource).toContain('phone,\n      country: MX_COUNTRY_CODE')
  })

  it('does not render social login providers', () => {
    expect(authSource).not.toContain('ppmx-auth-socials')
    expect(authSource).not.toContain('aria-label="Google"')
    expect(authSource).not.toContain('aria-label="Facebook"')
    expect(authSource).not.toContain('aria-label="Apple"')
    expect(authSource).not.toContain('fa-google')
    expect(authSource).not.toContain('fa-facebook-f')
    expect(authSource).not.toContain('fa-apple')
  })
})
