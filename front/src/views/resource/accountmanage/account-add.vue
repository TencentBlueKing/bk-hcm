<template>
  <div class="account-add-warp">
    <bk-form
      ref="formRef"
      class="example"
      :model="formModel"
      :rules="rules">
      <bk-form-item
        label="名称"
        property="name"
      >
        <bk-input
          class="w450"
          v-model="formModel.name"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="名称"
        property="name"
      >
        <bk-input
          class="w450"
          v-model="formModel.name"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="云厂商"
        property="cloudName"
      >
        <bk-select class="w450" v-model="formModel.cloudName">
          <bk-option
            v-for="(item, index) in cloudType"
            :key="index"
            :value="item.value"
            :label="item.key"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        label="主账号"
        property="account"
      >
        <bk-input
          class="w450"
          v-model="formModel.account"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="子账号ID"
        property="subAccountId"
      >
        <bk-input
          class="w450"
          v-model="formModel.subAccountId"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="子账号名称"
        property="subAccountName"
      >
        <bk-input
          class="w450"
          v-model="formModel.subAccountName"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="密钥ID"
        property="SecretId"
      >
        <bk-input
          class="w450"
          v-model="formModel.SecretId"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="SecretKey"
        property="SecretKey"
      >
        <bk-input
          class="w450"
          v-model="formModel.SecretKey"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="责任人"
        property="SecretKey"
      >
        <bk-input
          class="w450"
          v-model="formModel.SecretKey"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="组织架构"
        property="SecretKey"
      >
        <bk-input
          class="w450"
          v-model="formModel.SecretKey"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
      <bk-form-item
        label="使用业务"
        property="SecretKey"
      >
        <bk-select
          class="w450"
          v-model="formModel.cloudName"
          filterable
          multiple
          show-select-all
          multiple-mode="tag"
          collapse-tags>
          <bk-option
            v-for="(item, index) in cloudType"
            :key="index"
            :value="item.value"
            :label="item.key"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        label="描述"
        property="desc"
      >
        <bk-input
          class="w450"
          type="textarea"
          v-model="formModel.desc"
          placeholder="请输入"
          clearable
        />
      </bk-form-item>
    </bk-form>
    <bk-button class="w90 account-btn-cls" theme="primary" @click="submit">
      {{t('确认')}}
    </bk-button>
  </div>
</template>
<script lang="ts">
import { reactive, toRefs, defineComponent, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();
    const state = reactive({
      formModel: {
        name: '',
        cloudName: '',
        account: '',
        subAccountId: '',
        subAccountName: '',
        SecretId: '',
        SecretKey: '',
        desc: '',
      },
      rules: {},
      cloudType: [
        { key: '华为云', value: 'huawei' },
      ],
    });
    onMounted(async () => {
      console.log(11112334);
    });

    onUnmounted(() => {
    });

    const submit = () => {};

    return {
      ...toRefs(state),
      submit,
      t,
    };
  },
});
</script>
<style lang="scss">
    .account-btn-cls{
        margin-right: 150px;
    }
</style>
