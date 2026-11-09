from pathlib import Path
import copy, json, re, xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]
BASE = ROOT.parent / 'logo-geometric' / 'lGRBc.svg'
NS = '{http://www.w3.org/2000/svg}'
ET.register_namespace('', NS[1:-1])
symbol = ET.parse(BASE).getroot()
wordmark = ET.parse(ROOT / 'work/n8LMBS.svg').getroot()
for folder in ['svg', 'png', 'web', 'presentation']:
    (ROOT / folder).mkdir(exist_ok=True)

def rounded(value):
    return re.sub(r'-?\d+\.\d+', lambda m: str(round(float(m.group()), 2)).rstrip('0').rstrip('.'), value)

def group(source, color, small=False):
    g = ET.Element(NS+'g')
    for path in source.iter(NS+'path'):
        p = copy.deepcopy(path)
        for k in ['d', 'transform']:
            if k in p.attrib:
                p.set(k, rounded(p.get(k)))
        p.set('fill', color)
        if small:
            outer = ET.SubElement(g, NS+'g', {'transform': 'translate(5 0)' if '109.' in p.get('transform', '') else 'translate(-5 0)'})
            outer.append(p)
        else:
            g.append(p)
    return g

def transform(parent, child, spec):
    wrapper = ET.SubElement(parent, NS+'g', {'transform':spec})
    wrapper.append(child)

def save(name, w, h, elements, title):
    root = ET.Element(NS+'svg', {'viewBox':f'0 0 {w} {h}', 'width':str(w), 'height':str(h), 'role':'img', 'aria-labelledby':'title'})
    ET.SubElement(root, NS+'title', {'id':'title'}).text = title
    for element in elements:
        root.append(element)
    ET.ElementTree(root).write(ROOT/'svg'/f'{name}.svg', encoding='unicode', xml_declaration=False)

for label, color in [('plum','#7040BA'),('ink','#17151B'),('black','#000000'),('white','#FFFFFF')]:
    g=ET.Element(NS+'g')
    transform(g, group(symbol,color), 'scale(1.230769)')
    save(f'miniledger-symbol-{label}',256,256,[g],f'MiniLedger geometric Balanced M — {label}')
    h=ET.Element(NS+'g')
    transform(h, group(symbol,color), 'scale(1.230769)')
    transform(h, group(wordmark,color), 'translate(280 65.47) scale(4)')
    save(f'miniledger-horizontal-{label}',1000,256,[h],f'MiniLedger horizontal logo — {label}')
    s=ET.Element(NS+'g')
    transform(s, group(symbol,color), 'translate(192 0) scale(1.230769)')
    transform(s, group(wordmark,color), 'translate(44.93 292) scale(3.2)')
    save(f'miniledger-stacked-{label}',640,424,[s],f'MiniLedger stacked logo — {label}')
    w=ET.Element(NS+'g')
    transform(w,group(wordmark,color),'scale(4)')
    save(f'miniledger-wordmark-{label}',687.68,125.06,[w],f'MiniLedger outlined wordmark — {label}')

for label,color in [('plum','#7040BA'),('ink','#17151B'),('white','#FFFFFF')]:
    g=ET.Element(NS+'g')
    transform(g,group(symbol,color,small=True),'scale(1.230769)')
    save(f'miniledger-symbol-small-{label}',256,256,[g],f'MiniLedger small-size Balanced M — {label}')

spec={
    'brand':'MiniLedger','tagline':'Balanced by design.',
    'brief':'A mobile ledger for operations staff and accountants. Equal mirrored halves and an open center express balance, precision and clarity.',
    'adjectives':['geometric','dependable','clear','approachable'],
    'criteria':'Works in one color, preserves the center split, and uses the existing plum and mint palette.',
    'industry':'finance','mockups':['app-icon','website','favicon-tab','business-card','social','payment-card'],
    'nav':['Accounts','Transactions','Journal'],
    'brand_color':'#7040BA','tile_color':'#7040BA','greyscale':False,'final':True,
    'designer':'MiniLedger','round':2,
    'concepts':[{
        'name':'Geometric Balanced M','symbol':'../svg/miniledger-symbol-plum.svg',
        'lockup':'../svg/miniledger-horizontal-plum.svg',
        'stacked':'../svg/miniledger-stacked-plum.svg',
        'symbol_on_tile':'../svg/miniledger-symbol-white.svg',
        'symbol_on_dark':'../svg/miniledger-symbol-white.svg',
        'lockup_on_dark':'../svg/miniledger-horizontal-white.svg',
        'lockup_on_tile':'../svg/miniledger-horizontal-white.svg',
        'stacked_on_tile':'../svg/miniledger-stacked-white.svg',
        'idea':'Two mirrored halves form one balanced M, with a continuous central gap.',
        'rationale':['Angular forms and flat terminals give the mark a geometric character.',
                     'An outlined wordmark keeps the artwork independent of installed fonts.',
                     'A dedicated small-size cut opens the center for favicons.']
    }],
    'recommendation':'Use the plum horizontal logo on light surfaces. Use the white symbol on plum for app icons. Application mockups illustrate the identity and do not represent shipped products.',
    'next_steps':['Use the supplied SVG masters for interface work.','Follow the clear-space and minimum-size guidance.','Use the small-size symbol at 16–24 px.']
}
(ROOT/'presentation/spec.json').write_text(json.dumps(spec,indent=2))
print('Wrote 19 outlined SVG masters and presentation spec')
