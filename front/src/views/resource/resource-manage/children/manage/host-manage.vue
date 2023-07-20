<script setup lang="ts">
import type {
  // PlainObject,
  DoublePlainObject,
  FilterType,
} from '@/typings/resource';
import {
  Message,
} from 'bkui-vue';

import {
  PropType,
  h,
  ref,
  computed,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
// import {
//   AngleRight,
// } from 'bkui-vue/lib/icon';
// import useShutdown from '../../hooks/use-shutdown';
// import useReboot from '../../hooks/use-reboot';
// import usePassword from '../../hooks/use-password';
// import useRefund from '../../hooks/use-refund';
// import useBootUp from '../../hooks/use-boot-up';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import useFilter  from '@/views/resource/resource-manage/hooks/use-filter';
import { HostCloudEnum, CloudType } from '@/typings';
import {
  useResourceStore,
} from '@/store';
import HostOperations from '../../common/table/HostOperations';
import { useBusinessMapStore } from '@/store/useBusinessMap';

// use hook
const {
  t,
} = useI18n();

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  whereAmI: {
    type: String,
  },
});

const resourceStore = useResourceStore();

const isLoadingCloudAreas = ref(false);
const cloudAreaPage = ref(0);
const cloudAreas = ref([]);

const {
  searchData,
  searchValue,
  filter,
} = useFilter(props);

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
  triggerApi,
} = useQueryList({ filter: filter.value }, 'cvms');

const {
  selections,
  handleSelectionChange,
  resetSelections,
} = useSelection();

const isShowDistribution = ref(false);
const businessId = ref('');
const businessList = ref(useBusinessMapStore().businessList);
const columns = useColumns('cvms');

const hostSearchData = computed(() => {
  return [
    ...searchData.value,
    ...[{
      name: '管控区域',
      id: 'bk_cloud_id',
    }, {
      name: '操作系统',
      id: 'os_name',
    }, {
      name: '云地域',
      id: 'region',
    }],
  ];
});

const distribColumns = [
  {
    label: 'ID',
    field: 'id',
  },
  {
    label: '实例 ID',
    field: 'cloud_id',
  },
  {
    label: '云厂商',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          CloudType[data.vendor],
        ],
      );
    },
  },
  {
    label: '地域',
    field: 'region',
  },
  {
    label: '名称',
    field: 'name',
  },
  {
    label: '状态',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          HostCloudEnum[data.status] || data.status,
        ],
      );
    },
  },
  {
    label: '操作系统',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.os_name || '--',
        ],
      );
    },
  },
  {
    label: '云区域ID',
    field: 'bk_cloud_id',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.bk_cloud_id === -1 ? '未分配' : data.bk_cloud_id,
        ],
      );
    },
  },
];


const distributionCvm = async () => {
  const cvmIds = selections.value.map(e => e.id);
  try {
    await resourceStore.cvmAssignBizs({ cvm_ids: cvmIds, bk_biz_id: businessId.value });
    Message({
      message: t('操作成功'),
      theme: 'success',
    });
  } catch (error) {
    console.log(error);
  } finally {
  }
};

const handleDistributionConfirm = () => {
  isShowDistribution.value = true;
  distributionCvm();
};

const isRowSelectEnable = ({ row }: DoublePlainObject) => {
  if (!props.isResourcePage) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};

const getCloudAreas = () => {
  if (isLoadingCloudAreas.value) return;
  isLoadingCloudAreas.value = true;
  resourceStore
    .getCloudAreas({
      page: {
        start: cloudAreaPage.value,
        limit: 100,
      },
    })
    .then((res: any) => {
      cloudAreaPage.value += 1;
      cloudAreas.value.push(...res?.data?.info || []);
    })
    .finally(() => {
      isLoadingCloudAreas.value = false;
    });
};
getCloudAreas();

</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'">
      <slot></slot>
      <HostOperations :selections="selections" :on-finished="() => {
        triggerApi();
        resetSelections();
      }"></HostOperations>

      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <bk-search-select
          class="w500 ml10 mr15"
          clearable
          :conditions="[]"
          :data="hostSearchData"
          v-model="searchValue"
        />
        <slot name="recycleHistory"></slot>
      </div>

    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      :columns="columns"
      :data="datas"
      :pagination="pagination"
      remote-pagination
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @selection-change="handleSelectionChange"
      @column-sort="handleSort"
      row-key="id"
    />

    <bk-dialog
      v-model:is-show="isShowDistribution"
      width="820"
      :title="t('主机分配')"
      theme="primary"
      quick-close
      @confirm="handleDistributionConfirm">
      <section class="distribution-cls">
        目标业务
        <bk-select
          class="ml20"
          v-model="businessId"
          filterable
        >
          <bk-option
            v-for="item in businessList"
            :key="item.id"
            :value="item.id"
            :label="item.name"
          />
        </bk-select>
      </section>
      <bk-table
        class="mt20"
        row-hover="auto"
        :columns="distribColumns"
        :data="selections"
        show-overflow-tooltip
      />
    </bk-dialog>

  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
.distribution-cls{
  display: flex;
  align-items: center;
}
.mr15 {
  margin-right: 15px;
}
.search-selector-container {
  margin-left: auto;
}
</style>
