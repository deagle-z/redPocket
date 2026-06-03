# RedPocketH5 UX Optimization Notes

## Summary

RedPocketH5 is a mobile-first H5 and Telegram Mini-App experience. The main UX priorities are stable mobile scrolling, consistent red/gold visual language, predictable loading and failure states, and clear feedback for game and lucky packet flows.

## Layout Rules

- Use one intentional content scroll area per page whenever possible.
- Keep app chrome fixed: page header, search controls, bottom tabbar, and floating pagination should not depend on page body scroll.
- Reserve bottom space with `--bottom-safe-space` for floating controls.
- For split game pages, keep provider navigation and game cards inside one bounded content panel; each side may scroll independently inside that panel.

## Visual Rules

- Use the shared dark game tokens in `src/styles/themes/app-theme.css`.
- Keep the project red/gold theme across Home, Game List, Packet List, and Game Play pages.
- Prefer shared card radius, borders, and shadows instead of page-specific one-off values.
- Keep game and packet entry cards dense enough for mobile scanning.

## State Rules

- Use `AppSkeletonSection` for first-load card or row skeletons.
- Use `AppEmpty` for empty data.
- Use `AppRetryState` for retryable request or iframe launch failures.
- Do not duplicate generic request error toast handling already provided by `src/utils/request.ts`.

## Game UX

- Home category `Todo` actions should route to `/gameList?categoryCode=...`; Lottery remains `/prize`.
- Game cards should prefer `gameIcon`, then fall back to `horizontalImage`.
- `/gameList` keeps the header and search visible while provider and game lists scroll inside the content panel.
- `/gamePlay` shows launch loading, iframe loading, timeout failure, and retry.

## Verification

- Run `pnpm typecheck`.
- Run targeted eslint for touched Vue/CSS files.
- Manually verify 360px, 390px, and 430px mobile widths for Home, Game List, Packet List, and Game Play.
