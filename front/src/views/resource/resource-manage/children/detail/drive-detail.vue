<script lang="ts" setup>
import { CloudType } from '@/typings/account';

import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import DetailInfo from '../../common/info/detail-info';

import {
  ref,
} from 'vue';
import {
  useRoute,
} from 'vue-router';
import {
  InfoBox,
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

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];

const settingFields = ref<any[]>([
  {
    name: '资源 ID',
    prop: 'id',
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
    name: '磁盘类型',
    prop: 'disk_type',
  },
  {
    name: '磁盘容量',
    prop: 'disk_size',
  },
  {
    name: '是否加密',
    prop: '',
  },
  {
    name: '地域',
    prop: 'region',
  },
  {
    name: '可用区',
    prop: 'zone',
  },
  {
    name: '是否已挂载',
    prop: 'instance_id',
    render (instance_id: string) {
      return instance_id ? '已挂载' : '未挂载'
    }
  },
  {
    name: '挂载主机',
    prop: 'instance_id',
  },
  {
    name: '挂载主机名称',
    prop: '',
  },
  {
    name: '快照',
    prop: '',
  },
  {
    name: '标签',
    prop: '',
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
} = useDetail(
  'disks',
  route.query.id as string,
  (detail: any) => {
    switch (detail.vendor) {
      case 'tcloud':
        settingFields.value.push(...[
          {
            name: '是否随实例销毁',
            prop: '',
          },
          {
            name: '磁盘属性',
            prop: '',
          },
          {
            name: '到期欠费保护',
            prop: '',
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
      case 'azure':
        settingFields.value.push(...[
          {
            name: '资源组',
            prop: 'resource_group_name',
          },
        ]);
        break;
      case 'huawei':
        settingFields.value.push(...[
          {
            name: '磁盘属性',
            prop: '',
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
    title: '请确认是否删除',
    subTitle: `将删除【${detail.value.name}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    onConfirm() {
      return resourceStore
        .deleteBatch(
          'disks',
          {
            ids: [detail.value.id],
          },
        );
    },
  });
};
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
          @click="handleMountedDrive"
        >
          {{ t('挂载') }}
        </bk-button>
        <bk-button
          v-else
          class="w100 ml10"
          theme="primary"
          @click="handleUninstallDrive(detail)"
        >
          {{ t('卸载') }}
        </bk-button>
        <bk-button
          class="w100 ml10"
          theme="primary"
          @click="handleShowDelete"
        >
          {{ t('删除') }}
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
  />

  <uninstall-drive
    v-model:is-show="isShowUninstallDrive"
  />
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
