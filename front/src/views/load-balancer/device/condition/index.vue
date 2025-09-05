<script setup lang="ts">
import { reactive, inject, Ref, watch, computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { IAccountItem } from '@/typings';
import AccountSelector from '@/components/account-selector/index-new.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector.vue';
import { cloneDeep, isEqual } from 'lodash';
import { Info } from 'bkui-vue/lib/icon';

import { selectField, inputField, ILoadBalanceDeviceCondition } from '../common';
import { VendorEnum, ResourceTypeEnum } from '@/common/constant';

defineOptions({ name: 'device-condition' });
defineProps<{ loading: Boolean }>();

const emit = defineEmits(['save']);

let timeout: string | number | NodeJS.Timeout = null;

const businessId = inject<Ref<number>>('currentGlobalBusinessId');
const { t } = useI18n();

const formModel = reactive<ILoadBalanceDeviceCondition>({ account_id: '', vendor: VendorEnum.TCLOUD, lb_regions: [] });
const originFormModel: ILoadBalanceDeviceCondition = reactive(cloneDeep(formModel));
const isShow = ref(false);
const hasSaved = ref(false);

const hasChange = computed(() => !isEqual(formModel, originFormModel));
const disabled = computed(() => !hasSaved.value && !hasChange.value);

const handleAccountChange = (item: IAccountItem) => {
  formModel.vendor = item.vendor;
  formModel.lb_regions = [];
};
const handlePaste = (value: any) => value.split(/,|;|\n|\s/).map((tag: any) => ({ id: tag, name: tag }));
const handleSave = () => {
  if (!hasChange.value) return;
  Object.keys(formModel).forEach((key) => {
    originFormModel[key] = formModel[key];
  });
  hasSaved.value = true;
  isShow.value = false;
  emit('save', originFormModel);
};
const handleReset = () => {
  Object.keys(formModel).forEach((key) => {
    formModel[key] = originFormModel[key];
  });
};

watch(
  () => formModel.account_id,
  (val) => {
    originFormModel.account_id = val;
    originFormModel.vendor = formModel.vendor;
  },
  {
    once: true,
  },
);
watch(
  () => hasChange.value,
  (val) => {
    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
    if (val) {
      timeout = setTimeout(() => (isShow.value = true), 120000);
    }
  },
);
</script>

<template>
  <div class="device-condition">
    <div class="header">{{ t('检索条件') }}</div>
    <div class="condition">
      <bk-form ref="condition-form" class="condition-form g-expand" form-type="vertical" :model="formModel">
        <bk-form-item :label="t('云账号')" property="account_id" required>
          <account-selector
            v-model="formModel.account_id"
            :biz-id="businessId"
            :auto-select="true"
            :resource-type="ResourceTypeEnum.CLB"
            @change="handleAccountChange"
          />
        </bk-form-item>
        <bk-form-item :label="t('地域')" property="lb_regions">
          <region-selector v-model="formModel.lb_regions" :vendor="formModel.vendor" multiple clearable collapse-tags />
        </bk-form-item>
        <bk-form-item :label="t(item.name)" :property="item.id" v-for="item in selectField" :key="item.id">
          <bk-select
            v-model="formModel[item.id]"
            :list="item.list"
            clearable
            multiple
            multiple-mode="tag"
            collapse-tags
          />
        </bk-form-item>
        <bk-form-item :label="t(item.name)" :property="item.id" v-for="item in inputField" :key="item.id">
          <bk-tag-input
            v-model="formModel[item.id]"
            :paste-fn="handlePaste"
            placeholder="请输入"
            allow-create
            collapse-tags
          />
        </bk-form-item>
      </bk-form>
    </div>

    <div class="footer">
      <bk-popover theme="light" :is-show="isShow" trigger="manual">
        <bk-button
          class="mr10 save"
          theme="primary"
          @click="handleSave"
          :loading="loading"
          :disabled="disabled"
          v-bk-tooltips="{ content: t('请输入检索条件后点击'), disabled: !disabled }"
        >
          {{ t('查询') }}
        </bk-button>
        <template #content>
          <div class="tips">
            <info class="warning" />
            <div>{{ t('检索条件有更新，请点击下方查询按钮更新检索') }}</div>
          </div>
        </template>
      </bk-popover>
      <bk-button @click="handleReset">{{ t('重置') }}</bk-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.device-condition {
  height: 100%;
  padding: 16px 0px 16px 24px;
  position: relative;

  .header {
    font-weight: 700;
    color: #313238;
    margin-bottom: 16px;
  }
  .condition {
    height: calc(100% - 64px);
    overflow-y: auto;
    position: relative;

    .condition-form {
      padding-right: 24px;
    }
  }
  .footer {
    position: sticky;
    bottom: 0;
    width: calc(100% + 48px);
    margin-left: -24px;
    line-height: 48px;
    background: #fafbfd;
    border: 1px solid #dcdee5;
    padding-left: 24px;
    z-index: 999999;
  }
}
.tips {
  color: #4d4f56;
  display: flex;
  align-items: center;
  width: 180px;

  .warning {
    color: #f59500;
    margin-right: 5px;
  }
}
</style>
