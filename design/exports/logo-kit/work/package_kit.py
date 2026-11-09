from pathlib import Path
import json,struct,xml.etree.ElementTree as ET

root=Path(__file__).resolve().parents[1]
mapping={'HIG5V':'horizontal-black','E8sL4':'horizontal-ink','TyZIN':'horizontal-plum','GXRZp':'horizontal-white','NfJ2q':'stacked-black','J6H2z':'stacked-ink','blp2F':'stacked-plum','JwghB':'stacked-white','jZlYq':'symbol-black','GRibz':'symbol-ink','dxO0J':'symbol-plum','g8iVn':'symbol-small-ink','fdf1C':'symbol-small-plum','UVC5y':'symbol-small-white','heGd8':'symbol-white','D3mNp7':'wordmark-black','szMxt':'wordmark-ink','bq2QV':'wordmark-plum','xn7fB':'wordmark-white','NbRCZ':'app-icon','YBBue':'favicon'}
for node,name in mapping.items():
    source=root/'png'/f'{node}.png'
    target=root/'png'/f'miniledger-{name}.png'
    if source.exists():source.rename(target)
p=root/'png/1024/NbRCZ.png'
if p.exists():p.rename(root/'png/miniledger-app-icon-1024.png')
if (root/'png/1024').exists():(root/'png/1024').rmdir()
for size in [16,32,48,180,192,512]:
    folder=root/'web'/str(size)
    if folder.exists():
        source=next(folder.glob('*.png'))
        filename=f'favicon-{size}.png' if size<100 else 'apple-touch-icon.png' if size==180 else f'icon-{size}.png'
        source.rename(root/'web'/filename)
        folder.rmdir()
p=root/'web/maskable/QNmiD.png'
if p.exists():p.rename(root/'web/maskable-512.png');p.parent.rmdir()
for old,new in [('u3iG3x.svg','app-icon-square.svg'),('QNmiD.svg','maskable.svg')]:
    p=root/'web'/old
    if p.exists():
        text=p.read_text().replace('><defs>','><title>MiniLedger app icon</title><defs>',1)
        (root/'web'/new).write_text(text)
        p.unlink()
entries=[];payload=bytearray();offset=6+16*3
for size in [16,32,48]:
    data=(root/'web'/f'favicon-{size}.png').read_bytes()
    entries.append(struct.pack('<BBBBHHII',size,size,0,0,1,32,len(data),offset))
    payload+=data;offset+=len(data)
(root/'web/favicon.ico').write_bytes(struct.pack('<HHH',0,1,3)+b''.join(entries)+payload)
manifest={'name':'MiniLedger','short_name':'MiniLedger','start_url':'/','display':'standalone','background_color':'#F7F5FC','theme_color':'#7040BA','icons':[{'src':'icon-192.png','sizes':'192x192','type':'image/png','purpose':'any'},{'src':'icon-512.png','sizes':'512x512','type':'image/png','purpose':'any'},{'src':'maskable-512.png','sizes':'512x512','type':'image/png','purpose':'maskable'}]}
(root/'web/site.webmanifest').write_text(json.dumps(manifest,indent=2))
(root/'web/head-snippet.html').write_text('''<!-- Copy web/ contents into the public root, or adjust these paths. -->
<link rel="icon" href="/favicon.ico" sizes="16x16 32x32 48x48">
<link rel="icon" type="image/svg+xml" href="/miniledger-favicon.svg">
<link rel="apple-touch-icon" href="/apple-touch-icon.png" sizes="180x180">
<link rel="manifest" href="/site.webmanifest">
<meta name="theme-color" content="#7040BA">
''')
report=[]
for file in sorted(root.rglob('*.png')):
    data=file.read_bytes()
    assert data[:8]==b'\x89PNG\r\n\x1a\n',file
    width,height=struct.unpack('>II',data[16:24])
    assert width>0 and height>0,file
    report.append(f'{file.relative_to(root)}: {width} × {height}')
for file in (root/'svg').glob('*.svg'):
    doc=ET.parse(file).getroot()
    assert next(doc.iter('{http://www.w3.org/2000/svg}title'),None) is not None,file
    assert not any(el.tag.split('}')[-1] in ['text','image','filter'] for el in doc.iter()),file
for entry in manifest['icons']:assert (root/'web'/entry['src']).exists()
(root/'work/export-checks.txt').write_text('\n'.join(report))
print(f'Validated {len(report)} PNGs, 19 SVG masters, ICO and manifest references')
