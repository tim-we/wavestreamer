import { Component } from "preact";
import WavestreamerApi from "../wavestreamer-api";
import * as SongListModal from "./SongListModal";

import pauseIcon from "../../img/pause.svg";
import repeatIcon from "../../img/repeat.svg";
import skipIcon from "../../img/skip.svg";
import listIcon from "../../img/list.svg";

type ControlsProps = {
    radio: WavestreamerApi;
};

const SVG_ICONS = {
    pause: pauseIcon,
    repeat: repeatIcon,
    skip: skipIcon,
    list: listIcon,
} as const;

export default class Controls extends Component<ControlsProps> {
    public render() {
        const radio = this.props.radio;

        return (
            <section id="controls">
                <Button
                    id="pause"
                    tooltip="toggle pause"
                    onClick={() => radio.pause()}
                />
                <Button
                    id="repeat"
                    tooltip="repeat current clip"
                    onClick={() => radio.repeat()}
                />
                <Button
                    id="skip"
                    tooltip="skip current clip"
                    onClick={() => radio.skip()}
                />
                <Button
                    id="song-list-button"
                    tooltip="song list"
                    icon="list"
                    onClick={() => {
                        SongListModal.show(radio);
                        return Promise.resolve();
                    }}
                />
                <Button
                    id="news"
                    tooltip="Tagesschau in 100s"
                    onClick={() => {
                        // TODO
                        return Promise.resolve();
                    }}
                >🗞️</Button>
            </section>
        );
    }
}

type ButtonProps = {
    id?: string;
    tooltip: string;
    icon?: "pause" | "repeat" | "skip" | "list";
    onClick: () => Promise<any>;
};

type ButtonState = {
    active: boolean;
};

class Button extends Component<ButtonProps, ButtonState> {
    public constructor(props: ButtonProps) {
        super(props);
        this.state = { active: false };
        this.clickHandler = this.clickHandler.bind(this);
    }

    private clickHandler() {
        this.setState({ active: true }, async () => {
            await this.props.onClick().catch((e) => {
                console.error(e);
                alert(e.message || "operation failed");
            });
            this.setState({ active: false });
        });
    }

    public render() {
        const props = this.props;
        const state = this.state;
        const classes = [];

        if (state.active) {
            classes.push("active");
        }

        const tooltip = state.active ? "" : props.tooltip;

        const icon = props.icon ?? props.id;

        return (
            <button
                id={props.id}
                title={tooltip}
                type="button"
                onClick={this.clickHandler}
                className={classes.join(" ")}
            >
                {props.children ? (
                    props.children
                ) : (
                    <img src={SVG_ICONS[icon]} />
                )}
            </button>
        );
    }
}
