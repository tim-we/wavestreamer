import type { FunctionComponent } from "preact";
import { useEffect, useState } from "preact/hooks";
import listIcon from "../../img/list.svg";
import pauseIcon from "../../img/pause.svg";
import repeatIcon from "../../img/repeat.svg";
import skipIcon from "../../img/skip.svg";
import type WavestreamerApi from "../wavestreamer-api";
import * as SongListModal from "./SongListModal";

type ControlsProps = {
  radio: WavestreamerApi;
};

const SVG_ICONS = {
  pause: pauseIcon,
  repeat: repeatIcon,
  skip: skipIcon,
  list: listIcon,
} as const;

const Controls: FunctionComponent<ControlsProps> = ({ radio }) => {
  const [newsEnabled, setNewsEnabled] = useState<boolean>(false);

  useEffect(() => {
    radio.getConfig().then((config) => setNewsEnabled(config.news));
  }, [radio]);

  return (
    <section id="controls">
      <Button id="pause" tooltip="toggle pause" onClick={() => radio.pause()} />
      <Button
        id="repeat"
        label="repeat"
        tooltip="repeat current clip"
        onClick={() => radio.repeat()}
      />
      <Button
        id="skip"
        label="skip"
        tooltip="skip current clip"
        onClick={() => radio.skip()}
      />
      <Button
        id="song-list-button"
        label="song list"
        tooltip="song list"
        icon="list"
        onClick={() => {
          SongListModal.show(radio);
          return Promise.resolve();
        }}
      />
      {newsEnabled ? (
        <Button
          id="news"
          tooltip="Tagesschau in 100s"
          onClick={() => radio.news()}
        >
          🗞️
        </Button>
      ) : null}
    </section>
  );
};

export default Controls;

type ButtonProps = {
  id?: string;
  label?: string;
  tooltip: string;
  icon?: "pause" | "repeat" | "skip" | "list";
  onClick: () => Promise<unknown>;
};

const Button: FunctionComponent<ButtonProps> = ({
  id,
  label,
  tooltip,
  icon,
  onClick,
  children,
}) => {
  const [active, setActive] = useState<boolean>(false);

  const clickHandler = async () => {
    setActive(true);
    await onClick().catch((e) => {
      console.error(e);
      alert(e.message ?? "operation failed");
    });
    setActive(false);
  };

  const classes = [];

  if (active) {
    classes.push("active");
  }

  const iconSrc = SVG_ICONS[icon ?? id];

  return (
    <button
      id={id}
      title={tooltip}
      type="button"
      onClick={clickHandler}
      className={classes.join(" ")}
    >
      {children ? children : <img alt={label} src={iconSrc} />}
    </button>
  );
};
