<template>
  <div>
    <NuxtRouteAnnouncer />
    <NuxtPage />
  </div>
</template>

<script setup lang="ts">
// Add dark class to html element
onMounted(() => {
  document.documentElement.classList.add('dark')
})
</script>
