import { Component } from "preact";

type NPProps = {
    clip?: string;
};

export default class NowPlaying extends Component<NPProps> {
    public render() {
        const clip = this.props.clip;

        return (
            <div id="now">
                <div className="title">Now playing:</div>
                <div id="current-clip">{clip ? clip : "-"}</div>
            </div>
        );
    }
}
