<script setup lang="ts">
import type { FilterType } from '@/typings/resource';

import { PropType, h, computed, withDirectives } from 'vue';
import { useI18n } from 'vue-i18n';
import { bkTooltips, Button, InfoBox, Message } from 'bkui-vue';
import { useResourceStore } from '@/store/resource';
import useDelete from '../../hooks/use-delete';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import { VendorEnum } from '@/common/constant';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { useVerify } from '@/hooks';

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

const { t } = useI18n();

const { columns, settings } = useColumns('drive');
const simpleColumns = useColumns('drive', true).columns;
const resourceStore = useResourceStore();

const selectSearchData = computed(() => {
  return [
    {
      name: '云硬盘ID',
      id: 'cloud_id',
    },
    ...searchData.value,
  ];
});

const { searchData, searchValue, filter } = useFilter(props);

const emit = defineEmits(['auth']);

const isDisabledRecycle = (vendor: VendorEnum, status: string) => {
  let res = true;
  switch (vendor) {
    case VendorEnum.TCLOUD: {
      if (['UNATTACHED'].includes(status)) res = false;
      break;
    }
    case VendorEnum.HUAWEI: {
      if (['available'].includes(status)) res = false;
      break;
    }
    case VendorEnum.AWS: {
      if (!['attaching', 'attached', 'detaching'].includes(status)) res = false;
      break;
    }
    case VendorEnum.GCP: {
      if (!['CREATING', 'DELETING', 'RESTORING'].includes(status)) res = false;
      break;
    }
    case VendorEnum.AZURE: {
      if (['Unattached'].includes(status)) res = false;
      break;
    }
  }
  return res;
};

const unableRecycled = (data: { instance_id: string; vendor: VendorEnum; status: string; bk_biz_id: number }) => {
  return (
    !props.authVerifyData?.permissionAction[
      props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate'
    ] ||
    data.instance_id ||
    (props.isResourcePage && data.bk_biz_id !== -1) ||
    isDisabledRecycle(data?.vendor, data?.status)
  );
};

const generateTooltipsOptions = (data: any) => {
  const action_name = props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';

  if (!props.authVerifyData?.permissionAction[action_name])
    return {
      content: '当前用户无权限操作该按钮',
      disabled: props.authVerifyData?.permissionAction[action_name],
    };
  if (props.isResourcePage && data?.bk_biz_id !== -1)
    return {
      content: '该硬盘已分配到业务，仅可在业务下操作',
      disabled: data.bk_biz_id === -1,
    };
  if (data?.instance_id)
    return {
      content: '该硬盘已绑定主机，不可单独回收',
      disabled: !data.instance_id,
    };
  if (isDisabledRecycle(data?.vendor, data?.status))
    return {
      content: '该硬盘处于不可回收状态下',
      disabled: !isDisabledRecycle(data.vendor, data.status),
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
        h(
          'span',
          {
            onClick() {
              emit('auth', props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate');
            },
          },
          [
            withDirectives(
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  disabled: unableRecycled(data),
                  onClick() {
                    InfoBox({
                      title: '请确认是否回收',
                      subTitle: `将回收【${data.cloud_id}${data.name ? ` - ${data.name}` : ''}】`,
                      // @ts-ignore
                      theme: 'danger',
                      headerAlign: 'center',
                      footerAlign: 'center',
                      contentAlign: 'center',
                      extCls: 'recycle-resource-infobox',
                      onConfirm() {
                        resourceStore
                          .recycled('disks', {
                            infos: [{ id: data.id }],
                          })
                          .then(() => {
                            Message({
                              theme: 'success',
                              message: '回收成功',
                            });
                            triggerApi();
                          });
                      },
                    });
                  },
                },
                [t('回收')],
              ),
              [[bkTooltips, generateTooltipsOptions(data)]],
            ),
          ],
        ),
      );
    },
  },
];

const { selections, handleSelectionChange, resetSelections } = useSelection();

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'disks',
);

const { handleAuth } = useVerify();

const { handleShowDelete, DeleteDialog } = useDelete(
  simpleColumns,
  selections,
  'disks',
  '回收硬盘',
  true,
  'recycle',
  triggerApi,
);

/**
 * 资源下，未绑定 且 未分配 可删除；
 * 业务下，未绑定 可删除；
 */
const isRowSelectEnable = ({ row, isCheckAll }) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (!props.isResourcePage) return !unableRecycled(row);
  if (row.id) {
    return row.bk_biz_id === -1 && !unableRecycled(row);
  }
};
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
        :type="DResourceType.disks"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <bk-button
        class="mw88"
        :class="{ 'hcm-no-permision-btn': !authVerifyData?.permissionAction?.biz_iaas_resource_delete }"
        :disabled="authVerifyData?.permissionAction?.biz_iaas_resource_delete && selections.length <= 0"
        @click="
          () => {
            if (authVerifyData?.permissionAction?.biz_iaas_resource_delete) {
              handleShowDelete(
                selections
                  .filter((selection) => !isDisabledRecycle(selection?.vendor, selection?.status))
                  .map((selection) => selection.id),
              );
            } else {
              handleAuth('biz_iaas_resource_delete');
            }
          }
        "
      >
        {{ t('批量回收') }}
      </bk-button>
      <div class="flex-row align-items-center justify-content-arround mlauto">
        <bk-search-select
          class="w500"
          clearable
          :conditions="[]"
          :data="selectSearchData"
          v-model="searchValue"
          value-behavior="need-key"
        />
        <slot name="recycleHistory"></slot>
      </div>
    </section>

    <bk-table
      :settings="settings"
      row-hover="auto"
      remote-pagination
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      :pagination="pagination"
      :columns="renderColumns"
      :data="datas"
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
.mr15 {
  margin-right: 15px;
}
.mlauto {
  margin-left: auto;
}
</style>
