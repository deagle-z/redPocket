import type { Ref } from 'vue'

export interface UseNoticeCarouselOptions {
  count: Ref<number> | number
}

export function useNoticeCarousel(options: UseNoticeCarouselOptions) {
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

  return {
    activeIndex,
    goTo,
    next,
  }
}
