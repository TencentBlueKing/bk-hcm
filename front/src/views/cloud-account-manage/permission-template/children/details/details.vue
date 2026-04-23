<script setup lang="ts">
import { inject, ref, type Ref } from 'vue';
import { SecondaryAccountResourceTypeEnum, VendorEnum } from '@/common/constant';
import type { ModelPropertyDisplay } from '@/model/typings';
import {
  usePermissionTemplateStore,
  type IPermissionTemplateItem,
} from '@/store/cloud-account-manage/permission-template';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import { FieldFactory } from './field-factory';
import type { DetailsFieldTcloud } from './field-tcloud';
import routeAction from '@/router/utils/action';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { LinkPopoverItem } from '@/components/display-value/appearance/link-popover.vue';
import SecondaryAccountValue from '@/views/cloud-account-manage/components/secondary-account-value.vue';
import { Share } from 'bkui-vue/lib/icon';

defineProps<{
  data: IPermissionTemplateItem & DetailsFieldTcloud;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const model = FieldFactory.createModel(currentVendor.value);
const properties = model.getPropertiesByGroup<ModelPropertyDisplay>();

const permissionTemplateStore = usePermissionTemplateStore();
const { getBizsId } = useWhereAmI();

const handleGoToSecondaryAccount = (data: IPermissionTemplateItem) => {
  routeAction.open({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: { type: 'secondary-account', id: data?.account_id },
  });
};
const handleGoToTertiaryAccount = (item: LinkPopoverItem) => {
  routeAction.open({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: { type: 'tertiary-account', id: item.id as string },
  });
};

const getSubAccountLoadFn = (data: IPermissionTemplateItem) => async (): Promise<LinkPopoverItem[]> => {
  const sub_accounts = await permissionTemplateStore.getPermissionTemplateSubAccountIds(
    getBizsId(),
    currentVendor.value,
    data.id,
  );
  return sub_accounts.map(({ id, cloud_id }) => ({ id, label: cloud_id }));
};
</script>

<template>
  <div class="permission-template-details">
    <div v-for="(fields, group) in properties" :key="group" class="details-panel">
      <div class="panel-title">{{ group }}</div>
      <grid-container :column="1" :label-width="120">
        <grid-item v-for="field in fields" :key="field.id" :label="field.name">
          <template v-if="field.id === 'account_id'">
            <div class="link-button-container">
              <SecondaryAccountValue
                :value="data.cloud_account_id"
                :biz-id="getBizsId()"
                :vendor="currentVendor"
                :res-type="SecondaryAccountResourceTypeEnum.TEMPLATE"
              />

              <Share class="icon" @click="handleGoToSecondaryAccount(data)" />
            </div>
            <!-- <display-value
              :property="field"
              :value="data.cloud_account_id"
              :display="{
                on: 'info',
                appearance: 'link-button',
                appearanceProps: { isIcon: true, onClick: () => handleGoToSecondaryAccount(data) },
              }"
            /> -->
          </template>

          <template v-else-if="field.id === 'associated_sub_account_count'">
            <display-value
              :property="field"
              :value="data.associated_sub_account_count"
              :display="{
                appearance: 'link-popover',
                appearanceProps: {
                  loadFn: getSubAccountLoadFn(data),
                  onLinkClick: handleGoToTertiaryAccount,
                  emptyText: '未查询到关联三级账号',
                },
              }"
            />
          </template>

          <display-value
            v-else
            :property="field"
            :value="field.id === 'extension.cloud_type' ? data : data[field.id]"
            :display="{ ...field.meta?.display, on: 'info' }"
          />
        </grid-item>
      </grid-container>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.permission-template-details {
  display: flex;
  flex-direction: column;
  gap: 12px;

  .details-panel {
    background: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 4px 0 #1919290d;
    padding: 16px 24px;

    .panel-title {
      font-size: 14px;
      font-weight: 700;
      color: #313238;
      line-height: 22px;
      margin-bottom: 8px;
    }
  }
}

.link-button-container {
  display: flex;
  align-items: center;
  gap: 6px;

  .icon {
    font-size: 12px;
    color: #3a84ff;
    cursor: pointer;
  }
}
</style>
