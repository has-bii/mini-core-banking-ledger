# Still Bank — logo usage

**Direction B: Open Ground.** An open arch with an inward-turning foot creates a quiet, grounded identifier. This is a design rationale, not a product or security promise.

[One-page visual guide (PDF)](guidelines/still-bank-usage-guide.pdf) · [PNG](guidelines/still-bank-usage-guide.png)

## Choose the right file

- **Default:** `svg/still-bank-horizontal-color.svg` — compact symbol + readable name.
- **Tall spaces:** `svg/still-bank-stacked-color.svg`.
- **Name only:** `svg/still-bank-wordmark-color.svg` — do not recreate it by typing.
- **Identifier:** `svg/still-bank-symbol-color.svg` at 32 px and above; `svg/still-bank-symbol-small-color.svg` at 16–24 px.
- **Dark backgrounds:** the corresponding `-white.svg` file. It is optically lighter, not just a colour swap.
- **One-colour print:** `-black.svg` or `-mono-214f40.svg`. The full-colour logo is intentionally one forest green, so `color` and brand `mono` have identical visible artwork.

Use supplied lockups intact. Do not resize the mark relative to the name or recreate a stacked arrangement from separate files.

## Clear space and fixed relationships

Keep at least **one symbol stem-width X** clear around the **visible artwork** on all sides. In the 256-unit regular symbol X = 40. For the wordmark-only version, use one-third of its capital height. SVG canvas edges are not clear-space boundaries; add layout space where needed.

| Version | SVG canvas | Fixed construction |
| --- | --- | --- |
| Symbol | 256 × 256 | Visible bounds x40–216, y32–216; 40-unit stems, 44-unit return, 32-unit open exit. |
| Horizontal | 824 × 256 | Symbol scale 0.76; name em 136, baseline 174; visible symbol/name gap ≈60 units (about 2X). |
| Stacked | 384 × 256 | Symbol scale 0.60 above centred name; name em 80, baseline 224; visible vertical gap ≈32 units. |
| Wordmark | 736 × 256 | Name em 164, baseline 183; fixed custom spacing. |

Icons have their own built-in margins. Do not add a second tile, outline, inset border or rounded corner to a production icon file.

## Minimum sizes

| Artwork | Screen minimum | Print starting point, subject to proof |
| --- | --- | --- |
| Small identifier | 16 px; use at 16–24 px | Use regular artwork for print |
| Regular identifier | 32 px | 6 mm visible symbol width |
| Horizontal | 160 px canvas width | 32 mm canvas width |
| Wordmark | 128 px canvas width | 24 mm canvas width |
| Stacked | 128 px canvas width | 30 mm canvas width |

Screen minima were visually inspected in Chrome at actual pixel sizes. Increase size for noisy or low-quality reproduction. Print starting points are conservative recommendations, **not physically proofed production minima**. The small cut uses 48-unit stems (3 nominal pixels at 16 px) and a 32-unit open exit (2 pixels at 16 px); white geometry is slightly thinned.

## Colour and backgrounds

| Colour | HEX | RGB | Approx. unprofiled CMYK | Spot / Pantone |
| --- | --- | --- | --- | --- |
| Forest | **#214F40** | 33, 79, 64 | 58, 0, 19, 69 | Not assigned; match against physical proof |
| Mist | **#F5F6F3** | 245, 246, 243 | 0, 0, 1, 4 | Paper/stock or physical match |
| Utility white | #FFFFFF | 255, 255, 255 | 0, 0, 0, 0 | Paper or opaque white ink |
| Utility black | #000000 | 0, 0, 0 | 0, 0, 0, 100 | Printer-selected single black |

- Forest on white / Mist; corrected white on Forest / a sufficiently dark, calm background; black on white.
- Contrast: Forest/white **9.30:1**, Forest/Mist **8.58:1**. Logos are exempt from WCAG text-contrast requirements; these are value checks, not product accessibility certification.
- Do not put Forest directly on a dark background, or white on a light background. On photos/patterns, use a calm, solid approved container with clear space.
- SVG and PNG exports are RGB/screen artwork. CMYK figures above are simple estimates, not ICC-managed press specifications. Have a printer convert for the intended stock/profile and proof any spot/Pantone match before production.

## App, favicon and web assets

- `icons/app-icon-1024.png` — opaque RGB full square, no alpha channel and no baked-in corner mask. The platform supplies its mask. `app-icon.svg` is the vector source.
- `icons/apple-touch-icon.png` — 180 × 180; `icon-192.png` / `icon-512.png` — standard web tiles.
- `icons/maskable-512.png` — a separately inset foreground inside the central 80%-diameter safe circle. Use this, not the regular app tile, for `purpose: maskable`.
- `icons/favicon.svg`, `favicon-{16,24,32,48,64}.png`, `favicon.ico` — white small-size identifier on Forest, legible on light and dark browser chrome. ICO contains 16/32/48/64 px entries.
- `icons/site.webmanifest` is **icon metadata only**, with relative paths. Merge/adapt it to the deployment; it does not define an app entry route, installation behaviour or a service worker. No frontend/application code was changed or supplied.

## Type, licensing and misuse

The logo name is based on **Figtree**, weight 580 normal / 560 reversed, with custom a/k outlines and fixed normal-weight spacing in both versions. The symbol is an original drawn contour. All logo letterforms are paths; no installed font is required to display the logos.

The unmodified source font and complete **SIL Open Font License 1.1** are in `fonts/`. Source: [Google Fonts / Figtree](https://github.com/google/fonts/tree/main/ofl/figtree); copyright 2022 The Figtree Project Authors. No modified font binary is distributed. For non-logo brand copy, Figtree regular/semibold may be used under that licence; this kit does not define a UI typography system.

**Do not:** stretch, rotate, mirror, freely recolour, add shadows/gradients/strokes, close the open exit, retype the name, alter lockup proportions, or place the logo on busy imagery without a container. Misuse examples are illustrated on the one-page guide; those are not alternative assets.

Obtain professional trademark/name clearance and physical print/real-device proofs before launch. See `HANDOVER.md` for scope and limitations.
