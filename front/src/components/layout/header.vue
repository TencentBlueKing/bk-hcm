<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import Cookies from 'js-cookie';
import { jsonp } from '@/http';
import { LANGUAGE_TYPE } from '@/common/constant';
import { getMenus, type IMenu } from '@/common/menu-service';
import { useUserStore } from '@/store/user';
import ReleaseNote from '@/components/release-note/index.vue';

const { BK_COMPONENT_API_URL, BK_HCM_DOMAIN, BK_LOGIN_URL } = window.PROJECT_CONFIG;

const { t } = useI18n();
const userStore = useUserStore();

const LANG_COOKIE_NAME = 'blueking_language';

const language = Cookies.get(LANG_COOKIE_NAME) || LANGUAGE_TYPE.zh_cn;

const menus = getMenus();

const getNavLink = (nav: IMenu) => {
  const link = { name: nav.id };
  return link;
};

const handleChangeLang = async (lang: string) => {
  Cookies.remove(LANG_COOKIE_NAME, { path: '' });
  const cookieValue = lang;

  Cookies.set(LANG_COOKIE_NAME, cookieValue, {
    expires: 366,
    domain: BK_HCM_DOMAIN,
  });

  if (BK_COMPONENT_API_URL) {
    const url = `${BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fe_update_user_language`;
    try {
      await jsonp(url, { language: cookieValue });
    } finally {
      window.location.reload();
    }
  }

  window.location.reload();
};

const handleLogout = () => {
  window.location.href = `${BK_LOGIN_URL}/?is_from_logout=1&c_url=${window.location.href}`;
};
</script>

<template>
  <header class="hcm-header">
    <nav class="header-nav">
      <RouterLink v-for="nav in menus" :key="nav.id" class="header-nav-link" :to="getNavLink(nav)">
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
              <i class="hcm-icon bkhcm-icon-yuyanqiehuanzhongwen pr5" style="font-size: 16px" />
              中文
            </bk-dropdown-item>
            <bk-dropdown-item @click="handleChangeLang(LANGUAGE_TYPE.en)">
              <i class="hcm-icon bkhcm-icon-yuyanqiehuanyingwen pr5" style="font-size: 16px" />
              English
            </bk-dropdown-item>
          </bk-dropdown-menu>
        </template>
      </bk-dropdown>
      <ReleaseNote />
      <bk-dropdown class="dropdown">
        <span class="anchor anchor-user">
          {{ userStore.username }}
          <i class="hcm-icon bkhcm-icon-down-shape pl5"></i>
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
  }
}
</style>
