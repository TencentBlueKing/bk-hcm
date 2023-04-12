<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  defineExpose,
  h,
  computed,
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
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';

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

const emit = defineEmits(['auth']);


const hostSearchData = computed(() => {
  return [
    ...searchData.value,
    ...[{
      name: '蓝鲸云区域',
      id: 'bk_cloud_id',
    }, {
      name: '云地域',
      id: 'region',
    }],
  ];
});

const {
  searchData,
  searchValue,
} = useFilter(props);

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
      );
  };
  Promise
    .all([
      getRelateNum('cvms', 'vpc_ids', 'json_overlaps'),
      getRelateNum('subnets'),
      getRelateNum('route_tables'),
      getRelateNum('network_interfaces'),
    ])
    .then(([cvmsResult, subnetsResult, routeResult, networkResult]: any) => {
      // eslint-disable-next-line max-len
      if (cvmsResult?.data?.count || subnetsResult?.data?.count || routeResult?.data?.count || networkResult?.data?.count) {
        const getMessage = (result: any, name: string) => {
          if (result?.data?.count) {
            return `${result?.data?.count}个${name}，`;
          }
          return '';
        };
        Message({
          theme: 'error',
          message: `该VPC（name：${data.name}，id：${data.id}）关联${getMessage(cvmsResult, 'CVM')}${getMessage(subnetsResult, '子网')}${getMessage(routeResult, '路由表')}${getMessage(networkResult, '网络接口')}不能删除`,
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
            resourceStore
              .delete(
                'vpcs',
                data.id,
              ).then(() => {
                Message({
                  theme: 'success',
                  message: '删除成功',
                });
              });
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
      return h(h(
        'span',
        {
          onClick() {
            emit('auth', props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate');
          },
        },
        [
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              disabled: !props.authVerifyData?.permissionAction[props.isResourcePage ? 'iaas_resource_delete' : 'biz_iaas_resource_delete'],
              onClick() {
                handleDeleteVpc(data);
              },
            },
            [
              t('删除'),
            ],
          ),
        ],
      ));
    },
  },
];
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'">
      <slot>
      </slot>
      <bk-search-select
        class="w500 ml10"
        clearable
        :data="hostSearchData"
        v-model="searchValue"
      />
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
