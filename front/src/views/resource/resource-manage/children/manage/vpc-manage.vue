<script setup lang="ts">
import type { FilterType } from '@/typings/resource';

import { PropType, defineExpose, h, computed, ref, onMounted } from 'vue';
import { Button, Message } from 'bkui-vue';
import { useResourceStore } from '@/store/resource';
import useColumns from '../../hooks/use-columns';
import useQueryList from '../../hooks/use-query-list';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import useSelection from '../../hooks/use-selection';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  authVerifyData: {
    type: Object as PropType<any>,
  },
  whereAmI: {
    type: String,
  },
});

const { selections, handleSelectionChange, resetSelections } = useSelection();

const { whereAmI } = useWhereAmI();

// use hooks
// const { t } = useI18n();
const resourceStore = useResourceStore();
const { columns, settings } = useColumns('vpc');
const { searchData, searchValue, filter } = useFilter(props);
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'vpcs',
);

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};
defineExpose({ fetchComponentsData });

const emit = defineEmits(['auth']);
const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (!props.isResourcePage) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};
const curVpc = ref({});
const isDialogShow = ref(false);
const isDialogBtnLoading = ref(false);
const cloudAreaList = ref([]);
const curCloudArea = ref('');

const hostSearchData = computed(() => {
  return [
    {
      name: 'VPC ID',
      id: 'cloud_id',
    },
    ...searchData.value,
    ...[
      {
        name: '管控区域',
        id: 'bk_cloud_id',
      },
      {
        name: '云地域',
        id: 'region',
      },
    ],
  ];
});
const handleBindRegion = (data: any) => {
  isDialogShow.value = true;
  curVpc.value = data;
  curCloudArea.value = data.bk_cloud_id === -1 ? '' : data.bk_cloud_id;
};

const handleConfirm = async () => {
  isDialogBtnLoading.value = true;
  try {
    await resourceStore.bindVPCWithCloudArea([
      {
        vpc_id: curVpc.value.id,
        bk_cloud_id: curCloudArea.value,
      },
    ]);
    triggerApi();
    Message({
      message: '绑定成功',
      theme: 'success',
    });
  } finally {
    isDialogShow.value = false;
    isDialogBtnLoading.value = false;
  }
};

const getCloudAreas = async () => {
  isDialogBtnLoading.value = true;
  try {
    const res = await resourceStore.getCloudAreas({
      page: {
        start: 0,
        limit: 500,
      },
    });
    cloudAreaList.value = res.data?.info || [];
  } finally {
    isDialogBtnLoading.value = false;
  }
};

onMounted(() => {
  getCloudAreas();
});

/* const handleDeleteVpc = (data: any) => {
  const vpcIds = [data.id];
  const getRelateNum = (type: string, field = 'vpc_id', op = 'in') => {
    return resourceStore.list(
      {
        page: {
          count: true,
        },
        filter: {
          op: 'and',
          rules: [
            {
              field,
              op,
              value: vpcIds,
            },
          ],
        },
      },
      type,
    );
  };
  Promise.all([getRelateNum('subnets')]).then(([subnetsResult]) => {
    // eslint-disable-next-line max-len
    if (subnetsResult?.data?.count) {
      const getMessage = (result: any, name: string) => {
        if (result?.data?.count) {
          return `${result?.data?.count}个${name}，`;
        }
        return '';
      };
      Message({
        theme: 'error',
        message: `该VPC（id：${data.id}）关联${getMessage(
          subnetsResult,
          '子网',
        )}不能删除`,
      });
    } else {
      InfoBox({
        title: '请确认是否删除',
        subTitle: `将删除【${data.name}】`,
        theme: 'danger',
        headerAlign: 'center',
        footerAlign: 'center',
        contentAlign: 'center',
        onConfirm() {
          resourceStore.delete('vpcs', data.id).then(() => {
            triggerApi();
            Message({
              theme: 'success',
              message: '删除成功',
            });
          });
        },
      });
    }
  });
};*/

const renderColumns = [
  ...columns,
  {
    label: '操作',
    render({ data }: any) {
      return h(
        h('span', [
          whereAmI.value === Senarios.resource
            ? h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  class: `mr16 ${
                    props.authVerifyData?.permissionAction.iaas_resource_operate ? '' : 'hcm-no-permision-text-btn'
                  }`,
                  disabled: data.bk_cloud_id !== -1,
                  onClick() {
                    if (props.authVerifyData?.permissionAction.iaas_resource_operate) handleBindRegion(data);
                    else {
                      emit('auth', 'iaas_resource_operate');
                    }
                  },
                },
                ['绑定管控区'],
              )
            : null,
          // h(
          //   Popover,
          //   {
          //     content: 'VPC下有子网正在使用，不能直接删除',
          //     disabled: !(whereAmI.value === Senarios.resource && data.bk_cloud_id !== -1),
          //   },
          //   [
          //     h(
          //       Button,
          //       {
          //         text: true,
          //         theme: 'primary',
          //         disabled:
          //         !props.authVerifyData?.permissionAction[
          //           props.isResourcePage
          //             ? 'iaas_resource_delete'
          //             : 'biz_iaas_resource_delete'
          //         ] || (whereAmI.value === Senarios.resource && data.bk_cloud_id !== -1),
          //         onClick() {
          //           handleDeleteVpc(data);
          //         },
          //       },
          //       [t('删除')],
          //     ),
          //   ],
          // ),
        ]),
      );
    },
  },
];
</script>

<template>
  <bk-loading :loading="isLoading" opacity="1">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <BatchDistribution
        :selections="selections"
        :type="DResourceType.vpcs"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <bk-search-select
        class="w500 ml10 search-selector-container"
        clearable
        :conditions="[]"
        :data="hostSearchData"
        v-model="searchValue"
      />
    </section>

    <bk-table
      :settings="settings"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="renderColumns"
      :data="datas"
      :is-row-select-enable="isRowSelectEnable"
      show-overflow-tooltip
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>

  <bk-dialog
    :is-show="isDialogShow"
    title="VPC绑定管控区"
    :theme="'primary'"
    quick-close
    @closed="() => (isDialogShow = false)"
    @confirm="handleConfirm"
    :is-loading="isDialogBtnLoading"
  >
    <p class="bind-vpc-tips">
      注意：VPC绑定管控区后，VPC下的主机，会默认鄉定到该管控区。
      <br />
      当主机分配到业务后，主机也将同步到配置平台的该业务。
    </p>
    <bk-form>
      <bk-form-item label="VPC名称">
        {{ curVpc.name || '--' }}
      </bk-form-item>
      <bk-form-item label="管控区名称">
        <bk-select v-model="curCloudArea" :input-search="false" filterable>
          <bk-option v-for="(item, index) in cloudAreaList" :key="index" :value="item.id" :label="item.name" />
        </bk-select>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.search-selector-container {
  margin-left: auto;
}
.bind-vpc-tips {
  font-size: 12px;
  color: #979ba5;
  margin-bottom: 8px;
}
</style>
