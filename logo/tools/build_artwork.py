#!/usr/bin/env python3
"""Rebuild unapproved Still Bank concept artwork; never writes production assets.

Requires fonttools and uharfbuzz. Run from the repository root:
  /tmp/still-bank-logo-tools/bin/python logo/tools/build_artwork.py --iteration v1
"""
from __future__ import annotations

import argparse
from math import ceil
from io import BytesIO
from pathlib import Path
from xml.sax.saxutils import escape

import uharfbuzz as hb
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.recordingPen import RecordingPen
from fontTools.pens.transformPen import TransformPen
from fontTools.svgLib.path import parse_path
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont

ROOT = Path(__file__).resolve().parents[1]
FONT = ROOT / "research/type/Figtree[wght].ttf"
BLACK = "#000000"


def number(value):
    return f"{value:.2f}".rstrip("0").rstrip(".") or "0"


def path_data(data, transform=(1, 0, 0, 1, 0, 0)):
    pen = SVGPathPen(None, ntos=number)
    parse_path(data, TransformPen(pen, transform))
    return pen.getCommands()


def element(data, color=BLACK, ident=None):
    identity = f' id="{ident}"' if ident else ""
    return f'<path{identity} fill="{color}" d="{data}"/>'


def svg(title, body, width=256, height=256):
    return (
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {number(width)} {number(height)}" '
        f'width="{number(width)}" height="{number(height)}" role="img" aria-labelledby="title">\n'
        f'  <title id="title">{escape(title)}</title>\n'
        f'  <desc>Unapproved logo concept. Original vector construction; any type is outlined.</desc>\n'
        f'  {body}\n</svg>\n'
    )


class TypeOutliner:
    def __init__(self, weight):
        self.font = instantiateVariableFont(TTFont(FONT), {"wght": weight}, inplace=False)
        self.glyphs = self.font.getGlyphSet()
        self.names = self.font.getGlyphOrder()
        buffer = BytesIO()
        self.font.save(buffer)
        self.hbfont = hb.Font(hb.Face(buffer.getvalue()))
        self.hbfont.scale = (1000, 1000)
        hb.ot_font_set_funcs(self.hbfont)

    def adjusted_a(self, target):
        # Remove the source face's two-unit terminal taper, not its bowl curves.
        recording = RecordingPen()
        self.glyphs["a"].draw(recording)
        previous = None
        for command, points in recording.value:
            if command == "lineTo" and previous:
                px, py = previous
                nx, ny = points[0]
                if px > 400 and abs(nx - px) < 3 and py - ny > 250:
                    points = ((px, ny),)
            getattr(target, command)(*points)
            if points:
                previous = points[-1]

    def draw(self, text, size, x, baseline, *, tracking=0, replacements=None, refine=False):
        buf = hb.Buffer()
        buf.add_str(text)
        buf.guess_segment_properties()
        hb.shape(self.hbfont, buf, {"kern": True, "liga": False})
        cursor = 0
        pen = SVGPathPen(self.glyphs, ntos=number)
        for info, position in zip(buf.glyph_infos, buf.glyph_positions):
            transform = (
                size / 1000, 0, 0, -size / 1000,
                x + (cursor + position.x_offset) * size / 1000,
                baseline - position.y_offset * size / 1000,
            )
            target = TransformPen(pen, transform)
            name = self.names[info.codepoint]
            replacement = (replacements or {}).get(name)
            if refine and name == "a":
                self.adjusted_a(target)
            elif refine and name == "k":
                # Exact 45-degree arms, slightly lighter than the vertical stem.
                parse_path("M63 0V700H171V276L395 500H531L295 264L559 0H415L227 188L171 132V0Z", target)
            elif replacement:
                parse_path(replacement, target)
            else:
                self.glyphs[name].draw(target)
            cursor += (590 if refine and name == "k" else position.x_advance) + tracking
        return pen.getCommands(), cursor * size / 1000


def symbol_a(iteration, reverse=False):
    if iteration == "v1":
        return "M216 32H94C54 32 32 54 32 92C32 128 54 148 94 148H162C176 148 184 154 184 166C184 178 176 184 162 184H40V224H162C202 224 224 202 224 166C224 130 202 108 162 108H94C80 108 72 102 72 90C72 78 80 72 94 72H216Z"
    # Horizontals are 38 vs. the 40-unit end turns; gentle overshoot keeps the
    # squared S from feeling short. Matching 62-unit turns distribute its mass.
    if reverse:
        return "M216 33H94C54.6 33 33 54.6 33 92C33 127.4 54.6 145 94 145H162C176.6 145 185 153.4 185 166C185 178.6 176.6 185 162 185H40V223H162C201.4 223 223 201.4 223 166C223 130.6 201.4 109 162 109H94C79.4 109 71 102.6 71 90C71 77.4 79.4 71 94 71H216Z"
    return "M216 32H94C54 32 32 54 32 92C32 128 54 146 94 146H162C176 146 184 154 184 166C184 178 176 184 162 184H40V224H162C202 224 224 202 224 166C224 130 202 108 162 108H94C80 108 72 102 72 90C72 78 80 72 94 72H216Z"


def symbol_b(iteration, reverse=False):
    if iteration == "v1":
        return "M40 216V112C40 60 72 32 128 32C184 32 216 64 216 116V216H176V120C176 88 160 72 128 72C96 72 80 88 80 120V176H136V216Z"
    # A broad, flatter crown and an inward-turning left foot make a single open
    # contour, not a padlock sitting on a disconnected base.
    if reverse:
        return "M41 215V108C41 55 74.5 33 128 33C183.5 33 215 59 215 112V215H177V116C177 85.5 160.5 71 128 71C95.5 71 79 85.5 79 116V173H144V215Z"
    return "M40 216V108C40 54 74 32 128 32C184 32 216 58 216 112V216H176V116C176 86 160 72 128 72C96 72 80 86 80 116V172H144V216Z"


# Lowercase st shares a level head stroke. The rest of the word remains familiar.
ST_V1 = "M804 500H660V656H552V500H178C66 500 18 448 18 366C18 291 70 256 162 239L262 220C306 212 326 193 326 164C326 124 288 104 222 104C146 104 88 130 36 167L0 74C58 24 134 -12 222 -12C350 -12 430 48 430 150C430 235 376 277 278 296L178 315C134 323 114 337 114 364C114 396 142 406 188 406H552V140C552 67 591 27 632 8C663 -6 704 -12 741 -12C779 -12 813 1 842 24L802 107C781 94 770 88 754 88C691 88 660 107 660 160V406H804Z"
ST_V2 = "M804 500H668V656H560V500H178C68 500 18 450 18 368C18 291 70 256 162 239L262 220C306 212 326 193 326 164C326 124 288 104 222 104C146 104 88 130 36 167L0 74C58 24 134 -12 222 -12C350 -12 430 48 430 150C430 235 376 277 278 296L178 315C134 323 114 337 114 364C114 396 142 406 188 406H560V140C560 67 599 27 640 8C671 -6 712 -12 749 -12C787 -12 821 1 850 24L810 107C789 94 778 88 762 88C699 88 668 107 668 160V406H804Z"


ST_SMALL = "M804 500H668V656H556V500H178C68 500 18 450 18 368C18 291 70 256 162 239L262 220C302 212 322 193 322 164C322 126 288 108 222 108C146 108 88 130 32 164L0 68C58 20 134 -12 222 -12C350 -12 434 48 434 150C434 235 376 281 278 300L178 319C138 327 118 340 118 366C118 394 142 402 188 402H556V140C556 32 620 -12 749 -12H834V90H764C700 90 668 110 668 162V402H804Z"


def symbol_c(iteration, reverse=False, small=False):
    # The lettermark is an exact excerpt, not a separate stock initial, except
    # the documented favicon cut with a larger silhouette and simpler t foot.
    data = ST_SMALL if small else (ST_V1 if iteration == "v1" else ST_V2)
    scale = (224 if small else 208) / 850
    return path_data(data, (scale, 0, 0, -scale, 16 if small else 24, 216 if small else (210 if iteration == "v1" else 208)))


def wordmark_a(outliner, size, x, baseline, refine=False):
    # Level, square-ended i tittle echoes the S's level terminals. It is still
    # recognisably an i; no substituted pictogram or hidden product claim.
    replacements = {
        "i": "M69 0V500H178V0ZM62 603V718H185V603Z",
    }
    return outliner.draw("Still Bank", size, x, baseline, tracking=2, replacements=replacements, refine=refine)


def wordmark_c(outliner, iteration, size, x, baseline):
    scale = size / 1000
    st = ST_V1 if iteration == "v1" else ST_V2
    first = path_data(st, (scale, 0, 0, -scale, x, baseline))
    # The double l stays separated, with two gently turned feet, not a U-ligature.
    ell = "M63 700H171V128Q171 100 197 100H228V0H165Q63 0 63 100Z"
    # Font glyph advance remains 234; keep the feet short enough for open gaps.
    rest, width = outliner.draw("ill bank", size, x + 864 * scale, baseline,
                               tracking=12, replacements={"l": ell}, refine=iteration == "v2")
    return first + rest, 864 * scale + width


def build(iteration):
    types = {"a": TypeOutliner(560), "b": TypeOutliner(580), "c": TypeOutliner(580)}
    concepts = [
        ("a-level-s", "Level S", symbol_a),
        ("b-open-ground", "Open Ground", symbol_b),
        ("c-shared-line", "Shared Line", symbol_c),
    ]
    for slug, name, draw_symbol in concepts:
        key = slug[0]
        folder = ROOT / "concepts" / slug / "iterations" / iteration
        folder.mkdir(parents=True, exist_ok=True)
        data = draw_symbol(iteration)
        (folder / "symbol.svg").write_text(svg(f"Still Bank — {name} — symbol ({iteration})", element(data, ident="symbol")))
        if key == "a":
            draw_word = lambda size, x, baseline: wordmark_a(types[key], size, x, baseline, refine=iteration == "v2")
        elif key == "c":
            draw_word = lambda size, x, baseline: wordmark_c(types[key], iteration, size, x, baseline)
        else:
            draw_word = lambda size, x, baseline: types[key].draw("Still Bank", size, x, baseline, tracking=4, refine=iteration == "v2")
        text, width = draw_word(164, 16, 183)
        (folder / "wordmark.svg").write_text(svg(f"Still Bank — {name} — outlined name ({iteration})", element(text, ident="wordmark"), ceil((width + 32) / 8) * 8))
        if key == "c":
            # Wordmark-led direction: do not repeat the st excerpt before the name.
            full, lock_width = draw_word(164, 24, 183)
            lock_body = element(full, ident="wordmark")
            lock_width += 48
        else:
            compact = path_data(data, (0.76, 0, 0, 0.76, 0, 30.72))
            full, text_width = draw_word(136, 218, 174)
            lock_body = element(compact, ident="symbol") + "\n  " + element(full, ident="wordmark")
            lock_width = 218 + text_width + 20
        (folder / "horizontal.svg").write_text(svg(f"Still Bank — {name} — compact name treatment ({iteration})", lock_body, ceil(lock_width / 8) * 8))
        # Testing artwork only: not an approved colour or export kit.
        reverse_data = draw_symbol(iteration, reverse=True)
        (folder / "symbol-reversed-test.svg").write_text(svg(f"Still Bank — {name} — reversal proof only", element(reverse_data, "#FFFFFF", "symbol")))
        if key == "c" and iteration == "v2":
            (folder / "symbol-small-test.svg").write_text(svg(f"Still Bank — {name} — 16/24 px cut proof only", element(symbol_c(iteration, small=True), ident="symbol")))
        print(f"{slug}/{iteration}: symbol, outlined name, compact treatment, reversal proof")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--iteration", choices=["v1", "v2"], required=True)
    build(parser.parse_args().iteration)
