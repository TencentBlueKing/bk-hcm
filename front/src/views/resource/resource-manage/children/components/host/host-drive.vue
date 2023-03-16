<script lang="ts" setup>
import {
  h,
  watch,
  ref,
} from 'vue';
import {
  Button,
} from 'bkui-vue';

import {
  useRouter
} from 'vue-router';
import useMountedDrive from '../../../hooks/use-mounted-drive';
import useUninstallDrive from '../../../hooks/use-uninstall-drive';
import useQueryList from '../../../hooks/use-query-list'
import {
  useResourceStore,
} from '@/store/resource';

const props = defineProps({
  data: {
    type: Object,
  },
});

const resourceStore = useResourceStore();
const router = useRouter()

const {
  datas,
  triggerApi,
  isLoading,
} = useQueryList(
  {},
  'disk',
  () => {
    return Promise.all([resourceStore.getDiskListByCvmId(props.data.vendor, props.data.id)])
  }
);

const {
  isShowMountedDrive,
  handleMountedDrive,
  MountedDrive,
} = useMountedDrive();

const {
  isShowUninstallDrive,
  handleUninstallDrive,
  UninstallDrive,
} = useUninstallDrive();

const columns = ref([
  {
    label: '硬盘用途',
    field: '',
  },
  {
    label: '名称',
    field: 'name',
  },
  {
    label: 'ID',
    field: 'id',
    sort: true,
    render({ cell }: { cell: string }) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: { type: 'drive' },
              query: {
                id: cell,
              },
            });
          },
        },
        [
          cell || '--',
        ],
      );
    },
  },
  {
    label: '连接状态',
    field: '',
  },
  {
    label: '容量',
    field: 'disk_size',
  },
  {
    label: '已加密',
    field: '',
  },
  {
    label: '删除实例时',
    field: '',
  },
  {
    label: '操作',
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            handleUninstallDrive(data);
          },
        },
        [
          '卸载',
        ],
      );
    },
  },
]);

watch(
  () => props.data,
  () => {
    if (props.data.vendor === 'tcloud') {
      columns.value.splice(2, 4 , ...[
        {
          label: '硬盘类型',
          field: 'disk_type',
        },
        {
          label: '容量',
          field: 'disk_size',
        },
        {
          label: '计费类型',
          field: 'disk_charge_type',
        },
        {
          label: '到期时间',
          field: '',
        }
      ])
    }
    if (props.data.vendor === 'aws') {
      columns.value.splice(2, 4 , ...[
        {
          label: '硬盘类型',
          field: 'disk_type',
        },
        {
          label: '接口类型',
          field: '',
        },
        {
          label: '容量',
          field: 'disk_size',
        },
        {
          label: '加密类型',
          field: '',
        },
        {
          label: '模式',
          field: '',
        },
        {
          label: '删除实例时',
          field: '',
        }
      ])
    }
    if (props.data.vendor === 'azure') {
      columns.value.splice(5, 1 , ...[
        {
          label: '终止时删除',
          field: '',
        },
      ])
    }
  },
  {
    deep: true,
    immediate: true
  }
)
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <bk-button
      class="mt20 mr20 w100"
      theme="primary"
      @click="handleMountedDrive"
    >挂载</bk-button>

    <bk-table
      class="mt20"
      row-hover="auto"
      :columns="columns"
      :data="datas"
    />
  </bk-loading>

  <mounted-drive
    v-model:is-show="isShowMountedDrive"
    :detail="data"
    @success="triggerApi"
  />

  <uninstall-drive
    v-model:is-show="isShowUninstallDrive"
    @success="triggerApi"
  />
</template>

<style lang="scss" scoped>
  .w100 {
    width: 100px;
  }
</style>
