<script setup lang="ts">
import { ref, inject, computed, type Ref, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useCloudSecretStore } from '@/store/cloud-account-manage/cloud-secret';
import type { ISubAccountItem } from '@/store/cloud-account-manage/tertiary-account';
import { useAccountStore } from '@/store';
import { VendorEnum } from '@/common/constant';
import {
  AUTH_BIZ_CREATE_SUB_ACCOUNT_SECRET,
  AUTH_BIZ_UPDATE_SUB_ACCOUNT_SECRET,
  AUTH_BIZ_DELETE_SUB_ACCOUNT_SECRET,
  AUTH_BIZ_UPDATE_SUB_ACCOUNT,
  AUTH_BIZ_DELETE_SUB_ACCOUNT,
} from '@/constants/auth-symbols';
import { FLAG_OPTIONS, ACCOUNT_TYPE_OPTIONS } from '../../constants';
import Status from '@/components/display-value/appearance/status.vue';
import BusinessValue from '@/components/display-value/business-value.vue';
import DatetimeValue from '@/components/display-value/datetime-value.vue';
import SecretActionDialog from '@/views/cloud-account-manage/cloud-secret/children/secret-action-dialog/index.vue';
import type { ICloudSecretItem, SecretActionType } from '@/views/cloud-account-manage/cloud-secret/typings';

const model = defineModel<boolean>();

const props = defineProps<{
  rowData: ISubAccountItem | null;
}>();

const emit = defineEmits<{
  (e: 'update-success'): void;
  (e: 'edit' | 'delete', row: ISubAccountItem): void;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const cloudSecretStore = useCloudSecretStore();
const accountStore = useAccountStore();
const { getBizsId } = useWhereAmI();

const secretList = ref<ICloudSecretItem[]>([]);
const secretLoading = ref(false);
const showKeyLoading = ref(false);
const showKeyResult = ref(false);
const newSecretId = ref('');
const newSecretKey = ref('');
const keyAcknowledged = ref(false);

// 密钥操作二次确认弹窗状态
const showSecretActionDialog = ref(false);
const secretActionType = ref<SecretActionType>('enable');
const currentSecretData = ref<ICloudSecretItem | null>(null);

const loadSecretList = async () => {
  if (!props.rowData?.id) return;
  secretLoading.value = true;
  try {
    const result = await cloudSecretStore.getSubAccountSecretList(getBizsId(), currentVendor.value, {
      sub_account_ids: [props.rowData.id],
      page: { count: false, start: 0, limit: 500 },
    });
    secretList.value = result.list;
  } catch (error) {
    console.error('加载密钥列表失败:', error);
  } finally {
    secretLoading.value = false;
  }
};

watch(
  () => model.value,
  (val) => {
    if (val && props.rowData) {
      loadSecretList();
    }
  },
);

const handleClose = () => {
  model.value = false;
};

const handleEdit = () => {
  if (props.rowData) {
    emit('edit', props.rowData);
  }
};

const handleDelete = () => {
  if (props.rowData) {
    emit('delete', props.rowData);
  }
};

const maskSecretId = (id: string) => {
  if (!id) return '****';
  if (id.length <= 8) return id;
  return `${id.substring(0, 4)}****${id.substring(id.length - 4)}`;
};

const isProgramAccount = computed(() => {
  return props.rowData?.extension?.console_login === 0;
});

const getLoginFlagText = (flag?: string) => {
  if (!flag) return '--';
  return FLAG_OPTIONS[flag as keyof typeof FLAG_OPTIONS] || '--';
};

const getActionFlagText = (flag?: string) => {
  if (!flag) return '--';
  return FLAG_OPTIONS[flag as keyof typeof FLAG_OPTIONS] || '--';
};

const getMfaStatus = () => {
  const ext = props.rowData?.extension;
  if (!ext) return '未绑定';
  return ext.login_flag === 'stoken' || ext.action_flag === 'stoken' ? '已绑定' : '未绑定';
};

const maskPhone = (phone?: string) => {
  if (!phone) return '--';
  if (phone.length < 7) return phone;
  return `${phone.substring(0, 3)}****${phone.substring(phone.length - 4)}`;
};

const handleCreateSecret = async () => {
  if (!props.rowData?.id) return;
  if (!isProgramAccount.value) {
    Message({ theme: 'warning', message: '仅编程账号允许创建密钥' });
    return;
  }

  showKeyLoading.value = true;
  try {
    const res = await cloudSecretStore.createSubAccountSecret(getBizsId(), currentVendor.value, props.rowData.id);
    showKeyLoading.value = false;
    newSecretId.value = res?.extension?.cloud_secret_id || '--';
    newSecretKey.value = res?.extension?.cloud_secret_key || '--';
    showKeyResult.value = true;
    keyAcknowledged.value = false;
    await loadSecretList();
    emit('update-success');
  } catch (error) {
    showKeyLoading.value = false;
    console.error('创建密钥失败:', error);
  }
};

const handleToggleSecretStatus = (secret: ICloudSecretItem) => {
  secretActionType.value = secret.status === 'enabled' ? 'disable' : 'enable';
  currentSecretData.value = secret;
  showSecretActionDialog.value = true;
};

const handleDeleteSecret = (secret: ICloudSecretItem) => {
  secretActionType.value = 'delete';
  currentSecretData.value = secret;
  showSecretActionDialog.value = true;
};

const handleSecretActionSuccess = () => {
  showSecretActionDialog.value = false;
  loadSecretList();
};

const handleCopySecret = () => {
  const text = `密钥ID: ${newSecretId.value}\n密钥Key: ${newSecretKey.value}`;
  navigator.clipboard.writeText(text).then(() => {
    Message({ theme: 'success', message: '密钥已复制到剪贴板' });
  });
};

const handleDownloadCSV = () => {
  const csvContent = `密钥ID,密钥Key\n${newSecretId.value},${newSecretKey.value}`;
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = 'secret_key.csv';
  link.click();
};

const handleCloseKeyResult = () => {
  showKeyResult.value = false;
  newSecretId.value = '';
  newSecretKey.value = '';
  keyAcknowledged.value = false;
};

const formatTime = (time?: string) => {
  if (!time) return '--';
  return time.replace('T', ' ').replace('Z', '');
};
</script>

<template>
  <bk-sideslider
    :is-show="model"
    :width="960"
    title="三级账号详情"
    :before-close="handleClose"
    @closed="handleClose"
    ext-cls="detail-sideslider"
  >
    <template #header>
      <div class="detail-header">
        <span class="title">三级账号详情</span>
        <div class="header-actions">
          <hcm-auth
            v-if="accountStore.bizs"
            :sign="{ type: AUTH_BIZ_DELETE_SUB_ACCOUNT, relation: [accountStore.bizs] }"
            v-slot="{ noPerm }"
          >
            <bk-button :disabled="noPerm || rowData?.operable === false" @click="handleDelete">删除</bk-button>
          </hcm-auth>
        </div>
      </div>
    </template>
    <template #default>
      <div v-if="rowData" class="detail-content">
        <div class="info-card">
          <div class="card-header">
            <span class="card-title">基本信息</span>
            <hcm-auth
              v-if="accountStore.bizs"
              :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT, relation: [accountStore.bizs] }"
              v-slot="{ noPerm }"
            >
              <bk-button theme="primary" text :disabled="noPerm || rowData.operable === false" @click="handleEdit">
                编辑
              </bk-button>
            </hcm-auth>
          </div>
          <div class="card-body info-grid">
            <div class="info-item">
              <span class="info-label">云厂商：</span>
              <span class="info-value">腾讯云</span>
            </div>
            <div class="info-item">
              <span class="info-label">账号类型：</span>
              <span class="info-value">{{ ACCOUNT_TYPE_OPTIONS[rowData.extension?.console_login] || '--' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">三级账号名称：</span>
              <span class="info-value">{{ rowData.name || '--' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">三级账号ID：</span>
              <span class="info-value">{{ rowData.cloud_id || '--' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">所属业务：</span>
              <BusinessValue :value="rowData.bk_biz_ids" />
            </div>
            <div class="info-item">
              <span class="info-label">账号邮箱：</span>
              <span class="info-value">{{ rowData.email || '--' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">创建时间：</span>
              <DatetimeValue :value="rowData.cloud_created_at" />
            </div>
            <div class="info-item">
              <span class="info-label">手机号：</span>
              <span class="info-value">{{ maskPhone(rowData.phone_num) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">更新时间：</span>
              <DatetimeValue :value="rowData.updated_at" />
            </div>
            <div class="info-item">
              <span class="info-label">负责人：</span>
              <span class="info-value">{{ rowData.managers?.join(', ') || '--' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">备注：</span>
              <span class="info-value">{{ rowData.memo || '--' }}</span>
            </div>
          </div>
        </div>

        <div class="info-card">
          <div class="card-header">
            <span class="card-title">安全信息</span>
          </div>
          <div class="card-body info-grid">
            <div class="info-item">
              <span class="info-label">登录保护：</span>
              <span class="info-value">{{ getLoginFlagText(rowData.extension?.login_flag) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">操作保护：</span>
              <span class="info-value">{{ getActionFlagText(rowData.extension?.action_flag) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">MFA设备绑定：</span>
              <span class="info-value">{{ getMfaStatus() }}</span>
            </div>
          </div>
        </div>

        <div class="info-card">
          <div class="card-header">
            <span class="card-title">权限模板</span>
          </div>
          <div class="card-body">
            <div class="info-item">
              <span class="info-label">权限模板：</span>
              <span class="info-value permission-template-tags" v-if="rowData?.permission_templates?.length">
                <bk-tag v-for="(template, index) in rowData.permission_templates" :key="index">
                  {{ template.name }}
                </bk-tag>
              </span>
              <span class="info-value" v-else>--</span>
            </div>
          </div>
        </div>

        <div class="info-card">
          <div class="card-header">
            <div class="card-header-left">
              <span class="card-title">API密钥</span>
              <hcm-auth
                v-if="accountStore.bizs"
                :sign="{ type: AUTH_BIZ_CREATE_SUB_ACCOUNT_SECRET, relation: [accountStore.bizs] }"
                v-slot="{ noPerm }"
              >
                <bk-button
                  text
                  theme="primary"
                  class="create-secret-btn"
                  @click="handleCreateSecret"
                  :disabled="!isProgramAccount || noPerm || rowData?.operable === false"
                >
                  <i class="hcm-icon bkhcm-icon-plus-circle-shape"></i>
                  新建密钥
                </bk-button>
              </hcm-auth>
            </div>
          </div>
          <div class="card-body">
            <bk-loading :loading="secretLoading">
              <bk-table :data="secretList" row-key="id" :border="['row', 'outer']" show-overflow-tooltip>
                <bk-table-column label="密钥ID" min-width="130">
                  <template #default="{ row }">
                    {{ maskSecretId(row.extension?.cloud_secret_id || row.cloud_secret_id) }}
                  </template>
                </bk-table-column>
                <bk-table-column label="密钥Key" width="80">
                  <template #default>****</template>
                </bk-table-column>
                <bk-table-column label="密钥状态" min-width="100">
                  <template #default="{ row }">
                    <Status
                      :value="row.status === 'enabled' ? 'normal' : 'unknown'"
                      :display-value="row.status === 'enabled' ? '已启用' : '已禁用'"
                    />
                  </template>
                </bk-table-column>
                <bk-table-column label="创建时间" min-width="160">
                  <template #default="{ row }">
                    {{ formatTime(row.cloud_created_at) }}
                  </template>
                </bk-table-column>
                <bk-table-column label="最近访问时间" min-width="160">
                  <template #default="{ row }">
                    {{ formatTime(row.last_used_time) || '--' }}
                  </template>
                </bk-table-column>
                <bk-table-column label="禁用时间" min-width="160">
                  <template #default="{ row }">
                    {{ row.status === 'disabled' ? formatTime(row.disabled_time) : '--' }}
                  </template>
                </bk-table-column>
                <bk-table-column label="操作" width="100" fixed="right">
                  <template #default="{ row }">
                    <template v-if="accountStore.bizs">
                      <template v-if="row.status === 'enabled'">
                        <hcm-auth
                          :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT_SECRET, relation: [accountStore.bizs] }"
                          v-slot="{ noPerm }"
                        >
                          <bk-button
                            text
                            theme="primary"
                            :disabled="noPerm || row.operable === false"
                            @click="handleToggleSecretStatus(row)"
                          >
                            禁用
                          </bk-button>
                        </hcm-auth>
                      </template>
                      <template v-else>
                        <hcm-auth
                          :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT_SECRET, relation: [accountStore.bizs] }"
                          v-slot="{ noPerm }"
                        >
                          <bk-button
                            text
                            theme="primary"
                            :disabled="noPerm || row.operable === false"
                            @click="handleToggleSecretStatus(row)"
                          >
                            启用
                          </bk-button>
                        </hcm-auth>
                        <hcm-auth
                          :sign="{ type: AUTH_BIZ_DELETE_SUB_ACCOUNT_SECRET, relation: [accountStore.bizs] }"
                          v-slot="{ noPerm }"
                        >
                          <bk-button
                            text
                            theme="primary"
                            class="ml8"
                            :disabled="noPerm || row.operable === false"
                            @click="handleDeleteSecret(row)"
                          >
                            删除
                          </bk-button>
                        </hcm-auth>
                      </template>
                    </template>
                  </template>
                </bk-table-column>
              </bk-table>
            </bk-loading>
          </div>
        </div>
      </div>
    </template>
  </bk-sideslider>

  <bk-dialog
    :is-show="showKeyLoading"
    :show-footer="false"
    :close-icon="true"
    width="400"
    @closed="showKeyLoading = false"
  >
    <div class="key-loading-content">
      <bk-loading :loading="true" size="large" />
      <p class="loading-title">正在生成密钥</p>
      <p class="loading-subtitle">请耐心等待...</p>
    </div>
  </bk-dialog>

  <bk-dialog
    :is-show="showKeyResult"
    :show-footer="false"
    :close-icon="true"
    width="600"
    title="新建密钥"
    @closed="handleCloseKeyResult"
  >
    <div class="key-result-content">
      <div class="warning-box">
        <i class="bk-icon icon-exclamation-circle-shape warning-icon" />
        <span>新建的密钥，仅在创建时提供下载和复制，后续只可查询密钥ID，不可查询密钥key。请妥善保管密钥。</span>
      </div>
      <div class="key-info">
        <div class="key-item">
          <span class="key-label">密钥ID：</span>
          <span class="key-value">{{ newSecretId }}</span>
        </div>
        <div class="key-item">
          <span class="key-label">密钥Key：</span>
          <span class="key-value">{{ newSecretKey }}</span>
        </div>
      </div>
      <div class="key-actions">
        <bk-button theme="primary" @click="handleCopySecret">
          <i class="bk-icon icon-copy" style="margin-right: 4px" />
          复制密钥
        </bk-button>
        <bk-button @click="handleDownloadCSV">
          <i class="bk-icon icon-download" style="margin-right: 4px" />
          下载CSV文件
        </bk-button>
      </div>
      <div class="key-confirm">
        <bk-checkbox v-model="keyAcknowledged">我已知晓并保存密钥Key</bk-checkbox>
        <bk-button :disabled="!keyAcknowledged" @click="handleCloseKeyResult">关闭</bk-button>
      </div>
    </div>
  </bk-dialog>

  <!-- 密钥操作二次确认弹窗（启用/禁用/删除） -->
  <SecretActionDialog
    v-model="showSecretActionDialog"
    :action-type="secretActionType"
    :secret-data="currentSecretData"
    :vendor="currentVendor"
    @success="handleSecretActionSuccess"
  />
</template>

<style lang="scss" scoped>
.detail-sideslider {
  :deep(.bk-modal-body) {
    background: #f5f7fa;
  }
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 16px;

  .title {
    font-size: 16px;
    color: #313238;
  }
}

.detail-content {
  padding: 20px 24px;
}

.info-card {
  background: #fff;
  border-radius: 2px;
  margin-bottom: 16px;
  padding: 16px 24px 24px;

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;

    .card-title {
      font-size: 14px;
      font-weight: 600;
      color: #313238;
    }

    .card-header-right {
      display: flex;
      align-items: center;
      gap: 8px;

      .secret-count {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-width: 20px;
        height: 20px;
        padding: 0 6px;
        border-radius: 10px;
        background: #3a84ff;
        color: #fff;
        font-size: 12px;
      }
    }

    .card-header-left {
      display: flex;
      align-items: center;
      gap: 8px;

      .create-secret-btn {
        display: inline-flex;
        align-items: center;
        font-size: 12px;

        .hcm-icon {
          font-size: 14px;
          margin-right: 5px;
          margin-left: 18px;
        }
      }
    }
  }

  .card-body {
    &.info-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 12px 20px;
      padding-left: 10px;
    }
  }
}

.info-item {
  display: flex;
  align-items: flex-start;
  font-size: 12px;
  line-height: 20px;

  .info-label {
    color: #4d4f56;
    white-space: nowrap;
    min-width: 90px;
    text-align: right;
    margin-right: 8px;
  }

  .info-value {
    color: #313238;
    word-break: break-all;
  }

  .permission-template-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
}

.key-loading-content {
  text-align: center;
  padding: 40px 0;

  .loading-title {
    font-size: 16px;
    font-weight: 600;
    color: #313238;
    margin-top: 24px;
  }

  .loading-subtitle {
    font-size: 14px;
    color: #979ba5;
    margin-top: 8px;
  }
}

.key-result-content {
  .warning-box {
    display: flex;
    align-items: flex-start;
    padding: 12px 16px;
    background: #fff8e6;
    border: 1px solid #ffe8c3;
    border-radius: 2px;
    margin-bottom: 16px;
    font-size: 12px;
    color: #63656e;
    line-height: 20px;

    .warning-icon {
      color: #ff9c01;
      font-size: 16px;
      margin-right: 8px;
      flex-shrink: 0;
      margin-top: 2px;
    }
  }

  .key-info {
    padding: 16px;
    background: #f5f7fa;
    border-radius: 2px;
    margin-bottom: 16px;

    .key-item {
      display: flex;
      align-items: center;
      font-size: 14px;
      line-height: 28px;

      .key-label {
        color: #313238;
        font-weight: 600;
        white-space: nowrap;
      }

      .key-value {
        color: #313238;
        word-break: break-all;
      }
    }
  }

  .key-actions {
    display: flex;
    gap: 8px;
    margin-bottom: 24px;
  }

  .key-confirm {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 16px;
    border-top: 1px solid #dcdee5;
  }
}

.svg-icon {
  width: 14px;
  height: 14px;
  vertical-align: middle;

  /* 不要设置 fill 属性 */
}
</style>
