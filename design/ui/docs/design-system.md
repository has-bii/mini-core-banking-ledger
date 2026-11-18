# Still Bank — design system

The design system lives in [still-bank.pen](../still-bank.pen). It contains foundations and component specimens only, with no application screens.

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
| Radius | Base `0.625rem` = 10 px at a 16 px root |
| Derived radii | sm 6, md 8, lg 10, xl 14, 2xl 18, 3xl 22, 4xl 26 px |
| Pill radius | `radius-full`: 999 px |
| Control sizes | Buttons 52 px; compact and icon targets 44 px; inputs 56 px |
| Icons | Lucide, 20 px in controls and 24 px in bottom navigation |
| Elevation | Flat by default; borders and surface contrast provide grouping |

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

Button variants and states are connected instances of one base. Cards replace the content slot and hide the footer when unused. Navigation examples override the nested item states.

## Usage

- Use one primary action per task area. Keep button width stable during submission and prevent repeated submission.
- Keep labels visible after typing. Pair errors with a message and icon; preserve editable values.
- Focus is distinct from selection. Buttons use a 2 px ring with a 4 px offset; inputs use a 2 px boundary; navigation has a separate focus example.
- Use IDR examples with Indonesian grouping. The long-amount specimen reduces type from 40 to 32 px at a 320 px component width.
- Bottom navigation belongs to primary destinations. Focused forms and security tasks use contextual back or close navigation.
- Bottom navigation's 20 px lower inset is a design specimen. Implementation must use the device's actual safe-area inset.
- Cards group content with a shared purpose. Hide unused slots and allow content to determine height.
- Disabled colors deliberately differ from enabled text. They are excluded from the active-text contrast checks.

## Design review checklist

- [x] Only design-system boards and reusable masters appear at the document root.
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
- [x] No dark theme or application screens were created.

Contrast checks evaluate the specified pairs mathematically. The canvas shows visual states; keyboard interaction, loading animation, assistive-technology behavior, and device safe areas need verification when implementing the app.

The canvas editor dropped variable references supplied to width/height during creation. Control dimensions are therefore explicit pixel values matching the exported tokens. Other supported properties retain variable bindings. Update these explicit dimensions when changing their corresponding size tokens.
