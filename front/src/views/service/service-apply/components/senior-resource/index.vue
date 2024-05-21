<template>
  <div class="senior-resource">
    <ResourceCard :list="list" />
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import ResourceCard from '../resource-card/index.vue';

export default defineComponent({
  name: 'SeniorResource',
  components: {
    ResourceCard,
  },
  setup() {
    const list = ref(
      new Array(Math.ceil(Math.random() * 900) + 100).fill('').map((_, index) => ({
        name: '高级高级高级高级高级高级高级高级',
        id: index + 1,
        btnText: '立即录入',
      })),
    );

    const handleApply = () => {
      console.log('立即申请');
    };

    return {
      list,
      handleApply,
    };
  },
});
</script>
