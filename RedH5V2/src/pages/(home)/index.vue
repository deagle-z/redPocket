<script setup lang="ts">
import PpmxHome from '../ppmx-home/PpmxHome.vue'

definePage({
  name: 'home',
  meta: {
    title: 'PP.BET',
    keepAlive: true,
    tabbar: false,
    shell: 'ppmx',
  },
})
</script>

<template>
  <PpmxHome />
</template>
