<script setup lang="ts">
import { computed, inject, ref, Ref, useTemplateRef, watch, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { LocationQueryRaw, useRoute } from 'vue-router';
import { ILoadBalancerWithDeleteProtectionItem, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { ConditionKeyType, SearchConditionFactory } from '../children/search/condition-factory';
import { ISearchCondition, ISearchSelectValue } from '@/typings';
import { transformSimpleCondition } from '@/utils/search';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_LOAD_BALANCER_DETAILS, MENU_BUSINESS_LOAD_BALANCER_OVERVIEW } from '@/constants/menu-symbol';
import { getInstVip, parseIP } from '@/utils';

import { VirtualRender } from 'bkui-vue';
import Search from '../children/search/search.vue';
import allLoadBalancerIcon from '@/assets/image/all-lb.svg';
import loadBalancerIcon from '@/assets/image/loadbalancer.svg';

defineOptions({ name: 'load-balancer-list' });

const route = useRoute();
const { t } = useI18n();
const loadBalancerClbStore = useLoadBalancerClbStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const conditionProperties = SearchConditionFactory.createModel(ConditionKeyType.CLB).getProperties();
const conditionIds = ['vendor', 'name', 'lb_vip', 'domain', 'lb_type', 'ip_version'];
const searchFields = conditionIds.map((id) => conditionProperties.find((item) => item.id === id));

const condition = ref<Record<string, any>>({});
const loadBalancerList = ref<ILoadBalancerWithDeleteProtectionItem[]>([]);

const isSearch = computed(() => Object.keys(condition.value).length > 0);

const validateValues: ValidateValuesFunc = async (item, values) => {
  if (!item) return '请选择条件';
  if ('lb_vip' === item.id) {
    const { IPv4List, IPv6List } = parseIP(values[0].id);
    return Boolean(IPv4List.length || IPv6List.length) ? true : 'IP格式有误';
  }
  return true;
};

const handleSearch = (_val: ISearchSelectValue, cond: ISearchCondition) => {
  condition.value = cond;
};

const isLoading = ref(false);
watch(
  condition,
  async (condition) => {
    isLoading.value = true;
    try {
      const { list } = await loadBalancerClbStore.getLoadBalancerListWithDeleteProtection(
        { filter: transformSimpleCondition(condition, conditionProperties) },
        currentGlobalBusinessId.value,
      );
      loadBalancerList.value = list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    } catch (error) {
      console.error(error);
      loadBalancerList.value = [];
    } finally {
      isLoading.value = false;
    }
  },
  { immediate: true, deep: true },
);

const activeLoadBalancerId = ref<string>();
const handleClick = (id: string) => {
  activeLoadBalancerId.value = id;

  const query: LocationQueryRaw = { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value };
  if (id) {
    routerAction.redirect({ name: MENU_BUSINESS_LOAD_BALANCER_DETAILS, params: { id }, query });
  } else {
    routerAction.redirect({ name: MENU_BUSINESS_LOAD_BALANCER_OVERVIEW, query });
  }
};

watchEffect(() => {
  activeLoadBalancerId.value = route.params.id as string;
});

const virtualRenderRef = useTemplateRef<typeof VirtualRender>('virtual-render');
const fixToActive = (id: string) => {
  virtualRenderRef.value?.fixToTop({ id });
};

defineExpose({
  fixToActive,
});
</script>

<template>
  <section class="load-balancer-list-container">
    <div class="search">
      <search
        :placeholder="t('搜索负载均衡名称、VIP')"
        :fields="searchFields"
        :condition="condition"
        :validate-values="validateValues"
        @search="handleSearch"
      />
    </div>
    <div class="list-container">
      <div class="load-balancer-item is-all" :class="{ active: !activeLoadBalancerId }" @click="handleClick('')">
        <img :src="allLoadBalancerIcon" alt="" style="height: 20px; width: 20px; margin-right: 8px" />
        <span class="label">{{ t('全部负载均衡') }}</span>
        <span class="count">{{ loadBalancerList.length }}</span>
      </div>
      <bk-virtual-render
        v-if="isLoading || loadBalancerList.length"
        ref="virtual-render"
        class="load-balancer-list"
        :height="300"
        :line-height="36"
        :list="loadBalancerList"
        row-key="id"
        v-bkloading="{ loading: isLoading }"
      >
        <template #default="{ data }">
          <div
            v-for="loadBalancer in data"
            :key="loadBalancer.id"
            class="load-balancer-item"
            :class="{ active: loadBalancer.id === activeLoadBalancerId }"
            @click="handleClick(loadBalancer.id)"
          >
            <img :src="loadBalancerIcon" alt="" style="height: 20px; width: 20px; margin-right: 8px" />
            <bk-overflow-title type="tips" class="label">
              {{ `${loadBalancer.name} (${getInstVip(loadBalancer)})` }}
            </bk-overflow-title>
          </div>
        </template>
      </bk-virtual-render>
      <bk-exception
        v-else
        scene="part"
        :description="isSearch ? '搜索为空' : '没有数据'"
        :type="isSearch ? 'search-empty' : 'empty'"
      />
    </div>
  </section>
</template>

<style scoped lang="scss">
.load-balancer-list-container {
  height: 100%;
  display: flex;
  flex-direction: column;

  .search {
    flex-shrink: 0;
    padding: 12px 24px;

    :deep(.bk-search-select) {
      background: #f0f1f5;

      .bk-search-select-container {
        border-color: #fff;
      }
    }
  }

  .list-container {
    flex: 1;
    display: flex;
    flex-direction: column;

    .load-balancer-item {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      padding: 0 16px;
      height: 36px;
      line-height: 22px;
      color: #4d4f56;
      cursor: pointer;

      .label {
        max-width: calc(100% - 28px);
      }

      .count {
        margin-left: auto;
        font-size: 12px;
        color: #c4c6cc;
      }

      &.is-all {
        border: 1px solid #dcdee5;
      }

      &.active {
        background: #e1ecff !important;
        color: #3a84ff;
      }

      &:hover {
        background: #f0f1f5;
      }
    }

    .load-balancer-list {
      flex: 1;
    }
  }
}
</style>
