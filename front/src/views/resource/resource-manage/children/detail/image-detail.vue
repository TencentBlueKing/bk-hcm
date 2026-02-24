<script lang="ts" setup>
import DetailTab from '../../common/tab/detail-tab';
import ImageInfo from '../components/image/image-info.vue';

import { ref, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useBreadcrumb from '@/hooks/use-breadcrumb';

const route = useRoute();
const { t } = useI18n();
const { whereAmI } = useWhereAmI();
const { setTitle } = useBreadcrumb();

const imageId = ref<string>(route.params.id as string);
const vendor = ref<string>(route.query?.type as string);

watchEffect(() => {
  setTitle(`${t('镜像')}：ID（${imageId.value}）`);
});

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];
</script>

<template>
  <div class="detail-content-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
    <detail-tab :tabs="hostTabs">
      <template #default>
        <image-info :id="imageId" :vendor="vendor"></image-info>
      </template>
    </detail-tab>
  </div>
</template>
