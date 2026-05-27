<script setup lang="ts">
import { computed, inject, ref, type Ref } from 'vue';
import { VendorEnum, SecondaryAccountResourceTypeEnum } from '@/common/constant';
import type { ModelPropertyDisplay } from '@/model/typings';
import type { IPermissionPolicyItem } from '../../typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import {
  AUTH_APPLY_PERMISSION_POLICY_LIBRARY,
  AUTH_BIZ_APPLY_PERMISSION_POLICY_LIBRARY,
} from '@/constants/auth-symbols';
import { getAuthSignByBusinessId } from '@/utils';
import router from '@/router';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';
import { useRoute } from 'vue-router';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import { FieldFactory } from './field-factory';
import SecondaryAccountValue from '@/views/cloud-account-manage/components/secondary-account-value.vue';
import type { LinkPopoverItem } from '@/components/display-value/appearance/link-popover.vue';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

const props = withDefaults(
  defineProps<{
    policyData?: IPermissionPolicyItem;
  }>(),
  {
    policyData: undefined,
  },
);

const emit = defineEmits<{
  'apply-to-account': [row: IPermissionPolicyItem];
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const modelInstance = FieldFactory.createModel(currentVendor.value);
const properties = modelInstance.getPropertiesByGroup<ModelPropertyDisplay>();

const { isBusinessPage, getBizsId } = useWhereAmI();
const route = useRoute();

const bizId = computed(() => (isBusinessPage ? getBizsId() : 0));

// 跳转二级账号详情（新开标签页）
const handleGoToAccount = (item: LinkPopoverItem) => {
  router.push({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: {
      ...route.query,
      type: 'secondary-account',
      id: item.id,
    },
  });
};

// 应用到二级账号
const handleApplyToAccount = () => {
  emit('apply-to-account', props.policyData);
};
</script>

<template>
  <bk-sideslider
    v-model:is-show="model"
    title="权限策略库详情"
    render-directive="if"
    width="640"
    quick-close
    background-color="#f5f7fa"
  >
    <template #header>
      <div class="permission-policy-header">
        <div class="title">
          权限策略库详情
          <div class="separator"></div>
          <span class="name">{{ props.policyData?.name }}</span>
        </div>
        <hcm-auth
          :sign="
            getAuthSignByBusinessId(
              bizId,
              AUTH_APPLY_PERMISSION_POLICY_LIBRARY,
              AUTH_BIZ_APPLY_PERMISSION_POLICY_LIBRARY,
            )
          "
          v-slot="{ noPerm }"
        >
          <bk-button theme="primary" :disabled="noPerm" @click="handleApplyToAccount" outline>应用到二级账号</bk-button>
        </hcm-auth>
      </div>
    </template>
    <template #default>
      <div class="permission-policy-info">
        <div v-for="(fields, group) in properties" :key="group" class="details-panel">
          <div class="panel-title">{{ group }}</div>
          <grid-container :column="1" :label-width="125">
            <grid-item v-for="field in fields" :key="field.id" :label="field.name">
              <template v-if="field.id === 'associated_account_count'">
                <display-value
                  :property="field"
                  :value="props.policyData?.associated_account_count"
                  :display="{
                    appearance: 'link-popover',
                    appearanceProps: {
                      onLinkClick: handleGoToAccount,
                      emptyText: '未查询到关联二级账号',
                      list: props.policyData?.related_accounts?.map((id: string) => ({ id, label: id })),
                    },
                  }"
                >
                  <template #item-label="{ item }">
                    <SecondaryAccountValue
                      :value="item.id"
                      :vendor="currentVendor"
                      :res-type="SecondaryAccountResourceTypeEnum.PERMISSION"
                      :biz-id="bizId"
                      :label-formatter="(item) => item?.extension?.cloud_main_account_id"
                    />
                  </template>
                </display-value>
              </template>

              <display-value
                v-else
                :property="field"
                :value="props.policyData?.[field.id]"
                :display="{ ...field.meta?.display, on: 'info' }"
              />
            </grid-item>
          </grid-container>
        </div>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
:deep(.permission-policy-header) {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 24px;

  .title {
    font-size: 16px;
    color: #313238;
    display: flex;
    align-items: center;

    .separator {
      margin: 0 8px;
      width: 1px;
      height: 12px;
      background-color: #979ba5;
    }

    .name {
      font-size: 14px;
      color: #979ba5;
    }
  }
}

.permission-policy-info {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  &.apply {
    padding-bottom: 150px;
  }

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
      margin-bottom: 14px;
    }

    :deep(.grid-container) {
      .grid-item {
        align-items: center;

        .item-label,
        .item-content {
          font-size: 12px;
          padding: 7px 0;
        }

        .item-label {
          color: #4d4f56;
        }
      }
    }

    .relate-account-count {
      .num {
        color: #3a84ff;
        cursor: pointer;
      }
    }
  }
}
</style>
