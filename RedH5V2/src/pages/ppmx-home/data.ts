import type {
  LocalizedText,
  PpmxBanner,
  PpmxCategoryBrand,
  PpmxFooterColumn,
  PpmxFloatingWidget,
  PpmxHomeCategory,
  PpmxNavItem,
  PpmxNotice,
  PpmxSocialItem,
  PpmxTrustItem,
} from './types'

const text = (value: string) => ({
  en: value,
  es: value,
  zh: value,
})

const PPMX_HERO_BANNER_IMAGE_ROOT = '/images/ppmx/promos'
const PPMX_HERO_BANNER_OVERLAY = 'linear-gradient(95deg, rgba(6, 6, 8, 0.94) 0%, rgba(8, 7, 10, 0.8) 28%, rgba(26, 16, 4, 0.34) 56%, rgba(46, 30, 8, 0.05) 100%)'

const heroBannerImage = (id: string) => `${PPMX_HERO_BANNER_IMAGE_ROOT}/${id}.jpg`

const heroBannerBackground = (id: string, fallback: string) =>
  `${PPMX_HERO_BANNER_OVERLAY}, url("${heroBannerImage(id)}") right center / cover, ${fallback}`

const game = (
  name: string,
  imageId: number,
  flags: Pick<import('./types').PpmxGame, 'hot' | 'isNew'> = {},
) => ({
  id: name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, ''),
  imageId,
  name: text(name),
  ...flags,
})

export const ppmxNavItems: PpmxNavItem[] = [
  { id: 'home', icon: 'fa-house', label: { es: 'Inicio', en: 'Home', zh: '首页' }, target: 'home' },
  { id: 'wheel', icon: 'fa-dharmachakra', label: { es: 'Ruleta', en: 'Wheel', zh: '幸运转盘' }, target: 'wheel' },
  { id: 'team', icon: 'fa-user-group', label: { es: 'Mi equipo', en: 'My team', zh: '我的团队' }, target: 'team' },
  { id: 'promo', icon: 'fa-gift', label: { es: 'Promociones', en: 'Promotions', zh: '优惠' }, target: 'promo' },
  { id: 'account', icon: 'fa-user', label: { es: 'Mi Cuenta', en: 'My Account', zh: '我的账户' }, target: 'account' },
]

export const ppmxDrawerItems: PpmxNavItem[] = [
  { id: 'home', icon: 'fa-house', label: { es: 'Inicio', en: 'Home', zh: '首页' }, target: 'home' },
  { id: 'promo', icon: 'fa-gift', label: { es: 'Promociones', en: 'Promotions', zh: '优惠活动' }, target: 'promo' },
  { id: 'account', icon: 'fa-user', label: { es: 'Mi Cuenta', en: 'My Account', zh: '我的账户' }, target: 'account' },
  { id: 'help', icon: 'fa-circle-question', label: { es: 'Ayuda', en: 'Help', zh: '帮助' }, target: 'help' },
  { id: 'responsible', icon: 'fa-shield-heart', label: { es: 'Juego Responsable', en: 'Responsible Gaming', zh: '责任博彩' }, target: 'responsible' },
]

export const ppmxNotices: PpmxNotice[] = [
  {
    id: 'welcome-gift',
    icon: 'fa-gift',
    target: 'registro',
    text: {
      es: 'Regístrate y recibe $38 USD gratis',
      en: 'Sign up and get $38 USD free',
      zh: '注册立即赠送 38 美元',
    },
  },
  {
    id: 'bank-withdrawals',
    icon: 'fa-bolt',
    text: {
      es: 'Retiros bancarios acreditados con rapidez',
      en: 'Bank withdrawals credited quickly',
      zh: '银行提现快速到账',
    },
  },
  {
    id: 'reload-bonus',
    icon: 'fa-coins',
    target: 'recarga',
    text: {
      es: 'Recarga y recibe hasta 30% de reembolso',
      en: 'Deposit and get up to 30% rebate',
      zh: '用户充值即享 最高 30%返利',
    },
  },
  {
    id: 'referral-commission',
    icon: 'fa-user-group',
    target: 'invita',
    text: {
      es: 'Invita amigos y gana hasta 50% de comisión por cada recarga',
      en: 'Invite friends and earn up to 50% commission on every deposit',
      zh: '邀请好友 每笔充值最高返佣 50%',
    },
  },
  {
    id: 'regulated-gaming',
    icon: 'fa-shield-halved',
    text: {
      es: 'Juego 100% seguro y regulado · +21',
      en: '100% safe & regulated gaming · 21+',
      zh: '100% 安全合规博彩 · 21 岁以上',
    },
  },
]

export const ppmxBanners: PpmxBanner[] = [
  {
    id: 'registro',
    image: heroBannerImage('registro'),
    backgroundPosition: 'right center',
    background: heroBannerBackground('registro', 'radial-gradient(120% 130% at 12% 18%, #138a52 0%, #0c6b40 22%, #7a0f24 58%, #4a0413 80%, #2a020c 100%)'),
    thumbs: [11, 13, 14],
    amountHtml: text('<span class="cur">$</span>38<span class="suf">USD</span>'),
    tag: { es: 'Regalo de bienvenida', en: 'Welcome gift', zh: '注册礼金' },
    benefitHtml: {
      es: 'Gratis solo por registrarte <span class="pill">Sin depósito</span>',
      en: 'Free just for signing up <span class="pill">No deposit</span>',
      zh: '注册即送 · 无需充值 <span class="pill">免存款</span>',
    },
    title: { es: 'Regístrate y gana $38 USD gratis', en: 'Sign up and get $38 USD free', zh: '注册立即赠送 38 美元' },
    sub: {
      es: 'Crea tu cuenta en 1 minuto y recibe $38 USD al instante para empezar a jugar.',
      en: 'Create your account in 1 minute and get $38 USD instantly to start playing.',
      zh: '1 分钟完成注册，立即获得 38 美元，马上开玩。',
    },
    cta: { es: 'Registrarme y cobrar', en: 'Sign up & claim', zh: '注册领取' },
    valid: { es: 'Promoción para nuevos usuarios', en: 'New users only', zh: '仅限新用户' },
    howTo: {
      es: 'Regístrate con tu teléfono, verifica tu cuenta y los $38 USD se acreditan automáticamente.',
      en: 'Register with your phone, verify your account and the $38 USD is credited automatically.',
      zh: '点击下方立即激活优惠后输入兑换码领取 38 美元新用户注册奖励。',
    },
    terms: { es: '1 por persona/dispositivo. Rollover 10x. Aplican T&C. +21.', en: '1 per person/device. 10x rollover. T&C apply. 21+.', zh: '每人/每设备限领一次，10 倍流水，适用条款，21 岁以上。' },
  },
  {
    id: 'checkin',
    image: heroBannerImage('checkin'),
    backgroundPosition: 'right center',
    background: heroBannerBackground('checkin', 'radial-gradient(120% 130% at 14% 18%, #c8922b 0%, #8a5e16 34%, #6b4a12 62%, #1c1206 100%)'),
    thumbs: [21, 26, 28],
    amountHtml: text('<span class="cur">$</span>???'),
    tag: { es: 'Check-in de bienvenida', en: 'Welcome check-in', zh: '新人签到' },
    benefitHtml: {
      es: 'Entra cada día y reclama tu regalo <span class="pill">Cofre día 7</span>',
      en: 'Log in daily and claim your gift <span class="pill">Day-7 chest</span>',
      zh: '连续签到 7 天天天领奖 <span class="pill">第 7 天宝箱</span>',
    },
    title: { es: 'Check-in de nuevos: gana hasta $888 en 7 días', en: 'New-user check-in: win up to $888 in 7 days', zh: '新用户签到 7 天最高领 $???' },
    sub: {
      es: 'Regístrate y entra cada día para reclamar tu recompensa; el día 7 abre el cofre mayor.',
      en: 'Sign up and log in daily to claim your reward; day 7 unlocks the grand chest.',
      zh: '注册后每日登录即可签到领奖，连续 7 天第 7 天开启终极宝箱。',
    },
    cta: { es: 'Hacer check-in', en: 'Check in now', zh: '立即签到' },
    valid: { es: 'Para nuevos usuarios · 7 días', en: 'New users · 7 days', zh: '仅限新用户 · 7 天内' },
    howTo: {
      es: 'Tras registrarte, entra cada día y pulsa Check-in para acreditar la recompensa diaria. 7 días consecutivos liberan el cofre del día 7.',
      en: 'After signing up, log in each day and tap Check-in to credit your daily reward. 7 consecutive days unlock the day-7 chest.',
      zh: '注册后每天登录点击下方立即激活优惠或首页「新人签到」即可领取当日奖励，连续签到 7 天解锁第 7 天宝箱大奖。',
    },
    terms: { es: 'Solo nuevos usuarios. 1 check-in por día. Rollover 10x. +21.', en: 'New users only. 1 check-in per day. 10x rollover. 21+.', zh: '仅限新用户，每日限签到一次，10 倍流水，21 岁以上。' },
  },
  {
    id: 'vip',
    image: heroBannerImage('vip'),
    backgroundPosition: 'right center',
    background: heroBannerBackground('vip', 'radial-gradient(120% 130% at 16% 16%, #6d28d9 0%, #4c1d95 30%, #2a1163 62%, #160a33 100%)'),
    thumbs: [12, 15, 16],
    amountHtml: text('<span class="cur">$</span>188,888'),
    tag: { es: 'Programa VIP', en: 'VIP program', zh: 'VIP 俱乐部' },
    benefitHtml: {
      es: 'Premio por cada nivel que subes <span class="pill">+ Salario</span>',
      en: 'Reward for every level you climb <span class="pill">+ Salary</span>',
      zh: '每升一级即享奖励 <span class="pill">+ 周薪</span>',
    },
    title: { es: 'Sube de nivel VIP y reclama hasta $188,888', en: 'Level up VIP and claim up to $188,888', zh: 'VIP 升级奖励 最高 $188,888' },
    sub: {
      es: 'Más rango = más recompensas: bonos, salario ampliado y un gestor personal.',
      en: 'Higher rank = more rewards: bonuses, bigger salary and a personal host.',
      zh: '等级越高奖励越多：升级红利、更高周薪与专属客户经理。',
    },
    cta: { es: 'Ver beneficios VIP', en: 'See VIP perks', zh: '查看 VIP 权益' },
    valid: { es: 'Recompensas permanentes', en: 'Permanent rewards', zh: '长期权益' },
    howTo: {
      es: 'Acumula puntos jugando. Al alcanzar cada nivel, reclama tu bono de ascenso desde tu perfil VIP.',
      en: 'Earn points by playing. On reaching each level, claim your upgrade bonus from your VIP profile.',
      zh: '通过游戏累积积分，达成每个等级后即可在 VIP 中心领取升级红利。',
    },
    terms: { es: 'Bono según nivel alcanzado. Rollover 10x. +21.', en: 'Bonus by level reached. 10x rollover. 21+.', zh: '红利依达成等级而定，10 倍流水，21 岁以上。' },
  },
  {
    id: 'invita',
    image: heroBannerImage('invita'),
    backgroundPosition: 'right center',
    background: heroBannerBackground('invita', 'radial-gradient(120% 130% at 14% 18%, #0891b2 0%, #0e7490 26%, #b1480c 60%, #7a1f5c 88%, #3c0f30 100%)'),
    thumbs: [25, 20, 17],
    amountHtml: {
      es: '<span class="suf">Hasta</span>50<span class="cur">%</span><span class="suf">comisión</span>',
      en: '<span class="suf">Up to</span>50<span class="cur">%</span><span class="suf">rebate</span>',
      zh: '<span class="suf">最高</span>50<span class="cur">%</span><span class="suf">返佣</span>',
    },
    tag: { es: 'Programa de afiliados', en: 'Referral program', zh: '邀请返佣' },
    title: { es: 'Hasta 50% de comisión', en: 'Up to 50% rebate', zh: '最高50%返佣' },
    sub: {
      es: 'Comparte tu enlace exclusivo e invita a amigos a registrarse y recargar: gana hasta 50% de comisión sobre las recargas de cada usuario válido, ¡de por vida!',
      en: "Share your exclusive link and invite friends to register and deposit — earn up to 50% commission on each valid user's deposits, for life!",
      zh: '分享您的专属链接，邀请好友注册充值，每位有效用户最高可享充值金额50%佣金，终身有效！',
    },
    cta: { es: 'Invitar y ganar', en: 'Invite & earn', zh: '立即邀请' },
    valid: { es: 'Comisión recurrente sin límite', en: 'Recurring, uncapped commission', zh: '长期返佣 上不封顶' },
    howTo: {
      es: 'Copia tu enlace de referido desde Mi equipo, compártelo y cobra tus comisiones cada semana.',
      en: 'Copy your referral link from My team, share it and collect commissions every week.',
      zh: '点击下方立即激活优惠或在「我的团队」复制专属链接并分享，实时结算返佣。',
    },
    terms: { es: 'Comisión según el número de recarga del referido. +21.', en: 'Commission based on referral deposit count. 21+.', zh: '按下级充值次数计算返佣，21 岁以上。' },
  },
  {
    id: 'recarga',
    image: heroBannerImage('recarga'),
    backgroundPosition: 'right center',
    background: heroBannerBackground('recarga', 'radial-gradient(120% 130% at 14% 16%, #f0a93d 0%, #d11f33 30%, #9a1322 62%, #4a0413 88%, #2a020c 100%)'),
    thumbs: [30, 27, 29],
    amountHtml: {
      es: '<span class="suf" style="font-size:.55em">Hasta</span>30<span class="suf">% regalo</span>',
      en: '<span class="suf" style="font-size:.55em">Up to</span>30<span class="suf">% bonus</span>',
      zh: '<span class="suf" style="font-size:.55em">最高</span>30<span class="suf">%赠送</span>',
    },
    tag: { es: 'Bono de recarga', en: 'Reload bonus', zh: '充值赠送' },
    benefitHtml: {
      es: 'Los usuarios que recarguen reciben <span class="pill">hasta 30% de reembolso</span>',
      en: 'Users who deposit get <span class="pill">up to 30% rebate</span>',
      zh: '用户充值即享 <span class="pill">最高 30%返利</span>',
    },
    title: { es: 'Recarga y recibe hasta 30% de regalo', en: 'Deposit and get up to 30% bonus', zh: '最高30%赠送' },
    sub: {
      es: 'La primera recarga recibe 18% de reembolso, la segunda 20%, la tercera 25% y desde la cuarta 30%. Cuantas más recargas hagas, mayor será el reembolso.',
      en: 'First deposit gets 18% rebate, second deposit gets 20%, third deposit gets 25%, and the 4th deposit and above gets 30%. The more times you deposit, the higher your rebate.',
      zh: '首次充值享18%返利，二次充值享20%返利，三次充值享25%，≧充值享30%。充值的次数越多享受的返利越高！',
    },
    cta: { es: 'Recargar ahora', en: 'Deposit now', zh: '立即充值' },
    valid: { es: 'Válido todos los días', en: 'Valid every day', zh: '每日有效' },
    howTo: {
      es: 'Ve a Depositar, elige tu método, ingresa el monto y el reembolso según tu número de recarga se aplica automáticamente.',
      en: 'Go to Deposit, pick your method, enter the amount and the rebate for your deposit count applies automatically.',
      zh: '点击下方立即激活优惠或前往存款页选择方式并输入金额，红利自动到账。',
    },
    terms: { es: 'Bono según número de recarga. Rollover 10x. +21.', en: 'Bonus by deposit count. 10x rollover. 21+.', zh: '按充值次数赠送，10 倍流水，21 岁以上。' },
  },
  {
    id: 'download',
    image: heroBannerImage('download'),
    backgroundPosition: 'right center',
    background: heroBannerBackground('download', 'radial-gradient(120% 130% at 14% 18%, #3b82f6 0%, #1d4ed8 30%, #1e3a8a 62%, #0a0e26 100%)'),
    thumbs: [19, 22, 24],
    amountHtml: text('<span class="cur">$</span>8'),
    tag: { es: 'App móvil', en: 'Mobile app', zh: '下载 APP' },
    benefitHtml: {
      es: 'Descarga la app y recibe tu bono <span class="pill">+ Giros gratis</span>',
      en: 'Download the app and get your bonus <span class="pill">+ Free spins</span>',
      zh: '下载 APP 即送红利 <span class="pill">+ 免费旋转</span>',
    },
    title: { es: 'Descarga la app PP.BET y gana $8', en: 'Download the PP.BET app and get $30', zh: '下载 PP.BET APP 立即领 $8' },
    sub: {
      es: 'Juega donde quieras: descarga la app, inicia sesión y recibe $30 USD al instante.',
      en: 'Play anywhere: download the PP.BET app, log in and get $30 USD instantly.',
      zh: '随时随地畅玩，下载 APP 并登录即送 8 美元，还有专属推送福利。',
    },
    cta: { es: 'Descargar app', en: 'Download app', zh: '立即下载' },
    valid: { es: 'Solo nuevas instalaciones', en: 'New installs only', zh: '仅限新下载用户' },
    howTo: {
      es: 'Pulsa Descargar, instala la app PP.BET, inicia sesión y el bono de $30 se acredita solo.',
      en: 'Tap Download, install the PP.BET app, log in and the $30 bonus is credited automatically.',
      zh: '点击下载并安装 PP.BET APP，登录账户后 8 美元自动到账。',
    },
    terms: { es: '1 por cuenta/dispositivo. Rollover 10x. +21.', en: '1 per account/device. 10x rollover. 21+.', zh: '每账户/设备限领一次，10 倍流水，21 岁以上。' },
  },
  {
    id: 'wheel',
    image: heroBannerImage('wheel'),
    backgroundPosition: 'right center',
    background: heroBannerBackground('wheel', 'radial-gradient(120% 130% at 14% 18%, #f59e0b 0%, #d6259a 34%, #7e1f6d 64%, #1f0a1c 100%)'),
    thumbs: [21, 26, 28],
    amountHtml: text('<span class="cur">$</span>888,888'),
    tag: { es: 'Ruleta de la suerte', en: 'Lucky wheel', zh: '幸运转盘' },
    benefitHtml: {
      es: 'Gira cada día y gana premios <span class="pill">Giro gratis diario</span>',
      en: 'Spin daily and win prizes <span class="pill">Free daily spin</span>',
      zh: '参与游戏获得次数抽奖赢大奖 <span class="pill">赢大奖</span>',
    },
    title: { es: 'Gira la ruleta y gana hasta $10,000', en: 'Spin the wheel and win up to $10,000', zh: '幸运转盘每日转 最高赢 $888888' },
    sub: {
      es: 'Cada día tienes un giro gratis: bonos, giros y hasta $10,000 USD te esperan. Por cada recarga de ≥$200 recibes un giro gratis (con $400 son dos giros, y así sucesivamente).',
      en: 'Every day you get a free spin: bonuses, spins and up to $10,000 USD await. Each single deposit of ≥$200 grants one free spin ($400 gives two spins, and so on).',
      zh: '参与游戏获得免费旋转，获得奖励、免费旋转及最高 888888 美元等你拿。（每达到 10000 流水即可自动获得一次免费幸运转盘旋转次数）单次充值金额≧$200元赠送一次免费旋转。（若充值$400则赠送两次免费旋转依次类推）',
    },
    cta: { es: 'Girar ahora', en: 'Spin now', zh: '立即抽奖' },
    valid: { es: 'Un giro gratis al día', en: 'One free spin per day', zh: '2026 年' },
    howTo: {
      es: 'Entra a la Ruleta de la suerte, usa tu giro gratis diario y reclama el premio que caiga.',
      en: 'Open the Lucky Wheel, use your daily free spin and claim whatever prize lands.',
      zh: '点击下方激活优惠或进入「幸运转盘」页面，点击旋转，转到的奖励即可领取。',
    },
    terms: { es: '1 giro gratis/día. Rollover según premio. +21.', en: '1 free spin/day. Rollover varies by prize. 21+.', zh: '参与游戏免费获得旋转次数，10 倍流水，21 岁以上。' },
  },
]

export const ppmxHomeCategories: PpmxHomeCategory[] = [
  {
    id: 'populares',
    icon: 'fa-fire',
    target: 'casino',
    name: { es: 'Populares', en: 'Popular', zh: '热门的' },
    games: [
      game('Sweet Bonanza', 11, { hot: true }),
      game('Gates of Olympus', 12, { hot: true }),
      game('Crazy Time', 25, { hot: true }),
      game('Mega Moolah', 29, { hot: true }),
      game('Sugar Rush', 13, { isNew: true }),
      game('Lightning Roulette', 26, { hot: true }),
      game('Razor Shark', 19),
      game('Big Bass Bonanza', 14),
    ],
  },
  {
    id: 'slots',
    icon: 'fa-dice',
    target: 'casino',
    name: { es: 'Tragamonedas', en: 'Slots', zh: '老虎机' },
    games: [
      game('Sweet Bonanza', 11),
      game('Gates of Olympus', 12),
      game('Sugar Rush', 13, { isNew: true }),
      game('Big Bass Bonanza', 14),
      game('Starburst', 15),
      game('Book of Dead', 16),
      game('Wolf Gold', 17),
      game('Dog House', 20, { isNew: true }),
    ],
  },
  {
    id: 'pesca',
    icon: 'fa-fish-fins',
    target: 'casino',
    name: { es: 'Pesca', en: 'Fishing', zh: '捕鱼' },
    games: [
      game('Fishing God', 61, { hot: true }),
      game('Dragon Fishing', 62),
      game('Royal Fishing', 63),
      game('Mega Fishing', 64, { isNew: true }),
      game('Ocean King', 65),
      game('Fish Hunter', 66),
      game('Bombing Fishing', 67),
      game('Jackpot Fishing', 68, { hot: true }),
    ],
  },
  {
    id: 'deportes',
    icon: 'fa-futbol',
    target: 'sports',
    name: { es: 'Deportes', en: 'Sports', zh: '体育' },
    games: [
      game('World Cup', 71, { hot: true }),
      game('Belt of Champion', 72),
      game('Chained Fighters', 73, { isNew: true }),
      game('Wild Gladiators', 74, { hot: true }),
      game("Wrigley's World", 75),
      game('Wild Horses', 76),
      game("Bull's Club", 77),
      game('Battle Roosters', 78, { isNew: true }),
    ],
  },
  {
    id: 'envivo',
    icon: 'fa-tower-broadcast',
    target: 'casino',
    name: { es: 'En vivo', en: 'Live', zh: '真人' },
    games: [
      game('Crazy Time', 25, { hot: true }),
      game('Lightning Roulette', 26, { hot: true }),
      game('Monopoly Live', 27),
      game('Blackjack VIP', 21),
      game('Baccarat', 23),
      game('Mega Ball', 81),
      game('Dream Catcher', 82),
      game('CandyLand', 83, { isNew: true }),
    ],
  },
]

export const ppmxCategoryBrands: Record<string, PpmxCategoryBrand[]> = {
  slots: [
    { id: 'pragmatic', name: text('Pragmatic Play'), games: [game('Sweet Bonanza', 11, { hot: true }), game('Gates of Olympus', 12, { hot: true }), game('Sugar Rush', 13, { isNew: true }), game('Big Bass Bonanza', 14), game('The Dog House', 20), game('Wolf Gold', 17)] },
    { id: 'playngo', name: text("Play'n GO"), games: [game('Book of Dead', 16, { hot: true }), game('Rise of Olympus', 31), game('Reactoonz', 32, { hot: true }), game('Moon Princess', 33), game('Fire Joker', 34)] },
    { id: 'netent', name: text('NetEnt'), games: [game('Starburst', 15, { hot: true }), game("Gonzo's Quest", 35), game('Dead or Alive 2', 36, { hot: true }), game('Twin Spin', 37), game('Divine Fortune', 30)] },
    { id: 'microgaming', name: text('Microgaming'), games: [game('Mega Moolah', 29, { hot: true }), game('Immortal Romance', 38), game('Thunderstruck II', 39), game('Break da Bank', 40)] },
    { id: 'hacksaw', name: text('Hacksaw Gaming'), games: [game('Wanted Dead or a Wild', 41, { hot: true }), game('Le Bandit', 42, { isNew: true }), game('Chaos Crew', 43), game('RIP City', 44)] },
    { id: 'nolimit', name: text('Nolimit City'), games: [game('Mental', 45, { hot: true }), game('San Quentin', 46), game('Tombstone', 47, { isNew: true }), game('Fire in the Hole', 48, { hot: true })] },
  ],
  pesca: [
    { id: 'jili', name: text('JILI'), games: [game('Royal Fishing', 63), game('Dragon Fortune', 62, { hot: true }), game('Mega Fishing', 64, { isNew: true }), game('Boom Legend', 61), game('Jackpot Fishing', 68, { hot: true })] },
    { id: 'fachai', name: text('Fa Chai'), games: [game('Fishing Yilufa', 51), game('All-Star Fishing', 52), game('Dao Fishing', 53, { isNew: true }), game('Da Sheng Nao Hai', 54)] },
    { id: 'cq9', name: text('CQ9 Gaming'), games: [game('Fishing War', 55, { hot: true }), game('Money Tree Fishing', 56), game('Bao Chuan Fishing', 57)] },
    { id: 'jdb', name: text('JDB'), games: [game('Fishing Disco', 58), game('Cai Shen Fishing', 59, { hot: true }), game('Dragon Fishing II', 60)] },
    { id: 'spade', name: text('Spadegaming'), games: [game('Fishing God', 65, { hot: true }), game('Alien Hunter', 66), game('Zombie Party', 67)] },
  ],
  deportes: [
    { id: 'futbol', icon: 'fa-futbol', name: { es: 'Fútbol', en: 'Soccer', zh: '足球' }, games: [game('MLS', 71, { hot: true }), game('Champions League', 74), game('Premier League', 91), game('LaLiga', 92), game('US Open Cup', 93)] },
    { id: 'basket', icon: 'fa-basketball', name: { es: 'Baloncesto', en: 'Basketball', zh: '篮球' }, games: [game('NBA', 72, { hot: true }), game('Euroliga', 94), game('NBA G League', 95)] },
    { id: 'nfl', icon: 'fa-football', name: { es: 'Fútbol Am.', en: 'Am. Football', zh: '美式足球' }, games: [game('NFL', 73, { hot: true }), game('NCAA', 96)] },
    { id: 'beis', icon: 'fa-baseball', name: { es: 'Béisbol', en: 'Baseball', zh: '棒球' }, games: [game('MLB', 77), game('MiLB', 97, { isNew: true })] },
    { id: 'tenis', icon: 'fa-table-tennis-paddle-ball', name: { es: 'Tenis', en: 'Tennis', zh: '网球' }, games: [game('ATP', 75), game('WTA', 98), game('Grand Slam', 99, { hot: true })] },
    { id: 'ufc', icon: 'fa-hand-fist', name: { es: 'UFC / MMA', en: 'UFC / MMA', zh: 'UFC / 综合格斗' }, games: [game('UFC', 76, { hot: true }), game('Bellator', 100)] },
    { id: 'esports', icon: 'fa-gamepad', name: { es: 'E-Sports', en: 'E-Sports', zh: '电子竞技' }, games: [game('CS2', 101, { isNew: true }), game('Dota 2', 102), game('League of Legends', 103, { hot: true })] },
  ],
  envivo: [
    { id: 'evolution', name: text('Evolution'), games: [game('Crazy Time', 25, { hot: true }), game('Lightning Roulette', 26, { hot: true }), game('Monopoly Live', 27), game('Mega Ball', 81), game('Dream Catcher', 82), game('Crazy Coin Flip', 84, { isNew: true })] },
    { id: 'pragmaticlive', name: text('Pragmatic Live'), games: [game('Mega Wheel', 85, { hot: true }), game('Sweet Bonanza CandyLand', 83, { isNew: true }), game('Boom City', 86), game('ONE Blackjack', 87)] },
    { id: 'ezugi', name: text('Ezugi'), games: [game('Ezugi Roulette', 88), game('Teen Patti', 89, { hot: true }), game('Andar Bahar', 90)] },
    { id: 'vivo', name: text('Vivo Gaming'), games: [game('VIP Blackjack', 21), game('Baccarat', 23), game("Casino Hold'em", 24, { hot: true })] },
  ],
}

export const ppmxTrustItems: PpmxTrustItem[] = [
  {
    id: 'instant-payments',
    icon: 'fa-bolt',
    title: { es: 'Pagos al instante', en: 'Instant payments', zh: '即时支付' },
    description: { es: 'Retiros bancarios rápidos', en: 'Fast bank withdrawals', zh: '银行提现快速到账' },
  },
  {
    id: 'ssl',
    icon: 'fa-lock',
    title: { es: 'Cifrado SSL', en: 'SSL encryption', zh: 'SSL 加密' },
    description: { es: 'Tus datos van blindados', en: 'Your data is protected', zh: '数据安全保护' },
  },
  {
    id: 'support',
    icon: 'fa-headset',
    title: { es: 'Soporte 24/7', en: '24/7 support', zh: '7x24 支持' },
    description: { es: 'Siempre hay alguien', en: 'Someone is always available', zh: '随时有人响应' },
  },
  {
    id: 'license',
    icon: 'fa-certificate',
    title: { es: 'Responsible Gaming', en: 'Responsible gaming', zh: '责任博彩' },
    description: { es: 'Operación 100% legal', en: '100% legal operation', zh: '100% 合规运营' },
  },
]

export const ppmxFooterColumns: PpmxFooterColumn[] = [
  {
    id: 'play',
    title: { es: 'Juega', en: 'Play', zh: '游戏' },
    links: [
      { id: 'wheel', label: { es: 'Ruleta', en: 'Wheel', zh: '轮盘' }, target: 'wheel' },
      { id: 'casino', label: { es: 'Casino', en: 'Casino', zh: '娱乐场' }, target: 'casino' },
      { id: 'promotions', label: { es: 'Promociones', en: 'Promotions', zh: '优惠活动' }, target: 'promo' },
    ],
  },
  {
    id: 'support',
    title: { es: 'Soporte', en: 'Support', zh: '支持' },
    links: [
      { id: 'help', label: { es: 'Ayuda', en: 'Help', zh: '帮助' }, target: 'help' },
      { id: 'terms', label: { es: 'Términos', en: 'Terms', zh: '条款' }, target: 'terms' },
      { id: 'privacy', label: { es: 'Privacidad', en: 'Privacy', zh: '隐私' }, target: 'privacy' },
      {
        id: 'responsible',
        label: { es: 'Juego Responsable', en: 'Responsible Gaming', zh: '责任博彩' },
        target: 'responsible',
      },
    ],
  },
  {
    id: 'company',
    title: { es: 'Empresa', en: 'Company', zh: '公司' },
    links: [
      { id: 'about', label: { es: 'Sobre nosotros', en: 'About us', zh: '关于我们' }, target: 'about' },
      { id: 'affiliates', label: { es: 'Afiliados', en: 'Affiliates', zh: '联盟合作' }, target: 'affiliates' },
      { id: 'press', label: { es: 'Prensa', en: 'Press', zh: '媒体' }, target: 'press' },
    ],
  },
]

export const ppmxSocialItems: PpmxSocialItem[] = [
  {
    id: 'facebook',
    icon: 'fa-facebook',
    label: 'Facebook',
    target: 'facebook',
    url: 'https://www.facebook.com/profile.php?id=61591107827237',
    handle: '/ppmx',
    tone: 'facebook',
  },
  {
    id: 'instagram',
    icon: 'fa-instagram',
    label: 'Instagram',
    target: 'instagram',
    url: 'https://www.instagram.com/ppus/',
    handle: '@ppus',
    tone: 'instagram',
  },
  {
    id: 'telegram',
    icon: 'fa-telegram',
    label: 'Telegram',
    target: 'telegram',
    url: 'https://t.me/ppus',
    handle: '@ppus',
    tone: 'telegram',
  },
  {
    id: 'tiktok',
    icon: 'fa-tiktok',
    label: 'TikTok',
    target: 'tiktok',
    url: 'https://www.tiktok.com/@pp.us',
    handle: '@ppus',
    tone: 'tiktok',
  },
]

export const ppmxSocialText = {
  close: {
    es: 'Cerrar redes oficiales',
    en: 'Close official channels',
    zh: '关闭官方社媒弹窗',
  },
  title: {
    es: 'Redes oficiales',
    en: 'Official channels',
    zh: '官方社交媒体',
  },
  sub: {
    es: 'Síguenos para promos y ganadores. Cuidado con cuentas falsas.',
    en: 'Follow us for promos and winners. Beware of fake accounts.',
    zh: '关注我们获取优惠与中奖喜讯，谨防假冒账号。',
  },
  open: {
    es: 'Abrir canal',
    en: 'Open channel',
    zh: '打开渠道',
  },
} satisfies Record<string, LocalizedText>

export const ppmxFloatingWidgets: PpmxFloatingWidget[] = [
  { id: 'checkin', icon: 'fa-calendar-check', label: { es: 'Check-in', en: 'Check-in', zh: '签到' }, target: 'checkin' },
  { id: 'social', icon: 'fa-share-nodes', label: { es: 'Síguenos', en: 'Follow us', zh: '关注我们' }, target: 'social' },
  { id: 'download', icon: 'fa-mobile-screen-button', label: { es: 'App', en: 'App', zh: 'App' }, target: 'download' },
]

export const ppmxCheckinText = {
  close: {
    es: 'Cerrar check-in',
    en: 'Close check-in',
    zh: '关闭签到弹窗',
  },
  title: {
    es: 'Check-in diario',
    en: 'Daily check-in',
    zh: '每日签到',
  },
  sub: {
    es: 'Conéctate cada día y reclama tu bono. 7 días seguidos desbloquean el premio mayor.',
    en: 'Log in every day and claim your bonus. Seven consecutive days unlock the grand prize.',
    zh: '每天登录领取红利，连续 7 天可解锁终极大奖。',
  },
  loading: {
    es: 'Cargando check-in...',
    en: 'Loading check-in...',
    zh: '正在加载签到...',
  },
  failed: {
    es: 'No se pudo cargar el check-in. Inténtalo otra vez.',
    en: 'Could not load check-in. Try again.',
    zh: '签到加载失败，请重试。',
  },
  retry: {
    es: 'Reintentar',
    en: 'Retry',
    zh: '重试',
  },
  empty: {
    es: 'Todavía no hay recompensas configuradas.',
    en: 'No check-in rewards are configured yet.',
    zh: '暂无签到奖励配置。',
  },
  mystery: {
    es: 'Bono misterioso',
    en: 'Mystery bonus',
    zh: '神秘奖励',
  },
  revealed: {
    es: 'Bono revelado',
    en: 'Revealed bonus',
    zh: '已揭晓奖励',
  },
  pending: {
    es: 'Pendiente',
    en: 'Pending',
    zh: '待领取',
  },
  grand: {
    es: 'Premio mayor',
    en: 'Grand prize',
    zh: '终极大奖',
  },
  claim: {
    es: 'Reclamar hoy',
    en: 'Claim today',
    zh: '领取今日奖励',
  },
  claimed: {
    es: 'Reclamado hoy',
    en: 'Claimed today',
    zh: '今日已领取',
  },
  completed: {
    es: 'Ciclo completado',
    en: 'Cycle completed',
    zh: '周期已完成',
  },
  submitting: {
    es: 'Reclamando...',
    en: 'Claiming...',
    zh: '领取中...',
  },
  success: {
    es: 'Bono diario reclamado',
    en: 'Daily bonus claimed',
    zh: '每日红利已领取',
  },
  records: {
    es: 'Últimos reclamos',
    en: 'Recent claims',
    zh: '最近领取',
  },
} satisfies Record<string, LocalizedText>
