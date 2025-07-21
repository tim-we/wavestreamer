import type { FunctionComponent } from "preact";

import logo from "../../img/icon.svg";

const Header: FunctionComponent = () => (
  <header>
    <img alt="logo" src={logo} />
    <h1>wavestreamer</h1>
  </header>
);

export default Header;
