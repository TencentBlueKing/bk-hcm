<script setup lang="ts">
import BusinessSelector from '@/components/business-selector/index.vue';
import AccountSelector from '@/components/account-selector/index.vue';
import RegionSelector from './region-selector';
import ResourceGroupSelector from './resource-group-selector';
import { CloudType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { ref, PropType, computed, watch } from 'vue';
import { useAccountStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import CommonCard from '@/components/CommonCard';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';

const accountStore = useAccountStore();

const props = defineProps({
  type: String as PropType<string>,
  bizId: Number as PropType<number>,
  cloudAccountId: String as PropType<string>,
  vendor: String as PropType<string>,
  region: String as PropType<string>,
  resourceGroup: String as PropType<string>,
});

const emit = defineEmits([
  'update:bizId',
  'update:cloudAccountId',
  'update:vendor',
  'update:region',
  'update:resourceGroup',
]);

const vendorList = ref([]);

const selectedBizId = computed({
  get() {
    if (!props.bizId) emit('update:bizId', accountStore.bizs);
    return props.bizId || accountStore.bizs;
  },
  set(val) {
    emit('update:bizId', val);

    selectedCloudAccountId.value = '';
    selectedVendor.value = '';
    selectedRegion.value = '';
    vendorList.value = [];
  },
});

const selectedCloudAccountId = computed({
  get() {
    return props.cloudAccountId;
  },
  set(val) {
    emit('update:cloudAccountId', val);

    selectedVendor.value = '';
    selectedRegion.value = '';
  },
});

const selectedVendor = computed({
  get() {
    return props.vendor;
  },
  set(val) {
    emit('update:vendor', val);

    selectedRegion.value = '';
  },
});

const selectedRegion = computed({
  get() {
    return props.region;
  },
  set(val) {
    emit('update:region', val);
  },
});

const selectedResourceGroup = computed({
  get() {
    return props.resourceGroup;
  },
  set(val) {
    emit('update:resourceGroup', val);
  },
});

const handleChangeAccount = (account: any) => {
  vendorList.value = [
    {
      id: account.vendor,
      name: CloudType[account?.vendor],
    },
  ];

  // 默认选中第1个
  selectedVendor.value = vendorList.value?.[0]?.id ?? '';
  selectedRegion.value = '';
};

/**
 * 资源下申请主机、VPC、硬盘时无需选择业务，且无需走审批流程
 */
const { isResourcePage } = useWhereAmI();
const resourceAccountStore = useResourceAccountStore();

watch(
  () => resourceAccountStore.resourceAccount.id,
  (id) => {
    selectedCloudAccountId.value = id;
    handleChangeAccount(resourceAccountStore.resourceAccount);
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <CommonCard
    :title="() => '基本信息'"
    :layout="'grid'"
  >
    <div class="cond-item" v-if="!isResourcePage">
      <div class="mb8">业务</div>
      <div class="cond-content">
        <business-selector
          v-model="selectedBizId"
          :authed="true"
          :auto-select="true">
        </business-selector>
      </div>
    </div>
    <div>
      <div class="mb8">云账号</div>
      <div>
        <account-selector
          v-model="selectedCloudAccountId"
          :must-biz="!isResourcePage"
          :biz-id="selectedBizId"
          :type="'resource'"
          :disabled="!!resourceAccountStore?.resourceAccount?.id"
          @change="handleChangeAccount">
        </account-selector>
      </div>
    </div>
    <div>
      <div class="mb8">云厂商</div>
      <div>
        <bk-select
          :clearable="false"
          v-model="selectedVendor"
          :disabled="!!resourceAccountStore?.resourceAccount?.id"
        >
          <bk-option
            v-for="(item, index) in vendorList"
            :key="index"
            :value="item.id"
            :label="item.name"
          />
        </bk-select>
      </div>
    </div>
    <div v-if="selectedVendor === VendorEnum.AZURE">
      <div class="mb8">资源组</div>
      <div>
        <resource-group-selector :account-id="selectedCloudAccountId" v-model="selectedResourceGroup" />
      </div>
    </div>
    <div>
      <div class="mb8">云地域</div>
      <div>
        <region-selector
          v-model="selectedRegion"
          :type="type"
          :vendor="selectedVendor"
          :account-id="selectedCloudAccountId"
        />
      </div>
    </div>
    <slot />
    <slot name="appendix" />
  </CommonCard>
</template>

<style lang="scss" scoped>
.mb8 {
  margin-bottom: 8px;
}
</style>
