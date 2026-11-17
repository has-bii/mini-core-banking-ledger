#!/usr/bin/env python3
"""Use the skill overview, adding the brief's 24px samples and C's small cut."""
from copy import deepcopy
from pathlib import Path
import subprocess
import sys
from xml.etree import ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
SKILL = Path('/Users/hasbiyrobbiy/.agents/skills/logo-design')
NS = '{http://www.w3.org/2000/svg}'
ET.register_namespace('', NS[1:-1])
SLUGS = ['a-level-s', 'b-open-ground', 'c-shared-line']
folders = [ROOT / 'concepts' / slug / 'iterations/v2' for slug in SLUGS]
output = ROOT / 'concepts-overview.png'
command = [sys.executable, str(SKILL / 'scripts/concept_sheet.py'),
           *[str(p / 'symbol.svg') for p in folders], '--lockups',
           *[str(p / 'horizontal.svg') for p in folders],
           '--names', 'Level S', 'Open Ground', 'Shared Line',
           '--notes', 'A broad S with level counters and softened turns.',
           'An open arch with a deliberately grounded return.',
           'A joined st head stroke gives the name its own signature.',
           '--title', 'Still Bank — three directions', '--subtitle',
           'Quiet Confidence / unapproved concepts / black first',
           '--recommend', '1', '--greyscale', '--width', '1500', '-o', str(output)]
subprocess.run(command, check=True)
svg_path = output.with_suffix('.svg')
tree = ET.parse(svg_path)
root = tree.getroot()
root.set('role', 'img')
root.set('aria-label', 'Three unapproved Still Bank logo directions with compact name treatments and actual 64, 32, 24 and 16 pixel identifiers')
for index, folder in enumerate(folders):
    target = None
    for group in list(root):
        small = group.find(NS + 'svg')
        if small is not None and any(node.get('id') == f's{index}16-symbol' for node in small.iter()):
            target = group
            break
    if target is None:
        raise RuntimeError(f'Missing small-size slot for concept {index}')
    sixteen = target.find(NS + 'svg')
    old_x = float(sixteen.get('x'))
    sixteen.set('x', str(round(old_x + 50)))
    twenty_four = deepcopy(target)
    nested = twenty_four.find(NS + 'svg')
    nested.set('x', str(round(old_x)))
    nested.set('y', str(float(sixteen.get('y')) - 8))
    nested.set('width', '24')
    nested.set('height', '24')
    for node in nested.iter():
        if node.get('id'):
            node.set('id', node.get('id').replace(f's{index}16', f's{index}24'))
    if index == 2:
        small_path = ET.parse(folder / 'symbol-small-test.svg').getroot().find(NS + 'path').get('d')
        sixteen.find(NS + 'path').set('d', small_path)
        nested.find(NS + 'path').set('d', small_path)
    root.append(twenty_four)
    for label in list(root):
        if label.tag == NS + 'text' and label.text == '16px' and abs(float(label.get('x')) - (old_x + 8)) < .01:
            label.set('x', str(round(old_x + 58)))
            new_label = deepcopy(label)
            new_label.set('x', str(round(old_x + 12)))
            new_label.text = '24px'
            root.append(new_label)
            break
note = ET.SubElement(root, NS + 'text', {
    'x': '1029', 'y': '900', 'font-family': 'Arial,Helvetica,sans-serif',
    'font-size': '14', 'fill': '#6B6B6B',
})
note.text = '16 / 24 px: simplified st cut'
tree.write(svg_path, encoding='unicode', xml_declaration=False)
subprocess.run([sys.executable, str(SKILL / 'scripts/render_png.py'), str(svg_path), '-o', str(output),
                '--width', root.get('width'), '--height', root.get('height'), '--padding', '0', '--backend', 'chrome'], check=True)
print('Overview updated with true 24px samples and C small-cut artwork.')
