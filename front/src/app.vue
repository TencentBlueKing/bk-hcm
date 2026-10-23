<script lang="ts" setup>
import { computed, onMounted, reactive, useTemplateRef } from 'vue';
import { storeToRefs } from 'pinia';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { provideBreadcrumb } from '@/hooks/use-breadcrumb';
import { providePermissionDialog } from '@/hooks/use-permission-dialog';
import { getCurrentTopMenu } from '@/common/menu-service';
import { VendorEnum } from '@/common/constant';
import { MENU_BUSINESS_HOST_MANAGEMENT } from '@/constants/menu-symbol';
import { useCommonStore } from '@/store';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useChromeNotice } from '@/utils/chrome-notice';
import logo from '@/assets/image/logo.png';
import HcmHeader from '@/components/layout/header.vue';
import HcmMenu from '@/components/layout/menu.vue';
import HcmMainContent from '@/components/layout/main-content.vue';
import HcmFooter from '@/components/layout/footer.vue';
import Notice from '@/views/notice/index.vue';
import PermissionApplyDialog from '@/components/permission/apply-dialog.vue';

const { ENABLE_NOTICE } = window.PROJECT_CONFIG;

const { t } = useI18n();
const route = useRoute();
const { isNoticeAlert } = storeToRefs(useCommonStore());

// 面包屑
provideBreadcrumb();

// 权限申请弹窗
const permissionDialogContext = providePermissionDialog();

const permissionDialogRef = useTemplateRef<InstanceType<typeof PermissionApplyDialog>>('permission-dialog');

// 导航布局状态
const navigationState = reactive({
  collapse: false,
});

// 侧栏只看当前一级菜单是否声明了 menu：资源接入、资源选型不声明；404、error 等状态页不属于任何一级菜单，取不到即不显示
const needMenu = computed(() => !!getCurrentTopMenu(route)?.menu?.length);
const hasFooter = computed(() => route.name === MENU_BUSINESS_HOST_MANAGEMENT);

// toggle 回传侧栏当前是否展开（含悬停临时展开）
const handleCollapse = (expanded: boolean) => {
  navigationState.collapse = !expanded;
};

onMounted(() => {
  window.hcmPermissionDialog = permissionDialogRef.value;
  useChromeNotice();

  const { fetchRegions } = useRegionsStore();
  const { fetchBusinessMap } = useBusinessMapStore();
  const { fetchAllCloudAreas } = useCloudAreaStore();

  fetchRegions(VendorEnum.TCLOUD);
  fetchRegions(VendorEnum.HUAWEI);
  fetchBusinessMap();
  fetchAllCloudAreas();
});
</script>

<template>
  <!-- 跑马灯公告在文档流中占 40px，必须排在导航之前 -->
  <Notice v-if="ENABLE_NOTICE === 'true'" />
  <bk-navigation
    :class="['hcm-app', { 'has-notice': isNoticeAlert }]"
    navigation-type="top-bottom"
    :side-title="t('海垒')"
    :need-menu="needMenu"
    :default-open="!navigationState.collapse"
    @toggle="handleCollapse"
  >
    <template #side-icon>
      <img class="hcm-app-logo" :src="logo" />
    </template>
    <template #header>
      <HcmHeader />
    </template>
    <template #menu>
      <HcmMenu />
    </template>
    <template #default>
      <HcmMainContent />
    </template>
    <template v-if="hasFooter" #footer>
      <HcmFooter />
    </template>
  </bk-navigation>
  <PermissionApplyDialog
    ref="permission-dialog"
    v-model="permissionDialogContext.isShow"
    :permission="permissionDialogContext.permission"
    :done="permissionDialogContext.done"
  />
</template>

<style lang="scss" scoped>
.hcm-app {
  height: 100%;

  // 与 useCalcTopWithNotice 的跑马灯高度一致
  &.has-notice {
    height: calc(100% - 40px);
  }

  // 覆盖组件的 max-width: calc(100vw - 60px)，让导航容器跟随文档宽度（html/body 最小 1366px）而非视口，避免顶栏被压缩
  :deep(.bk-navigation-wrapper .navigation-container) {
    max-width: none !important;
  }

  :deep(.bk-navigation-wrapper .navigation-nav) {
    flex-shrink: 0;
  }

  :deep(.bk-navigation-wrapper .navigation-nav .nav-slider .business-selector-global) {
    margin: 0 8px 10px;
  }

  :deep(.bk-navigation-title .title-desc) {
    margin-left: 10px;
    font-size: 18px;
  }

  .hcm-app-logo {
    width: 28px;
    height: 28px;
  }
}
</style>
