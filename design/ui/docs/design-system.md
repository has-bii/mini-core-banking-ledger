# Still Bank — design system

The design system lives in [still-bank.pen](../still-bank.pen). It contains foundations, reusable masters, component specimens, the `AUTH-01`–`AUTH-06` authentication screens, device enrollment/unlock screens, and the `HOME-01` overview with supporting states.

## Token contract

- [Tailwind v4 / shadcn light-theme CSS](../tokens/still-bank.css)
- [Canvas variable export](../tokens/still-bank.tokens.json)
- [Measured contrast pairs](../tokens/contrast-audit.json)
- 89 variables; 49 semantic and supporting colors.
- Token names follow the user's supplied shadcn contract. Forest `#214F40` and Mist `#F5F6F3` adapt the values to the approved identity.
- All chart and sidebar token names are included. Sidebar tokens provide compatibility; no desktop sidebar component or screen is included.
- Small extensions cover success, warning, information, subdued state fills, interaction colors, disabled controls, and pill radius.
- This system has one light theme. The CSS contains no dark variant or dark theme block.

## Foundations

| Foundation | Definition |
| --- | --- |
| Typeface | Figtree: regular 400, medium 500, semibold 600 |
| Token annotations | IBM Plex Mono; documentation only |
| Type size / line height | 12/16, 14/20, 16/24, 20/28, 24/32, 32/40, 40/48 px |
| Spacing | 4 px base; 4, 8, 12, 16, 20, 24, 32, 40, 48, 64 px; 2 px optical adjustment |
| Radius | Base `0.875rem` = 14 px at a 16 px root |
| Derived radii | sm 8.4, md 11.2, lg 14, xl 19.6, 2xl 25.2, 3xl 30.8, 4xl 36.4 px |
| Pill radius | `radius-full`: 999 px; shared by buttons and bordered form inputs |
| Control sizes | Buttons 52 px; compact and icon targets 44 px; inputs 56 px |
| Icons | Lucide, 20 px in controls and 24 px in bottom navigation |
| Elevation | Flat; cards use Card / Surface's border treatment without shadows |

Figtree's licensed source and existing logo assets are in `../../logo/`. Load the font when implementing the interface. Use supplied logo artwork intact, with its original proportions and clear space.

## Reusable masters

| Component | Canvas ID | Use |
| --- | --- | --- |
| Button / Base | CoJBf | Primary, secondary, outline, ghost, destructive; optional icons and loading indicator |
| Button / Icon | kPjGo | 44 px target with a 20 px glyph |
| Input / Text | k8hMD4 | Persistent label, control, optional prefix/action, helper or error feedback |
| Input / Amount | xn6w8 | IDR prefix, prominent value, supporting balance context |
| Card / Surface | etfbu | Optional header, content, and footer slots |
| Navigation / Item | EiRnc | Labeled item with selected indicator |
| Navigation / Bottom bar | x6SHJq | Home, Wallets, Activity, Settings |
| Navigation / Top bar | SRwVz | Back or close action and contextual title |
| System / Status bar | QT5AI | Shared mobile status-bar specimen; device chrome is illustrative |
| Feedback / Inline notice | qPycb | Success, error, and information feedback with an icon, title, and description |
| Security / Method option | ZHfba | Account PIN and passkey choices with an icon, title, description, and navigation indicator |
| Wallet / Home preview | y0Zpji | Persistent wallet category, editable name, balance, and optional savings target progress and outgoing restriction |
| Activity / Home row | UmysQ | Transaction type, signed or neutral amount, affected wallet context, and date |

Button variants and states are connected instances of one base. Cards replace the content slot and hide the footer when unused. Navigation examples override the nested item states.

## Usage

- Use one primary action per task area. Keep button width stable during submission and prevent repeated submission.
- Keep labels visible after typing. Pair errors with a message and icon; preserve editable values.
- Focus is distinct from selection. Buttons use a 2 px ring with a 4 px offset; inputs use a 2 px boundary; navigation has a separate focus example.
- Use IDR examples with Indonesian grouping. The long-amount specimen reduces type from 40 to 32 px at a 320 px component width.
- Bottom navigation belongs to primary destinations. Focused forms and security tasks use contextual back or close navigation.
- Bottom navigation's 20 px lower inset is a design specimen. Implementation must use the device's actual safe-area inset.
- Cards group content with a shared purpose. Hide unused slots and allow content to determine height.
- Home uses individual wallet cards in a bento arrangement and a separate recent-activity card, following the flat Card / Surface master. Section headings and View all actions stay outside the cards. Each sub or savings wallet shows its own name, balance, and type; absent wallets add no explanation or creation CTA.
- Disabled colors deliberately differ from enabled text. They are excluded from the active-text contrast checks.
- Authentication uses contextual navigation and connected input/button instances. Registration verification and password recovery return to Sign in with a confirmation notice; neither signs the user in automatically.
- OTP entry uses one text field. Code examples do not prescribe length, format, expiry, resend timing, or password complexity. Those product details remain open.

## Design review checklist

- [x] Design-system boards, reusable masters, and individual authentication screen frames appear at the document root.
- [x] Master components are arranged above the foundation and specimen boards.
- [x] No visible node has a clipping problem in the resolved layout traversal.
- [x] Root boards do not overlap; all work-in-progress placeholder flags are cleared.
- [x] Text has visible fills, and parent layouts control wrapping and spacing.
- [x] Component variants use connected instances rather than independent redraws.
- [x] Actual button, icon-target, and input dimensions match 52 / 44 / 56 px specifications.
- [x] Text color pairs meet the chosen 4.5:1 target; input boundary and focus pairs meet 3:1.
- [x] Selected navigation uses an indicator and label weight as well as color.
- [x] Error feedback combines text and an icon.
- [x] A 320 px specimen accommodates a wrapped label and a long IDR amount.
- [x] Long card headings grow their surface naturally.
- [x] Screenshots were reviewed per foundation section and component family.
- [x] CSS and JSON share the canvas color values and the requested radius multipliers.
- [x] Authentication screens and states use the existing light theme.

## Authentication review

- [x] Six inventory screens, validation/retry states, pending states, resend feedback, and Sign-in confirmations were reviewed on the canvas.
- [x] Shared OTP specimens cover focused editing and missing, incomplete, and invalid code feedback; full-screen examples include expired codes.
- [x] A resolved traversal found no visible clipping or overlapping root frames. Input controls retain 56 px height; primary buttons retain 52 px and compact/back controls 44 px.
- [x] 320 px examples were visually inspected for password visibility, wrapped text, and a long email address.
- [x] Error and notice text uses existing semantic colors. Authentication link/error contrast pairs were added to the contrast audit; no token values changed.
- [x] Product intent, MVP scope, inventory sections 4.1/6/7.2, the logo kit, and the Wise/Phantom authentication references were consulted. The Coinbase onboarding reference was also inspected.

Long IDR amounts are not applicable to authentication. Canvas checks do not verify runtime focus order, keyboard behavior, screen-reader announcements, animations, safe areas, or real authentication delivery.

## Device enrollment and unlock review

- [x] `DEV-01`–`DEV-05`, `SEC-02` in enrollment/reset contexts, `OV-07`, and illustrative `OV-08` creation/unlock surfaces are on the canvas. PIN/passkey management and logout remain in section 4.7 scope.
- [x] PIN-only, passkey-only, and combined unlock variants, incorrect PIN, setup/unlock cancellation and failure, unavailable passkey creation, pending operations, and email-sign-in fallback are represented.
- [x] PIN mismatch, focused editing, recovery delivery retry, invalid/expired OTP, resend pending/confirmation, missing/incomplete OTP specimens, and PIN-reset completion were inspected visually.
- [x] A resolved traversal found no visible clipping or overlapping root frames. Connected button/input instances retain 52 / 44 / 56 px controls. The method option extends the existing library; foundation token values are unchanged.
- [x] 320 px PIN editing and long-email OTP layouts were visually reviewed. Existing semantic colors are reused; additional measured contrast pairs cover the method choices, section header, and illustrative platform surfaces.
- [x] Replacement confirmation explains that prior enrollment persists until setup succeeds. PIN reset returns to Quick unlock without opening Home. Passkey prompt examples identify the OS/browser-owned handoff in their canvas context and review notes.
- [x] Product intent, MVP scope, inventory sections 4.2/4.7/5/6/7.3, the design system, logo kit, and Wise/Phantom/Coinbase references were consulted.

PIN length/format and OTP timing remain unspecified. Long IDR amounts are not applicable to these screens. Platform prompts are illustrative, not a verified replica of a specific provider. Runtime focus, keyboards, assistive technology, safe areas, email delivery, and real passkey behavior require implementation checks.

## Home overview review

- [x] `HOME-01` has populated, first-visit/zero-balance, main-wallet-only, no-sub-wallet, no-savings-wallet, empty-activity, and 320 px long-name/long-IDR coverage across five screen frames.
- [x] Total balances equal the main, sub, and savings balances in each example. The main wallet is identified as everyday money; total is labeled across wallets, without an available-to-spend claim.
- [x] Wallet previews retain their type after renaming. Each sub or savings wallet shows its name above the balance and its type below it, supporting multiple wallets of either type. Home omits savings target amounts, progress, deadlines, and outgoing-lock labels; those details belong in the savings wallet.
- [x] Money actions, card surfaces, status bars, bottom navigation, wallet previews, and activity rows use connected component instances. Top up, Move, Transfer, and Others use 44 × 44 px icon buttons with accessible names recorded in node context; compact section links retain 44 px height. The populated Home (`O8Rgfi`, action row `EPgw4`) shows centered 12 px labels beneath the icons with an 8 px gap; the icon targets remain 44 × 44 px. The Others destination remains undecided.
- [x] The main wallet spans the available width; sub and savings wallets each have their own flat Card / Surface instance in the paired row. Missing wallets add no summary or creation CTA, and the 320 px variant stacks wallet cards to preserve long names and amounts.
- [x] Recent activity rows and the empty activity state are inside a flat Card / Surface instance. Recent activity and View all remain siblings of that card, outside its bounds.
- [x] An own-wallet move appears once in recent activity with a neutral amount and both wallet names; transfers and top-ups use debit and credit signs. Empty activity explains where future transactions appear.
- [x] Revised wallet and activity sections were inspected visually across all five variants. Resolved visible-node traversal found no clipping or overlapping siblings; affected root frames do not overlap. Long names wrap and the narrow total uses 32 px type. All work-in-progress placeholder flags are cleared.
- [x] Existing colors and foundation tokens are reused. The four text/icon contrast pairs used by the revised cards and actions were recomputed and match the contrast audit. CSS and token-export values are unchanged; the existing progress pair remains applicable to the reusable preview's optional progress treatment.
- [x] Product intent, MVP scope, inventory sections 2/3/4.3/6/7.1/7.8, the design system, supplied logo kit, and Wise/Neo/Chime Home references were consulted.

Home uses the existing bottom-navigation convention. Screen frames expand to show all content on the canvas. Navigation targets are documented in node context; this static design does not verify runtime scrolling, focus order, screen-reader announcements, touch behavior, device safe areas, or live balance updates.

Home cards use connected Card / Surface instances with its 2xl radius and 1 px border, without shadows. The main wallet uses the primary fill and matching border, spanning both columns. At 390 px, other wallets occupy one 169 px column and size to their content. Narrow cards stack and size to their content. Wallet types appear below individual balances, without category subsections or missing-wallet guidance. Foundation colors, CSS tokens, and JSON token values are unchanged; existing contrast pairs still apply.

Contrast checks evaluate the specified pairs mathematically. The canvas shows visual states; keyboard interaction, loading animation, assistive-technology behavior, and device safe areas need verification when implementing the app.

The canvas editor dropped variable references supplied to width/height during creation. Control dimensions are therefore explicit pixel values matching the exported tokens. Other supported properties retain variable bindings. Update these explicit dimensions when changing their corresponding size tokens.
