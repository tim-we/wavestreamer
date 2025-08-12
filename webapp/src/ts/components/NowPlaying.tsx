import type { FunctionComponent } from "preact";
import { nowDataSignal } from "../wavestreamer-api";

const NowPlaying: FunctionComponent<unknown> = () => {
  const clip = nowDataSignal.value?.current ?? "-";

  return (
    <section id="now">
      <div class="title">Now playing:</div>
      <div id="current-clip">{clip}</div>
    </section>
  );
};

export default NowPlaying;
