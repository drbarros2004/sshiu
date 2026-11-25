# monta os pets da capa e do fechamento a partir dos PNGs do app, com a mesma lógica do
# pet.Compose (internal/pet/appearance.go). Uso: python3 pets.py deck/pets previa.png
import json, sys, os
from PIL import Image
A=os.path.join(os.path.dirname(os.path.abspath(__file__)),'..','assets','pets')+'/'
OUT=sys.argv[1]
cat=json.load(open(A+'characters.json')); hats=json.load(open(A+'hats.json'))
lay=json.load(open(A+'sprite-layout.json')); main={m['id']:m['spriteSheet'] for m in json.load(open(A+'main-assets.json'))['assets']}
FW,FH=lay['frameWidth'],lay['frameHeight']; HW,HH=hats['width'],hats['height']; HEAD=6
def compose(char,variant,mood,hat=None,hv=None,frame=0):
    ch=next(c for c in cat['characters'] if c['id']==char); v=next(x for x in ch['variants'] if x['id']==variant)
    cl=json.load(open(A+ch['spriteLayout'])) if 'spriteLayout' in ch else {}
    order=lay['moodOrder']; row=order.index(mood); hatY=cl.get('hatY',lay.get('hatY'))
    W,H=FW,FH+HEAD; c=Image.new('RGBA',(W,H),(0,0,0,0))
    for layer in (ch['composition']['withHat'] if hat else ch['composition']['withoutHat']):
        if layer=='hats/selected':
            hh=next(x for x in hats['hats'] if x['id']==hat); img=Image.open(A+next(x for x in hh['variants'] if x['id']==hv)['image']).convert('RGBA')
            top=HEAD+FH-hatY[row][frame]-(HH-1); c.alpha_composite(img,(0,top)); continue
        if layer=='characters/body': src=v['spriteSheet']
        elif layer.startswith('characters/other_assets/'): src=v['otherAssets'][layer.split('/')[-1]]
        else: src=main[layer.split('/')[-1]]
        sheet=Image.open(A+src).convert('RGBA'); fr=sheet.crop((frame*FW,row*FH,frame*FW+FW,row*FH+FH))
        c.alpha_composite(fr,(0,HEAD))
    return c
TEAM=[('bernardo','fruit','purple','cowboy','normal'),('daniel','bunny','blue','top_hat','normal'),
      ('gabriel','fruit','green',None,None),('guilherme','bunny','pink','crown','normal')]
for mood,tag in [('good','foco'),('radiant','radiante')]:
    ims=[(n,compose(c,v,mood,h,hv)) for n,c,v,h,hv in TEAM]
    top=min(im.getbbox()[1] for _,im in ims)
    for n,im in ims:
        im=im.crop((0,top,FW,FH+HEAD))
        # salvo já ampliado 16× (vizinho mais próximo): no PDF o leitor só reduz, sem borrar os pixels
        im.resize((im.width*16,im.height*16),Image.NEAREST).save(f'{OUT}/{n}-{tag}.png')
    print(mood,'altura',FH+HEAD-top)
# prévia ampliada
import glob
fs=sorted(glob.glob(OUT+'/*.png')); prev=Image.new('RGBA',(len(fs)*20*8,22*8),(240,240,240,255))
for k,f in enumerate(fs): im=Image.open(f); prev.alpha_composite(im.resize((im.width//2,im.height//2),Image.NEAREST),(k*160,0))
prev.save(sys.argv[2]); print([os.path.basename(f) for f in fs])
