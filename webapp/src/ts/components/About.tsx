import { Component } from "preact";

export default class About extends Component {
    public render() {
        return (
            <footer id="about">
                The code for this project is&nbsp;
                <a href="https://github.com/tim-we/wavestreamer">
                    available on GitHub
                </a>
                .
            </footer>
        );
    }
}
