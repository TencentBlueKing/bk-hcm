<script setup lang="ts">
defineProps<{ value: string[] | number[] }>();
</script>

<template>
  {{ value?.join(', ') || '--' }}
</template>
