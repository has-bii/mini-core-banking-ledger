# MiniLedger — usage guide

**Primary logo:** plum horizontal lockup on white or paper. Use the symbol alone when the MiniLedger name is already visible or space is limited. Preserve the supplied proportions and spacing.

| Asset | Minimum screen width | Suggested minimum print width |
| --- | --- | --- |
| Standard symbol | 32 px | 6 mm |
| Small-size symbol | 16 px | Use the standard drawing for print |
| Horizontal lockup | 160 px | 35 mm |
| Stacked lockup | 140 px | 30 mm |
| Wordmark only | 120 px | 26 mm |

**Clear space:** let X be the width of an outer vertical stem, approximately 42 units in the 256-unit standard symbol canvas. Leave at least X beyond the visible logo on every side. Built-in SVG padding is part of this clearance; add the remaining space in the placement. Do not place borders, text or nearby icons inside it. App tiles use their supplied internal spacing.

**Small sizes:** use `symbol-small` at 16–24 px. It moves the two approved halves apart by about 6 units each on the 256-unit grid, keeping the center visible. The supplied favicon is separately fitted for browser tabs. Do not squeeze the full wordmark into an app tile or favicon. Use the normal symbol from 32 px upwards.

| Color | HEX | RGB | Role |
| --- | --- | --- | --- |
| Plum | #7040BA | 112, 64, 186 | Primary logo, app tile |
| Mint | #BFE8DA | 191, 232, 218 | Supporting background |
| Lilac | #EEE5FB | 238, 229, 251 | Supporting background |
| Ink | #17151B | 23, 21, 27 | Monochrome mark and text |
| Paper | #F7F5FC | 247, 245, 252 | Light background |
| White | #FFFFFF | 255, 255, 255 | Reversed mark |

Use plum or ink on white, mint, lilac or paper. Use white on plum or dark ink. The supplied black variant is for strict one-color reproduction. Keep mint and lilac as backgrounds, not pale logo foregrounds on white. Checked contrast includes plum/mint 5.04:1, plum/white 6.71:1 and ink/white 18.11:1.

**Typography:** the outlined wordmark is based on Plus Jakarta Sans Bold (700). Use the supplied paths rather than retyping it. Plus Jakarta Sans remains suitable for interface text; existing Bricolage Grotesque headings can remain in the product. The font is licensed under the [SIL Open Font License 1.1](https://raw.githubusercontent.com/google/fonts/main/ofl/plusjakartasans/OFL.txt); the license permits use in documents and artwork. No font software is bundled.

**App and web:** use the white M on plum. `miniledger-app-icon.svg` is a rounded presentation tile. `app-icon-square.svg` and the Apple/PWA PNGs have opaque full-bleed backgrounds for platform masking. `maskable-512.png` keeps the symbol within the central safe circle. Retain each asset's supplied internal padding.

**Avoid:** stretching, rotation, closing the center gap, changing the relationship between symbol and wordmark, gradients, shadows, outlines, or placing the mark on a busy image without a solid backing. The standard reversed artwork was visually checked; use its supplied geometry.

SVG and PNG masters are RGB assets. Print dimensions above are practical starting points; a printer should proof ink spread and convert color through the intended output profile. No Pantone match or CMYK separation is specified. This kit does not establish trademark clearance.
