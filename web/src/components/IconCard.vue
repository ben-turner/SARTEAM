<template>
  <v-card to="/incidents/new">
    <v-card-text>
      <svg class="giant" viewBox="0 0 24 24">
        <path fill="currentColor" :d="props.iconSVG" />
      </svg>
    </v-card-text>
    <v-card-title>{{ props.title }}</v-card-title>
  </v-card>
</template>

<script setup>
const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  to: {
    type: String,
    required: true,
  },
  iconSVG: {
    type: String,
    required: true,
  },
});
</script>

<style scoped>
svg.giant {
  display: block;
  width: 50%;
  max-height: 30vh;
  margin: 0 auto;
}

.v-card {
  text-align: center;
}
</style>
