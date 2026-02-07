import type { FunctionComponent } from "preact";
import logo from "../../img/icon.svg";
import { connectedSignal } from "../wavestreamer-api";

const Header: FunctionComponent = () => (
  <header>
    <img alt="logo" src={logo} />
    <h1>wavestreamer</h1>
    {connectedSignal.value ? null : (
      <div class="error connection-lost">Connection lost.</div>
    )}
  </header>
);

export default Header;
