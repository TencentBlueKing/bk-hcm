<script lang="ts" setup>
import { computed, reactive, onMounted, useTemplateRef } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { provideBreadcrumb } from '@/hooks/use-breadcrumb';
import { providePermissionDialog } from '@/hooks/use-permission-dialog';
import HcmHeader from '@/components/layout/header.vue';
import HcmMenu from '@/components/layout/menu.vue';
import HcmFooter from '@/components/layout/footer.vue';
import HcmMainContent from '@/components/layout/main-content.vue';
import Notice from '@/views/notice/index.vue';
import PermissionApplyDialog from '@/components/permission/apply-dialog.vue';
import { MENU_BUSINESS_HOST_MANAGEMENT } from '@/constants/menu-symbol';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { VendorEnum } from '@/common/constant';

const { ENABLE_NOTICE } = window.PROJECT_CONFIG;
const { t } = useI18n();
const route = useRoute();

const isStatusPage = computed(() => route.name === '404' || route.name === 'error');
const isNeedSideMenu = computed(() => !isStatusPage.value);
const hasFooter = computed(() => route.name === MENU_BUSINESS_HOST_MANAGEMENT);

// 面包屑
provideBreadcrumb();

// 权限申请弹窗
const permissionDialogContext = providePermissionDialog();

// 导航布局状态
const navigationState = reactive({
  collapse: false,
});

const permissionDialogRef = useTemplateRef<InstanceType<typeof PermissionApplyDialog>>('permission-dialog');

onMounted(() => {
  window.hcmPermissionDialog = permissionDialogRef.value;

  // TODO: 以下数据预加载未来将替换为组件内部按需获取
  // 例如：region-value 组件、business-value 组件等
  const { fetchRegions } = useRegionsStore();
  const { fetchBusinessMap } = useBusinessMapStore();
  const { fetchAllCloudAreas } = useCloudAreaStore();

  fetchRegions(VendorEnum.TCLOUD);
  fetchRegions(VendorEnum.HUAWEI);
  fetchBusinessMap();
  fetchAllCloudAreas();
});

const handleCollapse = (collapse: boolean) => {
  navigationState.collapse = !collapse;
};
</script>

<template>
  <bk-navigation
    :class="['hcm-app', { 'has-footer': hasFooter }]"
    navigation-type="top-bottom"
    :side-title="t('海垒')"
    :need-menu="isNeedSideMenu"
    :default-open="!navigationState.collapse"
    @toggle="handleCollapse"
  >
    <template #side-icon>
      <img src="@/assets/image/logo.png" width="28" />
    </template>
    <template #header>
      <HcmHeader />
    </template>
    <template #default>
      <HcmMainContent />
    </template>
    <template #menu>
      <HcmMenu />
    </template>
    <template v-if="hasFooter" #footer>
      <HcmFooter />
    </template>
  </bk-navigation>
  <Notice v-if="ENABLE_NOTICE === 'true'" />
  <PermissionApplyDialog
    ref="permission-dialog"
    v-model="permissionDialogContext.isShow"
    :permission="permissionDialogContext.permission"
    :done="permissionDialogContext.done"
  />
</template>

<style lang="scss" scoped>
.hcm-app {
  :deep(.bk-navigation-wrapper) {
    .navigation-container {
      // 覆盖组件内联的 max-width: calc(100vw - 60px)，使导航容器跟随文档宽度而非视口宽度
      max-width: none !important;

      .container-content {
        padding: 0;
      }
    }
  }

  &.has-footer {
    :deep(.main-content) {
      height: calc(100% - 52px);
    }
  }

  &:deep(.bk-navigation-title) {
    .title-desc {
      font-size: 18px;
    }
  }

  :deep(.bk-navigation-wrapper) {
    .navigation-nav {
      .nav-slider {
        // background-color: #2c354d;

        .business-selector-global {
          margin: 0 8px 10px 8px;
        }

        // .bk-menu {
        //   background-color: #2c354d;

        //   .menu-warp {
        //     background-color: #2c354d;
        //   }
        // }
      }
    }
  }
}
</style>
