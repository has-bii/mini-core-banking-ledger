#!/usr/bin/env python3
"""Inventory and package only the final authorised kit, never unused concepts."""
from datetime import datetime, timezone
import hashlib
import json
from pathlib import Path
import sys
from xml.etree import ElementTree as ET
import zipfile

sys.dont_write_bytecode=True
from PIL import Image

ROOT=Path(__file__).resolve().parents[1]
FINAL=ROOT/'final'
INVENTORY=FINAL/'FILE-INVENTORY.json'
ARCHIVE=ROOT/'still-bank-logo-kit.zip'


def main():
    report=json.loads((FINAL/'tests/validation.json').read_text())
    if report.get('status')!='pass':raise SystemExit('Run validate_final.py successfully before packaging')
    rows=[]
    for path in sorted(FINAL.rglob('*')):
        if not path.is_file() or path==INVENTORY or path.name=='.DS_Store':continue
        relative=path.relative_to(FINAL)
        row={'path':str(relative),'bytes':path.stat().st_size,'sha256':hashlib.sha256(path.read_bytes()).hexdigest()}
        if relative.parts[0] in {'tests','guidelines','presentation'}:
            row['purpose']='evidence / guidance / presentation, not logo artwork'
        elif relative.parts[0] in {'svg','png','icons'}:
            row['purpose']='production logo / icon asset or icon metadata'
        elif relative.parts[0]=='fonts':row['purpose']='unmodified licensed type source / licence'
        else:row['purpose']='handover document / kit overview'
        if path.suffix=='.png':
            with Image.open(path) as image:
                row.update({'width_px':image.width,'height_px':image.height,'mode':image.mode})
                if 'A' in image.getbands():row['alpha_range']=list(image.getchannel('A').getextrema())
        elif path.suffix=='.svg':row['viewBox']=ET.parse(path).getroot().get('viewBox')
        rows.append(row)
    INVENTORY.write_text(json.dumps({
        'brand':'Still Bank','direction':'B — Open Ground','stage':'final kit / user-authorised production',
        'generated_utc':datetime.now(timezone.utc).isoformat(),'checksum_algorithm':'SHA-256',
        'inventory_excludes':'This inventory file itself; the external ZIP/checksum; archived concepts and research.',
        'files':rows,
    },indent=2)+'\n')
    with zipfile.ZipFile(ARCHIVE,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=9) as archive:
        for path in sorted(FINAL.rglob('*')):
            if path.is_file() and path.name!='.DS_Store':
                archive.write(path,Path('still-bank-logo-kit')/path.relative_to(FINAL))
    digest=hashlib.sha256(ARCHIVE.read_bytes()).hexdigest()
    (ROOT/'still-bank-logo-kit.zip.sha256').write_text(f'{digest}  {ARCHIVE.name}\n')
    with zipfile.ZipFile(ARCHIVE) as archive:
        bad=archive.testzip()
        if bad:raise RuntimeError(f'ZIP CRC failure: {bad}')
        file_count=len(archive.namelist())
    print(f'Packaged {file_count} files; archive {ARCHIVE.stat().st_size/1024/1024:.2f} MiB: logo/still-bank-logo-kit.zip')
    print('Wrote final/FILE-INVENTORY.json and still-bank-logo-kit.zip.sha256; archive CRC check passed.')


if __name__=='__main__':main()
