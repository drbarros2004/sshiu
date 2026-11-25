# trilha original do vídeo demo do sshiu: chiptune calmo, 90 bpm, 12 compassos = 32 s
# composta do zero (sem amostras de terceiros). Para refazer o vídeo:
#   python3 soundtrack.py trilha.wav
#   ffmpeg -i demo-mudo.mp4 -i trilha.wav -map 0:v -map 1:a -c:v copy \
#     -af "loudnorm=I=-18:TP=-2:LRA=7" -c:a aac -b:a 128k -shortest -movflags +faststart deck/demo.mp4
import numpy as np
from scipy.signal import butter, sosfilt
from scipy.io import wavfile
SR=44100; BPM=90; BEAT=60/BPM; BAR=4*BEAT; NB=12; DUR=32.0
N=int(SR*DUR); mix=np.zeros((N,2))
rng=np.random.default_rng(7)
hz=lambda m:440*2**((m-69)/12)
def env(n,a=.005,d=.3,s=0.0,r=.05):
    t=np.arange(n)/SR; e=np.minimum(1,t/a)*((1-s)*np.exp(-t/d)+s)
    rl=int(r*SR); 
    if rl and rl<n: e[-rl:]*=np.linspace(1,0,rl)
    return e
def osc(kind,f,n,duty=.25):
    t=np.arange(n)/SR; ph=(f*t)%1
    if kind=='sin': return np.sin(2*np.pi*ph)
    if kind=='tri': return 2*np.abs(2*ph-1)-1
    if kind=='pulse': return np.where(ph<duty,1.0,-1.0)
def lp(x,fc): return sosfilt(butter(2,fc,fs=SR,output='sos'),x)
def add(x,t0,gain,pan=0):
    i=int(t0*SR); x=x[:max(0,N-i)]
    mix[i:i+len(x),0]+=x*gain*(1-pan)/1; mix[i:i+len(x),1]+=x*gain*(1+pan)/1
# acordes (MIDI): Fmaj7, Em7, Dm7, Cmaj7
CH=[[53,57,60,64],[52,55,59,62],[50,53,57,60],[48,52,55,59]]
ROOT=[53,52,50,48]  # baixo uma oitava acima: soa em caixa de notebook
for b in range(NB):
    c=CH[b%4]; t0=b*BAR; last=b==NB-1
    # pad: triângulo, ataque lento
    n=int(BAR*SR*(1.6 if last else 1.05))
    pad=sum(osc('tri',hz(m),n) for m in c)/4
    add(lp(pad*env(n,a=.35,d=99,s=1,r=.4),1800),t0,.10)
    # baixo: seno na raiz, tempos 1 e 3
    for k in ([0] if last else [0,2]):
        n=int(BEAT*1.8*SR); add(osc('sin',hz(ROOT[b%4]),n)*env(n,a=.01,d=.6),t0+k*BEAT,.13)
    # arpejo chiptune em colcheias, a partir do compasso 2
    if 1<=b<NB-1:
        seq=[c[0]+12,c[1]+12,c[2]+12,c[3]+12,c[2]+12,c[1]+12,c[3]+12,c[2]+12]
        for k,m in enumerate(seq):
            n=int(BEAT/2*SR*1.6); pan=.35 if k%2 else -.35
            add(lp(osc('pulse',hz(m),n,duty=.25)*env(n,d=.16),2600),t0+k*BEAT/2,.065,pan)
    # chimbal suave nos contratempos, a partir do compasso 4
    if 3<=b<NB-1:
        for k in range(4):
            n=int(.06*SR); h=rng.standard_normal(n)*env(n,a=.001,d=.018)
            h=sosfilt(butter(2,7000,'hp',fs=SR,output='sos'),h); add(h,t0+(k+.5)*BEAT,.03,.2)
# melodia de sino (seno + harmônico), compassos 7 a 12: o jardim e a loja
MEL=[(6,0,72),(6,1.5,74),(6,3,76),(7,0,74),(7,2,71),(8,0,69),(8,1.5,72),(8,3,74),(9,0,72),(9,2,67),
     (10,0,72),(10,1.5,74),(10,3,77),(11,0,76)]
for b,beat,m in MEL:
    n=int((3.5 if b==11 else 1.6)*SR); t=np.arange(n)/SR
    x=(np.sin(2*np.pi*hz(m)*t)+.3*np.sin(2*np.pi*hz(m)*2*t))*env(n,a=.004,d=.55 if b<11 else 1.4)
    add(x,b*BAR+beat*BEAT,.09,-.1)
# eco leve (delay de 3/8)
d=int(BEAT*1.5*SR); wet=np.zeros_like(mix); wet[d:]=mix[:-d]*.22; mix+=wet[:, ::-1]
# corta o subgrave
mix=sosfilt(butter(2,90,'hp',fs=SR,output='sos'),mix,axis=0)
# fades e normalização
f=np.ones(N); fi=int(1.2*SR); fo=int(2.5*SR); f[:fi]=np.linspace(0,1,fi)**2; f[-fo:]=np.linspace(1,0,fo)**1.5
mix*=f[:,None]; mix/=np.abs(mix).max(); mix*=10**(-3/20)
wavfile.write(__import__('sys').argv[1],SR,(mix*32767).astype(np.int16))
print('ok',len(mix)/SR)
