<script lang="ts" setup>
import {
  h,
} from 'vue';
import {
  Button,
} from 'bkui-vue';

import useMountedDrive from '../../../hooks/use-mounted-drive';
import useUninstallDrive from '../../../hooks/use-uninstall-drive';

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

const columns = [
  {
    label: 'ID',
    field: 'id',
  },
  {
    label: '名称',
    field: 'id',
  },
  {
    label: '属性',
    field: 'id',
  },
  {
    label: '类型',
    field: 'id',
  },
  {
    label: '容量（GB）',
    field: 'id',
  },
  {
    label: '创建时间',
    field: 'id',
  },
  {
    label: '计费模式',
    field: 'id',
  },
  {
    label: '过期时间',
    field: 'id',
  },
  {
    label: '运行状态',
    field: 'id',
  },
  {
    label: '操作',
    render() {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            handleUninstallDrive();
          },
        },
        [
          '卸载',
        ],
      );
    },
  },
];
const tableData = [
  {
    id: 233,
  },
];
</script>

<template>
  <bk-button
    class="mt20"
    theme="primary"
    @click="handleMountedDrive"
  >挂载</bk-button>

  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />

  <mounted-drive
    v-model:is-show="isShowMountedDrive"
  />

  <uninstall-drive
    v-model:is-show="isShowUninstallDrive"
  />
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin: 20px 0 5px;
  }
</style>
