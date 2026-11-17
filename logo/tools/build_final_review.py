#!/usr/bin/env python3
"""Render final identity preview and independent size/background/spacing proofs."""
from __future__ import annotations
import base64
import importlib.util
import json
from pathlib import Path
import sys
from xml.etree import ElementTree as ET

sys.dont_write_bytecode = True
from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
FINAL = ROOT / 'final'
SKILL = Path('/Users/hasbiyrobbiy/.agents/skills/logo-design')
spec = importlib.util.spec_from_file_location('still_evidence_helpers', ROOT / 'tools/build_evidence.py')
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)
text, rect, embed, render = helper.text, helper.rect, helper.embed, helper.render
GREEN, MIST, DARK = '#214F40', '#F5F6F3', '#171B19'


def logo(form='symbol', version='color'):
    return FINAL / 'svg' / f'still-bank-{form}-{version}.svg'


def doc(width, height, title, body, background=MIST):
    return f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}" role="img" aria-label="{title}"><title>{title}</title>{rect(0,0,width,height,background)}{"".join(body)}</svg>\n'


def save(name, body, width, height, background=MIST):
    path = FINAL / name
    path.parent.mkdir(parents=True, exist_ok=True)
    path.with_suffix('.svg').write_text(doc(width,height,'Still Bank — ' + path.stem,body,background))
    render(path.with_suffix('.svg'), path.with_suffix('.png'), width, height)
    return path.with_suffix('.png')


def image_uri(path, x, y, width, height=None):
    uri = base64.b64encode(Path(path).read_bytes()).decode()
    return f'<image x="{x}" y="{y}" width="{width}" height="{height or width}" href="data:image/png;base64,{uri}"/>'


def build():
    icon = FINAL / 'icons/app-icon.svg'
    # One concise, shareable final kit overview. Rounded corners here illustrate
    # an OS mask; the production app icon itself stays an opaque square.
    overview = [text(64,68,'Still Bank',34,GREEN,'700'), text(1130,64,'OPEN GROUND / LOGO KIT',18,GREEN),
                rect(48,110,1504,410,'#FFFFFF',24), embed(logo('horizontal'),88,156,1040,320),
                '<defs><clipPath id="app"><rect x="1270" y="214" width="192" height="192" rx="44"/></clipPath></defs>',
                f'<g clip-path="url(#app)">{embed(icon,1270,214,192)}</g>', text(1318,446,'App icon',19,GREEN),
                rect(48,550,580,280,'#FFFFFF',22), text(76,588,'STACKED / NAME ONLY',16,GREEN,'700'),
                embed(logo('stacked'),72,604,300,200), embed(logo('wordmark'),366,640,234,120),
                rect(656,550,896,280,GREEN,22), text(688,588,'REVERSED / OPTICALLY LIGHTENED',16,'#FFFFFF','700'),
                embed(logo('horizontal','white'),694,607,640,196), embed(logo('symbol','white'),1360,642,140),
                rect(48,858,580,150,'#FFFFFF',22), rect(76,888,82,82,GREEN,12),
                text(178,913,'Forest',20,GREEN,'700'), text(178,944,'#214F40',19),
                rect(338,888,82,82,MIST,12), text(440,913,'Mist',20,GREEN,'700'), text(440,944,'#F5F6F3',19),
                rect(656,858,896,150,'#FFFFFF',22), text(688,894,'SMALL IDENTIFIER / ACTUAL PIXEL SIZES',16,GREEN,'700')]
    for size,x in zip([16,24,32,64],[706,770,846,942]):
        overview += [embed(logo('symbol-small' if size<=24 else 'symbol'),x,920,size), text(x,994,f'{size}px',13)]
    overview += [text(1080,936,'16–24 px: small cut',19,GREEN), text(1080,968,'32 px and up: regular',18)]
    save('still-bank-logo-kit',overview,1600,1040)

    proofs = [text(40,54,'Still Bank / final production checks',32,GREEN,'700'),
              text(40,86,'Selected B / refined crown / pixel cut / light, dark and forest backgrounds / no application screens',18),
              rect(24,112,1552,250,'#FFFFFF',16), text(48,147,'Regular / 256-unit frame',18,GREEN,'700'),
              embed(logo('symbol','black'),66,166,160), text(270,147,'Small / 16–24 px cut',18,GREEN,'700'),
              embed(logo('symbol-small','black'),302,166,160), rect(530,164,180,180,DARK,14),
              embed(logo('symbol','white'),550,184,140), text(536,147,'Thinned reversal',18,GREEN,'700'),
              rect(756,164,180,180,GREEN,14), embed(logo('symbol','white'),776,184,140),
              text(762,147,'White on Forest',18,GREEN,'700'),
              embed(logo('horizontal'),1000,190,520,144), text(1000,147,'Outlined compact name',18,GREEN,'700'),
              rect(24,390,1050,302,'#FFFFFF',16), text(48,426,'Native sizes / black, then reversed',20,GREEN,'700'),
              rect(54,554,390,106,DARK,10), text(500,426,'16 px / 6× nearest pixels',18,GREEN,'700'),
              rect(1102,390,474,302,'#FFFFFF',16), text(1126,426,'App masks / safe foreground',20,GREEN,'700')]
    coords=[]
    for size,x in zip([16,24,32,64],[76,138,216,318]):
        form='symbol-small' if size<=24 else 'symbol'
        proofs += [embed(logo(form,'black'),x,462,size), text(x,544,f'{size}px',14), embed(logo(form,'white'),x,574,size)]
        coords.append((size,x,462,574))
    # Platform masking examples, with the PWA guaranteed 80%-diameter safe circle.
    proofs += ['<defs><clipPath id="circle-mask"><circle cx="1204" cy="520" r="64"/></clipPath></defs>',
               f'<g clip-path="url(#circle-mask)">{embed(icon,1140,456,128)}</g>',
               embed(FINAL/'icons/maskable-icon.svg',1338,456,128),
               '<circle cx="1402" cy="520" r="51.2" fill="none" stroke="#FFFFFF" stroke-dasharray="4 4" stroke-width="1"/>',
               text(1138,624,'Circle crop',16), text(1328,624,'80% safe circle',16),
               rect(24,720,1552,290,'#FFFFFF',16), text(48,758,'Minimum screen sizes / actual widths',21,GREEN,'700')]
    for label, form, width, x in [('Horizontal / 160 px','horizontal',160,72),('Wordmark / 128 px','wordmark',128,490),('Stacked / 128 px','stacked',128,872)]:
        proofs += [text(x,804,label,18,GREEN,'700'), embed(logo(form),x,822,width,96),
                   rect(x+184,820,width+16,116,GREEN,8), embed(logo(form,'white'),x+192,830,width,96)]
    proofs += [text(48,978,'These are screen minima, not print proofs. For crowded / low-quality contexts choose a larger size.',17),
               text(40,1054,'All logo SVGs: outlined paths, no fonts, images, strokes, masks or filters. Test sheets may use labels / diagnostic overlays.',17),
               text(40,1084,'Reference review is limited visual research, not trademark clearance. No real-device or physical print sign-off is implied.',17)]
    base=save('tests/final-proof-base',proofs,1600,1120)
    crops=FINAL/'tests/actual-size';crops.mkdir(parents=True,exist_ok=True)
    rows=[]
    with Image.open(base) as image:
        for size,x,y,ry in coords:
            for version,cy in [('black',y),('white',ry)]:
                crop=image.crop((x,cy,x+size,cy+size))
                destination=crops/f'symbol-{version}-{size}.png';crop.save(destination)
                rows.append({'version':version,'size':size,'components_at_128':helper.component_count(crop,reverse=version=='white')})
                if size==16:
                    crop.resize((96,96),Image.Resampling.NEAREST).save(crops/f'symbol-{version}-16-6x.png')
    proofs += [image_uri(crops/'symbol-black-16-6x.png',500,452,96),image_uri(crops/'symbol-white-16-6x.png',500,568,96),
               text(652,468,'16 / 24 px:',18,GREEN,'700'), text(652,497,'3 px nominal stems at 16.',16),
               text(652,526,'2 px nominal open exit.',16), text(652,610,'White cut is optically thinned.',16)]
    save('tests/final-proof',proofs,1600,1120)
    (FINAL/'tests/native-size-metrics.json').write_text(json.dumps(rows,indent=2)+'\n')

    # Compare the retained choice to its final refinement, then inspect transforms.
    comparison=[text(40,52,'Open Ground / refinement and readings',30,GREEN,'700')]
    labels=['Chosen concept v2','Final v3','Mirrored','180° rotation','Soft squint']
    for x,label in zip([48,320,592,864,1136],labels):comparison.append(text(x,108,label,19,GREEN,'700'))
    comparison += [embed(ROOT/'concepts/b-open-ground/iterations/v2/symbol.svg',44,148,200), embed(logo('symbol','black'),316,148,200),
                   f'<g transform="translate(1380 0) scale(-1 1)">{embed(logo("symbol","black"),592,148,200)}</g>',
                   f'<g transform="translate(1928 496) rotate(180)">{embed(logo("symbol","black"),864,148,200)}</g>',
                   '<defs><filter id="blur"><feGaussianBlur stdDeviation="3"/></filter></defs>',
                   f'<g filter="url(#blur)">{embed(logo("symbol","black"),1136,148,200)}</g>',
                   text(40,422,'Refinement keeps the selected open arch and inward foot; only crown balance / optical variants change.',19),
                   text(40,460,'Known reading risk: an n / doorway / vessel association. No independent audience or cultural test has been run.',18)]
    save('tests/refinement-and-transforms',comparison,1400,500)

    peers=[('monzo',ROOT/'research/references/monzo.svg'),('n26',ROOT/'research/references/n26.svg'),
           ('starlingbank',ROOT/'research/references/starlingbank.svg'),('revolut',ROOT/'research/references/revolut.svg'),
           ('wise',ROOT/'research/references/wise.svg'),('chase',ROOT/'research/references/chase.svg')]
    shelf=[text(40,52,'Open Ground / limited category comparison',30,GREEN,'700'),embed(logo('symbol','black'),56,112,144,normalize=True),
           embed(logo('horizontal','black'),250,142,530,120),text(830,166,'Selected B / final refinement',22,GREEN,'700'),
           text(830,204,'Equal visible-size category marks below.',17),text(830,237,'Snapshots, not verified current manuals.',17)]
    for i,(name,path) in enumerate(peers):
        x=60+i*220
        shelf += [embed(path,x,340,112,color='#000000',normalize=True),text(x,491,name,18)]
    shelf += [text(40,560,'The arch is not a conventional shield, arrow, coin or closed nested-square badge; n / doorway ambiguity remains.',18),
              text(40,594,'Other marks remain their owners’ trademarks. This limited comparison is not legal clearance or an exhaustive look-alike search.',17)]
    save('tests/final-shelf-test',shelf,1400,630)


if __name__=='__main__':build()
