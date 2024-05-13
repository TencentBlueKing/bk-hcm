<script lang="ts" setup>
// import { defineComponent, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { ref, watch } from 'vue';
import FormSelect from '@/views/business/components/form-select.vue';
import ResourceGroup from '@/components/resource-group/index.vue';
import { BusinessFormFilter } from '@/typings';
import { useBusinessStore } from '@/store';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const { t } = useI18n();
const formRef = ref(null);
const formSelectRef = ref(null);
const type = ref('tcloud');
const formFilter = ref<any>({});
const cloudVpcId = ref('');
const submitLoading = ref(false);
const filter = ref<any>({
  filter: { op: 'and', rules: [{ field: 'vendor', op: 'eq', value: 'aws' }] },
});

const useBusiness = useBusinessStore();
const props = defineProps<{
  isFormDataChanged: boolean;
}>();
const emit = defineEmits(['cancel', 'success', 'update:isFormDataChanged']);
const handleFormFilter = (value: BusinessFormFilter) => {
  formFilter.value = { ...value };
  type.value = value.vendor;
  !props.isFormDataChanged && emit('update:isFormDataChanged', true);
};
const formData = ref({ name: '', memo: '', resource_group_name: '' });

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

// 方法
const cancel = async () => {
  emit('cancel');
};
const submit = async () => {
  await Promise.all([formSelectRef.value[0](), formRef.value.validate()]);
  const params: any = { ...formData.value, ...formFilter.value };
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

const { datas, isLoading } = useQueryList(filter.value, 'vpcs'); // 只查aws的vpcs
</script>

<template>
  <div class="business-dialog-warp">
    <form-select @change="handleFormFilter" type="security" ref="formSelectRef"></form-select>
    <bk-form class="form-subnet" label-width="150" :model="formData" :rules="rules" ref="formRef">
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
  padding-right: 20px;
  .item-warp-button {
    min-width: 88px;
  }
}
</style>
