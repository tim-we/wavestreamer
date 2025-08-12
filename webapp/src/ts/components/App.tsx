import type { FunctionComponent } from "preact";
import About from "./About";
import Controls from "./Controls";
import Header from "./Header";
import History from "./History";
import NowPlaying from "./NowPlaying";

const App: FunctionComponent = () => {
  return (
    <>
      <Header />
      <div id="content">
        <NowPlaying />
        <Controls />
        <History />
        <section id="stats" />
      </div>
      <div class="space-filler"></div>
      <About />
    </>
  );
};

export default App;
