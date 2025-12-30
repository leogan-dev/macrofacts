import { createRoot } from "react-dom/client";
import App from "./App";

import "./styles/theme.css";
import "./styles/ui.css";

// Importing theme module applies the initial theme immediately.
import "./theme";

const rootEl = document.getElementById("root");
if (!rootEl) {
    throw new Error("Root element #root not found");
}

createRoot(rootEl).render(<App />);
