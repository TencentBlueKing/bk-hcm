<script lang="ts" setup>
import type { AwsDiskAttachment, Disk, VendorResource } from '@/typings/resource-detail';
import { h, inject, computed, withDirectives, resolveComponent } from 'vue';
import { bkTooltips, Button } from 'bkui-vue';

import { useRoute } from 'vue-router';
import useMountedDrive from '../../../hooks/use-mounted-drive';
import useUninstallDrive from '../../../hooks/use-uninstall-drive';
import useQueryList from '../../../hooks/use-query-list';
import { useResourceStore } from '@/store/resource';
import { INSTANCE_CHARGE_MAP, VendorEnum } from '@/common/constant';
import { formatDiskBoolean, formatDiskExpiry } from '../../../common/disk-format';
import routerAction from '@/router/utils/action';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { MENU_BUSINESS_DRIVE_DETAILS, MENU_RESOURCE_DETAIL } from '@/constants/menu-symbol';
import { useResourceDetailAuth } from '../../../hooks/use-resource-detail-route';

const props = defineProps({
  data: {
    type: Object,
  },
  isBindBusiness: {
    type: [Boolean, String],
  },
});

const resourceStore = useResourceStore();
const route = useRoute();
const { whereAmI } = useWhereAmI();

const isResourcePage: any = inject('isResourcePage');
const { authPending, updateSign } = useResourceDetailAuth(computed(() => props.data?.account_id));
const HcmAuthComp = resolveComponent('hcm-auth');

const { datas, triggerApi, isLoading } = useQueryList({}, 'disks', () => {
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

type DiskTypeLabels = Record<VendorEnum.TCLOUD | VendorEnum.AWS, Record<string, string>>;

const sysDiskTypeValues: DiskTypeLabels = {
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

const dataDiskTypeValues: DiskTypeLabels = {
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

const getHostAttachment = (attachments?: AwsDiskAttachment[]) =>
  attachments?.find((item) => item.instance_id === props.data.cloud_id);

const columns = computed(() => {
  const { vendor } = props.data as { vendor: VendorEnum };

  return [
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
    ...(vendor === VendorEnum.TCLOUD || vendor === VendorEnum.AWS
      ? [
          {
            label: '硬盘类型',
            field: 'disk_type',
            render({ data }: any) {
              const types = data.is_system_disk ? sysDiskTypeValues[vendor] : dataDiskTypeValues[vendor];
              return types[data.disk_type] ?? data.disk_type ?? '--';
            },
          },
        ]
      : []),
    ...(vendor === VendorEnum.AWS
      ? [
          {
            label: '设备名',
            field: 'device_name',
            render: ({ data }: { data: VendorResource<Disk, 'aws'> }) =>
              getHostAttachment(data.extension?.attachment)?.device_name || '--',
          },
        ]
      : []),
    ...(vendor !== VendorEnum.TCLOUD
      ? [
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
                    const routeInfo: any = {};
                    if (whereAmI.value === Senarios.business) {
                      Object.assign(routeInfo, {
                        name: MENU_BUSINESS_DRIVE_DETAILS,
                        params: { bizId: route.params.bizId, id: cell },
                        query: { vendor: props.data.vendor },
                      });
                    } else {
                      // 挂载硬盘与主机厂商相同
                      Object.assign(routeInfo, {
                        name: MENU_RESOURCE_DETAIL,
                        params: { resourceType: 'drive', id: cell },
                        query: { accountId: route.query.accountId, vendor: props.data.vendor },
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
        ]
      : []),
    {
      label: '容量(GB)',
      field: 'disk_size',
    },
    ...(vendor === VendorEnum.TCLOUD
      ? [
          {
            label: '计费类型',
            field: 'disk_charge_type',
            render: ({ data }: { data: VendorResource<Disk, 'tcloud'> }) =>
              INSTANCE_CHARGE_MAP[data.extension?.disk_charge_type] || '--',
          },
          {
            label: '到期时间',
            field: 'deadline_time',
            render: ({ data }: { data: VendorResource<Disk, 'tcloud'> }) =>
              formatDiskExpiry(data.extension?.deadline_time),
          },
        ]
      : [
          {
            label: '是否加密',
            field: 'encrypted',
            render: ({ data }: { data: Disk }) => formatDiskBoolean(data.extension?.encrypted),
          },
        ]),
    ...(vendor !== VendorEnum.AZURE
      ? [
          {
            label: '随主机销毁',
            field: 'delete_with_instance',
            render({ data }: { data: Disk }) {
              if (data.vendor === VendorEnum.TCLOUD) return formatDiskBoolean(data.extension?.delete_with_instance);
              if (data.vendor === VendorEnum.AWS)
                return formatDiskBoolean(getHostAttachment(data.extension?.attachment)?.delete_on_termination);
              return '--';
            },
          },
        ]
      : []),
    {
      label: '操作',
      render({ data }: any) {
        return h(
          HcmAuthComp,
          { sign: updateSign.value, ignore: authPending.value, tag: 'span' },
          {
            default: ({ noPerm }: { noPerm: boolean }) =>
              withDirectives(
                h(
                  Button,
                  {
                    text: true,
                    theme: 'primary',
                    disabled:
                      noPerm ||
                      authPending.value ||
                      data.is_system_disk ||
                      (isResourcePage.value && props.data?.bk_biz_id !== -1),
                    onClick() {
                      handleUninstallDrive({ ...data, instance_id: data.cvm_id });
                    },
                  },
                  ['卸载'],
                ),
                [[bkTooltips, noPerm ? { disabled: true } : generateTooltipsOptions(data)]],
              ),
          },
        );
      },
    },
  ];
});
</script>

<template>
  <bk-loading :loading="isLoading">
    <hcm-auth :sign="updateSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">
      <bk-button
        class="btn"
        theme="primary"
        :disabled="isBindBusiness || noPerm || authPending"
        @click="handleMountedDrive"
      >
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
