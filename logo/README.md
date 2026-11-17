# Still Bank — logo identity

**B — Open Ground selected; full-kit production explicitly authorised.** Production files are in `final/`. Earlier concepts remain archived separately.

## Start here

- [Download the complete kit](still-bank-logo-kit.zip) · [SHA-256](still-bank-logo-kit.zip.sha256)
- [Final kit overview](final/still-bank-logo-kit.png) · [Master files and inventory](final/README.md)
- [Usage guide (PDF)](final/guidelines/still-bank-usage-guide.pdf) · [Complete usage notes](final/USAGE.md)
- [Identity mockups](final/presentation/slides/slide-04.png) · [Final tests / limitations](final/HANDOVER.md)
- [Selection and production approval](final/APPROVAL.md)

## Concept archive

- [Three-direction overview](concepts-overview.png) — marks, compact name treatments and actual 64/32/24/16 px samples; greyscale first.
- [Rendered test evidence](tests/round-2/evidence.png) — small sizes, reversals, provisional forest green, isolated app tiles and names at 256 px wide.
- [Limited visual similarity comparison](tests/round-2/shelf-test.png).
- [Test report and honest limitations](tests/TEST-REPORT.md).
- [Brief interpretation, research, concept longlist and rationale](research/design-strategy.md).

## The three directions

| Direction | Type / idea | Main trade-off |
| --- | --- | --- |
| **A — Level S · original concept recommendation** | Custom S letterform: level open counters and softened turns. Directly connects the app identifier to Still. | S initials are crowded territory; detailed clearance is essential. |
| **B — Open Ground · selected** | Abstract arch: an inward-turning foot grounds an otherwise open centre. | May initially read as n or a doorway. |
| **C — Shared Line** | Name-led design: a custom s–t head stroke becomes the standalone st identifier. | The abbreviated icon needs association; a simplified 16/24 px cut is documented. |

The concept-stage recommendation was A; the user selected **B** and explicitly authorised its kit. That choice is respected in the final production identity.

Final kit palette: Forest **#214F40**, Mist **#F5F6F3**, with utility black/white. See `final/USAGE.md` for background pairs and print caveats.

## Artwork and evidence inventory

```text
final/                                Production kit, guides, presentation, tests and inventory
still-bank-logo-kit.zip               Packaged final kit only
still-bank-logo-kit.zip.sha256        Archive checksum
refinements/b-open-ground/v3/         Retained final-geometry snapshot
concepts-overview.{png,svg}              Checkpoint presentation, not a master
concepts/
  a-level-s/iterations/{v1,v2}/          Symbol, name, compact treatment, reversal proof
  b-open-ground/iterations/{v1,v2}/      Same
  c-shared-line/iterations/{v1,v2}/      Same; v2 also has a small-size test cut
tests/
  TEST-REPORT.md                       What was tested, results and limitations
  round-{1,2}/                         Proof PNG/SVG, HTML preview, audit logs,
                                       transforms, similarity shelf, exact pixel crops
research/
  design-strategy.md                   Strategy, longlist, construction, risks
  sources.json                        Reference/font URLs and SHA-256 hashes
  references/                         Six external marks, comparison only
  type/                               Unmodified Figtree source and OFL licence
tools/                                Concept / final build, validation and packaging scripts
```

Logo-artwork SVGs are self-contained outlined paths: no font installation is needed to view them. Test/presentation sheets may have live text labels and embedded pixel samples; they are not logo masters.

### Typography / third-party materials

Figtree, copyright 2022 The Figtree Project Authors, is used as an outline source at weights 560/580 under the **SIL Open Font License 1.1**. Its source and complete licence are in `research/type/`. The custom st, S and abstract mark are drawn constructions; name treatments also modify i, a, k and/or l shapes and spacing. No modified font binary is distributed.

Reference logos belong to their respective trademark owners and are kept only for research. Limited visual comparison is **not trademark clearance**. Obtain professional clearance before launch.

## Rebuilding the artwork

Requires Python 3, `fonttools`, `uharfbuzz`, `pillow`, Chrome/Chromium, and the installed logo-design skill. The scripts currently point to the local skill path; update `SKILL` when moving machines.

The build used an isolated environment at `/tmp/still-bank-logo-tools` (not committed). To recreate it:

```sh
python3 -m venv /tmp/still-bank-logo-tools
/tmp/still-bank-logo-tools/bin/python -m pip install fonttools uharfbuzz pillow
/tmp/still-bank-logo-tools/bin/python logo/tools/build_artwork.py --iteration v1
/tmp/still-bank-logo-tools/bin/python logo/tools/build_artwork.py --iteration v2
/tmp/still-bank-logo-tools/bin/python logo/tools/build_evidence.py --iteration v1
/tmp/still-bank-logo-tools/bin/python logo/tools/build_evidence.py --iteration v2
/tmp/still-bank-logo-tools/bin/python logo/tools/build_overview.py
/tmp/still-bank-logo-tools/bin/python logo/tools/build_final.py --vectors --png
/tmp/still-bank-logo-tools/bin/python logo/tools/build_final_review.py
/tmp/still-bank-logo-tools/bin/python logo/tools/build_final_presentation.py
/tmp/still-bank-logo-tools/bin/python logo/tools/validate_final.py
/tmp/still-bank-logo-tools/bin/python logo/tools/package_final.py
```

Run the skill's `svg_audit.py` and `preview_sheet.py` again after any artwork change. Browser/font-tool versions may affect rasterisation; inspect the rebuilt PNGs rather than assuming identical output.

## Completed approval gate

The user chose B (“I pick B”), then explicitly authorised the full kit (“go ahead”). Refined masters, variants, icons, PNGs, isolated identity mockups and usage guidance are delivered under `final/` and in the ZIP. No UI screens, frontend/application code or token system were created. Legal clearance, real-device checks and physical print proofs remain open as documented in `final/HANDOVER.md`.
