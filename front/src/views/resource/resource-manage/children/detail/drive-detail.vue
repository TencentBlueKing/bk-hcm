<script lang="ts" setup>
import { CloudType } from '@/typings/account';

import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import DetailInfo from '../../common/info/detail-info';
import { useAccountStore } from '@/store';

import {
  ref,
  computed,
  h,
  withDirectives,
} from 'vue';
import {
  useRoute,
  useRouter,
} from 'vue-router';
import {
  InfoBox,
  bkTooltips,
} from 'bkui-vue';
import {
  useI18n,
} from 'vue-i18n';
import {
  useResourceStore,
} from '@/store/resource';
import useDetail from '../../hooks/use-detail';
import useMountedDrive from '../../hooks/use-choose-host-drive';
import useUninstallDrive from '../../hooks/use-uninstall-drive';
import { useRegionsStore } from '@/store/useRegionsStore';

const { getRegionName } = useRegionsStore();

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];

const settingFields = ref<any[]>([
  {
    name: 'ID',
    prop: 'id',
  },
  {
    name: '资源 ID',
    prop: 'cloud_id',
    render(cell: string = '') {
      const index = cell.lastIndexOf('/') <= 0 ? 0 : cell.lastIndexOf('/') + 1;
      const value = cell.slice(index);
      return withDirectives(
        h(
          'span',
          [
            value || '--'
          ]
        ), 
        [
          [bkTooltips, cell],
        ]
      )
    },
  },
  {
    name: '资源名称',
    prop: 'name',
  },
  {
    name: '账号',
    prop: 'account_id',
    link(val: string) {
      return `/#/resource/account/detail/?id=${val}`;
    },
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
  },
  {
    name: '状态',
    prop: 'status',
  },
  {
    name: '云厂商',
    prop: 'vendor',
    render(cell: string) {
      return CloudType[cell] || '--';
    },
  },
  {
    name: '地域',
    prop: 'region',
    render(cell: string) {
      return getRegionName(detail.value.vendor, cell);
    }
  },
  {
    name: '可用区',
    prop: 'zone',
  },
  {
    name: '磁盘类型',
    prop: 'disk_type',
  },
  {
    name: '磁盘容量(GB)',
    prop: 'disk_size',
  },
  {
    name: '是否加密',
    prop: 'exencrypted',
  },
  {
    name: '挂载主机',
    prop: 'instance_id',
    txtBtn(id: string) {
      const type = 'host'
      const routeInfo: any = {
        query: {
          id,
          type: detail.value.vendor,
        },
      };
      // 业务下
      if (route.path.includes('business')) {
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
  {
    name: '是否挂载',
    prop: 'instance_id',
    render(cell: string) {
      return cell ? '是' : '否';
    },
  },
  {
    name: '创建时间',
    prop: 'created_at',
  },
  {
    name: '备注',
    type: 'textarea',
    prop: 'memo',
    // edit: true,
  },
]);
const resourceStore = useResourceStore();
const route = useRoute();
const router =  useRouter();
const accountStore = useAccountStore();

const isResourcePage = computed(() => {   // 资源下没有业务ID
  return !accountStore.bizs;
});

const {
  t,
} = useI18n();

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

const {
  loading,
  detail,
  getDetail,
} = useDetail(
  'disks',
  route.query.id as string,
  (detail: any) => {
    switch (detail.vendor) {
      case 'tcloud':
        settingFields.value.push(...[
          {
            name: '是否随实例销毁',
            prop: 'delete_with_instance',
            render(delete_with_instance: string) {
              return delete_with_instance ? '是' : '否';
            },
          },
          {
            name: '硬盘用途',
            prop: '',
            render() {
              return detail.is_system_disk ? '系统盘' : '数据盘';
            },
          },
          {
            name: '到期欠费保护',
            prop: 'backup_disk',
            render(backup_disk: boolean) {
              return backup_disk ? '是' : '否';
            },
          },
          {
            name: '计费模式',
            prop: 'disk_charge_type',
          },
          {
            name: '到期时间',
            prop: 'deadline_time',
          },
        ]);
        break;
      case 'azure':
        settingFields.value.splice(9, 1, ...[
          {
            name: '资源组',
            prop: 'resource_group_name',
          },
          {
            name: '磁盘类型',
            prop: 'sku_name',
          },
        ]);
        break;
      case 'huawei':
        settingFields.value.push(...[
          {
            name: '硬盘用途',
            prop: '',
            render() {
              return detail.is_system_disk ? '系统盘' : '数据盘';
            },
          },
          {
            name: '计费模式',
            prop: 'disk_charge_type',
          },
          {
            name: '到期时间',
            prop: '',
          },
        ]);
        break;
    }
  },
);

const handleShowDelete = () => {
  InfoBox({
    title: '请确认是否回收',
    subTitle: `将回收【${detail.value.name}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    onConfirm() {
      return resourceStore
        .recycled(
          'disks',
          {
            infos: [{ id: detail.value.id }],
          },
        ).then(() => {
          router.replace({
            path: location.href.includes('business') ? 'recyclebin/disk' : '/resource/recyclebin',
          })
        });
    },
  });
};

const disableOperation = computed(() => {
  return !location.href.includes('business') && detail.value.bk_biz_id !== -1
})
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      云硬盘：ID（{{ detail.id }}）
      <template #right>
        <bk-button
          v-if="!detail.instance_id"
          class="w100 ml10"
          theme="primary"
          :disabled="disableOperation"
          @click="handleMountedDrive"
        >
          {{ t('挂载') }}
        </bk-button>
        <bk-button
          v-else
          class="w100 ml10"
          theme="primary"
          :disabled="!!detail.is_system_disk || disableOperation"
          @click="handleUninstallDrive(detail)"
        >
          {{ t('卸载') }}
        </bk-button>
        <bk-button
          class="w100 ml10"
          theme="primary"
          :disabled="!!detail.instance_id || disableOperation"
          @click="handleShowDelete"
        >
          {{ t('回收') }}
        </bk-button>
      </template>
    </detail-header>

    <detail-tab
      :tabs="hostTabs"
    >
      <template #default>
        <detail-info
          :fields="settingFields"
          :detail="detail"
        />
      </template>
    </detail-tab>
  </bk-loading>

  <mounted-drive
    v-if="detail.id"
    v-model:is-show="isShowMountedDrive"
    :detail="detail"
    @success-attach="getDetail"
  />

  <uninstall-drive
    v-model:is-show="isShowUninstallDrive"
    @success="getDetail"
  />
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.f-right{
  float: right;
}
</style>
