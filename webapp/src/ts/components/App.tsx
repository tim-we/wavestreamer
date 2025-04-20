import { Component } from "preact";
import PyRadio, { HistoryEntry } from "../wavestreamer";
import About from "./About";
import Controls from "./Controls";
import Header from "./Header";
import History from "./History";
import NowPlaying from "./NowPlaying";

type AppProps = {
    radio: PyRadio;
};

type AppState = {
    now?: { current: string; history: HistoryEntry[] };
    extensions: { name: string; command: string }[];
};

export default class App extends Component<AppProps, AppState> {
    public constructor(props: AppProps) {
        super(props);
        this.state = { extensions: [] };

        const radio = props.radio;

        radio.on("update", (now) => this.setState({ now }));
        radio.extensions().then((extensions) => this.setState({ extensions }));
    }

    public render() {
        const props = this.props;
        const state = this.state;

        return (
            <>
                <Header />
                <div id="content">
                    <NowPlaying clip={state.now?.current} />
                    <Controls
                        radio={props.radio}
                        extensions={state.extensions}
                    />
                    <History data={state.now ? state.now.history : []} />
                    <div id="stats"></div>
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
