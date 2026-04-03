<script setup lang="ts">
import { ref, computed, defineAsyncComponent, provide } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import qs from 'qs';
import { VendorEnum } from '@/common/constant';
import VendorSelector from './components/vendor-selector.vue';

// Tab面板配置
const tabPanels = [
  { name: 'secondary-account', label: '二级账号' },
  { name: 'tertiary-account', label: '三级账号' },
  { name: 'cloud-secret', label: '云密钥' },
  // { name: 'cloud-permission', label: '云权限模版' },
  { name: 'permission-policy', label: '权限策略库' },
];

const route = useRoute();
const router = useRouter();
const tabActive = ref(route.query?.type || 'secondary-account');

// 当前选中的云厂商
const currentVendor = ref<VendorEnum>(VendorEnum.TCLOUD);

const tabComponents: Record<string, ReturnType<typeof defineAsyncComponent>> = {
  'secondary-account': defineAsyncComponent(() => import('./secondary-account/index.vue')),
  'tertiary-account': defineAsyncComponent(() => import('./tertiary-account/index.vue')),
  'cloud-secret': defineAsyncComponent(() => import('./cloud-secret/index.vue')),
  'permission-policy': defineAsyncComponent(() => import('./permission-policy/index.vue')),
  // 其他Tab组件待开发
  // 'cloud-permission': defineAsyncComponent(() => import('./cloud-permission/index.vue')),
};
const currentComponent = computed(() => tabComponents[tabActive.value as string]);

// Tab切换 — 清除上一个 Tab 遗留的搜索条件和分页参数，并将 type 写入 query
const handleTabChange = (name: string) => {
  router.replace({ query: { ...route.query, type: name, filter: undefined } });
};

// 云厂商切换
const handleVendorChange = (vendor: VendorEnum) => {
  currentVendor.value = vendor;
  // TODO: 触发数据刷新
};

// 提供云厂商信息给子组件
provide('currentVendor', currentVendor);

// 提供切换到三级账号Tab并带查询参数的方法
const switchToTertiaryTab = (filter?: Record<string, any>, detailCloudId?: string) => {
  tabActive.value = 'tertiary-account';
  const query: Record<string, string> = { type: 'tertiary-account', _t: String(Date.now()) };
  if (filter) {
    query.filter = qs.stringify(filter, { arrayFormat: 'comma', encode: false });
  }
  if (detailCloudId) {
    query.detailCloudId = detailCloudId;
  }
  router.replace({ query });
};
provide('switchToTertiaryTab', switchToTertiaryTab);

// 提供切换到二级账号Tab并打开详情弹窗的方法
const switchToSecondaryTab = (detailCloudId?: string) => {
  tabActive.value = 'secondary-account';
  const query: Record<string, string> = { type: 'secondary-account', _t: String(Date.now()) };
  if (detailCloudId) {
    query.detailCloudId = detailCloudId;
  }
  router.replace({ query });
};
provide('switchToSecondaryTab', switchToSecondaryTab);
</script>

<template>
  <div class="cloud-account-manage-page">
    <div class="page-header">
      <Teleport defer to="#breadcrumbLeft">
        <VendorSelector
          style="margin-left: 12px"
          v-model="currentVendor"
          :disabled="true"
          @change="handleVendorChange"
        />
      </Teleport>
    </div>
    <div class="page-content">
      <bk-tab v-model:active="tabActive" type="unborder-card" @update:active="handleTabChange">
        <bk-tab-panel v-for="panel in tabPanels" :key="panel.name" :name="panel.name" :label="panel.label">
          <template v-if="tabActive === panel.name && currentComponent">
            <component :is="currentComponent" />
          </template>
          <template v-else-if="tabActive === panel.name">
            <div class="empty-placeholder">
              <bk-exception type="building" scene="part">
                <span>{{ panel.label }}功能开发中...</span>
              </bk-exception>
            </div>
          </template>
        </bk-tab-panel>
      </bk-tab>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.cloud-account-manage-page {
  height: 100%;
  display: flex;
  flex-direction: column;

  .page-header {
    flex-shrink: 0;
    padding: 0 24px;
    margin-bottom: 16px;

    .breadcrumb-title {
      display: flex;
      align-items: center;
      gap: 16px;

      .title {
        font-size: 16px;
        font-weight: 600;
        color: #313238;
      }
    }
  }

  .page-content {
    flex: 1;
    overflow: hidden;

    :deep(.bk-tab) {
      height: 100%;

      .bk-tab-header {
        padding: 0 24px;
        background: #fff;
        border-bottom: none;
      }

      .bk-tab-content {
        height: calc(100% - 42px);
        padding: 0;
        background: none;

        // padding: 16px 24px;
        overflow: auto;
      }
    }

    .empty-placeholder {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 400px;
      background: #fff;
      border-radius: 2px;
    }
  }
}
</style>
