<script lang="ts" setup>
import InfoList from '../../../common/info-list/info-list';

const settingInfo = [
  {
    name: '操作系统',
    value: 'win',
  },
  {
    name: 'CPU',
    value: '100',
  },
  {
    name: '内存',
    value: '100',
  },
  {
    name: '镜像 ID',
    value: '100',
  },
  {
    name: '镜像名称',
    value: 'mirror.tencent',
  },
  {
    name: '镜像类型',
    value: 'image',
  },
];
</script>

<template>
  <info-list class="mt20" :fields="settingInfo"></info-list>
</template>
