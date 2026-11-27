import { Composition } from "remotion";
import { MUSICS, Trailer, trailerDuration } from "./Trailer";
import { Trailer2, Trailer2Flow, trailer2Duration, trailer2FlowDuration } from "./Trailer2";

// Uma composição por música e por versão: trailer-*, trailer2-* e trailer2-flow-* (a música não para no brainrot).
const VERSIONS = [
  { id: "trailer", component: Trailer, duration: trailerDuration },
  { id: "trailer2", component: Trailer2, duration: trailer2Duration },
  { id: "trailer2-flow", component: Trailer2Flow, duration: trailer2FlowDuration },
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
