import { C, mono } from "./theme";

// Janela do Terminal do macOS, com os três botões (fechar, minimizar, maximizar).
export const WIN = { width: 1640, height: 922, bar: 40, top: 22 };

type Props = {
  title: string;
  children: React.ReactNode;
  // posição e tamanho do conteúdo; sem eles, a janela grande centralizada
  left?: number;
  top?: number;
  width?: number;
  height?: number;
  bar?: number;
  style?: React.CSSProperties;
};

export const Window: React.FC<Props> = ({
  title,
  children,
  width = WIN.width,
  height = WIN.height,
  bar = WIN.bar,
  left = (1920 - width) / 2,
  top = WIN.top,
  style,
}) => {
  const k = bar / WIN.bar; // os botões e o título acompanham a altura da barra
  return (
    <div
      style={{
        position: "absolute",
        left,
        top,
        width,
        height: height + bar,
        borderRadius: 12 * Math.max(k, 0.7),
        overflow: "hidden",
        background: C.term,
        boxShadow: "0 30px 80px rgba(0,0,0,0.55), 0 0 0 1px rgba(255,255,255,0.08)",
        ...style,
      }}
    >
      <div style={{ height: bar, background: C.titlebar, position: "relative", display: "flex", alignItems: "center" }}>
        <div style={{ display: "flex", gap: 9 * k, marginLeft: 16 * k }}>
          {[C.red, C.yellow, C.lights].map((c) => (
            <div key={c} style={{ width: 13 * k, height: 13 * k, borderRadius: 7 * k, background: c }} />
          ))}
        </div>
        <div
          style={{
            position: "absolute",
            width: "100%",
            textAlign: "center",
            fontFamily: "-apple-system, 'SF Pro Text', 'Helvetica Neue', sans-serif",
            fontSize: 15 * k,
            fontWeight: 600,
            color: C.titleText,
          }}
        >
          {title}
        </div>
      </div>
      <div style={{ position: "relative", width, height, overflow: "hidden", fontFamily: mono }}>{children}</div>
    </div>
  );
};
