<script setup lang="ts">
import { ref, computed, defineAsyncComponent, provide } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { VendorEnum } from '@/common/constant';
import { useCloudAccountNavStore } from '@/store/cloud-account-nav';
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
const navStore = useCloudAccountNavStore();
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

// Tab 手动切换 — 清除遗留搜索条件/分页参数，同时清除残留的跨 Tab 导航意图
const handleTabChange = (name: string) => {
  navStore.clearNavIntent();
  router.replace({ query: { type: name } });
};

// 云厂商切换
const handleVendorChange = (vendor: VendorEnum) => {
  currentVendor.value = vendor;
  // TODO: 触发数据刷新
};

// 提供云厂商信息给子组件
provide('currentVendor', currentVendor);

/**
 * 切换到三级账号 Tab（跨 Tab 跳转）
 * @param filter  要注入的搜索条件，如 { 'extension.cloud_main_account_id': '123' }
 * @param detailCloudId  要自动打开详情的三级账号 cloud_id
 */
const switchToTertiaryTab = (filter?: Record<string, any>, detailCloudId?: string) => {
  // 将跳转参数写入 store，目标 Tab 在数据就绪后消费
  navStore.setNavIntent({
    targetTab: 'tertiary-account',
    filter: filter && Object.keys(filter).length > 0 ? filter : undefined,
    detailCloudId,
  });
  tabActive.value = 'tertiary-account';
  // URL 只保留 type + 时间戳触发 watcher，不再携带 filter/detailCloudId
  router.replace({ query: { type: 'tertiary-account', _t: String(Date.now()) } });
};
provide('switchToTertiaryTab', switchToTertiaryTab);

/**
 * 切换到二级账号 Tab 并打开指定账号的详情弹窗
 * @param detailCloudId  要自动打开详情的二级账号 cloud_main_account_id
 */
const switchToSecondaryTab = (detailCloudId?: string) => {
  navStore.setNavIntent({
    targetTab: 'secondary-account',
    detailCloudId,
  });
  tabActive.value = 'secondary-account';
  router.replace({ query: { type: 'secondary-account', _t: String(Date.now()) } });
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
