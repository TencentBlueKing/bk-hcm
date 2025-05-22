<script setup lang="ts">
import { computed, onBeforeMount, ref, useAttrs } from 'vue';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';
import rollRequest from '@blueking/roll-request';
import http from '@/http';

defineOptions({ name: 'hcm-form-region' });

interface IResourceGroupItem {
  id: string;
  name: string;
  type: string;
  location: string;
  account_id: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

const props = defineProps<{ accountId: string; multiple?: boolean; clearable?: boolean }>();
const model = defineModel<string | string[]>();
const attrs = useAttrs();

const localModel = computed({
  get() {
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    model.value = val;
  },
});

const loading = ref(false);
const resourceGroupList = ref<IResourceGroupItem[]>();
const getResourceGroupList = async () => {
  loading.value = true;
  try {
    const filter: QueryFilterType = {
      op: 'and',
      rules: [
        { field: 'type', op: QueryRuleOPEnum.EQ, value: 'Microsoft.Resources/resourceGroups' },
        { field: 'account_id', op: QueryRuleOPEnum.EQ, value: props.accountId },
      ],
    };
    const list = await rollRequest({
      httpClient: http,
      pageEnableCountKey: 'count',
    }).rollReqUseCount<IResourceGroupItem>(
      '/api/v1/cloud/vendors/azure/resource_groups/list',
      { filter },
      { limit: 500, listGetter: (res) => res.data.details, countGetter: (res) => res.data.count },
    );
    resourceGroupList.value = list;
  } catch (error) {
    console.error(error);
    return Promise.reject();
  } finally {
    loading.value = false;
  }
};

onBeforeMount(() => {
  getResourceGroupList();
});
</script>

<template>
  <bk-select
    v-model="localModel"
    :list="resourceGroupList"
    :clearable="clearable"
    :multiple="multiple"
    :multiple-mode="multiple ? 'tag' : 'default'"
    :id-key="'name'"
    :display-key="'name'"
    v-bind="attrs"
  />
</template>
