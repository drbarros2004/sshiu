import { TransitionSeries, type TransitionPresentation, type TransitionPresentationComponentProps } from "@remotion/transitions";
import { fade } from "@remotion/transitions/fade";
import { slide } from "@remotion/transitions/slide";
import { AbsoluteFill, Easing, interpolate, OffthreadVideo, Sequence, staticFile, useCurrentFrame } from "remotion";
import { C, mono, ramp, s } from "./theme";
import { ENTER, Music, Opening, Outro, Sfx, Shot, shotFrames, T, timing, Title } from "./Trailer";
import { Window } from "./Window";

// trailer2: um título em cima de cada tela (sem numerar, como um trailer). A câmera só anda quando há uma coisa
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
  title?: string; // em cima da tela; *palavra* fica laranja
  card?: "title" | "outro";
  shots?: Shot[];
  grid?: { titles: string[] };
  strips?: Strip[];
  stages?: number[]; // segundos de cada estágio da obra nas faixas
  cam?: Cam;
  rings?: Ring[];
  counter?: { at: number; from: number; to: number }; // contador de moedas grande, só do trailer
  dur?: number;
  sfx?: { at: number; src: string; volume?: number }[];
};

const FULL: Cam[number] = [0, 1, 0.5, 0.5];
const PROMPT: Cam[number] = [0, 2.6, 0.17, 0.17]; // a pergunta da tarefa e o relógio, no alto à esquerda
const at = (p: Cam[number], t: number): Cam[number] => [t, p[1], p[2], p[3]];

// a obra: 4 estágios com corte seco; o primeiro passa metade na transição, o último segura
// a obra em 4 estágios; o último segura o clipe inteiro (2,3 s): a obra pronta respira antes do perfil
const buildStages = (sc: Scene, stages: number[]): Scene => ({
  ...sc,
  stages,
  dur: stages.reduce((a, b) => a + b, 0),
  sfx: stages.slice(0, 3).map((_, k) => ({ at: stages.slice(0, k + 1).reduce((a, b) => a + b, 0), src: "dialog", volume: 0.25 })),
});
// timelapse das roupas (capture/scene/timelapse.sh)
// os clipes têm 2,3 s; o último segura (a roupa nova respira antes da obra) e começa mais cedo para caber
const dressScene = (title: string, d: number[], lastFrom: number): Scene => ({
  id: "dress",
  enter: "slide",
  title,
  shots: d.map((x, k) => ({ clip: `tl-room-${k}`, from: k === 4 ? lastFrom : 0.6, to: (k === 4 ? lastFrom : 0.6) + x })),
  sfx: d.slice(0, 4).map((_, k) => ({ at: d.slice(0, k + 1).reduce((a, b) => a + b, 0), src: "toggle", volume: 0.35 })),
});

// 00-task: pede a tarefa em 1,45 s, digita de 2,8 a 4,2 s, "how long?" em 5,4 s e o relógio em 7,1 s.
// Os trechos parados saem com cortes que não se veem (a tela não muda neles).
const TASK: Shot[] = [
  { clip: "00-task", from: 1.2, to: 4.4 },
  { clip: "00-task", from: 5.2, to: 5.9 },
  { clip: "00-task", from: 6.9, to: 7.9 },
];
const typing = Array.from({ length: 16 }, (_, k) => ({ at: 1.6 + k * 0.0875, src: `key${(k % 4) + 1}`, volume: 0.1 }));

const SCENES: Scene[] = [
  { id: "title", card: "title", dur: 4.0 }, // a frase fica um pouco na tela antes da tarefa
  {
    id: "task",
    enter: "window",
    title: "say the *one* thing you'll do next.",
    shots: TASK,
    cam: [FULL, [0.6, 1, 0.5, 0.5], at(PROMPT, 1.4)],
    sfx: [...typing, { at: 3.3, src: "enter", volume: 0.18 }, { at: 3.95, src: "enter", volume: 0.18 }],
  },
  {
    id: "room",
    enter: "cut",
    title: "a room where everyone focuses *together*.",
    shots: [{ clip: "00-task", from: 7.9, to: 11.2 }],
    cam: [PROMPT, at(PROMPT, 0.3), [1.4, 1, 0.5, 0.5]],
  },
  {
    id: "garden",
    enter: "slide",
    title: "just a timer, your task and a *view*.",
    dur: 4.6, // o grid inteiro fica na tela antes do brainrot
    // garden-0..3: trechos do 02-garden-long (ver o README)
    grid: { titles: ["coffee", "tree", "night road", "sleeping cat"] },
    sfx: [0, 1, 2, 3].map((i) => ({ at: (T + 4 + i * 7) / 30, src: "dialog", volume: 0.18 })),
  },
  {
    id: "brainrot",
    enter: "window",
    title: "it's 2026. of course there's *brainrot*.",
    // a duração faz as moedas entrarem num bumbo da música (ver BREAK, mais abaixo)
    shots: [{ clip: "04-brainrot", from: 0, to: 80 / 30 }],
  },
  {
    id: "coins",
    enter: "window",
    title: "done? earn your *coins*.",
    // 05-coins-v2: a mesma "ship the trailer", quase no fim (TRAILER_FAST_TASK); f em 1,15 s. O feed fica
    // parado e aproximado; o contador grande à direita não existe no app, é do trailer.
    shots: [{ clip: "05-coins-v2", from: 0, to: 4.6 }],
    cam: [[0, 1.35, 0.36, 0.72]],
    counter: { at: 1.3, from: 240, to: 283 },
    sfx: [{ at: 1.6, src: "coin", volume: 0.45 }],
  },
  dressScene("spend them on hats. and pets. and *more hats*.", [1.4, 0.8, 0.8, 0.8, 2.2], 0.1),
  buildStages(
  {
      id: "build",
      enter: "slideUp",
      title: "focus together. *build* together.",
      strips: [
        { obra: "trem", title: "train", cx: 828 },
        { obra: "acampamento", title: "campsite", cx: 612 },
        { obra: "orquestra", title: "orchestra", cx: 532 },
      ],
    },
    [1.4, 1.0, 1.0, 2.3],
  ),
  {
    id: "profile",
    enter: "window",
    title: "see where your hours *actually* went.",
    shots: [{ clip: "08-profile-v2", from: 0.2, to: 0.2 + 179 / 30 }], // o fim acerta o logo num bumbo (END)
    // a tela inteira para situar; depois passa devagar pelo alto (o pet, o calendário e o mês) e desce
    // até a hora em que você foca, o que completa e as últimas tarefas (a "ship the trailer" no topo),
    // e a tela fica parada antes do fecho
    cam: [FULL, [1.2, 1, 0.5, 0.5], [2.2, 1.45, 0.355, 0.34], [3.1, 1.45, 0.355, 0.34], [4.2, 1.4, 0.628, 0.62]],
  },
  { id: "outro", enter: "fade", card: "outro", dur: 4.4, sfx: [{ at: 1.9, src: "shh-voice", volume: 0.6 }] },
];

// A versão flow é só o app: sem a tela das moedas (elas vão para o título das roupas) e sem o brainrot,
// com a música direto, sem quebra. As outras telas ganham o tempo, com respiro no fim de cada uma, e o
// fecho corta no mesmo bumbo da versão principal (34,27 s), com o mesmo golpe ecoando.
const FLOW_SCENES: Scene[] = SCENES.filter((sc) => sc.id !== "coins" && sc.id !== "brainrot").map((sc) => {
  switch (sc.id) {
    case "title":
      return { ...sc, dur: 4.3 }; // mais respiro na frase da abertura
    case "room":
      return { ...sc, shots: [{ clip: "00-task", from: 7.9, to: 11.9 }] };
    case "garden":
      return { ...sc, dur: 5.5 };
    case "dress":
      return dressScene("finish tasks, earn coins, dress your *pet*.", [1.4, 1.1, 1.1, 1.1, 2.3], 0);
    case "build":
      return buildStages(sc, [1.4, 1.4, 1.4, 2.3]);
    case "profile":
      // o perfil é quase parado: tocado um pouco mais devagar, para o passeio da câmera ter mais tempo
      return {
        ...sc,
        shots: [{ clip: "08-profile-v2", from: 0.2, to: 6.7, rate: 195 / 239 }],
        cam: [FULL, [1.4, 1, 0.5, 0.5], [2.6, 1.45, 0.355, 0.34], [3.8, 1.45, 0.355, 0.34], [5.2, 1.4, 0.628, 0.62]],
      };
    default:
      return sc;
  }
});

const sceneFrames = (sc: Scene) => (sc.shots ? sc.shots.reduce((n, x) => n + shotFrames(x), 0) : s(sc.dur!));
const enterOf = (sc: Scene) => sc.enter ?? "slide";

// No brainrot a música quebra (a fita para) e volta acelerando até as moedas, que entram num bumbo que as
// duas faixas têm em 14,27 s (o próximo é 16,05 s; tools/beats.py). No fecho, ela corta no bumbo de
// 34,27 s, quando o logo aparece, e esse golpe ecoa até o "shh". Os fx-*.wav saem de tools/musicfx.py
// nos segundos da faixa: se a conta mudar, gere de novo e acerte as durações. from: o segundo da faixa
// no enter.
const plan = (scenes: Scene[], from: number, withBreak: boolean, endKick: number) => {
  const starts = scenes.reduce<number[]>(
    (acc, sc, i) => [...acc, i === 0 ? 0 : acc[i - 1] + sceneFrames(scenes[i - 1]) - (enterOf(sc) === "cut" ? 0 : T)],
    [],
  );
  const idx = (id: string) => scenes.findIndex((sc) => sc.id === id);
  const onKick = (frame: number, kick: number, what: string) => {
    const at = from + (frame - ENTER) / 30;
    if (Math.abs(at - kick) > 0.04) throw new Error(`${what} em ${at} s da música, fora do bumbo (${kick} s)`);
  };
  const brk: [number, number] | undefined = withBreak ? [starts[idx("brainrot")], starts[idx("coins")]] : undefined;
  const end = starts[idx("outro")] + T + 4; // o logo do fecho já aparecendo
  if (brk) onKick(brk[1], 14.27, "as moedas entram");
  onKick(end, endKick, "o logo aparece");
  return { scenes, starts, from, brk, end, duration: starts[starts.length - 1] + sceneFrames(scenes[scenes.length - 1]) };
};
const MAIN = plan(SCENES, 0, true, 34.27);
const FLOW = plan(FLOW_SCENES, 0, false, 34.27);
export const trailer2Duration = MAIN.duration;
export const trailer2FlowDuration = FLOW.duration;

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

// O título, com as palavras entrando uma a uma.
const Head: React.FC<{ title: string }> = ({ title }) => {
  const f = useCurrentFrame();
  const words = title.split(" ");
  let hl = false;
  return (
    <div style={{ position: "absolute", left: LEFT, top: 82, fontFamily: mono, fontSize: 54, fontWeight: 700, color: C.text, whiteSpace: "nowrap", letterSpacing: -1 }}>
      {words.map((w, i) => {
        const start = hl || w.startsWith("*");
        hl = start && !w.replace(/^\*/, "").includes("*");
        const a = ramp(f, 4 + i * 2, 14 + i * 2);
        return (
          <span key={i} style={{ display: "inline-block", marginRight: "0.6em", color: start ? C.accent : C.text, opacity: a, translate: `0px ${(1 - a) * 12}px` }}>
            {w.replace(/\*/g, "")}
          </span>
        );
      })}
    </div>
  );
};

// Os dígitos grandes do app (bigText, internal/tui/ui.go): 3 colunas de meio-bloco, cada uma um pixel
// quadrado de largura por dois de altura. Desenhados aqui como pixels.
const DIGITS = [
  ["█▀█", "█ █", "▀▀▀"], ["▀█ ", " █ ", "▀▀▀"], ["▀▀█", "█▀▀", "▀▀▀"], ["▀▀█", " ▀█", "▀▀▀"], ["█ █", "▀▀█", "  ▀"],
  ["█▀▀", "▀▀█", "▀▀▀"], ["█▀▀", "█▀█", "▀▀▀"], ["▀▀█", "  █", "  ▀"], ["█▀█", "█▀█", "▀▀▀"], ["█▀█", "▀▀█", "▀▀▀"],
];
const PixelNumber: React.FC<{ n: number; px: number; color: string }> = ({ n, px, color }) => {
  const cells: [number, number][] = [];
  [...String(n)].forEach((d, i) =>
    DIGITS[+d].forEach((row, r) =>
      [...row].forEach((ch, c) => {
        if (ch === "█" || ch === "▀") cells.push([i * 4 + c, r * 2]);
        if (ch === "█" || ch === "▄") cells.push([i * 4 + c, r * 2 + 1]);
      }),
    ),
  );
  const w = String(n).length * 4 - 1;
  return (
    <div style={{ position: "relative", width: w * px, height: 5 * px }}>
      {cells.map(([x, y], k) => (
        <div key={k} style={{ position: "absolute", left: x * px, top: y * px, width: px, height: px, background: color }} />
      ))}
    </div>
  );
};

// O contador de moedas grande, à direita do feed (não existe no app, mas usa a moeda e os dígitos dele):
// aparece com o saldo antigo, conta até o novo e dá um pulinho no fim, com o "+43" subindo em cima.
const COIN = { light: "#fff3a0", mid: "#ffd75f", dark: "#d9a400" };
const Counter: React.FC<{ at: number; from: number; to: number }> = ({ at, from, to }) => {
  const f = useCurrentFrame();
  const t0 = s(at);
  const show = ramp(f, t0, t0 + 8);
  const p = interpolate(f, [t0 + 9, t0 + 42], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: Easing.out(Easing.cubic) });
  const end = t0 + 42;
  const bump = 1 + 0.08 * (ramp(f, end, end + 4) - ramp(f, end + 4, end + 12));
  const plus = ramp(f, t0 + 9, t0 + 18);
  const px = 13; // um pixel do app, nesta escala (a moeda tem 2×2 e cada dígito 3×5)
  return (
    <div style={{ position: "absolute", left: 980, top: 440, width: 400, display: "flex", flexDirection: "column", alignItems: "center", gap: 16, opacity: show, translate: `0px ${(1 - show) * 20}px` }}>
      <div style={{ fontFamily: mono, fontSize: 32, fontWeight: 700, color: C.green, opacity: plus, translate: `0px ${(1 - plus) * 24}px` }}>+{to - from}</div>
      <div style={{ display: "flex", alignItems: "center", gap: px * 2, scale: `${bump}` }}>
        <div style={{ display: "grid", gridTemplateColumns: `${px * 2}px ${px * 2}px`, gridTemplateRows: `${px * 2}px ${px * 2}px` }}>
          {[COIN.light, COIN.mid, COIN.mid, COIN.dark].map((c, i) => <div key={i} style={{ background: c }} />)}
        </div>
        <PixelNumber n={Math.round(from + (to - from) * p)} px={px} color={COIN.mid} />
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
const Strips: React.FC<{ strips: Strip[]; stages: number[] }> = ({ strips, stages }) => {
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
            {stages.map((d, k) => {
              const el = (
                <Sequence key={k} from={s(from)} durationInFrames={s(d)} name={`${st.obra}-${k}`}>
                  <OffthreadVideo
                    src={staticFile(`clips/obra-${st.obra}-${k}.mp4`)}
                    trimBefore={0}
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
      {sc.title && <Head title={sc.title} />}
      {sc.grid && <Grid {...sc.grid} />}
      {sc.strips && <Strips strips={sc.strips} stages={sc.stages!} />}
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
          {sc.counter && <Counter {...sc.counter} />}
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

const Cut: React.FC<{ music: Music; p: ReturnType<typeof plan> }> = ({ music, p }) => (
  <AbsoluteFill style={{ background: C.page }}>
    <TransitionSeries>
      {p.scenes.flatMap((sc, i) => {
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
    <Opening music={{ ...music, from: p.from }} outroAt={p.starts[p.scenes.length - 1]} pause={p.brk} end={p.end} />
    {p.scenes.map((sc, i) => {
      const x = i > 0 ? ENTER_SFX[enterOf(sc)] : undefined;
      return x ? <Sfx key={`e-${sc.id}`} from={p.starts[i]} src={x.src} volume={x.volume} /> : null;
    })}
    {p.scenes.flatMap((sc, i) => (sc.sfx ?? []).map((x, k) => <Sfx key={`${sc.id}${k}`} from={p.starts[i] + s(x.at)} src={x.src} volume={x.volume} />))}
  </AbsoluteFill>
);

export const Trailer2: React.FC<{ music: Music }> = ({ music }) => <Cut music={music} p={MAIN} />;
// flow: a música não para no brainrot e não há a tela das moedas
export const Trailer2Flow: React.FC<{ music: Music }> = ({ music }) => <Cut music={music} p={FLOW} />;
