<script setup lang="ts">
import { computed } from 'vue';
import Panel from '@/components/panel';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import DisplayValue from '@/components/display-value/index.vue';
import { type ModelProperty } from '@/model/typings';
import { APPLICATION_TYPE_MAP } from '@/views/ticket/constants';
import { VendorMap } from '@/common/constant';
import { type IApplicationDetail } from '../index';

const props = defineProps<{
  details: IApplicationDetail;
  loading: boolean;
  cancelLoading: boolean;
  onCancel: () => void;
}>();

const baseFields: ModelProperty[] = [
  { id: 'type', name: '申请类型', type: 'enum', option: APPLICATION_TYPE_MAP },
  { id: 'creator', name: '申请人', type: 'user' },
  { id: 'memo', name: '申请单备注', type: 'string' },
  { id: 'created_at', name: '申请时间', type: 'datetime' },
  { id: 'updated_at', name: '更新时间', type: 'datetime' },
];

const paramsFields: ModelProperty[] = [
  { id: 'account_id', name: '账号', type: 'string' },
  { id: 'vendor', name: '云厂商', type: 'enum', option: VendorMap },
  { id: 'bk_biz_id', name: '业务名称', type: 'business' },
  { id: 'action', name: undefined, type: 'string' },
  { id: 'req', name: undefined, type: 'json' },
];

const detailsParams = computed(() => {
  try {
    const params = JSON.parse(props.details.content);
    return params;
  } catch (error) {
    console.error(error);
    return {};
  }
});
</script>

<template>
  <div class="common-apply-detail-container">
    <Panel title="基本信息">
      <GridContainer fixed :column="2" :content-min-width="300" :label-width="150">
        <GridItem v-for="field in baseFields" :key="field.id" :label="field.name">
          <DisplayValue
            :property="field"
            :value="details[field.id]"
            :display="{ ...field.meta?.display, on: 'info' }"
          />
        </GridItem>
      </GridContainer>
    </Panel>
    <Panel title="参数信息">
      <GridContainer fixed :column="2" :content-min-width="300" :label-width="150">
        <GridItem v-for="field in paramsFields" :key="field.id" :label="field.name ?? field.id">
          <DisplayValue
            :property="field"
            :value="detailsParams[field.id]"
            :display="{ ...field.meta?.display, on: 'info' }"
          />
        </GridItem>
      </GridContainer>
    </Panel>
  </div>
</template>

<style scoped lang="scss">
.common-apply-detail-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
