import { loadFont } from "@remotion/google-fonts/JetBrainsMono";
import { Easing, interpolate } from "remotion";

// A fonte que os clipes usam (o harness rasteriza com a JetBrains Mono).
export const mono = loadFont("normal", { weights: ["400", "700"], subsets: ["latin"] }).fontFamily;

// Cores do tema "sshiu" do app (internal/tui/ui.go) e da janela do Terminal do macOS.
export const C = {
  page: "#0a0a0b",
  term: "#111113", // o fundo dos clipes como sai no vídeo final (o defBg do harness, #121214, muda um pouco no h264)
  text: "#d8d8d8",
  accent: "#ff8700",
  muted: "#7a7a7a",
  green: "#5fd787",
  titlebar: "#2b2b2e",
  titleText: "#9a9a9e",
  red: "#ff5f57",
  yellow: "#febc2e",
  lights: "#28c840",
};

export const FPS = 30;
export const s = (sec: number) => Math.round(sec * FPS);

const ease = Easing.bezier(0.16, 1, 0.3, 1);

/** 0→1 entre os quadros a e b, com easing, fixo fora deles. */
export const ramp = (f: number, a: number, b: number) =>
  interpolate(f, [a, b], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: ease });
