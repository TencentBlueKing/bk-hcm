<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { ref, watch, nextTick } from 'vue';
import FormSelect from '@/views/business/components/form-select.vue';
import ResourceGroup from '@/components/resource-group/index.vue';
import { BusinessFormFilter } from '@/typings';
import { useBusinessStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { type IBusinessItem } from '@/store/business-global';
import SecurityGroupManagerSelector from '@/views/resource/resource-manage/children/components/security/manager-selector/index.vue';
import BusinessSelector from '@/components/business-selector/business.vue';
import { useAccountBusiness } from '@/views/resource/resource-manage/hooks/use-account-business';

const { t } = useI18n();
const useBusiness = useBusinessStore();
const resourceAccountStore = useResourceAccountStore();

const formRef = ref(null);
const formSelectRef = ref(null);
const personSelectorRef = ref(null);
const type = ref('tcloud');
const formFilter = ref<any>({});
const cloudVpcId = ref('');
const submitLoading = ref(false);
const filter = ref<any>({
  filter: { op: 'and', rules: [{ field: 'vendor', op: 'eq', value: 'aws' }] },
});
let usageBizList = ref<IBusinessItem[]>([]);

const props = defineProps<{
  isFormDataChanged: boolean;
  show: boolean;
}>();
const emit = defineEmits(['cancel', 'success', 'update:isFormDataChanged']);
const handleFormFilter = (value: BusinessFormFilter) => {
  formFilter.value = { ...value };
  type.value = value.vendor;
  !props.isFormDataChanged && emit('update:isFormDataChanged', true);
};
const formData = ref({ name: '', memo: '', resource_group_name: '', usage_biz_ids: [] });

const rules = {
  resource_group_name: [
    {
      validator: (value: string) => value.length > 0,
      message: '资源组必填',
      trigger: 'blur',
    },
  ],
};
const { isResourcePage } = useWhereAmI();

const reset = () => {
  formData.value.name = '';
  formData.value.memo = '';
  formData.value.resource_group_name = '';
  formData.value.usage_biz_ids = [];
  nextTick(() => formRef.value.clearValidate());
};
const resetAll = () => {
  reset();
  personSelectorRef?.value?.reset?.();
};
const cancel = async () => {
  emit('cancel');
};
const submit = async () => {
  const { formData: personSelectorParams, validate } = personSelectorRef.value;
  await Promise.all([formSelectRef.value[0](), formRef.value.validate(), validate()]);
  const params: any = { ...formData.value, ...formFilter.value, ...personSelectorParams };
  if (type.value === 'aws') {
    params.extension = {
      cloud_vpc_id: cloudVpcId.value,
    };
    params.extension.cloud_vpc_id = cloudVpcId.value;
  } else if (type.value === 'azure') {
    params.extension = {
      resource_group_name: formData.value.resource_group_name,
    };
  }
  delete params.resource_group_name;
  if (!isResourcePage) {
    delete params.usage_biz_ids;
  }
  try {
    submitLoading.value = true;
    await useBusiness.addSecurity(params, isResourcePage);
    Message({
      theme: 'success',
      message: t('新增成功'),
    });
    emit('success');
  } catch (error) {
  } finally {
    submitLoading.value = false;
  }
};
const getAccountList = () => {
  const accountId = resourceAccountStore.resourceAccount?.id ?? '';
  if (!isResourcePage || !accountId) return;

  const { accountBizList } = useAccountBusiness(accountId);
  usageBizList = accountBizList;

  // 默认填充后，清除表单校验结果
  nextTick(() => formRef.value.clearValidate());
};

watch(
  () => formFilter.value.region,
  (val) => {
    if (val) {
      filter.value.filter.rules = [
        { field: 'vendor', op: 'eq', value: 'aws' },
        { field: 'region', op: 'eq', value: val },
      ];
    }
  },
);

watch(
  () => formFilter,
  () => {
    cloudVpcId.value = '';
  },
  {
    deep: true,
  },
);

watch(
  () => formData,
  () => {
    !props.isFormDataChanged && emit('update:isFormDataChanged', true);
  },
  { deep: true },
);

watch(
  () => props.show,
  (val) => {
    resetAll();
    if (val) {
      getAccountList();
    }
  },
  {
    immediate: true,
  },
);

const { datas, isLoading } = useQueryList(filter.value, 'vpcs'); // 只查aws的vpcs
</script>

<template>
  <div class="business-dialog-warp">
    <div class="form-manage-type" v-if="isResourcePage">
      管理类型：
      <span>平台管理</span>
    </div>
    <form-select
      @change="handleFormFilter"
      type="security"
      ref="formSelectRef"
      :show="props.show"
      :hidden="['vendor']"
    ></form-select>
    <bk-form class="form-subnet" label-width="150" :model="formData" :rules="rules" form-type="vertical" ref="formRef">
      <bk-form-item :label="t('名称')" class="item-warp" required property="name">
        <bk-input class="item-warp-component" v-model="formData.name" :placeholder="t('请输入安全组名称')" />
      </bk-form-item>

      <bk-form-item :label="t('备注')" class="item-warp">
        <bk-input
          type="textarea"
          class="item-warp-component"
          v-model="formData.memo"
          :resize="false"
          :placeholder="t('请输入备注')"
        />
      </bk-form-item>

      <bk-form-item v-if="type === 'aws'" :label="t('所属的vpc')" :loading="isLoading" class="item-warp" required>
        <bk-select class="item-warp-component" v-model="cloudVpcId">
          <bk-option
            v-for="(item, index) in datas"
            :key="index"
            :value="item.cloud_id"
            :label="`${item.cloud_id}（${item.name || '--'}）`"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item v-if="type === 'azure'" label="资源组" property="resource_group_name" required>
        <resource-group
          :vendor="formFilter.vendor"
          :region="formFilter.region"
          v-model="formData.resource_group_name"
        />
      </bk-form-item>
      <bk-form-item label="使用业务" property="usage_biz_ids" required v-if="isResourcePage">
        <business-selector
          v-model="formData.usage_biz_ids"
          :multiple="true"
          :clearable="true"
          :filterable="false"
          :collapse-tags="true"
          :data="usageBizList"
          :show-all="usageBizList ? false : true"
          :all-option-id="-1"
        />
      </bk-form-item>
      <security-group-manager-selector ref="personSelectorRef" />

      <bk-form-item label-width="150" class="item-warp">
        <bk-button class="item-warp-button" theme="primary" @click="submit" :loading="submitLoading">
          {{ t('提交创建') }}
        </bk-button>
        <bk-button class="ml10 item-warp-button" @click="cancel">
          {{ t('取消') }}
        </bk-button>
      </bk-form-item>
    </bk-form>
  </div>
</template>
<style lang="scss" scoped>
.form-subnet {
  .item-warp-button {
    min-width: 88px;
  }
}

.business-dialog-warp {
  padding: 30px 40px;
}

.form-manage-type {
  font-size: 14px;
  color: #4d4f56;
  line-height: 22px;

  > span {
    font-size: 12px;
    background: #f0f1f5;
    border-radius: 11px;
    width: 64px;
    height: 22px;
    display: inline-block;
    text-align: center;
  }
}
</style>
