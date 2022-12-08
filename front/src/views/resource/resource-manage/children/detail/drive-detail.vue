<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import DriveInfo from '../components/drive/drive-info.vue';

import {
  useI18n,
} from 'vue-i18n';
import useMountedDrive from '../../hooks/use-mounted-drive';
import useUninstallDrive from '../../hooks/use-uninstall-drive';

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


const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
</script>

<template>
  <detail-header>
    云硬盘：ID（xxx）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleMountedDrive"
      >
        {{ t('挂载') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleUninstallDrive"
      >
        {{ t('卸载') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
      >
        {{ t('删除') }}
      </bk-button>
    </template>
  </detail-header>

  <detail-tab
    :tabs="hostTabs"
  >
    <template #default>
      <drive-info></drive-info>
    </template>
  </detail-tab>

  <mounted-drive
    v-model:is-show="isShowMountedDrive"
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
