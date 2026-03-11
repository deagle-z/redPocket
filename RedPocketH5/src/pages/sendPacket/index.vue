<script setup lang="ts">
import SendPacketForm from '@/components/SendPacketForm.vue'
</script>

<template>
  <div class="send-page">
    <SendPacketForm />
  </div>
</template>

<style scoped>
.send-page {
  min-height: 100vh;
  padding: 14px var(--page-padding-x) calc(26px + env(safe-area-inset-bottom));
  background-image:
    radial-gradient(circle at 18% 8%, rgba(212, 175, 55, 0.16), transparent 26%),
    repeating-linear-gradient(
      45deg,
      transparent,
      transparent 18px,
      rgba(212, 175, 55, 0.03) 18px,
      rgba(212, 175, 55, 0.03) 20px
    ),
    linear-gradient(180deg, #3e0000 0%, #230000 62%, #160000 100%);
}
</style>

<route lang="json5">
{
  name: 'SendPacket'
}
</route>
