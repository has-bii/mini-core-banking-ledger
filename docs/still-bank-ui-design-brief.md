# Still Bank — Design System and Authentication Only

## 1. Task and strict scope

Start from scratch.

Create only:

1. A reusable **design system** for Still Bank.
2. **Authentication screens** demonstrating that system.

**Do not design the rest of the banking application.** Do not generate Home, dashboards, Accounts, Savings, Transfers, Top-up, Activity, transaction receipts, Profile, or bottom-navigation screens—not even as examples or previews. Banking context does not authorize banking screens.

This brief is for a UI design agent such as pen.dev or Wonder Design. Deliver designs, component specifications, and token specifications only. Do not implement frontend code.

## 2. Audience and platform

- Still Bank is a personal banking app for customers, not internal staff or business users.
- Mobile only, with a native-mobile-app appearance—not a mobile website.
- Light mode only; no desktop layouts or dark-mode designs.
- Choose one reference phone size and document adaptation to smaller and larger phones.
- Respect safe areas and mobile keyboards.
- Future financial content uses IDR, formatted like **Rp150.000**, without decimals. This is context for reusable typography, not a request for transaction screens.

## 3. Visual direction

**Quiet Confidence:** calm, precise, trustworthy, and approachable without feeling cold, flashy, or exclusive.

- Use the existing **Still Bank** identity and approved logo assets in repository-root [`design/logo/`](../design/logo/); do not design a new logo.
- **Main canvas must be off-white**, not pure white: `--background: oklch(0.98 0.006 85)`.
- Use this shared background throughout auth screens, including exposed safe areas.
- Cards and popovers stay white: `--card: oklch(1 0 0)` and `--popover: oklch(1 0 0)`.
- Deep forest-green primary actions, charcoal text, subtle borders.
- Readable sans-serif typography, clear hierarchy, and tabular numerals for financial amounts.
- Generous but practical spacing, restrained rounding, minimal decoration.
- Reserve red for errors and destructive actions.
- Clear, reassuring language; no unnecessary financial jargon.
- Avoid excessive gradients, decorative banking dashboards, and generic website hero sections.

Preserve the specified canvas, card, and popover values. Choose and document the remaining values within this direction.

### Existing logo and brand assets

The approved logo is **Open Ground (Direction B)**. Start with the [logo kit README](../design/logo/README.md) and follow the [usage rules](../design/logo/USAGE.md) and [visual guide](../design/logo/guidelines/still-bank-usage-guide.pdf).

- Default auth-screen lockup: [`still-bank-horizontal-color.svg`](../design/logo/svg/still-bank-horizontal-color.svg), at least **160 px canvas width**.
- For symbol-only placements, use [`still-bank-symbol-color.svg`](../design/logo/svg/still-bank-symbol-color.svg) at **32 px and above**, or [`still-bank-symbol-small-color.svg`](../design/logo/svg/still-bank-symbol-small-color.svg) at **16–24 px**.
- Use supplied SVG lockups intact. Do not retype the wordmark, redraw, stretch, recolour, add effects, or change symbol/name proportions.
- Keep at least **one symbol stem-width** clear around the visible artwork; SVG canvas padding does not replace layout clear space. Follow the usage rules for other lockups and their minimum sizes.
- Logo Forest is **`#214F40`**. Coordinate UI primary colors with it, but preserve the required off-white canvas and white card/popover tokens above; the kit's Mist **`#F5F6F3`** does not override them.
- The logo uses outlined Figtree letterforms. Figtree is available in [`design/logo/fonts/`](../design/logo/fonts/), but the UI typography scale still needs to be specified; do not recreate the logo with a font.

Show appropriate logo placement within the authentication scope only. The supplied app icons and favicons are reference assets, not requests for additional screens or implementation work.

## 4. Required shadcn/ui and Tailwind token contract

The frontend will use **shadcn/ui with Tailwind CSS v4**. Use the supplied `globals.css` structure as the naming and mapping contract, not a separate token system. Its default neutral palette is not the final brand palette.

### CSS structure

Any illustrative CSS specification must preserve:

```css
@import "tailwindcss";
@import "tw-animate-css";
@import "shadcn/tailwind.css";
```

- Define light-theme colors as **OKLCH CSS custom properties in `:root`**.
- Preserve `@theme inline` mappings from `--color-<name>` to `var(--<name>)` for every color token below.
- Use semantic utilities such as `bg-background`, `text-foreground`, `bg-primary`, and `border-border`, not hardcoded per-component palette colors.
- Preserve the base styling contract: `border-border outline-ring/50` globally, and `bg-background text-foreground` on the body.
- The supplied `.dark` block and custom dark variant are compatibility scaffolding only. Do not design dark mode or add a theme switch.

### Required color tokens

Provide each token's value, semantic role, and utility mapping. Ensure readable foreground/background pairings.

| CSS custom properties | Role |
| --- | --- |
| `--background`, `--foreground` | Off-white canvas and default text |
| `--card`, `--card-foreground` | White card surfaces and their text |
| `--popover`, `--popover-foreground` | White floating surfaces and their text |
| `--primary`, `--primary-foreground` | Forest-green primary actions and contrasting content |
| `--secondary`, `--secondary-foreground` | Secondary actions and surfaces |
| `--muted`, `--muted-foreground` | Subdued surfaces and supporting text |
| `--accent`, `--accent-foreground` | Highlighted/selected interactive surfaces and their content |
| `--destructive`, `--destructive-foreground` | Errors and destructive actions with contrasting content |
| `--border`, `--input`, `--ring` | Borders, input boundaries, visible focus rings |
| `--chart-1` through `--chart-5` | Compatible visualization palette; no charts required |
| `--sidebar`, `--sidebar-foreground` | Compatible sidebar surface pair |
| `--sidebar-primary`, `--sidebar-primary-foreground` | Compatible sidebar primary-action pair |
| `--sidebar-accent`, `--sidebar-accent-foreground` | Compatible sidebar selected-surface pair |
| `--sidebar-border`, `--sidebar-ring` | Compatible sidebar border and focus colors |

Chart and sidebar tokens are retained for compatibility only. **Do not create charts or sidebars.** Explicitly define `--destructive-foreground`, which is mapped but missing from the supplied default `:root` values.

If success, warning, or informational colors are needed, document them as explicit semantic extensions with foreground pairs and matching `@theme inline` mappings. Do not repurpose chart colors as status colors.

### Radius contract

Keep `--radius: 0.625rem` as the base and preserve these `@theme inline` definitions:

| Theme token | Definition |
| --- | --- |
| `--radius-sm` | `calc(var(--radius) * 0.6)` |
| `--radius-md` | `calc(var(--radius) * 0.8)` |
| `--radius-lg` | `var(--radius)` |
| `--radius-xl` | `calc(var(--radius) * 1.4)` |
| `--radius-2xl` | `calc(var(--radius) * 1.8)` |
| `--radius-3xl` | `calc(var(--radius) * 2.2)` |
| `--radius-4xl` | `calc(var(--radius) * 2.6)` |

Use Tailwind-aligned spacing, typography, and elevation scales. Document values and utility mappings; avoid arbitrary per-screen values.

## 5. Design-system deliverables

Create clearly named foundation and component sheets, separate from auth screen frames.

### Foundations

- Semantic colors and foreground pairings.
- Typography scale, including tabular-number and IDR formatting specimens.
- Spacing, layout, radius, elevation, and icon conventions.
- Safe-area and small-phone adaptation guidance.
- Focus, validation, loading, disabled, error, and success conventions.

### Reusable components

Use shadcn/ui primitives where applicable:

- **Button:** primary, secondary, outline, ghost, destructive, icon; loading and disabled states.
- **Input and Label:** text, email, password, amount; helper text, validation, focus, disabled states.
- **Password field:** visibility control with accessible labeling.
- **Card, Badge, Separator:** reusable surfaces and hierarchy.
- **Checkbox:** reusable control with relevant states; do not add unnecessary registration consent or marketing fields merely to demonstrate it.
- **Alert and Sonner/toast:** error, informational, and success feedback.
- **Dialog, Alert Dialog, Sheet/Drawer:** isolated component specimens with neutral content, not banking workflows.
- **Skeleton and loading indicator:** reusable loading patterns.

Document component variants, states, semantic-token usage, and intended behavior. Component specimens are allowed; assembling them into additional product screens is not.

Do not expand into an exhaustive catalog of every shadcn/ui component. Focus on reusable foundations and the components listed above, especially those used by authentication.

## 6. Allowed screen examples — authentication only

Create the following four screen templates and their relevant state variants:

### A. Registration

- Full name, email, and password.
- Primary create-account action and link to sign-in.
- Password visibility control and clear requirements/validation.
- No email verification.
- No identity documents, KYC flow, phone verification, social sign-in, or extra onboarding screens.
- Successful registration can show a brief success state. Stop before any signed-in screen.

Product context: a Main account is created automatically after registration. This does not require an account-creation or account-details screen.

### B. Sign-in

- Email and password.
- Password visibility control.
- Sign-in action, forgot-password link, and route to registration.
- Invalid-credentials feedback and submission/loading states.
- Stop at successful authentication; do not design the destination screen.

### C. Forgot password

- Email entry and reset-link request action.
- Generic submission confirmation that does not disclose whether an email is registered.
- Clear route back to sign-in.

### D. Reset password

- New password and confirmation.
- Validation, submission/loading, and completion states.
- Invalid/expired-link feedback may be a variant of this template.
- Completion returns the customer to sign-in.

Show relevant empty/default, focus, filled, validation-error, loading/disabled, failure, and success states as component variants or focused screen variants. Do not invent unrelated screens to demonstrate states.

Connect only the auth flows. No email-verification screen is permitted; password-recovery email is not registration verification.

## 7. Accessibility and mobile behavior

- Minimum **44×44-point** touch targets, including password toggles and navigation links.
- Readable text and sufficient contrast.
- Visible focus treatment and non-color-only status indicators.
- Persistent field labels and actionable error messages.
- Keyboard-aware forms with reachable primary actions on smaller phones.
- No reliance on placeholder text as the only label.
- Design annotations should identify accessible component behavior, not merely visual styling.

## 8. Expected handoff and stopping point

Organize the output as:

1. **Foundations:** token sheet, scales, and usage guidance.
2. **Components:** reusable variants and states.
3. **Authentication:** registration, sign-in, forgot password, reset password, and their relevant variants.
4. **Specifications:** auth-flow connections, accessibility/mobile notes, proposed OKLCH `:root` token block, and any added `@theme inline` mappings. CSS snippets are token specifications, not application implementation.

Use consistent component instances in all auth screens. Label any added design assumptions.

### Completion checklist

- [ ] Main canvas is explicitly off-white; cards/popovers are white.
- [ ] Tokens follow the supplied shadcn/ui and Tailwind v4 contract.
- [ ] Forest-green Quiet Confidence branding and light-only scope are preserved.
- [ ] Reusable design-system components include relevant states.
- [ ] Only the four auth screen templates and their relevant variants are designed.
- [ ] No registration email verification or identity-verification flow is added.
- [ ] No signed-in banking screen, dashboard, Home preview, or bottom navigation is created.

**Stop after the design system and authentication screens. Wait for a separate explicit request before designing any banking screens.**
