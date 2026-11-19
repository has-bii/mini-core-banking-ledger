# Still Bank — design system

The design system lives in [still-bank.pen](../still-bank.pen). It contains foundations, reusable masters including Sheet primitives, component specimens, authentication and device enrollment/unlock screens, `HOME-01`, `WAL-01`–`WAL-04`, `SAV-01`–`SAV-02`, `BUD-01`–`BUD-02`, ordinary/scheduled `TX-03`, `TX-05`, `TX-10`, and `ACT-03`. Supporting surfaces include wallet-type, rename, target-removal, budgeting-cancellation, wallet-closure, and enrollment-replacement sheets, plus illustrative platform passkey prompts. The screen inventory records exact frame references and coverage limits.

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
| Wallet / Bento card | Oy4Yu | WAL-01 wallet destination tile; connected Card / Surface and wallet preview, type icon, name, balance, and persistent type |
| Wallet / Create card | kJs8k | WAL-01 Create wallet action as the final bento tile; whole-card button opens OV-01 |
| Wallet / Balance summary | kPXhQ | WAL-01 total across all open wallets, wallet count, and 44 px balance-visibility toggle; visible and masked specimens |
| Activity / Home row | UmysQ | Transaction type, signed or neutral amount, affected wallet context, and date |
| Sheet / Content | h5PRBF | Edge-to-edge bottom surface with connected header/footer and a replaceable body slot |
| Sheet / Header | Zkuos | Wrapping title, optional description/leading icon, and connected 44 px close control |
| Sheet / Footer | DX3rD | Stacked connected 52 px primary/secondary actions; hidden for immediate choice lists |
| Sheet / Overlay | nhldJ | Full-screen modal scrim above the calling screen and below the sheet |
| Savings / Target summary | q5LfmJ | WAL-04 target progress, optional deadline, outgoing status, and View target; connected Card / Surface and button |
| Savings / Budgeting summary | irkQi | WAL-04 manual contribution schedule, per-contribution source guidance, next/finish dates, status, Move money, and View budgeting; connected Card / Surface and buttons |
| Wallet / Actions menu | y0c6gw | Shared 208 px Others dropdown, connected 44 px Rename/Close rows; hide Close for main wallet |
| Transaction / Wallet effect | migy6 | Wallet identity, signed debit/credit, and before/after balance; shared by projected reviews and posted transaction details |

Button variants and states are connected instances of one base. Cards replace the content slot and hide the footer when unused. Navigation examples override the nested item states.

## Usage

- Use one primary action per task area. Keep button width stable during submission and prevent repeated submission.
- Keep labels visible after typing. Pair errors with a message and icon; preserve editable values.
- Focus is distinct from selection. Buttons use a 2 px ring with a 4 px offset; inputs use a 2 px boundary; navigation has a separate focus example.
- Use IDR examples with Indonesian grouping. The long-amount specimen reduces type from 40 to 32 px at a 320 px component width.
- Bottom navigation belongs to primary destinations. Focused forms and security tasks use contextual back or close navigation.
- Bottom navigation's 20 px lower inset is a design specimen. Implementation must use the device's actual safe-area inset.
- Cards group content with a shared purpose. Hide unused slots and allow content to determine height.
- Sheets span the full screen width at x = 0 and meet the bottom edge, with no outer side or bottom inset. Use a full-screen vertical positioning frame with zero padding and `justifyContent: "end"`; Content uses `fill_container` width and content-driven height. Its top corners use `radius-2xl`, while bottom corners are square. Internal padding is 24 px top and 20 px on the sides/bottom, with a 20 px gap between header, body, and footer. The lower inset is a canvas specimen; implementation must use the device's actual safe area.
- Compose Sheet / Overlay and Sheet / Content over the calling screen. Replace Content's body slot, override the connected Header/Footer, and hide unused slots. The close control and Cancel dismiss without committing data; contain modal focus, make the background inert, and restore focus to the invoker. These interactions require implementation verification. OS-owned `OV-08` prompts retain their platform presentation rather than using Still Bank Sheet primitives.
- Home uses individual wallet cards in a bento arrangement and a separate recent-activity card, following the flat Card / Surface master. Section headings and View all actions stay outside the cards. Each sub or savings wallet shows its own name, balance, and type; absent wallets add no explanation or creation CTA.
- Wallets uses a two-column bento grid at 390 px. Main, sub, savings, and Create wallet each occupy one 169 × 169 px square, with a 12 px gap. An unpaired last tile keeps its column width. Main uses Primary, Create wallet uses Secondary, and both sub and savings wallets use Card (white). Each wallet keeps its name, balance or mask, and type. Long-content 320 px examples use 280 × 280 px full-width squares. Short-content sheet backgrounds use 134 × 134 px squares with 12 px padding, 14 px names, and 16 px amounts. The whole create card is one button, with a decorative plus glyph and the existing 2 px focus ring / 4 px offset convention.
- Wallets shows Total balance above the grid, with an Across N wallets caption and a 44 × 44 px eye control. The total includes all open wallets and is not labeled available to spend. Use Hide balances with an eye-off glyph while amounts are visible, and Show balances with an eye glyph while hidden. The toggle masks the total and individual wallet balances with a fixed-length `Rp ••••••` string; wallet names, types, counts, and navigation remain visible. Hidden numeric values must be absent from accessible labels. Persistence outside this view is unspecified.
- Wallet detail money-action rows show Top up, Move, Transfer, and Others with connected 44 px icon controls and centered 12 px labels. Others opens a right-aligned dropdown 8 px below the row. Rename opens OV-02; eligible sub/savings Close opens OV-06, while unfinished-target savings show the existing closure block. Main wallets expose Rename only. Sheet backdrops retain the same action row. Bottom Rename/Close buttons are removed.
- Savings wallets and budgeting schedules do not save a source wallet. Budgeting setup contains amount, period, and finish date. Scheduled Move money opens `TX-03` with the amount and savings destination prefilled; each fresh contribution requires an explicit source choice before `TX-05` review. Show source eligibility/funds feedback within that transaction, with a way to choose another source. Ordinary Add money remains independent of schedule advancement.
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
- [x] Money actions, card surfaces, status bars, bottom navigation, wallet previews, and activity rows use connected component instances. Top up, Move, Transfer, and Others use 44 × 44 px icon buttons with accessible names recorded in node context; compact section links retain 44 px height. All five Home variants (`O8Rgfi`, `I2fbM`, `D9JFa`, `k1Qib`, `M0eP4`) show centered 12 px labels beneath the icons with an 8 px gap, matching the populated Home's action row `EPgw4`. Labels fit on one line at both 390 px and 320 px screen widths; the icon targets remain 44 × 44 px. The Others destination remains undecided.
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

## Wallet list and type-sheet review

- [x] `WAL-01` has populated, first-visit/zero-balance/main-wallet-only, no-sub-wallet, no-savings-wallet, multiple-wallet/long-name/long-IDR at 320 px, closed-wallet return, and hidden-balance examples across seven screen frames. `OV-01` has 390 px and 320 px examples.
- [x] Exactly one main wallet appears in each state. Every wallet has its own square bento card with a name, balance or mask, type label, type icon, and detail indicator; multiple sub and savings wallets are supported. Target and budgeting details remain in savings detail. The first-visit explanation treats wallet creation as optional.
- [x] Wallet / Bento card and Wallet / Create card use square dimensions and tighter internal spacing. Their surfaces remain connected to Card / Surface, and destination cards retain a connected Wallet / Home preview. Wallet / Balance summary extends the reusable library with visible and hidden specimens. All 390 px wallet states use two columns; the 320 px long-content case uses full-width squares. Status bars, bottom navigation, close controls, and closure feedback reuse existing masters. The descriptive Security / Method option master supplies the type choices. Foundation values are unchanged.
- [x] The Create wallet card occupies the final grid position and leads to `OV-01`; each type choice leads to name-only `WAL-02` creation. There is no main-wallet creation choice and no required target or budgeting step. The whole card is one button; its plus glyph is decorative. A connected focus specimen in the bento library shows the 2 px ring and 4 px offset. Both type-sheet backgrounds use the revised bento screen, and dismissal restores focus to the create card.
- [x] The closed-wallet return example removes Holiday fund and credits its Rp12.000.000 to Daily spending, changing main from Rp2.450.000 to Rp14.450.000 while Bills remains Rp4.000.000. Its unfinished target was removed before closure; no budgeting schedule is present in this example.
- [x] The user's color mapping is applied to all 29 wallet/create card instances, including hidden balances and sheet backgrounds: main uses `--primary`, creation uses `--secondary`, and sub/savings use `--card`. The reusable create card and its focus specimen use the same Secondary treatment. Representative populated, masked, savings, creation, and focus views were inspected visually; a resolved traversal found no visible clipping.
- [x] Visible totals were recomputed from every wallet example: populated Rp18.450.000, first visit Rp0, no sub Rp14.450.000, no savings Rp6.450.000, long-IDR Rp137.037.785.790, and closed return Rp18.450.000. Counts match the open wallets. `CN67s` masks the total and all three wallet amounts without changing the layout.
- [x] All nine screen/overlay frames, the three wallet masters, balance specimens, and the focus specimen were inspected visually. The long total and narrow main card were also reviewed at a closer scale. Names wrap and complete amounts fit; the narrow long total stays on one line at 24 px. A resolved traversal found no visible clipping or overlapping siblings in the nine frames and bento library, and no overlapping document-root frames. All 29 wallet/create instances in the nine frames were checked for equal width and height: 169 px at 390 px, 280 px in the long-content narrow screen, and 134 px in the narrow sheet background. Eye and sheet-close targets remain 44 × 44 px. Work-in-progress placeholder flags are cleared.
- [x] Total, scope, eye-control, and masked-main contrast checks still apply. The six pairs affected by the wallet color mapping were recomputed and match the audit: main text 9.3:1, create title 7.9:1, create caption 4.89:1, white-card text 15.87:1, white-card type label 5.76:1, and white-card icons 9.3:1. Canvas variables, CSS, token JSON, and contrast-audit values are unchanged.
- [x] Product intent, MVP scope, inventory sections 2/3/4.3/5/6/7.1/7.4, this design system, the logo kit, existing Home and confirmation designs, the Neo wallet-detail reference, the user's Kantong bento-grid reference, and the user's square-card and balance-visibility requirements were consulted. No new logo artwork was placed.

Screen frames show all content on the canvas. Node context records navigation, balance-visibility, and modal intent; runtime toggling, scrolling, keyboard and assistive-technology behavior, modal focus containment, dismissal/focus restoration, safe areas, and live balance changes remain unverified. `WAL-02` and wallet-detail designs remain separate inventory work.

## Sheet primitives review

- [x] Library board `huy67` contains the four reusable Sheet primitives and connected composition specimens. The 320 px stress specimen `O3sTcD` shows a two-line title, description, complete `Rp137.037.785.790` amount, close control, and stacked actions; it adds no product flow. The 320 px `OV-01` screen separately demonstrates description wrapping.
- [x] `OV-01` frames `Dg4eL` (390 px) and `xgYTu` (320 px), plus `OV-07` frame `Q5EnzS`, use connected Sheet / Content and Sheet / Overlay instances. All three surfaces begin at x = 0, equal the screen width, and end at y = 844. Choice rows still use Security / Method option. The enrollment confirmation retains both consequence messages and Continue setup/Cancel, with a close action equivalent to Cancel.
- [x] The library and all three affected screens were inspected visually, including close views of the surfaces and the 320 px long-title/long-IDR specimen. Visible resolved traversal found no clipping or overlapping layout siblings; document-root frames do not overlap. Close controls measure 44 × 44 px and action buttons remain 52 px high. Work-in-progress placeholder flags are cleared.
- [x] All 89 canvas variables match the token JSON. CSS color values match the JSON, and all 29 contrast-audit pairs were recomputed successfully. Sheet text on the white surface measures 15.87:1 for primary text and 5.76:1 for supporting text; primary and secondary action pairs measure 9.3:1 and 7.9:1. Popover colors equal the existing Card pairs. Foundation values and token exports are unchanged.
- [x] Product intent, MVP scope, inventory decisions/entry rules/section 5 and flow maps 7.3–7.4, this design system, existing wallet-type/enrollment/platform surfaces, and the Neo wallet-detail reference were consulted. No logo artwork was placed or changed. Platform-owned `OV-08` examples remain outside this component migration.

Runtime modal focus, background inertness, dismissal/focus restoration, keyboard and screen-reader behavior, scrolling, and device safe areas remain unverified by the static canvas.

## Wallet creation, detail, and management review

- [x] Thirty-one top-level screen/overlay frames cover `WAL-02` (five), `WAL-03` (six), `WAL-04` (twelve), `OV-02` (four), and `OV-06` (four). Journey/state labels sit above the screen rows; reusable savings summaries sit above the screen area in library board `d8UxU9`. Each new frame has `clip: true`.
- [x] Creation is name-only for both supported types, with editing, required-name feedback, pending submission, and a 320 px example. Detail keeps wallet type visible after rename and shows an illustrative unique wallet number without prescribing its format. Main has no closure action. New-wallet/zero-balance and empty-activity states are represented.
- [x] Savings covers neither feature, target only, budgeting only, both, no target deadline, reached/unfinished/passed-deadline targets, target removal, Active/Overdue/Ended budgeting and guidance to choose a source when contributing. Source/funds recovery is handled in transaction entry. Reached targets keep budgeting independent; ended schedules retain their summary and hide scheduled Move money. Ordinary Add money and scheduled Move money have distinct documented destinations and effects. Locked examples keep incoming actions available and show incoming historical moves.
- [x] All 26 wallet detail states and rename/close sheet backdrops have four actions ending in Others, and no bottom management groups. The shared Wallet / Actions menu master `y0c6gw` is instantiated by `WriyT` (locked savings), `r0CHc` (main, Rename only), and `wVXYu` (eligible sub, 320 px). Its 208 px surface sits 8 px below the row; menu items measure 44 px. All action rows were checked structurally, representative complete screens/backdrops and all three open-menu examples were reviewed visually, and resolved checks found no visible clipping or root overlap. Existing contrast pairs apply and foundation values are unchanged. Click, keyboard navigation, dismissal, and focus restoration require implementation.
- [x] Rename and Close use connected Sheet / Content, Header, Footer, and Overlay primitives. All eight sheets were measured at full screen width, x = 0, ending at y = 844. Closure shows the destination/main balance and relevant schedule cancellation; unfinished-target closure instead explains the block and offers View target. The existing wallet-list closure result remains available.
- [x] The original 29 screen/overlay frames and both summary masters were inspected visually; subsequent Others changes and the two added menu states were reviewed as described above. Narrow creation, renamed detail, savings long IDR, rename, and closure examples were reviewed at 320 px. Complete `Rp123.456.789.000` balances fit; target amount lines wrap naturally. Input controls retain 56 px, full actions 52 px, and compact/icon controls 44 px, within floating-point layout precision.
- [x] Resolved traversal excluding disabled branches found no visible clipping or overlapping layout siblings. Document-root bounds found no overlapping frames. Work-in-progress placeholder flags are cleared. Mutation diagnostics reported sizing warnings on hidden/replaced template slot IDs; the resolved visible layouts and screenshots were checked separately.
- [x] Existing variables, colors, radii, and explicit control dimensions are reused. Foundation tokens and exports are unchanged. All 29 existing contrast pairs were recomputed and passed, and CSS color values match the token JSON. Relevant pairs include white on primary 9.30:1, body text on white 15.87:1, supporting text on white 5.76:1, destructive on white 5.63:1, warning text on its subtle surface 5.48:1, and information text on its subtle surface 6.04:1.
- [x] Product intent, MVP wallet/target/budgeting scope, inventory decisions and sections 4.3–4.4/5/6/7.4–7.5, this design system, supplied logo rules, existing Home/Wallets/Sheet designs, and `../references/walletdetail-neo.png` were consulted. No logo artwork was placed or changed.

The canvas records navigation and submission intent, rather than implementing interactions. Target/budgeting forms and details are covered by the subsequent review below; transaction entry/review/failure and transaction detail are covered below; success results and activity lists remain separate inventory work. Runtime scrolling, focus order, modal focus containment and restoration, screen-reader announcements, keyboards, safe areas, persistence, and simulated balance/schedule updates remain unverified. No name length or uniqueness rule was introduced.

## Savings target and budgeting review

- [x] Thirty-one new top-level frames cover `SAV-01` (six), `SAV-02` (four), `BUD-01` (six), `BUD-02` (three), `TX-03` scheduled contribution (four), `OV-03` (three), `OV-04` (three), and two `WAL-04` operation returns. Journey headers and row labels separate the work to the right of wallet management. New screen/overlay frames use `clip: true`; reviewed work-in-progress flags are cleared.
- [x] Target entry includes empty/editing, optional deadline, invalid amount/date, saving, and 320 px long-content states. Target details show unfinished, reached, no-deadline, and passed-deadline states. Removal explicitly preserves money and budgeting. The return example retains the active weekly schedule and enables outgoing Move/Transfer.
- [x] Budgeting entry includes amount, daily/weekly/monthly choices, finish date, validation, saving, and 320 px long-content coverage. No source is saved or selected here. Selection adds a checkmark to the existing connected Button / Base treatment. Budgeting detail covers Active, Overdue, and Ended. Ended hides scheduled Move money; other states open scheduled transaction entry, where the source is chosen for that move and source/funds problems can be corrected. Schedule cancellation preserves the unfinished target, deadline, outgoing lock, and balance.
- [x] Screens reuse connected status/navigation, text/amount inputs, buttons, notices, Card / Surface, target/budgeting summaries, and all four Sheet primitives. The connected Savings / Budgeting summary now explains per-contribution source choice and routes scheduled Move money to `TX-03`. Foundation tokens are unchanged. Dates use the existing labeled text input; the calendar glyph indicates date content and does not define a custom picker.
- [x] Every new screen/overlay and both journey headers were inspected in canvas screenshots. Exported close views of both long-IDR 320 px forms and the 320 px cancellation sheet were inspected separately. `Rp123.456.789.000` fits completely; long wallet names and explanatory copy wrap. The long target example's balance is exactly 80% of its target. The budgeting example has no configured source balance.
- [x] The initial resolved traversal excluding disabled branches found no visible clipping or overlapping flex-layout siblings in all 31 frames, and document-root bounds found no overlapping frames. All 12 sheet footer actions measure 52 px. Each of the six sheets starts at x = 0, matches its 390/320 px viewport, and ends at y = 844. Current source-selection revision checks are recorded below. Initial hidden-slot sizing warnings were corrected on the target instance; replacement operations also emitted diagnostics for removed template nodes, while subsequent visible-layout checks and screenshots passed.
- [x] All 29 existing contrast-audit pairs were recomputed and passed; all 49 CSS colors match the token JSON. Foundation colors and token exports are unchanged. Relevant existing pairs include primary/white 9.30:1, destructive/white 5.63:1, destructive/subtle 4.94:1, warning/subtle 5.48:1, and information/subtle 6.04:1. Disabled actions use the documented disabled treatment and show a visible explanation.
- [x] Product intent, MVP savings/budgeting scope, inventory decisions and sections 4.3–4.4/5/6/7.5, this design system, logo README/usage, existing wallet/form/sheet designs, and `../references/walletdetail-neo.png` were consulted. No logo artwork was added or changed.

Changing the budgeting period has no specified rule for recalculating the next contribution date. The edit examples show the current date without prescribing a recalculation; clarification was requested. Scheduled contribution review/pending and failure are covered in the transaction review below; the `TX-08` success result remains separate work. Runtime date input, source selection, keyboard/focus behavior, modal containment/restoration, screen-reader announcements, device safe areas, scrolling, persistence, and actual schedule/balance updates remain unverified by this static design.

### Source selection revision

- [x] Applied the user's decision that savings wallets and budgeting schedules have no saved source. Removed the source field from all six `BUD-01` forms. Updated the connected Savings / Budgeting summary master and its 18 instances, including wallet details, budgeting details, specimens, and cancellation backdrops, to explain per-contribution source choice. Scheduled Move money now enters `TX-03` before review.
- [x] Reused four former source-specific budgeting frames as `TX-03` scheduled-contribution variants: `C7K0i` chooses an initially unselected source, `JFcg8` shows an eligible selected source at 320 px, `bOmPI` handles insufficient funds in the selected wallet, and `mqwJl` handles a selected wallet becoming unavailable. These choices belong to the current move only; retries can correct the source without changing budgeting. Source options retain the MVP's open main/sub restriction. Ordinary Move entry is covered in the transaction review below.
- [x] Replaced the obsolete wallet-level source/funds warnings in `AsawL` and `EFFaT` with guidance to choose a source when moving money; `EFFaT` now supplies a 320 px wallet example. The scheduled contribution action remains available. Existing schedule amount, dates, target independence, and success-only advancement rules are unchanged.
- [x] Visually reviewed affected masters, forms, Active/Overdue/Ended summaries, wallet examples, return states, cancellation sheets, and all four transaction-entry variants. Exported close views of the overdue wallet, 320 px selected-source entry, and 320 px long-IDR budgeting form were inspected. A resolved audit of 38 affected/review frames found no visible clipping, overlapping layout siblings, or saved-source controls; document-root frames do not overlap. The ten text-input controls in this audit measure 56 px, and all 18 remaining period controls measure 44 px. Replaced-node diagnostics were checked separately from the current visible layout. Work-in-progress flags are cleared.
- [x] Updated MVP scope, inventory entries/counts, navigation rules, transaction retry behavior, traceability, and flow map 7.5 to match this decision. Foundation variables, token exports, and colors are unchanged; existing measured contrast pairs still apply. The product docs and existing design-system/pen.dev conventions were used; no brand artwork changed. Runtime source selection, focus, assistive technology, and transaction/schedule persistence remain unverified.


## Move entry, review, failure, and transaction-detail review

- [x] Thirty-one new top-level screen frames cover `TX-03` (twelve ordinary/selection/recovery states), `TX-05` (seven reviews/pending states), `TX-10` (five failures), and `ACT-03` (seven completed-detail variants). The four existing scheduled `TX-03` frames remain intact. Journey header `qG4sb` and six row labels organize the work to the right of scheduled entry. New screen frames have `clip: true`; all work-in-progress flags are cleared.
- [x] Ordinary entry includes empty/ready, savings Add money, inline source/destination choices, invalid amount, source/destination conflict, insufficient funds, no eligible source, newly locked source, eligible savings source, and 320 px long content. Scheduled choices remain transaction-only and restricted to open main/sub wallets. Target-locked savings can receive money; View target, source changes, amount correction where editable, and source top-up provide recovery paths.
- [x] Reviews show complete wallet identities and projected debit/credit and balance effects. Ordinary savings Add money leaves budgeting unchanged. Scheduled, early, and overdue examples advance exactly one weekly period only after success. The pending variant uses a loading action and records disabled repeated submission and conflicting Back navigation. Confirmation requires no extra transaction PIN/passkey screen.
- [x] Failure shows no movement, unchanged affected balances, retained schedule date where applicable, Try again, and Cancel. The current Rp400.000 source balance in scheduled failure reflects a change before confirmation, not a debit caused by the failed attempt. Retry returns to entry to revalidate/correct the draft. A failed attempt does not create posted activity.
- [x] Transaction details cover ordinary/scheduled moves, top-up, outgoing/incoming transfers, renamed/closed sub-wallet identity, and long names/amounts/references. They show completed status, participants, date/time, reference, and posted own-wallet effects. Other users' balances remain private. Back returns to the caller, including the originating success result or activity scope. Example wallet numbers, people, and references do not define formats or new capabilities.
- [x] Connected status bars, contextual navigation, inputs, buttons, notices, Card / Surface, and descriptive choice rows reuse existing masters. Library board `gVQWh` adds Transaction / Wallet effect (`migy6`) for review projections and historical detail. The master is above the screens. No foundation colors or tokens changed.
- [x] Every new screen, the shared master, and journey header were inspected in canvas screenshots. Narrow entry, review, failure, and detail show full `Rp123.456.789.000` amounts. Long wallet names remain visible; narrow selectors use the wallet type/number in the control and the full selected name below. Narrow amount typography uses existing 24/20 px scale sizes to prevent a final digit wrapping onto its own line.
- [x] Resolved traversal excluding disabled branches found no visible clipping or overlapping layout siblings across all 39 new screen/library/header/label frames. A document-root audit found no overlapping frames. All 23 visible labeled input controls in the new work measure 56 px; full action buttons measure 52 px. Existing selection/back targets remain connected 44 px controls. Final checks follow the last narrow-amount and pending-navigation refinements.
- [x] All 29 existing contrast-audit pairs were recomputed with zero failures or recorded-ratio discrepancies. New work reuses those semantic pairs; CSS, token JSON, and contrast-audit values are unchanged. Ordinary and long-content before/after examples preserve total balance; top-ups increase and outgoing transfers reduce it by the shown amount.
- [x] Product intent, MVP scope, inventory decisions and sections 3/4.5/4.6/6/7.5/7.7/7.8, this design system, supplied logo rules, existing scheduled-entry/wallet designs, and `../references/transfermove-fuse.png` were consulted. No brand artwork was added or changed.

`TX-08` Move success remains undesigned and is still the intended destination after a successful review. Top-up/transfer entry, their reviews/results, and activity lists remain separate work. Navigation, draft preservation, selection, pending exclusion, announcements, focus/keyboard behavior, scrolling, safe areas, and actual balance/schedule updates are documented intent and require runtime verification.
