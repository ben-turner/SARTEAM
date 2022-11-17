<template>
  <v-container>
    <h1>{{ $t("home.welcome") }}</h1>
    <h2>{{ $t("home.cta") }}</h2>
    <v-row>
      <v-col md-6>
        <IconCard
          :title="$t('home.new-incident')"
          to="{ name: 'incident-wizard' }"
          color="primary"
          iconSVG="M12,1L3,5V11C3,16.55 6.84,21.74 12,23C17.16,21.74 21,16.55 21,11V5M11,7H13V13H11M11,15H13V17H11"
        />
      </v-col>
      <v-col md-6>
        <IconCard
          :title="$t('home.open-incident')"
          to="/incidents"
          color="primary"
          iconSVG="M19,20H4C2.89,20 2,19.1 2,18V6C2,4.89 2.89,4 4,4H10L12,6H19A2,2 0 0,1 21,8H21L4,8V18L6.14,10H23.21L20.93,18.5C20.7,19.37 19.92,20 19,20Z"
      /></v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import IconCard from "../components/IconCard.vue";
</script>

<style scoped>
.v-container {
  max-width: 800px;
  padding-top: 2em;
}

h2 {
  margin-bottom: 1em;
}
</style>
