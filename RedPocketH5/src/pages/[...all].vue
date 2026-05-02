<script setup lang="ts">
import { safeBack } from '@/utils/navigation'

const router = useRouter()

function onBack() {
  safeBack(router)
}
</script>

<template>
  <div text="center gray-300 dark:gray-200">
    <van-icon name="warn-o" size="3em" />
    <div> Not found </div>

    <div class="mt-2">
      <button van-haptics-feedback class="btn" @click="onBack">
        Back
      </button>
    </div>
  </div>
</template>

<route lang="json5">
{
  name: '404'
}
</route>
