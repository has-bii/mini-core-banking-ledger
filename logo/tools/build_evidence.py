#!/usr/bin/env python3
"""Render concept-only proofs, actual-size samples, transformation and shelf tests.

Requires Pillow and fonttools. Labels may use live text; logo artwork never does.
These test boards are not production masters, icon exports, or a usage guide.
"""
from __future__ import annotations

import argparse
import base64
import importlib.util
import json
import subprocess
import sys
from pathlib import Path
from xml.etree import ElementTree as ET
from xml.sax.saxutils import escape

from PIL import Image
from fontTools.pens.boundsPen import BoundsPen
from fontTools.svgLib.path import parse_path

ROOT = Path(__file__).resolve().parents[1]
SKILL = Path("/Users/hasbiyrobbiy/.agents/skills/logo-design")
RENDER = SKILL / "scripts/render_png.py"
LIBRARY = SKILL / "assets/library/svg"
SLUGS = ["a-level-s", "b-open-ground", "c-shared-line"]
NAMES = ["A / Level S", "B / Open Ground", "C / Shared Line"]
NS = "{http://www.w3.org/2000/svg}"
GREEN = "#214F40"
DARK = "#171B19"


def text(x, y, label, size=18, fill="#343B36", weight="400"):
    return f'<text x="{x}" y="{y}" font-family="Arial,Helvetica,sans-serif" font-size="{size}" font-weight="{weight}" fill="{fill}">{escape(label)}</text>'


def rect(x, y, width, height, fill, radius=0):
    return f'<rect x="{x}" y="{y}" width="{width}" height="{height}" rx="{radius}" fill="{fill}"/>'


def document(width, height, title, body):
    return f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} {height}" width="{width}" height="{height}"><title>{escape(title)}</title>{rect(0, 0, width, height, "#FFFFFF")}{body}</svg>\n'


def bounds_of(root):
    bounds = []
    # All comparison marks used here are path-only and have no geometry transforms.
    for child in root:
        nodes = child.iter() if child.tag != NS + "defs" else []
        for node in nodes:
            if node.tag == NS + "path":
                pen = BoundsPen(None)
                parse_path(node.attrib["d"], pen)
                if pen.bounds:
                    bounds.append(pen.bounds)
    if not bounds:
        return root.attrib.get("viewBox", "0 0 256 256")
    x0 = min(b[0] for b in bounds)
    y0 = min(b[1] for b in bounds)
    x1 = max(b[2] for b in bounds)
    y1 = max(b[3] for b in bounds)
    return f"{x0} {y0} {x1 - x0} {y1 - y0}"


def embed(path, x, y, width, height=None, color=None, normalize=False):
    root = ET.parse(path).getroot()
    viewbox = bounds_of(root) if normalize else root.attrib.get("viewBox", "0 0 256 256")
    children = []
    for node in root:
        if node.tag in {NS + "title", NS + "desc", NS + "defs"}:
            continue
        for item in node.iter():
            item.attrib.pop("id", None)
            item.attrib.pop("clip-path", None)
            if color and item.tag in {NS + "path", NS + "rect", NS + "circle", NS + "polygon"}:
                original_fill = item.attrib.get("fill", "").lower()
                # Keep opaque white knockout artwork in reference logos; forcing
                # it to black would hide the S in Svelte / Semantic UI.
                item.set("fill", "#FFFFFF" if color == "#000000" and original_fill in {"white", "#fff", "#ffffff"} else color)
        children.append(ET.tostring(node, encoding="unicode"))
    return f'<svg x="{x}" y="{y}" width="{width}" height="{height or width}" viewBox="{viewbox}" preserveAspectRatio="xMidYMid meet">{"".join(children)}</svg>'


def render(source, target, width, height):
    result = subprocess.run([sys.executable, str(RENDER), str(source), "-o", str(target),
                             "--width", str(width), "--height", str(height), "--padding", "0", "--backend", "chrome"],
                            capture_output=True, text=True)
    if result.returncode:
        raise RuntimeError(result.stdout + result.stderr)
    print(f"Rendered {target.relative_to(ROOT)} ({width}×{height})")


def nearest_sample(path, x, y):
    # An enlarged PNG is deliberately used only on the proof sheet. The SVG
    # masters themselves have no images. Nearest-neighbour keeps pixels honest.
    with Image.open(path) as img:
        img.resize((96, 96), Image.Resampling.NEAREST).save(path.with_name(path.stem + "-6x.png"))
    data = base64.b64encode(path.with_name(path.stem + "-6x.png").read_bytes()).decode()
    return f'<image x="{x}" y="{y}" width="96" height="96" href="data:image/png;base64,{data}"/>'


def component_count(image, reverse=False):
    gray = image.convert("L")
    points = {(x, y) for y in range(gray.height) for x in range(gray.width)
              if (gray.getpixel((x, y)) > 128 if reverse else gray.getpixel((x, y)) < 128)}
    count = 0
    while points:
        count += 1
        todo = [points.pop()]
        while todo:
            x, y = todo.pop()
            for other in [(x-1, y), (x+1, y), (x, y-1), (x, y+1)]:
                if other in points:
                    points.remove(other)
                    todo.append(other)
    return count


def build(iteration):
    round_number = 1 if iteration == "v1" else 2
    folder = ROOT / "tests" / f"round-{round_number}"
    samples = folder / "actual-size"
    samples.mkdir(parents=True, exist_ok=True)
    width, height = 1800, 1280
    pieces = [text(40, 52, f"Still Bank / visual proofs / {iteration}", 34, weight="700"),
              text(40, 88, "Actual pixel sizes, reversals, app-icon framing and compact name treatments. Unapproved exploration.", 20)]
    coords = []
    for index, (slug, name) in enumerate(zip(SLUGS, NAMES)):
        y = 120 + index * 340
        concept = ROOT / "concepts" / slug / "iterations" / iteration
        regular = concept / "symbol.svg"
        reversal = concept / "symbol-reversed-test.svg"
        small = concept / "symbol-small-test.svg"
        pieces += [rect(24, y, 1752, 320, "#F5F6F3", 16), text(48, y + 34, name, 23, weight="700"),
                   embed(regular, 58, y + 72, 160), text(48, y + 276, ["Letterform + name", "Abstract + name", "Wordmark + st excerpt"][index], 17),
                   text(294, y + 34, "Actual size / black", 18, weight="700"),
                   text(550, y + 34, "16 px / 6× pixels", 17, weight="700"),
                   text(705, y + 34, "White on charcoal", 18, weight="700"),
                   text(906, y + 34, "Forest-green proof", 18, weight="700"),
                   text(1110, y + 34, "App-tile proof", 18, weight="700"),
                   text(1260, y + 34, "Compact treatment / 256 px wide", 18, weight="700")]
        pieces += [rect(280, y + 205, 235, 100, DARK, 10), rect(698, y + 65, 178, 178, DARK, 12),
                   rect(899, y + 65, 178, 178, GREEN, 12), embed(reversal, 719, y + 86, 136),
                   embed(reversal, 920, y + 86, 136), text(700, y + 279, ("Weight compensated" if index < 2 else "Base weight retained") if round_number == 2 else "Initial reverse geometry", 15),
                   text(900, y + 279, "Candidate #214F40", 16)]
        for size, x in zip([16, 24, 32, 64], [294, 336, 386, 442]):
            source = small if round_number == 2 and index == 2 and size <= 24 else regular
            pieces += [embed(source, x, y + 104, size), text(x, y + 190, str(size), 15),
                       embed(reversal if index < 2 else source, x, y + 228, size, color="#FFFFFF")]
            coords.append({"concept": slug, "size": size, "x": x, "y": y + 104,
                           "reverse_y": y + 228, "small_cut": source == small})
        pieces += [rect(1110, y + 74, 96, 96, GREEN, 22), embed(reversal, 1118, y + 82, 80),
                   rect(1110, y + 214, 32, 32, GREEN, 7),
                   embed(small if small.exists() else regular, 1112, y + 216, 28, color="#FFFFFF"),
                   text(1150, y + 237, "32 px tile", 15),
                   embed(concept / "horizontal.svg", 1260, y + 74, 256, 90),
                   text(1260, y + 192, "Name only / 256 px wide", 17),
                   embed(concept / "wordmark.svg", 1260, y + 204, 256, 90)]
        if small.exists():
            pieces.append(text(294, y + 69, "16 / 24: simplified cut", 15))
    pieces += [text(40, 1190, "Pixel enlargements are nearest-neighbour crops of the real 16 px raster above; they are not enlarged vector impostors.", 18),
               text(40, 1220, "Green is provisional. These are isolated identity tests, not application screens or approved production assets.", 18)]
    base = folder / "evidence-base.svg"
    base.write_text(document(width, height, "Still Bank concept test proofs", "".join(pieces)))
    base_png = folder / "evidence-base.png"
    render(base, base_png, width, height)
    metrics = []
    with Image.open(base_png) as image:
        if image.size != (width, height):
            raise RuntimeError(f"Proof was rescaled unexpectedly: {image.size}")
        for record in coords:
            size = record["size"]
            normal = image.crop((record["x"], record["y"], record["x"] + size, record["y"] + size))
            reverse = image.crop((record["x"], record["reverse_y"], record["x"] + size, record["reverse_y"] + size))
            normal.save(samples / f'{record["concept"]}-{size}.png')
            reverse.save(samples / f'{record["concept"]}-{size}-reversed.png')
            metrics.append({"concept": record["concept"], "size_px": size, "small_cut": record["small_cut"],
                            "dark_ink_components_128_threshold": component_count(normal),
                            "white_ink_components_128_threshold": component_count(reverse, reverse=True)})
    for index, slug in enumerate(SLUGS):
        y = 120 + index * 340
        pieces += [nearest_sample(samples / f"{slug}-16.png", 550, y + 60),
                   nearest_sample(samples / f"{slug}-16-reversed.png", 550, y + 196)]
    board = folder / "evidence.svg"
    board.write_text(document(width, height, "Still Bank visual test evidence", "".join(pieces)))
    render(board, folder / "evidence.png", width, height)
    (folder / "pixel-metrics.json").write_text(json.dumps({
        "method": "Chrome, 1× CSS pixel scale, opaque light/dark crops; 4-connected ink at threshold 128. Components are diagnostic, not a readability score.",
        "samples": metrics,
    }, indent=2) + "\n")

    # Rotation / mirror / squint: only identity artwork is transformed.
    transform_parts = [text(40, 52, f"Still Bank / silhouette and transform tests / {iteration}", 32, weight="700")]
    for column, label in enumerate(["Original", "Mirrored", "180° rotation", "Soft squint"]):
        transform_parts.append(text(286 + column * 270, 100, label, 21, weight="700"))
    transform_parts.append('<defs><filter id="squint"><feGaussianBlur stdDeviation="3"/></filter></defs>')
    for index, (slug, name) in enumerate(zip(SLUGS, NAMES)):
        y = 128 + index * 190
        symbol = ROOT / "concepts" / slug / "iterations" / iteration / "symbol.svg"
        transform_parts += [text(40, y + 86, name, 21), embed(symbol, 280, y, 160),
                            f'<g transform="translate(1260 0) scale(-1 1)">{embed(symbol, 550, y, 160)}</g>',
                            f'<g transform="translate(1800 {2*y+160}) rotate(180)">{embed(symbol, 820, y, 160)}</g>',
                            f'<g filter="url(#squint)">{embed(symbol, 1090, y, 160)}</g>']
    transform_parts.append(text(40, 755, "Visual review only. No independent audience/cultural testing or legal clearance is implied.", 18))
    transform_svg = folder / "transform-tests.svg"
    transform_svg.write_text(document(1400, 800, "Still Bank transform tests", "".join(transform_parts)))
    render(transform_svg, folder / "transform-tests.png", 1400, 800)

    # Shelf and craft peers are presented at the same visible (not padded-canvas) size.
    shelf = [text(40, 52, f"Still Bank / visual similarity and craft comparisons / {iteration}", 30, weight="700"),
             text(40, 88, "Monochrome normalisation compares form, not colour. Reference marks are trademarks of their respective owners.", 18)]
    for index, (slug, name) in enumerate(zip(SLUGS, NAMES)):
        x = 90 + index * 420
        source = ROOT / "concepts" / slug / "iterations" / iteration / "symbol.svg"
        shelf += [rect(x - 30, 125, 300, 190, "#F5F6F3", 14), embed(source, x + 44, 141, 112, normalize=True), text(x, 290, name, 21, weight="700")]
    reference_names = ["monzo", "n26", "starlingbank", "revolut", "wise", "chase"]
    shelf.append(text(40, 365, "Relevant banking / fintech references", 22, weight="700"))
    for index, name in enumerate(reference_names):
        x = 60 + index * 220
        shelf += [embed(ROOT / "research/references" / f"{name}.svg", x, 390, 112, color="#000000", normalize=True),
                  text(x, 535, name, 18)]
    shelf.append(text(40, 600, "Letterform familiarity / typography / geometric craft peers", 22, weight="700"))
    peers = ["scribd-icon.svg", "svelte-icon.svg", "semantic-ui.svg", "adyen.svg", "stripe.svg", "square.svg"]
    for index, name in enumerate(peers):
        x = 60 + index * 220
        shelf += [embed(LIBRARY / name, x, 630, 112, color="#000000", normalize=True), text(x, 775, name.removesuffix(".svg"), 17)]
    shelf += [text(40, 850, "Scope: six retrieved category marks + six bundled craft/familiarity peers; not an exhaustive look-alike search.", 18),
              text(40, 880, "No trademark-database, reverse-image, app-store or market-wide search has been performed. Obtain professional clearance before launch.", 17)]
    shelf_svg = folder / "shelf-test.svg"
    shelf_svg.write_text(document(1400, 920, "Still Bank limited visual similarity comparison", "".join(shelf)))
    render(shelf_svg, folder / "shelf-test.png", 1400, 920)
    print(f"Saved 24 true-size light/dark crops, 6 pixel enlargements and pixel metrics in tests/round-{round_number}/")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--iteration", choices=["v1", "v2"], required=True)
    build(parser.parse_args().iteration)
