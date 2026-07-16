import { describe, expect, it } from 'vitest'

import {
  ppmxCommissionSummary,
  ppmxTeamKpis,
  ppmxWheelRules,
} from './pageData'

const { readFileSync } = await import('node:' + 'fs')

const viewSources = [
  '../wheel/index.vue',
  '../team/index.vue',
  '../promo/index.vue',
  './components/PpmxPageChrome.vue',
  './PpmxHome.vue',
].map((path) => ({
  path,
  source: readFileSync(new URL(path, import.meta.url), 'utf8') as string,
}))

const hardcodedCopy: Array<string | RegExp> = [
  'Premio diario',
  'Ruleta de la suerte',
  'Girar la ruleta',
  'Reglas de la Ruleta de la Suerte',
  'No tienes giros disponibles',
  'Ganaste $',
  'Programa de afiliados',
  'Mi equipo',
  'Comisión disponible',
  'Retiro de comisión próximamente',
  'Enlace copiado',
  'Invita amigos y gana comisiones',
  'Bonos vigentes',
  'Promociones',
  'Reclamar',
  'Activación próximamente',
  'Soporte 24/7',
  'Juego responsable',
  />\s*Login\s*<\/button>/,
  />\s*Registro\s*<\/button>/,
  'Notificaciones',
  'Cargando juegos',
  'Sin juegos disponibles',
  'Casino y apuestas deportivas premium en México.',
]

describe('PP.MX page copy localization', () => {
  it('does not hardcode page copy in the PP.MX view templates', () => {
    const offenders = viewSources.flatMap(({ path, source }) =>
      hardcodedCopy
        .filter((copy) => typeof copy === 'string' ? source.includes(copy) : copy.test(source))
        .map((copy) => `${path}: ${String(copy)}`),
    )

    expect(offenders).toEqual([])
  })

  it('keeps reusable wheel and team business copy translated', () => {
    for (const copy of [
      ...ppmxWheelRules,
      ...ppmxTeamKpis.map((item) => item.label),
      ...ppmxCommissionSummary.map((item) => item.label),
    ]) {
      expect(copy.en).not.toBe(copy.es)
      expect(copy.zh).not.toBe(copy.es)
    }
  })
})
