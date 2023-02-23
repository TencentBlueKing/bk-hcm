<template>
  <bk-loading
    :opacity="1"
    :loading="pageLoading"
  >
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
            :key="state.comInfo.comKey"
            :is="currCom"
            :params="state.comInfo.currentApplyData"
            :loading="state.comInfo.loading"
            :cancel-loading="state.comInfo.cancelLoading"
            @on-cancel="handleCancel"
          />
        </div>
      </slot>
    </div>
  </bk-loading>
</template>

<script lang="ts">
import { defineComponent, computed, reactive, ref, onMounted } from 'vue';
import LeftSide from './components/left-side/index.vue';
import ApplyDetail from './components/apply-detail/index.vue';
import { useAccountStore } from '@/store';
// 复杂数据结构先不按照泛型一个个定义，先从简
export default defineComponent({
  name: 'MyApply',
  components: {
    LeftSide,
    ApplyDetail,
  },
  setup() {
    const accountStore = useAccountStore();
    const COM_MAP = Object.freeze(new Map([[['add_account', 'service_apply'], 'ApplyDetail']]));

    const currCom = computed(() => {
      let com = '';
      for (const [key, value] of COM_MAP.entries()) {
        if (
          Object.keys(state.comInfo.currentApplyData).length > 0
          && key.includes(state.comInfo.currentApplyData.type)
        ) {
          com = value;
          break;
        }
      }
      return com;
    });

    const pageLoading = computed(() => {
      return initRequestQueue.value.length > 0;
    });

    const filterParams = ref({
      filter: {
        op: 'and',
        rules: [
          {
            field: 'created_at',
            op: 'gt',
            value: '2020-01-02 00:00:00',
          },
        ],
      },
      page: {
        count: false,
        start: 0,
        limit: 10,
      },
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
    const applyList = ref([]);
    const id = ref('');
    const isApplyLoading = ref(false);
    const initRequestQueue = ref(['applyList', 'applyDetail']);
    const state = reactive({
      comInfo: {
        comKey: -1,
        currentApplyData: {} as any,
        loading: false,
        cancelLoading: false,
      },
    });

    const handleFilterChange = (payload: Record<string, number>) => {
      selectValue.value = payload.value;
      pagination = Object.assign(pagination, { current: 1, currentBackup: 1 });
    };

    const handleChange = (id: string) => {
      getMyApplyDetail(id);
    };

    const handleLoadMore = () => {
      pagination = Object.assign(pagination, { current: pagination.current + 1 });
    };

    const handleCancel = async (id: string) => {
      state.comInfo.cancelLoading = true;
      const res = await accountStore.cancelApplyAccount(id);
      getMyApplyDetail(id);
      state.comInfo.cancelLoading = false;
      console.log(res);
    };

    // 获取我的申请列表
    const getMyApplyList = async () => {
      try {
        isApplyLoading.value = true;
        const res = await accountStore.getApplyAccountList(filterParams.value);
        applyList.value = res.data.details;
        id.value = res.data.details[0].id;
        getMyApplyDetail(id.value);
      } catch (error) {
        console.log('error', error);
      }  finally {
        isApplyLoading.value = false;
        initRequestQueue.value.length > 0 && initRequestQueue.value.shift();
      }
    };

    // 获取我的申请详情
    const getMyApplyDetail = async (id: string) => {
      state.comInfo.loading = true;
      try {
        const res = await accountStore.getApplyAccountDetail(id);
        state.comInfo.currentApplyData = res.data;
        state.comInfo.comKey = res.data.id;
      } catch (error) {
        console.log('error', error);
      } finally {
        initRequestQueue.value.length > 0 && initRequestQueue.value.shift();
        state.comInfo.loading = false;
      }
    };

    onMounted(() => {
      getMyApplyList();
    });

    return {
      filterData,
      applyList,
      selectValue,
      state,
      currCom,
      isApplyLoading,
      handleFilterChange,
      handleChange,
      handleCancel,
      handleLoadMore,
      pageLoading,
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
