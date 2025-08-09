import { createContext, type FunctionComponent, render } from "preact";
import { useContext, useEffect, useRef, useState } from "preact/hooks";
import type WavestreamerApi from "../wavestreamer-api";
import type { SearchResultEntry } from "../wavestreamer-api";

const portal = document.getElementById("modal-portal")!;
const highlight = new Highlight();
const UserQueryContext = createContext("");
const segmenter = new Intl.Segmenter(undefined, { granularity: "word" });
CSS.highlights.set("search-results", highlight);

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
    const inputListener = async () => {
      highlight.clear();
      const filter = input.value.trim();
      if (filter === "") {
        setClips([]);
      } else if (filter.length > 1) {
        setClips(await radio.search(filter));
      }
    };

    input.addEventListener("input", inputListener);
    input.focus();

    document.addEventListener("keydown", songListKeydownHandler);

    return () => {
      document.removeEventListener("keydown", songListKeydownHandler);
      input.removeEventListener("input", inputListener);
    };
  }, []);

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
          <UserQueryContext.Provider value={inputRef.current?.value ?? ""}>
            {clips.map((clip) => (
              <Clip key={clip.id} radio={radio} clip={clip} />
            ))}
          </UserQueryContext.Provider>
        </div>
      </div>
    </div>
  );
};

function closeSongListModal() {
  // Unmount: https://github.com/preactjs/preact/issues/53
  render("", portal);
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
  const userQuery = useContext(UserQueryContext);
  const nameRef = useRef<HTMLElement>();

  useEffect(() => {
    const loweredClipName = clip.name.toLocaleLowerCase();
    for (const { segment, isWordLike } of segmenter.segment(userQuery)) {
      if (!isWordLike) {
        continue;
      }

      const index = loweredClipName.indexOf(segment.toLocaleLowerCase());

      if (index < 0) {
        continue;
      }

      const range = new Range();
      const textNode = nameRef.current.childNodes[0];
      range.setStart(textNode, index);
      range.setEnd(textNode, index + segment.length);

      highlight.add(range);
    }
  }, [clip.name, userQuery]);

  return (
    <details class="song" name="clip">
      <summary>
        <span class="file" ref={nameRef}>
          {clip.name}
        </span>
      </summary>
      <div class="buttons">
        <button
          class="add"
          type="button"
          onClick={async (e) => {
            e.preventDefault();
            e.stopPropagation();
            await radio.schedule(clip.id);
            alert(`${clip.name} added to queue.`);
          }}
        >
          add to queue
        </button>
        <button
          class="download"
          type="button"
          title={`download ${clip.name}`}
          onClick={() => downloadClip(radio, clip, clip.name)}
        >
          download
        </button>
      </div>
    </details>
  );
};

function downloadClip(
  radio: WavestreamerApi,
  clip: SearchResultEntry,
  filename: string,
): void {
  const a = document.createElement("a");
  a.href = radio.getDownloadUrl(clip.id);
  a.onclick = (e) => e.stopPropagation();
  a.download = filename;
  a.click();
}
