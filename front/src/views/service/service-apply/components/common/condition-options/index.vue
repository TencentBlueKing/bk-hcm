<script setup lang="ts">
import AccountSelector from '@/components/account-selector/index-new.vue';
import RegionSelector from '../region-selector';
import ResourceGroupSelector from '../resource-group-selector';
import { IAccountItem } from '@/typings';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import { PropType, computed, watch } from 'vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import CommonCard from '@/components/CommonCard';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { Form } from 'bkui-vue';
import { accountFilter } from './account-filter.plugin';

const { FormItem } = Form;

const props = defineProps({
  type: String as PropType<string>,
  bizs: Number as PropType<number>,
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

const selectedCloudAccountId = computed({
  get() {
    return props.cloudAccountId;
  },
  set(val) {
    val && emit('update:cloudAccountId', val);

    selectedVendor.value = '';
    selectedRegion.value = '';
  },
});

const selectedVendor = computed({
  get() {
    return props.vendor;
  },
  set(val) {
    val && emit('update:vendor', val);

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
    val && emit('update:resourceGroup', val);
  },
});

const handleChangeAccount = (account: IAccountItem) => {
  selectedVendor.value = account?.vendor ?? '';
  selectedRegion.value = '';
};

/**
 * 资源下申请主机、VPC、硬盘时无需选择业务，且无需走审批流程
 */
const { isResourcePage } = useWhereAmI();
const resourceAccountStore = useResourceAccountStore();

watch(
  () => resourceAccountStore.resourceAccount?.id,
  (id) => {
    selectedCloudAccountId.value = id;
    handleChangeAccount(resourceAccountStore.resourceAccount as IAccountItem);
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <CommonCard class="mb16" :title="() => '基本信息'" :layout="'grid'">
    <FormItem
      label="云账号"
      required
      :property="[ResourceTypeEnum.SUBNET, ResourceTypeEnum.CLB].includes(type) ? 'account_id' : 'cloudAccountId'"
    >
      <account-selector
        v-model="selectedCloudAccountId"
        :biz-id="isResourcePage ? undefined : props.bizs"
        :filter="accountFilter"
        :disabled="isResourcePage"
        :placeholder="isResourcePage ? '请在左侧选择账号' : undefined"
        @change="handleChangeAccount"
      />
    </FormItem>
    <FormItem
      label="资源组"
      required
      :property="type === ResourceTypeEnum.SUBNET ? 'resource_group' : 'resourceGroup'"
      v-if="selectedVendor === VendorEnum.AZURE"
    >
      <resource-group-selector :account-id="selectedCloudAccountId" v-model="selectedResourceGroup" />
    </FormItem>
    <FormItem label="云地域" required property="region">
      <region-selector
        v-model="selectedRegion"
        :type="type"
        :vendor="selectedVendor"
        :account-id="selectedCloudAccountId"
      />
    </FormItem>
    <slot />
    <slot name="appendix" />
  </CommonCard>
</template>

<style lang="scss" scoped>
.mb8 {
  margin-bottom: 8px;
}
</style>
