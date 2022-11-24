<template>
  demo3-{{aaa}}
</template>

<script lang="ts">
import { reactive, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';
import { useUser } from '@/store/index';

export default defineComponent({
  name: 'DemoThree',
  setup() {
    const state = reactive({
      aaa: 111,
    });
    onMounted(async () => {
      const user = useUser();
      const res = await user.test();
      console.log(12213, res);
    });

    onUnmounted(() => {
    });

    return {
      ...toRefs(state),
    };
  },
});
</script>
