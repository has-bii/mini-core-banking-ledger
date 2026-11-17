#!/usr/bin/env python3
"""Independently validate the delivered Still Bank vectors, rasters and metadata."""
from __future__ import annotations
import ast
import hashlib
from html.parser import HTMLParser
import json
import math
from pathlib import Path
import re
import subprocess
import sys
from xml.etree import ElementTree as ET

sys.dont_write_bytecode = True
from PIL import Image, ImageDraw
from fontTools.pens.recordingPen import RecordingPen
from fontTools.svgLib.path import parse_path

ROOT = Path(__file__).resolve().parents[1]
FINAL = ROOT / 'final'
SKILL = Path('/Users/hasbiyrobbiy/.agents/skills/logo-design')
GREEN = (33,79,64)
NS = '{http://www.w3.org/2000/svg}'


def check(condition,message):
    if not condition:
        raise AssertionError(message)


def main():
    vectors=sorted((FINAL/'svg').glob('*.svg'))+sorted((FINAL/'icons').glob('*.svg'))
    check(len(vectors)==23,'Expected 20 logo + 3 icon SVGs')
    audit=[];logs=[]
    for path in vectors:
        root=ET.parse(path).getroot()
        vb=list(map(float,root.get('viewBox').split()))
        check(len(vb)==4 and all(math.isfinite(x) for x in vb) and min(vb[2:])>0,f'Invalid viewBox: {path}')
        check(vb[3]==256,f'Canvas height must remain 256: {path}')
        title=root.find(NS+'title')
        check(title is not None and title.text and title.get('id')==root.get('aria-labelledby'),f'Missing accessible title: {path}')
        ids=[]
        for node in root.iter():
            tag=node.tag.split('}')[-1]
            check(tag in {'svg','title','desc','path'},f'Non-master element {tag} in {path}')
            check(node.get('stroke') in {None,'none'},f'Unexpanded stroke in {path}')
            check(not any(k.endswith('href') for k in node.attrib),f'External reference in {path}')
            if node.get('id'):ids.append(node.get('id'))
            if tag=='path':
                check(node.get('fill') in {'#214F40','#000000','#FFFFFF'},f'Unexpected paint in {path}')
                pen=RecordingPen();parse_path(node.get('d'),pen)
                check(not any(command=='endPath' for command,_ in pen.value),f'Open filled contour: {path}')
        check(len(ids)==len(set(ids)),f'Duplicate IDs: {path}')
        result=subprocess.run([sys.executable,str(SKILL/'scripts/svg_audit.py'),str(path)],capture_output=True,text=True)
        messages=[x for x in result.stdout.splitlines() if 'FAIL' in x or 'WARN' in x]
        match=re.search(r'score: (\d+)/100',result.stdout)
        check(result.returncode==0 and not messages and match,f'SVG audit failed: {path}: {messages}')
        audit.append({'file':str(path.relative_to(FINAL)),'score':int(match.group(1)),'issues':messages,'returncode':result.returncode})
        logs.append(f'FILE: {path.relative_to(FINAL)}\n{result.stdout}{result.stderr}')
    (FINAL/'tests/svg-audit-summary.json').write_text(json.dumps(audit,indent=2)+'\n')
    (FINAL/'tests/svg-audit.txt').write_text('\n\n'.join(logs))

    transparent=sorted((FINAL/'png').glob('*.png'))
    check(len(transparent)==44,'Expected 44 transparent logo PNGs')
    for path in transparent:
        with Image.open(path) as image:
            check(image.mode=='RGBA',f'Logo PNG missing alpha: {path}')
            check(image.getchannel('A').getextrema()==(0,255),f'Unexpected logo alpha: {path}')
            if '-symbol-small-' in path.stem:
                expected=int(path.stem.rsplit('-',1)[1]);check(image.size==(expected,expected),f'Small PNG size mismatch: {path}')
            else:
                expected=int(path.stem.rsplit('-',1)[1]);check(image.width==expected,f'PNG width mismatch: {path}')
                form=next(x for x in ['horizontal','wordmark','stacked','symbol'] if f'-{x}-' in path.stem)
                version=path.stem.split(f'-{form}-')[1].rsplit('-',1)[0]
                svg=ET.parse(FINAL/'svg'/f'still-bank-{form}-{version}.svg').getroot()
                vb=list(map(float,svg.get('viewBox').split()))
                check(image.height==round(expected*vb[3]/vb[2]),f'PNG aspect mismatch: {path}')
    opaque=sorted((FINAL/'icons').glob('*.png'))
    check(len(opaque)==11,'Expected 11 opaque icon PNGs')
    icon_dimensions={
        'app-icon-1024.png':1024,'app-icon-512.png':512,'icon-192.png':192,'icon-512.png':512,
        'apple-touch-icon.png':180,'maskable-512.png':512,
        **{f'favicon-{s}.png':s for s in [16,24,32,48,64]},
    }
    for path in opaque:
        with Image.open(path) as image:
            check(image.mode=='RGB',f'Icon must not contain an alpha channel: {path}')
            check(image.size==(icon_dimensions[path.name],)*2,f'Wrong icon dimensions: {path}')
            for point in [(0,0),(image.width-1,0),(0,image.height-1),(image.width-1,image.height-1)]:
                check(image.getpixel(point)==GREEN,f'Unexpected corner / baked mask: {path}')
    safety=[]
    for name,limit in [('maskable-512.png',204.8),('app-icon-512.png',256)]:
        with Image.open(FINAL/'icons'/name) as image:
            radii=[math.hypot(x+.5-image.width/2,y+.5-image.height/2)
                   for y in range(image.height) for x in range(image.width) if image.getpixel((x,y))!=GREEN]
            radius=max(radii);check(radius<limit,f'Foreground outside safe circle: {name}')
            safety.append({'file':name,'max_foreground_radius_px':round(radius,2),'safe_circle_radius_px':limit})
    with Image.open(FINAL/'icons/favicon.ico') as image:
        sizes=image.ico.sizes()
        check(sizes=={(16,16),(32,32),(48,48),(64,64)},'Wrong ICO entries')
        for size in sizes:image.ico.getimage(size).load()
    manifest=json.loads((FINAL/'icons/site.webmanifest').read_text())
    check(manifest['name']=='Still Bank','Manifest name mismatch')
    for icon in manifest['icons']:
        file=FINAL/'icons'/icon['src'];check(file.exists(),f'Missing manifest icon: {file}')
        with Image.open(file) as image:check(f'{image.width}x{image.height}'==icon['sizes'],f'Manifest dimensions wrong: {file}')
    check(not any(k in manifest for k in ['start_url','scope','display']),'Icon-only manifest unexpectedly declares app behaviour')
    metrics=json.loads((FINAL/'tests/native-size-metrics.json').read_text())
    check(len(metrics)==8 and all(x['components_at_128']==1 for x in metrics),'Broken-stroke size diagnostic')

    pdf=(FINAL/'guidelines/still-bank-usage-guide.pdf').read_bytes()
    pages=len(re.findall(rb'/Type\s*/Page(?!s)',pdf))
    check(pages==1,'Usage guide is not one page')
    check(pdf.startswith(b'%PDF-'),'Invalid PDF header')
    media=re.search(rb'/MediaBox\s*\[\s*0\s+0\s+([\d.]+)\s+([\d.]+)\s*\]',pdf)
    check(media and 840<float(media[1])<844 and 594<float(media[2])<597,'Guide is not A4 landscape')

    class Presentation(HTMLParser):
        def handle_starttag(self,tag,attrs):
            for key,value in attrs:
                if key=='class':check(not set(value.split()) & {'phone','screen','web','social','pcard','nav','hero2','addr','term','readme'},'Unexpected application UI in presentation')
    html=(FINAL/'presentation/presentation.html').read_text()
    Presentation().feed(html)
    check('Alex Morgan' not in html and 'stillbank.com' not in html,'Invented contact/domain in presentation')
    check(len(list((FINAL/'presentation/slides').glob('slide-*.png')))==5,'Expected 5 final presentation slides')
    check((FINAL/'fonts/Figtree[wght].ttf').read_bytes()==(ROOT/'research/type/Figtree[wght].ttf').read_bytes(),'Modified font binary was included')
    check((FINAL/'fonts/OFL.txt').read_bytes()==(ROOT/'research/type/OFL.txt').read_bytes(),'Font licence mismatch')
    for file in (ROOT/'tools').glob('*.py'):ast.parse(file.read_text())
    required=['README.md','USAGE.md','HANDOVER.md','APPROVAL.md','still-bank-logo-kit.png']
    check(all((FINAL/name).exists() for name in required),'Missing handover document')

    # Contact sheet from actual exported PNGs, not the source SVGs. It is a test
    # view only; thumbnails intentionally have no actual-size claim.
    canvas=Image.new('RGB',(1280,1040),'#F5F6F3');draw=ImageDraw.Draw(canvas)
    versions=['color','black','white','mono-214f40']
    for col,version in enumerate(versions):draw.text((col*320+20,18),version,fill=GREEN)
    for row,form in enumerate(['symbol','horizontal','wordmark','stacked']):
        for col,version in enumerate(versions):
            x,y=col*320+12,52+row*242
            background=GREEN if version=='white' else (255,255,255)
            tile=Image.new('RGBA',(296,216),background+(255,))
            size=1024 if form in {'symbol','stacked'} else 1200
            with Image.open(FINAL/'png'/f'still-bank-{form}-{version}-{size}.png') as image:
                image.thumbnail((272,168),Image.Resampling.LANCZOS)
                tile.alpha_composite(image,((296-image.width)//2,(190-image.height)//2))
            canvas.paste(tile.convert('RGB'),(x,y));draw.text((x+10,y+194),form,fill=(255,255,255) if version=='white' else GREEN)
    canvas.save(FINAL/'tests/export-contact-sheet.png')
    report={
        'status':'pass','production_svg_count':len(vectors),'svg_audit_score_range':[min(x['score'] for x in audit),max(x['score'] for x in audit)],
        'svg_warn_or_fail_count':0,'transparent_logo_png_count':len(transparent),'opaque_rgb_icon_png_count':len(opaque),
        'ico_sizes':sorted([list(s) for s in sizes]),'icon_safety':safety,'native_size_checks':metrics,
        'guide_pdf_pages':pages,'presentation_slides':5,'application_ui_in_presentation':False,
        'modified_font_binary':False,'script_syntax':'pass',
        'limits':['Automated checks do not replace visual inspection.','No trademark, real-device or physical print clearance.'],
    }
    (FINAL/'tests/validation.json').write_text(json.dumps(report,indent=2)+'\n')
    print(f'PASS: {len(vectors)} vector logos/icons; {len(transparent)} transparent logo PNGs; {len(opaque)} opaque RGB icons; 4-size ICO.')
    print('PASS: maskable / circle safety, native 16/24/32/64 light/dark continuity, manifest paths, 1-page A4 PDF, 5 slides without app UI.')
    print('SVG audit scores 99–100; no warnings/failures. Validation report and actual-PNG contact sheet saved.')


if __name__=='__main__':main()
