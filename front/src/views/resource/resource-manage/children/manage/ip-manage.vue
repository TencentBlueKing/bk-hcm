<script setup lang="ts">
import type { FilterType } from '@/typings/resource';

import { PropType, h, computed, withDirectives } from 'vue';
import { bkTooltips, Button, InfoBox, Message } from 'bkui-vue';
import { useResourceStore } from '@/store/resource';
import useDelete from '../../hooks/use-delete';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useSelection from '../../hooks/use-selection';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import { EipStatus, IEip } from '@/typings/business';
import { CLOUD_VENDOR } from '@/constants/resource';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { AUTH_BIZ_DELETE_IAAS_RESOURCE, AUTH_DELETE_IAAS_RESOURCE } from '@/constants/auth-symbols';
import HcmAuth from '@/components/auth/auth.vue';

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
  bkBizId: Number,
});

// use hooks
const resourceStore = useResourceStore();

const { searchData, searchValue, filter } = useFilter(props);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'eips',
);

const selectSearchData = computed(() => {
  return [
    {
      name: 'IP资源ID',
      id: 'cloud_id',
    },
    ...searchData.value,
  ];
});

const { columns, settings } = useColumns('eips');

const { selections, handleSelectionChange, resetSelections } = useSelection();

const { handleShowDelete, DeleteDialog } = useDelete(
  columns,
  selections,
  'eips',
  '删除 EIP',
  true,
  'delete',
  triggerApi,
);

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};

const hasNoRelateResource = ({ vendor, status }: IEip): boolean => {
  let res = false;
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
  return res;
};
const deleteAuthType = computed(() =>
  props.isResourcePage ? AUTH_DELETE_IAAS_RESOURCE : AUTH_BIZ_DELETE_IAAS_RESOURCE,
);
const canDelete = (data: IEip): boolean => {
  // 分配到业务下面后不可删除
  const isInBusiness = data.cvm_id || (data.bk_biz_id !== -1 && !location.href.includes('business'));
  return hasNoRelateResource(data) && !isInBusiness;
};

const generateTooltipsOptions = (data: IEip) => {
  if (props.isResourcePage && data?.bk_biz_id !== -1)
    return {
      content: '该弹性IP已分配到业务，仅可在业务下操作',
      disabled: data.bk_biz_id === -1,
    };
  if (data?.cvm_id || !hasNoRelateResource(data))
    return {
      content: '该弹性IP已绑定资源，不可以删除',
      disabled: !(data?.cvm_id || !hasNoRelateResource(data)),
    };

  return {
    disabled: true,
  };
};

const renderColumns = [
  ...columns,
  {
    label: '操作',
    render({ data }: any) {
      return h(
        HcmAuth,
        { sign: { type: deleteAuthType.value, relation: [props.bkBizId] } },
        {
          default: ({ noPerm }: { noPerm: boolean }) =>
            withDirectives(
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  disabled: noPerm || !canDelete(data),
                  onClick() {
                    InfoBox({
                      title: '请确认是否删除',
                      subTitle: `将删除【${data.cloud_id}${data.name ? ` - ${data.name}` : ''}】`,
                      theme: 'danger',
                      headerAlign: 'center',
                      footerAlign: 'center',
                      contentAlign: 'center',
                      extCls: 'delete-resource-infobox',
                      async onConfirm() {
                        await resourceStore.deleteBatch('eips', { ids: [data.id] });
                        Message({ theme: 'success', message: '删除成功' });
                        triggerApi();
                      },
                    });
                  },
                },
                ['删除'],
              ),
              [[bkTooltips, generateTooltipsOptions(data)]],
            ),
        },
      );
    },
  },
];

/**
 * 资源下，未绑定 且 未分配 可删除；
 * 业务下，未绑定 可删除；
 */
const isRowSelectEnable = ({ row, isCheckAll }: { row: IEip; isCheckAll: boolean }) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (!props.isResourcePage) return canDelete(row);
  if (row.id) {
    return row.bk_biz_id === -1 && canDelete(row);
  }
};

defineExpose({ fetchComponentsData });
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
        :type="DResourceType.eips"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <hcm-auth :sign="{ type: deleteAuthType, relation: [props.bkBizId] }" v-slot="{ noPerm }">
        <bk-button
          class="mw88"
          :disabled="selections.length <= 0 || noPerm"
          @click="
            handleShowDelete(selections.filter((selection) => canDelete(selection)).map((selection) => selection.id))
          "
        >
          批量删除
        </bk-button>
      </hcm-auth>

      <bk-search-select
        class="w500 ml10 mlauto"
        clearable
        :conditions="[]"
        :data="selectSearchData"
        v-model="searchValue"
        value-behavior="need-key"
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
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
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
