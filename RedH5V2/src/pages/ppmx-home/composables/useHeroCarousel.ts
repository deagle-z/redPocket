import type { Ref } from 'vue'

export interface UseHeroCarouselOptions {
  count: Ref<number> | number
}

export function useHeroCarousel(options: UseHeroCarouselOptions) {
  const activeIndex = ref(0)
  const count = computed(() =>
    typeof options.count === 'number' ? options.count : options.count.value,
  )

  function goTo(index: number) {
    if (count.value <= 0) {
      activeIndex.value = 0
      return
    }

    activeIndex.value = (index + count.value) % count.value
  }

  function next() {
    goTo(activeIndex.value + 1)
  }

  function previous() {
    goTo(activeIndex.value - 1)
  }

  return {
    activeIndex,
    goTo,
    next,
    previous,
  }
}
