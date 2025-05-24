import { render, Component, createRef } from "preact";
import WavestreamerApi, { SearchResultEntry } from "../wavestreamer-api";
import { unmountComponentAtNode, type MouseEvent } from "preact/compat";

const portal = document.getElementById("modal-portal")!;

export function show(radio: WavestreamerApi) {
  render(<SongListModal radio={radio} />, portal);
}

type SLMProps = {
  radio: WavestreamerApi;
};

type SLMState = {
  clips: SearchResultEntry[];
  expandedClipId?: string;
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
                expanded={this.state.expandedClipId === clip.id}
                onClick={(e) => {
                  e.stopPropagation();
                  let selection = window.getSelection();
                  if (selection === null || selection.type !== "Range") {
                    this.setState({ expandedClipId: clip.id });
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
  clip: SearchResultEntry;
  radio: WavestreamerApi;
  expanded: boolean;
  onClick: (e: MouseEvent<HTMLDivElement>) => void;
};

class Clip extends Component<ClipProps> {
  public render() {
    const props = this.props;
    const radio = props.radio;

    const compontents = props.clip.name.split("/");
    const folder = compontents.slice(0, compontents.length - 1).join("/");
    const filename = compontents[compontents.length - 1];

    return (
      <div
        className={props.expanded ? "song expanded" : "song"}
        onClick={props.onClick}
      >
        <div className="main">
          {folder.length > 0 ? (
            <span className="folder">{folder + "/"}</span>
          ) : null}
          <span className="file">{filename}</span>
        </div>
        <div className="buttons">
          <a
            className="add"
            href="#"
            onClick={async (e) => {
              e.preventDefault();
              e.stopPropagation();
              await radio.schedule(props.clip.id);
              alert(filename + " added to queue.");
            }}
          >
            add to queue
          </a>
          <a
            className="download"
            download={filename}
            title={"download " + filename}
            onClick={(e) => e.stopPropagation()}
            href={radio.getDownloadUrl(props.clip.id)}
          >
            download
          </a>
        </div>
      </div>
    );
  }
}
