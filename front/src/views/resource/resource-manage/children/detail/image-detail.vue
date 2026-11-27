<script lang="ts" setup>
import DetailTab from '../../common/tab/detail-tab';
import ImageInfo from '../components/image/image-info.vue';

import { ref, computed, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { useResourceDetailRoute } from '../../hooks/use-resource-detail-route';

const route = useRoute();
const { t } = useI18n();
const { setTitle } = useBreadcrumb();
const { detailId } = useResourceDetailRoute();

const imageId = ref<string>(detailId.value);
const vendor = computed(() => route.query?.vendor as string);

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
  <div class="detail-content-wrap">
    <detail-tab :tabs="hostTabs">
      <template #default>
        <image-info :id="imageId" :vendor="vendor"></image-info>
      </template>
    </detail-tab>
  </div>
</template>
