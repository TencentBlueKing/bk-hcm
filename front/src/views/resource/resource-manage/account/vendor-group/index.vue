<script lang="ts" setup>
import { ref, watch, watchEffect } from 'vue';
import { storeToRefs } from 'pinia';
import { escapeRegExp } from 'lodash';
import { useRoute, useRouter } from 'vue-router';
import { VendorEnum, VendorMap } from '@/common/constant';
import { QueryRuleOPEnum, IAccountItem } from '@/typings';
import { AUTH_IMPORT_ACCOUNT } from '@/constants/auth-symbols';
import { useAccountSelectorStore } from '@/store/account-selector';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useResourceAccount } from '../accountList/useResourceAccount';
import CreateAccount from '../createAccount';
import { vendorProperty } from './vendor.plugin';

const route = useRoute();
const router = useRouter();
const accountSelectorStore = useAccountSelectorStore();
const resourceAccountStore = useResourceAccountStore();

const { authorizedResourceAccountList, authorizedResourceAccountLoading } = storeToRefs(accountSelectorStore);
const isCreateAccountDialogShow = ref(false);
const searchValue = ref('');
const expandedMap = ref<Map<VendorEnum, boolean>>(new Map());
const vendorAccountMap = ref<Map<VendorEnum, IAccountItem[]>>(new Map());
const isAccountAllEmpty = ref(false);

const { accountId, setAccountId } = useResourceAccount();

watch([authorizedResourceAccountList, searchValue], ([newList, searchValue]) => {
  vendorAccountMap.value.clear();
  for (const [vendor] of vendorProperty) {
    const accountList = newList.filter(
      (item) => item.vendor === vendor && item.name.toLowerCase().includes(searchValue.toLowerCase()),
    );
    vendorAccountMap.value.set(vendor, accountList);

    // 设置展开状态，保持上一次的状态或者在搜索命中时自动展开
    expandedMap.value.set(vendor, expandedMap.value.get(vendor) || (searchValue.length > 0 && accountList.length > 0));
  }

  // 检查是否所有vendor的账号都为空
  isAccountAllEmpty.value = Array.from(vendorAccountMap.value.values()).every(
    (accountList) => accountList.length === 0,
  );
});

watch(
  () => route.query.dialog,
  (dialog) => {
    isCreateAccountDialogShow.value = dialog?.length > 0;
  },
  {
    deep: true,
    immediate: true,
  },
);

watch(
  [accountId, vendorAccountMap],
  ([val, accountMap]) => {
    // 初始化时存在账号ID则展开对应的vendor
    if (val) {
      for (const [vendor, accountList] of accountMap) {
        if (accountList.some((item) => item.id === val)) {
          expandedMap.value.set(vendor, true);
          break;
        }
      }
    }
  },
  {
    deep: true,
    immediate: true,
  },
);

watchEffect(() => {
  // 滚动获取所有账号列表数据
  accountSelectorStore.getAuthorizedResourceAccountList({
    filter: {
      op: 'and',
      rules: [
        {
          field: 'type',
          op: QueryRuleOPEnum.EQ,
          value: 'resource',
        },
      ],
    },
  });
});

const highlightSearchValue = (text: string) => {
  if (!searchValue.value) {
    return text;
  }
  return text.replace(new RegExp(escapeRegExp(searchValue.value), 'gi'), (match) => `<span class="hl">${match}</span>`);
};

const toggleExpanded = (vendor: VendorEnum) => {
  expandedMap.value.set(vendor, !expandedMap.value.get(vendor));
};

const handleSelectAll = () => {
  setAccountId('');
  resourceAccountStore.clear();
};

const handleSelectVendor = (vendor: VendorEnum) => {
  resourceAccountStore.setCurrentVendor(vendor);
  toggleExpanded(vendor);
  setAccountId('');
};

const handleSelectAccount = (account: IAccountItem) => {
  resourceAccountStore.setCurrentAccountSimpleInfo(account);
  resourceAccountStore.setCurrentVendor(null);
  setAccountId(account.id);
};

const handleCreateShow = () => {
  router.push({
    query: {
      ...route.query,
      dialog: 'create_account',
    },
  });
};
const handleCreateCancel = () => {
  router.push({
    query: {
      ...route.query,
      dialog: undefined,
    },
  });
};
</script>

<template>
  <div class="account-vendor-group">
    <div class="search-bar">
      <bk-input type="search" clearable placeholder="搜索云账号" class="search-input" v-model="searchValue" />
    </div>
    <div class="toolbar">
      账号列表
      <hcm-auth :sign="{ type: AUTH_IMPORT_ACCOUNT }" v-slot="{ noPerm }">
        <bk-button text theme="primary" :disabled="noPerm" class="import-button" @click="handleCreateShow">
          <i class="hcm-icon bkhcm-icon-plus-circle" />
          接入
        </bk-button>
      </hcm-auth>
    </div>
    <div class="account-container" v-bkloading="{ loading: authorizedResourceAccountLoading }">
      <template v-if="!isAccountAllEmpty">
        <div
          :class="['account-all', { active: !resourceAccountStore.vendorInResourcePage }]"
          v-show="!searchValue.length"
          @click="handleSelectAll"
        >
          <img src="@/assets/image/all-vendors.png" alt="全部账号" class="icon" />
          全部账号
        </div>
        <div class="group-list g-scroller">
          <div class="group-item" v-for="[vendor, accountList] in vendorAccountMap" :key="vendor">
            <div
              :class="['vendor-title', { active: resourceAccountStore.currentVendor === vendor }]"
              @click="handleSelectVendor(vendor)"
              v-show="accountList.length"
            >
              <i :class="['hcm-icon', 'bkhcm-icon-right-shape', 'arrow-icon', { expanded: expandedMap.get(vendor) }]" />
              <img :src="vendorProperty.get(vendor).icon" :class="['vendor-icon', vendor]" />
              <div class="vendor-name">{{ VendorMap[vendor] }}</div>
              <em class="account-count">{{ accountList.length }}</em>
            </div>
            <div class="account-list" v-show="accountList.length && expandedMap.get(vendor)">
              <div
                :class="['account-item', { active: accountId === account.id }]"
                v-for="(account, index) in accountList"
                :key="index"
                @click="handleSelectAccount(account)"
              >
                <img
                  src="@/assets/image/success-account.png"
                  class="status-icon"
                  v-if="account.sync_status === 'sync_success'"
                />
                <img src="@/assets/image/failed-account.png" class="status-icon" v-else />
                <!-- eslint-disable-next-line vue/no-v-html -->
                <div class="account-name" :title="account.name" v-html="highlightSearchValue(account.name)"></div>
              </div>
            </div>
          </div>
        </div>
      </template>
      <div class="data-empty" v-else-if="!authorizedResourceAccountLoading">
        <bk-exception description="暂无账号" scene="part" type="empty" v-if="!searchValue" />
        <bk-exception description="无搜索结果" scene="part" type="search-empty" v-else />
      </div>
    </div>
  </div>
  <create-account :is-show="isCreateAccountDialogShow" :on-cancel="handleCreateCancel" />
</template>

<style lang="scss" scoped>
.account-vendor-group {
  height: 100%;
  background: #fff;
  padding: 10px 0;

  .search-bar {
    padding: 0 16px;
    margin-bottom: 10px;

    .search-input {
      border-radius: 2px;
      border-width: 0px;

      :deep(.bk-input--text),
      :deep(.bk-input--suffix-icon) {
        background-color: #f0f1f5;
      }
    }
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    padding: 0 16px;
    margin-bottom: 10px;
    color: #979ba5;

    .import-button {
      :deep(.bk-button-text) {
        gap: 2px;
      }
    }
  }

  .account-container {
    height: calc(100% - 65px);
    overflow: hidden;
  }

  .account-all {
    display: flex;
    align-items: center;
    height: 36px;
    gap: 8px;
    padding: 0 16px;
    border-bottom: 1px solid #eaebf0;
    border-top: 1px solid #eaebf0;
    cursor: pointer;

    &:hover {
      background: #f5f7fa;
    }

    &.active {
      background: #e1ecff;
      position: relative;

      &::before {
        position: absolute;
        content: '';
        left: 0;
        width: 3px;
        height: 100%;
        background: #3a84ff;
      }
    }

    .icon {
      width: 16px;
    }
  }
}

.group-list {
  height: calc(100% - 36px);

  .vendor-title {
    display: flex;
    align-items: center;
    gap: 6px;
    position: sticky;
    top: 0;
    height: 36px;
    padding: 0 16px;
    background: #fff;
    cursor: pointer;

    &:hover {
      background: #f5f7fa;
    }

    &.active {
      background: #e1ecff;
    }

    .arrow-icon {
      font-size: 12px;
      color: #979ba5;
      transition: transform 0.3s ease;

      &.expanded {
        transform: rotate(90deg);
      }
    }

    .vendor-name {
      margin-left: 2px;
    }

    .vendor-icon {
      width: 20px;

      &.other {
        width: 18px;
      }
    }

    .account-count {
      margin-left: auto;
      font-style: normal;
      font-size: 12px;
      color: #c4c6cc;
    }
  }
}

.account-list {
  overflow: hidden;

  .account-item {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 36px;
    padding: 0 16px 0 34px;
    cursor: pointer;

    &:hover {
      background: #f5f7fa;
    }

    &.active {
      background: #e1ecff;
      position: relative;

      &::before {
        position: absolute;
        content: '';
        left: 0;
        width: 3px;
        height: 100%;
        background: #3a84ff;
      }
    }

    .status-icon {
      width: 16px;
    }

    .account-name {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;

      :deep(.hl) {
        color: #3a84ff;
      }
    }
  }
}
</style>
