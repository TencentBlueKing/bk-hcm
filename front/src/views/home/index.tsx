import { defineComponent, computed, watch, ref, nextTick, onMounted } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';

import { Menu, Navigation, Dropdown, Button } from 'bkui-vue';
import ReleaseNote from './release-note/index.vue';
import Breadcrumb from './breadcrumb';
import BusinessSelector from './business-selector';
import NoPermission from '@/views/resource/NoPermission';
import AccountVendorGroup from '@/views/resource/resource-manage/account/vendor-group/index.vue';
import GlobalPermissionDialog from '@/components/global-permission-dialog';

import Cookies from 'js-cookie';
import { useI18n } from 'vue-i18n';
import { useVerify } from '@/hooks';
import { useUserStore, useAccountStore, useCommonStore } from '@/store';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import usePagePermissionStore from '@/store/usePagePermissionStore';

import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useChangeHeaderTab from './hooks/useChangeHeaderTab';
import { GLOBAL_BIZS_KEY, LANGUAGE_TYPE, VendorEnum } from '@/common/constant';
import { classes } from '@/common/util';

import { headRouteConfig } from '@/router/header-config';
import logo from '@/assets/image/logo.png';
import './index.scss';

import {
  MENU_BUSINESS_LOAD_BALANCER,
  MENU_BUSINESS_OPERATION_LOG,
  MENU_BUSINESS_TASK_MANAGEMENT,
  MENU_SERVICE_TICKET_MANAGEMENT,
} from '@/constants/menu-symbol';
import { jsonp } from '@/http';
import i18n from '@/language/i18n';

// import { CogShape } from 'bkui-vue/lib/icon';
// import { useProjectList } from '@/hooks';
// import AddProjectDialog from '@/components/AddProjectDialog';

const { DropdownMenu, DropdownItem } = Dropdown;
const { VERSION, BK_COMPONENT_API_URL, BK_DOMAIN, ENABLE_CLOUD_SELECTION, ENABLE_ACCOUNT_BILL } = window.PROJECT_CONFIG;

export default defineComponent({
  name: 'Home',
  setup() {
    const NAV_WIDTH = 240;
    const NAV_TYPE = 'top-bottom';

    const { t } = useI18n();
    const route = useRoute();
    const userStore = useUserStore();
    const accountStore = useAccountStore();
    const { fetchBusinessMap } = useBusinessMapStore();
    const { fetchAllCloudAreas } = useCloudAreaStore();
    const { fetchRegions } = useRegionsStore();
    const { whereAmI } = useWhereAmI();
    const { getAuthVerifyData, authVerifyData } = useVerify(); // 权限中心权限

    const openedKeys: string[] = [];
    const isRouterAlive = ref<Boolean>(true);
    const curYear = ref(new Date().getFullYear());
    const isMenuOpen = ref<boolean>(true);
    const language = ref(Cookies.get('blueking_language') || i18n.global.locale.value);

    const isNeedSideMenu = computed(
      () => ![Senarios.resource, Senarios.scheme, Senarios.unauthorized].includes(whereAmI.value),
    );

    const { hasPagePermission, permissionMsg, logout } = usePagePermissionStore();

    const { topMenuActiveItem, menus, curPath, handleHeaderMenuClick } = useChangeHeaderTab();

    const saveLanguage = async (language: string) => {
      return jsonp(`${BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fe_update_user_language`, { language });
    };

    // 过渡方式，最终希望所有路由通过name跳转
    const getRouteLinkParams = (config: any) => {
      if (
        [
          MENU_BUSINESS_TASK_MANAGEMENT,
          MENU_SERVICE_TICKET_MANAGEMENT,
          MENU_BUSINESS_OPERATION_LOG,
          MENU_BUSINESS_LOAD_BALANCER,
        ].includes(config.name)
      ) {
        return { name: config.name };
      }
      return { path: config.path };
    };

    // 切换路由
    const reload = () => {
      isRouterAlive.value = false;
      nextTick(() => {
        isRouterAlive.value = true;
      });
    };

    // 点击
    const handleToggle = (val: any) => {
      isMenuOpen.value = val;
    };

    const renderRouterView = () => {
      if (whereAmI.value !== Senarios.resource) return <RouterView />;
      return (
        <div class={'resource-manage-container'}>
          <div class='fixed-account-list-container'>
            <AccountVendorGroup />
          </div>
          <RouterView class={'router-view-content'} />
        </div>
      );
    };

    /**
     * 在这里获取项目公共数据并缓存
     */
    onMounted(() => {
      fetchRegions(VendorEnum.TCLOUD);
      fetchRegions(VendorEnum.HUAWEI);
      fetchBusinessMap();
      fetchAllCloudAreas();
    });

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

    watch(
      () => language.value,
      async (val) => {
        document.cookie = `blueking_language=${val}; domain=${BK_DOMAIN}`;
        await saveLanguage(val);
        location.reload();
      },
    );

    if (!hasPagePermission) return () => <NoPermission message={permissionMsg} />;

    return () => (
      <main class='home-page'>
        <Navigation
          navigationType={NAV_TYPE}
          hoverWidth={NAV_WIDTH}
          defaultOpen={isMenuOpen.value}
          needMenu={isNeedSideMenu.value}
          onToggle={handleToggle}
          class={['flex-1', { 'no-footer': route.path !== '/business/host' }]}>
          {{
            'side-header': () => (
              <div class='left-header flex-row justify-content-between align-items-center'>
                <img class='logo-icon' src={logo} />
                <div class='title-text'>{t('海垒')}</div>
              </div>
            ),
            header: () => (
              <header class='bk-hcm-header'>
                <section class='flex-row justify-content-between header-width'>
                  {headRouteConfig
                    .filter(
                      ({ id }) =>
                        ((ENABLE_CLOUD_SELECTION !== 'true' && id !== 'scheme') || ENABLE_CLOUD_SELECTION === 'true') &&
                        ((ENABLE_ACCOUNT_BILL !== 'true' && id !== 'bill') || ENABLE_ACCOUNT_BILL === 'true'),
                    )
                    .map(({ id, name, path }) => (
                      <Button
                        text
                        class={classes(
                          {
                            active: topMenuActiveItem.value === id,
                          },
                          'header-title',
                        )}
                        key={id}
                        aria-current='page'
                        onClick={() => handleHeaderMenuClick(id, path)}>
                        {t(name)}
                      </Button>
                    ))}
                </section>
                <aside class='header-lang'>
                  <Dropdown>
                    {{
                      default: () => (
                        <span class='cursor-pointer flex-row align-items-center '>
                          {language.value === LANGUAGE_TYPE.en ? (
                            <span class='hcm-icon bkhcm-icon-yuyanqiehuanyingwen'></span>
                          ) : (
                            <span class='hcm-icon bkhcm-icon-yuyanqiehuanzhongwen'></span>
                          )}
                        </span>
                      ),
                      content: () => (
                        <DropdownMenu>
                          <DropdownItem
                            onClick={() => {
                              language.value = LANGUAGE_TYPE.zh_cn;
                            }}>
                            <span
                              class='hcm-icon bkhcm-icon-yuyanqiehuanzhongwen pr5'
                              style={{ fontSize: '16px' }}></span>
                            {'中文'}
                          </DropdownItem>
                          <DropdownItem
                            onClick={() => {
                              language.value = LANGUAGE_TYPE.en;
                            }}>
                            <span
                              class='hcm-icon bkhcm-icon-yuyanqiehuanyingwen pr5'
                              style={{ fontSize: '16px' }}></span>
                            {'English'}
                          </DropdownItem>
                        </DropdownMenu>
                      ),
                    }}
                  </Dropdown>
                </aside>
                <ReleaseNote />
                <aside class='header-user'>
                  <Dropdown>
                    {{
                      default: () => (
                        <span class='cursor-pointer flex-row align-items-center '>
                          {userStore.username}
                          <span class='hcm-icon bkhcm-icon-down-shape pl5'></span>
                        </span>
                      ),
                      content: () => (
                        <DropdownMenu>
                          <DropdownItem onClick={logout}>{t('退出登录')}</DropdownItem>
                        </DropdownMenu>
                      ),
                    }}
                  </Dropdown>
                </aside>
              </header>
            ),
            menu: () => (
              <div class={'home-menu'}>
                {topMenuActiveItem.value === 'business' && isMenuOpen.value && <BusinessSelector reload={reload} />}
                <Menu
                  class='menu-warp'
                  style={{ width: `${NAV_WIDTH}px` }}
                  uniqueOpen={false}
                  openedKeys={openedKeys}
                  activeKey={route.meta.activeKey?.toString()}>
                  {menus.value
                    .map((menuItem) => {
                      // menuItem.children 是一个数组, 且没有配置 hasPageRoute(页面级子路由)
                      if (Array.isArray(menuItem.children) && !menuItem.meta?.hasPageRoute) {
                        const children = menuItem.children
                          // 过滤掉非菜单的路由项
                          .filter((child) => !child.meta?.notMenu)
                          // 构建子菜单项
                          .map((child) => {
                            // 如果配置了 checkAuth, 则检查菜单是否具有访问权限
                            if (
                              child.meta?.checkAuth &&
                              !authVerifyData.value?.permissionAction[child.meta?.checkAuth as string]
                            ) {
                              return null;
                            }

                            return (
                              <RouterLink
                                to={{
                                  ...getRouteLinkParams(child),
                                  query: {
                                    [GLOBAL_BIZS_KEY]:
                                      whereAmI.value === Senarios.business ? accountStore.bizs : undefined,
                                  },
                                }}>
                                <Menu.Item key={child.meta?.activeKey?.toString()}>
                                  {{
                                    icon: () => <i class={child.meta?.icon} />,
                                    default: () => (
                                      <p class='flex-row flex-1 justify-content-between align-items-center pr16'>
                                        <span class='flex-1 text-ov'>{child.meta?.title}</span>
                                      </p>
                                    ),
                                  }}
                                </Menu.Item>
                              </RouterLink>
                            );
                          })
                          // 过滤掉 null 项
                          .filter((item) => !!item);

                        // 如果构建的子菜单项为空, 则表明子菜单都不具备访问权限, 直接隐藏 group
                        if (!children.length) return null;

                        return (
                          <Menu.Group key={menuItem.path as string} name={menuItem.meta?.groupTitle as string}>
                            {{ default: () => children }}
                          </Menu.Group>
                        );
                      }

                      // 如果配置了 notMenu、或者配置了 checkAuth 且不具备访问权限, 则隐藏菜单
                      if (
                        menuItem.meta?.notMenu ||
                        (menuItem.meta?.checkAuth &&
                          !authVerifyData.value?.permissionAction[menuItem.meta.checkAuth as string])
                      ) {
                        return null;
                      }

                      // 正常显示菜单
                      return (
                        <RouterLink to={getRouteLinkParams(menuItem)}>
                          <Menu.Item key={menuItem.meta?.activeKey?.toString()}>
                            {{
                              icon: () => <i class={menuItem.meta.icon} />,
                              default: () => menuItem.meta?.title,
                            }}
                          </Menu.Item>
                        </RouterLink>
                      );
                    })
                    // 过滤掉 null 项
                    .filter((item) => !!item)}
                </Menu>
              </div>
            ),
            default: () => (
              <>
                {whereAmI.value === Senarios.resource ? null : <Breadcrumb></Breadcrumb>}
                <div class={['/service/ticket'].includes(curPath.value) ? 'view-warp no-padding' : 'view-warp'}>
                  {isRouterAlive.value ? renderRouterView() : null}
                </div>
              </>
            ),
            footer: () => `Copyright © ${curYear.value} Tencent BlueKing. All Rights Reserved. ${VERSION}`,
          }}
        </Navigation>
        <GlobalPermissionDialog />
      </main>
    );
  },
});
