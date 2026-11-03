<script lang="ts" setup>
import { computed } from 'vue';
import { useRouter, type RouteLocationRaw } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Cookies from 'js-cookie';
import BkLoginUserinfo, { ActionItem } from '@blueking/login-userinfo';
import '@blueking/login-userinfo/vue3/vue3.css';
import http from '@/http';
import { LANGUAGE_TYPE } from '@/common/constant';
import { Message } from 'bkui-vue';
import { getMenus, type IMenu } from '@/common/menu-service';
import { MENU_RESOURCE_ACCESS } from '@/constants/menu-symbol';
import { useUserStore } from '@/store';
import { gotoLoginPage } from '@/utils/login-helper';
import ReleaseNote from '@/components/release-note/index.vue';

const { USER_MANAGE_URL, BK_DOMAIN } = window.PROJECT_CONFIG;

const { t } = useI18n();
const router = useRouter();
const userStore = useUserStore();

const LANG_COOKIE_NAME = 'blueking_language';

const language = Cookies.get(LANG_COOKIE_NAME) || LANGUAGE_TYPE.zh_cn;
const userinfo = computed(() => ({
  name: userStore.userData?.display_name ?? userStore.username,
  email: '',
  organization: userStore.userData?.tenant_id,
  timezone: userStore.userData?.time_zone,
}));

// 声明了 route 的一级菜单是跳转入口：链到其它一级菜单下的页面，自身不高亮
const getNavTarget = (nav: IMenu) => nav.route?.name ?? nav.id;

// 一级菜单对应的路由不存在时只隐藏该项
const menus = getMenus().filter((menu) => router.hasRoute(getNavTarget(menu)));

const getNavLink = (nav: IMenu): RouteLocationRaw => ({ name: getNavTarget(nav) });

// 资源接入是跳转型过渡入口：点击时（含已在资源纳管的重复导航）都用 Message 提示迁移去向。
// 原生 click 每次点击都会触发，资源接入下线时连同本入口一并清理。
// bkui 的 Message 没有 close/closeAll 之类的实例控制 API（内部 id 也由组件自己生成），
// 所以用 onClose 复位标记，保证连续点击时同屏只保留一条提示
let isMigrationMessageShowing = false;

const handleNavClick = (nav: IMenu) => {
  if (nav.id !== MENU_RESOURCE_ACCESS || isMigrationMessageShowing) return;
  isMigrationMessageShowing = true;
  Message({
    theme: 'primary',
    // 默认宽度写死 560px 会折行：auto + min-width: max-content 让文案保持单行
    width: 'auto',
    minWidth: 'max-content',
    message: '「资源接入」已迁移至 「资源运营 → 云资源 → 资源纳管」，原入口将于 2026-11-27 下线',
    onClose: () => {
      isMigrationMessageShowing = false;
    },
  });
};

// 写到蓝鲸统一域名，与其它蓝鲸系统共用语言设置（#1897 用的 BK_HCM_DOMAIN 已在 #1647 改名为 BK_DOMAIN）
const handleChangeLang = async (lang: string) => {
  Cookies.remove(LANG_COOKIE_NAME, { path: '' });
  const cookieValue = lang;

  Cookies.set(LANG_COOKIE_NAME, cookieValue, {
    expires: 366,
    domain: BK_DOMAIN,
  });

  try {
    if (USER_MANAGE_URL) {
      await http.put(
        `${USER_MANAGE_URL}/api/v3/open-web/tenant/current-user/language/`,
        { language: cookieValue },
        {
          globalHeaders: false,
          globalError: false,
          headers: { 'X-Bk-Tenant-Id': userStore.tenantId },
        },
      );
    }
  } finally {
    window.location.reload();
  }
};

const handleLogout = () => {
  gotoLoginPage(undefined, true);
};

const handleOpenUserCenter = () => {
  if (userStore.userData?.user_center_url) {
    window.open(userStore.userData.user_center_url, '_blank');
  }
};

const handleOpenAuthCenter = () => {
  if (userStore.userData?.auth_center_url) {
    window.open(userStore.userData.auth_center_url, '_blank');
  }
};
</script>

<template>
  <header class="hcm-header">
    <nav class="header-nav">
      <RouterLink
        v-for="nav in menus"
        :key="nav.id"
        class="header-nav-link"
        :to="getNavLink(nav)"
        :active-class="nav.route ? '' : undefined"
        :exact-active-class="nav.route ? '' : undefined"
        @click="handleNavClick(nav)"
      >
        {{ t(nav.i18n) }}
      </RouterLink>
    </nav>
    <div class="header-end">
      <bk-dropdown class="dropdown">
        <span class="anchor anchor-lang">
          <i class="hcm-icon bkhcm-icon-yuyanqiehuanyingwen" v-if="language === LANGUAGE_TYPE.en" />
          <i class="hcm-icon bkhcm-icon-yuyanqiehuanzhongwen" v-else />
        </span>
        <template #content>
          <bk-dropdown-menu>
            <bk-dropdown-item @click="handleChangeLang(LANGUAGE_TYPE.zh_cn)">
              <i class="hcm-icon bkhcm-icon-yuyanqiehuanzhongwen lang-icon" />
              中文
            </bk-dropdown-item>
            <bk-dropdown-item @click="handleChangeLang(LANGUAGE_TYPE.en)">
              <i class="hcm-icon bkhcm-icon-yuyanqiehuanyingwen lang-icon" />
              English
            </bk-dropdown-item>
          </bk-dropdown-menu>
        </template>
      </bk-dropdown>
      <ReleaseNote />
      <BkLoginUserinfo :userinfo="userinfo">
        <template #action>
          <ActionItem @click="handleOpenUserCenter">
            <template #icon><span class="hcm-icon bkhcm-icon-user"></span></template>
            个人中心
          </ActionItem>
          <ActionItem @click="handleOpenAuthCenter">
            <template #icon><span class="hcm-icon bkhcm-icon-permission"></span></template>
            权限中心
          </ActionItem>
          <ActionItem theme="danger" @click="handleLogout">
            <template #icon><span class="hcm-icon bkhcm-icon-export"></span></template>
            退出登录
          </ActionItem>
        </template>
      </BkLoginUserinfo>
    </div>
  </header>
</template>

<style lang="scss" scoped>
.hcm-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-nav {
  display: flex;
  gap: 8px;
}

.header-nav-link {
  color: #96a2b9;
  font-size: 14px;
  padding: 0 16px;
  line-height: 52px;

  &.router-link-active,
  &:hover {
    color: #fff;
  }
}

.header-end {
  display: flex;
  align-items: center;
  gap: 24px;

  .dropdown {
    .anchor {
      color: #96a2b9;
      cursor: pointer;

      &:hover {
        color: #fff;
      }
    }

    .anchor-lang {
      font-size: 16px;
    }
  }
}

.lang-icon {
  padding-right: 5px;
  font-size: 16px;
}
</style>
