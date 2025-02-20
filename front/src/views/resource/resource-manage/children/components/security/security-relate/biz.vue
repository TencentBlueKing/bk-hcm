<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  type ISecurityGroupDetail,
  type ISecurityGroupRelBusiness,
  type ISecurityGroupRelResCountItem,
  SecurityGroupRelatedResourceName,
} from '@/store/security-group';
import { RELATED_RES_KEY_MAP } from '@/constants/security-group';

import tab from './tab/index.vue';
import collapseDataList from './data-list/collapse-data-list.vue';

defineProps<{
  detail: ISecurityGroupDetail;
  relatedResourcesCountList: ISecurityGroupRelResCountItem[];
  relatedBiz: ISecurityGroupRelBusiness;
}>();

const { t } = useI18n();

const tabActive = ref<SecurityGroupRelatedResourceName>(SecurityGroupRelatedResourceName.CVM);
</script>

<template>
  <div class="business-manage-module">
    <div class="tools-bar">
      <tab v-model="tabActive" :detail="detail" :related-resources-count-list="relatedResourcesCountList" />
      <bk-search-select class="search" :placeholder="t('请输入IP/主机名称等搜索')" />
    </div>

    <div class="rel-res-display-wrap">
      <collapse-data-list
        v-for="{ bk_biz_id: bkBizId, res_count: resCount } in relatedBiz?.[RELATED_RES_KEY_MAP[tabActive]]"
        :key="bkBizId"
        :detail="detail"
        :bk-biz-id="bkBizId"
        :tab-active="tabActive"
        :res-count="resCount"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.tools-bar {
  display: flex;
  align-items: center;

  .search {
    margin-left: auto;
    width: 320px;
  }
}

.rel-res-display-wrap {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
