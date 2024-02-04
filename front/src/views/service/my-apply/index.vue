<template>
  <bk-loading :opacity="1" :loading="pageLoading" class="my-apply-page">
    <div class="page-layout views-layout">
      <div class="left-layout">
        <LeftSide
          :list="applyList"
          :active="selectValue"
          :filter-data="filterData"
          :list-loading="isApplyLoading"
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
import moment from 'moment';
// 复杂数据结构先不按照泛型一个个定义，先从简
export default defineComponent({
  name: 'MyApply',
  components: {
    LeftSide,
    ApplyDetail,
  },
  setup() {
    const accountStore = useAccountStore();
    const COM_MAP = Object.freeze(
      new Map([[['add_account', 'service_apply', 'create_cvm', 'create_disk', 'create_vpc'], 'ApplyDetail']]),
    );

    const currCom = computed(() => {
      let com = '';
      for (const [key, value] of COM_MAP.entries()) {
        if (
          Object.keys(state.comInfo.currentApplyData).length > 0 &&
          key.includes(state.comInfo.currentApplyData.type)
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

    const filterParams = ref<any>({
      filter: {
        op: 'and',
        rules: [
          {
            field: 'created_at',
            op: 'gt',
            value: '2006-01-02T15:04:05Z',
          },
        ],
      },
      page: {
        count: false,
        start: 0,
        limit: 10,
        sort: 'id',
        order: 'DESC',
      },
    });

    const canScrollLoad = computed(() => {
      return backupData.value.length >= 10;
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
    const backupData = ref([]);
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

    const handleFilterChange = (payload: any) => {
      isApplyLoading.value = true;
      selectValue.value = payload.value;
      paramsReset();
      if (payload.value === '*') {
        filterParams.value.filter = { op: 'and', rules: [] };
      } else {
        filterParams.value.filter = {
          op: 'and',
          rules: [
            {
              field: 'created_at',
              op: 'gt',
              value: '',
            },
          ],
        };
        const value = moment().add(-payload.value, 'd').format('YYYY-MM-DD HH:mm:ss');
        const time = new Date(value).toISOString().replace('.000Z', 'Z');
        filterParams.value.filter.rules[0].value = time;
      }
      getMyApplyList();
    };

    // 参数还原
    const paramsReset = () => {
      applyList.value = [];
      filterParams.value.page = {
        count: false,
        start: 0,
        limit: 10,
        sort: 'id',
        order: 'DESC',
      };
    };

    const handleChange = (id: string) => {
      getMyApplyDetail(id);
    };

    // 滚动加载更多
    const handleLoadMore = async () => {
      filterParams.value.page.start = (filterParams.value.page.start + 1) * 10;
      try {
        isApplyLoading.value = true;
        const res = await accountStore.getApplyAccountList(filterParams.value);
        backupData.value = res.data.details;
        applyList.value.push(...res.data.details);
      } catch (error) {
        console.log('error', error);
      } finally {
        isApplyLoading.value = false;
      }
    };

    // 撤销
    const handleCancel = async (id: string) => {
      state.comInfo.cancelLoading = true;
      try {
        await accountStore.cancelApplyAccount(id);
        getMyApplyDetail(id);
        changeApplyitemStatus(id);
      } catch (error) {
        console.log(error);
      } finally {
        state.comInfo.cancelLoading = false;
      }
    };

    // 获取我的申请列表
    const getMyApplyList = async () => {
      try {
        const res = await accountStore.getApplyAccountList(filterParams.value);
        backupData.value = res.data.details;
        applyList.value.push(...res.data.details);
        id.value = res.data.details[0]?.id;
        if (id.value) {
          getMyApplyDetail(id.value);
        } else {
          initRequestQueue.value.length > 0 && initRequestQueue.value.shift();
        }
      } catch (error) {
        console.log('error', error);
      } finally {
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

    // 改变当前撤销的状态
    const changeApplyitemStatus = async (id: string) => {
      applyList.value = applyList.value.map((e) => {
        if (e.id === id) {
          e.status = 'cancelled';
        }
        return e;
      });
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
.my-apply-page {
  height: 100%;
}
.page-layout {
  height: 100%;
  display: flex;
  padding: 0;
}

.views-layout {
  min-height: 100%;
  min-width: 1120px;
}

.left-layout {
  flex: 0 0 280px;
  background: #fff;
  overflow: hidden;
  :deep(.bk-nested-loading) {
    height: 100%;
  }
}

.right-layout {
  padding: 30px;
  flex: 1 0 auto;
  width: calc(100% - 280px);
  background: #f5f6fa;
  overflow-y: auto;
}
</style>
