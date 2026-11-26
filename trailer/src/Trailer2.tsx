import { TransitionSeries } from "@remotion/transitions";
import { slide } from "@remotion/transitions/slide";
import { AbsoluteFill, Easing, interpolate, OffthreadVideo, Sequence, staticFile, useCurrentFrame } from "remotion";
import { C, mono, ramp, s } from "./theme";
import { cuts, moments, Music, Opening, Outro, Sfx, Shot, shotFrames, T, timing, Title } from "./Trailer";
import { Window } from "./Window";

// trailer2: a mesma ordem do trailer, com um passo numerado e um título em cima de cada tela, e a
// câmera andando dentro da tela do app (zoom no que importa em cada cena).

// Câmera: [segundo da cena, zoom, x, y], com x/y = o ponto da tela do clipe (0–1) que fica no centro.
// Entre dois pontos ela anda com easing; para segurar, repita o ponto.
type Cam = [number, number, number, number][];
// Destaque: uma moldura laranja em volta de [x, y, w, h], em pixels do clipe (1920×1080).
type Ring = { from: number; to: number; box: [number, number, number, number] };
type Scene = {
  id: string;
  step?: string; // o passo, no selo: "01 · ssh in"
  title?: string; // *palavra* fica laranja
  card?: "title" | "outro";
  shots?: Shot[];
  grid?: { clip: string; from: number[]; titles: string[] };
  cam?: Cam;
  rings?: Ring[];
  dur?: number;
  sfx?: { at: number; src: string; volume?: number }[];
};

const FULL: Cam[number] = [0, 1, 0.5, 0.5];

const SCENES: Scene[] = [
  { id: "title", card: "title", dur: 4.0 },
  {
    id: "room",
    step: "01 · ssh in",
    title: "pull up a chair. the room is *already* focusing.",
    shots: [{ clip: "01-room", from: 0.3, to: 4.3 }],
    // aproxima nos três que já estão focando (ana, bruno e duda)
    cam: [FULL, [1.0, 1, 0.5, 0.5], [2.6, 1.9, 0.39, 0.45]],
  },
  {
    id: "garden",
    step: "02 · focus",
    title: "the screen goes *quiet* while you work.",
    dur: 4.2,
    // garden-0..3: os trechos de 12,3 s, 16,3 s, 20,3 s e 28,3 s do 02-garden-long, cortados à parte (quadro-chave a cada
    // 0,5 s) porque quatro buscas no clipe longo derrubavam o compositor
    grid: { clip: "garden", from: [0, 0, 0, 0], titles: ["coffee", "tree", "night road", "sleeping cat"] },
    sfx: [0, 1, 2, 3].map((i) => ({ at: (T + 4 + i * 7) / 30, src: "dialog", volume: 0.2 })),
  },
  {
    id: "brainrot",
    step: "02 · focus, for real devs",
    title: "or, if you insist, *brainrot*.",
    shots: [{ clip: "04-brainrot", from: 0, to: 4.0 }],
    cam: [FULL, [0.8, 1, 0.5, 0.5], [4.0, 1.14, 0.5, 0.42]],
    sfx: [{ at: 0, src: "glitch", volume: 0.45 }],
  },
  {
    id: "coins",
    step: "03 · ship it",
    title: "finish the task. get paid in *coins*.",
    shots: [{ clip: "05-coins", from: 0.3, to: 4.1 }],
    // a linha "+43 coins gained" e depois o contador lá em cima
    cam: [FULL, [0.9, 1, 0.5, 0.5], [1.6, 1.9, 0.3, 0.74], [2.4, 1.9, 0.3, 0.74], [3.1, 1.9, 0.86, 0.12]],
    rings: [
      { from: 1.6, to: 2.5, box: [508, 788, 222, 34] },
      { from: 3.1, to: 3.8, box: [1814, 90, 92, 36] },
    ],
    sfx: [{ at: 0.85, src: "coin", volume: 0.45 }],
  },
  {
    id: "dress",
    step: "04 · spend it",
    title: "on hats, colors, a *whole new pet*.",
    shots: moments("tl-room", 0.6, 1.6, 1.0, 1.8),
    // passeia pelos cards, da esquerda para a direita
    cam: [[0, 1.4, 0.36, 0.47], [6.4, 1.4, 0.6, 0.47]],
    sfx: cuts(1.6, 1.0).map((at) => ({ at, src: "dialog", volume: 0.25 })),
  },
  {
    id: "build",
    step: "05 · build",
    title: "everyone's focus builds *the train*.",
    shots: moments("tl-mural", 0.4, 1.6, 0.9, 1.9),
    cam: [[0, 1.12, 0.44, 0.22], [6.1, 1.22, 0.44, 0.22]],
    sfx: cuts(1.6, 0.9).map((at) => ({ at, src: "toggle", volume: 0.35 })),
  },
  {
    id: "profile",
    step: "06 · look back",
    title: "see where your hours *actually* went.",
    // o perfil é quase parado: um pouco mais lento para dar tempo da câmera andar
    shots: [{ clip: "08-profile", from: 0.2, to: 3.3, rate: 0.85 }],
    cam: [FULL, [0.9, 1, 0.5, 0.5], [1.6, 1.65, 0.62, 0.2], [2.4, 1.65, 0.62, 0.2], [3.1, 1.65, 0.62, 0.56]],
  },
  { id: "outro", card: "outro", dur: 3.6, sfx: [{ at: 1.3, src: "shh-voice", volume: 0.6 }] },
];

const sceneFrames = (sc: Scene) => (sc.shots ? sc.shots.reduce((n, x) => n + shotFrames(x), 0) : s(sc.dur!));
const starts = SCENES.reduce<number[]>((acc, sc, i) => [...acc, i === 0 ? 0 : acc[i - 1] + sceneFrames(SCENES[i - 1]) - T], []);
export const trailer2Duration = starts[starts.length - 1] + sceneFrames(SCENES[SCENES.length - 1]);

// layout: selo e título em cima, a janela do app embaixo
const APP = { width: 1440, height: 810, bar: 36, top: 200 };
const LEFT = (1920 - APP.width) / 2;
const GRID = { w: 652, h: 367, bar: 28, gapX: 24, gapY: 18, top: 200 };

const ease = Easing.bezier(0.65, 0, 0.35, 1);

// Tudo dentro da janela, em pixels do clipe: a câmera move esta camada inteira.
const Camera: React.FC<{ cam?: Cam; rings?: Ring[]; children: React.ReactNode }> = ({ cam, rings, children }) => {
  const f = useCurrentFrame();
  const c = cam ?? [FULL];
  const at = (i: number) => (c.length === 1 ? c[0][i] : interpolate(f, c.map((p) => s(p[0])), c.map((p) => p[i]), { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: ease }));
  const [z, x, y] = [at(1), at(2), at(3)];
  const k = APP.width / 1920;
  return (
    <div
      style={{
        position: "absolute",
        width: 1920,
        height: 1080,
        transformOrigin: "0 0",
        transform: `scale(${k}) translate(960px, 540px) scale(${z}) translate(${-x * 1920}px, ${-y * 1080}px)`,
      }}
    >
      {children}
      {(rings ?? []).map((r, i) => {
        const a = Math.min(ramp(f, s(r.from), s(r.from) + 8), 1 - ramp(f, s(r.to) - 6, s(r.to)));
        const pad = 10;
        return (
          <div
            key={i}
            style={{
              position: "absolute",
              left: r.box[0] - pad,
              top: r.box[1] - pad,
              width: r.box[2] + pad * 2,
              height: r.box[3] + pad * 2,
              border: `${5 / z}px solid ${C.accent}`,
              borderRadius: 10,
              opacity: a,
              scale: `${1.15 - 0.15 * a}`,
            }}
          />
        );
      })}
    </div>
  );
};

// O selo do passo e o título, com as palavras entrando uma a uma.
const Head: React.FC<{ step: string; title: string }> = ({ step, title }) => {
  const f = useCurrentFrame();
  const [n, label] = step.split(" · ");
  const words = title.split(" ");
  let hl = false;
  return (
    <div style={{ position: "absolute", left: LEFT, top: 44, fontFamily: mono }}>
      <div
        style={{
          display: "inline-block",
          fontSize: 26,
          fontWeight: 700,
          padding: "8px 14px",
          borderRadius: 8,
          background: "#1f1f22",
          color: C.muted,
          opacity: ramp(f, 4, 14),
          translate: `${(1 - ramp(f, 4, 14)) * -16}px 0px`,
        }}
      >
        <span style={{ color: C.accent }}>{n}</span> · {label}
      </div>
      <div style={{ marginTop: 18, fontSize: 50, fontWeight: 700, color: C.text, whiteSpace: "nowrap", letterSpacing: -1 }}>
        {words.map((w, i) => {
          const start = hl || w.startsWith("*");
          hl = start && !w.replace(/^\*/, "").includes("*");
          const a = ramp(f, 10 + i * 2, 20 + i * 2);
          return (
            <span key={i} style={{ display: "inline-block", marginRight: "0.6em", color: start ? C.accent : C.text, opacity: a, translate: `0px ${(1 - a) * 12}px` }}>
              {w.replace(/\*/g, "")}
            </span>
          );
        })}
      </div>
    </div>
  );
};

const Grid: React.FC<{ clip: string; from: number[]; titles: string[] }> = ({ clip, from, titles }) => {
  const f = useCurrentFrame();
  const left0 = (1920 - (GRID.w * 2 + GRID.gapX)) / 2;
  return (
    <>
      {from.map((t, i) => {
        const a = ramp(f, T + 4 + i * 7, T + 14 + i * 7);
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
            <OffthreadVideo src={staticFile(`clips/${clip}-${i}.mp4`)} trimBefore={Math.round(t * 30)} muted style={{ width: GRID.w, height: GRID.h }} />
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
      {sc.step && <Head step={sc.step} title={sc.title!} />}
      {sc.grid && <Grid {...sc.grid} />}
      {sc.shots && (
        <Window title="maya — ssh sshiu — 160×45" {...APP} left={LEFT}>
          <Camera cam={sc.cam} rings={sc.rings}>
            {sc.shots.map((x, j) => {
              const d = shotFrames(x);
              const el = (
                <Sequence key={j} from={at} durationInFrames={d} name={x.clip}>
                  <OffthreadVideo
                    src={staticFile(`clips/${x.clip}.mp4`)}
                    trimBefore={Math.round(x.from * 30)}
                    playbackRate={x.rate ?? 1}
                    muted
                    style={{ width: 1920, height: 1080 }}
                  />
                </Sequence>
              );
              at += d;
              return el;
            })}
          </Camera>
        </Window>
      )}
    </AbsoluteFill>
  );
};

export const Trailer2: React.FC<{ music: Music }> = ({ music }) => (
  <AbsoluteFill style={{ background: C.page }}>
    <TransitionSeries>
      {SCENES.flatMap((sc, i) => {
        const seq = (
          <TransitionSeries.Sequence key={sc.id} durationInFrames={sceneFrames(sc)} name={sc.id}>
            <SceneView sc={sc} />
          </TransitionSeries.Sequence>
        );
        if (i === 0) return [seq];
        return [<TransitionSeries.Transition key={`t-${sc.id}`} presentation={slide({ direction: "from-right" })} timing={timing} />, seq];
      })}
    </TransitionSeries>
    <Opening music={music} outroAt={starts[SCENES.length - 1]} />
    {SCENES.map((sc, i) => (i > 0 ? <Sfx key={`w-${sc.id}`} from={starts[i]} src="whoosh" volume={0.2} /> : null))}
    {SCENES.flatMap((sc, i) => (sc.sfx ?? []).map((x, k) => <Sfx key={`${sc.id}${k}`} from={starts[i] + s(x.at)} src={x.src} volume={x.volume} />))}
  </AbsoluteFill>
);
