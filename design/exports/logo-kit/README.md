# MiniLedger logo kit

Approved direction: **Geometric Balanced M**. Two mirrored angular halves form one M with a continuous central gap.

## Start here

- `svg/miniledger-horizontal-plum.svg`: primary logo for light surfaces.
- `svg/miniledger-symbol-plum.svg`: standalone symbol at 32 px and above.
- `svg/miniledger-symbol-small-plum.svg`: wider center gap for 16–24 px use.
- `svg/*-white.svg`: reversed logos for dark surfaces.
- `png/`: transparent raster versions; symbols at 512 px, horizontal logos at 2000 px wide, stacked logos at 1280 px wide, and a 1024 px app icon.
- `web/`: favicon.ico, SVG favicon, 16/32/48 px PNG favicons, Apple touch icon, 192/512 px PWA icons, maskable icon, manifest and HTML head snippet.
- `USAGE.md`: clear space, minimum sizes, color and typography.
- `logo-kit.pdf` and `logo-kit.png`: visual reference board.
- `presentation/index.html`: standalone identity presentation with six illustrative application mockups.

The 19 SVG masters include horizontal, stacked, symbol-only and wordmark-only lockups in plum, ink, black and white, plus three small-size symbols. All logo lettering is outlined; installing a font is unnecessary to display the supplied artwork.

Copy the contents of `web/` into a site's public root and use `web/head-snippet.html`, or adjust the paths for the chosen asset directory. The manifest is a starting asset configuration; set the deployment's actual `start_url` and scope as needed.

The editable canvas is `design.pen`, board **09 · MiniLedger — Logo kit**. Four reusable production assets are available in its component library. Their source frames are retained inside the hidden **Logo kit · outlined export masters** frame, which can be enabled to inspect or edit them. Existing application screens are unchanged.

The phrase “Balanced by design.” is presentation copy, not part of the logo lockups. Mockups illustrate identity use; they do not represent shipped MiniLedger features or payment products.
