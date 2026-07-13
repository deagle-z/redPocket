import type { Ref } from 'vue'
import { onBeforeUnmount, onMounted, toValue, watch } from 'vue'

export interface VisibleCategoryEntry {
  id: string
  ratio: number
}

export function pickMostVisibleCategory(
  entries: VisibleCategoryEntry[],
  currentCategoryId: string,
) {
  if (entries.length === 0) return currentCategoryId

  return entries.reduce((best, entry) => (entry.ratio > best.ratio ? entry : best)).id
}

export function useCategoryScrollSpy(
  categoryIds: string[] | Ref<string[]>,
  activeCategoryId: Ref<string>,
) {
  let observer: IntersectionObserver | undefined

  function observeCategories() {
    if (!observer) return

    observer.disconnect()

    for (const categoryId of toValue(categoryIds)) {
      const section = document.getElementById(`ppmx-cat-${categoryId}`)
      if (section) observer.observe(section)
    }
  }

  onMounted(() => {
    if (typeof IntersectionObserver === 'undefined') return

    observer = new IntersectionObserver(
      (entries) => {
        const visibleEntries = entries
          .filter((entry) => entry.isIntersecting)
          .map((entry) => ({
            id: String((entry.target as HTMLElement).dataset.ppmxCategoryId || entry.target.id.replace(/^ppmx-cat-/, '')),
            ratio: entry.intersectionRatio,
          }))

        activeCategoryId.value = pickMostVisibleCategory(visibleEntries, activeCategoryId.value)
      },
      {
        rootMargin: '-140px 0px -55% 0px',
        threshold: [0.08, 0.16, 0.32, 0.48, 0.64, 0.8],
      },
    )

    observeCategories()
  })

  watch(
    () => toValue(categoryIds).join('|'),
    () => {
      window.requestAnimationFrame(observeCategories)
    },
  )

  onBeforeUnmount(() => {
    observer?.disconnect()
  })
}
