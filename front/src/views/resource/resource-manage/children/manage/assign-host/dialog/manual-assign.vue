<script setup lang="ts">
import { computed, h, onBeforeMount, watchEffect } from 'vue';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import useFormModel from '@/hooks/useFormModel';
import { useBusinessGlobalStore } from '@/store/business-global';
import { ICvmsAssignBizsPreviewItem, useHostStore, type CvmsAssignPreviewItem } from '@/store';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import type { FieldList } from '@/views/resource/resource-manage/common/info-list/types';
import { timeFormatter } from '@/common/util';

import { Tag } from 'bkui-vue';
import DisplayValue from '@/components/display-value/index.vue';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

const props = withDefaults(
  defineProps<{
    action: 'backfill' | 'submit'; // backfill：批量分配，submit：单个分配
    cvm: CvmsAssignPreviewItem;
    isLoading?: boolean;
  }>(),
  {
    isLoading: false,
  },
);
const emit = defineEmits<{
  (e: 'backfill', bkBizId: number, bkCloudId: number): void;
  (e: 'submit', cvm: { cvm_id: string; bk_biz_id: number; bk_cloud_id: number }): void;
}>();
const model = defineModel<boolean>();

const { t } = useI18n();
const cloudAreaStore = useCloudAreaStore();
const businessGlobalStore = useBusinessGlobalStore();
const hostStore = useHostStore();

const { cloudAreaMap } = storeToRefs(cloudAreaStore);
const { businessFullList } = storeToRefs(businessGlobalStore);

// 预览
const isPreviewScene = computed(() => props.cvm?.match_type === 'auto' && props.action === 'submit');
const cvmPreviewFields: FieldList = [
  { name: t('内网IP'), prop: 'private_ip_address' },
  { name: t('地域'), prop: 'region' },
  { name: t('公网IP'), prop: 'public_ip_address' },
  { name: t('所属vpc'), prop: 'cloud_vpc_ids', render: (val: string[]) => val.join(','), copy: true },
  {
    name: t('管控区域'),
    prop: 'bk_cloud_id',
    render: (value: number) =>
      h(DisplayValue, {
        value,
        property: { id: 'bk_cloud_id', name: t('管控区域'), type: 'cloud-area' },
      }),
    copy: false,
  },
  { name: t('主机名称'), prop: 'name' },
  {
    name: t('分配的目标业务'),
    prop: 'bk_biz_id',
    render: (value: number) =>
      h(DisplayValue, { value, property: { id: 'bk_biz_id', name: t('分配的目标业务'), type: 'business' } }),
    copy: false,
  },
  { name: t('实例规格'), prop: 'machine_type' },
  {
    name: t('是否与配置平台关联'),
    prop: 'match_type',
    render: (val: ICvmsAssignBizsPreviewItem['match_type']) => {
      const tagMap: Record<
        ICvmsAssignBizsPreviewItem['match_type'],
        { text: string; theme: 'success' | 'danger' | 'warning' }
      > = {
        no_match: { text: t('待关联'), theme: 'danger' },
        manual: { text: t('手动关联'), theme: 'warning' },
        auto: { text: t('自动关联'), theme: 'success' },
      };

      const { text, theme } = tagMap[val];

      return h(Tag, { theme }, text);
    },
    copy: false,
  },
  { name: t('操作系统'), prop: 'os_name' },
  { name: t('创建时间'), prop: 'created_at', render: (val: string) => timeFormatter(val) },
];

// 表单
const { formModel, resetForm } = useFormModel<{ bk_cloud_id: number; bk_biz_id: number }>({
  bk_cloud_id: undefined,
  bk_biz_id: undefined,
});
const isFormModelHasEmpty = computed(() => formModel.bk_biz_id === undefined || formModel.bk_cloud_id === undefined);
const hasFooter = computed(() => !props.isLoading);
const dialogWidth = computed(() => {
  if (props.isLoading) return '400';
  if (props.cvm?.match_type === 'auto') return '40%';
  return '480';
});
const cloudAreaOption = computed(() =>
  // 暂不支持0管控区
  Object.fromEntries(Array.from(cloudAreaMap.value.entries()).filter(([key]) => key !== 0)),
);
const businessOptionList = computed(() => {
  return businessFullList.value.filter((item) => props.cvm?.bk_biz_ids.includes(item.id));
});

const handleClosed = () => {
  resetForm();
  model.value = false;
};

const handleConfirm = async () => {
  if (props.action === 'backfill') {
    // 批量分配
    emit('backfill', formModel.bk_biz_id, +formModel.bk_cloud_id);
  } else {
    // 单个分配
    if (isPreviewScene.value) {
      const { id: cvm_id, bk_biz_id, bk_cloud_id } = props.cvm;
      emit('submit', { cvm_id, bk_biz_id, bk_cloud_id });
    } else {
      emit('submit', { cvm_id: props.cvm.id, bk_biz_id: formModel.bk_biz_id, bk_cloud_id: +formModel.bk_cloud_id });
    }
  }
  handleClosed();
};

watchEffect(() => {
  formModel.bk_biz_id = props.cvm?.bk_biz_id;
  formModel.bk_cloud_id = props.cvm?.bk_cloud_id;
});

onBeforeMount(() => {
  // 加载管控区域list
  cloudAreaStore.fetchAllCloudAreas();
});
</script>

<template>
  <bk-dialog
    :is-show="model"
    :title="isLoading ? '' : t('分配主机')"
    :dialog-type="hasFooter ? 'operation' : 'show'"
    :width="dialogWidth"
    @closed="handleClosed"
  >
    <template v-if="isLoading">
      <div class="loading-status" style="">
        <bk-loading mode="spin" theme="primary" loading>
          <div style="width: 100%; height: 48px" />
        </bk-loading>
        <div class="loading-text" style="margin-top: 16px; font-size: 20px; color: #313238">{{ t('主机分配中') }}</div>
      </div>
    </template>
    <!-- 根据关联状态，显示不同内容 -->
    <template v-else>
      <div class="info-preview" v-if="isPreviewScene">
        <detail-info :fields="cvmPreviewFields" :detail="props.cvm" label-width="150px" />
      </div>
      <div class="assign-host" v-else>
        <bk-form :model="formModel" form-type="vertical">
          <bk-form-item :label="t('管控区域')" property="bk_cloud_id">
            <hcm-form-enum v-model="formModel.bk_cloud_id" :option="cloudAreaOption" />
          </bk-form-item>
          <bk-form-item :label="t('分配的目标业务')" property="bk_biz_id">
            <hcm-form-business v-model="formModel.bk_biz_id" :data="businessOptionList" />
          </bk-form-item>
        </bk-form>
      </div>
    </template>

    <template v-if="hasFooter" #footer>
      <bk-button
        theme="primary"
        :loading="hostStore.isAssignCvmsToBizsLoading"
        :disabled="isFormModelHasEmpty"
        @click="handleConfirm"
      >
        {{ t('确认分配') }}
      </bk-button>
      <bk-button class="ml8" :disabled="hostStore.isAssignCvmsToBizsLoading" @click="handleClosed">
        {{ t('取消') }}
      </bk-button>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.loading-status {
  margin-top: -16px;
  min-height: 180px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;

  .loading-text {
    position: relative;
    &::after {
      content: '...';
      position: absolute;
    }
  }
}
</style>
