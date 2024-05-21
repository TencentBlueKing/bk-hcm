<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import ImageInfo from '../components/image/image-info.vue';

import { ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const route = useRoute();
const { t } = useI18n();
const { whereAmI } = useWhereAmI();

const imageId = ref<string>(route.query?.id as string);
const vendor = ref<string>(route.query?.type as string);

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
</script>

<template>
  <detail-header>{{ t('镜像') }}：ID（{{ imageId }}）</detail-header>

  <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
    <detail-tab :tabs="hostTabs">
      <template #default>
        <image-info :id="imageId" :vendor="vendor"></image-info>
      </template>
    </detail-tab>
  </div>
</template>
