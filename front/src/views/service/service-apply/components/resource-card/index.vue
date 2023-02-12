<template>
  <div class="resource-list">
    <div class="resource-item" v-for="item in list" :key="item?.id">
      <div class="resource-name">{{ item.name }}</div>
      <div class="bottom-btn">
        <bk-button theme="primary" size="small" @click="handleApply('applyAccount', item.id)">
          {{ item.btnText }}
        </bk-button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { useRouter } from 'vue-router';

export default defineComponent({
  name: 'ResourceCard',
  props: {
    list: {
      type: Array as PropType<any[]>,
    },
  },
  setup() {
    const router = useRouter();

    // 跳转页面
    const handleApply = (routerName: string, id?: string | number) => {
      const routerConfig = {
        query: {},
        name: routerName,
      };
      if (id) {
        routerConfig.query = {
          id,
        };
      }
      console.log(routerConfig, '配置');
      router.push(routerConfig);
    };

    return {
      handleApply,
    };
  },
});
</script>

<style lang="scss" scoped>
.resource-list {
  display: flex;
  flex-wrap: wrap;
  .resource-item {
    min-width: 240px;
    margin-top: 20px;
    margin-left: 20px;
    padding: 10px;
    border-radius: 4px;
    box-shadow: 2px 2px 4px 1px rgb(0 0 0 / 15%);
    .resource-name {
      max-width: 200px;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
  }
  .bottom-btn {
    margin-top: 20px;
    text-align: right;
  }
}
</style>
