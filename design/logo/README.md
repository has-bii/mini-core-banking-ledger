# Still Bank — Open Ground logo kit

**Selected direction: B. Kit production explicitly authorised by the user.** See [approval record](APPROVAL.md).

## Start here

- [Kit overview](still-bank-logo-kit.png)
- [One-page usage guide — PDF](guidelines/still-bank-usage-guide.pdf) · [PNG](guidelines/still-bank-usage-guide.png)
- [Complete usage rules](USAGE.md)
- [Handover and limitations](HANDOVER.md)

## Primary files

| Use | File |
| --- | --- |
| Default compact logo | `svg/still-bank-horizontal-color.svg` |
| Stacked logo | `svg/still-bank-stacked-color.svg` |
| Name only | `svg/still-bank-wordmark-color.svg` |
| Identifier, 32 px+ | `svg/still-bank-symbol-color.svg` |
| Identifier, 16–24 px | `svg/still-bank-symbol-small-color.svg` |
| White on dark | Corresponding `-white.svg` — optically corrected |
| Black / brand monochrome | Corresponding `-black.svg` / `-mono-214f40.svg` |
| App/store source | `icons/app-icon-1024.png` — opaque RGB, square |
| Browser favicon | `icons/favicon.svg` / `icons/favicon.ico` |
| Web/PWA icons | `icons/icon-{192,512}.png`, `icons/maskable-512.png` |

Every logo SVG is self-contained outlined vector artwork. No installed font is needed. All masters use flat fills, real open/negative spaces, and no live text, raster images, effects, masks or unexpanded strokes.

**Palette:** Forest `#214F40`, Mist `#F5F6F3`; utility black/white. Full colour intentionally uses only Forest, so colour and brand-monochrome versions have the same visible artwork.

## Inventory

```text
svg/            20 SVG logo files: 5 forms × colour / black / white / brand mono
png/            44 transparent PNGs: 32 regular logo exports + 12 native small-size exports
icons/          3 vector icon sources, 11 opaque RGB PNGs, 4-size ICO, icon-only webmanifest
guidelines/     One-page usage guide: SVG / PNG / PDF
fonts/          Unmodified Figtree variable source and complete SIL OFL 1.1 licence
FILE-INVENTORY.json  Per-file byte sizes, SHA-256 and raster dimensions/modes
```

### PNG sizes

- Symbol: 512 / 1024 px square, all four colour versions.
- Horizontal and wordmark: 600 / 1200 px wide, native aspect ratio, all four versions.
- Stacked: 512 / 1024 px wide, native aspect ratio, all four versions.
- Small symbol: 16 / 24 / 32 / 64 px square in colour, black and corrected white. Use at 16–24 px; regular artwork is the default from 32 px up.
- App icon: 1024 / 512 px; web tiles: 192 / 512 px; Apple touch: 180 px; maskable: 512 px.
- Favicon PNG: 16 / 24 / 32 / 48 / 64 px. ICO: 16 / 32 / 48 / 64 px.

The kit overview and usage guide are reference documents, not production logo variants.

## Important limits

- This is an identity kit, not UI screens or application code. The webmanifest contains icon metadata only; deployment integration remains with the application owner.
- Visual similarity review is limited research, **not trademark clearance**. The abstract n/doorway reading remains a known trade-off.
- The guide PDF is a usage document, not CMYK press-ready logo artwork. Printers must convert/proof the RGB vectors for the intended stock/profile; no verified Pantone/spot match is assigned.
- Real-device / app-store acceptance and physical print proofs have not been performed.

Production assets live in repository-root `design/logo/`. Earlier concepts, research, test evidence, presentations and build scripts were removed during cleanup; the supplied masters are the source of truth.
