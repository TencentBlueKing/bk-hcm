<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import ImageInfo from '../components/image/image-info.vue';

import {
  useI18n,
} from 'vue-i18n';

const {
  t,
} = useI18n();

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
</script>

<template>
  <detail-header>
    {{ t('镜像') }}：ID（xxx）
  </detail-header>

  <detail-tab
    :tabs="hostTabs"
  >
    <template #default>
      <image-info></image-info>
    </template>
  </detail-tab>
</template>
