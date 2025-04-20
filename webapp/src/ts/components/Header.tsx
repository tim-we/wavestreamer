import React, { Component } from "preact";

export default class Header extends Component {
    public render() {
        return (<div id="header">
            <img alt="logo" src="/static/img/icon.svg"/>
            <span>wavestreamer</span>
        </div>);
    }
}
