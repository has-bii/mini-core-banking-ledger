# Still Bank — UI Design Brief

## Objective

Design an original, mobile-only personal banking experience for **Still Bank**. Deliver a reusable design system and example screens demonstrating the complete agreed product scope.

This brief is intended for a UI design agent such as pen.dev or Wonder Design. Produce designs and design specifications only—not frontend code.

## 1. Audience and platform

- Customers managing their own personal money; not internal staff or business banking users.
- Native-mobile-app appearance and interaction patterns, not a mobile website.
- Mobile layouts only; no desktop designs.
- Use one reference phone size and specify adaptation to smaller and larger phones.
- Respect safe areas, mobile keyboards, and comfortable touch interaction.

## 2. Visual direction

**Direction: Quiet Confidence.** Calm, precise, trustworthy, and approachable without looking cold, flashy, or exclusive.

- Original identity using the name **Still Bank**.
- Warm off-white backgrounds, deep forest-green accents, charcoal text, subtle borders.
- Light mode only.
- Readable sans-serif typography, clear hierarchy, and tabular numerals for financial amounts.
- Generous but practical spacing, restrained rounding, and minimal decoration.
- Reserve red for errors and destructive actions.
- Distinguish incoming and outgoing money with signs and labels, not color alone.
- Use clear, reassuring, direct language. Avoid unnecessary financial jargon.

Choose exact visual values within this direction and document them as reusable tokens. Avoid generic dashboard styling, excessive gradients, and decorative elements that compete with account information.

## 3. Currency and sample content

- IDR only; no currency switching or conversion.
- Format money consistently as **Rp150.000**, without decimals.
- Use realistic fictional customers, account names, balances, and transactions.
- Keep sample balances and transactions consistent across connected screens.
- Do not invent interest rates or imply unsupported banking services.

## 4. Authentication

### Registration

- Fields: full name, email, password.
- No email verification.
- Automatically create one Main account after successful registration.

### Sign-in and recovery

- Email and password sign-in.
- Password visibility control and clear validation.
- Forgot-password and reset-password screens/states.
- Sign-out from Profile.

Keep registration simple. Do not add identity-document collection or a full identity-verification flow.

## 5. Account model

| Account type | Creation | Purpose | Closure |
| --- | --- | --- | --- |
| Main | One, created automatically | Primary balance and destination for top-ups | Cannot be closed through account management |
| Sub | Customer can create multiple | Separate everyday money | Allowed only with zero balance and no pending transfers |
| Savings | Customer can create multiple | Separate money with manual locking | Allowed only when unlocked, with zero balance and no pending transfers |

### Shared account behavior

- Show account type, nickname, balance, and relevant actions.
- Customers can update account nicknames—not balances or historical transactions.
- Present deletion as **Close account**, not erasing an account or its history.
- Preserve transaction history after closure.
- Explain closure restrictions and show a confirmation before closing.

### Savings locking

- New Savings accounts start unlocked.
- Customers manually lock or unlock them from account details.
- Locked Savings accounts can receive money.
- Locked Savings accounts cannot transfer money out or be closed.
- Show a visible Locked status and explain restrictions when an outgoing transfer is attempted.
- Unlocking requires a confirmation dialog explaining that the money becomes available to transfer.
- No PIN, waiting period, penalty, interest, savings goals, or target/progress tracking.
- Describe this as voluntary spending control—not a security lock or fixed-term deposit.

## 6. Money movement

### Transfers

Support:

1. Transfers between the customer's own accounts.
2. Transfers to another customer within Still Bank.

Do not include external-bank or international transfers.

Show source account selection, destination/recipient selection, amount entry, review, and a clear result or receipt. Respect Savings locking and make failures or unavailable actions understandable.

Recipient lookup details are not prescribed; choose a simple, consistent design without implying an external banking integration.

### Top-ups

- Top-ups go only to Main.
- Use a simulated bank-transfer funding flow: amount entry → review → confirm → receipt.
- Identify this as a demo funding flow in design annotations.
- Do not imply that real payment processing or bank integration exists.

## 7. Navigation and home

Use four bottom tabs:

1. **Home**
2. **Accounts**
3. **Activity**
4. **Profile**

### Home content

- Total balance.
- Account previews.
- Prominent **Send Money** and **Move Money** actions.
- Access to Main-account top-up.
- Recent transactions.
- A small services section with labeled shortcuts for future features, including bill payments.

### Future-service shortcuts

- Clearly label unavailable services **Coming soon**.
- Tapping a shortcut explains that it is not available yet.
- Do not let customers begin an unfinished payment flow.
- Keep this section small and secondary to working banking actions.

## 8. Example screen coverage

Use reusable screen templates and variants rather than drawing every combination as a separate screen.

| Area | Required examples |
| --- | --- |
| Access | Registration, sign-in, forgot/reset password |
| Home | Balance overview, account previews, shortcuts, recent activity, coming-soon feedback |
| Accounts | Account list, account details, create, rename, close and closure restrictions |
| Savings | Locked and unlocked states, lock/unlock interactions, unlock confirmation, restricted transfer feedback |
| Transfers | Own-account transfer, another-customer transfer, review, result/receipt |
| Top-up | Amount entry, review, confirmation/result and receipt |
| Activity | Transaction list, filters, transaction details |
| Profile | Customer details, sign-out |

Include representative **loading, empty, validation, failure, and success states** across the relevant flows. Do not limit examples to populated happy-path screens.

## 9. Reusable design system

Create foundations and components that directly support the example screens. The frontend will use **shadcn/ui with Tailwind CSS v4**. Design components and tokens for that stack, while retaining the native-mobile-app appearance.

### Required shadcn/ui token contract

Use the supplied `globals.css` structure as the naming and mapping contract—not a separate, incompatible token system. Its neutral color values are a starting point; choose branded light-theme values that express the agreed warm off-white, forest-green, and charcoal direction.

- Preserve the imports for `tailwindcss`, `tw-animate-css`, and `shadcn/tailwind.css` in any illustrative CSS specification.
- Define theme color values as **OKLCH CSS custom properties in `:root`**.
- Preserve the `@theme inline` mapping from `--color-<name>` to `var(--<name>)`. Components should use semantic utilities such as `bg-background`, `text-foreground`, `bg-primary`, and `border-border`, rather than hardcoded palette colors.
- Keep the base styling contract: `border-border outline-ring/50` globally and `bg-background text-foreground` on the body.
- Include every color token listed below in the token sheet, with its value and intended use. Keep foreground/background pairs legible at their actual usage sizes.

| CSS custom properties | Intended role |
| --- | --- |
| `--background`, `--foreground` | App canvas and default text |
| `--card`, `--card-foreground` | Account cards and other card surfaces |
| `--popover`, `--popover-foreground` | Floating surfaces and their text |
| `--primary`, `--primary-foreground` | Forest-green primary actions and contrasting content |
| `--secondary`, `--secondary-foreground` | Secondary actions and surfaces |
| `--muted`, `--muted-foreground` | Subdued surfaces and supporting text |
| `--accent`, `--accent-foreground` | Highlighted or selected interactive surfaces and their content |
| `--destructive`, `--destructive-foreground` | Errors and destructive actions with contrasting content |
| `--border`, `--input`, `--ring` | Borders, input boundaries, and visible focus rings |
| `--chart-1` through `--chart-5` | Compatible data-visualization palette; defining these does not require adding charts |
| `--sidebar`, `--sidebar-foreground` | Sidebar-compatible surface pair |
| `--sidebar-primary`, `--sidebar-primary-foreground` | Sidebar-compatible primary-action pair |
| `--sidebar-accent`, `--sidebar-accent-foreground` | Sidebar-compatible selected-surface pair |
| `--sidebar-border`, `--sidebar-ring` | Sidebar-compatible border and focus colors |

Sidebar tokens are retained for compatibility only. Do not add a sidebar to the mobile navigation.

Use the supplied radius contract, with `--radius: 0.625rem` as the base:

| Tailwind theme token | Definition |
| --- | --- |
| `--radius-sm` | `calc(var(--radius) * 0.6)` |
| `--radius-md` | `calc(var(--radius) * 0.8)` |
| `--radius-lg` | `var(--radius)` |
| `--radius-xl` | `calc(var(--radius) * 1.4)` |
| `--radius-2xl` | `calc(var(--radius) * 1.8)` |
| `--radius-3xl` | `calc(var(--radius) * 2.2)` |
| `--radius-4xl` | `calc(var(--radius) * 2.6)` |

Use Tailwind-aligned spacing, typography, and elevation scales; document chosen values and their utility mappings. Avoid arbitrary per-screen values.

If distinct success, warning, or informational colors are necessary, document them as explicit semantic extensions with foreground pairs and matching `@theme inline` mappings. Do not misuse chart colors as status tokens.

**Light mode remains the only design deliverable.** The supplied `.dark` values and custom dark variant are compatibility scaffolding, not a request for dark screens or a theme switch.

### Foundations

- Semantic color tokens, including text, surfaces, borders, actions, and status colors.
- Typography scale and financial-number styling.
- Spacing, layout, corner radius, elevation, and icon conventions.
- Safe-area and mobile layout guidance.
- Currency formatting and concise UI-copy conventions.

### Components

Cover the components needed by the flows, including:

- Buttons and icon buttons with relevant interaction states.
- Text, email, password, and amount inputs; labels, helper text, and errors.
- Account cards, balance displays, account selectors, and status badges.
- Transaction rows and transaction-detail patterns.
- Recipient selection and review/receipt patterns.
- Bottom navigation, headers, and back navigation.
- Dialogs or sheets for confirmations and focused actions.
- Loading indicators/skeletons, empty states, error feedback, and success feedback.
- Coming-soon service shortcuts.

Document variants, states, and intended usage. Reuse these components across screens rather than introducing one-off styling. Map applicable components to shadcn/ui primitives such as Button, Input, Label, Card, Badge, Select, Dialog, Alert Dialog, Sheet/Drawer, Skeleton, and Sonner. Compose banking-specific patterns from these primitives; identify custom patterns such as bottom navigation explicitly. Preserve accessible focus, disabled, validation, and destructive states rather than replacing them with purely visual approximations.

## 10. Accessibility and mobile behavior

- Minimum **44×44-point** touch targets.
- Readable text and sufficient contrast.
- Never communicate status through color alone.
- Clear field labels and actionable validation messages.
- Keyboard-aware forms with visible primary actions.
- Distinguish destructive actions from routine actions.
- Specify how content adapts across phone sizes without adding desktop layouts.

## 11. Expected handoff

Provide:

1. A foundations/token sheet using the exact shadcn/ui CSS variable names above, with OKLCH light-theme values, semantic roles, Tailwind utility mappings, and the supplied radius scale. Include a proposed `:root` token block and any added `@theme inline` mappings as design specifications, not application implementation.
2. A reusable component library with variants and states.
3. Example screens covering the agreed flows.
4. Flow connections or annotations explaining navigation, confirmations, and restrictions.
5. Responsive-mobile and accessibility notes.
6. Explicit annotations identifying simulated funding and simplified onboarding.

Use clear page, frame, and component names. Label any additional design assumptions rather than presenting them as confirmed product requirements.

## 12. Boundaries

This is a design concept, not a production banking or compliance specification.

Out of scope:

- Frontend implementation or backend/API design.
- Desktop layouts and dark mode.
- Business accounts and internal operations tools.
- External-bank transfers, international transfers, and currency conversion.
- Functional bill payments or other coming-soon services.
- Cards, loans, investments, and rewards.
- Savings interest, goals, automated rules, penalties, or time-based locking.
- Full identity verification or real funding integrations.

Do not silently expand the product beyond this brief.
