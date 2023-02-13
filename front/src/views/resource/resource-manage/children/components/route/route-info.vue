<script lang="ts" setup>
import {
  ref,
  watch,
} from 'vue';
import DetailInfo from '../../../common/info/detail-info';
import {
  useResourceStore,
} from '@/store/resource';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const resourceStore = useResourceStore();

// 基本信息字段配置
const fileds = ref<any[]>([
  {
    name: '路由表名称',
    prop: 'name',
  },
  {
    name: '路由表 ID',
    prop: 'id',
  },
  {
    name: '账号',
    prop: 'account_id',
    link(val: string) {
      return `/#/resource/account/detail/?id=${val}`;
    },
  },
]);
// 路由列表字段配置
const columns = ref<any[]>([]);
// 路由列表数据
const datas = ref([]);
// 加载状态
const isLoading = ref(false);
// 分页状态
const pagination = ref({
  current: 1,
  limit: 10,
  count: 0,
});

watch(
  () => props.detail,
  () => {
    switch (props.detail.vendor) {
      case 'tcloud':
        fileds.value.push(...[
          {
            name: '所属网络',
            prop: 'vpc_id',
            render (val: string) {
              return `VPC-${val}`
            }
          },
          {
            name: '地域',
            prop: '',
          },
          {
            name: '地域ID',
            prop: 'region',
          },
          {
            name: '路由表类型',
            value: '默认路由表'
          },
          {
            name: '备注',
            prop: 'memo',
          },
          {
            name: '创建时间',
            prop: 'created_at',
          },
        ]);
        columns.value.push(...[
          {
            label: '目的地址',
            field: 'destination_cidr_block',
          },
          {
            label: '下一跳类型',
            field: 'gateway_type',
          },
          {
            label: '下一跳地址',
            field: 'cloud_gateway_id',
          },
          {
            label: '状态',
            field: 'enabled',
            render({ cell }: { cell: string }) {
              return cell ? '启用' : '禁用'
            },
          },
          {
            label: '备注',
            field: 'memo',
          },
        ]);
        break;
      case 'azure':
        fileds.value.push(...[
          {
            name: '资源组',
            value: 'resource_group'
          },
          {
            name: '地域',
            prop: '',
          },
          {
            name: '地域ID',
            prop: 'region',
          },
          {
            name: '订阅',
            value: '默认路由表'
          },
          {
            name: '订阅ID',
            prop: 'memo',
          },
          {
            name: '备注',
            prop: 'memo',
          },
          {
            name: '创建时间',
            prop: 'created_at',
          },
        ]);
        columns.value.push(...[
          {
            label: '名称',
            field: 'name',
          },
          {
            label: '目的地址',
            field: 'address_prefix',
          },
          {
            label: '下一跳类型',
            field: 'next_hop_type',
          },
          {
            label: '下一跳地址',
            field: 'next_hop_ip_address',
          },
        ]);
        break;
      case 'aws':
        fileds.value.push(...[
          {
            name: '所属网络',
            prop: 'vpc_id',
            render (val: string) {
              return `VPC-${val}`
            }
          },
          {
            name: '地域',
            prop: '',
          },
          {
            name: '地域ID',
            prop: 'region',
          },
          {
            name: '路由表类型',
            value: '主路由表'
          },
          {
            name: '备注',
            prop: 'memo',
          },
          {
            name: '创建时间',
            prop: 'created_at',
          },
        ]);
        columns.value.push(...[
          {
            label: '目的地址',
            field: 'destination_cidr_block',
          },
          {
            label: '下一跳地址',
            field: 'cloud_carrier_gateway_id',
          },
          {
            label: '状态',
            field: 'state',
            render({ cell }: { cell: string }) {
              return cell === 'active' ? '可用' : '路由的目标不可用'
            },
          },
          {
            label: '已传播',
            field: 'propagated',
            render({ cell }: { cell: string }) {
              return cell ? '是' : '否'
            },
          },
          {
            label: '备注',
            field: 'memo',
          },
        ]);
        break;
      case 'gcp':
        fileds.value.push(...[
          {
            name: '所属网络',
            prop: 'vpc_id',
            render (val: string) {
              return `VPC-${val}`
            }
          },
          {
            name: '创建时间',
            prop: 'created_at',
          },
        ]);
        columns.value.push(...[
          {
            label: '名称',
            field: 'name',
          },
          {
            label: '目的地址',
            field: 'dest_range',
          },
          {
            label: '下一跳地址',
            field: 'next_hop_gateway',
          },
          {
            label: '优先级',
            field: 'priority',
          },
          {
            label: '实例标记',
            field: 'tags',
          },
          {
            label: '备注',
            field: 'memo',
          },
        ]);
        break;
      case 'huawei':
        fileds.value.push(...[
          {
            name: '所属网络',
            prop: 'vpc_id',
            render (val: string) {
              return `VPC-${val}`
            }
          },
          {
            name: '地域',
            prop: '',
          },
          {
            name: '地域ID',
            prop: 'region',
          },
          {
            name: '路由表类型',
            value: '默认路由表'
          },
          {
            name: '备注',
            prop: 'memo',
          },
          {
            name: '创建时间',
            prop: 'created_at',
          },
        ]);
        columns.value.push(...[
          {
            label: '目的地址',
            field: 'destination',
          },
          {
            label: '下一跳类型',
            field: '',
          },
          {
            label: '下一跳地址',
            field: 'nexthop',
          },
          {
            label: '类型',
            field: 'type',
          },
          {
            label: '备注',
            field: 'memo',
          },
        ]);
        break;
    }
    handleGetData()
  },
  {
    deep: true,
    immediate: true
  }
)

const handleGetData = () => {
  if (props.detail.id) {
    isLoading.value = true
    const filter = {
      op: 'and',
      rules: [{
        field: 'route_table_id',
        op: 'eq',
        value: props.detail.id
      }]
    }
    Promise.all([
      resourceStore
        .getRouteList(
          props.detail.vendor,
          props.detail.id,
          {
            filter,
            page: {
              count: false,
              start: (pagination.value.current - 1) * pagination.value.limit,
              limit: pagination.value.limit,
            },
          }
        ),
      resourceStore
        .getRouteList(
          props.detail.vendor,
          props.detail.id,
          {
            filter,
            page: {
              count: true,
            },
          }
        )
    ])
    .then(([listResult, countResult]) => {
      datas.value = listResult?.data?.details || []
      pagination.value.count = countResult?.data?.count || 0;
    })
    .finally(() => {
      isLoading.value = false
    })
  }
}

// 页码变化发生的事件
const handlePageChange = (current: number) => {
  pagination.value.current = current;
  handleGetData();
};

// 条数变化发生的事件
const handlePageSizeChange = (limit: number) => {
  pagination.value.limit = limit;
  handleGetData();
};
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <detail-info
      class="mt20"
      :fields="fileds"
      :detail="detail"
    ></detail-info>
    <bk-table
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
    />
  </bk-loading>
</template>
