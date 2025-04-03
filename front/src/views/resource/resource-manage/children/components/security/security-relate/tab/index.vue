<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  ISecurityGroupDetail,
  ISecurityGroupRelResCountItem,
  SecurityGroupRelatedResourceName,
} from '@/store/security-group';
import { VendorEnum } from '@/common/constant';

const props = defineProps<{
  detail: ISecurityGroupDetail;
  relatedResourcesCountList: ISecurityGroupRelResCountItem[];
}>();
const model = defineModel<SecurityGroupRelatedResourceName>();

const { t } = useI18n();

const tabRelRes = ref<{ name: SecurityGroupRelatedResourceName; label: string; count: number; disabled: boolean }[]>([
  { name: SecurityGroupRelatedResourceName.CVM, label: t('云主机'), count: 0, disabled: false },
  { name: SecurityGroupRelatedResourceName.CLB, label: t('负载均衡'), count: 0, disabled: false },
]);
const otherRelRes = ref([]);
const otherRelResCount = computed(() => otherRelRes.value.reduce((prev, curr) => prev + curr.count, 0));
const tabActive = computed({
  get() {
    return (model.value || tabRelRes.value[0].name) as SecurityGroupRelatedResourceName;
  },
  set(value) {
    model.value = value;
  },
});

watchEffect(() => {
  // 腾讯云，展示2个固定的tab：云主机，负载均衡，其他类型平铺展示
  // 其他云，展示1个固定的tab: 云主机，其他类型平铺展示
  props.relatedResourcesCountList[0]?.resources?.forEach(({ res_name, count }) => {
    if (res_name === 'CVM') {
      tabRelRes.value[0].count = count;
    } else if (res_name === 'CLB') {
      tabRelRes.value[1].count = count;
      tabRelRes.value[1].disabled = props.detail.vendor !== VendorEnum.TCLOUD;
    } else {
      otherRelRes.value.push({ name: res_name, count });
    }
  });
});
</script>

<template>
  <bk-radio-group class="tab-wrap" v-model="tabActive">
    <bk-radio-button v-for="{ name, label, count } in tabRelRes" :key="name" :label="name">
      {{ label }}
      <span class="number">{{ count }}</span>
    </bk-radio-button>
  </bk-radio-group>
  <bk-popover theme="light" placement="bottom" :disabled="!otherRelResCount">
    <div class="other-wrap" :class="{ 'is-disabled': !otherRelResCount }">
      {{ t('更多资源') }}
      <span class="number">({{ otherRelResCount }})</span>
    </div>
    <template #content>
      <div class="other-popover-content">
        <div v-for="{ name, count } in otherRelRes" :key="name">
          {{ name }}
          <span class="number">（{{ count }}）</span>
        </div>
      </div>
    </template>
  </bk-popover>
</template>

<style scoped lang="scss">
.tab-wrap {
  margin-right: 16px;
  display: flex;
  align-items: center;
  font-size: 14px;
  color: #4d4f56;

  .number {
    margin-left: 4px;
    padding: 0 8px;
    background: #eaebf0;
    font-size: 12px;
    border-radius: 2px;
    user-select: none;
  }

  :deep(.is-checked) {
    .number {
      color: #fff;
      background: #a3c5fd;
    }
  }
}

.other-wrap {
  margin-right: 16px;
  cursor: pointer;

  &.is-disabled {
    cursor: default;
  }
}

.other-popover-content {
  margin-right: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.other-wrap .number,
.other-popover-content .number {
  font-size: 12px;
  color: #979ba5;
}
</style>
