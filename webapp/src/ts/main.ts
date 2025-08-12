import { createElement, render } from "preact";
import App from "./components/App";
import * as WavestreamerApi from "./wavestreamer-api";

const root = document.getElementById("root")!;
WavestreamerApi.init();

render(createElement(App, {}), root);
