import { defineComponent, onMounted, onUnmounted } from 'vue';

export default defineComponent({
  setup() {
    onMounted(() => {});

    onUnmounted(() => {});

    return () => <span class='test'>do it</span>;
  },
});
