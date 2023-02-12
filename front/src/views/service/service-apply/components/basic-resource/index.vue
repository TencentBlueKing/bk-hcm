<template>
  <div class="basic-resource">
    <ResourceCard
      :list="list"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import ResourceCard from '../resource-card/index.vue';

export default defineComponent({
  name: 'BasicResource',
  components: {
    ResourceCard,
  },
  setup() {
    const list = ref(new Array(Math.ceil(Math.random() * 900) + 100).fill('')
      .map((_, index) => ({
        name:
          '主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机主机',
        id: index + 1,
        btnText: '立即申请',
      })));

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
