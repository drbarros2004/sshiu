import { TransitionSeries, type TransitionPresentation, type TransitionPresentationComponentProps } from "@remotion/transitions";
import { fade } from "@remotion/transitions/fade";
import { slide } from "@remotion/transitions/slide";
import { AbsoluteFill, Easing, interpolate, OffthreadVideo, Sequence, staticFile, useCurrentFrame } from "remotion";
import { C, mono, ramp, s } from "./theme";
import { Music, Opening, Outro, Sfx, Shot, shotFrames, T, timing, Title } from "./Trailer";
import { Window } from "./Window";

// trailer2: um passo numerado e um título em cima de cada tela. A câmera só anda quando há uma coisa
// específica para mostrar (a tarefa sendo digitada, as moedas); timelapses, brainrot e perfil ficam
// parados, e o perfil ganha molduras em vez de zoom.

// Câmera: [segundo da cena, zoom, x, y], com x/y = o ponto da tela do clipe (0–1) que fica no centro.
// Entre dois pontos ela anda com easing; para segurar, repita o ponto.
type Cam = [number, number, number, number][];
// Destaque: uma moldura laranja em volta de [x, y, w, h], em pixels do clipe (1920×1080).
type Ring = { from: number; to: number; box: [number, number, number, number] };
// Como a cena entra: cut = corte seco (a janela fica, só o título troca); window = a tela nova abre como
// uma janela do Mac; slide = desliza da direita (ou de baixo, slideUp); fade só no fecho.
type Enter = "cut" | "window" | "slide" | "slideUp" | "fade";
// Uma faixa: a obra recortada do clipe, numa janela estreita. cx = o centro da obra no clipe, em px.
type Strip = { obra: string; title: string; cx: number };
type Scene = {
  id: string;
  enter?: Enter;
  step?: string; // o passo, no selo: "01 · what's next?"
  title?: string; // *palavra* fica laranja
  card?: "title" | "outro";
  shots?: Shot[];
  grid?: { titles: string[] };
  strips?: Strip[];
  cam?: Cam;
  rings?: Ring[];
  dur?: number;
  sfx?: { at: number; src: string; volume?: number }[];
};

const FULL: Cam[number] = [0, 1, 0.5, 0.5];
const PROMPT: Cam[number] = [0, 2.6, 0.17, 0.17]; // a pergunta da tarefa e o relógio, no alto à esquerda
const at = (p: Cam[number], t: number): Cam[number] => [t, p[1], p[2], p[3]];

// a obra: 4 estágios com corte seco; o primeiro passa metade na transição, o último segura
const STAGES = [1.4, 1.0, 1.0, 1.8];
const stageCuts = STAGES.slice(0, 3).map((_, k) => STAGES.slice(0, k + 1).reduce((a, b) => a + b, 0));
// timelapse das roupas (capture/scene/timelapse.sh)
const DRESS = [1.4, 0.9, 0.9, 0.9, 1.5];
const dressCuts = DRESS.slice(0, 4).map((_, k) => DRESS.slice(0, k + 1).reduce((a, b) => a + b, 0));

// 00-task: pede a tarefa em 1,45 s, digita de 2,8 a 4,2 s, "how long?" em 5,4 s e o relógio em 7,1 s.
// Os trechos parados saem com cortes que não se veem (a tela não muda neles).
const TASK: Shot[] = [
  { clip: "00-task", from: 1.0, to: 4.4 },
  { clip: "00-task", from: 5.2, to: 5.9 },
  { clip: "00-task", from: 6.8, to: 7.9 },
];
const typing = Array.from({ length: 16 }, (_, k) => ({ at: 1.8 + k * 0.0875, src: `key${(k % 4) + 1}`, volume: 0.1 }));

const SCENES: Scene[] = [
  { id: "title", card: "title", dur: 4.0 },
  {
    id: "task",
    enter: "window",
    step: "01 · what's next?",
    title: "say the *one* thing you'll do next.",
    shots: TASK,
    cam: [FULL, [0.6, 1, 0.5, 0.5], at(PROMPT, 1.4)],
    sfx: [...typing, { at: 3.5, src: "enter", volume: 0.18 }, { at: 4.25, src: "enter", volume: 0.18 }],
  },
  {
    id: "room",
    enter: "cut",
    step: "02 · focus",
    title: "then focus. *everyone* else already is.",
    shots: [{ clip: "00-task", from: 7.9, to: 10.9 }],
    cam: [PROMPT, at(PROMPT, 0.3), [1.6, 1, 0.5, 0.5]],
  },
  {
    id: "garden",
    enter: "slide",
    step: "03 · zone out",
    title: "just a timer, your task and a *view*.",
    dur: 4.0,
    // garden-0..3: trechos do 02-garden-long (ver o README)
    grid: { titles: ["coffee", "tree", "night road", "sleeping cat"] },
    sfx: [0, 1, 2, 3].map((i) => ({ at: (T + 4 + i * 7) / 30, src: "dialog", volume: 0.18 })),
  },
  {
    id: "brainrot",
    enter: "cut",
    step: "03 · zone out, for real devs",
    title: "or, if you insist, *brainrot*.",
    shots: [{ clip: "04-brainrot", from: 0, to: 3.2 }],
    sfx: [{ at: 0, src: "glitch", volume: 0.45 }],
  },
  {
    id: "coins",
    enter: "window",
    step: "04 · ship it",
    title: "done? earn your *coins*.",
    // 05-coins-v2: a mesma "ship the trailer", quase no fim (TRAILER_FAST_TASK); f em 1,15 s
    shots: [{ clip: "05-coins-v2", from: 0.3, to: 5.3 }],
    cam: [FULL, [0.9, 1, 0.5, 0.5], [1.5, 1.9, 0.3, 0.74], [2.9, 1.9, 0.3, 0.74], [3.6, 2.2, 0.9, 0.1]],
    rings: [
      { from: 1.5, to: 2.9, box: [514, 789, 212, 30] },
      { from: 3.6, to: 5.0, box: [1816, 93, 88, 30] },
    ],
    sfx: [{ at: 0.85, src: "coin", volume: 0.45 }],
  },
  {
    id: "dress",
    enter: "slide",
    step: "05 · spend it",
    title: "on hats, colors, a *whole new pet*.",
    shots: DRESS.map((d, k) => ({ clip: `tl-room-${k}`, from: 0.6, to: 0.6 + d })),
    sfx: dressCuts.map((at) => ({ at, src: "toggle", volume: 0.35 })),
  },
  {
    id: "build",
    enter: "slideUp",
    step: "06 · build",
    title: "every focused minute grows the room's *mural*.",
    dur: STAGES.reduce((a, b) => a + b, 0),
    strips: [
      { obra: "trem", title: "train", cx: 828 },
      { obra: "acampamento", title: "campsite", cx: 612 },
      { obra: "orquestra", title: "orchestra", cx: 532 },
    ],
    sfx: stageCuts.map((at) => ({ at, src: "dialog", volume: 0.25 })),
  },
  {
    id: "profile",
    enter: "window",
    step: "07 · look back",
    title: "see where your hours *actually* went.",
    shots: [{ clip: "08-profile-v2", from: 0.2, to: 4.7 }],
    // a tarefa do trailer entre as últimas, o mês e a hora em que você foca, uma de cada vez
    rings: [
      { from: 1.0, to: 2.1, box: [535, 745, 790, 26] },
      { from: 2.1, to: 3.2, box: [1008, 122, 440, 96] },
      { from: 3.2, to: 4.4, box: [530, 392, 1356, 100] },
    ],
  },
  { id: "outro", enter: "fade", card: "outro", dur: 3.6, sfx: [{ at: 1.3, src: "shh-voice", volume: 0.6 }] },
];

const sceneFrames = (sc: Scene) => (sc.shots ? sc.shots.reduce((n, x) => n + shotFrames(x), 0) : s(sc.dur!));
const enterOf = (sc: Scene) => sc.enter ?? "slide";
const starts = SCENES.reduce<number[]>(
  (acc, sc, i) => [...acc, i === 0 ? 0 : acc[i - 1] + sceneFrames(SCENES[i - 1]) - (enterOf(sc) === "cut" ? 0 : T)],
  [],
);
export const trailer2Duration = starts[starts.length - 1] + sceneFrames(SCENES[SCENES.length - 1]);

// layout: selo e título em cima, a janela do app embaixo
const APP = { width: 1440, height: 810, bar: 36, top: 200 };
const LEFT = (1920 - APP.width) / 2;
const GRID = { w: 652, h: 367, bar: 28, gapX: 24, gapY: 18, top: 200 };
// faixas da obra: o alto do mural (y 170–392 no clipe), em escala 0,9
const STRIP = { w: 1500, top: 236, gap: 26, bar: 30, y0: 170, y1: 392, k: 0.9 };

const ease = Easing.bezier(0.65, 0, 0.35, 1);

// A tela nova abre como uma janela do Mac: cresce um pouco e aparece; a velha só escurece.
const WindowPop: React.FC<TransitionPresentationComponentProps<Record<string, never>>> = ({
  children,
  presentationDirection,
  presentationProgress: p,
}) =>
  presentationDirection === "entering" ? (
    <AbsoluteFill style={{ opacity: p, scale: `${0.88 + 0.12 * p}` }}>{children}</AbsoluteFill>
  ) : (
    <AbsoluteFill style={{ opacity: 1 - 0.7 * p }}>{children}</AbsoluteFill>
  );
const windowPop = (): TransitionPresentation<Record<string, never>> => ({ component: WindowPop, props: {} });

// cada apresentação tem as suas props; o TransitionSeries aceita qualquer uma
const presentation = (e: Enter) =>
  (e === "window" ? windowPop() : e === "fade" ? fade() : slide({ direction: e === "slideUp" ? "from-bottom" : "from-right" })) as unknown as ReturnType<typeof fade>;

// Tudo dentro da janela, em pixels do clipe: a câmera move esta camada inteira.
const Camera: React.FC<{ cam?: Cam; rings?: Ring[]; children: React.ReactNode }> = ({ cam, rings, children }) => {
  const f = useCurrentFrame();
  const c = cam ?? [FULL];
  const val = (i: number) =>
    c.length === 1
      ? c[0][i]
      : interpolate(f, c.map((p) => s(p[0])), c.map((p) => p[i]), { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: ease });
  const [z, x, y] = [val(1), val(2), val(3)];
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
              scale: `${1.06 - 0.06 * a}`,
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
  const [n, ...rest] = step.split(" · ");
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
        <span style={{ color: C.accent }}>{n}</span> · {rest.join(" · ")}
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

const opening = (f: number, i: number) => ramp(f, T + 4 + i * 7, T + 14 + i * 7); // janelinha i abrindo

// Grid do jardim: uma janelinha do Mac por wallpaper (garden-0..3), abrindo uma depois da outra.
const Grid: React.FC<{ titles: string[] }> = ({ titles }) => {
  const f = useCurrentFrame();
  const left0 = (1920 - (GRID.w * 2 + GRID.gapX)) / 2;
  return (
    <>
      {titles.map((t, i) => {
        const a = opening(f, i);
        return (
          <Window
            key={i}
            title={t}
            left={left0 + (i % 2) * (GRID.w + GRID.gapX)}
            top={GRID.top + Math.floor(i / 2) * (GRID.h + GRID.bar + GRID.gapY)}
            width={GRID.w}
            height={GRID.h}
            bar={GRID.bar}
            style={{ opacity: a, scale: `${0.92 + 0.08 * a}` }}
          >
            <OffthreadVideo src={staticFile(`clips/garden-${i}.mp4`)} muted style={{ width: GRID.w, height: GRID.h }} />
          </Window>
        );
      })}
    </>
  );
};

// Faixas da obra: cada uma é o mural de uma sala com outra obra, crescendo nos 4 estágios com corte seco.
const Strips: React.FC<{ strips: Strip[] }> = ({ strips }) => {
  const f = useCurrentFrame();
  const h = (STRIP.y1 - STRIP.y0) * STRIP.k;
  return (
    <>
      {strips.map((st, i) => {
        const a = opening(f, i);
        let from = 0;
        return (
          <Window
            key={st.obra}
            title={st.title}
            left={(1920 - STRIP.w) / 2}
            top={STRIP.top + i * (h + STRIP.bar + STRIP.gap)}
            width={STRIP.w}
            height={h}
            bar={STRIP.bar}
            style={{ opacity: a, scale: `${0.94 + 0.06 * a}` }}
          >
            {STAGES.map((d, k) => {
              const el = (
                <Sequence key={k} from={s(from)} durationInFrames={s(d)} name={`${st.obra}-${k}`}>
                  <OffthreadVideo
                    src={staticFile(`clips/obra-${st.obra}-${k}.mp4`)}
                    trimBefore={6}
                    muted
                    style={{
                      position: "absolute",
                      width: 1920 * STRIP.k,
                      height: 1080 * STRIP.k,
                      left: STRIP.w / 2 - st.cx * STRIP.k,
                      top: -STRIP.y0 * STRIP.k,
                      maxWidth: "none",
                    }}
                  />
                </Sequence>
              );
              from += d;
              return el;
            })}
          </Window>
        );
      })}
    </>
  );
};

const SceneView: React.FC<{ sc: Scene }> = ({ sc }) => {
  let pos = 0;
  return (
    <AbsoluteFill style={{ background: C.page }}>
      {sc.card === "title" && <Title />}
      {sc.card === "outro" && <Outro />}
      {sc.step && <Head step={sc.step} title={sc.title!} />}
      {sc.grid && <Grid {...sc.grid} />}
      {sc.strips && <Strips strips={sc.strips} />}
      {sc.shots && (
        <Window title="maya — ssh sshiu — 160×45" {...APP} left={LEFT}>
          <Camera cam={sc.cam} rings={sc.rings}>
            {sc.shots.map((x, j) => {
              const d = shotFrames(x);
              const el = (
                <Sequence key={j} from={pos} durationInFrames={d} name={x.clip}>
                  <OffthreadVideo
                    src={staticFile(`clips/${x.clip}.mp4`)}
                    trimBefore={Math.round(x.from * 30)}
                    playbackRate={x.rate ?? 1}
                    muted
                    style={{ width: 1920, height: 1080 }}
                  />
                </Sequence>
              );
              pos += d;
              return el;
            })}
          </Camera>
        </Window>
      )}
    </AbsoluteFill>
  );
};

// som de cada tipo de transição: o whoosh baixinho só no slide, um swish suave na janela
const ENTER_SFX: Partial<Record<Enter, { src: string; volume: number }>> = {
  slide: { src: "whoosh", volume: 0.07 },
  slideUp: { src: "whoosh", volume: 0.07 },
  window: { src: "swish", volume: 0.1 },
};

export const Trailer2: React.FC<{ music: Music }> = ({ music }) => (
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
        return [<TransitionSeries.Transition key={`t-${sc.id}`} presentation={presentation(e)} timing={timing} />, seq];
      })}
    </TransitionSeries>
    <Opening music={music} outroAt={starts[SCENES.length - 1]} />
    {SCENES.map((sc, i) => {
      const x = i > 0 ? ENTER_SFX[enterOf(sc)] : undefined;
      return x ? <Sfx key={`e-${sc.id}`} from={starts[i]} src={x.src} volume={x.volume} /> : null;
    })}
    {SCENES.flatMap((sc, i) => (sc.sfx ?? []).map((x, k) => <Sfx key={`${sc.id}${k}`} from={starts[i] + s(x.at)} src={x.src} volume={x.volume} />))}
  </AbsoluteFill>
);
