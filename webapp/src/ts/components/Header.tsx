import { Component } from "preact";

import logo from "../../img/icon.svg";

export default class Header extends Component {
    public render() {
        return (<header>
            <img alt="logo" src={logo}/>
            <h1>wavestreamer</h1>
        </header>);
    }
}
