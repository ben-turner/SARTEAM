<template>
  <v-card-text align="center">
    <v-progress-circular
      v-if="recommended === ''"
      indeterminate
      :size="200"
      :width="24"
    ></v-progress-circular>
    <svg v-else class="giant" viewBox="0 0 24 24">
      <path fill="currentColor" :d="icons[recommended]" />
    </svg>
    <p>{{ messages[recommended] }}</p>
    <p v-if="recommended !== ''">
      <v-btn variant="flat" @click="checkConnection">
        Click here to rerun test
      </v-btn>
    </p>
  </v-card-text>

  <v-card-actions v-if="recommended === 'online'">
    <v-spacer />
    <v-btn variant="text" color="error" :to="linkOffline">
      Use Offline Mode Instead
    </v-btn>
    <v-btn variant="flat" color="primary" :to="linkOnline">
      Continue Online
    </v-btn>
  </v-card-actions>

  <v-card-actions v-if="recommended === 'offline'">
    <v-spacer />
    <v-btn variant="text" color="error" :to="linkOnline">
      Use Online Mode Instead
    </v-btn>
    <v-btn variant="flat" color="primary" :to="linkOffline">
      Continue Offline
    </v-btn>
  </v-card-actions>

  <v-card-actions v-if="recommended === 'failed'">
    <v-spacer />
    <v-btn variant="text" color="primary" :to="linkOnline">
      Continue Online
    </v-btn>
    <v-btn variant="text" color="primary" :to="linkOffline">
      Continue Offline
    </v-btn>
  </v-card-actions>
</template>

<script setup>
import { ref, onMounted } from "vue";
import http from "@/http";

const recommended = ref("");

async function checkConnection() {
  recommended.value = "";

  try {
    const res = await http({
      url: "/networkStatus",
    });

    if (res.data.online) {
      recommended.value = "online";
    } else {
      recommended.value = "offline";
    }
  } catch (err) {
    recommended.value = "failed";
  }
}

onMounted(() => {
  checkConnection();
});

const linkOnline = {
  name: "wizard-create-map",
  query: {
    online: true,
  },
};

const linkOffline = {
  name: "wizard-create-map",
  query: {
    online: false,
  },
};

const icons = {
  online:
    "M5,14H19L17.5,9.5H6.5L5,14M17.5,19A1.5,1.5 0 0,0 19,17.5A1.5,1.5 0 0,0 17.5,16A1.5,1.5 0 0,0 16,17.5A1.5,1.5 0 0,0 17.5,19M6.5,19A1.5,1.5 0 0,0 8,17.5A1.5,1.5 0 0,0 6.5,16A1.5,1.5 0 0,0 5,17.5A1.5,1.5 0 0,0 6.5,19M18.92,9L21,15V23A1,1 0 0,1 20,24H19A1,1 0 0,1 18,23V22H6V23A1,1 0 0,1 5,24H4A1,1 0 0,1 3,23V15L5.08,9C5.28,8.42 5.85,8 6.5,8H17.5C18.15,8 18.72,8.42 18.92,9M12,0C14.12,0 16.15,0.86 17.65,2.35L16.23,3.77C15.11,2.65 13.58,2 12,2C10.42,2 8.89,2.65 7.77,3.77L6.36,2.35C7.85,0.86 9.88,0 12,0M12,4C13.06,4 14.07,4.44 14.82,5.18L13.4,6.6C13.03,6.23 12.53,6 12,6C11.5,6 10.97,6.23 10.6,6.6L9.18,5.18C9.93,4.44 10.94,4 12,4Z",
  offline:
    "M18,3V16.18L21,19.18V3H18M4.28,5L3,6.27L10.73,14H8V21H11V14.27L13,16.27V21H16V19.27L19.73,23L21,21.72L4.28,5M13,9V11.18L16,14.18V9H13M3,18V21H6V18H3Z",
  failed:
    "M13 11H11V5H13M13 15H11V13H13M20 2H4C2.9 2 2 2.9 2 4V22L6 18H20C21.1 18 22 17.1 22 16V4C22 2.9 21.1 2 20 2Z",
};

const messages = {
  "": "Checking connection...",
  online:
    "Internet connection detected. Continue in online mode unless you are expecting to move outside of internet coverage or expect to loose internet connectivity prior to completing the incident.",
  offline:
    "No internet connection or poor quality connection detected. Continuing in offline mode is recommended.",
  failed:
    "Failed to check internet connection. Please try again, or you may be able to continue in offline mode.",
};
</script>

<style scoped>
p {
  margin: 1em 0;
}

svg {
  width: 200px;
  height: 200px;
}
</style>
