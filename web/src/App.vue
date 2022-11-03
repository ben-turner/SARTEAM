<template>
  <v-app>
    <v-main>
      <component :is="layout">
        <RouterView @confirm="handleConfirm" />
      </component>
    </v-main>
  </v-app>
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";

import * as layouts from "@/layouts";

const route = useRoute();

const layout = computed(() => {
  return layouts[route.meta.layout] || "div";
});

function handleConfirm(e) {
  const res = window.confirm(e.text);
  const handler = res ? e.onConfirm : e.onCancel;
  if (handler) {
    handler();
  }
}
</script>
