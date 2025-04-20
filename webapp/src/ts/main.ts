
import { createElement, render } from "preact";
import App from "./components/App";
import Wavestreamer from "./wavestreamer";

const root = document.getElementById("root")!;
const radio = new Wavestreamer();

render(createElement(App, { radio }), root);
