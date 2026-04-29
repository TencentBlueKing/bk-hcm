<script setup lang="ts">
import { ref, computed } from 'vue';
import { Share } from 'bkui-vue/lib/icon';
import { AUTH_BIZ_UPDATE_SUB_ACCOUNT_SECRET } from '@/constants/auth-symbols';
import { useAccountStore } from '@/store';
import { SECRET_STATUS_MAP } from '../../constants';
import Status from '@/components/display-value/appearance/status.vue';
import type { ICloudSecretItem, SecretActionType } from '../../typings';
import SecretActionDialog from '../secret-action-dialog/index.vue';
import ArrayValue from '@/components/display-value/array-value.vue';
import DatetimeValue from '@/components/display-value/datetime-value.vue';
import routeAction from '@/router/utils/action';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';

const model = defineModel<boolean>();

const props = defineProps<{
  secretData: ICloudSecretItem | null;
  vendor: string;
}>();

const emit = defineEmits<{
  'action-success': [];
}>();

const accountStore = useAccountStore();

const showActionDialog = ref(false);
const currentActionType = ref<SecretActionType>('disable');

const subAccountId = computed(
  () => props.secretData?.cloud_sub_account_id || props.secretData?.extension?.cloud_sub_account_id,
);
const mainAccountId = computed(
  () => props.secretData?.cloud_main_account_id || props.secretData?.extension?.cloud_main_account_id,
);

const handleGoToTertiaryAccount = () => {
  routeAction.open({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: { type: 'tertiary-account', id: props.secretData?.sub_account_id },
  });
};

const handleGoToSecondaryAccount = () => {
  routeAction.open({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: { type: 'secondary-account', id: props.secretData?.account_id },
  });
};

const statusConfig = computed(() => {
  if (!props.secretData) return null;
  return SECRET_STATUS_MAP[props.secretData.status];
});

const canDisable = computed(() => props.secretData?.status === 'enabled');
const canEnable = computed(() => props.secretData?.status === 'disabled');

const handleToggleStatus = () => {
  currentActionType.value = canDisable.value ? 'disable' : 'enable';
  showActionDialog.value = true;
};

const handleActionSuccess = () => {
  emit('action-success');
  model.value = false;
};
</script>

<template>
  <bk-sideslider v-model:is-show="model" title="云密钥详情" :width="640" quick-close :before-close="() => true">
    <template #header>
      <div class="slider-header">
        <span class="title">云密钥详情</span>
        <div class="header-actions">
          <template v-if="accountStore.bizs">
            <hcm-auth
              v-if="canDisable"
              :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT_SECRET, relation: [accountStore.bizs] }"
              v-slot="{ noPerm }"
            >
              <bk-button
                theme="primary"
                outline
                :disabled="noPerm || secretData?.operable === false"
                @click="handleToggleStatus"
              >
                禁用
              </bk-button>
            </hcm-auth>
            <hcm-auth
              v-if="canEnable"
              :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT_SECRET, relation: [accountStore.bizs] }"
              v-slot="{ noPerm }"
            >
              <bk-button
                theme="primary"
                outline
                :disabled="noPerm || secretData?.operable === false"
                @click="handleToggleStatus"
              >
                启用
              </bk-button>
            </hcm-auth>
          </template>
        </div>
      </div>
    </template>

    <div class="secret-detail-container">
      <div class="detail-content">
        <div class="detail-item">
          <span class="label">云密钥ID：</span>
          <span class="value">{{ secretData?.cloud_secret_id || secretData?.extension?.cloud_secret_id || '--' }}</span>
        </div>

        <div class="detail-item">
          <span class="label">云密钥状态：</span>
          <span class="value status-value" v-if="statusConfig">
            <Status :value="statusConfig.iconName" :display-value="statusConfig.text" />
          </span>
          <span class="value" v-else>--</span>
        </div>

        <div class="detail-item">
          <span class="label">所属三级账号：</span>
          <span class="value link-value">
            <template v-if="subAccountId">
              {{ secretData?.sub_account_name ? `${secretData.sub_account_name}（${subAccountId}）` : subAccountId }}
              <Share class="icon-link" @click="handleGoToTertiaryAccount" />
            </template>
            <template v-else>--</template>
          </span>
        </div>

        <div class="detail-item">
          <span class="label">三级账号负责人：</span>
          <span class="value">
            <ArrayValue :value="secretData?.sub_account_managers" :display="{ showOverflowTooltip: true }" />
          </span>
        </div>

        <div class="detail-item">
          <span class="label">所属二级账号：</span>
          <span class="value link-value">
            <template v-if="mainAccountId">
              {{ secretData?.account_name ? `${secretData.account_name}（${mainAccountId}）` : mainAccountId }}
              <Share class="icon-link" @click="handleGoToSecondaryAccount" />
            </template>
            <template v-else>--</template>
          </span>
        </div>

        <div class="detail-item">
          <span class="label">二级账号负责人：</span>
          <span class="value">
            <ArrayValue :value="secretData?.account_managers || []" :display="{ showOverflowTooltip: true }" />
          </span>
        </div>

        <div class="detail-item">
          <span class="label">创建时间：</span>
          <span class="value"><DatetimeValue :value="secretData?.cloud_created_at" /></span>
        </div>

        <div class="detail-item">
          <span class="label">更新时间：</span>
          <span class="value"><DatetimeValue :value="secretData?.updated_at" /></span>
        </div>

        <div class="detail-item">
          <span class="label">最近访问时间：</span>
          <span class="value"><DatetimeValue :value="secretData?.last_used_time" /></span>
        </div>
      </div>
    </div>

    <SecretActionDialog
      v-model="showActionDialog"
      :action-type="currentActionType"
      :secret-data="secretData"
      :vendor="vendor"
      @success="handleActionSuccess"
    />
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.slider-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 40px;

  .title {
    font-size: 16px;
    color: #313238;
  }

  .header-actions {
    display: flex;
    gap: 8px;
  }
}

.secret-detail-container {
  padding: 24px 52px;

  .detail-content {
    .detail-item {
      display: flex;
      align-items: center;
      font-size: 12px;
      line-height: 32px;
      padding: 2px 0;

      .label {
        color: #4d4f56;
        min-width: 96px;
        text-align: right;
        flex-shrink: 0;
      }

      .value {
        color: #313238;
        margin-left: 8px;

        &.status-value {
          display: flex;
          align-items: center;
        }

        &.link-value {
          display: flex;
          align-items: center;

          .icon-link {
            cursor: pointer;
            color: #3a84ff;
            margin-left: 8px;
            font-size: 14px;
          }
        }
      }
    }
  }
}
</style>
