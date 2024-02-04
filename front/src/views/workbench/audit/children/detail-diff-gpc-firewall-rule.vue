<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import isEqual from 'lodash/isEqual';
const { t } = useI18n();

const props = defineProps<{
  action: string;
  detail: { data: any; changed?: any };
  businessList: any[];
}>();

const properties = [
  { id: 'target_tags', name: t('目标') },
  { id: 'priority', name: t('优先级') },
  { id: 'allowed', name: t('协议和端口') },
];

const isShowBefore = computed(() => props.action !== 'create');
const isShowAfter = computed(() => props.action !== 'delete');

const rows = computed(() =>
  properties.map((item) => {
    const before = props.detail?.data?.[item.id];
    const after = props.detail?.changed?.[item.id];
    return {
      prop: item,
      field: item.name,
      before: before || '--',
      after: after || before || '--',
      changed: before && after && !isEqual(before, after),
    };
  }),
);
const getCellStyle = (column, index, row) => {
  if (index > 0 && row.changed) {
    return {
      backgroundColor: '#e9faf0',
    };
  }
  return {};
};
</script>

<template>
  <bk-table :data="rows" :cell-style="getCellStyle">
    <bk-table-column :label="t('属性')" prop="field" />
    <bk-table-column :label="t('变更前')" v-if="isShowBefore" prop="before" show-overflow-tooltip>
      <template #default="{ cell, row }">
        <template v-if="row?.prop?.id === 'target_tags'">
          {{ cell?.join?.(',') || '--' }}
        </template>
        <template v-else-if="row?.prop?.id === 'allowed'">
          <div v-for="(item, index) in cell" :key="index">{{ item.protocol }}:{{ item.port?.join?.(',') }}</div>
        </template>
        <span v-else>{{ cell }}</span>
      </template>
    </bk-table-column>
    <bk-table-column :label="t('变更后')" v-if="isShowAfter" prop="after" show-overflow-tooltip>
      <template #default="{ cell, row }">
        <template v-if="row?.prop?.id === 'target_tags'">
          {{ cell?.join?.(',') }}
        </template>
        <template v-else-if="row?.prop?.id === 'allowed'">
          <div v-for="(item, index) in cell" :key="index">{{ item.protocol }}:{{ item.port?.join?.(',') }}</div>
        </template>
        <span v-else>{{ cell }}</span>
      </template>
    </bk-table-column>
  </bk-table>
</template>
