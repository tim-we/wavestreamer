import type { FunctionComponent } from "preact";

declare const __BUILD_DATE__: string;

const buildDate = new Date(__BUILD_DATE__);

const About: FunctionComponent = () => (
  <footer id="about">
    The code for this project is&nbsp;
    <a href="https://github.com/tim-we/wavestreamer">available on GitHub</a>.
    Build date {buildDate.toLocaleDateString()}.
  </footer>
);
export default About;
