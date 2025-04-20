import { Component } from "preact";

export default class About extends Component {
    public render() {
        return (
            <div id="about">
                The code for this project is&nbsp;
                <a href="https://github.com/tim-we/wavestreamer">
                    available on GitHub
                </a>
                .
            </div>
        );
    }
}
