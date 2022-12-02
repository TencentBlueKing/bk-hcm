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
const priceInfo = [
  {
    name: '实例计费模式',
    value: 'ddd',
  },
  {
    name: '创建时间',
    value: '2021-11-12',
  },
  {
    name: '网络计费模式',
    value: 'fff',
  },
  {
    name: '到期时间',
    value: '2021-12-12',
  },
];
</script>

<template>
  <h3 class="info-title">配置信息</h3>
  <info-list :fields="settingInfo"></info-list>
  <h3 class="info-title">计费信息</h3>
  <info-list :fields="priceInfo"></info-list>
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin: 20px 0 5px;
  }
</style>
