<script setup lang="ts">
import { computed, defineAsyncComponent, provide, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { VendorEnum } from '@/common/constant';
import VendorSelector from './components/vendor-selector.vue';
import { TAB_PANELS } from './constants';

const tabPanels = TAB_PANELS;

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
};
const currentComponent = computed(() => tabComponents[tabActive.value as string]);

// 用户点击 tab 时，更新 URL query.type（@change 仅在用户点击时触发，不会在代码修改 active 时触发）
const handleTabChange = (name: string) => {
  router.replace({ query: { ...route.query, type: name, filter: undefined } });
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
