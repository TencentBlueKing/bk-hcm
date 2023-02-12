<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import ImageInfo from '../components/image/image-info';

import { ref } from 'vue';
import {
  useRoute,
} from 'vue-router';
import {
  useI18n,
} from 'vue-i18n';

const route = useRoute();
const {
  t,
} = useI18n();

const imageId = ref(route.query?.id);
const vendor = ref(route.query?.vendor);

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
</script>

<template>
  <detail-header>
    {{ t('镜像') }}：ID（{{imageId}}）
  </detail-header>

  <detail-tab
    :tabs="hostTabs"
  >
    <template #default>
      <image-info :id="imageId" :vendor="vendor"></image-info>
    </template>
  </detail-tab>
</template>
