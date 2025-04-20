import { Component } from "preact";

import logo from "../../img/icon.svg";

export default class Header extends Component {
    public render() {
        return (<div id="header">
            <img alt="logo" src={logo}/>
            <span>wavestreamer</span>
        </div>);
    }
}
