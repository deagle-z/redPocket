import type {
  PpmxAccountHubItem,
  PpmxDownloadFeature,
  PpmxDownloadStep,
  PpmxGuideCard,
  PpmxPromoItem,
  PpmxProfileField,
  PpmxTeamKpi,
  PpmxWheelPrize,
} from './types'

const text = (value: string) => ({
  en: value,
  es: value,
  zh: value,
})

export const ppmxWheelPrizes: PpmxWheelPrize[] = [
  { amount: 50, color: '#138a52', weight: 20 },
  { amount: 20, color: '#c8102e', weight: 22 },
  { amount: 200, color: '#138a52', weight: 10 },
  { amount: 10, color: '#c8102e', weight: 24 },
  { amount: 100, color: '#138a52', weight: 14 },
  { amount: 500, color: '#c8102e', weight: 6 },
  { amount: 30, color: '#138a52', weight: 22 },
  { amount: 1000, color: '#e6a417', gold: true, weight: 3 },
]

export const ppmxWheelRules = [
  {
    es: 'Un depósito único de ≥S/200 otorga 1 giro gratis. (Con S/400 son dos giros, y así sucesivamente.)',
    en: 'A single deposit of ≥S/200 earns 1 free spin. (S/400 gives 2 spins, and so on.)',
    zh: '单次充值金额 ≧S/200 元赠送一次免费旋转。（若充值 S/400 则赠送两次免费旋转，依次类推）',
  },
  {
    es: 'Por cada S/10,000 en apuestas válidas obtienes 1 giro gratis.',
    en: 'Every S/10,000 in valid wagers earns 1 free spin.',
    zh: '每达到 10000 投注额可享一次免费旋转。',
  },
  {
    es: 'Todos los premios son monedas de oro PP que se acreditan como saldo real en tu cuenta.',
    en: 'All prizes are PP gold coins credited as real balance to your account.',
    zh: '所有奖励均为 PP 金币，将作为真实余额存入账户。',
  },
  {
    es: 'Al detenerse la ruleta, el premio se acredita automáticamente a tu saldo.',
    en: 'Once the wheel stops, winnings are automatically credited to your balance.',
    zh: '转盘停止后，奖金将自动存入您的余额。',
  },
  {
    es: 'Todos los premios requieren completar un rollover de 10x.',
    en: 'All prizes are subject to a 10x wagering requirement.',
    zh: '所获的奖励需完成 10 倍流水。',
  },
  {
    es: 'Promoción exclusiva para mayores de 21 años. Apuesta responsablemente.',
    en: 'Promotion for users aged 21 and above. Please gamble responsibly.',
    zh: '活动仅限 18 岁以上用户，请理性博彩。',
  },
]

export const ppmxTeamKpis: PpmxTeamKpi[] = [
  {
    id: 'invited',
    icon: 'fa-user-plus',
    tone: 'gold',
    value: '128',
    label: { es: 'Usuarios invitados', en: 'Invited users', zh: '邀请用户' },
  },
  {
    id: 'valid',
    icon: 'fa-user-check',
    tone: 'green',
    value: '86',
    label: { es: 'Usuarios válidos', en: 'Valid users', zh: '有效用户' },
  },
  {
    id: 'today',
    icon: 'fa-bolt',
    tone: 'red',
    value: '5',
    label: { es: 'Invitados hoy', en: 'Invited today', zh: '今日邀请' },
  },
  {
    id: 'total',
    icon: 'fa-users',
    tone: 'blue',
    value: '342',
    label: { es: 'Invitación acumulada', en: 'Total invitations', zh: '累计邀请' },
  },
]

export const ppmxCommissionSummary = [
  { id: 'accumulated', value: 'S/12,480.00', label: { es: 'Acumulada', en: 'Accumulated', zh: '累计佣金' } },
  { id: 'month', value: 'S/1,920.00', label: { es: 'Este mes', en: 'This month', zh: '本月佣金' } },
  { id: 'pending', value: 'S/640.00', label: { es: 'Pendiente', en: 'Pending', zh: '待结算' } },
]

export const ppmxPromoItems: PpmxPromoItem[] = [
  {
    id: 'registro',
    tag: { es: 'Regalo de bienvenida', en: 'Welcome gift', zh: '注册礼金' },
    title: { es: 'Regístrate y gana S/3 PEN gratis', en: 'Sign up and get S/3 PEN free', zh: '注册立即赠送 3 比索' },
    sub: {
      es: 'Crea tu cuenta en 1 minuto y recibe S/3 PEN al instante para empezar a jugar.',
      en: 'Create your account in 1 minute and get S/3 PEN instantly to start playing.',
      zh: '1 分钟完成注册，立即获得 3 比索，马上开玩。',
    },
    howTo: {
      es: 'Regístrate con tu teléfono, verifica tu cuenta y los S/3 PEN se acreditan automáticamente.',
      en: 'Register with your phone, verify your account and the S/3 PEN is credited automatically.',
      zh: '点击下方立即激活优惠后前往我的账户领取 3 比索新用户注册奖励。',
    },
    valid: { es: 'Promoción para nuevos usuarios', en: 'New users only', zh: '仅限新用户' },
    terms: {
      es: '1 por persona/dispositivo. Rollover 10x. Aplican T&C. +18.',
      en: '1 per person/device. 10x rollover. T&C apply. 18+.',
      zh: '每人/每设备限领一次，10 倍流水，适用条款，18 岁以上。',
    },
    tone: 'green',
    backgroundImage: '/images/ppmx/promos/registro.jpg',
    backgroundPosition: 'right center',
  },
  {
    id: 'checkin',
    tag: { es: 'Check-in de bienvenida', en: 'Welcome check-in', zh: '新人签到' },
    title: { es: 'Check-in de nuevos: recibe S/1 cada día', en: 'New-user check-in: get S/1 every day', zh: '新用户签到每天送 1 比索' },
    sub: {
      es: 'Regístrate y entra cada día para reclamar tu recompensa; el día 7 abre el cofre mayor.',
      en: 'Sign up and log in daily to claim your reward; day 7 unlocks the grand chest.',
      zh: '注册后每日登录即可签到领奖，连续 7 天第 7 天开启终极宝箱。',
    },
    howTo: {
      es: 'Tras registrarte, entra cada día y pulsa Check-in para acreditar la recompensa diaria. 7 días consecutivos liberan el cofre del día 7.',
      en: 'After signing up, log in each day and tap Check-in to credit your daily reward. 7 consecutive days unlock the day-7 chest.',
      zh: '注册后每天登录点击下方立即激活优惠或首页「新人签到」即可领取当日奖励，连续签到 7 天解锁第 7 天宝箱大奖。',
    },
    valid: { es: 'Para nuevos usuarios · 7 días', en: 'New users · 7 days', zh: '仅限新用户 · 7 天内' },
    terms: {
      es: 'Solo nuevos usuarios. 1 check-in por día. Rollover 10x. +18.',
      en: 'New users only. 1 check-in per day. 10x rollover. 18+.',
      zh: '仅限新用户，每日限签到一次，10 倍流水，18 岁以上。',
    },
    tone: 'gold',
    backgroundImage: '/images/ppmx/promos/checkin.jpg',
    backgroundPosition: 'right center',
  },
  {
    id: 'vip',
    tag: { es: 'Programa VIP', en: 'VIP program', zh: 'VIP 俱乐部' },
    title: { es: 'Sube de nivel VIP y reclama hasta S/188,888', en: 'Level up VIP and claim up to S/188,888', zh: 'VIP 升级奖励 最高 S/188,888' },
    sub: {
      es: 'Más rango = más recompensas: bonos, salario ampliado y un gestor personal.',
      en: 'Higher rank = more rewards: bonuses, bigger salary and a personal host.',
      zh: '等级越高奖励越多：升级红利、更高周薪与专属客户经理。',
    },
    howTo: {
      es: 'Acumula puntos jugando. Al alcanzar cada nivel, reclama tu bono de ascenso desde tu perfil VIP.',
      en: 'Earn points by playing. On reaching each level, claim your upgrade bonus from your VIP profile.',
      zh: '通过游戏累积积分，达成每个等级后即可在 VIP 中心领取升级红利。',
    },
    valid: { es: 'Recompensas permanentes', en: 'Permanent rewards', zh: '长期权益' },
    terms: {
      es: 'Bono según nivel alcanzado. Rollover 10x. +18.',
      en: 'Bonus by level reached. 10x rollover. 18+.',
      zh: '红利依达成等级而定，10 倍流水，18 岁以上。',
    },
    tone: 'violet',
    backgroundImage: '/images/ppmx/promos/vip.jpg',
    backgroundPosition: 'right center',
  },
  {
    id: 'invita',
    tag: { es: 'Programa de afiliados', en: 'Referral program', zh: '邀请返佣' },
    title: { es: 'Hasta 50% de comisión', en: 'Up to 50% rebate', zh: '最高50%返佣' },
    sub: {
      es: 'Comparte tu enlace exclusivo e invita a amigos a registrarse y recargar: gana hasta 50% de comisión sobre las recargas de cada usuario válido, ¡de por vida!\n\nPrimera recarga: 40% de comisión\nSegunda recarga: 45% de comisión\nTercera recarga o más: 50% de comisión\n\nLas recargas de todos los miembros de tu equipo te generan ingresos continuos: ¡mientras más invites, más ganas!',
      en: "Share your exclusive link and invite friends to register and deposit — earn up to 50% commission on each valid user's deposits, for life!\n\nFirst deposit: 40% rebate\nSecond deposit: 45% rebate\nThird deposit and above: 50% rebate\n\nEvery team member's deposits bring you ongoing income — the more you invite, the more you earn!",
      zh: '分享您的专属链接，邀请好友注册充值，每位有效用户最高可享充值金额50%佣金，终身有效！\n\n首次充值：40% 返佣\n第二次充值：45% 返佣\n第三次及以上充值：50% 返佣\n\n团队所有成员的充值均可为您带来持续收益，邀请越多，赚得越多！',
    },
    howTo: {
      es: 'Comparte tu enlace exclusivo desde Mi equipo e invita a tus amigos a registrarse y recargar.',
      en: 'Copy your referral link from My team, share it and collect commissions every week.',
      zh: '点击下方立即激活优惠或在「我的团队」复制专属链接并分享，实时结算返佣。',
    },
    valid: { es: 'Comisión recurrente en cada recarga', en: 'Recurring commission on every deposit', zh: '每笔充值都返佣 上不封顶' },
    terms: {
      es: 'Comisión según el número de recarga del referido. +18.',
      en: 'Commission based on referral deposit count. 18+.',
      zh: '按下级充值次数计算返佣，18 岁以上。',
    },
    tone: 'pink',
    backgroundImage: '/images/ppmx/promos/invita.jpg',
    backgroundPosition: 'right center',
  },
  {
    id: 'recarga',
    tag: { es: 'Bono de recarga', en: 'Reload bonus', zh: '充值赠送' },
    title: { es: 'Recarga y recibe hasta 18% de regalo', en: 'Deposit and get up to 18% bonus', zh: '最高18%赠送' },
    sub: {
      es: 'La primera recarga recibe un 8% de reembolso, la segunda un 9%, la tercera un 10% y desde la cuarta un 18%. Cuantas más recargas hagas, mayor será el reembolso.',
      en: 'The first deposit gets an 8% rebate, the second 9%, the third 10%, and the fourth and subsequent deposits 18%. The more times you deposit, the higher your rebate.',
      zh: '首次充值享8%返利，二次充值享9%返利，三次充值享10%返利，第四次及以上充值享18%返利。充值的次数越多享受的返利越高！',
    },
    howTo: {
      es: 'Ve a Depositar, elige tu método, ingresa el monto y el reembolso según tu número de recarga se aplica automáticamente.',
      en: 'Go to Deposit, pick your method, enter the amount and the rebate for your deposit count applies automatically.',
      zh: '点击下方立即激活优惠或前往存款页选择方式并输入金额，红利自动到账。',
    },
    valid: { es: 'Válido todos los días', en: 'Valid every day', zh: '每日有效' },
    terms: {
      es: 'Bono según número de recarga. Rollover 10x. +18.',
      en: 'Bonus by deposit count. 10x rollover. 18+.',
      zh: '按充值次数赠送，10 倍流水，18 岁以上。',
    },
    tone: 'red',
    backgroundImage: '/images/ppmx/promos/recarga.jpg',
    backgroundPosition: 'right center',
  },
  {
    id: 'download',
    tag: { es: 'App móvil', en: 'Mobile app', zh: '下载 APP' },
    title: { es: 'Descarga la app PP.PE y gana S/8', en: 'Download the PP.PE app and get S/30', zh: '下载 PP.PE APP 立即领 S/8' },
    sub: {
      es: 'Juega donde quieras: descarga la app, inicia sesión y recibe S/30 PEN al instante.',
      en: 'Play anywhere: download the app, log in and get S/30 PEN instantly.',
      zh: '随时随地畅玩，下载 APP 并登录即送 8 PEN，还有专属推送福利。',
    },
    howTo: {
      es: 'Pulsa Descargar, instala la app PP.PE, inicia sesión y el bono de S/30 se acredita solo.',
      en: 'Tap Download, install the PP.PE app, log in and the S/30 bonus is credited automatically.',
      zh: '点击下载并安装 PP.PE APP，登录账户后 8 PEN自动到账。',
    },
    valid: { es: 'Solo nuevas instalaciones', en: 'New installs only', zh: '仅限新下载用户' },
    terms: {
      es: '1 por cuenta/dispositivo. Rollover 10x. +18.',
      en: '1 per account/device. 10x rollover. 18+.',
      zh: '每账户/设备限领一次，10 倍流水，18 岁以上。',
    },
    tone: 'blue',
    backgroundImage: '/images/ppmx/promos/download.jpg',
    backgroundPosition: 'right center',
  },
  {
    id: 'wheel',
    tag: { es: 'Ruleta de la suerte', en: 'Lucky wheel', zh: '幸运转盘' },
    title: { es: 'Gira la ruleta y gana hasta S/10,000', en: 'Spin the wheel and win up to S/10,000', zh: '幸运转盘每日转 最高赢 S/888888' },
    sub: {
      es: 'Cada día tienes un giro gratis: bonos, giros y hasta S/10,000 PEN te esperan. Por cada recarga de ≥S/200 recibes un giro gratis (con S/400 son dos giros, y así sucesivamente).',
      en: 'Every day you get a free spin: bonuses, spins and up to S/10,000 PEN await. Each single deposit of ≥S/200 grants one free spin (S/400 gives two spins, and so on).',
      zh: '参与游戏获得免费旋转，获得奖励、免费旋转及最高 888888 PEN等你拿。（每达到 10000 流水即可自动获得一次免费幸运转盘旋转次数）单次充值金额≧S/200元赠送一次免费旋转。（若充值S/400则赠送两次免费旋转依次类推）',
    },
    howTo: {
      es: 'Entra a la Ruleta de la suerte, usa tu giro gratis diario y reclama el premio que caiga.',
      en: 'Open the Lucky Wheel, use your daily free spin and claim whatever prize lands.',
      zh: '点击下方激活优惠或进入「幸运转盘」页面，点击旋转，转到的奖励即可领取。',
    },
    valid: { es: 'Un giro gratis al día', en: 'One free spin per day', zh: '2026 年' },
    terms: {
      es: '1 giro gratis/día. Rollover según premio. +18.',
      en: '1 free spin/day. Rollover varies by prize. 18+.',
      zh: '参与游戏免费获得旋转次数，10 倍流水，18 岁以上。',
    },
    tone: 'pink',
    backgroundImage: '/images/ppmx/promos/wheel.jpg',
    backgroundPosition: 'right center',
  },
]

export const ppmxAccountHubItems: PpmxAccountHubItem[] = [
  {
    id: 'records',
    icon: 'fa-receipt',
    tone: 'green',
    label: { es: 'Historial', en: 'History', zh: '交易记录' },
    description: {
      es: 'Depósitos, retiros y apuestas',
      en: 'Deposits, withdrawals, and bets',
      zh: '存款、提现和投注记录',
    },
  },
  {
    id: 'profile',
    icon: 'fa-id-card',
    tone: 'blue',
    label: { es: 'Perfil', en: 'Profile', zh: '个人资料' },
    description: {
      es: 'Usuario, avatar y datos',
      en: 'Username, avatar, and details',
      zh: '用户名、头像和资料',
    },
  },
  {
    id: 'vip',
    icon: 'fa-crown',
    tone: 'gold',
    target: 'vip',
    label: { es: 'Nivel VIP', en: 'VIP level', zh: 'VIP 等级' },
    description: {
      es: 'Plata · recompensas',
      en: 'Silver · rewards',
      zh: '白银 · 奖励权益',
    },
  },
  {
    id: 'security',
    icon: 'fa-shield-halved',
    tone: 'red',
    label: { es: 'Seguridad', en: 'Security', zh: '账户安全' },
    description: {
      es: 'Contraseña, correo y celular',
      en: 'Password, email, and mobile',
      zh: '密码、邮箱和手机号',
    },
  },
  {
    id: 'support',
    icon: 'fa-headset',
    tone: 'teal',
    label: { es: 'Atención al cliente', en: 'Customer support', zh: '客户支持' },
    description: {
      es: 'Telegram y WhatsApp 24/7',
      en: 'Telegram and WhatsApp 24/7',
      zh: 'Telegram 与 WhatsApp 24/7 在线',
    },
  },
  {
    id: 'social',
    icon: 'fa-share-nodes',
    tone: 'violet',
    label: { es: 'Redes oficiales', en: 'Official channels', zh: '官方社媒' },
    description: {
      es: 'Instagram, Facebook y más',
      en: 'Instagram, Facebook, and more',
      zh: 'Instagram、Facebook 等',
    },
  },
  {
    id: 'promo-code',
    icon: 'fa-ticket',
    tone: 'gold',
    label: { es: 'Canjear código', en: 'Redeem code', zh: '兑换优惠码' },
    description: {
      es: 'Ingresa un código de 6 dígitos',
      en: 'Enter a 6-digit code',
      zh: '输入 6 位数字优惠码',
    },
  },
]

export const ppmxAccountExtraItems: PpmxAccountHubItem[] = [
  {
    id: 'download',
    icon: 'fa-cloud-arrow-down',
    tone: 'red',
    label: { es: 'Descargar App', en: 'Download app', zh: '下载 App' },
    description: {
      es: 'iOS y Android · bonos exclusivos en la app',
      en: 'iOS and Android · app-exclusive bonuses',
      zh: 'iOS 与 Android · App 专属奖励',
    },
  },
  {
    id: 'guide',
    icon: 'fa-circle-question',
    tone: 'gold',
    label: { es: 'Guía de actividades', en: 'Activity guide', zh: '活动指南' },
    description: {
      es: 'Reglas de juego y cómo participar',
      en: 'Game rules and how to join',
      zh: '游戏规则与参与方式',
    },
  },
]

export const ppmxDownloadText = {
  eyebrow: { es: 'App oficial PP.PE', en: 'Official PP.PE app', zh: 'PP.PE 官方 App' },
  title: { es: 'Descarga la App PP.PE', en: 'Download the PP.PE App', zh: '下载 PP.PE App' },
  sub: {
    es: 'Juega más rápido, recibe bonos exclusivos y notificaciones de pago al instante. Disponible para iOS y Android.',
    en: 'Play faster, get exclusive bonuses and instant payment notifications. Available for iOS and Android.',
    zh: '更快进入游戏，领取 App 专属奖励，并即时接收支付通知。支持 iOS 与 Android。',
  },
  appStore: { es: 'App Store', en: 'App Store', zh: 'App Store' },
  android: { es: 'Android APK', en: 'Android APK', zh: 'Android APK' },
  qr: { es: 'Escanea para descargar', en: 'Scan to download', zh: '扫码下载' },
  soon: { es: 'Muy pronto', en: 'Coming very soon', zh: '即将开放' },
  featuresTitle: { es: 'Ventajas de la app', en: 'App advantages', zh: 'App 优势' },
  stepsTitle: { es: 'Cómo instalar', en: 'How to install', zh: '如何安装' },
} satisfies Record<string, import('./types').LocalizedText>

export const ppmxDownloadFeatures: PpmxDownloadFeature[] = [
  {
    id: 'instant',
    icon: 'fa-bolt',
    title: { es: 'Acceso instantáneo', en: 'Instant access', zh: '快速进入' },
    description: {
      es: 'Abre tus juegos favoritos con un toque, sin navegador.',
      en: 'Open your favorite games with one tap, without the browser.',
      zh: '无需浏览器，一键打开你喜欢的游戏。',
    },
  },
  {
    id: 'bonuses',
    icon: 'fa-gift',
    title: { es: 'Bonos solo en app', en: 'App-only bonuses', zh: 'App 专属奖励' },
    description: {
      es: 'Promociones y giros exclusivos para usuarios de la app.',
      en: 'Exclusive promotions and spins for app users.',
      zh: 'App 用户可享专属优惠和免费旋转。',
    },
  },
  {
    id: 'payments',
    icon: 'fa-shield-halved',
    title: { es: 'Pagos bancarios seguros', en: 'Secure bank payments', zh: '安全银行支付' },
    description: {
      es: 'Depósitos y retiros cifrados con confirmación inmediata.',
      en: 'Encrypted deposits and withdrawals with instant confirmation.',
      zh: '存款与提现加密处理，并即时确认。',
    },
  },
  {
    id: 'alerts',
    icon: 'fa-bell',
    title: { es: 'Alertas en tiempo real', en: 'Real-time alerts', zh: '实时提醒' },
    description: {
      es: 'Recibe avisos de bonos, resultados y movimientos al instante.',
      en: 'Get instant alerts for bonuses, results and balance movements.',
      zh: '即时接收红利、结果和资金变动提醒。',
    },
  },
]

export const ppmxDownloadSteps: PpmxDownloadStep[] = [
  {
    id: 'scan',
    title: { es: 'Escanea el código QR', en: 'Scan the QR code', zh: '扫描二维码' },
    description: {
      es: 'Apunta la cámara de tu celular al QR de arriba.',
      en: 'Point your phone camera at the QR above.',
      zh: '用手机摄像头对准上方二维码。',
    },
  },
  {
    id: 'download',
    title: { es: 'Descarga el instalador', en: 'Download the installer', zh: '下载安装包' },
    description: {
      es: 'Elige App Store (iOS) o el archivo APK (Android).',
      en: 'Choose App Store for iOS or the APK file for Android.',
      zh: '选择 App Store（iOS）或 APK 文件（Android）。',
    },
  },
  {
    id: 'allow',
    title: { es: 'Permite la instalación', en: 'Allow installation', zh: '允许安装' },
    description: {
      es: 'En Android, activa orígenes desconocidos si se solicita.',
      en: 'On Android, enable unknown sources if prompted.',
      zh: 'Android 如有提示，请允许未知来源安装。',
    },
  },
  {
    id: 'login',
    title: { es: 'Inicia sesión y juega', en: 'Sign in and play', zh: '登录并开始游戏' },
    description: {
      es: 'Entra con tu cuenta y reclama tu bono de bienvenida en la app.',
      en: 'Sign in with your account and claim your welcome bonus in the app.',
      zh: '使用账号登录，并在 App 中领取欢迎奖励。',
    },
  },
]

export const ppmxGuideText = {
  eyebrow: { es: 'Centro de actividades', en: 'Activity center', zh: '活动中心' },
  title: { es: 'Guía de actividades', en: 'Activity guide', zh: '活动讲解' },
  sub: {
    es: 'Conoce cada promoción, sus reglas y cómo participar para aprovechar al máximo tus recompensas.',
    en: 'Learn about each promotion, its rules and how to take part to make the most of your rewards.',
    zh: '了解每项活动的规则与参与方式，把奖励最大化。',
  },
  foot: {
    es: 'Aplican términos y condiciones. Juega de forma responsable · +18.',
    en: 'Terms and conditions apply. Play responsibly · 18+.',
    zh: '适用条款与条件，请理性博彩 · 18+。',
  },
} satisfies Record<string, import('./types').LocalizedText>

export const ppmxGuideCards: PpmxGuideCard[] = [
  {
    id: 'checkin',
    icon: 'fa-calendar-check',
    tagIcon: 'fa-coins',
    tone: 'green',
    title: { es: 'Check-in diario', en: 'Daily check-in', zh: '每日签到' },
    description: {
      es: 'Inicia sesión cada día y reclama tu recompensa. Mientras más días seguidos, mayor el premio acumulado de 7 días.',
      en: 'Sign in every day and claim your reward. The more consecutive days, the bigger the 7-day cumulative prize.',
      zh: '每天登录领取奖励，连续签到天数越多，7 天累计奖励越丰厚。',
    },
    tag: { es: 'S/1 por día', en: 'S/1 per day', zh: '每天签到送 S/1' },
  },
  {
    id: 'wheel',
    icon: 'fa-dharmachakra',
    tagIcon: 'fa-bolt',
    tone: 'gold',
    title: { es: 'Ruleta de la suerte', en: 'Lucky wheel', zh: '幸运转盘' },
    description: {
      es: 'Juega para ganar giros gratis: premios, giros y hasta S/888,888 PEN te esperan. (Por cada S/10,000 de apuesta acumulada obtienes un giro gratis automático.)',
      en: 'Play to earn free spins: prizes, free spins and up to S/888,888 PEN await. (Every S/10,000 wagered automatically grants one free wheel spin.)',
      zh: '参与游戏获得免费旋转，获得奖励、免费旋转及最高 888888 PEN等你拿。（每达到 10000 流水即可自动获得一次免费幸运转盘旋转次数）',
    },
    tag: { es: 'Giro gratis diario', en: 'Daily free spin', zh: '每日免费旋转' },
  },
  {
    id: 'invita',
    icon: 'fa-user-group',
    tagIcon: 'fa-chart-line',
    tone: 'blue',
    title: { es: 'Comisión por referidos', en: 'Referral commission', zh: '邀请返佣' },
    description: {
      es: 'Comparte tu enlace exclusivo y gana comisión continua cuando tus amigos participan.',
      en: 'Share your exclusive link and keep earning commission when your friends play.',
      zh: '分享专属链接，好友参与后可持续获得佣金。',
    },
    tag: { es: 'Hasta 50% por recarga', en: 'Up to 50% per deposit', zh: '每笔充值最高返佣 50%' },
  },
  {
    id: 'vip',
    icon: 'fa-crown',
    tagIcon: 'fa-gem',
    tone: 'violet',
    title: { es: 'Club VIP', en: 'VIP club', zh: 'VIP 俱乐部' },
    description: {
      es: 'Sube de nivel según tu actividad y los requisitos de ascenso; desbloquea salario semanal, bonos de nivel y servicio prioritario.',
      en: 'Level up through your activity and upgrade requirements to unlock weekly salary, level-up bonuses and priority service.',
      zh: '通过活跃度及升级要求提升等级，解锁周薪、升级奖励和优先服务。',
    },
    tag: { es: 'Salario semanal + bonos', en: 'Weekly salary + bonuses', zh: '周薪 + 升级奖励' },
  },
  {
    id: 'recarga',
    icon: 'fa-hand-holding-dollar',
    tagIcon: 'fa-gift',
    tone: 'red',
    title: { es: 'Bono de bienvenida', en: 'Welcome bonus', zh: '欢迎奖金' },
    description: {
      es: 'Completa tu registro para activar tu recompensa de bienvenida de nuevo usuario.',
      en: 'Complete your registration to activate your new-user welcome reward.',
      zh: '完成注册即可激活新人欢迎奖励。',
    },
    tag: { es: '100% · S/3 PEN al instante', en: '100% · S/3 PEN instantly', zh: '100% 立即获得 S/3 比索' },
  },
  {
    id: 'download',
    icon: 'fa-mobile-screen-button',
    tagIcon: 'fa-gift',
    tone: 'blue',
    title: { es: 'App móvil', en: 'Mobile app', zh: '下载 APP' },
    description: {
      es: 'Juega donde quieras: descarga la app, inicia sesión y recibe S/8 PEN, con beneficios exclusivos por notificaciones.',
      en: 'Play anywhere: download the app, log in and get S/8 PEN, plus exclusive push perks.',
      zh: '随时随地畅玩，下载 APP 并登录即送 8 PEN，还有专属推送福利。',
    },
    tag: { es: 'Inicia sesión y gana S/8 PEN', en: 'Log in & get S/8 PEN', zh: '登录即送 S/8 PEN' },
  },
]

export const ppmxProfileFields: PpmxProfileField[] = [
  { id: 'first-name', label: { es: 'Nombre', en: 'First name', zh: '名字' }, model: 'firstName' },
  { id: 'email', label: { es: 'Email', en: 'Email', zh: '邮箱' }, model: 'email', type: 'email' },
  { id: 'phone', label: { es: 'Teléfono', en: 'Phone', zh: '电话' }, model: 'phone', type: 'tel' },
]
