<script lang="ts" setup>
import { reactive, watch, ref } from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import { useAccountStore, useResourceStore } from '@/store';
import { BusinessFormFilter } from '@/typings';
import { CLOUD_TYPE } from '@/constants';

const { t } = useI18n();
const accountStore = useAccountStore();
const resourceStore = useResourceStore();
const emit = defineEmits(['change']);
const accountList = ref([]);
const cloudRegionsList = ref([]);
const accountLoading = ref(false);
const cloudRegionsLoading = ref(false);
const cloudAreaPage = ref(0);
const state = reactive<{filter: BusinessFormFilter}>({
  filter: {
    vendor: 'tcloud',
    account_id: '',
    region: '',
  },
});

watch(
  () => state.filter,
  (value) => {
    emit('change', value);
  },
  { deep: true },
);

watch(
  () => state.filter.vendor,
  () => {
    state.filter.region = '';
  },
);

const getAccountList = async () => {
  try {
    accountLoading.value = true;
    const res = await accountStore.getAccountList({
      filter: { op: 'and', rules: [] },
      page: {
        count: false,
        start: 0,
        limit: 500,
      },
    });
    accountList.value = res?.data?.details;
  } catch (error) {
    console.log(error);
  } finally {
    accountLoading.value = false;
  }
};

const getCloudRegionList = () => {
  if (cloudRegionsLoading.value) return;
  cloudRegionsLoading.value = true;
  resourceStore
    .getCloudRegion(state.filter.vendor, {
      filter: { op: 'and', rules: [] },
      page: {
        count: false,
        start: cloudAreaPage.value,
        limit: 100,
      },
    })
    .then((res: any) => {
      cloudAreaPage.value += 1;
      cloudRegionsList.value.push(...res?.data?.details || []);
    })
    .finally(() => {
      cloudRegionsLoading.value = false;
    });
};

// 选择云厂商
const handleCloudChange = () => {
  cloudRegionsList.value = [];
  getCloudRegionList();
};

getAccountList();
handleCloudChange();
</script>
<template>
  <bk-form class="mt20 pt20 bussine-form">
    <bk-form-item
      :label="t('云厂商')"
      class="item-warp"
    >
      <bk-select
        class="item-warp-component"
        v-model="state.filter.vendor"
        @change="handleCloudChange"
      >
        <bk-option
          v-for="(item, index) in CLOUD_TYPE"
          :key="index"
          :value="item.id"
          :label="item.name"
        />
      </bk-select>
    </bk-form-item>
    <bk-form-item
      :label="t('云账号')"
      class="item-warp"
    >
      <bk-select
        class="item-warp-component"
        :loading="accountLoading"
        v-model="state.filter.account_id"
      >
        <bk-option
          v-for="(item, index) in accountList"
          :key="index"
          :value="item.id"
          :label="item.name"
        />
      </bk-select>
    </bk-form-item>
    <bk-form-item
      :label="t('云区域')"
      class="item-warp"
    >
      <bk-select
        class="item-warp-component"
        :disabled="!state.filter.vendor"
        :loading="cloudRegionsLoading"
        v-model="state.filter.region"
      >
        <bk-option
          v-for="(item, index) in cloudRegionsList"
          :key="index"
          :value="item.region_id"
          :label="item.region_name"
        />
      </bk-select>
    </bk-form-item>
  </bk-form>
</template>
<style lang="scss" scoped>
  .bussine-form{
    padding-right: 20px;
  }
</style>
