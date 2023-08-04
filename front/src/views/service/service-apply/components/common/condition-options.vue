<script setup lang="ts">
import BusinessSelector from '@/components/business-selector/index.vue';
import AccountSelector from '@/components/account-selector/index.vue';
import RegionSelector from './region-selector';
import ResourceGroupSelector from './resource-group-selector';
import { CloudType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { ref, PropType, computed } from 'vue';
import { useAccountStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';

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

</script>

<template>
  <div class="cond-list">
    <div class="cond-item" v-if="!isResourcePage">
      <div class="cond-label">业务</div>
      <div class="cond-content">
        <business-selector
          v-model="selectedBizId"
          :authed="true"
          :auto-select="true">
        </business-selector>
      </div>
    </div>
    <div class="cond-item">
      <div class="cond-label">云账号</div>
      <div class="cond-content">
        <account-selector
          v-model="selectedCloudAccountId"
          :must-biz="!isResourcePage"
          :biz-id="selectedBizId"
          :type="'resource'"
          @change="handleChangeAccount">
        </account-selector>
      </div>
    </div>
    <div class="cond-item">
      <div class="cond-label">云厂商</div>
      <div class="cond-content">
        <bk-select :clearable="false" v-model="selectedVendor">
          <bk-option
            v-for="(item, index) in vendorList"
            :key="index"
            :value="item.id"
            :label="item.name"
          />
        </bk-select>
      </div>
    </div>
    <div class="cond-item" v-if="selectedVendor === VendorEnum.AZURE">
      <div class="cond-label">资源组</div>
      <div class="cond-content">
        <resource-group-selector :account-id="selectedCloudAccountId" v-model="selectedResourceGroup" />
      </div>
    </div>
    <div class="cond-item">
      <div class="cond-label">云地域</div>
      <div class="cond-content">
        <region-selector
          v-model="selectedRegion"
          :type="type"
          :vendor="selectedVendor"
          :account-id="selectedCloudAccountId"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.cond-list {
  display: flex;
  gap: 8px 32px;
  border-bottom: 1px solid rgba(0, 0, 0, .15);
  padding: 0 36px 12px 8px;
  margin-bottom: 8px;
  .cond-item {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 0 12px;
    .cond-label {
      flex: none;
    }
    .cond-content {
      flex: 1;
    }
  }
}
</style>
