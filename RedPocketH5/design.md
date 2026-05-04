# Design: Mobile H5 Feature/Page Integration

## Background

This project is a mobile-first Vue 3 H5 SPA for BaseGoUni lucky packet flows, with Telegram Mini-App usage as a primary runtime. The current implementation already provides the core app shell, route guards, API wrapper, Pinia user state, root tab navigation, Vant-based UI patterns, and multilingual copy.

Relevant implementation is concentrated in:

- `src/pages/`: file-based pages consumed by `unplugin-vue-router`
- `src/api/user.ts`: centralized API types and request functions
- `src/router/index.ts`: route guards, public route policy, page tracking, source/channel capture
- `src/config/routes.ts`: root tab route names used by navbar/tabbar visibility
- `src/stores/modules/user.ts`: authenticated user state, token lifecycle, current TG user loading
- `src/utils/request.ts`: Axios instance, token injection, generic HTTP/business error handling
- `src/locales/*.json`: user-facing copy for `zh-CN`, `en-US`, `pt-BR`, `es-MX`, and `id-ID`
- `src/styles/themes/app-theme.css`: shared theme tokens used by page/component styles

The design goal is to document how a new feature page, or a minimal modification to an existing page, should be added without changing the existing architecture.

## Goals

- Keep new feature/page work aligned with the existing Vue 3 + TypeScript + Vite + Vant + Pinia + UnoCSS stack.
- Reuse file-based routing under `src/pages/<feature>/index.vue`.
- Extend `src/api/user.ts` for app API contracts instead of introducing a second API module.
- Use `useUserStore` only for shared authenticated user state; keep page-specific UI state local to the page/component.
- Use Vant components, existing shared components, UnoCSS utilities, scoped Less, and theme variables.
- Add all major user-facing text to every locale file under `src/locales`.
- Preserve existing route names, API contracts, token storage, and persisted user state.

## Non-Goals

- Do not introduce a second router, request client, state manager, CSS framework, or i18n mechanism.
- Do not manually edit generated typed-router files under `src/types/`.
- Do not move existing pages or broadly refactor page structure.
- Do not duplicate global request error handling in callers unless page-specific UX needs a custom message.
- Do not add root tab entries unless the feature is explicitly part of the main bottom navigation.
- Do not change hook behavior, commit tooling, or build tooling.

## Current Implementation

Routing:

- Routes are generated from `src/pages/` by `unplugin-vue-router`.
- Pages define route names with `defineOptions({ name: '...' })`.
- Public routes are defined in `src/router/index.ts` by route name and path. Current public route names are `Home`, `Login`, `Register`, `ForgotPassword`, and `luckyDetail`; current public paths include `/`, `/login`, `/register`, `/resetpwd`, `/resetpwd/`, and `/luckyDetail`.
- Unauthenticated users visiting protected routes are redirected to `Login` with `redirect=to.fullPath`.
- Authenticated users visiting `/login` or `/register` are redirected to `/`.
- Route guard side effects include route cache tracking, source channel capture, Facebook Pixel capture/init, page title updates, and page view tracking.

Root navigation:

- `src/config/routes.ts` currently lists root route names: `Home`, `History`, `Prize`, `Wallet`, `Profile`.
- `src/components/NavBar.vue`, `src/components/AppTopHeader.vue`, and `src/components/TabBar.vue` use this list to decide whether to hide the back arrow and show the tabbar.
- `TabBar.vue` currently links to `Home`, `History`, `Prize`, `/wallet`, and `Profile`.

Pages and components:

- Root/business pages include `index.vue` (`Home`), `history`, `packetList`, `luckyDetail`, `prize`, `wallet`, `profile`, `team`, `invite`, `recharge`, `withdraw`, `withdrawAccount`, `transform`, `rebateWithdraw`, `questions`, `setting`, auth pages, and simple activity pages.
- Shared components include `AppTopHeader`, `NavBar`, `TabBar`, `AppEmpty`, `AppConfirmDialog`, `LuckyGrabModal`, `ParityChoiceDialog`, `SendPacketForm`, `CurrencyText`, and `CoinAmount`.
- Existing pages use local `ref`, `reactive`, `computed`, `onMounted`, and Vant components such as `van-list`, `van-popup`, `van-picker`, `van-field`, `van-button`, `van-loading`, `van-skeleton`, `van-overlay`, `van-collapse`, and `van-icon`.

API:

- `src/api/user.ts` defines `ApiResult<T>`, feature-specific request/response interfaces, and request functions.
- Existing app endpoints use `/api/v1/app/...`, Telegram user endpoints use `/api/v1/app/tg/...`, and some admin history endpoints still live under `/api/v1/admin/...`.
- `src/utils/request.ts` sets `baseURL` from `VITE_APP_API_BASE_URL`, injects `Authorization` from local storage, times out after 6000 ms, shows generic Vant toasts for HTTP errors, and rejects business responses where `success === false` or `code` is not `0` or `200`.

State:

- Pinia is initialized with persisted state in `src/stores/index.ts`.
- `useUserStore` persists `userInfo`, handles phone login, Telegram login, registration, password reset, logout, and `loadCurrentUserInfo()`.
- The router guard calls `userStore.loadCurrentUserInfo()` when logged in and user info is incomplete.
- `useRouteCacheStore` stores route names with `meta.keepAlive`.

i18n:

- Locale files are JSON files in `src/locales/`.
- Current locale files are `zh-CN.json`, `en-US.json`, `pt-BR.json`, `es-MX.json`, and `id-ID.json`.
- Pages commonly call `const { t } = useI18n()` and read page-scoped keys such as `walletPage.*`, `historyPage.*`, `profilePage.*`, `sendPacketPage.*`, and shared keys under `common.*`.

Styling:

- UnoCSS is enabled and used alongside scoped Less.
- Shared theme tokens live in `src/styles/themes/app-theme.css`, including primary, gold, text, background, border, and shadow variables.
- Existing feature pages sometimes define scoped page styles, but new work should prefer existing variables and Vant CSS variables before adding raw colors.

## Proposed Design

For a new feature page, create `src/pages/<feature>/index.vue` and keep the page as the orchestration layer: read route/query params, call API functions from `src/api/user.ts`, own local loading/error/empty state, and render Vant-first mobile UI. Extract a child component only when the UI is reused or the page becomes hard to scan.

UI:

- Use the existing app shell behavior. Non-root pages should rely on the normal navbar/back behavior; root tab pages must be added to `rootRouteList` and `TabBar.vue` only when product requirements explicitly say they belong in the bottom navigation.
- Use Vant controls for mobile interactions: `van-list` for paginated lists, `van-popup` + `van-picker` for selectors, `van-field` for forms, `van-button` for primary actions, `van-loading` or `van-skeleton` for loading, and `AppEmpty` or the current page's empty-state pattern for no data.
- Use existing shared components where they fit: `CurrencyText`/`CoinAmount` for money display, `AppConfirmDialog` or Vant confirm dialogs for confirmations, `AppTopHeader` where the page needs current balance context.
- Keep touch targets and content density mobile-first. Avoid desktop-only layouts or large marketing sections.
- Use scoped Less for page-specific styling and theme tokens from `src/styles/themes/app-theme.css`.

Data flow:

1. User enters `/<feature>` through file-based routing.
2. `src/router/index.ts` starts NProgress, captures attribution/source data, sets the title, checks auth, and loads current user info if needed.
3. The page initializes local state in `setup`.
4. On `onMounted`, the page calls one or more API functions from `src/api/user.ts`.
5. `src/utils/request.ts` injects `Authorization`, normalizes business errors, and shows generic request toasts.
6. The page maps `ApiResult<T>.data` into local render state.
7. User actions call API functions, show page-specific success toasts where useful, then refresh local data or update a small local slice.
8. If user balance or identity changes, call `userStore.loadCurrentUserInfo()` after the successful mutation so shared header/profile state stays fresh.

State management:

- Keep filters, form inputs, pagination, selected tab, popup visibility, and submit/loading flags local to the page.
- Use `useUserStore` for authenticated user summary only.
- Add a new Pinia module only if multiple unrelated pages need the same mutable state and local state would cause real duplication.

API design:

- Add request/response interfaces next to related types in `src/api/user.ts`.
- Name functions by action, following existing style: `getXxx`, `createXxx`, `updateXxx`, `deleteXxx`, `setXxx`, `claimXxx`.
- Return typed `request.get<ApiResult<T>>()` or `request.post<ApiResult<T>>()`.
- Keep backend business payload shape compatible with existing `ApiResult<T>`.
- For paginated data, follow existing request shapes with `currentPage` and `pageSize`, and response shapes with `list`, `total`, `pageSize`, and `currentPage`.

Error handling:

- Let `src/utils/request.ts` handle generic HTTP and business errors.
- Add local validation before submitting forms and show localized Vant toasts for missing/invalid fields.
- Catch API errors only when the page must set local state, stop loading, or show a more specific localized fallback message.

Loading state:

- Use a page-level `pageLoading` or section-specific loading flag for initial requests.
- Use button `:loading` and `:disabled` for submit actions to prevent duplicate requests.
- For lists, use `van-list` with `loading`, `finished`, and localized `finished-text`.
- For important summary areas, prefer `van-skeleton` or `van-loading` patterns already used by `packetList`, `luckyDetail`, `wallet`, and `recharge`.

Empty state:

- Show a localized empty state when API returns an empty list or missing optional data.
- Prefer existing page-specific empty styling or `AppEmpty` for simple states.
- Empty text should live under the page's locale namespace, unless `common.noMore` or another shared key already fits.

Modification of existing pages:

| Target file | Current behavior | Target behavior | Minimal change plan | Verification |
| --- | --- | --- | --- | --- |
| `src/pages/<existing>/index.vue` | Page owns local API calls, render state, Vant UI, scoped Less, and page-specific toasts. | Add the requested feature while preserving the page's route name, auth behavior, and existing data flow. | Add local state/computed/actions near related code, import any new API function from `src/api/user.ts`, add localized copy, and reuse existing layout sections. | Run `pnpm typecheck`, `pnpm lint`, and manually verify the page in `pnpm dev`. |
| `src/api/user.ts` | Central API file contains app request types and functions. | Add only the missing endpoint types/functions required by the page. | Append interfaces and functions near the closest related feature block. Avoid creating a new API file unless the codebase convention changes. | Typecheck confirms callers and response types. |
| `src/locales/*.json` | Locale files hold page-scoped and shared copy. | Add matching keys to all locale files. | Add a page namespace such as `<feature>Page` or extend the existing namespace. Keep key names consistent across all languages. | Open page in each supported language or inspect missing key warnings. |
| `src/config/routes.ts` | Root tab list is `Home`, `History`, `Prize`, `Wallet`, `Profile`. | Only update if the feature becomes a root tab. | Add the route name to `rootRouteList` and update `TabBar.vue` item mapping if a visible tab is required. | Verify tabbar visibility, active state, and back-arrow behavior. |
| `src/router/index.ts` | Protected-by-default, with explicit public routes. | Only update if the new page must be public. | Add the page route name/path to the public route sets. | Verify anonymous access and protected-route redirect behavior. |

## File Changes

| File | Change type | Reason |
| --- | --- | --- |
| `src/pages/<feature>/index.vue` | Add | New feature page using file-based routing, Vant UI, local state, and scoped Less. |
| `src/api/user.ts` | Modify | Add typed request/response interfaces and API functions for the feature. |
| `src/locales/zh-CN.json` | Modify | Add Simplified Chinese page copy and toast/empty/loading labels. |
| `src/locales/en-US.json` | Modify | Add English page copy and keep key parity with `zh-CN`. |
| `src/locales/pt-BR.json` | Modify | Add Portuguese copy for all new keys. |
| `src/locales/es-MX.json` | Modify | Add Spanish copy for all new keys. |
| `src/locales/id-ID.json` | Modify | Add Indonesian copy for all new keys. |
| `src/config/routes.ts` | Optional modify | Add route name only if the feature is a root tab page. |
| `src/components/TabBar.vue` | Optional modify | Add a tab item only if the feature appears in bottom navigation. |
| `src/router/index.ts` | Optional modify | Add to public route sets only if anonymous access is required. |
| `mock/` | Optional add/modify | Add mock handlers only if local development cannot proceed against the backend. |

## i18n

Use a page-scoped namespace for new copy, for example `<feature>Page`. Keep all locale files structurally identical.

Recommended keys:

```json
{
  "<feature>Page": {
    "title": "",
    "loading": "",
    "emptyText": "",
    "submit": "",
    "submitting": "",
    "toastLoadFailed": "",
    "toastSubmitSuccess": "",
    "toastSubmitFailed": "",
    "requiredField": "{field} cannot be empty",
    "invalidField": "{field} format is invalid"
  }
}
```

If modifying an existing page, extend its existing namespace instead of creating a parallel namespace. For example, wallet-related copy belongs under `walletPage`, invite-related copy under `invitePage`, and packet list copy under `packetListPage` or `homeLucky` depending on the current page usage.

Shared labels should reuse `common.*` when the existing meaning fits, such as `common.noMore`, `common.cancel`, `common.confirm`, `common.requestFailed`, `common.networkError`, and `common.ok`.

## Compatibility

- Existing routes are unaffected unless `src/router/index.ts`, `src/config/routes.ts`, or `TabBar.vue` is explicitly changed.
- Existing API contracts are unaffected when new functions are appended to `src/api/user.ts`.
- Existing localStorage token behavior is unchanged because `src/utils/request.ts` continues to inject `Authorization` from the current storage key.
- Existing persisted Pinia user state is unchanged unless `UserState` or `useUserStore` is intentionally extended.
- Existing historical data is not migrated or transformed by frontend-only page additions.
- Existing Telegram Mini-App login/source channel behavior remains intact because the router guard continues to call `captureSourceChannelCode`, `captureFacebookPixelId`, and `initFacebookPixel`.
- Adding a public page changes anonymous access behavior and must be done explicitly in `src/router/index.ts`.
- Adding a root tab changes global navigation and must be reflected consistently in `rootRouteList` and `TabBar.vue`.

## Validation

Run the smallest relevant checks after implementation:

- `pnpm typecheck`
- `pnpm lint`
- `pnpm build:dev` when adding routes, root navigation, API contracts, or integration-sensitive UI

Manual validation:

- Start local dev server with `pnpm dev`.
- Visit the new or modified route on a mobile viewport.
- Verify anonymous/protected access behavior.
- Verify loading, success, error, empty, and pagination states.
- Verify submit buttons cannot double-submit while loading.
- Verify all new i18n keys render in `zh-CN`, `en-US`, `pt-BR`, `es-MX`, and `id-ID`.
- Verify Telegram Mini-App specific flows if the feature depends on Telegram user data or invite/source channel parameters.

## Risks

- Missing locale keys can leak raw key names into the UI. Mitigation: add matching keys to all locale files and scan page output in each supported language.
- Duplicating API clients or local error handling can create inconsistent auth/error UX. Mitigation: extend `src/api/user.ts` and rely on `src/utils/request.ts`.
- Adding a route as public by path but not by name, or the reverse, can create inconsistent guard behavior. Mitigation: update both public route sets when a page must be public.
- Adding a root tab without updating `TabBar.vue` can make navbar/tabbar state inconsistent. Mitigation: update `rootRouteList` and tab item rendering together.
- Updating shared user data only locally can leave header/profile balance stale. Mitigation: call `userStore.loadCurrentUserInfo()` after mutations that affect user profile, balance, country, avatar, username, phone, email, or rebate amount.
- Backend field names may differ between camelCase and snake_case, as existing lucky packet APIs already contain both patterns. Mitigation: type the actual backend response shape in `src/api/user.ts` and normalize only inside the page/component where needed.
- Raw colors in new scoped styles can drift from the app theme. Mitigation: use `src/styles/themes/app-theme.css` tokens and existing Vant CSS variables first.
