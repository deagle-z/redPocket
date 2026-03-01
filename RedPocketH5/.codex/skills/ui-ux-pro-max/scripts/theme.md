Apply the RedPocketH5 unified theme system to the target file (or files described in $ARGUMENTS).

## Your job

Read the target Vue file(s). For every hardcoded color value (hex `#xxxxxx`, `rgb(…)`, `rgba(…)`) or repeated magic number (font-size, border-radius, spacing, height) in the `<style scoped>` block, replace it with the matching CSS variable from the table below.

Do **not** change logic, template markup, or script code — only the `<style scoped>` section.
Do **not** add new CSS rules. Only replace raw values with variables.
Do **not** use a variable if the hardcoded value has no close match; leave it as-is and add a `/* TODO: add to app-theme.css */` comment.

---

## Full variable reference — `src/styles/themes/app-theme.css`

### Primary colors
| Variable | Value | Use case |
|---|---|---|
| `--color-primary` | `#08ad5e` | Active tab underline, active state |
| `--color-primary-dark` | `#07bd63` | Feature icon background |
| `--color-primary-medium` | `#3d9b4f` | Withdraw / recharge accent |
| `--color-primary-btn` | `#4cae63` | Main action button background |
| `--color-primary-link` | `#49a86d` | Links, register button border |
| `--color-primary-soft` | `#e6f5ea` | Green soft background, active item |
| `--color-primary-active` | `#f2f8f4` | Language option selected bg |
| `--color-border-active` | `#5db97b` | Selected border |

### Functional colors
| Variable | Value | Use case |
|---|---|---|
| `--color-danger` | `#e84040` | Send-code button, destructive actions |
| `--color-telegram` | `#4da3e5` | Telegram login button |

### Text colors
| Variable | Value | Use case |
|---|---|---|
| `--color-text-title` | `#071a2b` | Page/header title |
| `--color-text-primary` | `#102236` | Popup title, first-level body |
| `--color-text-body` | `#0d1b2d` | Normal body text |
| `--color-text-form` | `#384f6b` | Form row label |
| `--color-text-input` | `#30455f` | Input field text |
| `--color-text-sub` | `#6b7280` | Secondary / sub-text |
| `--color-text-tab` | `#2d3f54` | Inactive tab label |
| `--color-text-feature` | `#24354a` | Feature card title |
| `--color-text-muted` | `#9ca8b8` | Placeholder, light description |
| `--color-text-light` | `#b7c1cc` | Close button, very light elements |
| `--color-text-en` | `#4b5e73` | English sub-label |

### Background colors
| Variable | Value | Use case |
|---|---|---|
| `--color-bg-page` | `#edf1f5` | Page background |
| `--color-bg-card` | `#ffffff` | Card / form / popup background |
| `--color-bg-feature` | `#eef4ee` | Feature card background |
| `--color-bg-gradient-a` | `#f7f9fc` | Gradient start (recharge/withdraw) |
| `--color-bg-gradient-b` | `#eef2f8` | Gradient end |

### Border & shadow
| Variable | Value | Use case |
|---|---|---|
| `--color-border` | `#edf1f5` | Row divider, inline border |
| `--color-stroke` | `#e4e9f0` | Card border, input border |
| `--shadow-card` | `0 12px 28px rgba(21,32,56,0.08)` | Card shadow |
| `--shadow-accent` | `0 6px 14px rgba(61,155,79,0.20)` | Accent shadow (light) |
| `--shadow-accent-md` | `0 8px 18px rgba(61,155,79,0.16)` | Accent shadow (medium) |

### Border radius
| Variable | Value |
|---|---|
| `--radius-sm` | `6px` |
| `--radius-md` | `10px` |
| `--radius-lg` | `12px` |
| `--radius-xl` | `14px` |
| `--radius-2xl` | `16px` |
| `--radius-3xl` | `18px` |
| `--radius-full` | `999px` |

### Spacing & sizing
| Variable | Value | Use case |
|---|---|---|
| `--page-padding-x` | `16px` | Horizontal page padding |
| `--header-height` | `72px` | AppPageHeader height |
| `--btn-height-md` | `52px` | Medium button height |
| `--btn-height-lg` | `62px` | Large button height |
| `--icon-size-sm` | `26px` | Small icon (26×26) |
| `--icon-size-md` | `34px` | Medium icon (34×34) |

### Font sizes
| Variable | Value |
|---|---|
| `--font-xs` | `13px` |
| `--font-sm` | `14px` |
| `--font-base` | `15px` |
| `--font-md` | `16px` |
| `--font-lg` | `18px` |
| `--font-xl` | `20px` |
| `--font-2xl` | `22px` |

---

## Steps to follow

1. Read the target file(s) specified in `$ARGUMENTS` (if empty, use the currently open file).
2. Scan every property value in `<style scoped>`.
3. Replace matching raw values with the CSS variable from the table.
4. Use the Edit tool to apply changes — one edit per file, replacing the entire `<style scoped>` block.
5. After editing, briefly list which variables were applied and flag any values that had no match.

Target: $ARGUMENTS
