#!/usr/bin/env python3
"""Build the explicitly authorised Still Bank / Open Ground production assets.

Run with the isolated fonttools/uharfbuzz/Pillow environment from the repo root.
  python logo/tools/build_final.py --vectors
  python logo/tools/build_final.py --png
Only writes under logo/. The original concept files remain unchanged.
"""
from __future__ import annotations

import argparse
from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy
import importlib.util
import json
from math import ceil
from pathlib import Path
import sys
from xml.etree import ElementTree as ET
from xml.sax.saxutils import escape

sys.dont_write_bytecode = True
import uharfbuzz as hb
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
from fontTools.svgLib.path import parse_path

ROOT = Path(__file__).resolve().parents[1]
FINAL = ROOT / 'final'
SKILL = Path('/Users/hasbiyrobbiy/.agents/skills/logo-design')
NS = '{http://www.w3.org/2000/svg}'
GREEN = '#214F40'
MIST = '#F5F6F3'


def load(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


art = load('still_concept_artwork', ROOT / 'tools/build_artwork.py')
sys.path.insert(0, str(SKILL / 'scripts'))
variants = load('still_export_variants', SKILL / 'scripts/export_variants.py')
renderer = load('still_render_png', SKILL / 'scripts/render_png.py')

# v3 preserves B's selected silhouette and inward foot. Both crown springlines
# are now y112; the outer crown is symmetrical rather than accidentally offset.
NORMAL = 'M40 216V112C40 56 74 32 128 32C182 32 216 56 216 112V216H176V116C176 86 160 72 128 72C96 72 80 86 80 116V172H144V216Z'
REVERSED = 'M41 215V112C41 57 74.5 33 128 33C181.5 33 215 57 215 112V215H177V116C177 85.5 160.5 71 128 71C95.5 71 79 85.5 79 116V173H144V215Z'
SMALL = 'M32 224V112C32 56 68 32 128 32C188 32 224 56 224 112V224H176V116C176 90 160 80 128 80C96 80 80 90 80 116V176H144V224Z'
SMALL_REVERSED = 'M33 223V112C33 57 68.5 33 128 33C187.5 33 223 57 223 112V223H177V116C177 89.5 160.5 79 128 79C95.5 79 79 89.5 79 116V177H144V223Z'
K = 'M63 0V700H171V276L395 500H531L295 264L559 0H415L227 188L171 132V0Z'
K_REVERSED = 'M65 0V700H169V280L389 500H521L291 270L561 0H421L229 192L169 132V0Z'


def svg(title, body, width=256, height=256, description=None):
    description = description or 'Still Bank, Open Ground. Self-contained vector logo; all letterforms are outlined.'
    return (f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}" role="img" aria-labelledby="title">\n'
            f'  <title id="title">{escape(title)}</title>\n  <desc>{escape(description)}</desc>\n  {body}\n</svg>\n')


def path(data, fill, ident):
    return f'<path id="{ident}" fill="{fill}" d="{data}"/>'


class FinalType:
    def __init__(self):
        self.regular = art.TypeOutliner(580)
        self.reverse = art.TypeOutliner(560)
        buffer = hb.Buffer()
        buffer.add_str('Still Bank')
        buffer.guess_segment_properties()
        hb.shape(self.regular.hbfont, buffer, {'kern': True, 'liga': False})
        self.slots = list(zip(buffer.glyph_infos, buffer.glyph_positions))

    def draw(self, size, x, baseline, reverse=False):
        source = self.reverse if reverse else self.regular
        pen = SVGPathPen(source.glyphs, ntos=art.number)
        cursor = 0
        for info, position in self.slots:
            glyph_name = self.regular.names[info.codepoint]
            target = TransformPen(pen, (size / 1000, 0, 0, -size / 1000,
                                       x + (cursor + position.x_offset) * size / 1000,
                                       baseline - position.y_offset * size / 1000))
            if glyph_name == 'a':
                source.adjusted_a(target)
            elif glyph_name == 'k':
                parse_path(K_REVERSED if reverse else K, target)
            else:
                source.glyphs[glyph_name].draw(target)
            # Identical advance/kerning positions in both optical weights.
            cursor += (590 if glyph_name == 'k' else position.x_advance) + 4
        return pen.getCommands(), cursor * size / 1000


def vectors():
    folder = FINAL / 'svg'
    icons = FINAL / 'icons'
    folder.mkdir(parents=True, exist_ok=True)
    icons.mkdir(parents=True, exist_ok=True)
    source = FinalType()
    for reverse in [False, True]:
        fill = '#FFFFFF' if reverse else GREEN
        mark = REVERSED if reverse else NORMAL
        small = SMALL_REVERSED if reverse else SMALL
        word, word_width = source.draw(164, 16, 183, reverse)
        horizontal_word, horizontal_width = source.draw(136, 218, 174, reverse)
        stack_word, stack_width = source.draw(80, 0, 224, reverse)
        # Centre using the reference advance rather than changing inter-letter spacing.
        stack_word = art.path_data(stack_word, (1, 0, 0, 1, (384 - stack_width) / 2, 0))
        horizontal_mark = art.path_data(mark, (.76, 0, 0, .76, 0, 30.72))
        stack_mark = art.path_data(mark, (.6, 0, 0, .6, 115.2, 4.8))
        bodies = {
            'symbol': (path(mark, fill, 'symbol'), 256),
            'symbol-small': (path(small, fill, 'symbol'), 256),
            'wordmark': (path(word, fill, 'wordmark'), ceil((word_width + 32) / 8) * 8),
            'horizontal': (path(horizontal_mark, fill, 'symbol') + '\n  ' + path(horizontal_word, fill, 'wordmark'), ceil((218 + horizontal_width + 20) / 8) * 8),
            'stacked': (path(stack_mark, fill, 'symbol') + '\n  ' + path(stack_word, fill, 'wordmark'), 384),
        }
        for form, (body, width) in bodies.items():
            suffix = 'white' if reverse else 'color'
            (folder / f'still-bank-{form}-{suffix}.svg').write_text(svg(f'Still Bank — {form} — {suffix}', body, width))
            if not reverse:
                for version, color in [('black', '#000000'), ('mono-214f40', GREEN)]:
                    # Use the skill's recolouring helper, retaining our own clean
                    # titles/metadata. White variants above have distinct geometry.
                    recolored = variants.recolor(body, color)
                    # recolor expects a whole SVG; it also works on our explicit path fills.
                    (folder / f'still-bank-{form}-{version}.svg').write_text(svg(f'Still Bank — {form} — {version}', recolored, width))

    # Production app icons are opaque, full-bleed squares. Platform masks are not
    # baked into app-store / Apple touch artwork.
    icon_mark = art.path_data(REVERSED, (1, 0, 0, 1, 3, 0))
    icon_body = f'<path fill="{GREEN}" d="M0 0H256V256H0Z"/>' + path(icon_mark, '#FFFFFF', 'symbol')
    (icons / 'app-icon.svg').write_text(svg('Still Bank — app icon', icon_body))
    mask_mark = art.path_data(REVERSED, (.78, 0, 0, .78, 128 * .22 + 3 * .78, 128 * .22))
    mask_body = f'<path fill="{GREEN}" d="M0 0H256V256H0Z"/>' + path(mask_mark, '#FFFFFF', 'symbol')
    (icons / 'maskable-icon.svg').write_text(svg('Still Bank — maskable app icon', mask_body))
    fav_body = f'<path fill="{GREEN}" d="M0 0H256V256H0Z"/>' + path(SMALL_REVERSED, '#FFFFFF', 'symbol')
    (icons / 'favicon.svg').write_text(svg('Still Bank — favicon', fav_body))
    manifest = {
        'name': 'Still Bank', 'short_name': 'Still Bank', 'theme_color': GREEN, 'background_color': GREEN,
        'icons': [
            {'src': 'icon-192.png', 'sizes': '192x192', 'type': 'image/png', 'purpose': 'any'},
            {'src': 'icon-512.png', 'sizes': '512x512', 'type': 'image/png', 'purpose': 'any'},
            {'src': 'maskable-512.png', 'sizes': '512x512', 'type': 'image/png', 'purpose': 'maskable'},
        ],
    }
    (icons / 'site.webmanifest').write_text(json.dumps(manifest, indent=2) + '\n')
    refinement = ROOT / 'refinements/b-open-ground/v3'
    refinement.mkdir(parents=True, exist_ok=True)
    for form in ['symbol', 'symbol-small', 'wordmark', 'horizontal', 'stacked']:
        # A retained production-refinement snapshot, separate from the final kit.
        (refinement / f'{form}.svg').write_bytes((folder / f'still-bank-{form}-black.svg').read_bytes())
    print('Built 20 outlined logo SVGs, 3 square icon SVGs, icon-only webmanifest and v3 refinement snapshots.')


def render_job(job):
    source, destination, width, height = job
    destination.parent.mkdir(parents=True, exist_ok=True)
    renderer.render(str(source), str(destination), width, height, padding=0, backend='chrome')
    return destination


def pngs():
    from PIL import Image
    jobs = []
    for form, sizes in [('symbol', [512, 1024]), ('wordmark', [600, 1200]),
                        ('horizontal', [600, 1200]), ('stacked', [512, 1024])]:
        for version in ['color', 'black', 'white', 'mono-214f40']:
            source = FINAL / 'svg' / f'still-bank-{form}-{version}.svg'
            vb = list(map(float, ET.parse(source).getroot().get('viewBox').split()))
            for width in sizes:
                height = round(width * vb[3] / vb[2])
                jobs.append((source, FINAL / 'png' / f'{source.stem}-{width}.png', width, height))
    icon_jobs = [
        ('app-icon.svg', 'app-icon-1024.png', 1024), ('app-icon.svg', 'app-icon-512.png', 512),
        ('app-icon.svg', 'icon-192.png', 192), ('app-icon.svg', 'icon-512.png', 512),
        ('app-icon.svg', 'apple-touch-icon.png', 180),
        ('maskable-icon.svg', 'maskable-512.png', 512),
    ] + [('favicon.svg', f'favicon-{s}.png', s) for s in [16, 24, 32, 48, 64]]
    for source, target, size in icon_jobs:
        jobs.append((FINAL / 'icons' / source, FINAL / 'icons' / target, size, size))
    with ThreadPoolExecutor(max_workers=3) as executor:
        results = list(executor.map(render_job, jobs))
    # One transparent sprite renders small-symbol exports at native integer
    # positions and exact 1× sizes. No downsampling from a large PNG is involved.
    nodes = []
    crops = []
    for row, version in enumerate(['color', 'black', 'white']):
        x = 0
        for size in [16, 24, 32, 64]:
            source = ET.parse(FINAL / 'svg' / f'still-bank-symbol-small-{version}.svg').getroot()
            source.attrib.update({'x': str(x), 'y': str(row * 80), 'width': str(size), 'height': str(size)})
            nodes.append(ET.tostring(source, encoding='unicode'))
            crops.append((version, size, x, row * 80))
            x += 80
    test = FINAL / 'tests'
    test.mkdir(parents=True, exist_ok=True)
    sprite_svg = test / 'small-symbol-sprite.svg'
    sprite_svg.write_text(svg('Still Bank small-symbol raster source (test sheet)', ''.join(nodes), 320, 240))
    sprite_png = test / 'small-symbol-sprite.png'
    render_job((sprite_svg, sprite_png, 320, 240))
    with Image.open(sprite_png) as sprite:
        if sprite.size != (320, 240):
            raise RuntimeError('Small-symbol sprite was unexpectedly scaled')
        for version, size, x, y in crops:
            destination = FINAL / 'png' / f'still-bank-symbol-small-{version}-{size}.png'
            sprite.crop((x, y, x + size, y + size)).save(destination)
            results.append(destination)
    # App-store/touch/web tiles must have no alpha channel, not merely opaque
    # pixels inside an RGBA container. Transparent logo PNGs remain RGBA.
    for _, target, _ in icon_jobs:
        file = FINAL / 'icons' / target
        with Image.open(file) as image:
            if image.convert('RGBA').getchannel('A').getextrema() != (255, 255):
                raise RuntimeError(f'Unexpected transparency in square icon: {file}')
            opaque = image.convert('RGB')
        opaque.save(file)
    renderer.write_ico([str(FINAL / 'icons' / f'favicon-{s}.png') for s in [16, 32, 48, 64]], str(FINAL / 'icons/favicon.ico'))
    print(f'Rendered {len(results)} production PNGs, native small-size exports and a 4-size favicon.ico.')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--vectors', action='store_true')
    parser.add_argument('--png', action='store_true')
    args = parser.parse_args()
    if not (args.vectors or args.png):
        parser.error('Choose --vectors and/or --png')
    if args.vectors:
        vectors()
    if args.png:
        pngs()
