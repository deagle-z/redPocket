import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const css = readFileSync(new URL('./pp-mx-home.css', import.meta.url), 'utf8') as string
const themeCss = readFileSync(new URL('./theme.css', import.meta.url), 'utf8') as string
const mainCss = readFileSync(new URL('./main.css', import.meta.url), 'utf8') as string
const indexHtml = readFileSync(new URL('../../../index.html', import.meta.url), 'utf8') as string
const offlineHtml = readFileSync(new URL('../../../public/offline.html', import.meta.url), 'utf8') as string
const publicLogoSvg = readFileSync(new URL('../../../public/images/ppmx/logo.svg', import.meta.url), 'utf8') as string
const symbolsSource = readFileSync(
  new URL('../../pages/ppmx-home/components/PpmxSymbols.vue', import.meta.url),
  'utf8',
) as string
const dialogSource = readFileSync(new URL('../../components/PpmxDialog.vue', import.meta.url), 'utf8') as string
const supportSource = readFileSync(new URL('../../components/PpmxSupportModal.vue', import.meta.url), 'utf8') as string
const selectInputSource = readFileSync(new URL('../../components/PpmxSelectInput.vue', import.meta.url), 'utf8') as string
const paymentNoticeSource = readFileSync(
  new URL('../../components/PpmxPaymentSuccessNotice.vue', import.meta.url),
  'utf8',
) as string
const pageChromeSource = readFileSync(
  new URL('../../pages/ppmx-home/components/PpmxPageChrome.vue', import.meta.url),
  'utf8',
) as string
const logoSource = readFileSync(
  new URL('../../pages/ppmx-home/components/PpmxLogo.vue', import.meta.url),
  'utf8',
) as string
const montserratCss = readFileSync(
  new URL('../../../public/fonts/montserrat/montserrat.css', import.meta.url),
  'utf8',
) as string

const normalizeCss = (source: string) => source.replace(/\s+/g, ' ')
const referenceFontStack = '"Montserrat", system-ui, -apple-system, "Segoe UI", sans-serif'
const legacyFontNames = ['Ari' + 'al', 'In' + 'ter', 'Ro' + 'boto']
const legacyAppleStack = '-apple-system, ' + 'BlinkMacSystemFont'
const legacyOfflineHeadingColor = '#111' + '827'
const legacyOfflineBodyColor = '#6b' + '7280'
const checkedTypographySources = [
  themeCss,
  css,
  indexHtml,
  offlineHtml,
  dialogSource,
  supportSource,
]


describe('PP.PE home CSS contract', () => {
  it('keeps the reference palette and content width in the namespaced home stylesheet', () => {
    expect(css).toContain('#060608')
    expect(css).toContain('#C8102E')
    expect(css).toContain('#FFD700')
    expect(css).toContain('#f5f5f7')
    expect(css).toContain('--ppmx-content-max: 1320px')
  })

  it('defines the required namespaced structural classes', () => {
    for (const className of [
      '.ppmx-page',
      '.ppmx-top-strip',
      '.ppmx-header',
      '.ppmx-logo',
      '.ppmx-drawer',
      '.ppmx-notif-bar',
      '.ppmx-hero-carousel',
      '.ppmx-game-nav',
      '.ppmx-hcat',
      '.ppmx-htile',
      '.ppmx-float-widget',
      '.ppmx-footer',
      '.ppmx-footer-brand',
      '.ppmx-footer-columns',
      '.ppmx-modal',
      '.ppmx-modal__card',
      '.ppmx-modal__backdrop',
      '.ppmx-modal__close',
      '.ppmx-auth-card',
      '.ppmx-auth-field',
      '.ppmx-auth-step-dot',
      '.ppmx-glass',
      '.ppmx-glass-strong',
      '.ppmx-btn-red',
      '.ppmx-btn-gold',
    ]) {
      expect(css).toContain(className)
    }
  })

  it('keeps hero banner images cover-cropped like the reference carousel', () => {
    expect(css).toContain('.ppmx-hero-bg')
    expect(css).toContain('background-position: right center')
    expect(css).toContain('background-size: cover')
  })

  it('preserves the reference responsive breakpoints', () => {
    expect(css).toContain('@media (max-width: 1023px)')
    expect(css).toContain('@media (max-width: 640px)')
    expect(css).toContain('@media (max-width: 380px)')
    expect(css).toContain('@media (min-width: 768px)')
    expect(css).toContain('@media (min-width: 1100px)')
  })

  it('keeps the responsive acceptance rules for mobile and desktop home', () => {
    const normalizedMain = normalizeCss(mainCss)

    expect(css).toContain('overflow-x: clip')
    expect(normalizedMain).toContain('html, body, #app { width: 100%; max-width: 100%; min-height: 100%; margin: 0; overflow-x: hidden; overflow-x: clip; overscroll-behavior: none; }')
    expect(normalizedMain).toContain('body { min-width: 320px; touch-action: pan-y;')
    expect(css).toContain('height: clamp(240px, 64vw, 280px)')
    expect(css).toContain('.ppmx-header__desktop-nav')
    expect(css).toContain('display: none;')
    expect(css).toContain('.ppmx-header__menu')
    expect(css).toContain('display: grid;')
    expect(css).toContain('top: 67px')
    expect(css).toContain('scrollbar-width: none')
  })

  it('keeps the nav menu button visible on desktop as well as mobile', () => {
    const normalizedCss = normalizeCss(css)

    expect(normalizedCss).toContain('.ppmx-header__menu { display: grid; }')
    expect(normalizedCss).not.toContain('.ppmx-header__menu { display: none; }')
  })

  it('prevents iOS keyboard zoom and keeps fixed mobile UI out of the keyboard', () => {
    const normalizedMain = normalizeCss(mainCss)
    const normalizedCss = normalizeCss(css)
    const normalizedSelectInput = normalizeCss(selectInputSource)
    const normalizedDialog = normalizeCss(dialogSource)

    expect(indexHtml).toContain('interactive-widget=resizes-content')
    expect(normalizedMain).toContain('--app-visual-height: 100dvh; --app-keyboard-inset: 0px;')
    expect(normalizedMain).toContain('@media (max-width: 1023px) and (pointer: coarse)')
    expect(normalizedMain).toContain('input, textarea, select { font-size: 16px; }')
    expect(normalizedCss).toContain('.ppmx-auth-field, .ppmx-recharge-field input, .ppmx-recharge-field textarea, .ppmx-withdraw-field input, .ppmx-bank-field input, .ppmx-bank-field textarea, .ppmx-user-profile-username input, .ppmx-user-profile-field input, .ppmx-casino-search input, .ppmx-responsible-control input, .ppmx-responsible-control select, .ppmx-promo-code-input { font-size: 16px; }')
    expect(normalizedCss).toContain('html.is-ios-keyboard-open .ppmx-mobile-tabbar, html.is-ios-keyboard-open .ppmx-floating-widgets { opacity: 0; pointer-events: none; transform: translateY(120%); }')
    expect(normalizedCss).toContain('.ppmx-modal { min-height: var(--app-visual-height, 100dvh); }')
    expect(normalizedCss).toContain('.ppmx-modal__card { max-height: calc(var(--app-visual-height, 100dvh) - 32px); }')
    expect(normalizedSelectInput).toContain('@media (max-width: 1023px) and (pointer: coarse)')
    expect(normalizedSelectInput).toContain('.ppmx-select-input__button { font-size: 16px; }')
    expect(normalizedSelectInput).toContain('max-height: min(64dvh, calc(var(--app-visual-height, 100dvh) - 24px), 520px);')
    expect(normalizedDialog).toContain('min-height: var(--app-visual-height, 100dvh);')
    expect(normalizedDialog).toContain('max-height: calc(var(--app-visual-height, 100dvh) - 32px);')
  })

  it('keeps the dark withdraw modal labels and disabled controls readable', () => {
    const normalizedCss = normalizeCss(css)

    expect(normalizedCss).toContain('.ppmx-withdraw-balance span, .wd-fee-k, .ppmx-withdraw-field > span { color: rgba(245, 245, 247, 0.7);')
    expect(normalizedCss).toContain('.wd-wager-cur { margin-top: 5px; color: rgba(245, 245, 247, 0.66);')
    expect(normalizedCss).toContain('.ppmx-withdraw-field input::placeholder { color: rgba(245, 245, 247, 0.56); }')
    expect(normalizedCss).toContain('.ppmx-withdraw-card .ppmx-select-input__button.is-empty span { color: rgba(245, 245, 247, 0.58); }')
    expect(normalizedCss).toContain('.ppmx-withdraw-card .ppmx-select-input.is-disabled .ppmx-select-input__button { opacity: 0.92; }')
    expect(normalizedCss).toContain('.ppmx-withdraw-account-note { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; gap: 10px; align-items: center; padding: 12px; color: rgba(245, 245, 247, 0.9);')
    expect(normalizedCss).toContain('.ppmx-withdraw-submit:disabled { color: rgba(255, 230, 168, 0.76); cursor: not-allowed; background: rgba(255, 215, 0, 0.12);')
  })

  it('matches the v6 game-nav-wrap light theme contract', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('.ppmx-game-nav { position: sticky; top: 67px; z-index: 30;')
    expect(normalized).toContain('background: rgba(8, 8, 11, 0.72);')
    expect(normalized).toContain('backdrop-filter: blur(14px) saturate(140%);')
    expect(normalized).toContain('.ppmx-game-nav__inner { display: flex; gap: 9px; overflow-x: auto; padding: 11px 0;')
    expect(normalized).toContain('scroll-snap-type: x proximity;')
    expect(normalized).toContain('.ppmx-game-nav__item { display: inline-flex; flex: none; gap: 8px;')
    expect(normalized).toContain('height: 40px; padding: 0 16px;')
    expect(normalized).toContain('color: rgba(255, 255, 255, 0.72);')
    expect(normalized).toContain('background: rgba(255, 255, 255, 0.05);')
    expect(normalized).toContain('border: 1px solid rgba(255, 255, 255, 0.08);')
    expect(normalized).toContain('.ppmx-game-nav__item i { color: #f0bc3d;')
    expect(normalized).toContain('.ppmx-game-nav__item.is-active { color: #1c1102; background: linear-gradient(180deg, #ffe488, #f0bc3d);')
    expect(normalized).toContain('.ppmx-game-nav__item.is-active i { color: #1c1102; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-game-nav { background: rgba(255, 255, 255, 0.82); border-bottom-color: rgba(255, 26, 26, 0.1);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-game-nav__item { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-game-nav__item i { color: #e11d2e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-game-nav__item:hover { color: #000; }')
  })

  it('keeps game tile provider labels gray in the PP.PE light theme', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('.ppmx-htile-provider { display: block; max-width: 100%; margin-top: 3px;')
    expect(normalized).toContain('color: rgba(240, 188, 61, 0.62);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-htile-name, html[data-theme="light"] .ppmx-htile-provider { color: #4a4a52; }')
  })

  it('matches the reference trust-info and footer composition', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('.ppmx-trust-section { margin-top: 54px; border-top: 1px solid rgba(255, 255, 255, 0.05);')
    expect(normalized).toContain('.ppmx-trust-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 32px; width: min(100% - 40px, var(--ppmx-content-max)); padding: 64px 0;')
    expect(normalized).toContain('.ppmx-trust-card { display: flex; flex-direction: column; gap: 12px;')
    expect(normalized).toContain('.ppmx-trust-card i { color: var(--ppmx-gold); font-size: 24px;')
    expect(normalized).toContain('.ppmx-trust-card strong { color: rgba(245, 245, 247, 0.94); font-size: 14px; font-weight: 600;')
    expect(normalized).toContain('.ppmx-trust-card p { margin: 4px 0 0; color: #71717a; font-size: 12px;')
    expect(normalized).toContain('@media (min-width: 768px) { .ppmx-hcat-grid { grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 14px; } .ppmx-trust-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-trust-section { border-top-color: rgba(255, 26, 26, 0.12); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-trust-card i { color: #e11d2e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-trust-card strong { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-trust-card p { color: #76767e; }')
    expect(normalized).toContain('.ppmx-footer { margin-top: 40px; border-top: 1px solid rgba(255, 255, 255, 0.05); }')
    expect(normalized).toContain('.ppmx-footer__inner { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 40px; width: min(100% - 40px, var(--ppmx-content-max)); padding: 64px 0;')
    expect(normalized).toContain('.ppmx-footer-brand { grid-column: span 2; min-width: 0; }')
    expect(normalized).toContain('.ppmx-footer-brand .ppmx-logo svg { width: 160px; }')
    expect(normalized).toContain('.ppmx-footer-brand p { max-width: 320px; margin: 16px 0 0; color: #71717a; font-size: 14px;')
    expect(normalized).toContain('.ppmx-footer-socials { display: flex; gap: 16px; align-items: center; margin-top: 20px; font-size: 18px;')
    expect(normalized).toContain('.ppmx-footer-columns section { min-width: 0; }')
    expect(normalized).toContain('.ppmx-footer-columns h4 { margin: 0 0 16px; color: rgba(245, 245, 247, 0.94); font-size: 14px; font-weight: 600;')
    expect(normalized).toContain('.ppmx-footer-columns ul { display: grid; gap: 10px; padding: 0; margin: 0;')
    expect(normalized).toContain('.ppmx-footer-columns button { justify-self: start; color: #71717a; font-size: 14px; font-weight: 500;')
    expect(normalized).toContain('.ppmx-footer-legal { border-top: 1px solid rgba(255, 255, 255, 0.05); }')
    expect(normalized).toContain('.ppmx-footer-legal__inner { width: min(100% - 40px, var(--ppmx-content-max)); padding: 28px 0;')
    expect(normalized).toContain('.ppmx-footer-legal p { max-width: 1024px; margin: 0; color: #71717a; font-size: 12px;')
    expect(normalized).toContain('.ppmx-footer-bottom { color: #52525b; font-size: 12px; border-top: 1px solid rgba(255, 255, 255, 0.05); }')
    expect(normalized).toContain('.ppmx-footer-bottom__inner { display: flex; flex-direction: column; gap: 12px; align-items: center; justify-content: space-between;')
    expect(normalized).toContain('.ppmx-footer-badges { display: flex; gap: 12px; align-items: center;')
    expect(normalized).toContain('.ppmx-footer-badges span { padding: 4px 10px; border: 1px solid rgba(255, 255, 255, 0.08); border-radius: 8px;')
    expect(normalized).toContain('@media (min-width: 768px) { .ppmx-hcat-grid { grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 14px; } .ppmx-trust-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } .ppmx-footer__inner { grid-template-columns: repeat(5, minmax(0, 1fr)); } .ppmx-footer-bottom__inner { flex-direction: row; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-footer { color: rgba(23, 23, 27, 0.66); border-top-color: rgba(0, 0, 0, 0.08); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-footer-brand p, html[data-theme="light"] .ppmx-footer-columns button { color: #8a8a92; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-footer-legal, html[data-theme="light"] .ppmx-footer-bottom { border-top-color: rgba(0, 0, 0, 0.08); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-footer-legal p { color: #6b6b72; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-footer-badges span { border-color: rgba(0, 0, 0, 0.08); }')
  })

  it('matches the reference floating widget motion effects', () => {
    expect(css).toContain('.ppmx-float-inner::after')
    expect(css).toContain('.ppmx-float-widget--checkin .ppmx-float-inner::after')
    expect(css).toContain('.ppmx-float-widget--social .ppmx-float-inner::after')
    expect(css).toContain('.ppmx-float-widget--download .ppmx-float-inner::after')
    expect(css).toContain('@keyframes ppmxFloatHalo')
    expect(css).toContain('@keyframes ppmxFloatIcon')
    expect(css).toContain('@keyframes ppmxFloatSlideInRight')
    expect(css).toContain('.ppmx-floating-widgets.is-scroll-hiding .ppmx-float-widget')
    expect(css).toContain('translateX(120%) scale(0.92)')
    expect(css).toContain('pointer-events: none')
    expect(css).toContain('.ppmx-floating-widgets:not(.is-scroll-hiding) .ppmx-float-widget:nth-child(2)')
    expect(css).toContain('animation-delay: 60ms')
    expect(css).toContain('animation-delay: 120ms')
    expect(css).toContain('animation: ppmxFloatIcon 3.4s ease-in-out infinite')
    expect(css).toContain('transform: scale(0.95)')
  })

  it('matches the reference login and register modal composition', () => {
    expect(css).toContain('.ppmx-modal__backdrop')
    expect(css).toContain('backdrop-filter: blur(10px)')
    expect(css).toContain('.ppmx-modal__card')
    expect(css).toContain('animation: ppmxModalIn 0.4s cubic-bezier(0.2, 0.8, 0.2, 1) both')
    expect(css).not.toContain('.ppmx-auth-socials')
    expect(css).toContain('@keyframes ppmxModalIn')
  })

  it('themes Vant dialogs with the PP.PE dark red-gold visual system', () => {
    expect(themeCss).toContain('.van-dialog')
    expect(themeCss).toContain('.van-dialog__header')
    expect(themeCss).toContain('.van-dialog__message')
    expect(themeCss).toContain('.van-dialog__footer')
    expect(themeCss).toContain('.van-dialog__confirm')
    expect(themeCss).toContain('linear-gradient(180deg, #e43b45, #b9152a)')
    expect(themeCss).toContain('border-radius: 28px')
  })

  it('themes Vant toasts as compact PP.PE premium notifications', () => {
    const normalized = normalizeCss(themeCss)

    expect(themeCss).toContain('.van-toast')
    expect(themeCss).toContain('.van-toast::before')
    expect(themeCss).toContain('.van-toast--text')
    expect(themeCss).toContain('.van-toast__icon')
    expect(themeCss).toContain('.van-toast__text')
    expect(themeCss).toContain('border-radius: 18px')
    expect(themeCss).toContain('rgba(255, 215, 0, 0.22)')
    expect(themeCss).toContain('animation: ppmxToastIn')
    expect(themeCss).toContain('@keyframes ppmxToastIn')
    expect(normalized).toContain('.van-toast--success, .van-toast--fail, .van-toast--loading { display: flex;')
    expect(normalized).toContain('flex-direction: row;')
    expect(normalized).toContain('align-items: center;')
    expect(normalized).toContain('.van-toast__icon, .van-toast__loading { flex: none; margin: 0;')
    expect(normalized).toContain('.van-toast__text { min-width: 0; margin: 0;')
  })

  it('adds the PP.PE light theme contract without changing the dark default', () => {
    expect(themeCss).toContain('html[data-theme="light"]')
    expect(themeCss).toContain('--app-bg: #ffffff')
    expect(themeCss).toContain('--van-toast-text-color: #1a1a1a')
    expect(themeCss).toContain('html[data-theme="light"] .van-dialog')
    expect(themeCss).toContain('html[data-theme="light"] .ppmx-dialog__card')
    expect(css).toContain('.ppmx-theme-toggle')
    expect(css).toContain('html[data-theme="light"] .ppmx-shell')
    expect(css).toContain('--ppmx-bg: #ffffff')
    expect(css).toContain('html[data-theme="light"] .ppmx-mobile-tabbar')
    expect(css).toContain('html[data-theme="light"] .ppmx-account-balance')
    expect(css).toContain('html[data-theme="light"] .ppmx-play-page')
    expect(css).toContain('html[data-theme="light"] .ppmx-play-title strong')
    expect(css).toContain('html[data-theme="light"] .ppmx-play-title span')
  })

  it('keeps auth error text readable in the PP.PE light theme', () => {
    const normalized = normalizeCss(css)

    expect(css).toContain('.ppmx-auth-error')
    expect(css).toContain('color: #ffd7dc')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-auth-error { color: #8f1020;')
    expect(normalized).toContain('background: rgba(200, 16, 46, 0.08);')
    expect(normalized).toContain('border-color: rgba(200, 16, 46, 0.26);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-auth-error i { color: #e11d2e; }')
  })

  it('matches the v6 light desktop chrome structure for account pages', () => {
    expect(pageChromeSource).not.toContain('class="ppmx-icon-button ppmx-header__back"')
    expect(pageChromeSource).not.toContain('showDesktopBack')
    expect(pageChromeSource).toContain(':class="{ \'is-active\': isTargetActive(item.target) }"')
    expect(pageChromeSource).toContain('class="ppmx-notification-button__label"')
    expect(pageChromeSource).toContain("{{ t('ppmx.chrome.notifications') }}")
    expect(pageChromeSource).toContain('class="ppmx-top-strip__age"')
    expect(pageChromeSource).toContain('PpmxThemeToggle class="ppmx-theme-toggle--top"')
  })

  it('matches the v6 light desktop chrome visual contract', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('.ppmx-header { position: sticky; top: 0; z-index: 40; height: 68px; min-height: 68px; background: rgba(17, 17, 20, 0.72); border-bottom: 1px solid rgba(255, 255, 255, 0.05);')
    expect(normalized).toContain('.ppmx-header__inner { position: relative; display: flex; align-items: center; gap: 24px; width: min(100% - 40px, var(--ppmx-content-max)); height: 68px;')
    expect(normalized).toContain('.ppmx-header-nav__item { display: inline-flex; gap: 8px; align-items: center; min-height: 36px; padding: 8px 14px; color: #d4d4d8; font-size: 14px; font-weight: 500;')
    expect(normalizeCss(themeCss)).toContain('--app-mode-accent: #ffd86b;')
    expect(normalizeCss(themeCss)).toContain('--app-mode-accent-strong: var(--app-gold);')
    expect(normalizeCss(themeCss)).toContain('--ppmx-mode-accent: var(--app-mode-accent);')
    expect(normalized).toContain('--ppmx-mode-accent: var(--app-mode-accent);')
    expect(normalized).toContain('--ppmx-mode-accent-strong: var(--app-mode-accent-strong);')
    expect(normalized).toContain('.ppmx-header-nav__item i { color: var(--ppmx-mode-accent); font-size: 14px;')
    expect(normalized).toContain('.ppmx-auth-button { height: 40px; padding: 0 20px; font-size: 14px; font-weight: 600; border-radius: 999px; }')
    expect(normalized).toContain('.ppmx-notification-button { position: relative; display: inline-flex; gap: 8px; align-items: center; justify-content: center; width: auto; height: 40px;')
    expect(normalized).toContain('.ppmx-logo svg { display: block; width: 150px;')
    expect(normalizeCss(themeCss)).toContain('--app-mode-accent: #ff1a1a;')
    expect(normalizeCss(themeCss)).toContain('--app-mode-accent-strong: #ff3b3b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-eyebrow, html[data-theme="light"] .ppmx-top-strip__mobile, html[data-theme="light"] .ppmx-top-strip__age, html[data-theme="light"] .ppmx-top-strip__support-button i { color: var(--ppmx-mode-accent-strong); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-top-strip__support span i { color: #34c56e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-header { background: rgba(255, 255, 255, 0.82); border-bottom-color: rgba(255, 26, 26, 0.12);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-header-nav__item i { color: var(--ppmx-mode-accent); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-header-nav__item.is-active { color: #1a1a1a; background: rgba(255, 26, 26, 0.07);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-header-nav__item.is-active i { color: var(--ppmx-mode-accent); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-auth-button.ppmx-auth-button--login { color: #c8102e; background: transparent; border-color: rgba(255, 26, 26, 0.3); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-notification-button > i { color: var(--ppmx-mode-accent); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-hcat-title i { color: var(--ppmx-mode-accent); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-rules h2 { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-rules h2 i, html[data-theme="light"] .ppmx-wheel-rules li i { color: var(--ppmx-mode-accent); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-rules li { color: rgba(23, 23, 27, 0.66); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-state { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-state h2 { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-state p { color: rgba(23, 23, 27, 0.58); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-jackpot { background: radial-gradient(80% 120% at 50% 0%, rgba(255, 26, 26, 0.1), transparent 68%), linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(255, 255, 255, 0.92)), #fff;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-jackpot span { color: var(--ppmx-mode-accent); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-wheel-jackpot strong { color: #17171b; text-shadow: none; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-bank-prompt { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-bank-prompt__icon { color: var(--ppmx-mode-accent); background: rgba(255, 26, 26, 0.08);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-bank-prompt strong { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-bank-prompt p { color: rgba(23, 23, 27, 0.68); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-bank-prompt__cta.ppmx-btn-gold { color: #fff; background: linear-gradient(135deg, #ff5a5a, #ff1a1a);')
    expect(normalized).toContain('.ppmx-account-hub { grid-template-columns: repeat(2, minmax(0, 1fr));')
    expect(normalized).toContain('.ppmx-account-hub-card { grid-template-columns: minmax(0, 1fr) auto; grid-template-rows: auto minmax(0, 1fr);')
    expect(normalized).toContain('.ppmx-account-hub-card > span:not(.ppmx-account-hub-card__icon) { grid-row: 2; grid-column: 1 / -1;')
    expect(normalized).toContain('.ppmx-account-hub-card--promo { grid-column: 1 / -1;')
    expect(normalized).toContain('.ppmx-account-hub-card--promo > span:not(.ppmx-account-hub-card__icon) { grid-row: 1; grid-column: 2;')
    expect(normalized).toContain('.ppmx-promo-code-input { width: 100%; height: 58px;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-hub-card { background:')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-hub-card--promo .ppmx-account-hub-card__icon { color: var(--ppmx-mode-accent); background: transparent; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-hub-card small { color: #76767e; }')
    expect(normalized).toContain('.ppmx-account-extra__head h2 { margin: 3px 0 0; color: #fff; font-size: 1.125rem; font-weight: 800;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-extra { color: #17171b; background: transparent; border-color: transparent; box-shadow: none; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-extra-card { color: #17171b; background:')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-extra-card small { color: #6b6b72; }')
    expect(normalized).toContain('.ppmx-account-logout { display: flex; align-items: center; justify-content: center; width: 100%; min-height: 52px;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-account-logout { color: #c8102e;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-code-input { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-code-modal small { color: rgba(23, 23, 27, 0.5); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-logo .logo-word, html[data-theme="light"] .ppmx-logo .ppmx-logo-word { fill: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-header .ppmx-logo .logo-word, html[data-theme="light"] .ppmx-header .ppmx-logo .ppmx-logo-word { fill: var(--ppmx-mode-accent); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-theme-toggle .ppmx-theme-toggle__icon { color: #fff; background: linear-gradient(135deg, #ff5a5a, #ff1a1a);')
    expect(normalized).toContain('.ppmx-account-page .ppmx-account-stack { width: min(100% - 40px, 768px); padding: 40px 0 96px;')
    expect(normalized).toContain('.ppmx-account-head h1 { margin: 0; color: #fff; font-size: clamp(2.25rem, 4.8vw, 3rem); font-weight: 800;')
    expect(normalized).toContain('.ppmx-account-theme .ppmx-theme-toggle__icon { display: grid; width: 24px; height: 24px;')
    expect(normalized).toContain('.ppmx-account-guest { display: grid; align-content: center; justify-items: center; min-height: 0; padding: 80px 24px;')
    expect(normalized).toContain('.ppmx-account-guest button { min-width: 114px; min-height: 46px; padding: 0 24px; font-weight: 600; font-size: 0.875rem;')
  })

  it('uses the PP.PE reference Montserrat font stack for visible non-icon text', () => {
    for (const source of checkedTypographySources) {
      const normalized = normalizeCss(source)

      expect(normalized).toContain(referenceFontStack)
      for (const fontName of legacyFontNames) {
        expect(normalized).not.toContain(fontName)
      }
      expect(normalized).not.toContain('ui-sans-serif')
      expect(normalized).not.toContain('BlinkMacSystemFont')
      expect(normalized).not.toContain(legacyAppleStack)
    }
  })

  it('self-hosts distinct Montserrat weight assets for the reference typography effect', () => {
    const requiredWeights = ['400', '500', '600', '700', '800', '900']
    const weightSources = new Map(
      Array.from(montserratCss.matchAll(/font-weight:\s*(400|500|600|700|800|900);[\s\S]*?src:\s*url\(\.\/([^)]+)\)/g))
        .map((match) => [match[1], match[2]]),
    )

    for (const weight of requiredWeights) {
      expect(weightSources.get(weight)).toMatch(/\.(?:ttf|woff2)$/)
    }

    expect(new Set(requiredWeights.map((weight) => weightSources.get(weight))).size).toBe(requiredWeights.length)
  })

  it('keeps SVG wordmarks on the PP.PE Montserrat stack without legacy font fallbacks', () => {
    for (const source of [publicLogoSvg, symbolsSource]) {
      expect(source).toContain('font-family="Montserrat, system-ui, -apple-system, Segoe UI, sans-serif"')
      expect(source).not.toContain(legacyFontNames[0])
    }

    expect(logoSource).toContain('viewBox="0 0 240 64"')
    expect(logoSource).toContain('width="150" height="40"')
    expect(logoSource).toContain('font-weight="600" font-size="6" letter-spacing="3" fill="#9a9aa3"')
  })

  it('aligns dark and light text tokens with the reference HTML palette', () => {
    expect(themeCss).toContain('--app-text: #f5f5f7')
    expect(themeCss).toContain('--app-muted: rgba(245, 245, 247, 0.72)')
    expect(themeCss).toContain('--app-soft: rgba(245, 245, 247, 0.52)')
    expect(css).toContain('--ppmx-text: #f5f5f7')
    expect(css).toContain('--ppmx-muted: rgba(245, 245, 247, 0.72)')
    expect(css).toContain('--ppmx-soft: rgba(245, 245, 247, 0.52)')
    expect(themeCss).toContain('--app-text: #1a1a1a')
    expect(themeCss).toContain('--app-muted: #5a5a62')
    expect(themeCss).toContain('--app-soft: #76767e')
    expect(css).toContain('--ppmx-text: #1a1a1a')
    expect(css).toContain('--ppmx-muted: #5a5a62')
    expect(css).toContain('--ppmx-soft: #76767e')
  })

  it('maps PP.PE light utility colors to the v6 readable grayscale and red accent scale', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('html[data-theme="light"] .ppmx-shell .text-white { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-shell .text-zinc-300 { color: #3a3a40; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-shell .text-zinc-400 { color: #5a5a62; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-shell .text-zinc-500 { color: #76767e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-shell .text-oro-400')
    expect(normalized).toContain('color: #e11d2e;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-shell .bg-white\\/5 { background-color: rgba(255, 26, 26, 0.04); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-shell .border-white\\/8 { border-color: rgba(0, 0, 0, 0.1); }')
  })

  it('keeps light-mode dark promotional surfaces readable instead of white-on-white', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-card { color: #fff;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-card .ppmx-display, html[data-theme="light"] .ppmx-promo-detail .ppmx-display { color: #fff; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-detail { color: #fff;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-invite h2')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-vip-current h2')
    expect(normalized).toContain('color: var(--ppmx-text);')
  })

  it('matches the v6 promo page list/detail structure and readable light detail surface', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('.ppmx-promo-stack { display: grid; gap: 0; }')
    expect(normalized).toContain('.ppmx-promo-list { display: block; }')
    expect(normalized).toContain('.ppmx-promo-list.is-hidden { display: none; }')
    expect(normalized).toContain('.ppmx-promo-head { max-width: 720px; margin-bottom: 40px; }')
    expect(normalized).toContain('.ppmx-promo-detail { position: relative; display: grid;')
    expect(normalized).toContain('margin-bottom: 48px;')
    expect(normalized).toContain('.ppmx-promo-back { display: inline-flex; gap: 8px; align-items: center; margin-bottom: 20px;')
    expect(normalized).toContain('.ppmx-promo-detail__media { position: absolute; inset: 0; z-index: 0;')
    expect(normalized).toContain('opacity: 0.4;')
    expect(normalized).toContain('background: linear-gradient(90deg, rgba(6, 6, 8, 0.94), rgba(6, 6, 8, 0.7) 55%, rgba(74, 4, 17, 0.35));')
    expect(normalized).toContain('.ppmx-promo-highlight { color: #ffd24a; font-weight: 800;')
    expect(normalized).toContain('.ppmx-promo-highlight--percent { color: #ff6b74;')
    expect(normalized).toContain('.ppmx-promo-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 20px;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-detail { color: #fff;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-detail .ppmx-promo-highlight { color: #ffd700; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-detail .ppmx-promo-highlight--percent { color: #ff6b74; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-back { color: #5a5a62; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-promo-back:hover { color: #1a1a1a; }')
  })

  it('keeps the team page readable in the v6 light theme', () => {
    const normalized = normalizeCss(css)

    expect(normalized).not.toContain('html[data-theme="light"] .ppmx-team-commission, html[data-theme="light"] .ppmx-account-guest')
    expect(normalized).toContain('.ppmx-team-invite { padding: 24px; background: rgba(255, 255, 255, 0.035); border-color: var(--ppmx-border); border-radius: 28px; box-shadow: none; }')
    expect(normalized).toContain('.ppmx-team-invite__body { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 24px; align-items: center; margin-top: 20px; }')
    expect(normalized).toContain('.ppmx-team-invite__link .ppmx-field { min-height: 48px; padding: 0 14px;')
    expect(normalized).toContain('.ppmx-team-rebate { padding: 24px; background: rgba(255, 255, 255, 0.035); border-color: var(--ppmx-border); border-radius: 28px; box-shadow: none; }')
    expect(normalized).toContain('.ppmx-team-rebate__table { display: grid; gap: 8px; margin-top: 20px; font-size: 0.75rem; }')
    expect(normalized).toContain('.ppmx-team-rebate__head { display: grid; grid-template-columns: minmax(0, 1.5fr) minmax(72px, 1fr) minmax(72px, 1fr);')
    expect(normalized).toContain('.ppmx-team-rebate__rows { display: grid; gap: 8px; margin-top: 0; }')
    expect(normalized).toContain('.ppmx-team-rebate__row { display: grid; grid-template-columns: minmax(0, 1.5fr) minmax(72px, 1fr) minmax(72px, 1fr);')
    expect(normalized).toContain('.ppmx-team-rebate__row.is-top { background: linear-gradient(135deg, #2a1d05, #3f2b08 55%, #1a1205);')
    expect(normalized).toContain('.ppmx-team-rebate__amount { display: inline-flex; justify-content: flex-end; gap: 4px;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-kpi-card { color: #1a1a1a; background: rgba(255, 255, 255, 0.74); border-color: rgba(26, 26, 32, 0.1); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-kpi-card strong, html[data-theme="light"] .ppmx-team-invite h2, html[data-theme="light"] .ppmx-team-rules summary { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-kpi-card p, html[data-theme="light"] .ppmx-team-invite label, html[data-theme="light"] .ppmx-team-qr p, html[data-theme="light"] .ppmx-team-rules p, html[data-theme="light"] .ppmx-team-rules ul { color: #76767e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-commission { color: #fff; background: linear-gradient(135deg, #241803, #3a2807 48%, #0e0e11);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-commission__stats article { background: rgba(0, 0, 0, 0.25); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-rebate { color: #1a1a1a; background: rgba(255, 255, 255, 0.74); border-color: rgba(26, 26, 32, 0.1); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-rebate__row { color: #1a1a1f; background: rgba(20, 20, 22, 0.025); border-color: rgba(20, 20, 22, 0.08); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-rebate__row.is-top { background: linear-gradient(135deg, #fff2c8, #ffd24a 60%, #f4ba1c);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-rebate__amount { color: #9a6b00; text-shadow: none; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-team-share button { color: #1a1a1a; background: rgba(255, 255, 255, 0.72); border-color: rgba(26, 26, 32, 0.1); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-field { color: #1a1a1a; background: rgba(255, 255, 255, 0.78); border-color: rgba(26, 26, 32, 0.12); }')
  })

  it('keeps remaining account, bank, recharge, and VIP surfaces readable in light mode', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('html[data-theme="light"] .ppmx-user-profile-head h1, html[data-theme="light"] .ppmx-user-profile-details__head h2, html[data-theme="light"] .ppmx-user-profile-summary strong { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-user-profile-head p, html[data-theme="light"] .ppmx-user-profile-summary p, html[data-theme="light"] .ppmx-user-profile-summary small, html[data-theme="light"] .ppmx-user-profile-username span, html[data-theme="light"] .ppmx-user-profile-field span { color: #76767e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-user-profile-username input, html[data-theme="light"] .ppmx-user-profile-field input { color: #1a1a1a; background: rgba(255, 255, 255, 0.78); border-color: rgba(26, 26, 32, 0.12); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-bank-country, html[data-theme="light"] .ppmx-bank-empty, html[data-theme="light"] .ppmx-bank-form { color: #1a1a1a; background: rgba(255, 255, 255, 0.74); border-color: rgba(26, 26, 32, 0.1); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-bank-section-head h2, html[data-theme="light"] .ppmx-bank-form__head h2 { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-bank-card { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-bank-card h2, html[data-theme="light"] .ppmx-bank-field-row strong { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-bank-country span, html[data-theme="light"] .ppmx-bank-field > span, html[data-theme="light"] .ppmx-bank-field-row span, html[data-theme="light"] .ppmx-bank-no-fields { color: #76767e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-bank-field input, html[data-theme="light"] .ppmx-bank-field textarea')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-recharge-panel__head h2, html[data-theme="light"] .ppmx-recharge-promo-card strong,')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-recharge-pay-modal__details dd { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-recharge-promo-card small, html[data-theme="light"] .ppmx-recharge-field > span,')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-recharge-pay-modal__details dt { color: #76767e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-recharge-state, html[data-theme="light"] .ppmx-bank-state { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-recharge-country { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-download-hero { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-download-hero h1 { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-recharge-error { color: #8f1020;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-withdraw-head h2, html[data-theme="light"] .ppmx-withdraw-balance strong, html[data-theme="light"] .wd-wager-t, html[data-theme="light"] .wd-fee-v, html[data-theme="light"] .wd-rules summary, html[data-theme="light"] .ppmx-withdraw-account-note { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-withdraw-balance, html[data-theme="light"] .wd-wager, html[data-theme="light"] .wd-fees, html[data-theme="light"] .wd-rules { background:')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-withdraw-balance span, html[data-theme="light"] .wd-fee-k, html[data-theme="light"] .ppmx-withdraw-field > span, html[data-theme="light"] .wd-wager-cur, html[data-theme="light"] .wd-rules ul, html[data-theme="light"] .wd-rules summary > i { color: #76767e; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-vip-current h2, html[data-theme="light"] .ppmx-vip-level-card__body strong, html[data-theme="light"] .ppmx-vip-rewards__head h2 { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-vip-progress__meta, html[data-theme="light"] .ppmx-vip-progress p, html[data-theme="light"] .ppmx-vip-level-card__body small, html[data-theme="light"] .ppmx-vip-rewards__head p, html[data-theme="light"] .ppmx-vip-reward-card p { color: #76767e; }')
  })

  it('keeps PP.PE light overlays and remaining subpages readable', () => {
    const normalized = normalizeCss(css)
    const normalizedSupport = normalizeCss(supportSource)
    const normalizedDialog = normalizeCss(dialogSource)

    expect(normalized).toContain('html[data-theme="light"] .ppmx-social-close, html[data-theme="light"] .ppmx-checkin-close { color: rgba(23, 23, 27, 0.52);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-social-card-wrap h2, html[data-theme="light"] .ppmx-checkin-body h2 { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-social-card, html[data-theme="light"] .ppmx-checkin-state, html[data-theme="light"] .ppmx-checkin-tile, html[data-theme="light"] .ppmx-checkin-records span { color: #17171b;')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-social-card.is-x, html[data-theme="light"] .ppmx-social-card.is-tiktok { --ppmx-social-rgb: 23, 23, 27; --ppmx-social-color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-checkin-tile strong, html[data-theme="light"] .ppmx-checkin-records strong { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-records-head h1, html[data-theme="light"] .ppmx-security-head h1, html[data-theme="light"] .ppmx-records-summary strong, html[data-theme="light"] .ppmx-records-card__main strong, html[data-theme="light"] .ppmx-records-card__amount > strong, html[data-theme="light"] .ppmx-download-panel h2, html[data-theme="light"] .ppmx-download-features strong, html[data-theme="light"] .ppmx-download-steps strong, html[data-theme="light"] .ppmx-vip-progress b { color: #17171b; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-records-chip, html[data-theme="light"] .ppmx-records-card, html[data-theme="light"] .ppmx-records-empty { background:')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-records-chip.is-active { color: #fff; background: linear-gradient(180deg, #ff5a5a, #ff1a1a);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-security-card:hover { background: #fff; border-color: rgba(255, 26, 26, 0.16);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-vip-progress__track { background: rgba(26, 26, 32, 0.08); border-color: rgba(26, 26, 32, 0.08); }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-vip-level-card__body em, html[data-theme="light"] .ppmx-download-steps li::before { color: #fff; background: linear-gradient(180deg, #ff5a5a, #ff1a1a); }')
    expect(normalizedSupport).toContain('html[data-theme="light"] .ppmx-support-state { background: rgba(255, 255, 255, 0.72);')
    expect(normalizedSupport).toContain('html[data-theme="light"] .ppmx-support-channel:hover { background: #fff; border-color: rgba(255, 26, 26, 0.16);')
    expect(normalizedSupport).toContain('html[data-theme="light"] .ppmx-support-safe i { color: var(--app-mode-accent, #ff1a1a); }')
    expect(normalizedDialog).toContain('html[data-theme="light"] .ppmx-dialog__close { color: rgba(23, 23, 27, 0.54);')
    expect(normalizedDialog).toContain('html[data-theme="light"] .ppmx-dialog__card.is-gold .ppmx-dialog__button--confirm { color: #fff; background: linear-gradient(180deg, #ff5a5a, #ff1a1a);')
  })

  it('keeps the casino grid loading state readable in light mode', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-grid.is-loading { opacity: 1; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-grid.is-loading .ppmx-game-tile { opacity: 1; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-grid.is-loading .ppmx-htile-thumb { background-image: radial-gradient(circle at 18% 0%, rgba(226, 31, 51, 0.08), transparent 34%), linear-gradient(145deg, rgba(34, 34, 40, 0.1), rgba(34, 34, 40, 0.045)) !important; border-color: rgba(226, 31, 51, 0.12); box-shadow: 0 10px 22px -16px rgba(36, 38, 45, 0.28); filter: none; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-grid.is-loading .ppmx-htile-badge { opacity: 0; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-grid.is-loading .ppmx-htile-name, html[data-theme="light"] .ppmx-casino-grid.is-loading .ppmx-htile-provider { color: #5a5a62; }')
  })

  it('keeps the casino loading state panel on a light surface in light mode', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-state { color: #4a4a52; background: radial-gradient(circle at top right, rgba(255, 215, 0, 0.1), transparent 32%), rgba(255, 255, 255, 0.9); border-color: rgba(226, 31, 51, 0.1);')
  })

  it('keeps casino search and sort controls readable in light mode', () => {
    const normalized = normalizeCss(css)

    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-search, html[data-theme="light"] .ppmx-casino-sort { color: #1a1a1a; background: rgba(255, 255, 255, 0.92); border-color: rgba(26, 26, 32, 0.12);')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-search input, html[data-theme="light"] .ppmx-casino-sort__trigger strong { color: #1a1a1a; }')
    expect(normalized).toContain('html[data-theme="light"] .ppmx-casino-search input::placeholder, html[data-theme="light"] .ppmx-casino-search i, html[data-theme="light"] .ppmx-casino-sort > span { color: #76767e; }')
  })

  it('keeps offline and shared overlay text on PP.PE reference colors', () => {
    expect(offlineHtml).toContain('color: #f5f5f7')
    expect(offlineHtml).toContain('color: rgba(245, 245, 247, 0.72)')
    expect(offlineHtml).toContain('font-family: "Montserrat", system-ui, -apple-system, "Segoe UI", sans-serif')
    expect(offlineHtml).not.toContain(legacyOfflineHeadingColor)
    expect(offlineHtml).not.toContain(legacyOfflineBodyColor)
    expect(selectInputSource).toContain('color: var(--ppmx-text, #f5f5f7)')
    expect(selectInputSource).toContain('html[data-theme="light"] .ppmx-select-input__button')
    expect(selectInputSource).toContain('html[data-theme="light"] .ppmx-select-input__menu')
    expect(selectInputSource).toContain('html[data-theme="light"] .ppmx-select-input__popup')
    expect(paymentNoticeSource).toContain('color: var(--app-gold)')
  })
})
