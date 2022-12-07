<script setup lang="ts">
import {
  ref,
  watch,
  computed,
} from 'vue';

import HostManage from './children/manage/host-manage.vue';
import VpcManage from './children/manage/vpc-manage.vue';
import SubnetManage from './children/manage/subnet-manage.vue';
import SecurityManage from './children/manage/security-manage.vue';
import DriveManage from './children/manage/drive-manage.vue';
import IpManage from './children/manage/ip-manage.vue';
import RoutingManage from './children/manage/routing-manage.vue';
import ImageManage from './children/manage/image-manage.vue';

import {
  RESOURCE_TYPES,
  VENDORS,
} from '@/common/constant';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRouter,
  useRoute,
} from 'vue-router';
import useSteps from './hooks/use-steps';

// use hooks
const {
  t,
} = useI18n();
const router = useRouter();
const route = useRoute();
const {
  isShowDistribution,
  handleDistribution,
  ResourceDistribution,
} = useSteps();

// 状态
const currentAccount = '';
const accounts: any[] = [];
const isAccurate = ref(false);

// 组件map
const componentMap = {
  host: HostManage,
  vpc: VpcManage,
  subnet: SubnetManage,
  security: SecurityManage,
  drive: DriveManage,
  ip: IpManage,
  routing: RoutingManage,
  image: ImageManage,
};

// 标签相关数据
const tabs = RESOURCE_TYPES.map((type) => {
  return {
    name: type.type,
    type: t(type.name),
    component: componentMap[type.type],
  };
});
const activeTab = ref(route.query.type || tabs[0].type);

// 搜索过滤相关数据
const filter = ref([]);
const computedSearchList = computed(() => {
  const datas = [
    {
      name: '名称',
      id: 'name',
      multiable: false,
    },
    {
      name: '云厂商',
      id: 'vendor',
      children: VENDORS,
    },
    {
      name: '云区域',
      id: 'k_cloud_id',
      multiable: false,
    },
    {
      name: '地域',
      id: 'region',
      multiable: false,
    },
  ];
  return datas.filter((data) => {
    return !filter.value.find(val => val.id === data.id);
  });
});

// 状态保持
watch(
  activeTab,
  () => {
    router.replace({
      query: {
        type: activeTab.value,
      },
    });
  },
);
</script>

<template>
  <section class="flex-center resource-header">
    <section class="flex-center">
      <bk-select
        v-model="currentAccount"
        filterable
      >
        <bk-option
          v-for="item in accounts"
          :key="item.value"
          :value="item.value"
          :label="item.label"
        />
      </bk-select>
      <bk-select
        v-model="currentAccount"
        filterable
        class="ml10"
      >
        <bk-option
          v-for="item in accounts"
          :key="item.value"
          :value="item.value"
          :label="item.label"
        />
      </bk-select>
      <bk-button
        theme="primary"
        class="ml10"
        @click="handleDistribution"
      >
        {{ t('快速分配') }}
      </bk-button>
    </section>
    <section class="flex-center">
      <bk-checkbox
        v-model="isAccurate"
      >
        {{ t('精确') }}
      </bk-checkbox>
      <bk-search-select
        class="search-filter ml10"
        clearable
        :conditions="[]"
        :show-condition="false"
        :show-popover-tag-change="true"
        :data="computedSearchList"
        v-model="filter"
      />
    </section>
  </section>
  <bk-tab
    v-model:active="activeTab"
    type="card"
    class="resource-main g-scroller"
  >
    <bk-tab-panel
      v-for="item in tabs"
      :key="item.name"
      :name="item.name"
      :label="item.type"
    >
      <component
        :is="item.component"
        :filter="filter"
      />
    </bk-tab-panel>
  </bk-tab>

  <resource-distribution
    v-model:is-show="isShowDistribution"
    :choose-resource-type="true"
    :title="t('快速分配')"
  />
</template>

<style lang="scss" scoped>
.flex-center {
  display: flex;
  align-items: center;
}
.resource-header {
  justify-content: space-between;
  background: #fff;
  box-shadow: 1px 2px 3px 0 rgb(0 0 0 / 5%);
  padding: 20px;
}
.resource-main {
  margin-top: 20px;
  background: #fff;
  box-shadow: 1px 2px 3px 0 rgb(0 0 0 / 5%);
  height: calc(100vh - 236px);
  :deep(.bk-tab-content) {
    border-left: 1px solid #dcdee5;;
    border-right: 1px solid #dcdee5;;
    border-bottom: 1px solid #dcdee5;;
    padding: 20px;
  }
}
.search-filter {
  width: 500px;
}
</style>
