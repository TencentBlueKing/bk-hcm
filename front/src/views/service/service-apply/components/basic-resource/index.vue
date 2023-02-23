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
    const list = ref([{
      name:
          '主机',
      id: 1,
      btnText: '立即申请',
    }]);

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
