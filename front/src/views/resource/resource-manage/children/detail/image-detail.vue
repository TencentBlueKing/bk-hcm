<script lang="ts" setup>
import DetailTab from '../../common/tab/detail-tab';
import ImageInfo from '../components/image/image-info.vue';

import { ref, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import useBreadcrumb from '@/hooks/use-breadcrumb';

const route = useRoute();
const { t } = useI18n();
const { setTitle } = useBreadcrumb();

const imageId = ref<string>(route.params.id as string);
// vendor 是镜像详情 API 的前置依赖，用于构造 /vendors/{vendor}/images/{id} 请求路径
const vendor = ref<string>(route.query?.vendor as string);

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
