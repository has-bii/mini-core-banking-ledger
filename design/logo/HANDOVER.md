# Still Bank — Open Ground handover

## Approval and scope

The user chose **B — Open Ground** and authorised production of the logo kit. See [approval record](APPROVAL.md). This is not trademark clearance or regulatory approval.

The production kit lives in repository-root `design/logo/`. Concepts, research snapshots, test evidence, presentations, build scripts and the old ZIP were removed at the user's request. The supplied SVG masters are the source of truth; this folder is not a reproducible build pipeline.

## Final design decisions

- Open arch with an inward-turning foot: a quiet, grounded identifier, not a security promise.
- Separate small-size identifier for 16–24 px; regular artwork from 32 px.
- Optically corrected white artwork, not a simple colour swap.
- Outlined Figtree lettering with custom a/k shapes and fixed spacing; no installed font is needed to display the logos.
- Forest `#214F40`, Mist `#F5F6F3`, and utility black/white. Colour and brand-monochrome use the same single-green artwork.
- Opaque square app/web icon exports, separate maskable artwork, and white-on-Forest favicons.

## Delivered files

- 20 SVG logo masters across symbol, small symbol, wordmark, horizontal and stacked variants.
- 44 transparent logo PNG exports.
- Three icon SVGs, 11 opaque RGB icon PNGs, multi-size ICO and icon-only webmanifest.
- Kit overview and one-page visual usage guide in SVG/PNG, plus PDF usage guide.
- Complete [usage rules](USAGE.md), original Figtree font and SIL OFL licence.
- `FILE-INVENTORY.json`: current relative paths, byte sizes, SHA-256 checksums and raster metadata where available. The inventory excludes itself.

## Historical validation and limitations

Before cleanup, production artwork underwent SVG structural audits, browser-rendered visual checks, native-size identifier checks, PNG transparency/dimension checks, ICO and manifest checks, maskable safe-area checks, and limited visual similarity research. The original test evidence is no longer included; these are historical checks, not newly performed visual validation.

No product features, customer niche, regulatory status, security guarantees or financial outcomes are implied. The kit contains identity assets, not application UI or deployment integration.

## Open production items

1. **Legal:** obtain professional trademark/name clearance. Limited visual research does not establish registrability or exclusivity.
2. **Print:** arrange ICC-aware conversion for the intended stock/press and physical proofing. RGB vectors and approximate CMYK values are not CMYK press masters or verified Pantone matches. Print minimums remain unproofed recommendations.
3. **Platform:** test real devices and app-store requirements; integrate the icon-only webmanifest into the application as needed.
4. **Recognition:** validate the abstract n/doorway reading with independent viewers if launch-critical.
