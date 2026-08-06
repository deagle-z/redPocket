import { describe, expect, it } from 'vitest'

import {
  createDynamicHomeCategories,
  mapAppHomeGame,
} from './usePpmxGameHome'

describe('usePpmxGameHome data mapping', () => {
  it('maps backend game items into PP.PE game tiles', () => {
    expect(
      mapAppHomeGame({
        gameId: 1001,
        gameName: 'Lucky Jaguar',
        categoryCode: 'slots',
        manufacturer: 'pgsoft',
        gameIcon: 'https://cdn.example.com/jaguar.png',
        sort: 1,
        showIndex: 1,
      }),
    ).toMatchObject({
      id: 'game-1001',
      gameId: 1001,
      categoryCode: 'slots',
      cover: 'https://cdn.example.com/jaguar.png',
      brand: 'PGSOFT',
      subtitle: 'PGSOFT',
      hot: true,
      name: {
        es: 'Lucky Jaguar',
        en: 'Lucky Jaguar',
        zh: 'Lucky Jaguar',
      },
    })
  })

  it('creates ordered dynamic categories from backend home data', () => {
    const categories = createDynamicHomeCategories({
      sports: [
        {
          gameId: 2,
          gameName: 'MLS',
          categoryCode: 'sports',
          manufacturer: 'sportsbook',
          gameIcon: '',
        },
      ],
      hot: [
        {
          gameId: 1,
          gameName: 'Top Game',
          categoryCode: 'hot',
          manufacturer: 'pp',
          gameIcon: '',
        },
      ],
    })

    expect(categories.map(category => category.id)).toEqual(['hot', 'sports'])
    expect(categories[0].name.es).toBe('Populares')
    expect(categories[1].name.en).toBe('Sports')
  })

  it('returns empty categories for empty backend data instead of static fallback data', () => {
    expect(createDynamicHomeCategories({})).toEqual([])
    expect(createDynamicHomeCategories({ hot: [] })).toEqual([])
  })
})
