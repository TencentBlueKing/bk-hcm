<script setup lang="ts">
import { computed, inject, onMounted, Ref, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import routerAction from '@/router/utils/action';
import { ActiveQueryKey } from '../constants';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { type ILoadBalancerDetails, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { getInstVip } from '@/utils';
import { useBack } from '@/router/hooks/use-back';
import { MENU_BUSINESS_LOAD_BALANCER_OVERVIEW } from '@/constants/menu-symbol';
import { ModelPropertyDisplay } from '@/model/typings';
import { DisplayFieldFactory, DisplayFieldType } from '../children/display/field-factory';

import GridDetails from '../children/display/grid-details.vue';
import ListenerTable from '../listener/listener-table.vue';
import LoadBalancerInfo from './load-balancer-info.vue';
import SecurityGroup from '../clb/security-group/index.vue';

const emit = defineEmits<{ 'details-show': [id: string] }>();

const route = useRoute();
const { t } = useI18n();
const loadBalancerClbStore = useLoadBalancerClbStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');
watch(currentGlobalBusinessId, (val) => {
  routerAction.redirect({ name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW, query: { [GLOBAL_BIZS_KEY]: val } });
});

const { handleBack } = useBack();
const hasBackAll = computed(() => Object.hasOwn(route.query, '_f'));

const details = ref<ILoadBalancerDetails>();

const displayFieldProperties = DisplayFieldFactory.createModel(DisplayFieldType.CLB).getProperties();
const fieldIds = ['name', 'cloud_id', 'lb_vip', 'region'];
const fieldConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  lb_vip: {
    render: (data: ILoadBalancerDetails) => getInstVip(data),
  },
};
const fields = fieldIds.map((id) => {
  const property = displayFieldProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...fieldConfig[id] };
});

watch(
  () => route.params.id,
  async (id) => {
    details.value = await loadBalancerClbStore.getLoadBalancerDetails(id as string, currentGlobalBusinessId.value);
  },
  { immediate: true },
);

const tabs = computed(() => {
  return [
    { label: t('监听器'), name: 'listener', component: ListenerTable },
    { label: t('基本信息'), name: 'info', component: LoadBalancerInfo },
    { label: t('安全组'), name: 'security', component: SecurityGroup },
  ];
});
const active = ref(route.query?.[ActiveQueryKey.DETAILS] || tabs.value[0].name);

const handleTabChange = (tabName: string) => {
  routerAction.redirect({ query: { ...route.query, [ActiveQueryKey.DETAILS]: tabName } });
};

onMounted(() => {
  emit('details-show', route.params.id as string);
});
</script>

<template>
  <div class="details-container">
    <section
      v-if="hasBackAll"
      class="back"
      @click="handleBack({ query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId } })"
    >
      <i class="hcm-icon bkhcm-icon-arrows--left-line"></i>
      <span>{{ t('返回全部') }}</span>
    </section>
    <grid-details
      class="overview"
      layout="vertical"
      :content-min-width="200"
      :column="4"
      :fields="fields"
      :details="details"
      :is-loading="loadBalancerClbStore.loadBalancerDetailsLoading"
    />
    <bk-tab
      class="tab-container"
      :class="{ 'has-back-all': hasBackAll }"
      v-model:active="active"
      type="card-grid"
      @change="handleTabChange"
    >
      <bk-tab-panel v-for="tab in tabs" :key="tab.name" :label="tab.label" :name="tab.name" render-directive="if">
        <component :is="tab.component" :lb-id="route.params.id" :details="details"></component>
      </bk-tab-panel>
    </bk-tab>
  </div>
</template>

<style scoped lang="scss">
.details-container {
  height: 100%;
  padding: 24px;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;

  .back {
    margin-bottom: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 120px;
    height: 32px;
    border-radius: 16px;
    background: #fff;
    box-shadow: 0 2px 4px 0 #1919290d;
    cursor: pointer;

    .hcm-icon {
      margin-right: 8px;
      color: #3a84ff;
      font-weight: 700;
    }
  }

  .overview {
    margin-bottom: 24px;

    :deep(.grid-item) {
      .item-label {
        padding: 0 !important;
        color: #979ba5;
      }

      .item-content {
        padding: 4px 0 0 !important;
        color: #313238;
      }
    }
  }

  .tab-container {
    flex: 1;
    max-height: calc(100% - 77px);

    &.has-back-all {
      max-height: calc(100% - 133px);
    }

    :deep(.bk-tab-content) {
      height: calc(100% - 40px);
      padding: 0;

      .bk-tab-panel {
        height: 100%;
      }
    }
  }
}
</style>
