<template>
  <div class="resource-list">
    <div class="resource-item" v-for="item in list" :key="item?.id">
      <div class="resource-title">
        <img src="@/assets/image/serviceCard.png" alt="" />
        <span class="resource-name pl20">{{ item.name }}</span>
      </div>
      <div class="sub-resource-title">申请 {{ item.name }}</div>
      <div
        class="bottom-btn"
        @click="handleAuthClick(item.routeName === 'applyAccount' ? 'account_import' : 'biz_iaas_resource_create')"
      >
        <bk-button
          theme="primary"
          :disabled="
            !authVerifyData?.permissionAction[
              item.routeName === 'applyAccount' ? 'account_import' : 'biz_iaas_resource_create'
            ]
          "
          outline
          size="small"
          @click="handleApply(item.routeName)"
        >
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
    authVerifyData: {
      type: Object,
    },
  },
  emits: ['handleAuthClick'],
  setup(props, { emit }) {
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

    const handleAuthClick = (action: string) => {
      emit('handleAuthClick', action);
    };

    return {
      handleApply,
      handleAuthClick,
    };
  },
});
</script>

<style lang="scss" scoped>
.resource-list {
  display: flex;
  flex-wrap: wrap;
  .resource-item {
    cursor: pointer;
    width: 25%;
    height: 160px;
    margin-top: 20px;
    margin-left: 20px;
    border-radius: 10px;
    box-shadow: 2px 2px 4px 1px rgb(0 0 0 / 15%);
    .resource-title {
      display: flex;
      padding: 15px 30px 0 30px;
      align-items: center;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
      .resource-name {
        font-size: 16px;
        font-weight: bold;
      }
    }
    .sub-resource-title {
      color: #63656e;
      font-size: 12px;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
      padding: 28px 30px;
    }
  }
  .resource-item:hover {
    box-shadow: 4px 4px 8px 2px rgb(0 0 0 / 15%);
  }
  .bottom-btn {
    text-align: right;
    height: 40px;
    line-height: 40px;
    background-color: #fafbfd;
    border-radius: 0 0 10px 10px;
    padding: 0 15px;
  }
}
</style>
