<script lang="ts" setup>
import { useRouter, type RouteLocationRaw } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Cookies from 'js-cookie';
import { jsonp } from '@/http';
import i18n from '@/language/i18n';
import { GLOBAL_BIZS_KEY, LANGUAGE_TYPE } from '@/common/constant';
import { getMenus, type IMenu } from '@/common/menu-service';
import { MENU_BUSINESS, MENU_RESOURCE } from '@/constants/menu-symbol';
import { useAccountStore, useUserStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { gotoLoginPage } from '@/utils/login-helper';
import ReleaseNote from '@/components/release-note/index.vue';

const { BK_COMPONENT_API_URL, BK_DOMAIN } = window.PROJECT_CONFIG;

const { t } = useI18n();
const router = useRouter();
const userStore = useUserStore();
const accountStore = useAccountStore();

const language = Cookies.get('blueking_language') || i18n.global.locale.value;

// 一级菜单对应的路由不存在时只隐藏该项
const menus = getMenus().filter((menu) => router.hasRoute(menu.id));

const getNavLink = (nav: IMenu): RouteLocationRaw => {
  if (nav.id === MENU_BUSINESS) return { name: nav.id, query: { [GLOBAL_BIZS_KEY]: accountStore.bizs } };
  return { name: nav.id };
};

const handleNavClick = (nav: IMenu) => {
  // TODO(菜单整合二期·组装)：资源接入下线后去掉账号选择的清理
  if (nav.id !== MENU_RESOURCE) useResourceAccountStore().clear();
};

const handleLogout = () => {
  gotoLoginPage(undefined, true);
};

const handleChangeLang = async (lang: string) => {
  if (lang === language) return;
  document.cookie = `blueking_language=${lang}; domain=${BK_DOMAIN}`;
  await jsonp(`${BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fe_update_user_language`, { language: lang });
  window.location.reload();
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
      <bk-dropdown class="dropdown">
        <span class="anchor anchor-user">
          {{ userStore.username }}
          <i class="hcm-icon bkhcm-icon-down-shape user-icon"></i>
        </span>
        <template #content>
          <bk-dropdown-menu>
            <bk-dropdown-item @click="handleLogout">
              {{ t('退出登录') }}
            </bk-dropdown-item>
          </bk-dropdown-menu>
        </template>
      </bk-dropdown>
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

    .user-icon {
      padding-left: 5px;
    }
  }
}

.lang-icon {
  padding-right: 5px;
  font-size: 16px;
}
</style>
