<script lang="ts" setup>
import { CloudType } from '@/typings/account';

import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import DetailInfo from '../../common/info/detail-info';
// import { useAccountStore } from '@/store';

import { ref, computed, h, withDirectives, inject } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { InfoBox, bkTooltips } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useResourceStore } from '@/store/resource';
import useDetail from '../../hooks/use-detail';
import useMountedDrive from '../../hooks/use-choose-host-drive';
import useUninstallDrive from '../../hooks/use-uninstall-drive';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { timeFormatter } from '@/common/util';

const { getRegionName } = useRegionsStore();
const { getNameFromBusinessMap } = useBusinessMapStore();
const isResourcePage: any = inject('isResourcePage');
const { whereAmI } = useWhereAmI();

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
    name: '资源ID',
    prop: 'cloud_id',
    render(cell = '') {
      const index = cell.lastIndexOf('/') <= 0 ? 0 : cell.lastIndexOf('/') + 1;
      const value = cell.slice(index);
      return withDirectives(h('span', [value || '--']), [[bkTooltips, cell]]);
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
      return `/#/resource/account/detail/?accountId=${val}&id=${val}`;
    },
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render: (val: number) => (val === -1 ? '未分配' : `${getNameFromBusinessMap(val)} (${val})`),
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
    },
  },
  {
    name: '可用区',
    prop: 'zone',
  },
  {
    name: '硬盘分类',
    prop: 'is_system_disk',
    render(cell: boolean) {
      return cell ? '系统盘' : '数据盘';
    },
  },
  {
    name: '硬盘类型',
    prop: 'disk_type',
  },
  {
    name: '硬盘容量(GB)',
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
      const type = 'host';
      const routeInfo: any = {
        query: {
          id,
          type: detail.value.vendor,
        },
      };
      // 业务下
      if (route.path.includes('business')) {
        Object.assign(routeInfo, {
          name: `${type}BusinessDetail`,
        });
      } else {
        Object.assign(routeInfo, {
          name: 'resourceDetail',
          params: {
            type,
          },
        });
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
    render: (cell: string) => timeFormatter(cell),
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
const router = useRouter();
// const accountStore = useAccountStore();

// const isResourcePage = computed(() => {   // 资源下没有业务ID
//   return !accountStore.bizs;
// });

const { t } = useI18n();

const { isShowMountedDrive, handleMountedDrive, MountedDrive } = useMountedDrive();

const { isShowUninstallDrive, handleUninstallDrive, UninstallDrive } = useUninstallDrive();

const { loading, detail, getDetail } = useDetail('disks', route.query.id as string, (detail: any) => {
  switch (detail.vendor) {
    case 'tcloud':
      settingFields.value.push(
        ...[
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
            prop: 'extension.charge_type',
            render() {
              return detail.extension.charge_type === 'PREPAID' ? '包年/包月' : '按量计费' || '--';
            },
          },
          {
            name: '到期时间',
            prop: 'extension.expire_time',
            render() {
              return timeFormatter(detail.extension.expire_time) || '--';
            },
          },
        ],
      );
      break;
    case 'azure':
      settingFields.value.splice(
        9,
        1,
        ...[
          {
            name: '资源组',
            prop: 'resource_group_name',
          },
          {
            name: '硬盘类型',
            prop: 'sku_name',
          },
        ],
      );
      break;
    case 'huawei':
      settingFields.value.push(
        ...[
          {
            name: '硬盘用途',
            prop: '',
            render() {
              return detail.is_system_disk ? '系统盘' : '数据盘';
            },
          },
          {
            name: '计费模式',
            prop: 'extension.charge_type',
            render() {
              return detail.extension.charge_type === 'prePaid' ? '包年/包月' : '按量计费' || '--';
            },
          },
          {
            name: '到期时间',
            prop: 'extension.expire_time',
            render() {
              return timeFormatter(detail.extension.expire_time) || '--';
            },
          },
        ],
      );
      break;
  }
});

const handleShowDelete = () => {
  InfoBox({
    title: '请确认是否回收',
    subTitle: `将回收【${detail.value.cloud_id}${detail.value.name ? ` - ${detail.value.name}` : ''}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    extCls: 'recycle-resource-infobox',
    onConfirm() {
      return resourceStore
        .recycled('disks', {
          infos: [{ id: detail.value.id }],
        })
        .then(() => {
          router.replace({
            path: location.href.includes('business') ? 'recyclebin/disk' : '/resource/resource/recycle',
            query: {
              type: 'disk',
            },
          });
        });
    },
  });
};

const disabledOption = computed(() => {
  // 无权限，直接禁用按钮
  // if (!authVerifyData.value?.permissionAction?.[actionName.value]) return true;
  // 业务下，判断是否已被回收
  if (!isResourcePage.value) return detail.value?.recycle_status === 'recycling';
  // 资源下，判断是否分配业务，是否已被回收
  return detail.value?.bk_biz_id !== -1 || detail.value?.recycle_status === 'recycling';
});
const bkTooltipsOptions = computed(() => {
  // 无权限
  // if (!authVerifyData.value?.permissionAction?.[actionName.value]) return {
  //     content: '当前用户无权限操作该按钮',
  //     disabled: authVerifyData.value.permissionAction[actionName.value],
  // }
  // 资源下，是否分配业务
  if (isResourcePage.value && detail.value?.bk_biz_id !== -1)
    return {
      content: '该硬盘仅可在业务下操作',
      disabled: detail.value.bk_biz_id === -1,
    };
  // 业务/资源下，是否已被回收
  if (detail.value?.recycle_status === 'recycling')
    return {
      content: '已回收的资源，不支持操作',
      disabled: detail.value.recycle_status !== 'recycling',
    };

  return null;
});

// const disableOperation = computed(() => {
//   return !location.href.includes('business') && detail.value.bk_biz_id !== -1;
// });
// const disableToolTips = computed(() => {
//   return {
//     content: '已回收的资源，不支持操作',
//     disabled: !disableOperation.value && detail.recycle_status !== 'recycling',
//   };
// });
</script>

<template>
  <bk-loading :loading="loading">
    <detail-header>
      云硬盘：ID（{{ detail.id }}）
      <template #right>
        <bk-button
          v-if="!detail.instance_id"
          v-bk-tooltips="bkTooltipsOptions || { disabled: true }"
          class="w100 ml10"
          theme="primary"
          :disabled="disabledOption"
          @click="handleMountedDrive"
        >
          {{ t('挂载') }}
        </bk-button>
        <bk-button
          v-else
          class="w100 ml10"
          theme="primary"
          v-bk-tooltips="
            bkTooltipsOptions ||
            (detail.is_system_disk
              ? {
                  content: '该硬盘是系统盘，不允许卸载',
                  disabled: !detail.is_system_disk,
                }
              : { disabled: true })
          "
          :disabled="disabledOption || detail.is_system_disk"
          @click="handleUninstallDrive(detail)"
        >
          {{ t('卸载') }}
        </bk-button>
        <bk-button
          v-bk-tooltips="
            bkTooltipsOptions || {
              content: '该硬盘已绑定主机，不可单独回收',
              disabled: !detail.instance_id,
            }
          "
          class="w100 ml10"
          theme="primary"
          :disabled="!!detail.instance_id || disabledOption"
          @click="handleShowDelete"
        >
          {{ t('回收') }}
        </bk-button>
      </template>
    </detail-header>

    <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
      <detail-tab :tabs="hostTabs">
        <template #default>
          <detail-info :fields="settingFields" :detail="detail" />
        </template>
      </detail-tab>
    </div>
  </bk-loading>

  <mounted-drive v-if="detail.id" v-model:is-show="isShowMountedDrive" :detail="detail" @success-attach="getDetail" />

  <uninstall-drive v-model:is-show="isShowUninstallDrive" @success="getDetail" />
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}

.w60 {
  width: 60px;
}

.f-right {
  float: right;
}
</style>
