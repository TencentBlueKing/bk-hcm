<script lang="ts" setup>
import { type PropType } from 'vue';
import { useRouter } from 'vue-router';
import { AUTH_IMPORT_ACCOUNT, AUTH_BIZ_CREATE_IAAS_RESOURCE } from '@/constants/auth-symbols';

defineProps({
  list: {
    type: Array as PropType<any[]>,
  },
});

const router = useRouter();

const getAuthSign = (routeName: string) => {
  if (routeName === 'applyAccount') return { type: AUTH_IMPORT_ACCOUNT };
  return { type: AUTH_BIZ_CREATE_IAAS_RESOURCE };
};

const handleApply = (routerName: string, id?: string | number) => {
  const routerConfig: { query: Record<string, any>; name: string } = {
    query: {},
    name: routerName,
  };
  if (id) {
    routerConfig.query = { id };
  }
  router.push(routerConfig);
};
</script>

<template>
  <div class="resource-list">
    <div class="resource-item" v-for="item in list" :key="item?.id">
      <div class="resource-title">
        <img src="@/assets/image/serviceCard.png" alt="" />
        <span class="resource-name pl20">{{ item.name }}</span>
      </div>
      <div class="sub-resource-title">申请 {{ item.name }}</div>
      <div class="bottom-btn">
        <hcm-auth :sign="getAuthSign(item.routeName)" v-slot="{ noPerm }">
          <bk-button theme="primary" :disabled="noPerm" outline size="small" @click="handleApply(item.routeName)">
            {{ item.btnText }}
          </bk-button>
        </hcm-auth>
      </div>
    </div>
  </div>
</template>

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
