<script lang="ts" setup>
import { h, watch, ref, inject, computed, withDirectives, resolveComponent } from 'vue';
import { bkTooltips, Button } from 'bkui-vue';

import useMountedDrive from '../../../hooks/use-mounted-drive';
import useUninstallDrive from '../../../hooks/use-uninstall-drive';
import useQueryList from '../../../hooks/use-query-list';
import { useResourceStore } from '@/store/resource';
import { INSTANCE_CHARGE_MAP, VendorEnum } from '@/common/constant';
import { timeFormatter } from '@/common/util';
import { MENU_BUSINESS_DRIVE_DETAILS, MENU_RESOURCE_DETAIL } from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { AUTH_UPDATE_IAAS_RESOURCE, AUTH_BIZ_UPDATE_IAAS_RESOURCE } from '@/constants/auth-symbols';

const props = defineProps({
  data: {
    type: Object,
  },
  isBindBusiness: {
    type: [Boolean, String],
  },
});

const resourceStore = useResourceStore();
const { whereAmI, getBizsId } = useWhereAmI();

const isResourcePage: any = inject('isResourcePage');
const bizId = computed(() => getBizsId());
const updateSign = computed(() => {
  if (bizId.value) return { type: AUTH_BIZ_UPDATE_IAAS_RESOURCE, relation: [bizId.value] };
  return { type: AUTH_UPDATE_IAAS_RESOURCE, relation: [props.data?.account_id] };
});
const HcmAuthComp = resolveComponent('hcm-auth');

const { datas, triggerApi, isLoading } = useQueryList({}, 'disk', () => {
  return Promise.all([resourceStore.getDiskListByCvmId(props.data.vendor, props.data.id)]);
});

const { isShowMountedDrive, handleMountedDrive, MountedDrive } = useMountedDrive();

const { isShowUninstallDrive, handleUninstallDrive, UninstallDrive } = useUninstallDrive();

const generateTooltipsOptions = (data: any) => {
  if (isResourcePage.value && props.data?.bk_biz_id !== -1)
    return {
      content: '该主机已分配到业务，仅可在业务下操作',
      disabled: props.data.bk_biz_id === -1,
    };
  if (data?.is_system_disk)
    return {
      content: '系统盘不可以卸载',
      disabled: !data.is_system_disk,
    };

  return {
    disabled: true,
  };
};

const columns = ref([
  {
    label: '硬盘用途',
    field: '',
    render({ data }: any) {
      return data.is_system_disk ? '系统盘' : '数据盘';
    },
  },
  {
    label: '硬盘名称',
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
            const routeInfo: any = {
              query: {
                type: props.data.vendor,
              },
            };
            if (whereAmI.value === Senarios.business) {
              Object.assign(routeInfo, {
                name: MENU_BUSINESS_DRIVE_DETAILS,
                params: { id: cell },
              });
            } else {
              Object.assign(routeInfo, {
                name: MENU_RESOURCE_DETAIL,
                params: { resourceType: 'drive', id: cell },
              });
            }
            routerAction.redirect(routeInfo, { history: true });
          },
        },
        [cell || '--'],
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
      return h('span', [data.exencrypted ? '是' : '否']);
    },
  },
  {
    label: '随主机销毁',
    field: '',
    render({ data }: any) {
      const attachment = data?.extension?.attachment;
      const host = attachment?.find((x: any) => x.instance_id === props.data.cloud_id);
      // eslint-disable-next-line no-nested-ternary
      return host ? (host.delete_on_termination ? '是' : '否') : '--';
    },
  },
  {
    label: '操作',
    render({ data }: any) {
      return h(
        HcmAuthComp,
        { sign: updateSign.value, tag: 'span' },
        {
          default: ({ noPerm }: { noPerm: boolean }) => [
            withDirectives(
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  disabled: noPerm || data.is_system_disk || (isResourcePage.value && props.data?.bk_biz_id !== -1),
                  onClick() {
                    handleUninstallDrive(data);
                  },
                },
                ['卸载'],
              ),
              [[bkTooltips, generateTooltipsOptions(data)]],
            ),
          ],
        },
      );
    },
  },
]);

const sysDiskTypeValues = {
  [VendorEnum.TCLOUD]: {
    CLOUD_PREMIUM: '高性能云硬盘',
    CLOUD_BSSD: '通用型SSD云硬盘',
    CLOUD_SSD: 'SSD云硬盘',
  },
  [VendorEnum.AWS]: {
    gp3: '通用型SSD卷(gp3)',
    gp2: '通用型SSD卷(gp2)',
    io1: '预置IOPS SSD卷(io1)',
    io2: '预置IOPS SSD卷(io2)',
    st1: '吞吐量优化型HDD卷(st1)',
    sc1: 'Cold HDD卷(sc1)',
    standard: '上一代磁介质卷(standard)',
  },
};

const dataDiskTypeValues = {
  [VendorEnum.TCLOUD]: {
    CLOUD_PREMIUM: '高性能云硬盘',
    CLOUD_BSSD: '通用型SSD云硬盘',
    CLOUD_SSD: 'SSD云硬盘',
    CLOUD_HSSD: '增强型SSD云硬盘',
  },
  [VendorEnum.AWS]: {
    gp3: '通用型SSD卷(gp3)',
    gp2: '通用型SSD卷(gp2)',
    io1: '预置IOPS SSD卷(io1)',
    io2: '预置IOPS SSD卷(io2)',
    st1: '吞吐量优化型HDD卷(st1)',
    sc1: 'Cold HDD卷(sc1)',
    standard: '上一代磁介质卷(standard)',
  },
};

watch(
  () => props.data,
  () => {
    if (props.data.vendor === 'tcloud') {
      columns.value.splice(
        2,
        4,
        ...[
          {
            label: '硬盘类型',
            field: 'disk_type',
            render({ data }: any) {
              if (data?.is_system_disk) {
                return sysDiskTypeValues[VendorEnum.TCLOUD][data?.disk_type];
              }
              return dataDiskTypeValues[VendorEnum.TCLOUD][data?.disk_type];
            },
          },
          {
            label: '容量(GB)',
            field: 'disk_size',
          },
          {
            label: '计费类型',
            field: 'disk_charge_type',
            render({ cell }: any) {
              return INSTANCE_CHARGE_MAP[cell];
            },
          },
          {
            label: '到期时间',
            field: '',
            render({ cell }: any) {
              return timeFormatter(cell) || '--';
            },
          },
        ],
      );
    }
    if (props.data.vendor === 'aws') {
      columns.value.splice(
        2,
        0,
        ...[
          {
            label: '硬盘类型',
            field: 'disk_type',
            render({ data }: any) {
              if (data?.is_system_disk) {
                return sysDiskTypeValues[VendorEnum.AWS][data?.disk_type];
              }
              return dataDiskTypeValues[VendorEnum.AWS][data?.disk_type];
            },
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
        ],
      );
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
  <bk-loading :loading="isLoading">
    <hcm-auth :sign="updateSign" tag="span" v-slot="{ noPerm }">
      <bk-button class="btn" theme="primary" :disabled="isBindBusiness || noPerm" @click="handleMountedDrive">
        挂载
      </bk-button>
    </hcm-auth>
    <bk-table class="mt16" row-hover="auto" :columns="columns" :data="datas" show-overflow-tooltip />
  </bk-loading>

  <mounted-drive v-model:is-show="isShowMountedDrive" :detail="data" @success="triggerApi" />

  <uninstall-drive v-model:is-show="isShowUninstallDrive" @success="triggerApi" />
</template>

<style lang="scss" scoped>
.btn {
  min-width: 88px;
}
</style>
