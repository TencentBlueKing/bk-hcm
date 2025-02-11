<script setup lang="ts">
import { computed, h } from 'vue';
import { useI18n } from 'vue-i18n';
import { type ISecurityGroupOperateItem } from '@/store/security-group';

import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';

const props = defineProps<{ fieldIds: string[]; detail: ISecurityGroupOperateItem }>();

const { t } = useI18n();

const fields = [
  { id: 'name', name: t('安全组名称'), type: 'string' },
  { id: 'rule_count', name: t('规则数'), render: () => `${props.detail?.rule_count} ${t('个')}` },
  {
    id: 'rel_res',
    name: t('绑定实例'),
    render: () => {
      const { resources = [] } = props.detail ?? {};
      return h(
        'div',
        { class: 'rel-res-wrap' },
        resources.map(({ res_name, count }: any) =>
          h('span', null, [res_name, h('span', { class: `number${count > 0 ? ' danger' : ''}` }, count)]),
        ),
      );
    },
  },
  { id: 'manager', name: t('主负责人'), type: 'user' },
  { id: 'bak_manager', name: t('备份负责人'), type: 'user' },
  { id: 'mgmt_biz_id', name: t('管理业务'), type: 'business' },
  { id: 'usage_biz_ids', name: t('使用业务'), type: 'business' },
];

const renderFields = computed(() => fields.filter((field) => props.fieldIds.includes(field.id)));
</script>

<template>
  <grid-container class="detail-wrap" :column="1" :label-width="120">
    <grid-item v-for="field in renderFields" :key="field.id" :label="field.name">
      <template v-if="field.render">
        <component :is="field.render" />
      </template>
      <display-value v-else :property="field" :value="detail?.[field.id]" :display="{ on: 'info' }" />
    </grid-item>
  </grid-container>
</template>

<style scoped lang="scss">
.detail-wrap {
  :deep(.rel-res-wrap) {
    display: flex;
    align-items: center;

    .number {
      padding: 0 8px;
      display: inline-block;
      height: 16px;
      line-height: 16px;
      background: #eaebf0;
      border-radius: 2px;
      font-size: 12px;
      color: #4d4f56;

      &.danger {
        background: #ea3636;
        color: #fff;
      }
    }

    & > span:not(:last-of-type)::after {
      content: ';';
      margin: 0 4px;
    }
  }
}
</style>
