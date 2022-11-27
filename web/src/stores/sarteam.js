import { defineStore } from "pinia";
import http from "@/http";

export const useSarteamStore = defineStore("sarteam", {
  state: () => ({
    incidents: [],
    activeIncident: null,
    config: {},
  }),
  actions: {
    async getIncidents() {
      const res = await http({
        method: "GET",
        url: "/incident",
      });
      this.incidents = res.data;
    },
    async getActiveIncident() {
      try {
        const res = await http({
          method: "GET",
          url: "/incident/active",
        });

        this.activeIncident = res.data;
      } catch (e) {
        if (e.response.status === 404) {
          this.activeIncident = null;
        } else {
          throw e;
        }
      }
    },
    async getConfig() {
      const res = await http({
        method: "GET",
        url: "/config",
      });

      this.config = res.data;
    },
    async createIncident(incident) {
      const res = await http({
        method: "POST",
        url: "/incident",
        data: incident,
      });

      this.activeIncident = res.data;
    },
  },
});
