<script setup lang="ts">
defineOptions({ name: 'PpmxSymbols' })
</script>

<template>
  <svg
    width="0"
    height="0"
    class="ppmx-symbols"
    aria-hidden="true"
    focusable="false"
  >
    <defs>
      <symbol id="ppCoin" viewBox="0 0 100 100">
        <defs>
          <radialGradient id="ppcRim" cx="50%" cy="28%" r="74%">
            <stop offset="0" stop-color="#ffeaa6" />
            <stop offset=".5" stop-color="#e8a91e" />
            <stop offset="1" stop-color="#7e5206" />
          </radialGradient>
          <radialGradient id="ppcFace" cx="37%" cy="29%" r="80%">
            <stop offset="0" stop-color="#fffae6" />
            <stop offset=".4" stop-color="#ffd86b" />
            <stop offset=".74" stop-color="#f0bc3d" />
            <stop offset="1" stop-color="#bd880f" />
          </radialGradient>
          <linearGradient id="ppcClover" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stop-color="#69e893" />
            <stop offset="1" stop-color="#0c7043" />
          </linearGradient>
          <linearGradient id="ppcText" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stop-color="#fff6cf" />
            <stop offset="1" stop-color="#e09f17" />
          </linearGradient>
        </defs>
        <circle cx="50" cy="50" r="48" fill="url(#ppcRim)" />
        <circle
          cx="50"
          cy="50"
          r="43.6"
          fill="none"
          stroke="#7a4f06"
          stroke-width="3.1"
          stroke-dasharray="1.5 2.2"
          opacity=".5"
        />
        <circle cx="50" cy="50" r="39.5" fill="url(#ppcFace)" />
        <circle
          cx="50"
          cy="50"
          r="39.5"
          fill="none"
          stroke="#fff3c4"
          stroke-width="1.1"
          opacity=".75"
        />
        <circle
          cx="50"
          cy="50"
          r="35.4"
          fill="none"
          stroke="#9a6a06"
          stroke-width="1.2"
          opacity=".4"
        />
        <g>
          <path
            d="M50,49 C50.6,54 51.4,57 53.2,60.4"
            fill="none"
            stroke="#0c7043"
            stroke-width="2.4"
            stroke-linecap="round"
          />
          <circle cx="45.4" cy="38.6" r="6.7" fill="url(#ppcClover)" />
          <circle cx="54.6" cy="38.6" r="6.7" fill="url(#ppcClover)" />
          <circle cx="45.4" cy="47.6" r="6.7" fill="url(#ppcClover)" />
          <circle cx="54.6" cy="47.6" r="6.7" fill="url(#ppcClover)" />
          <circle cx="50" cy="43.1" r="3.1" fill="#0c7043" opacity=".55" />
          <ellipse cx="43.6" cy="36.4" rx="2.4" ry="1.5" fill="#ffffff" opacity=".4" />
        </g>
        <text
          x="50"
          y="76"
          text-anchor="middle"
          font-family="Montserrat, system-ui, -apple-system, Segoe UI, sans-serif"
          font-weight="800"
          font-size="17.5"
          letter-spacing="-.5"
          fill="url(#ppcText)"
          stroke="#7a4f06"
          stroke-width=".5"
        >PP</text>
        <ellipse cx="38" cy="27" rx="19" ry="11" fill="#fffefb" opacity=".38" />
      </symbol>

      <linearGradient id="lgRed" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#ff616b" />
        <stop offset="46%" stop-color="#e8243a" />
        <stop offset="100%" stop-color="#9c0b1c" />
      </linearGradient>
      <linearGradient id="lgGold" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stop-color="#fff3c4" />
        <stop offset="48%" stop-color="#ffd700" />
        <stop offset="100%" stop-color="#b8860b" />
      </linearGradient>
      <linearGradient id="lgGloss" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#ffffff" stop-opacity=".92" />
        <stop offset="100%" stop-color="#ffffff" stop-opacity="0" />
      </linearGradient>
      <linearGradient id="lgMono" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#ffffff" />
        <stop offset="55%" stop-color="#fff0f1" />
        <stop offset="100%" stop-color="#ffcfd3" />
      </linearGradient>
      <filter id="lgSh" x="-25%" y="-25%" width="150%" height="162%">
        <feDropShadow dx="0" dy="1.6" stdDeviation="1.7" flood-color="#6e0712" flood-opacity=".5" />
      </filter>
      <g id="ppEmblem" filter="url(#lgSh)">
        <rect x="4" y="6" width="52" height="52" rx="15" fill="url(#lgRed)" />
        <path
          d="M6 24 Q6 7 30 7 Q54 7 54 24 Q40 15 30 15 Q20 15 6 24 Z"
          fill="url(#lgGloss)"
          opacity=".5"
        />
        <path
          d="M11.5 50.5 Q30 57.5 48.5 50.5"
          stroke="url(#lgGold)"
          stroke-width="1.7"
          fill="none"
          stroke-linecap="round"
          opacity=".95"
        />
        <path
          fill-rule="evenodd"
          d="M20 17.5 H33.5 a10.2 10.2 0 0 1 0 20.4 H26 V46.5 H20 Z M26 23.7 H32.3 a4.3 4.3 0 0 1 0 8.6 H26 Z"
          fill="url(#lgMono)"
        />
        <rect
          x="4"
          y="6"
          width="52"
          height="52"
          rx="15"
          fill="none"
          stroke="url(#lgGold)"
          stroke-width="1.8"
        />
        <circle cx="47" cy="15" r="1.2" fill="#fff" opacity=".85" />
      </g>
    </defs>
  </svg>
</template>
