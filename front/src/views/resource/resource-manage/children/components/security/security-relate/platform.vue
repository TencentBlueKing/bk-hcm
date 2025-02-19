<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  type ISecurityGroupDetail,
  type ISecurityGroupRelBusiness,
  type ISecurityGroupRelResCountItem,
  type SecurityGroupRelResourceByBizItem,
  useSecurityGroupStore,
} from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import usePage from '@/hooks/use-page';
import { transformSimpleCondition } from '@/utils/search';
import { RELATED_RES_KEY_MAP } from './constants';
import securityGroupRelatedResourcesViewProperties from '@/model/security-group/related-resources.view';
import type { ITab } from './typings';

import { Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import tab from './tab/index.vue';
import bind from './bind/index.vue';
import batchUnbind from './unbind/batch.vue';
import dataList from './data-list/index.vue';
import singleUnbind from './unbind/single.vue';

const props = defineProps<{
  detail: ISecurityGroupDetail;
  relatedResourcesCountList: ISecurityGroupRelResCountItem[];
  relatedBiz: ISecurityGroupRelBusiness;
}>();

const { t } = useI18n();
const { getBizsId, isBusinessPage } = useWhereAmI();
const { getBusinessNames } = useBusinessGlobalStore();
const securityGroupStore = useSecurityGroupStore();

const tabActive = ref<ITab>('CVM');
// 当前业务所关联资源
const currentBizRelatedResources = computed(
  () =>
    props.relatedBiz?.[RELATED_RES_KEY_MAP[tabActive.value]]?.find((item) => item.bk_biz_id === getBizsId()) || {
      res_count: 0,
    },
);

// 关联资源table
const list = ref<SecurityGroupRelResourceByBizItem[]>([]);
const { pagination, getPageParams } = usePage();
const condition = ref<Record<string, any>>({});

// 业务下的平台管理：只拉取当前业务所关联的实例列表；其他业务只展示业务数量。
// 账号下的平台管理：拉取所有业务所关联的实例列表
const loading = ref(false);
const getList = async (sort = 'created_at', order = 'DESC') => {
  try {
    const { id } = props.detail;
    const api =
      tabActive.value === 'CVM' ? securityGroupStore.queryRelCvmByBiz : securityGroupStore.queryRelLoadBalancerByBiz;
    const bizIds = isBusinessPage
      ? [getBizsId()]
      : props.relatedBiz[RELATED_RES_KEY_MAP[tabActive.value]].map(({ bk_biz_id }) => bk_biz_id);

    const res = await Promise.all(
      bizIds.map((bk_biz_id) =>
        api(id, bk_biz_id, {
          filter: transformSimpleCondition(condition.value, securityGroupRelatedResourcesViewProperties),
          page: getPageParams(pagination, { sort, order }),
        }),
      ),
    );

    list.value = res.flatMap((item) => item.list);
    // 设置页码总条数
    pagination.count = isBusinessPage ? res[0].count : res.reduce((acc, cur) => acc + cur.count, 0);
  } finally {
    loading.value = false;
  }
};

const selected = ref<SecurityGroupRelResourceByBizItem[]>([]);

const handleBind = async (ids: string[]) => {
  // TODO：当前只支持CVM
  await securityGroupStore.batchAssociateCvms({ security_group_id: props.detail.id, cvm_ids: ids });
  Message({ theme: 'success', message: t('绑定成功') });
  getList();
};

const handleBatchUnbind = async (ids: string[]) => {
  // TODO：当前只支持CVM
  await securityGroupStore.batchDisassociateCvms({ security_group_id: props.detail.id, cvm_ids: ids });
  Message({ theme: 'success', message: t('解绑成功') });
  getList();
};

onBeforeMount(() => {
  getList();
});
</script>

<template>
  <div class="platform-manage-module">
    <div class="tools-bar">
      <tab v-model="tabActive" :detail="detail" :related-resources-count-list="relatedResourcesCountList" />

      <!-- TODO：目前只支持CVM -->
      <div class="operate-btn-wrap" v-if="tabActive === 'CVM'">
        <bind :tab-active="tabActive" :detail="detail" @confirm="handleBind">
          <template #icon>
            <plus width="26" height="26" />
          </template>
        </bind>
        <batch-unbind
          :selections="selected"
          :disabled="!selected.length"
          :tab-active="tabActive"
          :handle-confirm="handleBatchUnbind"
        />
      </div>

      <bk-search-select class="search" :placeholder="t('请输入IP/主机名称等搜索')" />
    </div>

    <div v-if="isBusinessPage" class="overview">
      {{ t(`当前业务（${getBusinessNames(getBizsId())}）下共有`) }}
      <span class="number">{{ currentBizRelatedResources?.res_count }}</span>
      {{ t(`台${RELATED_RES_KEY_MAP[tabActive]}，还有`) }}
      <span class="number">
        {{
          relatedBiz?.[RELATED_RES_KEY_MAP[tabActive]]?.filter(({ bk_biz_id: bkBizId }) => bkBizId !== getBizsId())
            .length
        }}
      </span>
      {{ t(`个业务也在使用`) }}
    </div>

    <div class="rel-res-display-wrap">
      <data-list
        v-bkloading="{ loading }"
        :list="list"
        :column-key="`${tabActive}-base`"
        :pagination="pagination"
        @select="(selections) => (selected = selections)"
      >
        <template v-if="tabActive === 'CVM'" #operate>
          <bk-table-column :label="'操作'">
            <template #default="{ row }">
              <single-unbind :row="row" :tab-active="tabActive" @confirm="handleBatchUnbind([row.id])" />
            </template>
          </bk-table-column>
        </template>
      </data-list>
    </div>
  </div>
</template>

<style scoped lang="scss">
.tools-bar {
  display: flex;
  align-items: center;

  .operate-btn-wrap {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .search {
    margin-left: auto;
    width: 320px;
  }
}

.overview {
  margin-top: 12px;
  font-size: 12px;
  color: #4d4f56;

  .number {
    font-weight: 700;
  }
}

.rel-res-display-wrap {
  margin-top: 12px;
}
</style>
