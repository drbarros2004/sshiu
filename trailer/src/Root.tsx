import { Composition } from "remotion";
import { MUSICS, Trailer, trailerDuration } from "./Trailer";
import { Trailer2, trailer2Duration } from "./Trailer2";

// Uma composição por música e por versão: trailer-warm, trailer-8bit, trailer2-warm, trailer2-8bit.
const VERSIONS = [
  { id: "trailer", component: Trailer, duration: trailerDuration },
  { id: "trailer2", component: Trailer2, duration: trailer2Duration },
];

export const RemotionRoot = () => (
  <>
    {VERSIONS.flatMap((v) =>
      Object.entries(MUSICS).map(([id, music]) => (
        <Composition
          key={`${v.id}-${id}`}
          id={`${v.id}-${id}`}
          component={v.component}
          defaultProps={{ music }}
          durationInFrames={v.duration}
          fps={30}
          width={1920}
          height={1080}
        />
      )),
    )}
  </>
);
