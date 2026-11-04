<script setup lang="ts">
import type { DoublePlainObject, FilterType } from '@/typings/resource';
import { PropType, defineExpose, computed, reactive } from 'vue';
import { Message } from 'bkui-vue';
import { useAccountStore } from '@/store/account';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { ResourceTypeEnum } from '@/common/constant';

import useColumns from '../../hooks/use-columns';
import useQueryList from '../../hooks/use-query-list';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import useSelection from '../../hooks/use-selection';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import SyncAccountResource from '@/components/sync-account-resource/index.vue';

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

const { selections, handleSelectionChange, resetSelections } = useSelection();
const accountStore = useAccountStore();
const resourceAccountStore = useResourceAccountStore();
const businessId = computed(() => (props.isResourcePage ? undefined : accountStore.bizs));
const currentAccount = computed(() => (props.isResourcePage ? resourceAccountStore.resourceAccount : null));
const isSyncDisabled = computed(
  () => selections.value.length > 0 || (props.isResourcePage && !currentAccount.value?.id),
);
const syncDialogState = reactive<{ isShow: boolean; isHidden: boolean; initialModel: any }>({
  isShow: false,
  isHidden: true,
  initialModel: null,
});
const syncDialogDesc = computed(() =>
  props.isResourcePage
    ? '从云上同步该账号的子网数据，不会同步 VPC、负载均衡等其他资源'
    : '从云上同步该业务的子网数据，不会同步 VPC、负载均衡等其他资源',
);
const handleSync = () => {
  syncDialogState.initialModel = currentAccount.value
    ? { account_id: currentAccount.value.id, vendor: currentAccount.value.vendor }
    : null;
  syncDialogState.isHidden = false;
  syncDialogState.isShow = true;
};
const handleSyncHidden = () => {
  syncDialogState.isShow = false;
  syncDialogState.isHidden = true;
  syncDialogState.initialModel = null;
};
const handleSyncError = (error: { message?: string }) => {
  Message({ theme: 'error', message: error?.message || '同步失败' });
};

// const resourceStore = useResourceStore();
const { columns, settings } = useColumns('subnet');

const { searchData, searchValue, filter } = useFilter(props);

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

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'subnets',
);

const hostSearchData = computed(() => {
  return [
    {
      name: '子网ID',
      id: 'cloud_id',
    },
    ...searchData.value,
    {
      name: '所属VPC ID',
      id: 'cloud_vpc_id',
    },
  ];
});

// const emit = defineEmits(['auth']);

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};

// const handleDeleteSubnet = (data: any) => {
//   if (data.vendor === 'gcp') {
//     const subnetIds = [data.id];
//     const getRelateNum = (type: string, field = 'subnet_id', op = 'in') => {
//       return resourceStore.list(
//         {
//           page: {
//             count: true,
//           },
//           filter: {
//             op: 'and',
//             rules: [
//               {
//                 field,
//                 op,
//                 value: subnetIds,
//               },
//             ],
//           },
//         },
//         type,
//       );
//     };
//     Promise.all([
//       getRelateNum('cvms', 'subnet_ids', 'json_overlaps'),
//       getRelateNum('network_interfaces'),
//     ]).then(([cvmsResult, networkResult]: any) => {
//       if (cvmsResult?.data?.count || networkResult?.data?.count) {
//         const getMessage = (result: any, name: string) => {
//           if (result?.data?.count) {
//             return `${result?.data?.count}个${name}，`;
//           }
//           return '';
//         };
//         Message({
//           theme: 'error',
//           message: `该子网（name：${data.name}，id：${
//             data.id
//           }）关联${getMessage(cvmsResult, 'CVM')}${getMessage(
//             networkResult,
//             '网络接口',
//           )}不能删除`,
//         });
//       } else {
//         handledelete(data);
//       }
//     });
//   } else {
//     resourceStore.countSubnetIps(data.id as string).then((res: any) => {
//       if (res?.data?.used_ip_count) {
//         Message({
//           theme: 'error',
//           message: `该子网（name：${data.name}，id：${data.id} IPv4已经被使用${res?.data?.used_ip_count}不能删除`,
//         });
//       } else {
//         handledelete(data);
//       }
//     });
//   }
// };

// const handledelete = (data: any) => {
//   InfoBox({
//     title: '请确认是否删除',
//     subTitle: `将删除【${data.cloud_id}${data.name ? ` - ${data.name}` : ''}】`,
//     theme: 'danger',
//     headerAlign: 'center',
//     footerAlign: 'center',
//     contentAlign: 'center',
//     extCls: 'delete-resource-infobox',
//     onConfirm() {
//       resourceStore
//         .deleteBatch('subnets', {
//           ids: [data.id],
//         })
//         .then(() => {
//           Message({
//             theme: 'success',
//             message: '删除成功',
//           });
//           triggerApi();
//         });
//     },
//   });
// };

// const generateTooltipsOptions = (data: any) => {
//   const action_name = props.isResourcePage ? 'iaas_resource_delete' : 'biz_iaas_resource_delete';

//   if (!props.authVerifyData?.permissionAction[action_name]) return {
//     content: '当前用户无权限操作该按钮',
//     disabled: props.authVerifyData?.permissionAction[action_name],
//   };
//   if (props.isResourcePage && data?.bk_biz_id !== -1) return {
//     content: '该子网已分配到业务，仅可在业务下操作',
//     disabled: data.bk_biz_id === -1,
//   };

//   return {
//     disabled: true,
//   };
// };

const renderColumns = [
  ...columns,
  // {
  //   label: '操作',
  //   render({ data }: any) {
  //     return h(h(
  //       'span',
  //       {
  //         onClick() {
  //           emit(
  //             'auth',
  //             props.isResourcePage
  //               ? 'iaas_resource_delete'
  //               : 'biz_iaas_resource_delete',
  //           );
  //         },
  //       },
  //       [
  //         withDirectives(h(
  //           Button,
  //           {
  //             text: true,
  //             theme: 'primary',
  //             disabled:
  //                   !props.authVerifyData?.permissionAction[
  //                     props.isResourcePage
  //                       ? 'iaas_resource_delete'
  //                       : 'biz_iaas_resource_delete'
  //                   ] || (whereAmI.value !== Senarios.business && data.bk_biz_id !== -1),
  //             onClick() {
  //               handleDeleteSubnet(data);
  //             },
  //           },
  //           ['删除'],
  //         ), [
  //           [bkTooltips, generateTooltipsOptions(data)],
  //         ]),
  //       ],
  //     ));
  //   },
  // },
];

defineExpose({ fetchComponentsData });
</script>

<template>
  <bk-loading :loading="isLoading" opacity="1">
    <section class="toolbar" :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'">
      <slot></slot>
      <BatchDistribution
        :selections="selections"
        :type="DResourceType.subnets"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <bk-button :disabled="isSyncDisabled" @click="handleSync">同步子网</bk-button>
      <bk-search-select
        class="w500 ml10 search-selector-container"
        clearable
        :conditions="[]"
        :data="hostSearchData"
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
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>

  <template v-if="!syncDialogState.isHidden">
    <sync-account-resource
      v-model="syncDialogState.isShow"
      title="同步子网"
      :desc="syncDialogDesc"
      :resource-type="ResourceTypeEnum.SUBNET"
      :business-id="businessId"
      resource-name="subnet"
      :initial-model="syncDialogState.initialModel"
      :error-handler="handleSyncError"
      @hidden="handleSyncHidden"
    />
  </template>
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

.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;

  > *:not(.search-selector-container) {
    flex-shrink: 0;
  }
}

.search-selector-container {
  margin-left: auto;
  min-width: 0;
}
</style>
