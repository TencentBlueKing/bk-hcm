<script setup lang="ts">
import { computed, defineAsyncComponent, provide, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { GLOBAL_BIZS_KEY, VendorEnum } from '@/common/constant';
import VendorSelector from './components/vendor-selector.vue';

// Tab面板配置
const tabPanels = [
  { name: 'secondary-account', label: '二级账号' },
  { name: 'tertiary-account', label: '三级账号' },
  { name: 'cloud-secret', label: '云密钥' },
  { name: 'permission-template', label: '云权限模板' },
  { name: 'permission-policy', label: '权限策略库' },
];

const route = useRoute();
const router = useRouter();

// tab 状态直接从 URL query.type 派生，单向数据流
const tabActive = computed(() => (route.query.type as string) || 'secondary-account');

// 当前选中的云厂商
const currentVendor = ref<VendorEnum>(VendorEnum.TCLOUD);

const tabComponents: Record<string, ReturnType<typeof defineAsyncComponent>> = {
  'secondary-account': defineAsyncComponent(() => import('./secondary-account/index.vue')),
  'tertiary-account': defineAsyncComponent(() => import('./tertiary-account/index.vue')),
  'cloud-secret': defineAsyncComponent(() => import('./cloud-secret/index.vue')),
  'permission-policy': defineAsyncComponent(() => import('./permission-policy/index.vue')),
  'permission-template': defineAsyncComponent(() => import('./permission-template/index.vue')),
};
const currentComponent = computed(() => tabComponents[tabActive.value as string]);

// 用户点击 tab 时，更新 URL query.type（@change 仅在用户点击时触发，不会在代码修改 active 时触发）
const handleTabChange = (name: string) => {
  router.replace({ query: { [GLOBAL_BIZS_KEY]: route.query[GLOBAL_BIZS_KEY], type: name } });
};

// 云厂商切换
const handleVendorChange = (vendor: VendorEnum) => {
  currentVendor.value = vendor;
};
provide('currentVendor', currentVendor);
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
      <bk-tab v-model:active="tabActive" type="unborder-card" @change="handleTabChange">
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
