import { defineComponent, computed, watch, ref, nextTick, onMounted } from 'vue';
import { RouterLink, RouterView, useRoute } from 'vue-router';

import { Menu, Navigation, Dropdown, Button } from 'bkui-vue';
import Breadcrumb from './breadcrumb';
import BusinessSelector from './business-selector';
import NoPermission from '@/views/resource/NoPermission';
import AccountList from '../resource/resource-manage/account/accountList';

import cookie from 'cookie';
import { useI18n } from 'vue-i18n';
import { useVerify } from '@/hooks';
import { useUserStore, useAccountStore, useCommonStore } from '@/store';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import usePagePermissionStore from '@/store/usePagePermissionStore';

import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useChangeHeaderTab from './hooks/useChangeHeaderTab';
import { LANGUAGE_TYPE, VendorEnum } from '@/common/constant';
import { classes } from '@/common/util';

import { headRouteConfig } from '@/router/header-config';
import logo from '@/assets/image/logo.png';
import './index.scss';

const { ENABLE_CLOUD_SELECTION } = window.PROJECT_CONFIG;
// import { CogShape } from 'bkui-vue/lib/icon';
// import { useProjectList } from '@/hooks';
// import AddProjectDialog from '@/components/AddProjectDialog';

const { DropdownMenu, DropdownItem } = Dropdown;
const { VERSION } = window.PROJECT_CONFIG;

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

    const openedKeys: string[] = [];
    const isRouterAlive = ref<Boolean>(true);
    const curYear = ref(new Date().getFullYear());
    const isMenuOpen = ref<boolean>(true);
    const language = ref(cookie.parse(document.cookie).blueking_language || 'zh-cn');

    const isNeedSideMenu = computed(() => ![Senarios.resource, Senarios.scheme].includes(whereAmI.value));

    const { hasPagePermission, permissionMsg, logout } = usePagePermissionStore();

    const { topMenuActiveItem, menus, curPath, handleHeaderMenuClick } = useChangeHeaderTab();

    const saveLanguage = async (val: string) => {
      return new Promise((resovle) => {
        const { BK_COMPONENT_API_URL } = window.PROJECT_CONFIG;
        const url = `${BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fe_update_user_language/language=${val}`;

        const scriptTag = document.createElement('script');
        scriptTag.setAttribute('type', 'text/javascript');
        scriptTag.setAttribute('src', url);
        const headTag = document.getElementsByTagName('head')[0];
        headTag.appendChild(scriptTag);
        resovle(val);
      });
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
            <AccountList />
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
        const { getAuthVerifyData } = useVerify(); // 权限中心权限
        await getAuthVerifyData(bizsPageAuthData);
      },
      { immediate: true },
    );

    watch(
      () => language.value,
      async (val) => {
        document.cookie = `blueking_language=${val}; domain=${window.PROJECT_CONFIG.BK_DOMAIN}`;
        await saveLanguage(val);
        location.reload();
      },
    );

    if (!hasPagePermission) return () => <NoPermission message={permissionMsg} />;

    return () => (
      <main class='flex-column full-page home-page'>
        {/* <Header></Header> */}
        <div class='flex-1'>
          {
            <Navigation
              navigationType={NAV_TYPE}
              hoverWidth={NAV_WIDTH}
              defaultOpen={isMenuOpen.value}
              needMenu={isNeedSideMenu.value}
              onToggle={handleToggle}
              class={route.path !== '/business/host' ? 'no-footer' : ''}>
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
                            (ENABLE_CLOUD_SELECTION !== 'true' && id !== 'scheme') || ENABLE_CLOUD_SELECTION === 'true',
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
                      style={{
                        width: `${NAV_WIDTH}px`,
                      }}
                      uniqueOpen={false}
                      openedKeys={openedKeys}
                      activeKey={route.meta.activeKey as string}>
                      {menus.value.map((menuItem) =>
                        Array.isArray(menuItem.children) && !menuItem.meta?.hasPageRoute ? (
                          <Menu.Group key={menuItem.path as string} name={menuItem.name as string}>
                            {{
                              default: () =>
                                menuItem.children
                                  .filter((child) => !child.meta?.notMenu)
                                  .map((child) => (
                                    <RouterLink to={{ path: `${child.path}`, query: { bizs: accountStore.bizs } }}>
                                      <Menu.Item key={child.meta?.activeKey as string}>
                                        {/* {route.meta.activeKey} */}
                                        {{
                                          icon: () => <i class={child.meta?.icon} />,
                                          default: () => (
                                            <p class='flex-row flex-1 justify-content-between align-items-center pr16'>
                                              <span class='flex-1 text-ov'>{child.name as string}</span>
                                            </p>
                                          ),
                                        }}
                                      </Menu.Item>
                                    </RouterLink>
                                  )),
                            }}
                          </Menu.Group>
                        ) : (
                          !menuItem.meta?.notMenu && (
                            <RouterLink to={`${menuItem.path}`}>
                              <Menu.Item key={menuItem.meta.activeKey as string}>
                                {/* {menuItem.meta.activeKey} */}
                                {{
                                  icon: () => <i class={menuItem.meta.icon} />,
                                  default: () => menuItem.name as string,
                                }}
                              </Menu.Item>
                            </RouterLink>
                          )
                        ),
                      )}
                    </Menu>
                  </div>
                ),
                default: () => (
                  <>
                    {whereAmI.value === Senarios.resource ? null : <Breadcrumb></Breadcrumb>}
                    <div class={['/service/my-apply'].includes(curPath.value) ? 'view-warp no-padding' : 'view-warp'}>
                      {isRouterAlive.value ? renderRouterView() : null}
                    </div>
                  </>
                ),
                footer: () => `Copyright © ${curYear.value} Tencent BlueKing. All Rights Reserved. ${VERSION}`,
              }}
            </Navigation>
          }
        </div>
        {/* <AddProjectDialog isShow={showAddProjectDialog.value} onClose={toggleAddProjectDialog} /> */}
      </main>
    );
  },
});
