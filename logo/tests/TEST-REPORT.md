# Still Bank — concept-stage test report

**Archive: concept-stage tests, not production assets.** Two visual/refinement rounds are retained. The user subsequently selected B and explicitly authorised its kit. Current final tests are in `../final/tests/` and `../final/HANDOVER.md`. The supplied brief remains unchanged.

## What was actually run

- The skill's SVG audit on all 12 v1 and all 13 v2 concept SVGs.
- The skill's HTML preview generator in both rounds, with six retrieved banking/fintech reference marks.
- Headless Chrome raster rendering of both rounds' concept/proof material, actual-size samples, reversal tests, transform/squint boards and reference shelves.
- Visual inspection of the rendered overview and the evidence, shelf and transform PNG boards; not just their SVG source.
- Exact-size 16, 24, 32 and 64 px light/dark sample crops from 1800 × 1280 proof boards, checked to be unscaled at output. Enlargements use nearest-neighbour pixels, not vectors redrawn at a larger size.
- A four-connected-component diagnostic at a 128 luminance threshold on each raster sample. All 12 normal and 12 reverse samples per round retained one connected ink component. This confirms stroke continuity only; it is not an automatic readability or distinctiveness score.

## Round 1 → Round 2 refinements

| Direction | Observation | Change |
| --- | --- | --- |
| A / Level S | Equal 40-unit horizontals felt slightly heavy in the centre; reversing increased apparent weight. | Central horizontal reduced to 38 units; reverse proof reduces approximately 40-unit bands to 38 units and opens the inner turns. |
| B / Open Ground | The arch could look too much like a plain n; its return foot needed to survive small reproduction. | Flatter crown, larger inward foot (ends at x144 instead of x136), 44-unit foot depth and a deliberately open 32-unit exit. Reverse stems reduced from 40 to 38 units. |
| C / Shared Line | The normal st works, but its small s counter and curved t foot get busy at favicon scale. | Shared junction adjusted; optically raised identifier; a 16/24 px test cut increases visible width from 208 to approximately 220 units and simplifies the t foot. The 32/64 px identifier remains the regular drawing. |
| Name treatments | The source face produced three audit angle warnings: a slight a-stem taper and k-arm angles close to 45°. | Straightened the a's outer terminal and constructed an exact-45° k with optically lighter diagonal strokes. Rounded lockup canvas widths to an 8-unit grid. |

The separate v1 artwork remains available for comparison. Proofs use candidate colour only; they are not the final variant set.

## Current outcomes

| Check | A / Level S | B / Open Ground | C / Shared Line |
| --- | --- | --- | --- |
| 16 / 24 / 32 / 64 px | S silhouette and counters remain recognisable in the inspected crops. | Open arch and return remain visible; n/doorway ambiguity remains. | st remains readable in the inspected crops; use the simplified test cut at 16/24 px. |
| Monochrome / greyscale | Single filled contour; does not rely on colour. | Single filled contour; does not rely on colour. | Single filled st contour and outlined name; does not rely on colour. |
| White on dark | Inspected with a slightly thinned reversal proof. | Inspected with a slightly thinned reversal proof. | Inspected at base weight; no separate optical thinning was needed for these screen proofs. Recheck after final refinement. |
| Light / dark / forest | Inspected on white/light neutral, charcoal and candidate green. | Same. | Same. |
| Compact name | Readable at 256 px-wide proof size. | Readable at 256 px-wide proof size. | Readable at 256 px-wide proof size; lowercase styling is intentional. |
| App tile | Symbol fits the 96 and 32 px isolated tile proofs without clipped terminals. | Fits; the arch's return remains visible. | Fits; small-cut lettermark is used in the smaller tile proof. |
| Large form / junctions | Inspected in large overview, filled contour and custom name treatment. | Inspected crown, stem transitions and foot; above-centre placement is deliberate. | Inspected shared st bar, t foot and separate l pair; no accidental joining of ll. |
| Mirror / 180° / blur | Recognisable S mass; rotation preserves much of the shape. | Rotation resembles a vessel/U; this is a noted metaphor/reading risk. | Rotation/mirror change the letters; no additional concerning reading observed in this limited review. |

No independent reader/audience test was performed. “Readable” and “recognisable” above are the designer's inspection, not a measured user-study result. Screen proofs do not establish print or embroidery minimums.

### Refined SVG audit

- **13 v2 SVGs:** no `FAIL` or `WARN` findings; scores range **99–100/100**.
- A: symbol/reversal 100; horizontal/name 99 (point-count information only).
- B: all four 99 (intentional above-centre symbol or name point-count information).
- C: all five 100, including the 16/24 px test cut.
- Concept logo SVGs contain only self-contained vector paths: no live text, embedded images, filters, masks, external font dependencies or strokes.
- The test boards/overview intentionally use live presentation labels; the evidence board embeds PNG pixel enlargements. Those are not logo masters and are not included in the logo-artwork audit.

Raw audit output and structured scores: `round-2/svg-audit.txt` and `round-2/svg-audit-summary.json`.

## Limited similarity review

See `round-2/shelf-test.png` for equal-visible-size comparisons. White knockout shapes in multicolour references are preserved rather than incorrectly filled black.

- **Bank/fintech sample:** Monzo, N26, Starling Bank, Revolut, Wise, Chase (retrieved Simple Icons SVG snapshots).
- **Craft/familiarity sample:** Scribd, Svelte, Semantic UI, Adyen, Stripe, Square (bundled library).
- **A:** shares the common S initial with Starling/Scribd/Svelte/Semantic UI, but uses a continuous broad, horizontally level contour rather than disconnected hooks, diagonal links or a stock curved S in a carrier. No direct match was observed in this small set; this is not a uniqueness claim.
- **B:** does not use Chase's four angled modules or Square's nested square. Its principal unresolved recognition issue is a possible n/doorway reading, not an observed direct match in the sample.
- **C:** lowercase sans is familiar fintech territory; the joined st head and separate turned l feet provide the differentiating detail. Its shortened st still requires brand association.

These marks were not traced. Other brands' marks remain their respective owners' trademarks. Download provenance and hashes are in `../research/sources.json`.

### Explicit limits / open items

- No trademark-database search, reverse-image search, app-store scan or exhaustive market-wide look-alike search was performed. Recommend professional name/device clearance in the intended jurisdictions before launch.
- The six retrieved category references were not independently verified against current official brand manuals.
- No independent cultural/linguistic or audience testing, physical print proofs, real-device icon tests or production-platform masking checks were performed.
- The green is a candidate: `#214F40`; white contrast calculates to approximately 9.30:1. Print conversions, spot-colour matches and colour approval remain open.
- Final geometry, reverse optical corrections, production minimum sizes, clear space, stacked lockup and final font/usage guidance are deferred to the approved direction's kit.
- No PNG export kit, favicon.ico, webmanifest or installed app assets have been delivered at this stage. PNGs here are tests and presentation renders only.

## Evidence paths

- `../concepts-overview.png` — three-direction checkpoint, with 64/32/24/16 px samples.
- `round-1/concepts-v1.png` — first-iteration comparison.
- `round-{1,2}/evidence.png` — size, reversal, background, app-tile and compact-name proofs.
- `round-{1,2}/actual-size/` — exact-size light/dark crops and 6× nearest-neighbour enlargements.
- `round-{1,2}/transform-tests.png` — mirror, 180° and squint checks.
- `round-{1,2}/shelf-test.png` — limited category/craft comparisons.
- `round-{1,2}/preview.html` — additional generated browser test pages; the custom PNG boards above are the primary inspected evidence.

**Approval gate completed:** the user selected B (“I pick B”), then authorised the offered kit (“go ahead”). The concept files above remain historical explorations; use the refined production artwork in `../final/`, not these test files.
