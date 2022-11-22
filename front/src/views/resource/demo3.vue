<template>
  demo3-{{aaa}}
</template>

<script lang="ts">
import { reactive, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';

export default defineComponent({
  name: 'DemoThree',
  setup() {
    const state = reactive({
      aaa: 111,
    });
    onMounted(() => {
    });

    onUnmounted(() => {
    });

    return {
      ...toRefs(state),
    };
  },
});
</script>
