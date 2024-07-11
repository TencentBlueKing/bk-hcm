<script setup lang="ts">
import type { FilterType } from '@/typings/resource';
import { PropType, defineExpose, computed } from 'vue';
// import { Message, InfoBox } from 'bkui-vue';
// import { useResourceStore } from '@/store/resource';

import useColumns from '../../hooks/use-columns';
import useQueryList from '../../hooks/use-query-list';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import useSelection from '../../hooks/use-selection';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';

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
    ...[
      {
        name: '所属VPC ID',
        id: 'cloud_vpc_id',
      },
      {
        name: '云地域',
        id: 'region',
      },
    ],
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
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
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
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
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
.search-selector-container {
  margin-left: auto;
}
</style>
