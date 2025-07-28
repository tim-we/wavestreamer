import { type FunctionComponent, render } from "preact";
import {
  unmountComponentAtNode,
  useEffect,
  useRef,
  useState,
} from "preact/compat";
import type WavestreamerApi from "../wavestreamer-api";
import type { SearchResultEntry } from "../wavestreamer-api";

const portal = document.getElementById("modal-portal")!;

export function show(radio: WavestreamerApi) {
  render(<SongListModal radio={radio} />, portal);
}

type SLMProps = {
  radio: WavestreamerApi;
};

const SongListModal: FunctionComponent<SLMProps> = ({ radio }) => {
  const inputRef = useRef<HTMLInputElement>();
  const [clips, setClips] = useState<SearchResultEntry[]>([]);

  useEffect(() => {
    const input = inputRef.current;

    input.addEventListener("input", async () => {
      const filter = input.value.trim();
      if (filter === "") {
        setClips([]);
      } else if (filter.length > 1) {
        setClips(await radio.search(filter));
      }
    });

    input.focus();

    document.addEventListener("keydown", songListKeydownHandler);
    // TODO remove event listener
  }, [inputRef.current?.value]);

  return (
    <div
      id="song-list-container"
      className="show"
      onClick={(e) => {
        e.stopPropagation();
        closeSongListModal();
      }}
    >
      {/** biome-ignore lint/a11y/noStaticElementInteractions: This is just a event boundary */}
      <div id="song-list-modal" onClick={(e) => e.stopPropagation()}>
        <div id="song-list-controls">
          <input
            id="song-filter"
            type="text"
            placeholder="filter"
            ref={inputRef}
          />
          <button
            id="song-list-close"
            type="button"
            onClick={closeSongListModal}
          />
        </div>
        <div id="song-list">
          {clips.map((clip) => (
            <Clip key={clip} radio={radio} clip={clip} />
          ))}
        </div>
      </div>
    </div>
  );
};

function closeSongListModal() {
  unmountComponentAtNode(portal);
}

function songListKeydownHandler(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.preventDefault();
    closeSongListModal();
  }
}

type ClipProps = {
  clip: SearchResultEntry;
  radio: WavestreamerApi;
};

const Clip: FunctionComponent<ClipProps> = ({ clip, radio }) => {
  const compontents = clip.name.split("/");
  const folder = compontents.slice(0, compontents.length - 1).join("/");
  const filename = compontents[compontents.length - 1];

  return (
    <details class="song" name="clip">
      <summary>
        {folder.length > 0 ? <span class="folder">{`${folder}/`}</span> : null}
        <span class="file">{filename}</span>
      </summary>
      <div class="buttons">
        <button
          class="add"
          type="button"
          onClick={async (e) => {
            e.preventDefault();
            e.stopPropagation();
            await radio.schedule(clip.id);
            alert(`${filename} added to queue.`);
          }}
        >
          add to queue
        </button>
        <button
          class="download"
          type="button"
          title={`download ${filename}`}
          onClick={() => downloadClip(radio, clip, filename)}
        >
          download
        </button>
      </div>
    </details>
  );
};

function downloadClip(radio: WavestreamerApi, clip: SearchResultEntry, filename: string): void {
  const a = document.createElement("a");
  a.href = radio.getDownloadUrl(clip.id);
  a.onclick = (e) => e.stopPropagation();
  a.download = filename;
  //a.click(); // TODO: re-enable once server side is implemented
}