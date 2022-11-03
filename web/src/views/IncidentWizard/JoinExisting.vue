<template>
  <v-card-text align="center">
    {{ props.existingIncident }}
  </v-card-text>
</template>

<script setup>
const props = defineProps({
  existingIncident: {
    type: Object,
    required: true,
  },
});
</script>

<style scoped>
p {
  margin: 1em 0;
}
</style>
