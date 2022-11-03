import { createRouter, createWebHashHistory } from "vue-router";
import HomeView from "../views/HomeView.vue";
import * as IncidentWizard from "../views/IncidentWizard";
import { useSarteamStore } from "../stores/sarteam";

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: HomeView,
    },
    {
      path: "/incidents/new",
      name: "wizard",
      async beforeEnter() {
        const sarteamStore = useSarteamStore();
        await sarteamStore.getActiveIncident();

        return sarteamStore.activeIncident
          ? { name: "wizard-join-existing" }
          : { name: "wizard-details" };
      },
    },
    {
      path: "/incidents/new/details",
      name: "wizard-details",
      component: IncidentWizard.IncidentDetails,
      meta: {
        title: "wizard.details.title",
        layout: "CardLayout",
      },
    },
    {
      path: "/incidents/new/join-existing",
      name: "wizard-join-existing",
      component: IncidentWizard.JoinExisting,
      meta: {
        title: "Checking for Active Incidents",
        layout: "CardLayout",
      },
    },
    {
      path: "/incidents/new/speedtest",
      name: "wizard-speedtest",
      component: IncidentWizard.SpeedTest,
      meta: {
        title: "Checking Online Connectivity",
        layout: "CardLayout",
      },
    },
    {
      path: "/incidents/new/create-map",
      name: "wizard-create-map",
      component: IncidentWizard.CreateMap,
      props: (route) => ({ online: route.query.online === "true" }),
      meta: {
        title: "Create Incident SARTopo Map",
        layout: "CardLayout",
      },
    },
  ],
});

export default router;
