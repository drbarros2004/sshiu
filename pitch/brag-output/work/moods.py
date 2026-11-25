# pets do vídeo: o mesmo compose do pets.py, um pet em cada humor, nos 4 quadros da animação
import sys, os
src=open(os.path.join(os.path.dirname(__file__),'..','..','pets.py')).read().split('TEAM=')[0]
sys.argv=[sys.argv[0],'x']; os.chdir(os.path.join(os.path.dirname(__file__),'..','..'))
__file__=os.path.abspath("pets.py"); exec(src)
from PIL import Image
out='brag-output/work/pets'
for mood in ['wilted','tired','normal','good','radiant']:
    for f in range(4):
        im=compose('fruit','blue',mood,frame=f)
        im.resize((im.width*16,im.height*16),Image.NEAREST).save(f'{out}/fruit-{mood}-{f}.png')
print('ok')
