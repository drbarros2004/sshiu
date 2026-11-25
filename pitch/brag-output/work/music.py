# trilha do vídeo de lançamento: chiptune original, 96 bpm, 16 compassos = 40 s
# música e efeitos numa grade só (contada do quadro 0), no mesmo tom (fá maior).
# uso: python3 music.py trilha.wav
import sys
import numpy as np
from scipy.signal import butter, sosfilt
from scipy.io import wavfile
SR=44100; BPM=96; BEAT=60/BPM; BAR=4*BEAT; NB=16; DUR=40.0; S16=BEAT/4
N=int(SR*DUR); mix=np.zeros((N,2)); fx=np.zeros((N,2))
rng=np.random.default_rng(11)
hz=lambda m:440*2**((m-69)/12)
def env(n,a=.005,d=.3,s=0.0,r=.05):
    t=np.arange(n)/SR; e=np.minimum(1,t/a)*((1-s)*np.exp(-t/d)+s)
    rl=int(r*SR)
    if rl and rl<n: e[-rl:]*=np.linspace(1,0,rl)
    return e
def osc(kind,f,n,duty=.25):
    t=np.arange(n)/SR; ph=(f*t)%1
    if kind=='sin': return np.sin(2*np.pi*ph)
    if kind=='tri': return 2*np.abs(2*ph-1)-1
    if kind=='pulse': return np.where(ph<duty,1.0,-1.0)
def lp(x,fc): return sosfilt(butter(2,fc,fs=SR,output='sos'),x)
def hp(x,fc): return sosfilt(butter(2,fc,'hp',fs=SR,output='sos'),x)
def add(x,t0,gain,pan=0,bus=None):
    bus=mix if bus is None else bus
    i=int(round(t0*SR)); x=x[:max(0,N-i)]
    bus[i:i+len(x),0]+=x*gain*(1-pan); bus[i:i+len(x),1]+=x*gain*(1+pan)
def bell(m,t0,gain,d=.5,pan=0,bus=None):
    n=int((d*3+.2)*SR); t=np.arange(n)/SR
    x=(np.sin(2*np.pi*hz(m)*t)+.25*np.sin(2*np.pi*hz(m)*2*t))*env(n,a=.003,d=d)
    add(x,t0,gain,pan,bus)
CH=[[53,57,60,64],[52,55,59,62],[50,53,57,60],[48,52,55,59]]   # Fmaj7 Em7 Dm7 Cmaj7
ROOT=[41,40,38,36]
for b in range(NB):
    c=CH[b%4]; t0=b*BAR; last=b>=NB-2
    if b==0: continue                                  # compasso 1: só a digitação
    # pad
    n=int(BAR*SR*(2.2 if b==NB-2 else 1.05))
    if b<NB-1:
        pad=sum(osc('tri',hz(m),n) for m in (CH[0] if b==NB-2 else c))/4
        add(lp(pad*env(n,a=.25 if b>1 else .05,d=99,s=1,r=.5),1900),t0,.10)
    if b==1: continue                                  # compasso 2: entra o pad, o resto espera
    if b>=NB-2: continue
    # baixo: seno+tri uma oitava acima da raiz, em 1, 2e e 3
    for k in [0,1.5,2]:
        n=int(BEAT*1.2*SR); x=osc('sin',hz(ROOT[b%4]+12),n)+.25*osc('tri',hz(ROOT[b%4]+12),n)
        add(lp(x*env(n,a=.008,d=.35),900),t0+k*BEAT,.16)
    # bumbo macio em 1 e 3
    for k in [0,2]:
        n=int(.25*SR); t=np.arange(n)/SR; f=55+70*np.exp(-t/.03)
        add(np.sin(2*np.pi*np.cumsum(f)/SR)*env(n,a=.001,d=.09),t0+k*BEAT,.22)
    # palma leve em 2 e 4
    for k in [1,3]:
        n=int(.15*SR); x=hp(lp(rng.standard_normal(n),3500),900)*env(n,a=.001,d=.05)
        add(x,t0+k*BEAT,.05,.1)
    # chimbal nos contratempos
    for k in range(4):
        n=int(.05*SR); h=hp(rng.standard_normal(n)*env(n,a=.001,d=.015),7000)
        add(h,t0+(k+.5)*BEAT,.035,.25)
    # arpejo em colcheias
    seq=[c[0]+12,c[2]+12,c[1]+12,c[3]+12,c[2]+12,c[1]+12,c[3]+12,c[2]+12]
    for k,m in enumerate(seq):
        n=int(BEAT/2*SR*1.5)
        add(lp(osc('pulse',hz(m),n,duty=.25)*env(n,d=.14),2600),t0+k*BEAT/2,.05,.35 if k%2 else -.35)
# melodia de sino no jardim e no pet (compassos 9–14)
MEL=[(8,0,72),(8,1.5,74),(8,3,76),(9,0,74),(9,2,71),(10,0,69),(10,1.5,72),(10,3,74),
     (11,0,72),(11,2,67),(12,0,72),(12,1.5,74),(12,3,77),(13,0,76),(13,2,74)]
for b,beat,m in MEL: bell(m,b*BAR+beat*BEAT,.075,d=.5,pan=-.1)
# fecho: acorde de fá ao sino, no tempo do logo
for i,m in enumerate([65,69,72,76]): bell(m,14.5*BAR+i*S16,.06,d=1.6,pan=(i-1.5)*.2)
bell(77,36.25+BEAT*2,.06,d=1.8)

# ---- efeitos (mesmo tom, mesma grade) ----
# digitação de "ssh sshiu": semicolcheias a partir do tempo 2 (bate com o vídeo)
TYPE0=BEAT
for k in range(9):
    n=int(.04*SR); x=hp(rng.standard_normal(n),2500)*env(n,a=.0005,d=.008)
    add(x,TYPE0+k*S16,.10,.1*(k%2*2-1),fx)
# enter: um "toc" grave no compasso 2
n=int(.18*SR); t=np.arange(n)/SR
add(np.sin(2*np.pi*np.cumsum(180*np.exp(-t/.05)+70)/SR)*env(n,a=.001,d=.06),BAR,.22,0,fx)
bell(77,BAR,.05,d=.7,bus=fx)
# sopro curto antes de cada corte, e um sino baixinho no corte
CUTS=[5,10,15,20,22.5,25,27.5,32.5,36.25]
NOTE={5:65,10:64,15:62,20:60,22.5:65,25:64,27.5:62,32.5:65,36.25:65}
for c in CUTS:
    n=int(.32*SR); x=lp(hp(rng.standard_normal(n),1200),5000)*np.linspace(0,1,n)**2
    add(x,c-.32,.025,0,fx)
    bell(NOTE[c]+12,c,.03,d=.25,bus=fx)
# pet: um blip por humor, subindo (fá maior), nos tempos depois do corte
for i,m in enumerate([72,74,76,77,81]):
    n=int(.2*SR); add(lp(osc('pulse',hz(m),n,duty=.5)*env(n,d=.06),3000),27.5+BEAT*(i+1),.04,0,fx)
# loja: tique de seleção a cada troca de chapéu (tempos aproximados da tomada)
mix+=fx
# eco 3/8
d=int(BEAT*1.5*SR); wet=np.zeros_like(mix); wet[d:]=mix[:-d]*.18; mix+=wet[:, ::-1]
mix=sosfilt(butter(2,60,'hp',fs=SR,output='sos'),mix,axis=0)
# compressão suave (tanh) e fade final
mix=np.tanh(mix*1.4)/1.4
f=np.ones(N); fo=int(1.6*SR); f[-fo:]=np.linspace(1,0,fo)**1.5; mix*=f[:,None]
mix/=np.abs(mix).max(); mix*=10**(-1.5/20)
wavfile.write(sys.argv[1],SR,(mix*32767).astype(np.int16))
print('ok',len(mix)/SR)
