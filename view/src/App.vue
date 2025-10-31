<template>
  <v-app>
    <v-app-bar color="primary" dark elevation="4">
      <v-app-bar-title class="font-weight-bold">
        <v-icon class="mr-2">mdi-database-search</v-icon>
        SQLfuse
      </v-app-bar-title>
    </v-app-bar>

    <v-main class="main-gradient">
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup>
  //
</script>

<style>
.main-gradient {
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}
</style>
