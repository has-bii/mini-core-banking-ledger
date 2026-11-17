# Still Bank — concept-stage design strategy

**Archive: stage-one strategy, written before selection.** The user subsequently chose **B — Open Ground** and explicitly authorised kit production. See `../final/APPROVAL.md` and `../final/README.md`. The original concept exploration and recommendation below are retained as history.

## Working brief

- **Brand:** Still Bank, a personal mobile banking app for everyday customers; no tagline.
- **Audience/application:** everyday banking customers encountering an app icon, favicon and compact name treatment.
- **Personality:** Quiet Confidence — calm, precise, trustworthy, approachable.
- **Avoid:** cold technical severity, flash, exclusive/luxury cues and unsupported product or security claims.
- **Constraint:** explore forest green, but establish the forms in black first.
- **Success:** a recognisable small identifier, a readable name, clean vector construction, one-colour/reverse flexibility and meaningful separation from sampled category references.
- **Decision-maker:** the user selects a direction and explicitly authorises any final kit.

### Non-blocking assumptions

1. No existing Still Bank logo equity was supplied; these are new directions, not a redesign.
2. No market or jurisdiction was specified. There is no invented demographic, regulatory status or local banking positioning.
3. Title case is used for A/B. C proposes lowercase styling of the same name, not a change to the business name.
4. “Level”, “ground” and “shared line” are creative rationales, not established business promises.
5. The precise forest green and neutral shown on test boards are candidates, not an approved palette.

## Category research and exclusions

The bundled `payments-fintech` category contains 37 examples. Its summary reported blue as the largest single primary-colour family (11/37); abstract, wordmark and combination marks each account for 8/37. Negative space and containment each appear in 9/37. This is a payment/technology-skewed reference sample, not a statistical survey of consumer banks.

Six more relevant banking/fintech references were retrieved as SVG snapshots from Simple Icons: Monzo, N26, Starling Bank, Revolut, Wise and Chase. Their current official identities were not independently verified. Exact URLs and SHA-256 hashes are in `sources.json`.

**Category clichés deliberately excluded:** shields/padlocks as security promises; upward arrows/growth charts as outcome promises; coins/currency symbols; classical columned bank buildings; globes; interwoven prestige initials; generic blue gradients. Also rejected: a default pause icon, water-ripple illustration and generic leaves used solely to justify green.

**Useful open territory:** horizontal stability, comfortably open spaces, softened but decisive turns and an unhurried readable name.

## Word map

| Starting point | Associations | Visual intersections worth exploring |
| --- | --- | --- |
| Still | settled, steady, uninterrupted, composed | level counters; a continuous S rather than a moving ribbon |
| Calm | breathing room, a resting place, low visual noise | an open centre; a broad arch with a stable foot |
| Precise | alignment, deliberate spacing, simple geometry | level terminals; a shared head stroke; exact 45° k arms |
| Approachable | familiar lettering, soft turns, room to enter | open apertures; humanist proportions; lowercase option |
| Trustworthy | consistency, legibility, grounded weight | repeatable stroke logic, not a shield or financial claim |

## Concept longlist and selection

Scores are design judgements, not measured user research. Each dimension is out of five: clarity / distinction / simplicity / relevance / small-size strength.

| Idea | Type | One-sentence concept | Scores | Total | Decision |
| --- | --- | --- | --- | --- | --- |
| Level S | Letterform | A continuous, broad S holds two level open counters between softened turns. | 5/4/5/5/5 | 24 | Build A |
| Open Ground | Abstract combination | A single soft arch turns one foot inward, holding an open centre without closing it into a lock. | 4/4/5/4/5 | 22 | Build B |
| Shared Line | Wordmark-led | The s and t share one head stroke, giving the name a signature that can be extracted as its app identifier. | 5/4/4/5/4 | 22 | Build C |
| Handrail | Letterform | An S extends one straight terminal into a quiet horizontal rail. | 4/3/4/4/4 | 19 | Too close to A; do not build |
| Common Measure | Abstract | Two unequal lengths meet on one shared baseline. | 3/2/5/4/4 | 18 | Too close to generic bars/pause |
| Inward Frame | Negative-space | Opposing soft corners define an unfilled centre rather than a badge. | 3/3/4/4/4 | 18 | Familiar bracket/device territory |
| Stillpoint | Letterform | A broad i tittle rests above an otherwise open initial. | 3/2/5/3/4 | 17 | Not enough ownership; wrong primary initial |
| SB Joint | Monogram | S and B share a rounded central joint without overlapping outlines. | 3/3/3/3/3 | 15 | Unnecessary decoding and institutional tone |
| Resting Stones | Pictorial | Two broad, flat-topped stones settle into an asymmetric stack. | 4/2/4/2/4 | 16 | Too close to wellness/spa identities |
| Still Water | Pictorial | A shallow reflection is enclosed in a level circular silhouette. | 4/1/4/3/4 | 16 | Literal water/horizon cliché |
| Double-l Seat | Wordmark | The final l pair shares a continuous base. | 3/3/3/4/2 | 15 | Risks reading as u; retain separate l feet instead |
| Folded Note | Pictorial | A single banknote fold becomes a quiet rectangular sign. | 4/2/3/2/3 | 14 | Product literalism and financial cliché |

The selected three test different recognition mechanisms: initial, abstract silhouette and a name-led typographic signature.

## Construction studies and familiarity checks

- **Adyen / Stripe:** custom name treatment, letter spacing and solid outlined type. Study only; no tracing.
- **Square / Framer:** economical geometry, solid-vs-negative construction and app-icon silhouette.
- **Relay:** a continuous turning path and its open internal spacing; its long, dotted S was deliberately not adopted.
- **Scribd / Svelte / Semantic UI:** S-letterform familiarity. These were compared visually with A, including their white knockout artwork.
- Bundled letterform catalogue subjects were also screened for S treatments (including Sublime Text, Snupps, Storybook, Smashing Magazine, Slides, SolidJS and Susy). This subject screening is not equivalent to visually reviewing every result.

### Construction intent, before the SVGs

- **A:** 256-unit square; 192 × 192 visible footprint; opposite 62-unit outer turns; 40-unit end bands with a slightly lighter 38-unit central horizontal. Squared horizontal terminals preserve a strong level rhythm. The name has an intentionally square-ended i dot.
- **B:** 256-unit square; 176 × 184 footprint; a broad curved crown, 40-unit stems, and a single inward return. In v2 the return is 44 units deep, with a 32-unit open bottom gap. The mark sits four units above the canvas centre deliberately.
- **C:** bespoke connected lowercase st built on a 1000-unit type grid; a 108-unit t stem and approximately 94-unit shared head stroke. The standalone identifier is the same st excerpt. The full name keeps two separated l feet. A larger, simpler cut is tested at 16/24 px.

All concept logo SVGs are path-only. The name treatments use Figtree as a source, with custom outline changes; presentation labels are ordinary live text, not logo artwork.

## Provisional colour direction

- **Forest candidate:** `#214F40` (RGB 33, 79, 64).
- **Supporting light neutral candidate:** `#F5F6F3` (RGB 245, 246, 243).
- White and charcoal appear as reversal-test backgrounds, not an approved expanded palette.

Forest/white contrast calculates to approximately **9.30:1** using the sRGB relative-luminance formula. This is a useful value check, not a claim that the logo or product is accessibility-certified. CMYK, spot-colour matching and print proofs are deferred until direction approval.

## Direction rationales and principal risks

### A — Level S · letterform combination · recommended

**Idea:** A broad S holds level counters between softened turns.

- Calm horizontals and generous open counters fit Quiet Confidence without using a pause or water illustration.
- The initial links directly to the name and occupies the app/favIcon square efficiently.
- The mixed square terminals/soft turns balance precision and approachability.

**Principal risk:** standalone S marks are crowded territory. The drawn geometry is custom, but its basic initial is not exclusive; detailed legal/market clearance remains necessary.

### B — Open Ground · abstract combination

**Idea:** One inward-turning foot gives an open arch a grounded, welcoming centre.

- A quiet silhouette and a deliberately open interior suggest room to breathe.
- A single filled contour remains clear in small one-colour tests.

**Principal risk:** the shape can read as a lowercase n or a doorway before the Still Bank association is learned. Architectural/holding-space meanings are metaphors, not security claims.

### C — Shared Line · wordmark-led, with st lettermark excerpt

**Idea:** A shared s–t head stroke makes the name itself the signature.

- Familiar lowercase lettering is approachable and leads with the brand name.
- The st excerpt carries the same distinctive joint into a small app identifier; the l pair stays readable as two letters.

**Principal risk:** st needs name association and can initially suggest another abbreviation (including “saint”). Its two-letter footprint needs the documented simplified cut at 16/24 px.

## Recommendation and approval gate

Recommend **A**, primarily for its direct name association, most efficient near-square app footprint and clearest immediate initial recognition at the smallest tested sizes. B offers a less letter-led alternative; C is strongest when the name itself should dominate.

At the concept checkpoint no direction was selected on the user's behalf, and work stopped for selection and explicit kit approval. The user subsequently selected B and authorised production. Its final masters, lockups, variants, icon exports, presentation and usage guide are now in `../final/`; the other concepts remain archived.
