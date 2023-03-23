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
    render({ data }: any) {
      return data.is_system_disk ? '系统盘' : '数据盘';
    },
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
    label: '状态',
    field: 'status',
  },
  {
    label: '容量(GB)',
    field: 'disk_size',
  },
  {
    label: '是否加密',
    field: 'exencrypted',
    render({ data }: any) {
      return h(
        'span',
        [
          data.exencrypted ? '是' : '否'
        ],
      );
    },
  },
  {
    label: '随主机销毁',
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
          disabled: data.is_system_disk,
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
      columns.value.splice(2, 4, ...[
        {
          label: '硬盘类型',
          field: 'disk_type',
        },
        {
          label: '容量(GB)',
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
      columns.value.splice(2, 1, ...[
        {
          label: '硬盘类型',
          field: 'disk_type',
        },
        {
          label: '设备名',
          field: 'device_name',
        },
        {
          label: '容量(GB)',
          field: 'disk_size',
        },
        {
          label: '是否加密',
          field: 'exencrypted',
        }
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
