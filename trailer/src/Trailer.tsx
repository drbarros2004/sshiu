import { Audio } from "@remotion/media";
import { linearTiming, TransitionSeries } from "@remotion/transitions";
import { fade } from "@remotion/transitions/fade";
import { slide } from "@remotion/transitions/slide";
import { AbsoluteFill, Easing, interpolate, OffthreadVideo, Sequence, staticFile, useCurrentFrame } from "remotion";
import { C, mono, ramp, s } from "./theme";
import { WIN, Window } from "./Window";

// Um plano = um trecho de um clipe gravado pelo harness (capture/), dentro da janela.
// from/to em segundos do clipe; rate acelera; zoom aproxima num ponto (0–1 da tela do clipe).
export type Shot = { clip: string; from: number; to: number; rate?: number; zoom?: { scale: number; x: number; y: number } };
// enter: como a cena entra. slide = a tela anterior sai para a esquerda, como trocar de mesa no Mac.
type Enter = "cut" | "slide" | "fade";
// card: abertura e fecho, telas do trailer (sem a janela do app).
// grid: trechos do mesmo clipe, cada um numa janelinha (2×2), a partir de "from" segundos.
type Scene = {
  id: string;
  enter?: Enter;
  caption?: string;
  card?: "title" | "outro";
  shots?: Shot[];
  grid?: { clip: string; from: number[]; titles: string[] };
  dur?: number;
  sfx?: { at: number; src: string; volume?: number }[];
};

export const T = 24; // quadros de cada transição; elas sobrepõem as duas cenas
const gridOpen = (i: number) => T + 4 + i * 7; // quadro em que a janelinha i abre

// abertura: "$ ssh sshiu" digitado em letra grande; no enter entram a frase e a música
export const CMD = "ssh sshiu";
export const TYPE_FROM = 14; // quadros antes da primeira tecla
export const PER_KEY = 4;
export const ENTER = TYPE_FROM + CMD.length * PER_KEY + 10;

// o mural ocupa só o alto da tela do app: aproxima e centraliza
const MURAL = { scale: 1.12, x: 0, y: -1.7 };

// timelapse (capture/scene/timelapse.sh): 5 momentos do dia, com corte seco. O primeiro dura mais
// porque metade dele passa durante a transição; o último segura para fechar a ideia.
export const moments = (clip: string, from: number, first: number, mid: number, last: number): Shot[] =>
  [0, 1, 2, 3, 4].map((k) => ({ clip: `${clip}-${k}`, from, to: from + (k === 0 ? first : k === 4 ? last : mid) }));
export const cuts = (first: number, mid: number) => [0, 1, 2, 3].map((k) => first + k * mid); // segundos de cada corte

const SCENES: Scene[] = [
  { id: "title", card: "title", dur: 4.0, enter: "cut" },
  { id: "room", caption: "a room where everyone focuses together", shots: [{ clip: "01-room", from: 0.3, to: 4.3 }] },
  {
    id: "garden",
    caption: "while you focus, the screen rests",
    dur: 4.2,
    // café, árvore, estrada e gato: 4 s cada em 02-garden-long (capture/harness/record-garden.sh)
    grid: { clip: "02-garden-long", from: [12.3, 16.3, 20.3, 28.3], titles: ["coffee", "tree", "night road", "sleeping cat"] },
    sfx: [0, 1, 2, 3].map((i) => ({ at: gridOpen(i) / 30, src: "dialog", volume: 0.2 })),
  },
  { id: "brainrot", caption: "for real devs", shots: [{ clip: "04-brainrot", from: 0, to: 4.0 }], sfx: [{ at: 0, src: "glitch", volume: 0.45 }] },
  {
    id: "coins",
    caption: "finish a task, earn coins",
    shots: [{ clip: "05-coins", from: 0.3, to: 4.1 }],
    sfx: [{ at: 0.85, src: "coin", volume: 0.45 }],
  },
  {
    id: "dress",
    caption: "dress up your pet",
    shots: moments("tl-room", 0.6, 1.6, 1.0, 1.8),
    sfx: cuts(1.6, 1.0).map((at) => ({ at, src: "dialog", volume: 0.25 })),
  },
  {
    id: "build",
    caption: "build something together",
    shots: moments("tl-mural", 0.4, 1.6, 0.9, 1.9).map((x) => ({ ...x, zoom: MURAL })),
    sfx: cuts(1.6, 0.9).map((at) => ({ at, src: "toggle", volume: 0.35 })),
  },
  { id: "profile", caption: "and see where your time went", shots: [{ clip: "08-profile", from: 0.2, to: 3.3 }] },
  // o "shh" é o começo de sfx-shh/elevenlabs-shush-source.mp3
  { id: "outro", card: "outro", dur: 3.6, sfx: [{ at: 1.3, src: "shh-voice", volume: 0.6 }] },
];

export const shotFrames = (x: Shot) => Math.round(((x.to - x.from) * 30) / (x.rate ?? 1));
const sceneFrames = (sc: Scene) => (sc.shots ? sc.shots.reduce((n, x) => n + shotFrames(x), 0) : s(sc.dur!));
const enterOf = (sc: Scene) => sc.enter ?? "slide";
const starts = SCENES.reduce<number[]>(
  (acc, sc, i) => [...acc, i === 0 ? 0 : acc[i - 1] + sceneFrames(SCENES[i - 1]) - (enterOf(sc) === "cut" ? 0 : T)],
  [],
);
export const trailerDuration = starts[starts.length - 1] + sceneFrames(SCENES[SCENES.length - 1]);

const Clip: React.FC<{ shot: Shot }> = ({ shot }) => {
  const z = shot.zoom;
  return (
    <AbsoluteFill
      style={{
        transform: z ? `scale(${z.scale})` : undefined,
        transformOrigin: z ? `${z.x * 100}% ${z.y * 100}%` : undefined,
      }}
    >
      <OffthreadVideo
        src={staticFile(`clips/${shot.clip}.mp4`)}
        trimBefore={Math.round(shot.from * 30)}
        playbackRate={shot.rate ?? 1}
        muted
        style={{ width: WIN.width, height: WIN.height }}
      />
    </AbsoluteFill>
  );
};

// o bloco do cursor de terminal, piscando a cada meio segundo depois que para de digitar
export const Cursor: React.FC<{ on: boolean; size: number }> = ({ on, size }) => (
  <span
    style={{
      display: "inline-block",
      width: size * 0.6,
      height: size * 0.95,
      marginLeft: size * 0.08,
      verticalAlign: "-0.12em",
      background: C.accent,
      opacity: on ? 1 : 0,
    }}
  />
);

export const Title: React.FC = () => {
  const f = useCurrentFrame();
  const n = Math.max(0, Math.min(CMD.length, Math.floor((f - TYPE_FROM) / PER_KEY) + 1));
  const typing = f < TYPE_FROM + CMD.length * PER_KEY;
  const up = ramp(f, ENTER, ENTER + 12);
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", flexDirection: "column", fontFamily: mono }}>
      <div style={{ fontSize: 150, fontWeight: 700, color: C.text, translate: `0px ${-40 * up}px` }}>
        <span style={{ color: C.accent }}>$</span> {f < TYPE_FROM ? "" : CMD.slice(0, n)}
        <Cursor on={typing || f >= ENTER || Math.floor(f / 15) % 2 === 0} size={150} />
      </div>
      <div style={{ fontSize: 46, color: C.muted, opacity: up, translate: `0px ${(1 - up) * 16 - 10}px` }}>
        a quiet place to focus together
      </div>
    </AbsoluteFill>
  );
};

export const Outro: React.FC = () => {
  const f = useCurrentFrame();
  const a = ramp(f, T, T + 12);
  return (
    <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", flexDirection: "column", gap: 28, fontFamily: mono }}>
      <div style={{ fontSize: 170, fontWeight: 700, color: C.accent, opacity: a, translate: `0px ${(1 - a) * 18}px` }}>
        sshiu
        <Cursor on={Math.floor(f / 15) % 2 === 0} size={170} />
      </div>
      <div style={{ fontSize: 46, color: C.text, opacity: ramp(f, T + 6, T + 18) }}>focus together, from the terminal</div>
      <div
        style={{
          marginTop: 16,
          fontSize: 42,
          color: C.accent,
          border: `3px solid ${C.accent}`,
          borderRadius: 10,
          padding: "10px 28px",
          opacity: ramp(f, T + 12, T + 24),
        }}
      >
        $ ssh sshiu
      </div>
    </AbsoluteFill>
  );
};

const Caption: React.FC<{ text: string; dur: number }> = ({ text, dur }) => {
  const f = useCurrentFrame();
  const a = Math.min(ramp(f, 2, 12), 1 - ramp(f, dur - 6, dur));
  return (
    <div
      style={{
        position: "absolute",
        top: WIN.top + WIN.height + WIN.bar + 16,
        width: "100%",
        textAlign: "center",
        fontFamily: mono,
        fontSize: 40,
        color: C.text,
        opacity: a,
        translate: `0px ${(1 - ramp(f, 2, 12)) * 14}px`,
      }}
    >
      {text}
    </div>
  );
};

export const Sfx: React.FC<{ from: number; src: string; volume?: number }> = ({ from, src, volume = 0.6 }) => (
  <Sequence from={from} layout="none" name={`sfx: ${src}`}>
    <Audio src={staticFile(`sfx/${src}.wav`)} volume={volume} />
  </Sequence>
);

// Grid do jardim: uma janelinha do Mac por wallpaper, abrindo uma depois da outra quando a
// tela já terminou de deslizar.
const GRID = { w: 780, h: 439, bar: 30, gapX: 24, gapY: 20, top: 24 };
const Grid: React.FC<{ clip: string; from: number[]; titles: string[] }> = ({ clip, from, titles }) => {
  const f = useCurrentFrame();
  const left0 = (1920 - (GRID.w * 2 + GRID.gapX)) / 2;
  return (
    <>
      {from.map((t, i) => {
        const a = ramp(f, gridOpen(i), gridOpen(i) + 10);
        return (
          <Window
            key={i}
            title={titles[i]}
            left={left0 + (i % 2) * (GRID.w + GRID.gapX)}
            top={GRID.top + Math.floor(i / 2) * (GRID.h + GRID.bar + GRID.gapY)}
            width={GRID.w}
            height={GRID.h}
            bar={GRID.bar}
            style={{ opacity: a, scale: `${0.92 + 0.08 * a}` }}
          >
            <OffthreadVideo
              src={staticFile(`clips/${clip}.mp4`)}
              trimBefore={Math.round(t * 30)}
              muted
              style={{ width: GRID.w, height: GRID.h }}
            />
          </Window>
        );
      })}
    </>
  );
};

const SceneView: React.FC<{ sc: Scene }> = ({ sc }) => {
  let at = 0;
  return (
    <AbsoluteFill style={{ background: C.page }}>
      {sc.card === "title" && <Title />}
      {sc.card === "outro" && <Outro />}
      {sc.grid && <Grid {...sc.grid} />}
      {sc.shots && (
        <Window title="maya — ssh sshiu — 160×45">
          {sc.shots.map((x, j) => {
            const d = shotFrames(x);
            const el = (
              <Sequence key={j} from={at} durationInFrames={d} name={x.clip}>
                <Clip shot={x} />
              </Sequence>
            );
            at += d;
            return el;
          })}
        </Window>
      )}
      {sc.caption && <Caption text={sc.caption} dur={sceneFrames(sc)} />}
    </AbsoluteFill>
  );
};

// A música entra no enter da abertura, no primeiro tempo do loop (o bumbo em 0 s), e some devagar no
// fecho, deixando o "shh" no silêncio. A abertura é sempre a primeira cena; outroAt: quadro do fecho.
const MUSIC_FADE = s(2.4);
export const Opening: React.FC<{ music: Music; outroAt: number }> = ({ music, outroAt }) => {
  const len = outroAt + s(1.3) - ENTER;
  return (
    <>
      <Sequence from={ENTER} durationInFrames={len} name={`música: ${music.src}`}>
        <Audio
          src={staticFile(music.src)}
          trimBefore={Math.round(music.from * 30)}
          volume={(f) =>
            music.volume *
            interpolate(f, [0, 3], [0, 1], { extrapolateRight: "clamp" }) *
            interpolate(f, [len - MUSIC_FADE, len], [1, 0], {
              extrapolateLeft: "clamp",
              extrapolateRight: "clamp",
              easing: Easing.inOut(Easing.sin),
            })
          }
        />
      </Sequence>
      {[...CMD].map((_, k) => (
        <Sfx key={k} from={TYPE_FROM + k * PER_KEY} src={`key${(k % 4) + 1}`} volume={0.14} />
      ))}
      <Sfx from={ENTER} src="enter" volume={0.22} />
    </>
  );
};

// Cada versão do trailer tem uma música. from: o tempo forte (em segundos da faixa) que cai no enter,
// achado com tools/beats.py. volume: iguala a loudness das duas.
export type Music = { src: string; from: number; volume: number };
export const MUSICS: Record<string, Music> = {
  warm: { src: "music/bed-warm.wav", from: 0, volume: 0.345 }, // earshot bed/warm
  "8bit": { src: "music/bed-warm-8bit.wav", from: 0, volume: 0.307 }, // a mesma partitura em chip (tools/warm8bit.py)
};

export const timing = linearTiming({ durationInFrames: T, easing: Easing.bezier(0.45, 0, 0.2, 1) });

export const Trailer: React.FC<{ music: Music }> = ({ music }) => (
  <AbsoluteFill style={{ background: C.page }}>
    <TransitionSeries>
      {SCENES.flatMap((sc, i) => {
        const e = enterOf(sc);
        const seq = (
          <TransitionSeries.Sequence key={sc.id} durationInFrames={sceneFrames(sc)} name={sc.id}>
            <SceneView sc={sc} />
          </TransitionSeries.Sequence>
        );
        if (i === 0 || e === "cut") return [seq];
        return [
          <TransitionSeries.Transition
            key={`t-${sc.id}`}
            presentation={e === "fade" ? fade() : slide({ direction: "from-right" })}
            timing={timing}
          />,
          seq,
        ];
      })}
    </TransitionSeries>
    <Opening music={music} outroAt={starts[SCENES.length - 1]} />
    {SCENES.map((sc, i) => (enterOf(sc) === "slide" ? <Sfx key={`w-${sc.id}`} from={starts[i]} src="whoosh" volume={0.2} /> : null))}
    {SCENES.flatMap((sc, i) => (sc.sfx ?? []).map((x, k) => <Sfx key={`${sc.id}${k}`} from={starts[i] + s(x.at)} src={x.src} volume={x.volume} />))}
  </AbsoluteFill>
);
