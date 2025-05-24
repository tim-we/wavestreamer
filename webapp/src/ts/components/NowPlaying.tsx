import type { FunctionComponent } from "preact";

type NPProps = {
  clip?: string;
};

const NowPlaying: FunctionComponent<NPProps> = ({ clip }) => (
  <section id="now">
    <div class="title">Now playing:</div>
    <div id="current-clip">{clip ? clip : "-"}</div>
  </section>
);

export default NowPlaying;
