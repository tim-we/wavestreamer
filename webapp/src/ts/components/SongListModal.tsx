import { render, Component, createRef } from "preact";
import PyRadio from "../wavestreamer";
import { unmountComponentAtNode } from "preact/compat";

const portal = document.getElementById("modal-portal")!;

export function show(radio: PyRadio) {
    render(<SongListModal radio={radio} />, portal);
}

type SLMProps = {
    radio: PyRadio;
};

type SLMState = {
    clips: string[];
    expandedClip?: string;
};

class SongListModal extends Component<SLMProps, SLMState> {
    private inputRef = createRef<HTMLInputElement>();

    public constructor(props: SLMProps) {
        super(props);
        this.state = { clips: [] };
        this.escapeHandler = this.escapeHandler.bind(this);
    }

    public render() {
        const radio = this.props.radio;

        return (
            <div
                id="song-list-container"
                className="show"
                onClick={(e) => {
                    e.stopPropagation();
                    this.close();
                }}
            >
                <div id="song-list-modal" onClick={(e) => e.stopPropagation()}>
                    <div id="song-list-controls">
                        <input
                            id="song-filter"
                            type="text"
                            placeholder="filter"
                            ref={this.inputRef}
                        />
                        <button
                            id="song-list-close"
                            type="button"
                            onClick={() => this.close()}
                        ></button>
                    </div>
                    <div id="song-list">
                        {this.state.clips.map((clip) => (
                            <Clip
                                key={clip}
                                radio={radio}
                                clip={clip}
                                expanded={this.state.expandedClip === clip}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    let selection = window.getSelection();
                                    if (
                                        selection === null ||
                                        selection.type !== "Range"
                                    ) {
                                        this.setState({ expandedClip: clip });
                                    }
                                }}
                            />
                        ))}
                    </div>
                </div>
            </div>
        );
    }

    public componentDidMount() {
        const input = this.inputRef.current!;
        const radio = this.props.radio;

        input.addEventListener("input", async () => {
            let filter = input.value.trim();
            if (filter === "") {
                this.clearResults();
            } else if (filter.length > 1) {
                this.setState({
                    clips: await radio.search(filter),
                });
            }
        });

        input.focus();

        document.addEventListener("keydown", this.escapeHandler);
    }

    public componentWillUnmount() {
        document.removeEventListener("keydown", this.escapeHandler);
    }

    private clearResults() {
        this.setState({ clips: [] });
    }

    private close() {
        unmountComponentAtNode(portal);
    }

    private escapeHandler(e: KeyboardEvent) {
        if (e.key === "Escape") {
            e.preventDefault();
            this.close();
        }
    }
}

type ClipProps = {
    clip: string;
    radio: PyRadio;
    expanded: boolean;
    onClick: (e: React.MouseEvent<HTMLDivElement>) => void;
};

type ClipState = {
    folder: string;
    filename: string;
};

class Clip extends Component<ClipProps, ClipState> {
    constructor(props: ClipProps) {
        super(props);
        let compontents = props.clip.split("/");
        let folder = compontents.slice(0, compontents.length - 1).join("/");
        let filename = compontents[compontents.length - 1];
        this.state = { folder, filename };
    }

    public render() {
        const props = this.props;
        const state = this.state;
        const radio = props.radio;

        return (
            <div
                className={props.expanded ? "song expanded" : "song"}
                onClick={props.onClick}
            >
                <div className="main">
                    {state.folder.length > 0 ? (
                        <span className="folder">{state.folder + "/"}</span>
                    ) : null}
                    <span className="file">{state.filename}</span>
                </div>
                <div className="buttons">
                    <a
                        className="add"
                        href="#"
                        onClick={async (e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            await radio.schedule(props.clip);
                            alert(state.filename + " added to queue.");
                        }}
                    >
                        add to queue
                    </a>
                    <a
                        className="download"
                        download={state.filename}
                        title={"download " + state.filename}
                        onClick={(e) => e.stopPropagation()}
                        href={radio.download_url(props.clip)}
                    >
                        download
                    </a>
                </div>
            </div>
        );
    }
}
