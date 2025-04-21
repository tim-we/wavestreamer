import { Component } from "preact";

type NPProps = {
    clip?: string;
};

export default class NowPlaying extends Component<NPProps> {
    public render() {
        const clip = this.props.clip;

        return (
            <section id="now">
                <div class="title">Now playing:</div>
                <div id="current-clip">{clip ? clip : "-"}</div>
            </section>
        );
    }
}
