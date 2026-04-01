<template>
  <ReadOnlyServiceList
    title="Apple System Services"
    permission-key="apple-system"
    warning-text="Some Apple system services may not be visible. LaunchPal needs disk access permission to read system LaunchDaemons."
    :list-fn="listFn"
  />
</template>

<script setup lang="ts">
const listFn = () => window.go.main.App.ListAppleSystemServices()
</script>
