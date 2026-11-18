# Still Bank UI design

## Scope

This folder governs UI design. Edit `still-bank.pen` as the pen.dev design artifact; application implementation is outside this guide's scope. Preserve existing foundation boards, reusable masters, and unrelated work when adding or refining designs.

## Consult the right sources

Paths below are relative to this folder.

- **Product intent and users:** Read [docs/problem-and-users.md](docs/problem-and-users.md) before defining journeys or evaluating usability.
- **Feature scope:** Read [docs/mvp-list.md](docs/mvp-list.md) before adding or changing product capabilities. It takes precedence over the earlier problem definition where wallet or savings behavior differs.
- **Screens, navigation, and states:** Read the relevant inventory entries, agreed decisions, and flow maps in [docs/screen-inventory-and-flow-map.md](docs/screen-inventory-and-flow-map.md) before designing a journey. Use its screen IDs for traceability and cover the applicable overlays, errors, and transient states.
- **Components and foundations:** Read [docs/design-system.md](docs/design-system.md) before editing the canvas. Reuse its existing masters, variables, and interaction conventions; extend the system when needed rather than creating disconnected alternatives.
- **Brand assets:** Read [../logo/README.md](../logo/README.md) and [../logo/USAGE.md](../logo/USAGE.md) before placing or changing logos. Use the supplied production artwork intact and follow its variant, size, background, and clear-space rules.
- **Visual inspiration:** Inspect relevant images in [../references/](../references/) when exploring layout or interaction presentation. These are inspiration, not authority for product behavior, brand identity, or feature scope.

Product docs govern behavior; the design-system documentation governs UI conventions; the logo kit governs brand usage. Ask for clarification when authoritative sources conflict or a requirement is missing rather than inventing behavior.

## Design workflow

1. Identify the requested journey, screen IDs, or components and consult the applicable sources above.
2. Inspect the existing canvas and reusable masters using pen.dev editor/tooling when available. Make targeted changes that preserve component connections and supported variable bindings; avoid blind edits to the `.pen` data. If safe editing tools are unavailable, explain the limitation before proceeding.
3. Keep foundation changes consistent across the canvas and affected exports: [tokens/still-bank.tokens.json](tokens/still-bank.tokens.json), [tokens/still-bank.css](tokens/still-bank.css), and [tokens/contrast-audit.json](tokens/contrast-audit.json). Recompute affected contrast checks when colors change. Consult the design-system note about explicit control dimensions when changing size tokens.
4. Inspect affected layouts and states visually. Apply the relevant checks in the design-system review checklist, including clipping, wrapping, component sizing, contrast, and long IDR amounts. Review affected screens against their inventory rules. Distinguish canvas checks from runtime accessibility or device behavior that a static design cannot verify.
5. Finish with a brief summary of changed screens/components, applicable screen IDs, consulted references, and unresolved issues or unavailable checks. Claim verification only for checks actually performed; no separate report file is required.
