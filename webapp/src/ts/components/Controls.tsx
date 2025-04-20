import { Component } from "preact";
import PyRadio from "../wavestreamer";
import * as SongListModal from "./SongListModal";

type ControlsProps = {
    radio: PyRadio;
    extensions: { name: string; command: string }[];
};

export default class Controls extends Component<ControlsProps> {
    public render() {
        const radio = this.props.radio;

        return (
            <div id="controls">
                {this.props.extensions.map((ext) => (
                    <Button
                        key={ext.command}
                        tooltip={ext.name}
                        onClick={() =>
                            radio.api_request(
                                `/extensions/${ext.command}/schedule`,
                                "PUT"
                            )
                        }
                    >
                        {ext.command}
                    </Button>
                ))}
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
            </div>
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
                    <img src={`/static/img/${props.icon || props.id}.svg`} />
                )}
            </button>
        );
    }
}
