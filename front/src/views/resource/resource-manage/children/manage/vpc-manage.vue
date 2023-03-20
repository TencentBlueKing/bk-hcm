<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  defineExpose,
  h,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import {
  InfoBox,
  Message,
  Button,
} from 'bkui-vue';
import {
  useResourceStore,
} from '@/store/resource';
import useColumns from '../../hooks/use-columns';
import useQueryList from '../../hooks/use-query-list';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

// use hooks
const {
  t,
} = useI18n();
const resourceStore = useResourceStore();
const columns = useColumns('vpc');
const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'vpcs');

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};
defineExpose({ fetchComponentsData });

const handleDeleteVpc = (data: any) => {
  const vpcIds = [data.id];
  const getRelateNum = (type: string, field = 'vpc_id', op = 'in') => {
    return resourceStore
      .list(
        {
          page: {
            count: true,
          },
          filter: {
            op: 'and',
            rules: [{
              field,
              op,
              value: vpcIds,
            }],
          },
        },
        type,
      )
  }
  Promise
    .all([
      getRelateNum('cvms', 'vpc_ids', 'json_overlaps'),
      getRelateNum('subnets'),
      getRelateNum('route_tables'),
      getRelateNum('network_interfaces'),
    ])
    .then(([cvmsResult, subnetsResult, routeResult, networkResult]: any) => {
      if (cvmsResult?.data?.count || subnetsResult?.data?.count || routeResult?.data?.count || networkResult?.data?.count) {
        const getMessage = (result: any, name: string) => {
          if (result?.data?.count) {
            return `${result?.data?.count}个${name}，`
          }
          return ''
        }
        Message({
          theme: 'error',
          message: `该VPC关联${getMessage(cvmsResult, 'CVM')}${getMessage(subnetsResult, '子网')}${getMessage(routeResult, '路由表')}${getMessage(networkResult, '网络接口')}不能删除`
        })
      } else {
        InfoBox({
          title: '请确认是否删除',
          subTitle: `将删除【${data.name}】`,
          theme: 'danger',
          headerAlign: 'center',
          footerAlign: 'center',
          contentAlign: 'center',
          onConfirm() {
            resourceStore
              .delete(
                'vpcs',
                data.id,
              );
          },
        });
      }
    });
};

const renderColumns = [
  ...columns,
  {
    label: '操作',
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            handleDeleteVpc(data)
          },
        },
        [
          t('删除'),
        ],
      );
    },
  }
]
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section>
      <slot>
      </slot>
    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="renderColumns"
      :data="datas"
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
</style>
