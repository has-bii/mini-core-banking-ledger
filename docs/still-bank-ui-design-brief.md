# Still Bank — Design System, Reusable Components, and Authentication Examples

## 1. Task and strict scope

Start from scratch.

Create only:

1. A reusable **design system** for Still Bank.
2. **Reusable component definitions** with variants and states.
3. **Authentication example pages** demonstrating that system through shared component instances.

**Do not design the rest of the banking application.** Do not generate Home, dashboards, Accounts, Savings, Transfers, Top-up, Activity, transaction receipts, Profile, or bottom-navigation screens—not even as examples or previews. Banking context does not authorize banking screens.

This brief is for **Wonder Desktop**, working in its own repository worktree. Use repository-relative paths to access the existing brand assets. If a required asset is unavailable, flag the missing dependency; do not recreate it.

Deliver editable design foundations, reusable components, and authentication example pages in Wonder, with component and token specifications. This is design work only: do not create or modify application code, stylesheets, backend code, dependencies, or configuration. CSS snippets belong in the design specifications only.

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

The frontend will use **shadcn/ui with Tailwind CSS v4**. The token names, values, and mappings specified below are the authoritative design contract; no external `globals.css` is required. Do not introduce a separate token system or treat the default shadcn neutral palette as the final brand palette.

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
- Specify the light theme only. Do not generate a `.dark` block, design dark mode, or add a theme switch.

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

Chart and sidebar tokens are retained for compatibility only. **Do not create charts or sidebars.** Explicitly define `--destructive-foreground` and its utility mapping alongside the other tokens.

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

Create the following four example-page templates and their relevant state variants. These pages demonstrate the design system; they are not a request to implement authentication.

### Backend-aligned behavior and password policy

The current behavior was checked against `backend/internal/modules/auth/validator.go`, `service.go`, and `http.go`. Wonder does not need to inspect or change application code.

- Registration requires a full name, a valid email address, and a password. Name and email have surrounding whitespace trimmed.
- Registration and sign-in passwords require at least **8 characters** and at most **72 UTF-8 bytes**, and cannot be entirely whitespace. Do not describe the byte limit as 72 characters; multibyte characters can reach it sooner.
- Passwords are not trimmed. Do not add uppercase, lowercase, number, or symbol requirements, a strength-based acceptance rule, or an invented maximum name length.
- Suggested password helper copy: “Use at least 8 characters.” For a byte-limit error: “Password is too long. Use no more than 72 bytes; some characters use more than one byte.”
- Include duplicate-email registration feedback and generic “Invalid email or password” sign-in feedback. Do not identify which credential was incorrect.
- Registration does not create a signed-in session. Its success state should offer a route to sign-in, not imply automatic sign-in.
- The auth module currently exposes registration, sign-in, sign-out, and current-user endpoints only. Forgot-password and reset-password examples below are **proposed design behavior**, not existing backend functionality. Label them accordingly in the specifications; do not add endpoints or implementation work.
- For the proposed reset-password example, reuse the registration password policy and require matching confirmation. Label this as proposed behavior rather than an existing backend contract.

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

1. **Foundations:** editable foundation sheets with tokens, scales, and usage guidance.
2. **Components:** reusable component definitions with named variants and states, not merely duplicated drawings.
3. **Authentication examples:** editable registration, sign-in, forgot-password, and reset-password page frames and their relevant variants, assembled from the shared components.
4. **Specifications:** linked auth-flow connections, accessibility/mobile notes, proposed OKLCH `:root` token block, and `@theme inline` mappings for every required color token and any extensions. CSS snippets are token specifications, not application implementation.

Use consistent component instances in all auth screens. Label any added design assumptions.

### Completion checklist

- [ ] Main canvas is explicitly off-white; cards/popovers are white.
- [ ] Tokens follow the self-contained shadcn/ui and Tailwind v4 contract in this brief.
- [ ] Forest-green Quiet Confidence branding and light-only scope are preserved.
- [ ] Reusable design-system components include relevant states.
- [ ] Only the four auth screen templates and their relevant variants are designed.
- [ ] No registration email verification or identity-verification flow is added.
- [ ] No signed-in banking screen, dashboard, Home preview, or bottom navigation is created.

**Stop after the design system and authentication screens. Wait for a separate explicit request before designing any banking screens.**
