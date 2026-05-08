<script setup lang="ts">
import { inject, computed, type Ref, ref, watch, useTemplateRef } from 'vue';
import { Form } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';
import { formatJSON } from '@/utils';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { QueryRuleOPEnum } from '@/typings';
import { usePermissionPolicyStore, type IPermissionPolicyItem } from '@/store/cloud-account-manage/permission-policy';
import { useSecondaryAccountStore } from '@/store/cloud-account-manage/secondary-account';

import type { FieldTcloud } from './field-tcloud';
import { FieldFactory } from './field-factory';
import type { ModelPropertyForm } from '@/model/typings';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';

const props = defineProps<{
  data: IPermissionTemplateItem & FieldTcloud;
  isEdit?: boolean;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const permissionPolicyStore = usePermissionPolicyStore();
const secondaryAccountStore = useSecondaryAccountStore();
const { getBizsId } = useWhereAmI();

const fieldModel = computed(() => FieldFactory.createModel(currentVendor.value));
const properties = computed(() => fieldModel.value.getProperties<ModelPropertyForm>());
const fields = computed(() => properties.value.filter((field) => !field.apiOnly));

const formData = ref(fieldModel.value.createInstance());

const formRef = useTemplateRef<typeof Form>('formRef');

watch(
  () => props.data,
  (newVal) => {
    formData.value.id = newVal?.id; // 仅编辑时存在
    formData.value.account_id = newVal?.account_id;
    formData.value.name = newVal?.name;
    formData.value.type = newVal?.type;
    formData.value.policy_library_id = newVal?.policy_library_id;
    formData.value.policy_document = newVal?.policy_document ? formatJSON(newVal?.policy_document) : '';
    formData.value.memo = newVal?.memo;
  },
  { deep: true, immediate: true },
);

const policyLibraryListGenerator = computed(() =>
  permissionPolicyStore.createPolicyLibraryListGenerator(currentVendor.value, getBizsId()),
);

const getFormCompProps = (field: ModelPropertyForm) => {
  const compProps = { ...(field.meta?.display?.props || {}) };
  if (field.id === 'account_id') {
    compProps.list = async () =>
      await secondaryAccountStore.getSecondaryAccountFullList(getBizsId(), {
        op: 'and',
        rules: [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: currentVendor.value }],
      });
  }
  if (field.id === 'policy_library_id') {
    compProps.listGenerator = policyLibraryListGenerator.value;
  }
  if ((field.id === 'account_id' || field.id === 'name') && props.isEdit) {
    compProps.disabled = true;
  }
  return compProps;
};

const getFormCompEvents = (field: ModelPropertyForm) => {
  if (field.id === 'policy_library_id') {
    return {
      change: (_value: string, item: IPermissionPolicyItem | undefined) => {
        formData.value.policy_document = item?.policy_document ? formatJSON(item.policy_document) : '';
      },
    };
  }
};

defineExpose({
  getFormData: () => formData.value,
  validate: () => formRef.value.validate(),
});
</script>

<template>
  <bk-form ref="formRef" :model="formData" form-type="vertical">
    <bk-form-item
      v-for="field in fields"
      :key="field.name"
      :label="field.name"
      :property="field.id"
      :required="field.required"
      :rules="field.rules"
    >
      <component
        :is="`hcm-form-${field.type}`"
        v-model="formData[field.id]"
        :option="field.option"
        :display="field.meta?.display"
        v-bind="getFormCompProps(field)"
        v-on="getFormCompEvents(field)"
      />
    </bk-form-item>
  </bk-form>
</template>
