#!/usr/bin/env python3
"""Build a compact final presentation with isolated identity mockups, never UI.

Uses the skill presentation builder, overriding its default screen/contact-data
mockups. Also renders a one-page visual usage guide and an A4 landscape PDF.
"""
from __future__ import annotations
import base64
import datetime
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parents[1]
FINAL = ROOT / 'final'
SKILL = Path('/Users/hasbiyrobbiy/.agents/skills/logo-design')
sys.path.insert(0,str(SKILL/'scripts'))


def load(name,path):
    spec=importlib.util.spec_from_file_location(name,path)
    module=importlib.util.module_from_spec(spec);sys.modules[name]=module;spec.loader.exec_module(module)
    return module


review=load('still_final_review_helpers',ROOT/'tools/build_final_review.py')
board=load('still_presentation_board',SKILL/'scripts/presentation_board.py')
renderer=load('still_presentation_renderer',SKILL/'scripts/render_png.py')
GREEN,MIST='#214F40','#F5F6F3'


def data(path):
    return 'data:image/svg+xml;base64,'+base64.b64encode(Path(path).read_bytes()).decode()


def mock(kind,ctx):
    # Real exported artwork; no invented address, employee, product, card, URL,
    # financial claim, store rating or application screen.
    icon=data(FINAL/'icons/app-icon.svg')
    favicon=data(FINAL/'icons/favicon.svg')
    lock=data(FINAL/'svg/still-bank-horizontal-color.svg')
    white=data(FINAL/'svg/still-bank-horizontal-white.svg')
    mark=data(FINAL/'svg/still-bank-symbol-color.svg')
    stack=data(FINAL/'svg/still-bank-stacked-mono-214f40.svg')
    if kind=='app-icon':
        body=f'<div style="text-align:center"><img src="{icon}" style="width:142px;height:142px;border-radius:32px;box-shadow:0 12px 24px #214f4020"><div style="margin-top:14px;color:{GREEN};font-size:16px">Still Bank</div></div>'
        label='OS app tile · illustrated mask'
    elif kind=='favicon-tab':
        body=f'<div style="width:310px;background:#dfe3df;padding:10px 10px 0;border-radius:12px"><div style="display:flex;align-items:center;gap:10px;background:#fff;width:244px;height:42px;border-radius:9px 9px 0 0;padding:0 12px"><img src="{favicon}" width="16" height="16"><span style="font-size:13px;color:{GREEN}">Still Bank</span></div></div>'
        label='Favicon · isolated browser tab'
    elif kind=='launcher-masks':
        body=f'<div style="display:flex;gap:28px"><img src="{icon}" style="width:112px;height:112px;border-radius:50%"><img src="{icon}" style="width:112px;height:112px;border-radius:27px"></div>'
        label='Launcher identity · circle / rounded masks'
    elif kind=='letterhead':
        body=f'<div style="width:192px;height:224px;background:#fff;padding:18px;box-shadow:0 10px 24px #214f401c;transform:rotate(-4deg)"><img src="{lock}" style="width:150px;height:auto"><div style="margin-top:16px;height:1px;background:#d9e3dc"></div></div>'
        label='Blank stationery · identity only'
    elif kind=='signature':
        body=f'<div style="width:318px;padding:24px 16px;background:{GREEN};border-radius:5px;box-shadow:0 8px 20px #214f4020"><img src="{white}" style="display:block;width:286px;height:auto"></div>'
        label='Compact brand signature'
    elif kind=='one-colour':
        body=f'<div style="text-align:center;background:#fff;width:242px;padding:24px;box-shadow:0 8px 20px #214f4014"><img src="{stack}" style="display:block;margin:0 auto;width:190px;height:auto"></div>'
        label='One-colour document imprint'
    else:
        raise KeyError(kind)
    return f'<div class="m" style="background:{MIST}">{body}<span class="lbl">{label}</span></div>'


def presentation():
    folder=FINAL/'presentation';folder.mkdir(parents=True,exist_ok=True)
    chosen=['app-icon','favicon-tab','launcher-masks','letterhead','signature','one-colour']
    spec={
        'brand':'Still Bank','tagline':'','deck_title':'Open Ground — final logo kit','final':True,'greyscale':False,
        'date':datetime.date.today().isoformat(),'industry':'consumer-app','mockups':chosen,
        'brand_color':GREEN,'tile_color':GREEN,
        'brief':'A personal mobile banking app for everyday customers. Quiet Confidence: calm, precise, trustworthy and approachable; never cold, flashy or exclusive.',
        'adjectives':['calm','precise','trustworthy','approachable'],
        'criteria':'Mobile-first identity: recognisable app / favicon identifier, a readable Still Bank name, clear 16/24/32/64 px forms, and one-colour / reversed flexibility. Mockups are isolated branding applications, not UI screens or product promises.',
        'concepts':[{
            'name':'B — Open Ground','symbol':'../svg/still-bank-symbol-color.svg',
            'lockup':'../svg/still-bank-horizontal-color.svg','stacked':'../svg/still-bank-stacked-color.svg',
            'symbol_on_tile':'../svg/still-bank-symbol-white.svg','symbol_on_dark':'../svg/still-bank-symbol-white.svg',
            'lockup_on_dark':'../svg/still-bank-horizontal-white.svg','stacked_on_dark':'../svg/still-bank-stacked-white.svg',
            'idea':'An open arch with an inward-turning foot gives the mark a quiet, grounded centre.',
            'rationale':['The selected silhouette is retained, with a symmetrical crown and open 32-unit exit.',
                         'Forest #214F40 with Mist #F5F6F3; one-colour form carries recognition.',
                         '16/24 px use a sturdier small cut. Reversed mark and name have optical weight correction.',
                         'The abstract n / doorway reading remains a known trade-off; use the name while recognition builds.'],
        }],
        'recommendation':'Direction B was selected by the user, and kit production was explicitly authorised. This kit contains refined vectors, raster exports, icon assets and usage guidance. No separate legal clearance or product capability is implied.',
        'next_steps':['Use the SVG / PNG / icon files documented in README.md and USAGE.md.',
                      'Run professional name/device clearance in the intended markets before launch.',
                      'Proof print colours and real-device launcher masks before production.'],
    }
    spec_file=folder/'presentation-spec.json';spec_file.write_text(json.dumps(spec,indent=2)+'\n')
    board.mock_html=mock
    board.pick_mockups=lambda _:chosen
    small=data(FINAL/'svg/still-bank-symbol-small-color.svg')
    board.CSS += f'\n.sizes>div:nth-child(-n+2)>img{{content:url("{small}")}}\n'
    sys.argv=[str(SKILL/'scripts/presentation_board.py'),str(spec_file),'-o',str(folder/'presentation.html'),'--png-dir',str(folder/'slides')]
    board.main()


def guide():
    text,rect,embed=review.text,review.rect,review.embed
    logo=review.logo
    parts=[text(48,58,'Still Bank / logo usage',34,GREEN,'700'), text(48,94,'Open Ground · one-page quick guide',19),
           embed(logo('horizontal'),1000,10,552,140),rect(48,146,920,220,'#FFFFFF',18),
           text(76,181,'Use the supplied versions',22,GREEN,'700'),embed(logo('horizontal'),72,200,440,136),
           embed(logo('stacked'),568,198,220,146),embed(logo('symbol'),814,222,112),
           rect(1000,146,552,360,'#FFFFFF',18),text(1028,181,'Clear space',22,GREEN,'700'),
           embed(logo('symbol'),1120,230,160),
           '<rect x="1120" y="225" width="160" height="165" fill="none" stroke="#214F40" stroke-dasharray="6 5" stroke-width="1.5"/>',
           text(1310,300,'X',24,GREEN,'700'),text(1028,434,'X = one symbol stem width.',19),
           text(1028,465,'Keep X clear around visible ink, not SVG edges.',17),
           rect(48,394,920,232,'#FFFFFF',18),text(76,430,'Colour / use one colour at a time',22,GREEN,'700'),
           rect(76,458,80,80,GREEN,10),text(178,480,'Forest · #214F40',21,GREEN,'700'),
           text(178,509,'RGB 33 / 79 / 64',18),text(178,540,'CMYK ≈ 58 / 0 / 19 / 69*',17),
           rect(530,458,80,80,MIST,10),text(632,480,'Mist · #F5F6F3',21,GREEN,'700'),
           text(632,509,'RGB 245 / 246 / 243',18),text(632,540,'CMYK ≈ 0 / 0 / 1 / 4*',17),
           text(76,583,'Forest on white / Mist; corrected white on Forest / dark. Black on white.',17),
           text(76,608,'*Unprofiled estimates; print and spot / Pantone matching require a physical proof.',15),
           rect(1000,534,552,264,'#FFFFFF',18),text(1028,572,'Minimum screen sizes',22,GREEN,'700')]
    for y,label,value in [(611,'Small identifier','16 px; small cut at 16–24'),(648,'Regular identifier','32 px'),(685,'Horizontal','160 px wide'),(722,'Wordmark / stacked','128 px wide')]:
        parts += [text(1028,y,label,18),text(1260,y,value,17,GREEN,'700')]
    parts += [text(1028,769,'Print: starting sizes in USAGE.md; proof first.',16),
              rect(48,654,920,144,'#FFFFFF',18),text(76,691,'Type and lockups',22,GREEN,'700'),
              text(76,725,'Figtree-based outlined name: 580 normal / 560 reversed, custom a and k.',17),
              text(76,756,'Never retype the wordmark. Keep the supplied spacing and proportions.',17),
              text(76,782,'Font: SIL OFL 1.1; source and licence notes included. No fonts needed to view SVGs.',15),
              rect(48,826,1504,218,'#FFFFFF',18),text(76,864,'Do not',22,GREEN,'700')]
    # Intentionally incorrect applications are examples in a guide, never assets.
    parts += [f'<g transform="translate(100 884) scale(1.55 .68)">{embed(logo("symbol"),0,0,130)}</g>',
              text(294,928,'Stretch or squash',19),
              embed(logo('symbol'),620,882,130,color='#C76A36'),text(768,928,'Recolour freely',19),
              f'<g transform="rotate(25 1150 947)">{embed(logo("symbol"),1085,882,130)}</g>',text(1240,928,'Rotate the mark',19)]
    for x in [282,756,1228]:
        parts.append(f'<path d="M{x} 966l16 16m0-16l-16 16" stroke="#B83D34" stroke-width="3" fill="none"/>')
    parts += [text(48,1080,'Use SVG masters; no shadows, outlines, gradients, retyping or altered lockup ratios. Add a calm container on busy imagery.',17),
              text(48,1112,'Masters: svg/ · Icons: icons/ · Full rules: USAGE.md · Limited visual research is not trademark clearance.',16)]
    review.save('guidelines/still-bank-usage-guide',parts,1600,1130)
    # A real one-page PDF handout. This is a usage document, not CMYK press art.
    svg=FINAL/'guidelines/still-bank-usage-guide.svg'
    pdf=FINAL/'guidelines/still-bank-usage-guide.pdf'
    html=f'<!doctype html><html><head><style>@page{{size:297mm 210mm;margin:0}}html,body{{margin:0;padding:0}}img{{display:block;width:297mm;height:210mm}}</style></head><body><img src="{data(svg)}"></body></html>'
    chrome=renderer.find_chrome()
    with tempfile.TemporaryDirectory(prefix='still-bank-guide-') as tmp:
        page=Path(tmp)/'guide.html';page.write_text(html)
        command=[chrome,'--headless=new','--disable-gpu','--no-first-run','--no-default-browser-check',
                 '--use-mock-keychain','--password-store=basic','--disable-extensions',
                 '--no-pdf-header-footer',f'--print-to-pdf={pdf}',f'--user-data-dir={tmp}/profile',page.as_uri()]
        result=subprocess.run(command,capture_output=True,text=True,timeout=45)
        if result.returncode or not pdf.exists():raise RuntimeError('Usage-guide PDF export failed: '+result.stderr[-1000:])
    print('Wrote one-page visual usage guide: SVG / PNG / PDF.')


if __name__=='__main__':
    presentation()
    guide()
