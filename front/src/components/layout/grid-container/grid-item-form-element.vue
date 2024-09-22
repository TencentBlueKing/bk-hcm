<template>
  <grid-item v-bind="attrs">
    <div class="form-element">
      <slot />
    </div>
  </grid-item>
</template>
<script setup lang="ts">
import { useAttrs } from 'vue';
import GridItem from './grid-item.vue';

const attrs = useAttrs();
</script>
