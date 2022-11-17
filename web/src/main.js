import { createApp } from "vue";
import { createPinia } from "pinia";
import vuetify from "./vuetify";
import { i18n } from "./i18n";

import App from "./App.vue";
import router from "./router";

import "./assets/main.scss";

const app = createApp(App);

app.use(createPinia());
app.use(router);
app.use(vuetify);
app.use(i18n);

app.mount("#app");
