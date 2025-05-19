<script setup lang="ts">
import { h, VNode } from 'vue';
import { useI18n } from 'vue-i18n';
import { type ISecurityGroupOperateItem } from '@/store/security-group';
import type { ModelProperty } from '@/model/typings';

import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import DisplayValue from '@/components/display-value/index.vue';
import RelResourcesDisplay from '../../components/security/rel-resources-display.vue';
import hintIcon from '@/assets/image/hint.svg';

const props = defineProps<{ detail: ISecurityGroupOperateItem; loading: boolean }>();
const emit = defineEmits(['confirm']);
const model = defineModel<boolean>();

const { t } = useI18n();

const fields = [
  { id: 'name', name: t('安全组名称'), type: 'string' },
  { id: 'rule_count', name: t('规则数'), render: () => `${props.detail?.rule_count} ${t('个')}` },
  {
    id: 'rel_res',
    name: t('绑定实例'),
    render: () => {
      const { resources = [] } = props.detail ?? {};
      const displayResources = resources?.filter(({ count }) => count > 0) || [];

      if (!displayResources.length) return '--';
      return h(RelResourcesDisplay, { resources: displayResources });
    },
  },
  {
    id: 'usage_biz_ids',
    name: t('使用业务'),
    render: () => {
      const showIcon = props.detail?.usage_biz_ids?.length > 1;
      return h('div', { class: 'usage-biz-wrap' }, [
        h(DisplayValue, {
          property: { id: 'usage_biz_ids', name: t('使用业务'), type: 'business' },
          value: props.detail?.usage_biz_ids,
        }),
        showIcon ? h('i', { class: 'hcm-icon bkhcm-icon-alert ml4', style: 'color: #E71818' }) : null,
      ]);
    },
  },
] as Array<ModelProperty & { render: () => VNode }>;

const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog :is-show="model" footer-align="center" @confirm="emit('confirm')" @closed="handleClosed">
    <div class="hint-wrap">
      <img :src="hintIcon" />
      <div>{{ t('变更影响确认') }}</div>
    </div>

    <template v-if="loading">
      <bk-loading mode="spin" theme="primary" loading>
        <div style="width: 100%; height: 160px"></div>
      </bk-loading>
    </template>
    <template v-else>
      <grid-container class="detail-preview-wrap" :column="1" :label-width="120">
        <grid-item v-for="field in fields" :key="field.id" :label="field.name">
          <template v-if="field.render">
            <component :is="field.render" />
          </template>
          <display-value v-else :property="field" :value="detail?.[field.id]" :display="{ on: 'info' }" />
        </grid-item>
      </grid-container>

      <bk-alert v-if="detail?.usage_biz_ids?.length > 1" theme="danger" :show-icon="false">
        {{
          t(
            '该安全组跨多业务使用，变更会影响多个业务的服务，建议将当前安全组克隆后绑定到相关实例，避免跨业务操作影响其他业务。',
          )
        }}
      </bk-alert>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.hint-wrap {
  margin-top: -35px;
  text-align: center;

  img {
    width: 42px;
    height: 42px;
  }

  div {
    font-size: 20px;
    color: #313238;
  }
}
</style>
