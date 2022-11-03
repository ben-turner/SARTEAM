import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import { mdi, aliases } from "vuetify/iconsets/mdi";

import "@mdi/font/css/materialdesignicons.css";
import "vuetify/styles";

const light = {
  dark: false,
  colors: {
    primary: "#2c5da7",
    "on-primary": "#ffffff",
    secondary: "#616200",
    "on-secondary": "#ffffff",
    surface: "#fdfbff",
    "on-surface": "#1a1b1f",
    background: "#fdfbff",
  },
};
const dark = {
  dark: true,
  colors: {
    primary: "#abc7ff",
    "on-primary": "#002f65",
    "primary-container": "#03458e",
    "on-primary-container": "#d7e3ff",
    secondary: "#cbcc58",
    "on-secondary": "#323200",
    "secondary-container": "#494a00",
    "on-secondary-container": "#e7e970",
    error: "#ffb4ab",
    "on-error": "#690005",
    "error-container": "#93000a",
    "on-error-container": "#ffdad6",
    background: "#1a1b1f",
    "on-background": "#e3e2e6",
    surface: "#1a1b1f",
    "on-surface": "#e3e2e6",
    outline: "#8e9099",
    "surface-variant": "#44474e",
    "on-surface-variant": "#c4c6d0",
  },
};

export default createVuetify({
  components,
  directives,
  icons: {
    defaultSet: "mdi",
    aliases,
    sets: {
      mdi,
    },
  },
  theme: {
    defaultTheme: "dark",
    themes: {
      light,
      dark,
    },
  },
});
