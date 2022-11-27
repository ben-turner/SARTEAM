<template>
  <v-card-text>
    <v-form ref="form">
      <v-row>
        <v-col cols="12">
          <v-text-field
            v-model="incident.date"
            :label="$t('wizard.details.start-date')"
            :hint="$t('wizard.details.start-date-hint')"
            prepend-inner-icon="mdi-calendar"
            :rules="[rules.required, rules.date]"
            @change="fixDate"
          />
        </v-col>
        <v-col cols="12">
          <v-text-field
            v-model="incident.location"
            :label="$t('wizard.details.incident-location')"
            :hint="$t('wizard.details.incident-location-hint')"
            prepend-inner-icon="mdi-map-marker"
            :rules="[rules.required]"
          />
        </v-col>
        <v-col cols="4">
          <v-switch
            :label="$t('wizard.details.training')"
            v-model="incident.training"
            color="secondary"
          />
        </v-col>
        <v-col cols="8">
          <v-text-field
            v-if="!incident.training"
            v-model="incident.caseNumber"
            :label="$t('wizard.details.incident-case-number')"
            :hint="$t('wizard.details.incident-case-number-hint')"
            prepend-inner-icon="mdi-file-document-outline"
          />
          <p v-else>{{ $t("wizard.details.no-case-number") }}</p>
        </v-col>
      </v-row>
    </v-form>
    <v-alert type="error" :value="error">
      {{ $t("wizard.details.failed") }}
    </v-alert>
  </v-card-text>
  <v-card-actions>
    <v-spacer />
    <v-btn color="primary" @click="submit" :loading="loading">
      {{ $t("generic.next") }}
    </v-btn>
  </v-card-actions>
</template>

<script setup>
import { reactive, ref } from "vue";
import { useSarteamStore } from "@/stores/sarteam";
import { useI18n } from "vue-i18n";
const { t } = useI18n();
const sarteamStore = useSarteamStore();

const incident = reactive({
  date: new Date().toISOString().substring(0, 10),
});

const loading = ref(false);
const failed = ref(false);
const form = ref(null);

const rules = reactive({
  required: (v) => !!v || t("generic.required"),
  date: (v) => {
    const date = new Date(v);
    return !isNaN(date) || t("generic.invalid-date");
  },
});

function fixDate() {
  const date = new Date(incident.date);
  if (date.getFullYear() < 2010) {
    date.setFullYear(new Date().getFullYear());
  }
  if (!isNaN(date)) {
    incident.date = date.toISOString().substring(0, 10);
  }
}

async function submit() {
  loading.value = true;
  failed.value = false;
  const { valid } = await form.value.validate();
  console.log(valid);
  if (!valid) {
    loading.value = false;
    return;
  }

  try {
    await sarteamStore.createIncident(incident);
  } catch (e) {
    console.log(e);
    failed.value = true;
  } finally {
    loading.value = false;
  }
}
</script>
