from pathlib import Path
import json, re, xml.etree.ElementTree as ET

root=Path(__file__).resolve().parents[1]
assets=[]
for file in sorted((root/'svg').glob('*.svg'))+[root/'web/miniledger-app-icon.svg',root/'web/miniledger-favicon.svg']:
    svg=ET.parse(file).getroot()
    vb=[float(x) for x in svg.get('viewBox').split()]
    width,height=vb[2:]
    nodes=[]
    def walk(el,sx=1,sy=1,tx=0,ty=0):
        for op,args in re.findall(r'(translate|scale)\(([^)]+)\)',el.get('transform','')):
            nums=[float(x) for x in re.split('[ ,]+',args.strip())]
            if op=='translate':
                tx+=sx*nums[0];ty+=sy*(nums[1] if len(nums)>1 else 0)
            else:
                sx*=nums[0];sy*=nums[1] if len(nums)>1 else nums[0]
        tag=el.tag.split('}')[-1]
        if tag=='path':
            nodes.append({'type':'path','name':f'{file.stem} outline {len(nodes)+1}', 'width':width,'height':height,'viewBox':[-tx/sx,-ty/sy,width/sx,height/sy],'geometry':el.get('d'),'fill':el.get('fill','#17151B')})
        elif tag=='rect':
            nodes.append({'type':'rectangle','name':f'{file.stem} background','x':tx+sx*float(el.get('x',0)),'y':ty+sy*float(el.get('y',0)),'width':sx*float(el.get('width')),'height':sy*float(el.get('height')),'cornerRadius':sx*float(el.get('rx',0)),'fill':el.get('fill')})
        elif tag in ['svg','g']:
            for child in el:walk(child,sx,sy,tx,ty)
    walk(svg)
    assets.append({'name':file.stem,'width':width,'height':height,'nodes':nodes})
print(json.dumps(assets))
