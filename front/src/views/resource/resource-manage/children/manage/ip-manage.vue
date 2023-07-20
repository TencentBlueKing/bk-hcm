<script setup lang="ts">
import type { FilterType } from '@/typings/resource';

import { PropType, h } from 'vue';
import { Button, InfoBox } from 'bkui-vue';
import { useResourceStore } from '@/store/resource';
import useDelete from '../../hooks/use-delete';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useSelection from '../../hooks/use-selection';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import { EipStatus, IEip } from '@/typings/business';
import { CLOUD_VENDOR } from '@/constants/resource';

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

// use hooks
const resourceStore = useResourceStore();

const { searchData, searchValue, filter } = useFilter(props);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'eips',
);

const columns = useColumns('eips');
const emit = defineEmits(['auth']);

const { selections, handleSelectionChange } = useSelection();

const {
  handleShowDelete,
  DeleteDialog,
} = useDelete(columns, selections, 'eips', '删除 EIP', true, 'delete', triggerApi);

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};

const canDelete = (data: IEip): boolean => {
  let res = false;
  // 分配到业务下面后不可删除
  const isInBusiness =    !props.authVerifyData?.permissionAction[
    props.isResourcePage ? 'iaas_resource_delete' : 'biz_iaas_resource_delete'
  ]
    || data.cvm_id
    || (data.bk_biz_id !== -1 && !location.href.includes('business'));
  const { status, vendor } = data;

  switch (vendor) {
    case CLOUD_VENDOR.tcloud:
      if (status === EipStatus.UNBIND) res = true;
      break;
    case CLOUD_VENDOR.huawei:
      if ([EipStatus.BIND_ERROR, EipStatus.DOWN, EipStatus.ERROR].includes(status)) res = true;
      break;
    case CLOUD_VENDOR.aws:
      if (status === EipStatus.UNBIND) res = true;
      break;
    case CLOUD_VENDOR.gcp:
      if (status === EipStatus.RESERVED) res = true;
      break;
    case CLOUD_VENDOR.azure:
      if (status === EipStatus.UNBIND) res = true;
      break;
  }
  return res && !isInBusiness;
};

const renderColumns = [
  ...columns,
  {
    label: '操作',
    render({ data }: any) {
      return h(h(
        'span',
        {
          onClick() {
            emit('auth', props.isResourcePage ? 'iaas_resource_delete' : 'biz_iaas_resource_delete');
          },
        },
        [
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              disabled: !canDelete(data),
              onClick() {
                InfoBox({
                  title: '请确认是否删除',
                  subTitle: `将删除【${data.id}】`,
                  theme: 'danger',
                  headerAlign: 'center',
                  footerAlign: 'center',
                  contentAlign: 'center',
                  onConfirm() {
                    resourceStore
                      .deleteBatch('eips', {
                        ids: [data.id],
                      })
                      .then(() => {
                        triggerApi();
                      });
                  },
                });
              },
            },
            ['删除'],
          ),
        ],
      ));
    },
  },
];

/**
 * 资源下，未绑定 且 未分配 可删除；
 * 业务下，未绑定 可删除；
 */
const isRowSelectEnable = ({ row }: { row: IEip }) => {
  if (!props.isResourcePage) return canDelete(row);
  if (row.id) {
    return row.bk_biz_id === -1 && canDelete(row);
  }
};

defineExpose({ fetchComponentsData });
</script>

<template>
  <bk-loading :loading="isLoading">
    <section
      class="flex-row align-items-center mb20"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleShowDelete(selections.filter(selection => canDelete(selection)).map((selection) => selection.id))"
      >
        删除
      </bk-button>
      <bk-search-select class="w500 ml10 mlauto" clearable :conditions="[]" :data="searchData" v-model="searchValue" />
    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="renderColumns"
      :data="datas"
      :is-row-select-enable="isRowSelectEnable"
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="handleSelectionChange"
      row-key="id"
    />
  </bk-loading>
  <delete-dialog />
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
.mlauto {
  margin-left: auto;
}
</style>
