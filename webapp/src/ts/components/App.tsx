import type { FunctionComponent } from "preact";
import WavestreamerApi, { HistoryEntry } from "../wavestreamer-api";
import About from "./About";
import Controls from "./Controls";
import Header from "./Header";
import History from "./History";
import NowPlaying from "./NowPlaying";
import { useEffect, useState } from "preact/hooks";

type AppProps = {
  radio: WavestreamerApi;
};

type NowInfo = { current: string; history: HistoryEntry[] };

const App: FunctionComponent<AppProps> = ({ radio }) => {
  const [now, setNow] = useState<NowInfo | undefined>();

  useEffect(() => radio.on("update", setNow), [radio]);

  return (
    <>
      <Header />
      <div id="content">
        <NowPlaying clip={now?.current} />
        <Controls radio={radio} />
        <History data={now ? now.history : []} />
        <section id="stats" />
      </div>
      <About />
    </>
  );
};

export default App;
