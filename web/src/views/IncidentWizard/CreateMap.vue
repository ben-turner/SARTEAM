<template>
  <v-card-text>
    <p>
      A new SARTopo map needs to be created and saved with some data. To do
      this, follow the steps below:
    </p>
    <ol>
      <li>
        <a
          :href="`${baseMapAddress.protocol}${baseMapAddress.host}/map.html`"
          target="_blank"
          >Open SARTopo</a
        >.
      </li>
      <li v-if="props.online">
        If there is a "Log In" button visible in the top-right corner, click it
        and sign in using the jdfsar google account.
      </li>
      <li>Click the "Add" button in the top left.</li>
      <li>Select "Marker" to create a new marker.</li>
      <li>
        Drag the red dot that appears to the approximate location of the ICP.
      </li>
      <li>
        Use the window in the bottom right of the screen to label the point as
        "ICP".
      </li>
      <li>
        Click the red dot beside "Style" and select the ICP icon under
        "ICS/Fire".
      </li>
      <li>Click OK in the bottom right.</li>
      <li>A new window appears, prompting you to save the map.</li>
      <li>
        Name the map with the date, type of incident, and area. Eg. `2022-10-18
        Rescue East Sooke Park`
      </li>
      <li>Click "Save" in the bottom right.</li>
    </ol>
    <p>
      Once you have saved the map, your address bar should change to include a
      map ID. Eg. `{{ baseMapAddress.protocol
      }}{{ baseMapAddress.host }}/m/ABCDEF`. Copy this address and paste it into
      the field below, then click "Next".
    </p>
    <v-text-field
      label="Map Address"
      v-model="mapAddress"
      :error-messages="addressErrors"
      @input="(v) => handleMapAddress(false)"
      @keyup.enter="(v) => handleMapAddress(true)"
      :loading="loading"
    />
  </v-card-text>
  <v-card-actions>
    <v-spacer />
    <v-btn color="primary" @click="() => handleMapAddress(true)"> Next </v-btn>
  </v-card-actions>
</template>

<script setup>
import { ref, computed } from "vue";
import http from "@/http";

const emit = defineEmits(["confirm"]);
const props = defineProps({
  online: Boolean,
});
const addressErrors = ref([]);
const mapAddress = ref("");
const loading = ref(false);

const addressRegex = /^(https?:\/\/)?(?:(.*)\/m\/)?([0-9A-Z]+)$/;

const onlineMapBaseAddress = {
  protocol: "https://",
  host: "sartopo.com",
};
const offlineMapBaseAddress = {
  protocol: "http://",
  host: "sartopo.cmd.jdfsar.net",
};

const baseMapAddress = computed(() => {
  if (props.online) {
    return onlineMapBaseAddress;
  } else {
    return offlineMapBaseAddress;
  }
});

async function handleMapAddress(navigateOnSuccess) {
  if (mapAddress.value === "") {
    addressErrors.value = ["You must enter a map address"];
    return;
  }

  const match = mapAddress.value.match(addressRegex);
  if (!match) {
    addressErrors.value = ["Invalid map address"];
    return;
  }

  const proto = match[1];
  const host = match[2];

  if (
    (host && host !== baseMapAddress.value.host) ||
    (proto && proto !== baseMapAddress.value.protocol)
  ) {
    if (navigateOnSuccess) {
      const continueAnyway = await new Promise((resolve) => {
        emit("confirm", {
          title: "Map Address Mismatch",
          text: `The map address you entered does not look like an ${
            props.online ? "online" : "offline"
          } map address. This is probably a mistake unless you know what you're doing. Do you want to continue anyway?`,
          confirm: "Continue with unrecognized address",
          cancel: "Cancel",
          onConfirm: () => resolve(true),
          onCancel: () => resolve(false),
        });
      });

      if (!continueAnyway) {
        return;
      }
    } else {
      addressErrors.value = [
        "Unexpected map host/protocol. This is probably a mistake unless you know what you're doing.",
      ];

      return;
    }
  }

  addressErrors.value = [];

  if (!navigateOnSuccess) {
    return;
  }

  loading.value = true;
  const res = await http({
    method: "post",
    url: "/newIncident",
    body: {
      mapURL: mapAddress.value,
    },
  });

  console.log(res);
}
</script>

<style scoped lang="scss">
ol {
  margin: 1em;

  li {
    margin: 0 1em;
  }
}

p {
  margin: 1em 0;
}
</style>
