<script setup lang="ts">
defineOptions({ name: 'PpmxLogo' })
</script>

<template>
  <span class="ppmx-logo" aria-label="PP.MX">
    <svg width="150" height="40" viewBox="0 0 240 64" role="img" aria-hidden="true">
      <use href="#ppEmblem" />
      <text class="logo-word" x="72" y="38" font-family="Montserrat,sans-serif" font-weight="800" font-size="26" letter-spacing="-0.5" fill="#FFFFFF">
        PP<tspan fill="#FFD700">.</tspan>MX
      </text>
      <text class="logo-sub" x="72" y="51" font-family="Montserrat,sans-serif" font-weight="600" font-size="6" letter-spacing="3" fill="#9a9aa3">
        CASINO · SPORTS · MÉXICO
      </text>
    </svg>
  </span>
</template>
