<template>
  <v-container>
    <v-card color="primary-container">
      <v-card-title>{{ $t(title) }}</v-card-title>
      <slot></slot>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const title = computed(() => {
  return route.meta.title;
});
</script>

<style scoped>
.v-container {
  max-width: 800px;
}
</style>
