<script setup lang="ts">
import { inject, ref, type Ref } from 'vue';
import { VendorEnum } from '@/common/constant';
import type { ModelPropertyDisplay } from '@/model/typings';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import { FieldFactory } from './field-factory';
import type { DetailsFieldTcloud } from './field-tcloud';

defineProps<{
  data: IPermissionTemplateItem & DetailsFieldTcloud;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const model = FieldFactory.createModel(currentVendor.value);
const properties = model.getPropertiesByGroup<ModelPropertyDisplay>();
</script>

<template>
  <div class="permission-template-details">
    <div v-for="(fields, group) in properties" :key="group" class="details-panel">
      <div class="panel-title">{{ group }}</div>
      <grid-container :column="1" :label-width="120">
        <grid-item v-for="field in fields" :key="field.id" :label="field.id === 'policy_document' ? null : field.name">
          <display-value
            :property="field"
            :value="field.id === 'extension.cloud_type' ? data : data[field.id]"
            :display="{ ...field.meta?.display, on: 'info' }"
          />
        </grid-item>
      </grid-container>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.permission-template-details {
  display: flex;
  flex-direction: column;
  gap: 12px;

  .details-panel {
    background: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 4px 0 #1919290d;
    padding: 16px 24px;

    .panel-title {
      font-size: 14px;
      font-weight: 700;
      color: #313238;
      line-height: 22px;
      margin-bottom: 8px;
    }
  }
}
</style>
