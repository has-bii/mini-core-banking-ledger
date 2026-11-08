# MiniLedger design system

Light-only “Fintech Fresh” system for mobile operations staff and accountants.

The editable source is [design.pen](./design.pen). It contains seven library boards, **152 reusable components**, **146 tokens**, and isolated component specimens. No screen templates are included.

## Files

- [mini-ledger.tokens.json](./mini-ledger.tokens.json): token values, CSS names and canvas component IDs.
- [mini-ledger.css](./mini-ledger.css): light variables, Tailwind v4 theme mappings and component recipes.
- [exports/export.pdf](./exports/export.pdf): static reference of the seven boards.

## Foundation contract

| Topic | Decision |
| --- | --- |
| Platform | Mobile, 390 × 844 base; 360px minimum viewport |
| Content gutter | 16px; content width is 358px at base and 328px at minimum |
| Touch targets | At least 44 × 44px; badges and timeline markers are passive |
| Safe areas | Runtime device insets; canvas samples use 47px top and 34px bottom |
| Surfaces | Background #F6F7F9; white Cards; comfortable density |
| Radius | 12px controls and Cards; 8px small marks; 16px large surfaces; 999px badges |
| Font | Plus Jakarta Sans, weights 400–800 |
| Numerals | Tabular, lining numerals for amounts, balances, account numbers and IDs |
| Overlays | Full-width bottom Drawers |
| Data density | Transaction list items replace mobile data tables |

All component styling comes from tokens. Color token names match shadcn CSS variables: `primary` → `--primary` → `bg-primary`. Other canvas paths map to prefixed CSS names: `text/body` → `--ml-text-body`, `spacing/16` → `--ml-spacing-16`. The JSON records every mapping.

The canvas engine resolves dimension variables to numeric values. The design generator reads those values from the token collection. Reusable component nodes retain `metadata.dimensionTokens` so their source binding is explicit; the foundation reference board documents its values visually. Other supported style properties bind directly to canvas variables.

The .pen schema does not expose OpenType feature controls. Numeric text is marked with typography metadata; the CSS handoff applies `font-variant-numeric: tabular-nums lining-nums` and `font-feature-settings: "tnum" on, "lnum" on`. Confirm font loading in the real application.

## Typography

| Canvas style | Size | Line height | Weight | Purpose |
| --- | --- | --- | --- | --- |
| text/display | 40px | 1.2 | 700 | Library or prominent headings |
| text/h1 | 32px | 1.3 | 700 | Main task heading |
| text/h2 | 24px | 1.3 | 700 | Section heading |
| text/h3 | 20px | 1.3 | 600 | Subsection or Card heading |
| text/body | 16px | 1.5 | 400 | Content and input values |
| text/body-sm | 14px | 1.5 | 400 | Supporting content |
| text/caption | 12px | 1.4 | 500 | Metadata and helpers |
| text/amount-lg | 32px | 1.3 | 700 | Prominent balances and amounts |
| text/amount-md | 20px | 1.3 | 600 | Compact amounts |

Amounts have a separate currency line and are right-aligned. Account names, descriptions and references wrap. Financial amounts must remain complete and readable; move a long amount to its own row or use a predefined smaller amount style.

## Color and contrast

The requested palette is preserved exactly. Semantic tints and darker semantic foregrounds support readable small labels.

| Meaning | Base | Tint | Small-text foreground |
| --- | --- | --- | --- |
| Credit | #16A34A | #F0FDF4 | #166534 |
| Debit | #E11D48 | #FFF1F2 | #BE123C |
| Warning | #D97706 | #FFFBEB | #92400E |
| Information | #2563EB | #EFF6FF | #1D4ED8 |

`muted-foreground` (#64748B) has 4.76:1 contrast on white, but 4.44:1 on the specified app background. `subtle-foreground` (#5B6B80) is the stronger alternative for small text on background and muted surfaces. Base credit green is reserved for large bold amounts and icons; compact credit amounts use `credit-foreground`.

Measured examples: foreground/background 16.65:1; primary white text 6.29:1; destructive white text 4.70:1; semantic foreground/tint pairs 5.72:1 or higher. Color always has a second cue such as a sign, icon, label or status.

Border and input colors are structural surface tokens. Do not use them for text. Focus uses the ring token and an explicit outline.

## shadcn/ui mapping

Use the CSS-variable approach described in [shadcn theming](https://ui.shadcn.com/docs/theming). Merge the handoff CSS into the existing Tailwind v4 stylesheet and load the font through the application's font loader.

| Design component | React / shadcn composition | MiniLedger customization |
| --- | --- | --- |
| Button | Button | default, secondary, outline, ghost, destructive; sm 44px, default 48px, lg 56px, icon 48px |
| Input | Input | 16px value text, 48px target, persistent label |
| Textarea | Textarea | 112px initial minimum height; grow or scroll in the application |
| Label | Label | Visible text with programmatic association |
| Select | Select trigger styling + controlled mobile Drawer | Single-choice options, 48px rows, selected checkmark |
| Checkbox | Checkbox + Label | 24px mark inside a 72 × 44px target; supports mixed state |
| Switch | Switch + Label | 48 × 28px track inside a 72 × 44px target |
| Tabs | Tabs, TabsList, TabsTrigger, TabsContent | 44px trigger minimum; selected label and surface |
| Badge | Badge | default, secondary, outline, destructive; passive pill |
| Card | Card, CardHeader, CardContent, CardFooter | 12px radius, 16px padding, soft shadow |
| Avatar | Avatar, AvatarFallback | 44px and 56px specimens; nearby full name |
| Separator | Separator | One-pixel structural divider |
| Skeleton | Skeleton | Reserves transaction-row layout; one loading announcement |
| Alert | Alert, AlertTitle, AlertDescription | info, success, warning, error |
| Progress | Progress | Numeric label and task context; default, complete, error |
| Calendar | Calendar | Single-date selection, 44px minimum days |
| Date picker | Button + Drawer + Calendar | No centered popover on mobile |
| Drawer | DrawerHeader, DrawerTitle, DrawerDescription, content, DrawerFooter | Bottom origin, safe-area footer, focus trap and return |
| Sonner | Toaster + Sonner notifications | Success, info, persistent dismissible error |
| Form field | Label + control + description + message | React Hook Form FormField adapter or equivalent accessible composition |

The documented [Drawer anatomy](https://ui.shadcn.com/docs/components/base/drawer), [Calendar](https://ui.shadcn.com/docs/components/base/calendar) and [Sonner](https://ui.shadcn.com/docs/components/radix/sonner) supply the implementation primitives. Keep the MiniLedger visual contract if the installed shadcn primitive family differs.

For mobile Select, use a Button styled as the Select trigger with a controlled Drawer and the chosen library's supported single-selection semantics. Do not mount an independently opening Select popup on top of the Drawer.

## States

| State | Implementation contract |
| --- | --- |
| Default | Base token recipe |
| Pressed | `:active`; primary-pressed, secondary-pressed, accent or muted surface |
| Focus | `:focus-visible`; 2px ring with 2px offset |
| Disabled | Actual disabled prop; no events or focus; disabled surface/text |
| Error | `aria-invalid` for fields, persistent helper/error text, or an Alert for failed actions |
| Loading | `aria-busy`, stable width, prevent repeated submission |
| Selected | Selected surface, stronger text and exposed selection state |

All sizes inherit the same state recipe; the button board shows states at default size and every variant in every requested size. Input, Textarea, Select, Checkbox, Switch, Tabs, currency inputs, date controls, transaction rows and filter chips have state specimens.

A failed action is contextual feedback rather than a new Button variant. Preserve the intended action variant and pair it with an Alert. A disabled item and an unavailable data value are different conditions: unavailable numeric data is shown as “—”.

Passive components—Label, Badge, Avatar, Separator, Skeleton, Amount and AuditItem—use content or semantic variants. Pressed, focus and disabled states apply to their interactive wrappers or actions. Drawer states are open/closed plus child-control states. Alerts and Progress include error variants.

## Ledger components

| Component | Composition and behavior |
| --- | --- |
| Amount | Right-aligned value and ISO currency code. Explicit + credit / − debit; no sign for zero. Neutral unavailable value is —. A balance uses its actual sign. |
| CurrencyInput | Currency prefix + right-aligned Input. `type=text`, `inputMode=decimal`. Keep the sign outside the editable amount. |
| StatusBadge | Badge + icon + Pending / Posted / Reversed / Failed label |
| AccountCard | Name, masked account number, available balance and currency. Error state uses an unavailable balance plus Alert. |
| TransactionItem | Direction icon, wrapped title, exact date/time, status and signed amount. Whole row is interactive. |
| JournalLine | Account name/ID, explicit DEBIT · Dr or CREDIT · Cr, unsigned amount and currency |
| Difference | Debit total minus credit total, per currency. Balanced and unbalanced variants include icon, label and exact amount. |
| AuditItem | Event, actor and role, timestamp with time zone, immutable reference and connector |
| TopAppBar | Safe top inset, 44px back/action targets, wrapped task title |
| BottomTabBar | Four- and five-item variants, selected icon/label, safe bottom inset |
| EmptyState | Initial and filtered-empty cases with one useful action |
| FilterChip | 44px pill; selected state removes its filter when tapped |

Keep monetary values as exact decimal strings or integer minor units. Formatting follows the currency's precision (for example USD 2, JPY 0, KWD 3); do not calculate ledger differences using binary floating-point. Preserve incomplete typed decimals while editing and validate on blur or submit.

Journal posting remains disabled until the exact difference is zero. Reversals create an opposite entry and preserve the original record. Failed transactions show the attempted amount and failed status; they do not imply a posted balance change.

Keep the currency code visible and never total different currencies without an explicit conversion workflow. Full account numbers belong only in authorized details. Announce values with their currency; do not make assistive technology infer meaning from +/− glyphs or color.

Audit timestamps include a time zone. References wrap and can be copied with Sonner feedback. A toast confirms an action, while the persistent ledger status remains authoritative.

## Mobile overlay and keyboard behavior

- Add `viewport-fit=cover` to the viewport configuration.
- Read safe-area insets at runtime; the canvas's 47px/34px values are specimens.
- Drawer height is capped to the dynamic usable viewport. Its body scrolls while the action footer remains available.
- Keep the footer above the native keyboard; verify with the real viewport and focus behavior on iOS and Android.
- Use the Drawer primitive's focus trapping and focus return. Every close action has a 44px target.
- Selecting a mobile Select option commits the value, closes the Drawer and restores focus.
- Date selection is explicit and confirmed with Apply date. Expose selected and disabled days.
- Place Sonner below the top safe area. Errors persist until dismissed; success and informational messages announce politely.
- Respect reduced motion for Skeleton and overlay transitions.

## Verification checklist

- [x] Seven root library boards; no dark tokens or screen templates.
- [x] 152 reusable components; all 1,339 nodes have human-readable names.
- [x] Structural pass after layout settled: no clipping or collapsed frames.
- [x] All root placeholders cleared after section completion.
- [x] Style-property audit: no literal color, font, spacing, radius, opacity or stroke values in component nodes.
- [x] Numeric dimensions originate from tokens; reusable component dimensions retain binding metadata.
- [x] Reusable button/input/currency/date/choice/chip/day targets are at least 44 × 44px.
- [x] 360px component specimens: 328px content, wrapping account/transaction names and visible amounts.
- [x] Compact Calendar day width is 46.9px; five-item navigation width is 65.6px.
- [x] Contrast checked; stronger small-text token added for the app background.
- [x] Visual review of foundations, actions, controls, surfaces, transactions, navigation, Drawers, compact examples and toast.
- [x] Exact 75% progress label agrees with the rendered fill.
- [x] CSS supplies the OpenType numeric feature absent from the canvas schema.

These checks validate the design and handoff files. React runtime behavior, actual font delivery, native keyboards and assistive-technology behavior must be verified when implementing the app; no application or automated interaction suite is included.
