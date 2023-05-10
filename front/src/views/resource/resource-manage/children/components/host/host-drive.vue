<script lang="ts" setup>
import {
  h,
  watch,
  ref,
  inject,
  computed,
} from 'vue';
import {
  Button,
} from 'bkui-vue';

import {
  useRouter,
  useRoute,
} from 'vue-router';
import useMountedDrive from '../../../hooks/use-mounted-drive';
import useUninstallDrive from '../../../hooks/use-uninstall-drive';
import useQueryList from '../../../hooks/use-query-list';
import bus from '@/common/bus';
import {
  useResourceStore,
} from '@/store/resource';
import {
  useAccountStore,
} from '@/store/account';

const props = defineProps({
  data: {
    type: Object,
  },
  isBindBusiness: {
    type: [Boolean, String],
  },
});

const resourceStore = useResourceStore();
const accountStore = useAccountStore();
const router = useRouter();
const route = useRoute();


const isResourcePage: any = inject('isResourcePage');
const authVerifyData: any = inject('authVerifyData');


const actionName = computed(() => {   // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});


// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

const {
  datas,
  triggerApi,
  isLoading,
} = useQueryList(
  {},
  'disk',
  () => {
    return Promise.all([resourceStore.getDiskListByCvmId(props.data.vendor, props.data.id)]);
  },
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
            const type = 'drive';
            const routeInfo: any = {
              query: {
                id: cell,
                type: props.data.vendor,
              },
            };
            // 业务下
            if (route.path.includes('business')) {
              routeInfo.query.bizs = accountStore.bizs;
              Object.assign(
                routeInfo,
                {
                  name: `${type}BusinessDetail`,
                },
              );
            } else {
              Object.assign(
                routeInfo,
                {
                  name: 'resourceDetail',
                  params: {
                    type,
                  },
                },
              );
            }
            router.push(routeInfo);
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
          data.exencrypted ? '是' : '否',
        ],
      );
    },
  },
  {
    label: '随主机销毁',
    field: '',
    render({ data }: any) {
      const attachment = data?.extension?.attachment;
      const host = attachment?.find((x: any) => x.instance_id === props.data.cloud_id);
      return host ? (host.delete_on_termination ? '是' : '否') : '--';
    },
  },
  {
    label: '操作',
    render({ data }: any) {
      return h(
        'span',
        {
          onClick() {
            showAuthDialog(actionName.value);
          },
        },
        [
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              disabled: data.is_system_disk || !authVerifyData.value?.permissionAction[actionName.value],
              onClick() {
                handleUninstallDrive(data);
              },
            },
            [
              '卸载',
            ],
          )],
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
        },
      ]);
    }
    if (props.data.vendor === 'aws') {
      columns.value.splice(2, 0, ...[
        {
          label: '硬盘类型',
          field: 'disk_type',
        },
        {
          label: '设备名',
          field: 'device_name',
          render({ data }: any) {
            const attachment = data?.extension?.attachment;
            const host = attachment.find((x: any) => x.instance_id === props.data.cloud_id);
            return host.device_name;
          },
        },
        {
          label: '容量(GB)',
          field: 'disk_size',
        },
      ]);
    }
    if (props.data.vendor === 'azure') {
      columns.value.splice(6, 1);
    }
  },
  {
    deep: true,
    immediate: true,
  },
);
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <span @click="showAuthDialog(actionName)">
      <bk-button
        class="mt20 mr20 w100"
        theme="primary"
        :disabled="isBindBusiness || !authVerifyData?.permissionAction[actionName]"
        @click="handleMountedDrive"
      >挂载</bk-button>
    </span>
    <bk-table
      class="mt20"
      row-hover="auto"
      :columns="columns"
      :data="datas"
      show-overflow-tooltip
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
