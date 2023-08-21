import {
  defineComponent,
  reactive,
  watch,
  ref,
  nextTick,
  onMounted,
} from 'vue';
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import { Menu, Navigation, Dropdown, Dialog, Exception, Button } from 'bkui-vue';
import { headRouteConfig } from '@/router/header-config';
import Breadcrumb from './breadcrumb';
import workbench from '@/router/module/workbench';
import resource from '@/router/module/resource';
import service from '@/router/module/service';
import business from '@/router/module/business';
import { classes } from '@/common/util';
import logo from '@/assets/image/logo.png';
import './index.scss';
import { useUserStore, useAccountStore, useCommonStore } from '@/store';
import { useVerify } from '@/hooks';
import { useI18n } from 'vue-i18n';
import { useRegionsStore } from '@/store/useRegionsStore';
import { LANGUAGE_TYPE, RESOURCE_TABS, VendorEnum } from '@/common/constant';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import cookie from 'cookie';
import NoPermission from '@/views/resource/NoPermission';
import usePagePermissionStore from '@/store/usePagePermissionStore';
// @ts-ignore
import AppSelect from '@blueking/app-select';
import '@blueking/app-select/dist/style.css';
import { getFavoriteList, useFavorite } from '@/hooks/useFavorite';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import BkTab, { BkTabPanel } from 'bkui-vue/lib/tab';

// import { CogShape } from 'bkui-vue/lib/icon';
// import { useProjectList } from '@/hooks';
// import AddProjectDialog from '@/components/AddProjectDialog';

const { DropdownMenu, DropdownItem } = Dropdown;
const { VERSION } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const NAV_WIDTH = 240;
    const NAV_TYPE = 'top-bottom';

    const { t } = useI18n();
    const route = useRoute();
    const router = useRouter();
    const userStore = useUserStore();
    const accountStore = useAccountStore();
    const { fetchBusinessMap } = useBusinessMapStore();
    const { fetchAllCloudAreas } = useCloudAreaStore();
    const { whereAmI } = useWhereAmI();

    let topMenuActiveItem = '';
    let menus: RouteRecordRaw[] = [];
    const openedKeys: string[] = [];
    let path = '';
    const curPath = ref('');
    const businessId = ref<number>(0);
    const businessList = ref<any[]>([]);
    const loading = ref<Boolean>(false);
    const isRouterAlive = ref<Boolean>(true);
    const curYear = ref(new Date().getFullYear());
    const isMenuOpen = ref<boolean>(true);
    const language = ref(cookie.parse(document.cookie).blueking_language || 'zh-cn');
    const isDialogShow = ref(false);
    const favoriteList = ref([]);
    const activeResourceTab = ref(RESOURCE_TABS[0].key);

    const {
      favoriteSet,
      addToFavorite,
      removeFromFavorite,
    } = useFavorite(businessId.value, favoriteList.value);

    const { hasPagePermission, permissionMsg, logout } = usePagePermissionStore();
    // 获取业务列表
    const getBusinessList = async () => {
      try {
        loading.value = true;
        const res = await accountStore.getBizListWithAuth();
        loading.value = false;
        businessList.value = res?.data;
        if (!businessList.value.length) {
          // 没有权限
          router.push({
            name: '403',
            params: {
              id: 'biz_access',
            },
          });
          return;
        }
        businessId.value = accountStore.bizs || res?.data[0].id; // 默认取第一个业务
        accountStore.updateBizsId(businessId.value); // 设置全局业务id
      } catch (error) {
        console.log(error);
      }
    };

    const changeMenus = (id: string, ...subPath: string[]) => {
      console.log('subPath', subPath, id);
      openedKeys.push(`/${id}`);
      switch (id) {
        case 'business':
          topMenuActiveItem = 'business';
          menus = reactive(business);
          path = '/business/host';
          if (!accountStore.bizs) accountStore.updateBizsId(useBusinessMapStore().businessList?.[0]?.id);
          getBusinessList(); // 业务下需要获取业务列表
          break;
        case 'resource':
          topMenuActiveItem = 'resource';
          menus = reactive(resource);
          path = '/resource/account';
          accountStore.updateBizsId(0); // 初始化业务ID
          break;
        case 'service':
          topMenuActiveItem = 'service';
          menus = reactive(service);
          path = '/service/service-apply';
          break;
        case 'workbench':
          topMenuActiveItem = 'workbench';
          menus = reactive(workbench);
          // path = '/workbench/auto';
          path = '/workbench/audit';
          accountStore.updateBizsId(0); // 初始化业务ID
          break;
        default:
          if (subPath[0] === 'biz_access') {
            topMenuActiveItem = 'business';
            menus = reactive(business);
            path = '/business/host';
          } else {
            topMenuActiveItem = 'resource';
            menus = reactive(resource);
            path = '/resource/account';
          }
          accountStore.updateBizsId(''); // 初始化业务ID
          break;
      }
    };

    watch(
      () => route,
      (val) => {
        const { bizs } = val.query;
        if (bizs) {
          businessId.value = Number(bizs); // 取地址栏的业务id
          accountStore.updateBizsId(businessId.value); // 设置全局业务id
        }
        curPath.value = route.path;
        const pathArr = val.path.slice(1, val.path.length).split('/');
        changeMenus(pathArr.shift(), ...pathArr);
      },
      {
        immediate: true,
        deep: true,
      },
    );

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

    const handleHeaderMenuClick = (id: string, routeName: string): void => {
      if (route.name !== routeName) {
        changeMenus(id);
        router.push({
          path,
        });
      }
    };

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

    watch(
      () => language.value,
      async (val) => {
        document.cookie = `blueking_language=${val}; domain=${window.PROJECT_CONFIG.BK_DOMAIN}`;
        await saveLanguage(val);
        location.reload();
      },
    );

    // 选择业务
    const handleChange = async (val: {id: number}) => {
      businessId.value = val.id;
      accountStore.updateBizsId(businessId.value); // 设置全局业务id
      // @ts-ignore
      const isbusinessDetail = route.name?.includes('BusinessDetail');
      if (isbusinessDetail) {
        const businessListPath = route.path.split('/detail')[0];
        router.push({
          path: businessListPath,
        });
      } else {
        reload();
      }
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

    const { fetchRegions } = useRegionsStore();


    watch(
      () => businessId.value,
      async () => {
        favoriteList.value = await getFavoriteList(businessId.value);
        for (const id of favoriteList.value) favoriteSet.value.add(id);
      },
    );

    /**
     * 在这里获取项目公共数据并缓存
     */
    onMounted(() => {
      fetchRegions(VendorEnum.TCLOUD);
      fetchRegions(VendorEnum.HUAWEI);
      fetchBusinessMap();
      fetchAllCloudAreas();
    });

    if (!hasPagePermission) return () => <NoPermission message={permissionMsg} />;

    return () => (
      <main class='flex-column full-page'>
        {/* <Header></Header> */}
        <div class='flex-1'>
          {
            <Navigation
              navigationType={NAV_TYPE}
              hoverWidth={NAV_WIDTH}
              defaultOpen={isMenuOpen.value}
              onToggle={handleToggle}>
              {{
                'side-header': () => (
                  <div class='left-header flex-row justify-content-between align-items-center'>
                    <div class='logo'>
                      <img class='logo-icon' src={logo} />
                    </div>
                    <div class='title-text'>{t('海垒')}</div>
                  </div>
                ),
                header: () => (
                  <header class='bk-hcm-header'>
                    <section class='flex-row justify-content-between header-width'>
                      {headRouteConfig.map(({ id, route, name, href }) => (
                        <a
                          class={classes(
                            {
                              active: topMenuActiveItem === id,
                            },
                            'header-title',
                          )}
                          key={id}
                          aria-current='page'
                          href={href}
                          onClick={() => handleHeaderMenuClick(id, route)}>
                          {t(name)}
                        </a>
                      ))}
                    </section>
                    <aside class='header-lang'>
                      <Dropdown trigger='click'>
                        {{
                          default: () => (
                            <span class='cursor-pointer flex-row align-items-center '>
                              {language.value === LANGUAGE_TYPE.en
                                ? 'English'
                                : '中文'}
                              <i
                                class={
                                  'icon hcm-icon bkhcm-icon-down-shape pl5'
                                }
                              />
                            </span>
                          ),
                          content: () => (
                            <DropdownMenu>
                              <DropdownItem
                                onClick={() => {
                                  language.value = LANGUAGE_TYPE.zh_cn;
                                }}>
                                {'中文'}
                              </DropdownItem>
                              <DropdownItem
                                onClick={() => {
                                  language.value = LANGUAGE_TYPE.en;
                                }}>
                                {'English'}
                              </DropdownItem>
                            </DropdownMenu>
                          ),
                        }}
                      </Dropdown>
                    </aside>
                    <aside class='header-user'>
                      <Dropdown trigger='click'>
                        {{
                          default: () => (
                            <span class='cursor-pointer flex-row align-items-center '>
                              {userStore.username}
                              <i
                                class={
                                  'icon hcm-icon bkhcm-icon-down-shape pl5'
                                }
                              />
                            </span>
                          ),
                          content: () => (
                            <DropdownMenu>
                              <DropdownItem onClick={logout}>
                                {t('退出')}
                              </DropdownItem>
                            </DropdownMenu>
                          ),
                        }}
                      </Dropdown>
                    </aside>
                  </header>
                ),
                menu: () => (
                  <>
                    {topMenuActiveItem === 'business' && isMenuOpen.value ? (
                      <AppSelect
                        data={businessList.value}
                        onChange={handleChange}
                        theme={'dark'}
                        class={'bk-hcm-app-selector'}
                        value={businessList.value?.[0]}
                        minWidth={360}>
                        {{
                          default: ({
                            data,
                          }: {
                            data: { id: number; name: string };
                          }) => (
                            <div class='bk-hcm-app-selector-item'>
                              <div class='bk-hcm-app-selector-item-content'>
                                <span class={'bk-hcm-app-selector-item-content-name'}>{`${data.name}`}</span>
                                &nbsp;&nbsp;&nbsp;
                                <span class={'bk-hcm-app-selector-item-content-id'}>{`(#${data.id})`}</span>
                              </div>

                              <div class='bk-hcm-app-selector-item-star'>
                                {
                                  favoriteSet.value.has(data.id)
                                    ? <i class={'icon bk-icon icon-collect'} style={{ color: '#CC933A' }} onClick={(event) => {
                                      removeFromFavorite(data.id);
                                      event.stopPropagation();
                                    }}/>
                                    : <i class={'icon bk-icon icon-not-favorited'} onClick={(event) => {
                                      addToFavorite(data.id);
                                      event.stopPropagation();
                                    }}/>
                                }
                              </div>
                            </div>
                          ),
                          append: () => (
                            <div class={'app-action-content'} onClick={() => isDialogShow.value = true}>
                              <i class={'icon bk-icon icon-plus-circle app-action-content-icon'}/>
                              <span class={'app-action-content-text'}>新建业务</span>
                            </div>
                          ),
                        }}
                      </AppSelect>
                    ) : (
                      ''
                    )}

                    <Menu
                      class='menu-warp'
                      style={{
                        width: `${NAV_WIDTH}px`,
                      }}
                      uniqueOpen={false}
                      openedKeys={openedKeys}
                      activeKey={route.meta.activeKey as string}>
                      {menus.map(menuItem => (Array.isArray(menuItem.children) ? (
                          <Menu.Group
                            key={menuItem.path as string}
                            name={menuItem.name as string}>
                            {{
                              default: () => menuItem.children.map(child => (
                                  <RouterLink to={`${child.path}`}>
                                    <Menu.Item
                                      key={child.meta.activeKey as string}>
                                      {/* {route.meta.activeKey} */}
                                      {{
                                        icon: () => <i class={'icon bk-icon icon-automatic-typesetting'}/>,
                                        default: () => <p class='flex-row flex-1 justify-content-between align-items-center pr16'>
                                          <span class='flex-1 text-ov'>
                                            {child.name as string}
                                          </span>
                                        </p>,
                                      }}
                                    </Menu.Item>
                                  </RouterLink>
                              )),
                            }}
                          </Menu.Group>
                      ) : (
                          <RouterLink to={`${menuItem.path}`}>
                            <Menu.Item key={menuItem.meta.activeKey as string}>
                              {/* {menuItem.meta.activeKey} */}
                              {{
                                icon: () => <i class={'icon bk-icon icon-automatic-typesetting'}/>,
                                default: () => menuItem.name as string,
                              }}
                            </Menu.Item>
                          </RouterLink>
                      )))}
                    </Menu>
                  </>
                ),
                default: () => (
                  <>
                    {whereAmI.value === Senarios.resource ? (
                      <div class='navigation-resource'>
                        <div class={'card-layout'}>
                          <p class={'resource-title'}>云账号1</p>
                          <BkTab
                            class={'ml15'}
                            type={'unborder-card'}
                            v-model={activeResourceTab}
                          >
                            {
                              RESOURCE_TABS.map(({ key, label }) => (
                                <BkTabPanel
                                  label={label}
                                  key={key}
                                  name={key}
                                />
                              ))
                            }
                          </BkTab>
                        </div>
                      </div>
                    ) : (
                      <div class='navigation-breadcrumb'>
                        <Breadcrumb></Breadcrumb>
                      </div>
                    )}
                    <div
                      class={
                        ['/service/my-apply'].includes(curPath.value)
                          ? 'view-warp no-padding'
                          : 'view-warp'
                      }>
                      {isRouterAlive.value ? <RouterView></RouterView> : ''}
                    </div>
                  </>
                ),

                footer: () => (
                  // eslint-disable-next-line max-len
                  <div class='mt20'>
                    Copyright © 2012-{curYear.value} BlueKing - Hybrid Cloud
                    Management System. All Rights Reserved.{VERSION}
                  </div>
                ),
              }}
            </Navigation>
          }
        </div>
        {/* <AddProjectDialog isShow={showAddProjectDialog.value} onClose={toggleAddProjectDialog} /> */}

        <Dialog
          isShow={isDialogShow.value}
          title=''
          dialogType='show'
          theme='primary'
          onConfirm={ () => isDialogShow.value = false }
          onClosed={ () => isDialogShow.value = false }
        >
            <Exception
              type='building'
              class={'hcm-create-business-dialog-exception-building-picture'}
              title={'新建业务参考以下指引'}
            >
              <div class={'hcm-create-business-dialog-exception-building-tips'}>
                {/* <p class={'hcm-create-business-dialog-exception-building-tips-text1'}>
                  可以按照以下方式进行查看
                </p> */}
                <p class={'hcm-create-business-dialog-exception-building-tips-text2'}>
                  业务是蓝鲸配置平台的管理空间，可以满足不同团队，不同项目的资源隔离管理需求。
                  <Button
                    theme='primary'
                    text
                  >
                    新建业务指引
                  </Button>
                </p>
              </div>
            </Exception>
        </Dialog>
      </main>
    );
  },
});
