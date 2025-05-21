import { Component } from "preact";
import WavestreamerApi, { HistoryEntry } from "../wavestreamer-api";
import About from "./About";
import Controls from "./Controls";
import Header from "./Header";
import History from "./History";
import NowPlaying from "./NowPlaying";

type AppProps = {
    radio: WavestreamerApi;
};

type AppState = {
    now?: { current: string; history: HistoryEntry[] };
};

export default class App extends Component<AppProps, AppState> {
    public constructor(props: AppProps) {
        super(props);

        const radio = props.radio;

        radio.on("update", (now) => this.setState({ now }));
    }

    public render() {
        const props = this.props;
        const state = this.state;

        return (
            <>
                <Header />
                <div id="content">
                    <NowPlaying clip={state.now?.current} />
                    <Controls radio={props.radio} />
                    <History data={state.now ? state.now.history : []} />
                    <section id="stats"/>
                </div>
                <About />
            </>
        );
    }

    private async update() {
        if (document.visibilityState !== "visible") {
            return;
        }

        const info = await this.props.radio.now();

        this.setState({
            now: {
                current: info.current,
                history: info.history,
            },
        });
    }
}
