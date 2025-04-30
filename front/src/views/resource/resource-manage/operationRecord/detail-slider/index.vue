<template>
  <bk-sideslider v-model:isShow="model" title="操作详情" :width="640" @closed="handleClosed">
    <grid-container
      class="info-display-container"
      :column="2"
      :content-min-width="200"
      :label-width="80"
      :gap="[0, 12]"
    >
      <grid-item v-for="field in infoFields" :key="field.id" :label="field.name">
        <display-value :property="field" :value="info[field.id]" :display="{ ...field.meta?.display, on: 'info' }" />
      </grid-item>
    </grid-container>
    <div class="json-display-container">
      <pre><code>{{detailJsonStr}}</code></pre>
    </div>
  </bk-sideslider>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { ModelProperty } from '@/model/typings';
import { IAuditItem } from '@/store/audit';

import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';

const props = defineProps<{ properties: ModelProperty[]; info: IAuditItem }>();
const model = defineModel<boolean>();
const emit = defineEmits(['closed']);

const infoFieldIds = [
  'res_type',
  'vendor',
  'account_id',
  'bk_biz_id',
  'res_id',
  'res_name',
  'cloud_res_id',
  'action',
  'operator',
  'created_at',
  'rid',
  'source',
];
const infoFields = infoFieldIds.map((id) => props.properties.find((p) => p.id === id));
const detailJsonStr = computed(() => {
  if (props.info?.detail?.data) {
    return JSON.stringify(props.info.detail.data, null, 2);
  }
  return '';
});

const handleClosed = () => {
  model.value = false;
  emit('closed');
};
</script>

<style scoped lang="scss">
.info-display-container {
  padding: 20px;
}

.json-display-container {
  padding: 0 20px 20px;
  font-size: 12px;
  color: #bfc6e0;

  pre {
    border-radius: 2px;
    background: #455070;
    overflow-x: auto;
  }
}
</style>
