# PP.PE 首页精确复刻分步文档

目标：让当前 Vue3/Vant H5 项目的首页 `/` 与 `pp-mx-apple-v2 (2).html` 的首页做到视觉、结构、响应式和交互尽可能一致。这里的“一模一样”不是只套主题色，而是要复刻参考页的全局顶部栏、PP.PE Logo、通知条、Hero 轮播、游戏分类导航、分类卡片、信任区、浮动入口、移动抽屉和响应式断点。

## 0. 当前差距

当前项目已经完成了基础主题能力：

- Montserrat 字体已本地化到 `public/fonts/`。
- Font Awesome CSS 和 webfonts 已本地化到 `public/vendor/fontawesome/`。
- `src/assets/styles/theme.css` 已抽离黑曜石、红金、玻璃质感主题。
- `src/components/AppShell.vue` 已支持手机/PC 自动切换。
- `src/pages/index.vue` 当前仍是红包大厅业务首页，结构和参考页差距很大。

要做到完全一致，不能只改现有 `index.vue` 的几张卡片。参考页首页包含自己的 Header、Drawer、通知弹层、Hero Carousel、Sticky Game Nav、Lobby 分类渲染和浮动按钮。当前 `AppShell` 的 PC 顶部栏、上下文栏和移动端 `AppTabbar` 会干扰复刻结果，因此首页需要进入专用 PP.PE Shell 或直接绕开默认 Shell。

## 1. 锁定参考版本

先固定参考文件，避免后续对比时参考源变化。

```powershell
Get-FileHash "pp-mx-apple-v2 (2).html" -Algorithm SHA256
New-Item -ItemType Directory -Force "docs/reference/ppmx-home"
Copy-Item "pp-mx-apple-v2 (2).html" "docs/reference/ppmx-home/pp-mx-apple-v2.reference.html"
```

建议记录 6 个基准视口截图：

- `375x812`
- `390x844`
- `430x932`
- `768x1024`
- `1024x768`
- `1440x900`

验收时必须拿 Vue 首页截图和这些截图逐屏对比。

## 2. 明确首页复刻范围

本次只复刻参考 HTML 中 `page-home` 可见首页和它依赖的全局结构：

- 顶部细栏：客服、合规、责任博彩、语言切换。
- 主 Header：菜单、PP.PE Logo、桌面导航、登录/注册、余额态、通知入口。
- 通知弹层：铃铛点击后的通知列表。
- 移动抽屉：菜单、主题切换、语言切换。
- 首页通知条：`notifBar` 自动轮播、可点击、可滑动。
- Hero 轮播：`heroCarousel`、左右箭头、dots、自动播放、触摸滑动、金币/缩略图装饰。
- 游戏分类导航：`gameNavRow` sticky 横向滚动。
- Lobby 分类区：`HOME_CATS` 渲染出的分类标题、更多按钮、游戏卡片。
- Trust 信任区：4 个能力说明。
- 首页浮动入口：Check-in、Social、App，可拖动，可关闭。

暂不复刻参考页里其他独立页面，例如 sports、casino、promo、account、download、wheel、wallet、bank 等。首页按钮可以先保留视觉和点击反馈，业务跳转后续再接入真实路由。

## 3. 调整 Shell 策略

推荐新增首页专用 Shell，而不是把 `AppShell.vue` 改成 PP.PE 首页样式。

新增路由元信息：

```ts
definePage({
  name: 'home',
  meta: {
    title: 'PP.PE',
    keepAlive: true,
    tabbar: false,
    shell: 'ppmx',
  },
})
```

需要扩展路由类型，例如：

```ts
declare module 'vue-router' {
  interface RouteMeta {
    shell?: 'default' | 'ppmx' | 'none'
  }
}
```

在 `App.vue` 或当前 Shell 入口处按 `route.meta.shell` 分流：

- `default`：继续使用当前 `AppShell`，服务红包业务页。
- `ppmx`：使用 `PpmxHomeShell`，完全承载参考页首页结构。
- `none`：给维护页、强更页或全屏异常页使用。

这是精确复刻的关键步骤。否则当前 PC 顶部栏、上下文栏、移动底部 Tabbar、页面 padding 都会让首页无法和参考页对齐。

## 4. 建议目录结构

新增独立 feature，避免把参考页的大量结构散落在全局组件中。

```text
src/components/ppmx-home/
  PpmxHome.vue
  data.ts
  types.ts
  composables/
    useHeroCarousel.ts
    useNoticeCarousel.ts
    useFloatingWidgets.ts
  components/
    PpmxTopStrip.vue
    PpmxHeader.vue
    PpmxLogo.vue
    PpmxMobileDrawer.vue
    PpmxNotificationPopover.vue
    PpmxNoticeBar.vue
    PpmxHeroCarousel.vue
    PpmxGameNav.vue
    PpmxLobbySection.vue
    PpmxGameTile.vue
    PpmxTrustSection.vue
    PpmxFloatingWidgets.vue
src/assets/styles/
  pp-mx-home.css
```

`src/pages/index.vue` 只做路由声明和挂载：

```vue
<script setup lang="ts">
import PpmxHome from '@/components/ppmx-home/PpmxHome.vue'

definePage({
  name: 'home',
  meta: {
    title: 'PP.PE',
    keepAlive: true,
    tabbar: false,
    shell: 'ppmx',
  },
})
</script>

<template>
  <PpmxHome />
</template>
```

## 5. 抽取参考数据

不要在 Vue 模板里写长数组。把参考 HTML 的数据抽到 `data.ts`：

- `NOTICES`
- `BANNERS`
- `HOME_CATS`
- `CAT_BRANDS`
- 首页 Header nav items
- drawer nav items
- trust items
- floating widget items

类型建议：

```ts
export type LocaleCode = 'es' | 'en' | 'zh'

export interface LocalizedText {
  es: string
  en: string
  zh: string
}

export interface HomeBanner {
  id: string
  tag: LocalizedText
  title: LocalizedText
  sub: LocalizedText
  cta: LocalizedText
  benefit?: LocalizedText
  amountHtml?: string
  background: string
  thumbs?: number[]
}

export interface HomeCategory {
  id: string
  icon: string
  name: LocalizedText
  games: HomeGame[]
  go?: 'sports' | 'casino' | 'wheel'
  wheel?: boolean
}
```

注意：参考页里部分文案是 HTML 字符串，例如 `<span class="pill">Sin depósito</span>`。如果继续用 `v-html`，必须保证来源是本地静态常量或后端已净化内容；不要直接渲染未处理的接口文本。

## 6. 迁移样式

当前 `theme.css` 只负责全局主题。参考首页专属样式应放入 `src/assets/styles/pp-mx-home.css`，并使用命名空间降低污染：

- `.ppmx-page`
- `.ppmx-top-strip`
- `.ppmx-header`
- `.ppmx-logo`
- `.ppmx-drawer`
- `.ppmx-notif-bar`
- `.ppmx-hero-carousel`
- `.ppmx-game-nav`
- `.ppmx-hcat`
- `.ppmx-htile`
- `.ppmx-float-widget`

迁移原则：

- 参考页里的 `.glass`、`.glass-strong`、`.btn-red`、`.btn-gold` 可迁移，但建议改成 `.ppmx-glass`、`.ppmx-glass-strong`、`.ppmx-btn-red`、`.ppmx-btn-gold`。
- 保留参考色值：`#060608`、`#C8102E`、`#FFD700`、`#f5f5f7`。
- 保留参考断点：`max-width: 1023px`、`max-width: 640px`、`max-width: 380px`、`min-width: 768px`、`min-width: 1100px`。
- 保留参考布局宽度：`max-width: 1320px`。
- 不要复用当前 `.page-surface`、`.responsive-grid`、`.pc-topbar`，这些是项目默认 Shell 的布局，会造成差异。

在 `PpmxHome.vue` 中引入专属 CSS：

```ts
import '@/assets/styles/pp-mx-home.css'
```

## 7. 本地化视觉资源

已完成：

- `public/fonts/montserrat/montserrat.css`
- `public/vendor/fontawesome/css/all.min.css`
- `public/vendor/fontawesome/webfonts/*`

还需要处理：

- 参考页里的 Unsplash 图片 URL。
- 参考页里的 SVG symbol：`ppCoin`、`ppEmblem`。
- 游戏卡片缩略图、Hero 装饰图和 fallback 渐变。

要做到“一模一样”，必须把参考页实际使用的远程图片下载到本地，例如：

```text
public/images/ppmx/
  banners/
  games/
  symbols/
```

然后把 `data.ts` 中图片地址改成本地路径：

```ts
const asset = (path: string) => `${import.meta.env.BASE_URL}images/ppmx/${path}`
```

如果图片授权无法确认，不要直接用于商业环境；可以先用于内部复刻验收，生产版再换成有授权的同尺寸资源。

## 8. 组件实现顺序

按下面顺序实现，避免一开始就处理所有交互导致难以对齐。

1. `PpmxLogo.vue`
   - 迁移参考页 `ppEmblem` SVG symbol。
   - 保持 PP.PE 字标尺寸：桌面约 `150x40`，移动约 `120x32`。

2. `PpmxTopStrip.vue`
   - 复刻顶部 40px 细栏。
   - 包含客服、合规、责任博彩、语言切换。
   - 移动端隐藏左侧辅助信息。

3. `PpmxHeader.vue`
   - 复刻 68px sticky Header。
   - 桌面显示 nav，移动显示汉堡菜单。
   - 登录/注册、余额态、通知按钮按参考结构保留。

4. `PpmxMobileDrawer.vue`
   - 用 Vue `ref` 控制打开关闭，替代 `toggleDrawer()`。
   - 保持 `280px` 宽度、玻璃背景、遮罩、语言折叠面板。

5. `PpmxNotificationPopover.vue`
   - 点击铃铛打开。
   - 点击外部关闭。
   - 通知未读点按本地 mock 状态展示。

6. `PpmxNoticeBar.vue`
   - 使用 `NOTICES` 数据。
   - 实现 4.5 秒自动轮播、hover 暂停、touch swipe。

7. `PpmxHeroCarousel.vue`
   - 使用 `BANNERS` 数据。
   - 实现 6 秒自动轮播、左右箭头、dots、hover 暂停、touch swipe。
   - 迁移金币 SVG、装饰缩略图、`GO!` 按钮、进入动画。

8. `PpmxGameNav.vue`
   - 使用 `HOME_CATS` 数据。
   - sticky 位置要按参考页：顶部 Header 下方。
   - 横向滚动隐藏滚动条。
   - 点击分类时滚动到对应 `PpmxLobbySection`。

9. `PpmxLobbySection.vue` 和 `PpmxGameTile.vue`
   - 复刻分类标题、更多按钮、卡片比例、HOT/NEW badge、hover overlay。
   - 使用 Font Awesome class 渲染图标。

10. `PpmxTrustSection.vue`
    - 复刻底部 4 栏 trust 区，移动端 2 列。

11. `PpmxFloatingWidgets.vue`
    - 复刻 3 个浮动入口。
    - 支持关闭、拖拽、resize 后回到安全位置。
    - 移动端注意不遮挡 PWA 安装按钮和系统安全区。

## 9. 把原生 JS 改成 Vue 状态

不要直接复制参考页 `<script>` 中的全局函数。需要映射为 Vue composable：

| 参考函数 | Vue 实现 |
| --- | --- |
| `toggleDrawer()` | `const drawerOpen = ref(false)` |
| `toggleLangMenu()` | `const langMenuOpen = ref(false)` |
| `renderBanners()` | `computed` + `v-for` |
| `goSlide()` / `nextSlide()` / `prevSlide()` | `useHeroCarousel()` |
| `startBannerAuto()` / `stopBannerAuto()` | `useIntervalFn` 或 `window.setInterval` + `onBeforeUnmount` |
| `renderNotif()` | `PpmxNoticeBar.vue` 的 `v-for` |
| `lobbyGo()` | `scrollIntoView` 或 `window.scrollTo` |
| `updateNavSpy()` | `IntersectionObserver` 或 scroll listener |
| `initFloat()` | `useFloatingWidgets()` |

所有定时器、事件监听、拖拽监听必须在 `onBeforeUnmount()` 清理，避免 KeepAlive 或路由切换后重复注册。

## 10. i18n 接入策略

参考页内置 `es`、`en`、`zh`。当前项目已有 `vue-i18n`。

推荐做法：

- 首页数据先保留 `LocalizedText`，通过当前语言选择字段。
- 当前语言不是 `es/en/zh` 时回退 `es` 或 `zh`，按产品目标决定。
- 不要把参考页全部文案硬塞进全局 i18n 文件，首页 mock 数据保留在 `data.ts` 更容易维护。

示例：

```ts
function pickText(text: LocalizedText, locale: string) {
  if (locale.startsWith('zh')) return text.zh
  if (locale.startsWith('en')) return text.en
  return text.es
}
```

## 11. 响应式验收标准

必须按参考断点验收：

- `<640px`：Hero 高度约 `240px-280px`，箭头隐藏，按钮紧凑，导航横向滚动。
- `<1024px`：显示移动菜单抽屉，不显示桌面 nav。
- `>=1024px`：显示桌面 nav，Header 居中最大宽 `1320px`，无移动底部 Tabbar。
- `>=1440px`：内容仍限制在 `1320px`，背景延展。

重点检查：

- 页面不能出现横向滚动。
- Header sticky 后不能遮挡分类导航。
- Hero dots 和 `GO!` 按钮位置与参考页一致。
- 浮动入口不能压住主要按钮。
- 字体必须是 Montserrat。
- Font Awesome 图标必须来自本地 CSS，不再依赖 CDN。

## 12. 像素对比流程

先打开参考 HTML 截图，再打开 Vue 首页截图。

建议本地启动：

```powershell
pnpm dev -- --host 0.0.0.0
```

逐个视口对比：

```text
390x844
430x932
768x1024
1024x768
1440x900
```

对比顺序：

1. 背景颜色和径向光是否一致。
2. 顶部细栏高度、文字、图标、间距是否一致。
3. Header 高度、Logo 尺寸、按钮位置是否一致。
4. 通知条高度、圆角、图标、文字滚动是否一致。
5. Hero 高度、圆角、内部文案、金币装饰、dots、按钮是否一致。
6. 分类导航 sticky 位置、按钮宽度、激活态是否一致。
7. 游戏卡片比例、间距、badge、hover 态是否一致。
8. Trust 区间距和移动端列数是否一致。
9. 浮动入口大小、位置、拖拽和关闭行为是否一致。

如果差异超过肉眼明显程度，优先调 CSS 尺寸和间距，不要先改业务逻辑。

## 13. 构建与回归验证

每一阶段完成后至少运行：

```powershell
pnpm typecheck
pnpm build
pnpm check:dist
```

如果新增了 composable 测试，补充：

```powershell
pnpm test:unit
```

还需要手动验证：

- 首页 `/` 正常打开。
- `/mine` 等现有页面仍走默认 `AppShell`。
- 维护中/强制更新页仍能覆盖首页。
- PWA 安装按钮不遮挡首页浮动入口。
- 断网或接口失败不会影响首页静态渲染。

## 14. 推荐实施里程碑

### Milestone 1：静态结构对齐

- 新增 `PpmxHome.vue` 和 `pp-mx-home.css`。
- 完成顶部细栏、Header、Drawer、NoticeBar、Hero、GameNav、Lobby、Trust、FloatingWidgets 的静态模板。
- 暂时使用 mock 数据。
- 验收：静态截图和参考页主体布局接近。

### Milestone 2：交互对齐

- 接入 Hero 自动播放、dots、箭头、swipe。
- 接入通知条自动轮播和 swipe。
- 接入 Drawer、语言菜单、通知弹层。
- 接入分类导航 scroll-spy。
- 接入浮动入口拖拽和关闭。
- 验收：主要交互行为与参考页一致。

### Milestone 3：资源对齐

- 下载并本地化参考页使用的图片资源。
- 转换 SVG symbol 到 Vue 组件。
- 检查字体、图标、图片都不再依赖外部 CDN。
- 验收：构建产物中无 Google Fonts、cdnjs、Unsplash 运行时依赖。

### Milestone 4：像素级修正

- 逐视口截图对比。
- 修正高度、间距、圆角、阴影、断点、动画时长。
- 验收：`390x844` 和 `1440x900` 两个核心视口达到肉眼一致。

### Milestone 5：接入真实业务

- 登录/注册按钮接现有认证页面或弹层。
- 余额态接 Pinia 用户状态。
- 游戏卡片和活动入口接真实接口。
- 点击事件接自研埋点。
- 保持 UI 结构不变，避免业务接入后破坏像素结果。

## 15. 风险与取舍

- 参考 HTML 是单文件原型，包含大量全局函数和 DOM 操作，直接复制进 Vue 会带来重复 ID、状态失控和卸载不清理问题。
- 当前默认 `AppShell` 与参考首页不兼容，必须新增专用 shell 或 route-level bypass。
- 图片资源如果来自第三方，需要确认授权；商业化 To C 项目不能长期依赖无授权图。
- 参考页使用 Tailwind CDN 原型写法，项目内是 Tailwind v4 + Vite，最终应沉淀为本地 CSS 和 Vue 组件。
- “一模一样”会牺牲部分现有红包首页信息架构。如果仍要保留红包业务入口，应把它融入 PP.PE 首页的分类/活动卡片，而不是继续保留当前红包大厅布局。

## 16. 最终验收口径

满足以下条件后，才能认为首页已按参考页完成：

- `/` 首屏在移动端和 PC 端都与 `pp-mx-apple-v2 (2).html` 首页布局一致。
- `AppShell` 默认顶部栏和移动底部 Tabbar 不再出现在 PP.PE 首页。
- Header、Drawer、通知条、Hero、分类导航、Lobby、Trust、浮动入口都已组件化。
- 自动轮播、触摸滑动、抽屉、通知弹层、分类滚动、浮动入口关闭/拖拽均可用。
- 本地字体和 Font Awesome 正常加载，无外部字体/CDN 依赖。
- `pnpm typecheck`、`pnpm build`、`pnpm check:dist` 全部通过。
- 关键视口无横向滚动、无文本溢出、无元素遮挡。
