# Still Bank — Open Ground handover

## Approval and scope

The user chose **B** (“I pick B”) and then explicitly authorised the offered full kit (“go ahead”). See `APPROVAL.md`. The original brief remains unchanged.

This delivery refines and exports that chosen direction. It does not invent product features, a customer niche, regulatory status, branches, payment cards, security guarantees or financial outcomes. Presentation applications are isolated identity mockups, not interface designs. No backend, frontend/application component or UI token system was modified.

## Final design decisions

- **Keep the choice:** the single open arch and inward-turning foot remain the identifier. The creative rationale is a quiet, grounded centre, not a security promise.
- **Refine the crown:** v3 aligns both outer springlines at y112 and makes the outer crown symmetrical. It retains the 40-unit stems, 44-unit return and 32-unit bottom exit. The normal symbol's visible footprint is 176 × 184 on a 256-unit canvas, deliberately slightly above geometric centre.
- **Make small use robust:** the separate 16–24 px cut uses a 192 × 192 footprint and 48-unit stems. At 16 px that gives 3 nominal stem pixels and a 2 px open exit. It is drawn and rendered at native size, not merely reduced from a large PNG.
- **Correct reversal:** normal 40-unit symbol bands become about 38 units; small 48-unit bands become about 46. The outlined name uses Figtree 560 instead of 580 for white, but retains the same advance/kerning positions. The custom k also has a lighter white outline.
- **Keep type readable:** title-case Still Bank, source Figtree, custom straightened a terminal and 45° k arms, fixed spacing. No substituted letters or pictograms. Normal and reversed name treatments are both paths.
- **Colour:** Forest `#214F40` plus Mist `#F5F6F3`, with utility black/white. Colour and brand-mono are intentionally the same single-green design. RGB/approximate CMYK specifications and print caveats are in `USAGE.md`.
- **Icons:** full-square opaque RGB app/touch/web PNGs; no pre-baked platform corner mask. A separate maskable asset has a safely inset foreground. Favicons use the sturdier white small cut on Forest so they remain visible in light/dark browser chrome.

Meaningful concept v1/v2 and final-refinement v3 files remain separate in the repository. The production kit uses the final SVGs under `svg/`, not the archived concepts.

## Delivered files

- 20 SVG logo files: symbol, small symbol, name, horizontal and stacked, each in colour/black/corrected-white/brand-mono.
- 44 transparent logo PNGs: regular forms at useful 512/1024 or 600/1200 widths, plus native 16/24/32/64 small cuts.
- Three icon SVGs, 11 opaque RGB icon PNGs, multi-size favicon.ico and icon-only webmanifest.
- Five-slide presentation with six isolated applications: OS tile, browser-tab favicon, launcher masks, blank stationery, compact signature and one-colour imprint. No fictional contacts, URL, card data, ratings or UI screens.
- One-page visual usage guide as SVG/PNG/PDF, complete usage notes, original licensed Figtree source, test evidence and machine-readable checksummed inventory.

`FILE-INVENTORY.json` records exact paths, dimensions, modes and SHA-256 values. The guide PDF is a document; no CMYK press-master PDF, AI or EPS is claimed.

## Tests performed after final refinement

### Vector checks

- Skill SVG audit on all **23 production logo/icon SVGs**: **99–100/100**, no `FAIL` or `WARN` findings. Remaining information is intentional optical placement or point-count guidance.
- Independent structural checks: XML parses; finite, positive viewBoxes; accessible titles; closed path contours; only flat vector paths; no live type, rasters, filters, masks, styles, external dependencies or strokes in production logos/icons.
- All four primary logo types are separate self-contained masters; background transparency is not simulated with white rectangles.

### Raster and size checks

- Headless Chrome render and visual inspection of `still-bank-logo-kit.png`, `tests/final-proof.png`, `tests/refinement-and-transforms.png`, `tests/final-shelf-test.png`, the visual usage guide, and the rendered presentation/contact sheet.
- Actual-size positive/reversed identifier tests at **16 / 24 / 32 / 64 px**. Small cut at 16/24, regular at 32/64. Nearest-neighbour pixel enlargements show the true 16 px raster.
- Four-connected-component diagnostics on all eight light/dark size crops: a single connected mark in each. This is a stroke-continuity diagnostic, not an automatic readability score or audience test.
- Compact name treatments inspected at the documented minima: horizontal 160 px wide; wordmark and stacked 128 px wide.
- All 44 production logo PNGs checked for genuine transparency and expected dimensions. All 11 icon PNGs checked for opaque RGB with Forest corner pixels. No alpha channel in app/store/touch/web tile PNGs.
- ICO decoded with the expected 16/32/48/64 px entries. Every webmanifest icon path resolves to a supplied file of the stated dimensions.
- Maskable foreground measured from the actual 512 px raster: maximum radius **195.19 px**, inside the **204.8 px** safe circle. Regular app foreground maximum radius **249.64 px**, inside a 256 px circular launcher crop.
- The usage PDF has one A4-landscape page. Presentation artwork was inspected after ensuring every mockup uses a supplied fixed lockup.

### Colour, balance and readings

- Forest on white/Mist; corrected white on Forest/charcoal; black on white; monochrome/greyscale; large contour and letter junction checks.
- sRGB contrast values: Forest/white **9.30:1**, Forest/Mist **8.58:1**, white/test charcoal **17.40:1**. These are identity value checks, not accessibility certification for a product.
- Mirror, 180° rotation and soft-blur checks rendered and inspected. The known n/doorway/vessel association remains; the full name should accompany the new identifier while recognition builds.

## Visual similarity research and limits

The final silhouette was compared again against the six retrieved Monzo, N26, Starling Bank, Revolut, Wise and Chase snapshots in `tests/final-shelf-test.png`. Earlier concept-stage comparisons also included bundled geometric/type peers, notably Square, Adyen, Stripe, Scribd, Svelte and Semantic UI. Third-party marks were studied, not traced; rights remain with their owners.

This was **limited visual research**, not an exhaustive look-alike check. Retrieved category assets were not verified against current official brand manuals. No trademark-database, reverse-image, market-wide, app-store, independent audience or jurisdiction-specific cultural search was performed. Obtain professional name/device clearance in the intended markets before launch.

## Open production items

1. **Legal:** professional trademark/name clearance; similarity research does not establish registrability or exclusivity.
2. **Print:** ICC-aware conversion for the actual stock/press, physical proof, and verified custom-spot/Pantone match. CMYK estimates are not press specifications. Suggested print minimums are not physically tested.
3. **Platform:** real-device launcher/touch tests, app-store requirements and final deployment integration. The webmanifest is icon metadata only and makes no installability claim.
4. **Recognition:** validate the abstract n/doorway reading with independent viewers if this is a launch-critical concern.

## Rebuilding in the repository

Use the existing isolated Python environment with fonttools, uharfbuzz and Pillow plus Chrome and the installed logo-design skill:

```sh
/tmp/still-bank-logo-tools/bin/python logo/tools/build_final.py --vectors --png
/tmp/still-bank-logo-tools/bin/python logo/tools/build_final_review.py
/tmp/still-bank-logo-tools/bin/python logo/tools/build_final_presentation.py
/tmp/still-bank-logo-tools/bin/python logo/tools/validate_final.py
```

Inspect regenerated renders, then refresh the inventory/archive with `logo/tools/package_final.py`. Fonts and browser versions can change rasterisation; do not assume a rebuild is visually identical.
