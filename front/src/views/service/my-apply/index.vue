<template>
  <div class="page-layout views-layout">
    <div class="left-layout">
      <LeftSide
        :list="applyList"
        :active="selectValue"
        :filter-data="filterData"
        :is-loading="isApplyLoading"
        :can-scroll-load="canScrollLoad"
        @on-change="handleChange"
        @on-filter-change="handleFilterChange"
        @on-load="handleLoadMore"
      />
    </div>
    <slot name="apply-right">
      <div class="right-layout">
        <component
          :key="comInfo.comKey"
          :is="currCom"
          :params="comInfo.currentApplyData"
          :loading="comInfo.cancelLoading"
          @on-cancel="handleCancel"
        />
      </div>
    </slot>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, reactive, ref, onMounted } from 'vue';
import LeftSide from './components/left-side/index.vue';
import ApplyDetail from './components/apply-detail/index.vue';
// 复杂数据结构先不按照泛型一个个定义，先从简
export default defineComponent({
  name: 'MyApply',
  components: {
    LeftSide,
    ApplyDetail,
  },
  setup() {
    const COM_MAP = Object.freeze(new Map([[['account_apply', 'service_apply'], 'ApplyDetail']]));

    const currCom = computed(() => {
      let com = '';
      for (const [key, value] of COM_MAP.entries()) {
        if (
          Object.keys(comInfo.currentApplyData).length > 0
          && key.includes(comInfo.currentApplyData.type)
        ) {
          com = value;
          break;
        }
      }
      return com;
    });

    const canScrollLoad = computed(() => {
      const { totalPage, currentBackup } = pagination;
      return totalPage > currentBackup;
    });

    // 做滚动分页用得到
    let pagination = reactive({
      totalPage: 0,
      current: 1,
      currentBackup: 1,
    });
    const filterData = ref([
      {
        label: '3天',
        value: 3,
      },
      {
        label: '一周',
        value: 7,
      },
      {
        label: '一个月',
        value: 30,
      },
      {
        label: '全部',
        value: '*',
      },
    ]);
    const selectValue = ref(3);
    const applyList = ref([
      {
        id: 94,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'pending',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 95,
        sn: 'REQ20230207000001',
        type: 'service_apply',
        applicant: 'poloohuang',
        status: 'pass',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入2',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 96,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'reject',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入3',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 97,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'cancelled',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 98,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'pending',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 88,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'pass',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 87,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'reject',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 86,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'cancelled',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
      {
        id: 85,
        sn: 'REQ20230207000001',
        type: 'account_apply',
        applicant: 'poloohuang',
        status: 'cancelled',
        created_time: '2023-02-07 16:10:45',
        resource: '账户申请',
        remarks: '账户接入',
        extra_info: {
          system_name: '持续集成平台',
        },
      },
    ]);
    const isApplyLoading = ref(false);
    let comInfo = reactive({
      comKey: -1,
      currentApplyData: {} as any,
      cancelLoading: false,
    });

    const handleFilterChange = (payload: Record<string, number>) => {
      selectValue.value = payload.value;
      pagination = Object.assign(pagination, { current: 1, currentBackup: 1 });
    };

    const handleChange = (payload: string) => {
      comInfo = Object.assign(comInfo, {
        currentApplyData: payload,
      });
    };

    const handleLoadMore = () => {
      pagination = Object.assign(pagination, { current: pagination.current + 1 });
    };

    const handleCancel = () => {
      comInfo.cancelLoading = true;
    };

    onMounted(() => {
      comInfo = Object.assign(comInfo,  { currentApplyData: applyList.value.length ? applyList.value[0] : {} });
    });

    return {
      filterData,
      applyList,
      selectValue,
      comInfo,
      currCom,
      isApplyLoading,
      handleFilterChange,
      handleChange,
      handleCancel,
      handleLoadMore,
      canScrollLoad,
    };
  },
});
</script>
<style lang="scss" scoped>
.page-layout {
  display: flex;
  padding: 0;
}

.views-layout {
  min-height: 100%;
  min-width: 1120px;
}

.left-layout {
  flex: 0 0 280px;
  height: calc(100vh - 61px);
  background: #fff;
  overflow: hidden;
}

.right-layout {
  padding: 30px;
  flex: 1 0 auto;
  width: calc(100% - 280px);
  height: calc(100vh - 61px);
  background: #f5f6fa;
  overflow-y: auto;
}
</style>
