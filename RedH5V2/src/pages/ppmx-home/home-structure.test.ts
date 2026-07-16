import { describe, expect, it } from 'vitest'

const { readFileSync } = await import('node:' + 'fs')
const homeSource = readFileSync(new URL('./PpmxHome.vue', import.meta.url), 'utf8') as string
const chromeSource = readFileSync(
  new URL('./components/PpmxPageChrome.vue', import.meta.url),
  'utf8',
) as string

describe('PP.MX home replica structure', () => {
  it('assembles the reference home sections through dedicated components', () => {
    for (const componentName of [
      'PpmxPageChrome',
      'PpmxNoticeBar',
      'PpmxHeroCarousel',
      'PpmxGameNav',
      'PpmxLobbySection',
      'PpmxGameTile',
      'PpmxTrustSection',
      'PpmxFloatingWidgets',
    ]) {
      expect(homeSource).toContain(componentName)
    }

    for (const componentName of [
      'PpmxTopStrip',
      'PpmxHeader',
      'PpmxMobileDrawer',
      'PpmxNotificationPopover',
      'PpmxAuthModal',
    ]) {
      expect(chromeSource).toContain(componentName)
    }
  })

  it('renders the PP.MX reference footer under the trust section', () => {
    expect(homeSource).toContain('ppmx-footer')
    expect(homeSource).toContain('ppmxFooterColumns')
    expect(homeSource).toContain('ppmxSocialItems')
    expect(homeSource).toContain('ppmx.home.footerDescription')
    expect(homeSource).toContain('ppmx.home.footerLegal')
    expect(homeSource).toContain('ppmx.home.footerRights')
    expect(homeSource).toContain('class="ppmx-footer-legal"')
    expect(homeSource).toContain('class="ppmx-footer-bottom"')
    expect(homeSource).toContain('class="ppmx-footer-badges"')
    expect(homeSource).toContain('<h4>{{ pickText(column.title) }}</h4>')
    expect(homeSource).toContain('<ul>')
    expect(homeSource).toContain('<li v-for="link in column.links"')
    expect(homeSource).toContain('Responsible Gaming')
    expect(homeSource).toContain('PP.MX SUPPORT')
  })

  it('opens PP.MX auth modals instead of routing home login actions away', () => {
    expect(chromeSource).toContain("openAuth('login')")
    expect(homeSource).toContain("openAuth('register')")
    expect(chromeSource).toContain('authModalOpen')
    expect(chromeSource).toContain('authModalMode')
    expect(homeSource).toContain('handleFloatingWidgetTap')
  })

  it('opens registration from the registration gift hero slide for guests', () => {
    expect(homeSource).toContain("if (bannerId === 'registro') {")
    expect(homeSource).toContain("openAuth('register')")
    expect(homeSource).toContain("void router.push('/promo/registro')")
  })

  it('uses a single language dropdown instead of separate locale buttons', () => {
    expect(chromeSource).toContain('ppmxLanguageOptions')
    expect(chromeSource).toContain('triggerLabel')
    expect(chromeSource).toContain('languageMenuOpen')
    expect(chromeSource).toContain('ppmx-language-trigger')
    expect(chromeSource).toContain('ppmx-language-menu')
    expect(chromeSource).toContain('ppmx-language-option')
    expect(chromeSource).not.toContain('class="ppmx-lang"')
    expect(chromeSource).toContain('id="drawerLangPanel"')
    expect(chromeSource).toContain('drawerLanguageOpen')
    expect(chromeSource).toContain('class="ppmx-drawer-lang"')
  })

  it('keeps language dropdown styles available on mobile top strip', () => {
    const css = readFileSync(new URL('../../assets/styles/pp-mx-home.css', import.meta.url), 'utf8') as string

    expect(css).toContain('.ppmx-language-switcher')
    expect(css).toContain('.ppmx-language-trigger')
    expect(css).toContain('.ppmx-language-menu')
    expect(css).toContain('.ppmx-language-option.is-active')
    expect(css).toContain('.ppmx-top-strip__actions')
    expect(css).toContain('margin-left: auto')
    expect(css).toContain('.ppmx-top-strip__mobile {\n    display: none;')
  })

  it('keeps the language dropdown compact enough for mobile account pages', () => {
    const css = readFileSync(new URL('../../assets/styles/pp-mx-home.css', import.meta.url), 'utf8') as string

    expect(css).toContain('width: min(218px, calc(100vw - 24px))')
    expect(css).toContain('padding: 10px')
    expect(css).toContain('min-height: 48px')
    expect(css).toContain('font-size: 15px')
    expect(css).not.toContain('width: min(264px, calc(100vw - 32px))')
  })

  it('switches the mobile header menu button to back navigation outside tabbar pages', () => {
    expect(chromeSource).toContain('isPublicTabbarRoute')
    expect(chromeSource).toContain('const isTabbarRoute')
    expect(chromeSource).toContain('headerMenuIcon')
    expect(chromeSource).toContain("'fa-arrow-left'")
    expect(chromeSource).toContain('handleHeaderMenuClick')
    expect(chromeSource).toContain('goBack')
    expect(chromeSource).toContain('router.back()')
    expect(chromeSource).toContain("router.push('/profile')")
    expect(chromeSource).toContain(':class="headerMenuIcon"')
    expect(chromeSource).toContain(':aria-label="headerMenuAriaLabel"')
  })

  it('keeps the reference nav menu and drawer DOM ids', () => {
    expect(chromeSource).toContain('id="navMenu"')
    expect(chromeSource).toContain('id="drawer"')
  })

  it('loads dynamic game categories while keeping the PP.MX tile structure', () => {
    expect(homeSource).toContain("import { usePpmxGameHome } from './composables/usePpmxGameHome'")
    expect(homeSource).toContain('const gameHome = usePpmxGameHome()')
    expect(homeSource).toContain('void gameHome.loadGameHome()')
    expect(homeSource).toContain('v-for="category in homeCategories"')
    expect(homeSource).toContain('getGameTileStyle(game, gameIndex)')
    expect(homeSource).toContain('openGame(game)')
    expect(homeSource).toContain('v-else-if="gameHome.empty.value"')
    expect(homeSource).toContain('status="empty"')
  })

  it('hides the floating widgets during home page scrolling and restores them after scrolling stops', () => {
    expect(homeSource).toContain(':hiding="floating.isScrollHiding.value"')
    expect(homeSource).toContain("window.addEventListener('scroll', floating.handlePageScroll")
    expect(homeSource).toContain("window.removeEventListener('scroll', floating.handlePageScroll")
    expect(homeSource).toContain('floating.stopScrollWatcher()')
  })
})
