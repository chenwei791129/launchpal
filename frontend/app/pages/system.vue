<template>
  <ReadOnlyServiceList
    title="System Services"
    permission-key="system"
    warning-text="Some system services may not be visible. LaunchPal needs disk access permission to read system LaunchDaemons."
    :list-fn="listFn"
  />
</template>

<script setup lang="ts">
const listFn = () => window.go.main.App.ListSystemServices()
</script>
