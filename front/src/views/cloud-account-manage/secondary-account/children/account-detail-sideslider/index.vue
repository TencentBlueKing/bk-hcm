<script setup lang="ts">
import { ref, computed, watch, h } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Message, InfoBox } from 'bkui-vue';
import { Share } from 'bkui-vue/lib/icon';
import type { ISecondaryAccountItem, IAccountSecretItem } from '@/store/cloud-account-manage/secondary-account';
import { useSecondaryAccountStore } from '@/store/cloud-account-manage/secondary-account';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import DisplayValue from '@/components/display-value/index.vue';
import SecretKeyDialog from './secret-key-dialog.vue';
import AccountFormSideslider from '../account-form-sideslider/index.vue';
import AccountCreateSideslider from '@/views/cloud-account-manage/tertiary-account/children/account-create-sideslider/index.vue';
import type { ModelProperty } from '@/model/typings';
import type { DisplayType } from '@/components/display-value/typings';
import {
  AUTH_BIZ_UPDATE_SECONDARY_ACCOUNT,
  AUTH_BIZ_CREATE_SECONDARY_ACCOUNT,
  AUTH_BIZ_DELETE_SECONDARY_ACCOUNT,
} from '@/constants/auth-symbols';
import { useAccountStore } from '@/store/account';
import { VendorMap } from '@/common/constant';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

// Props 定义
const props = defineProps<{
  rowData: ISecondaryAccountItem | null;
}>();

// Emits 定义
const emit = defineEmits<{
  'update-success': [];
  'edit-account': [row: ISecondaryAccountItem];
  'update:rowData': [data: ISecondaryAccountItem | null];
}>();

// Store 和 Hooks
const secondaryAccountStore = useSecondaryAccountStore();
const accountStore = useAccountStore();
const { getBizsId } = useWhereAmI();
const route = useRoute();
const router = useRouter();

// 密钥列表数据
const secretList = ref<IAccountSecretItem[]>([]);
const secretLoading = ref(false);

// 录入/编辑密钥弹窗状态
const showSecretDialog = ref(false);
const editingSecret = ref<IAccountSecretItem | null>(null);
const isEditMode = ref(false);

// 编辑账号弹窗状态
const showAccountFormSideslider = ref(false);

// 新建三级账号弹窗状态
const showCreateSubAccount = ref(false);

// 当前展示的账号数据（用于支持编辑后实时更新）
const currentRowData = ref<ISecondaryAccountItem | null>(null);

// 监听 props.rowData 变化，同步到本地
watch(
  () => props.rowData,
  (newVal) => {
    currentRowData.value = newVal ? { ...newVal } : null;
  },
  { immediate: true, deep: true },
);

// 获取密钥列表
const fetchSecretList = async () => {
  if (!currentRowData.value) return;

  try {
    secretLoading.value = true;
    const { list } = await secondaryAccountStore.getAccountSecretList(getBizsId(), currentRowData.value.vendor, {
      filter: {
        op: 'and',
        rules: [{ field: 'account_id', op: 'eq', value: currentRowData.value.id }],
      },
      page: { count: false, start: 0, limit: 500 },
    });
    secretList.value = list;
  } catch (error) {
    console.error('获取密钥列表失败:', error);
    secretList.value = [];
  } finally {
    secretLoading.value = false;
  }
};

// 监听显示状态和数据变化，确保用最新数据加载密钥列表
watch([model, () => props.rowData], ([isShow, rowData]) => {
  if (isShow && rowData) {
    // 先同步更新 currentRowData，再发起请求，避免时序错位
    currentRowData.value = { ...rowData };
    fetchSecretList();
  } else if (!isShow) {
    // 退出详情时清空资源密钥列表
    secretList.value = [];
  }
});

// 基本信息字段配置
const baseInfoFields = computed<Array<{ label: string; value: any; property: ModelProperty; display?: DisplayType }>>(
  () => [
    {
      label: '云厂商',
      value: currentRowData.value?.vendor,
      property: { id: 'vendor', name: '云厂商', type: 'enum', option: VendorMap },
    },
    {
      label: '管理业务',
      value: currentRowData.value?.bk_biz_id,
      property: { id: 'bk_biz_id', name: '管理业务', type: 'business' },
    },
    {
      label: '二级账号名称',
      value: currentRowData.value?.name,
      property: { id: 'name', name: '二级账号名称', type: 'string' },
    },

    {
      label: '二级账号ID',
      value: currentRowData.value?.extension?.cloud_main_account_id,
      property: { id: 'extension.cloud_main_account_id', name: '二级账号ID', type: 'string' },
    },
    {
      label: '使用业务',
      value: currentRowData.value?.usage_biz_ids,
      property: { id: 'usage_biz_ids', name: '使用业务', type: 'business' },
      display: { appearance: 'tag' },
    },
    {
      label: '账号邮箱',
      value: currentRowData.value?.email,
      property: { id: 'email', name: '账号邮箱', type: 'string' },
    },
    {
      label: '创建时间',
      value: currentRowData.value?.cloud_created_at,
      property: { id: 'cloud_created_at', name: '创建时间', type: 'datetime' },
    },
    {
      label: '负责人',
      value: currentRowData.value?.managers,
      property: { id: 'managers', name: '负责人', type: 'user' },
    },
    {
      label: '更新时间',
      value: currentRowData.value?.updated_at,
      property: { id: 'updated_at', name: '更新时间', type: 'datetime' },
    },
    {
      label: '安全负责人',
      value: currentRowData.value?.security_managers,
      property: { id: 'security_managers', name: '安全负责人', type: 'user' },
    },
    {
      label: '备注',
      value: currentRowData.value?.memo,
      property: { id: 'memo', name: '备注', type: 'string' },
    },
  ],
);

// 获取密钥类型文本
const getSecretTypeText = (type: string) => {
  const typeMap: Record<string, string> = {
    resource: '资源管理',
    security: '安全管理',
  };
  return typeMap[type] || type;
};

// 获取密钥状态配置
const getSecretStatus = (status: string) => {
  const statusMap: Record<string, { theme: string; text: string }> = {
    normal: { theme: 'success', text: '正常' },
    invalid: { theme: 'danger', text: '失败' },
  };
  return statusMap[status] || { theme: 'default', text: status };
};

// 打开录入密钥弹窗
const handleAddSecret = () => {
  isEditMode.value = false;
  editingSecret.value = null;
  showSecretDialog.value = true;
};

// 打开编辑密钥弹窗
const handleEditSecret = (row: IAccountSecretItem) => {
  isEditMode.value = true;
  editingSecret.value = { ...row };
  showSecretDialog.value = true;
};

// 删除密钥确认
const handleDeleteSecret = (row: IAccountSecretItem) => {
  InfoBox({
    title: '确定删除该密钥？',
    subTitle: '删除后将无法恢复，请谨慎操作',
    type: 'warning',
    confirmText: '确定',
    cancelText: '取消',
    onConfirm: async () => {
      try {
        await secondaryAccountStore.deleteAccountSecret(getBizsId(), row.vendor, [row.id]);
        Message({ theme: 'success', message: '删除成功' });
        fetchSecretList();
      } catch (error) {
        console.error('删除密钥失败:', error);
        Message({ theme: 'error', message: '删除失败' });
      }
    },
  });
};

// 密钥操作成功回调
const handleSecretSuccess = () => {
  showSecretDialog.value = false;
  fetchSecretList();
  emit('update-success');
};

// 编辑基本信息
const handleEditBaseInfo = () => {
  if (currentRowData.value) {
    showAccountFormSideslider.value = true;
  }
};

// 跳转到三级账号列表（按当前二级账号ID筛选）
const handleGoToTertiaryAccount = () => {
  const cloudMainAccountId = (currentRowData.value as any)?.extension?.cloud_main_account_id;
  if (!cloudMainAccountId) return;
  // 构建目标 query：切换 tab + 带上筛选条件 + 移除 id
  const { id, ...restQuery } = route.query;
  router.push({
    query: {
      ...restQuery,
      type: 'tertiary-account',
      filter: `extension.cloud_main_account_id=${cloudMainAccountId}`,
    },
  });
};

// 新建三级账号成功回调
const handleCreateSubAccountSuccess = () => {
  showCreateSubAccount.value = false;
  emit('update-success');
};

// 编辑基本信息成功回调
const handleAccountFormSuccess = (updatedData?: ISecondaryAccountItem) => {
  if (updatedData) {
    // 更新本地数据
    currentRowData.value = { ...updatedData };
    // 通知父组件更新列表
    emit('update-success');
    // 同步更新 rowData（如果父组件需要）
    emit('update:rowData', updatedData);
  }
};

// 同步账号功能
const handleSyncAccount = () => {
  if (!currentRowData.value) {
    Message({ theme: 'warning', message: '暂无账号信息' });
    return;
  }

  const accountName = currentRowData.value.name || currentRowData.value.extension?.cloud_main_account_id || '';
  const SyncContent = () =>
    h('div', { class: 'sync-info-content' }, [
      h('p', { class: 'sync-info-title' }, `即将同步账号：${accountName}`),
      h('p', { class: 'sync-info-subtitle' }, '同步信息包含：'),
      h('ul', { class: 'sync-info-list' }, [
        h('li', '二级账号本身的信息（邮箱、保护状态、MFA等）'),
        h('li', '二级账号下的三级账号'),
        h('li', '二级账号下的权限模板'),
      ]),
      h('p', { class: 'sync-info-tip' }, '同步操作可能需要几分钟，请耐心等待'),
    ]);

  InfoBox({
    title: '确定同步该二级账号信息',
    type: 'warning',
    subTitle: SyncContent,
    width: 480,
    contentAlign: 'left',
    confirmText: '确定',
    cancelText: '取消',
    beforeClose: (action: string) =>
      new Promise(async (resolve) => {
        if (action === 'confirm') {
          const loadingBox = InfoBox({
            type: 'loading',
            title: '同步二级账号信息中...',
            subTitle: '请耐心等待',
            width: 400,
            closeIcon: false,
            showMask: true,
            quickClose: false,
            escClose: false,
            confirmText: '',
            cancelText: '关闭',
          });

          try {
            const bkBizId = getBizsId();
            const { vendor } = currentRowData.value!;
            const accountIds = [currentRowData.value!.id];

            const results = await secondaryAccountStore.syncSecondaryAccounts(bkBizId, vendor, accountIds);
            loadingBox.hide();

            if (results.failed.length === 0) {
              Message({ theme: 'success', message: '账号同步成功' });
              // 重新拉取详情以刷新数据
              const detail = await secondaryAccountStore.getSecondaryAccountDetail(
                getBizsId(),
                currentRowData.value!.id,
              );
              if (detail) {
                currentRowData.value = { ...detail };
                emit('update:rowData', detail);
                emit('update-success');
              }
              resolve(true);
            } else {
              Message({ theme: 'error', message: '账号同步失败，请稍后重试' });
            }
          } catch (error) {
            console.error('同步失败:', error);
            loadingBox.hide();
            Message({ theme: 'error', message: '账号同步失败，请稍后重试' });
            resolve(true);
          }
        }
        if (action === 'cancel') {
          resolve(true);
        }
      }),
  });
};
</script>

<template>
  <bk-sideslider v-model:is-show="model" title="二级账号详情" :width="960" quick-close background-color="#f5f7fa">
    <template #header>
      <div class="detail-header">
        <span class="title">二级账号详情</span>
        <div class="header-actions">
          <bk-button @click="handleSyncAccount">同步</bk-button>
        </div>
      </div>
    </template>

    <div class="account-detail-container">
      <!-- 基本信息 -->
      <bk-card class="info-card" :border="false" :disable-header-style="true">
        <template #header>
          <div class="card-header btw">
            <span class="card-title">基本信息</span>
            <hcm-auth
              v-if="getBizsId()"
              :sign="{ type: AUTH_BIZ_UPDATE_SECONDARY_ACCOUNT, relation: [getBizsId()] }"
              v-slot="{ noPerm }"
            >
              <bk-button theme="primary" outline size="small" :disabled="noPerm" @click="handleEditBaseInfo">
                编辑
              </bk-button>
            </hcm-auth>
          </div>
        </template>
        <div class="info-grid">
          <template v-for="(field, index) in baseInfoFields" :key="index">
            <div class="info-item">
              <span class="info-label">{{ field.label }}：</span>
              <span class="info-value">
                <display-value :value="field.value" :property="field.property" :display="field.display" />
              </span>
            </div>
          </template>
        </div>
      </bk-card>

      <!-- 三级账号 -->
      <bk-card class="info-card" :border="false" :disable-header-style="true">
        <template #header>
          <div class="card-header">
            <span class="card-title">三级账号</span>
            <bk-button theme="primary" text class="add-btn" @click="showCreateSubAccount = true">
              <i class="hcm-icon bkhcm-icon-plus-circle-shape"></i>
              新建三级账号
            </bk-button>
          </div>
        </template>
        <div class="sub-account-info">
          <span class="label">三级账号数量：</span>
          <span class="count">{{ currentRowData?.sub_account_count ?? 0 }} 个</span>
          <Share class="icon-link" @click="handleGoToTertiaryAccount" />
        </div>
      </bk-card>

      <!-- 资源密钥 -->
      <bk-card class="info-card" :border="false" :disable-header-style="true">
        <template #header>
          <div class="card-header">
            <span class="card-title">资源密钥</span>
            <hcm-auth
              :sign="{ type: AUTH_BIZ_CREATE_SECONDARY_ACCOUNT, relation: [accountStore.bizs] }"
              v-slot="{ noPerm }"
            >
              <bk-button theme="primary" text class="add-btn" :disabled="noPerm" @click="handleAddSecret">
                <i class="hcm-icon bkhcm-icon-plus-circle-shape"></i>
                录入密钥
              </bk-button>
            </hcm-auth>
          </div>
        </template>
        <bk-loading :loading="secretLoading">
          <bk-table :data="secretList" :border="['row']" show-overflow-tooltip row-hover="auto">
            <bk-table-column label="密钥用途" prop="type" min-width="100">
              <template #default="{ row }">
                {{ getSecretTypeText(row.type) }}
              </template>
            </bk-table-column>
            <bk-table-column label="密钥ID" prop="extension.cloud_secret_id" min-width="140">
              <template #default="{ row }">
                {{ row.extension?.cloud_secret_id || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="密钥Key" min-width="80">
              <template #default>****</template>
            </bk-table-column>
            <bk-table-column label="所属三级账号ID" prop="extension.cloud_sub_account_id" min-width="140">
              <template #default="{ row }">
                {{ row.extension?.cloud_sub_account_id || '--' }}
              </template>
            </bk-table-column>
            <bk-table-column label="密钥状态" prop="status" min-width="100">
              <template #default="{ row }">
                <span :class="['status-dot', `status-${row.status}`]"></span>
                {{ getSecretStatus(row.status).text }}
              </template>
            </bk-table-column>
            <bk-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <hcm-auth
                  :sign="{ type: AUTH_BIZ_UPDATE_SECONDARY_ACCOUNT, relation: [accountStore.bizs] }"
                  v-slot="{ noPerm }"
                >
                  <bk-button theme="primary" text class="mr8" :disabled="noPerm" @click="handleEditSecret(row)">
                    编辑
                  </bk-button>
                </hcm-auth>
                <hcm-auth
                  :sign="{ type: AUTH_BIZ_DELETE_SECONDARY_ACCOUNT, relation: [accountStore.bizs] }"
                  v-slot="{ noPerm }"
                >
                  <bk-button
                    theme="danger"
                    text
                    :disabled="row.status === 'normal' || noPerm"
                    @click="handleDeleteSecret(row)"
                  >
                    删除
                  </bk-button>
                </hcm-auth>
              </template>
            </bk-table-column>
          </bk-table>
        </bk-loading>
      </bk-card>
    </div>

    <!-- 录入/编辑密钥弹窗 -->
    <SecretKeyDialog
      v-model="showSecretDialog"
      :is-edit="isEditMode"
      :secret-data="editingSecret"
      :account-id="currentRowData?.id || ''"
      @success="handleSecretSuccess"
    />

    <!-- 编辑二级账号弹窗 -->
    <AccountFormSideslider
      v-model="showAccountFormSideslider"
      :is-edit="true"
      :account-data="currentRowData"
      @success="handleAccountFormSuccess"
    />

    <!-- 新建三级账号弹窗 -->
    <AccountCreateSideslider
      v-model:model-value="showCreateSubAccount"
      :default-account-id="currentRowData?.id || ''"
      @success="handleCreateSubAccountSuccess"
    />
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.account-detail-container {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  .info-card {
    :deep(.bk-card-head) {
      padding: 16px 24px;
    }

    :deep(.bk-card-body) {
      padding: 5px 0 32px 40px;
    }
  }

  .card-header {
    display: flex;
    align-items: center;

    &.btw {
      justify-content: space-between;
    }

    width: 100%;

    .card-title {
      font-size: 14px;
      font-weight: 700;
      color: #313238;
      margin-right: 16px;
    }

    .add-btn {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;

      i {
        margin-right: 4px;
      }

      .icon-plus {
        font-size: 14px;
      }
    }
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 15px 48px;
    color: #4d4f56;

    .info-item {
      display: flex;
      align-items: flex-start;
      font-size: 12px;
      line-height: 20px;

      .info-label {
        min-width: 84px;
        flex-shrink: 0;
        text-align: right;
      }

      .info-value {
        flex: 1;
        word-break: break-all;
        min-width: 0;
        overflow: hidden;

        .tag-item {
          margin-right: 4px;
          margin-bottom: 4px;
        }
      }
    }
  }

  .sub-account-info {
    display: flex;
    align-items: center;
    font-size: 12px;

    .label {
      color: #4d4f56;
    }

    .count {
      color: #313238;
      font-weight: 700;
      margin-right: 8px;
    }

    .icon-link {
      color: #3a84ff;
      cursor: pointer;
    }
  }

  .status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 6px;

    &.status-normal {
      background-color: #2dcb56;
    }

    &.status-invalid {
      background-color: #ea3636;
    }
  }

  .mr8 {
    margin-right: 8px;
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
</style>

<style lang="scss">
.sync-info-content {
  text-align: left;
  padding: 12px 16px;
  background-color: rgb(245 247 250);

  .sync-info-title {
    font-size: 12px;
    color: #313238;
    line-height: 20px;
    font-weight: 700;
    margin-bottom: 12px;
  }

  .sync-info-subtitle {
    font-size: 12px;
    color: #4d4f56;
    line-height: 20px;
    font-weight: 700;
  }

  .sync-info-list {
    margin: 0;
    padding-left: 22px;

    li {
      font-size: 12px;
      color: #4d4f56;
      line-height: 20px;
      list-style-type: disc;
    }
  }

  .sync-info-tip {
    font-size: 12px;
    margin-top: 22px;
    line-height: 20px;
  }
}
</style>
