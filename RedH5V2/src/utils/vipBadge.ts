let _uid = 0

export function vipBadgeSVG(i: number, cls?: string): string {
  const u = ++_uid
  const royal = i >= 10
  const PI = Math.PI
  const t = Math.max(0, Math.min(1, i / 10))

  const lerpC = (a: string, b: string, k: number): string => {
    const pa = a.match(/../g)!.map(h => parseInt(h, 16))
    const pb = b.match(/../g)!.map(h => parseInt(h, 16))
    return '#' + pa.map((v, j) => Math.round(v + (pb[j] - v) * k).toString(16).padStart(2, '0')).join('')
  }

  const gTop = lerpC('e7c878', 'fff4c6', t)
  const gMid = lerpC('bb8b2c', 'ecc457', t)
  const gBot = lerpC('5b3e10', '8c6618', t)
  const gHi  = lerpC('fff4cf', 'fffbe9', t)

  const glowOp  = (0.12 + 0.05 * i).toFixed(3)
  const nLeaf   = 5 + Math.round(i * 0.4)
  const beads    = i >= 5
  const crownGem = i >= 6
  const studRuby = i >= 7
  const sideGems = i >= 9
  const extraRim = i >= 4

  const P2 = (r: number, d: number): [number, number] => [
    +(60 + r * Math.cos(d * PI / 180)).toFixed(2),
    +(60 + r * Math.sin(d * PI / 180)).toFixed(2),
  ]

  let defs = ''
  let g    = ''

  // gradients
  defs += `<radialGradient id="vglow${u}" cx="50%" cy="32%" r="62%"><stop offset="0" stop-color="#ffe9a6" stop-opacity="${glowOp}"/><stop offset="55%" stop-color="#f4b63e" stop-opacity="${(parseFloat(glowOp) * 0.4).toFixed(3)}"/><stop offset="100%" stop-color="#f4b63e" stop-opacity="0"/></radialGradient>`

  const goldStops = `<stop offset="0" stop-color="${gHi}"/><stop offset=".15" stop-color="${gTop}"/><stop offset=".4" stop-color="${gMid}"/><stop offset=".52" stop-color="${gBot}"/><stop offset=".7" stop-color="${gMid}"/><stop offset=".9" stop-color="${gTop}"/><stop offset="1" stop-color="${lerpC('f0d488', 'fff0bf', t)}"/>`
  defs += `<linearGradient id="vgold${u}" x1="0" y1="0" x2="0" y2="1">${goldStops}</linearGradient>`
  defs += `<linearGradient id="vgoldr${u}" x1="0" y1="1" x2="0" y2="0">${goldStops}</linearGradient>`
  defs += `<radialGradient id="vdisc${u}" cx="50%" cy="36%" r="66%"><stop offset="0" stop-color="#352818"/><stop offset="42%" stop-color="#1b140b"/><stop offset="78%" stop-color="#0c0805"/><stop offset="100%" stop-color="#040201"/></radialGradient>`
  defs += `<radialGradient id="vsheen${u}" cx="50%" cy="28%" r="56%"><stop offset="0" stop-color="#fff" stop-opacity=".13"/><stop offset="55%" stop-color="#fff" stop-opacity=".02"/><stop offset="100%" stop-color="#fff" stop-opacity="0"/></radialGradient>`
  defs += `<radialGradient id="vspec${u}" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="#fff" stop-opacity="${(0.10 + 0.03 * i).toFixed(3)}"/><stop offset="60%" stop-color="#fff" stop-opacity="${(0.03 + 0.01 * i).toFixed(3)}"/><stop offset="100%" stop-color="#fff" stop-opacity="0"/></radialGradient>`
  defs += `<linearGradient id="vtext${u}" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#fff7d4"/><stop offset=".42" stop-color="${gMid}"/><stop offset=".56" stop-color="${gBot}"/><stop offset="1" stop-color="${lerpC('caa044', 'f0c659', t)}"/></linearGradient>`
  defs += `<radialGradient id="vstud${u}" cx="40%" cy="32%" r="70%"><stop offset="0" stop-color="#fff6cf"/><stop offset=".5" stop-color="${gMid}"/><stop offset="1" stop-color="${gBot}"/></radialGradient>`
  defs += `<linearGradient id="vleaf${u}" x1="0" y1="1" x2="1" y2="0"><stop offset="0" stop-color="${gBot}"/><stop offset=".5" stop-color="${gMid}"/><stop offset="1" stop-color="${gTop}"/></linearGradient>`
  defs += `<radialGradient id="vruby${u}" cx="40%" cy="32%" r="66%"><stop offset="0" stop-color="#ff8a78"/><stop offset=".5" stop-color="#d4182a"/><stop offset="1" stop-color="#6e0a12"/></radialGradient>`

  // ambient halo
  g += `<circle cx="60" cy="60" r="56" fill="url(#vglow${u})"/>`
  if (royal) g += `<circle cx="60" cy="60" r="59" fill="none" stroke="#ffe1a0" stroke-opacity=".45" stroke-width="1.5"/>`

  // outer polished-gold bevel
  g += `<circle cx="60" cy="60" r="58" fill="url(#vgold${u})"/>`
  g += `<circle cx="60" cy="60" r="58" fill="none" stroke="#3a2706" stroke-opacity=".5" stroke-width=".7"/>`
  if (extraRim) g += `<circle cx="60" cy="60" r="56.3" fill="none" stroke="#fff6d6" stroke-opacity=".4" stroke-width=".7"/>`

  // dark groove
  g += `<circle cx="60" cy="60" r="54.4" fill="#0d0905"/>`

  // beaded course (42 soft spheres)
  let bd = ''
  for (let k = 0; k < 42; k++) {
    const p = P2(51.4, k * 360 / 42)
    bd += `<circle cx="${p[0]}" cy="${p[1]}" r="1.55" fill="url(#vstud${u})" stroke="${gBot}" stroke-opacity=".5" stroke-width=".3"/>`
  }
  g += bd

  // inner polished-gold ring
  g += `<circle cx="60" cy="60" r="48.6" fill="url(#vgoldr${u})"/>`
  g += `<circle cx="60" cy="60" r="48.6" fill="none" stroke="#3a2706" stroke-opacity=".45" stroke-width=".5"/>`

  // dark domed disc
  g += `<circle cx="60" cy="60" r="44.4" fill="#070402"/>`
  g += `<circle cx="60" cy="60" r="43.6" fill="url(#vdisc${u})"/>`

  // fine guilloche dot ring (high tiers ≥ VIP5)
  if (beads) {
    let gd = ''
    for (let k = 0; k < 60; k++) {
      const p = P2(41, k * 360 / 60)
      gd += `<circle cx="${p[0]}" cy="${p[1]}" r=".55" fill="${gTop}" opacity=".32"/>`
    }
    g += gd
  }

  // ring sheen: warm top highlight + cool bottom shadow arcs
  g += `<path d="M ${P2(56, 210).join(' ')} A 56 56 0 0 1 ${P2(56, 330).join(' ')}" fill="none" stroke="#fff6d6" stroke-opacity=".55" stroke-width="1.4" stroke-linecap="round"/>`
  g += `<path d="M ${P2(56, 30).join(' ')} A 56 56 0 0 1 ${P2(56, 150).join(' ')}" fill="none" stroke="#1c1206" stroke-opacity=".5" stroke-width="1.4" stroke-linecap="round"/>`

  // disc reflections
  g += `<circle cx="60" cy="60" r="43.6" fill="url(#vsheen${u})"/>`
  g += `<ellipse cx="60" cy="40" rx="27" ry="12" fill="#fff" opacity=".05"/>`
  g += `<path d="M ${P2(40, 42).join(' ')} A 40 40 0 0 1 ${P2(40, 138).join(' ')}" fill="none" stroke="${gMid}" stroke-opacity=".2" stroke-width="2"/>`

  // compass gem studs set into inner ring
  const stud = (cx: number, cy: number): string => {
    let s = `<g transform="translate(${cx},${cy}) rotate(45)"><rect x="-2.7" y="-2.7" width="5.4" height="5.4" rx="1.1" fill="url(#vstud${u})" stroke="#3a2706" stroke-width=".5"/><rect x="-2.7" y="-2.7" width="5.4" height="2.4" rx="1.1" fill="#fff" opacity=".22"/></g>`
    if (studRuby) s += `<circle cx="${cx}" cy="${cy}" r="1.35" fill="url(#vruby${u})"/>`
    return s
  }
  g += stud(60, 11.4) + stud(60, 108.6) + stud(11.4, 60) + stud(108.6, 60)

  // laurel wreath: organic two-lobe fronds
  const laurel = (side: number): string => {
    const R  = 37.5
    const a0 = side < 0 ? 152 : 28
    const a1 = side < 0 ? 216 : -36
    let s    = ''
    const pa = P2(R, a0)
    let stem = `M ${pa.join(' ')}`
    for (let sp = 1; sp <= 12; sp++) stem += ` L ${P2(R, a0 + (a1 - a0) * sp / 12).join(' ')}`
    s += `<path d="${stem}" fill="none" stroke="url(#vgold${u})" stroke-opacity=".85" stroke-width="1.5" stroke-linecap="round"/>`
    for (let k = 0; k < nLeaf; k++) {
      const f  = k / (nLeaf - 1)
      const ad = a0 + (a1 - a0) * f
      const p  = P2(R, ad)
      const rot = (ad + (side < 0 ? -180 : 0)).toFixed(1)
      const sc  = (0.66 + 0.42 * Math.sin(f * PI)).toFixed(2)
      s += `<g transform="translate(${p[0]},${p[1]}) rotate(${rot}) scale(${sc})">` +
           `<path d="M0 0 Q 4.2 -3 1.5 -9.2 Q -1.4 -3.6 0 0 Z" fill="url(#vleaf${u})" stroke="#3a2706" stroke-width=".4" stroke-linejoin="round"/>` +
           `<path d="M0 0 Q -3.4 -2.6 -1.1 -7.4 Q 1.3 -3 0 0 Z" fill="url(#vleaf${u})" stroke="#3a2706" stroke-width=".35" stroke-linejoin="round" opacity=".9"/>` +
           `<path d="M0 -.4 L .9 -7.6" fill="none" stroke="${gHi}" stroke-opacity=".5" stroke-width=".35"/></g>`
    }
    s += `<circle cx="${pa[0]}" cy="${pa[1]}" r="1.5" fill="url(#vstud${u})" stroke="#3a2706" stroke-width=".3"/>`
    return s
  }
  g += laurel(-1) + laurel(1)

  // crown shadow + body
  g += `<ellipse cx="60" cy="55" rx="19" ry="3.6" fill="#000" opacity=".22"/>`
  let cr = `<g transform="translate(60,42.5) scale(0.82)">`
  cr += `<rect x="-22" y="4" width="44" height="9" rx="3" fill="url(#vgold${u})" stroke="#3a2706" stroke-width=".7"/>`
  cr += `<path d="M -22 5 Q -21 -6 -20 -8 Q -16 -3 -14 0 Q -12 -10 -10 -13 Q -7 -6 -4.5 -2 Q -2 -15 0 -19 Q 2 -15 4.5 -2 Q 7 -6 10 -13 Q 12 -10 14 0 Q 16 -3 20 -8 Q 21 -6 22 5 Z" fill="url(#vgold${u})" stroke="#3a2706" stroke-width=".7" stroke-linejoin="round"/>`
  ;([[-20, -8], [-10, -13], [10, -13], [20, -8]] as [number, number][]).forEach(p => {
    cr += `<circle cx="${p[0]}" cy="${p[1]}" r="2.2" fill="url(#vstud${u})" stroke="#3a2706" stroke-width=".4"/>`
  })
  cr += `<circle cx="0" cy="-19" r="2.6" fill="url(#vstud${u})" stroke="#3a2706" stroke-width=".4"/>`
  cr += `<rect x="-1.1" y="-29" width="2.2" height="9" rx=".8" fill="url(#vgold${u})" stroke="#3a2706" stroke-width=".4"/>`
  cr += `<rect x="-4" y="-26.4" width="8" height="2.2" rx=".8" fill="url(#vgold${u})" stroke="#3a2706" stroke-width=".4"/>`
  let pr = ''
  for (let k = -4; k <= 4; k++) pr += `<circle cx="${(k * 4.6).toFixed(1)}" cy="8.5" r="1.05" fill="#fff3c0" stroke="#7a5512" stroke-width=".3" opacity=".9"/>`
  cr += pr
  if (crownGem) cr += `<circle cx="0" cy="8.5" r="2.3" fill="url(#vruby${u})" stroke="#3a2706" stroke-width=".4"/>`
  if (sideGems) cr += `<circle cx="-13" cy="8.5" r="1.8" fill="url(#vruby${u})"/><circle cx="13" cy="8.5" r="1.8" fill="url(#vruby${u})"/>`
  cr += `<path d="M -20 -7.5 Q -15 -11 -10 -12.5 Q -5 -16 0 -18.5 Q 5 -16 10 -12.5 Q 15 -11 20 -7.5" fill="none" stroke="#fff6d6" stroke-opacity=".5" stroke-width=".7" stroke-linejoin="round" stroke-linecap="round"/>`
  cr += `</g>`
  g += cr

  // center wordmark "PP.BET" + level number (embossed gold serif)
  const ff = `font-family='Georgia,"Times New Roman",serif' font-weight="700" text-anchor="middle" dominant-baseline="central"`
  const emb = (txt: string, x: number, y: number, sz: number, ls: string | number): string => {
    const a = `${ff} font-size="${sz}" letter-spacing="${ls}"`
    return `<text x="${x + 0.8}" y="${y + 1}" ${a} fill="#221703" opacity=".75">${txt}</text>` +
           `<text x="${x - 0.5}" y="${y - 0.6}" ${a} fill="#fff7d2" opacity=".5">${txt}</text>` +
           `<text x="${x}" y="${y}" ${a} paint-order="stroke" stroke="#3a2706" stroke-opacity=".55" stroke-width=".6" fill="url(#vtext${u})">${txt}</text>`
  }
  g += emb('PP.BET', 60, 66, 20, '0.5')
  g += `<line x1="42" y1="79" x2="78" y2="79" stroke="${gMid}" stroke-opacity=".7" stroke-width=".8"/>`
  g += `<path d="M 42 79 l -3.4 -2.1 v4.2 z" fill="${gTop}" opacity=".85"/>`
  g += `<path d="M 78 79 l 3.4 -2.1 v4.2 z" fill="${gTop}" opacity=".85"/>`
  g += `<circle cx="60" cy="79" r="1.1" fill="${gTop}"/>`
  g += emb(`VIP ${i}`, 60, 88.8, 10, '1')

  // living specular: soft reflection slowly travelling over the metal
  g += `<ellipse cx="60" cy="17" rx="17" ry="9" fill="url(#vspec${u})"><animateTransform attributeName="transform" type="rotate" from="0 60 60" to="360 60 60" dur="${royal ? 7 : 10}s" begin="${(-(u % 30) * 0.37).toFixed(2)}s" repeatCount="indefinite"/></ellipse>`

  const className = `${cls ?? ''}${royal ? ' vip-royal' : ''}`.trim()
  return `<svg class="${className}" viewBox="0 0 120 120" xmlns="http://www.w3.org/2000/svg" aria-hidden="true"><defs>${defs}</defs>${g}</svg>`
}

// Pre-generate all 11 level SVGs once at module load
export const VIP_SVGS: string[] = Array.from({ length: 11 }, (_, i) => vipBadgeSVG(i, 'vip-gem'))
