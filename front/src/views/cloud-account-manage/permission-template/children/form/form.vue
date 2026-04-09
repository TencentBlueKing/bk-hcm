<script setup lang="ts">
import { inject, computed, type Ref, ref, watch, useTemplateRef, watchEffect } from 'vue';
import { Form } from 'bkui-vue';
import http from '@/http';
import rollRequest from '@blueking/roll-request';
import { type IListResData, QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { formatJSON } from '@/utils';
import type { IPermissionPolicyItem } from '@/store/cloud-account-manage/permission-policy';
import type { FieldTcloud } from './field-tcloud';
import { FieldFactory } from './field-factory';
import type { ModelPropertyForm } from '@/model/typings';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';

const props = defineProps<{
  data: IPermissionTemplateItem & FieldTcloud;
  isEdit?: boolean;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

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

const fetchPolicyLibraryList = (name?: string) => {
  return rollRequest({ httpClient: http, pageEnableCountKey: 'count' }).rollReqUseCount<
    IListResData<IPermissionPolicyItem[]>
  >(
    `/api/v1/cloud/vendors/${currentVendor.value}/permission_policy_libraries/list`,
    {
      filter: name
        ? {
            op: QueryRuleOPEnum.AND,
            rules: [
              {
                field: 'name',
                op: QueryRuleOPEnum.CS,
                value: name,
              },
            ],
          }
        : {},
    },
    {
      limit: 5,
      countGetter: (res) => res.data.count,
      listGetter: (res) => res.data.details,
      generator: true,
    },
    true,
  );
};

// TODO: 考虑新增 scroll-list 组件，把这套包起来
const policyLibraryListGenerator = ref<Generator<Promise<IListResData<IPermissionPolicyItem[]>>, void, any>>(null);
const policyLibraryList = ref<IPermissionPolicyItem[]>([]);
const policyLibraryListLoading = ref(false);
const policyLibraryListScrollLoading = ref(false);
watchEffect(async () => {
  policyLibraryListLoading.value = true;
  policyLibraryListGenerator.value = await fetchPolicyLibraryList();
  policyLibraryListLoading.value = false;
  const result = policyLibraryListGenerator.value.next();
  if (!result.done) {
    const res = await (result.value as Promise<IListResData<IPermissionPolicyItem[]>>);
    policyLibraryList.value = res?.data?.details ?? [];
  }
});

const getFormCompProps = (field: ModelPropertyForm) => {
  const compProps = field.meta?.display?.props || {};
  if (field.id === 'account_id' && props.isEdit) {
    compProps.disabled = true;
  }
  if (field.id === 'policy_library_id') {
    compProps.filterable = true;
    compProps.loading = policyLibraryListLoading.value;
    compProps.scrollLoading = policyLibraryListScrollLoading.value;
    compProps.list = policyLibraryList.value;
    compProps.remoteMethod = async (query: string) => {
      policyLibraryListLoading.value = true;
      policyLibraryListGenerator.value = await fetchPolicyLibraryList(query);
      policyLibraryListLoading.value = false;
      const result = policyLibraryListGenerator.value.next();
      if (!result.done) {
        const res = await (result.value as Promise<IListResData<IPermissionPolicyItem[]>>);
        policyLibraryList.value = res?.data?.details ?? [];
      }
    };
  }
  return compProps;
};

const getFormCompEvents = (field: ModelPropertyForm) => {
  if (field.id === 'policy_library_id') {
    return {
      'scroll-end': async () => {
        if (!policyLibraryListGenerator.value) return;
        const result = policyLibraryListGenerator.value.next();
        if (!result.done) {
          policyLibraryListScrollLoading.value = true;
          const res = await (result.value as Promise<IListResData<IPermissionPolicyItem[]>>);
          policyLibraryList.value = [...policyLibraryList.value, ...(res?.data?.details ?? [])];
          policyLibraryListScrollLoading.value = false;
        }
      },
      change: (id: string) => {
        const policyDocument = policyLibraryList.value.find((item) => item.id === id)?.policy_document;
        formData.value.policy_document = policyDocument ? formatJSON(policyDocument) : '';
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
