<script lang="ts" setup>
// import { defineComponent, reactive } from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import {
  Message,
} from 'bkui-vue';
import {
  ref,
} from 'vue';
import FormSelect from '@/views/business/components/form-select.vue';
import { BusinessFormFilter } from '@/typings';
import { useBusinessStore } from '@/store';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';

const { t } = useI18n();
const type = ref('tcloud');
const formFilter = ref({});
const cloudVpcId = ref('');
const submitLoading = ref(false);
const useBusiness = useBusinessStore();
const emit = defineEmits(['cancel', 'success']);
const handleFormFilter = (value: BusinessFormFilter) => {
  console.log('value', value);
  formFilter.value = { ...value };
  type.value = value.vendor;
};
const formData = ref({ name: '', memo: '' });

// 方法
const cancel = async () => {
  emit('cancel');
};
const submit = async () => {
  const params: any = { ...formData.value, ...formFilter.value };
  if (type.value === 'aws') {
    console.log('cloudVpcId.value', cloudVpcId.value, params.extension);
    params.extension = {
      cloud_vpc_id: cloudVpcId.value,
    };
    params.extension.cloud_vpc_id = cloudVpcId.value;
  }
  try {
    submitLoading.value = true;
    await useBusiness.addSecurity(params);
    Message({
      theme: 'success',
      message: t('新增成功'),
    });
    emit('success');
  } catch (error) {
    console.log(error);
  } finally {
    submitLoading.value = false;
  }
};

const {
  datas,
  isLoading,
} = useQueryList({ filter: { op: 'and', rules: [{ field: 'vendor', op: 'eq', value: 'aws' }] } }, 'vpcs');  // 只查aws的vpcs
</script>

<template>
  <div class="business-dialog-warp">
    <form-select @change="handleFormFilter"></form-select>
    <bk-form class="form-subnet">
      <bk-form-item
        :label="t('名称')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" v-model="formData.name" :placeholder="t('请输入子网名称')" />
      </bk-form-item>

      <bk-form-item
        :label="t('备注')"
        class="item-warp"
      >
        <bk-input type="textarea" class="item-warp-component" v-model="formData.memo" :placeholder="t('请输入备注')" />
      </bk-form-item>

      <bk-form-item
        v-if="type === 'aws'"
        :label="t('所属的vpc')"
        :loading="isLoading"
        class="item-warp"
      >
        <bk-select
          class="item-warp-component"
          v-model="cloudVpcId"
        >
          <bk-option
            v-for="(item, index) in datas"
            :key="index"
            :value="item.cloud_id"
            :label="item.name"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        label-width="50"
        class="item-warp mt40"
      >
        <bk-button class="item-warp-button" theme="primary" @click="submit" :loading="submitLoading">
          {{t('提交创建')}}
        </bk-button>
        <bk-button class="ml20 item-warp-button" @click="cancel">
          {{t('取消')}}
        </bk-button>
      </bk-form-item>
    </bk-form>
  </div>
</template>
<style lang="scss" scoped>
.form-subnet{
  padding-right: 20px;
  .item-warp-button{
    width: 100px;
  }
  .item-button-group{
    .item-button{
      margin-left: 10px;
    }
  }
}
</style>
