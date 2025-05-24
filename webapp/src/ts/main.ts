import { createElement, render } from "preact";
import App from "./components/App";
import WavestreamerApi from "./wavestreamer-api";

const root = document.getElementById("root")!;
const radio = new WavestreamerApi();

render(createElement(App, { radio }), root);
