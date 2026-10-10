<script lang="ts" setup>
import { computed, nextTick, onMounted, ref, useTemplateRef, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { provideBreadcrumb } from '@/hooks/use-breadcrumb';
import { providePermissionDialog } from '@/hooks/use-permission-dialog';
import { useVerify } from '@/hooks';
import { getCurrentTopMenu } from '@/common/menu-service';
import { VendorEnum } from '@/common/constant';
import { MENU_BUSINESS_HOST_MANAGEMENT, MENU_RESOURCE, MENU_SCHEME } from '@/constants/menu-symbol';
import { useAccountStore, useCommonStore } from '@/store';
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
const accountStore = useAccountStore();
const { getAuthVerifyData } = useVerify();
const { isNoticeAlert } = storeToRefs(useCommonStore());

// 面包屑
provideBreadcrumb();

// 权限申请弹窗
const permissionDialogContext = providePermissionDialog();

const permissionDialogRef = useTemplateRef<InstanceType<typeof PermissionApplyDialog>>('permission-dialog');

const isMenuOpen = ref(true);
const isRouterAlive = ref(true);

const currentTopMenu = computed(() => getCurrentTopMenu(route));
// 资源接入、资源选型不声明侧栏，状态页与 403 不属于任何一级菜单
const needMenu = computed(() => !!currentTopMenu.value?.menu?.length);
const hasFooter = computed(() => route.name === MENU_BUSINESS_HOST_MANAGEMENT);

const reload = () => {
  isRouterAlive.value = false;
  nextTick(() => {
    isRouterAlive.value = true;
  });
};

const handleToggle = (open: boolean) => {
  isMenuOpen.value = open;
};

watch(
  () => accountStore.bizs,
  async (bizs) => {
    if (!bizs) return;
    const commonStore = useCommonStore();
    const { pageAuthData } = commonStore; // 所有需要检验的查看权限数据
    const bizsPageAuthData = pageAuthData.map((e: any) => {
      // eslint-disable-next-line no-prototype-builtins
      if (e.hasOwnProperty('bk_biz_id')) {
        e.bk_biz_id = bizs;
      }
      return e;
    });
    commonStore.updatePageAuthData(bizsPageAuthData);
    await getAuthVerifyData(bizsPageAuthData);
  },
  { immediate: true },
);

// 切换一级菜单时初始化全局业务 ID
watch(
  () => route.path,
  () => {
    const topMenuId = currentTopMenu.value?.id;
    if (topMenuId === MENU_RESOURCE || topMenuId === MENU_SCHEME) {
      accountStore.updateBizsId(0);
    } else if (!topMenuId) {
      accountStore.updateBizsId('');
    }
  },
  { immediate: true },
);

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
    :hover-width="240"
    :need-menu="needMenu"
    :default-open="isMenuOpen"
    @toggle="handleToggle"
  >
    <template #side-icon>
      <img class="hcm-app-logo" :src="logo" />
    </template>
    <template #header>
      <HcmHeader />
    </template>
    <template #menu>
      <HcmMenu :is-open="isMenuOpen" :reload="reload" />
    </template>
    <template #default>
      <HcmMainContent :is-router-alive="isRouterAlive" />
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

  // 覆盖组件的 max-width: calc(100vw - 60px)，使导航容器跟随文档宽度（html/body 最小 1366px）而非视口宽度
  :deep(.bk-navigation-wrapper .navigation-container) {
    max-width: none !important;
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
