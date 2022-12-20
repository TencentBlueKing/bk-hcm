import { defineComponent, onMounted, reactive, watch } from 'vue';
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import { Menu, Navigation, Dropdown } from 'bkui-vue';
import { headRouteConfig } from '@/router/header-config';
import Breadcrumb from './breadcrumb';
import workbench from '@/router/module/workbench';
import resource from '@/router/module/resource';
import service from '@/router/module/service';
import business from '@/router/module/business';
import { classes, deleteCookie } from '@/common/util';
import logo from '@/assets/image/logo.png';
import './index.scss';
import { useUserStore } from '@/store';
import { useI18n } from 'vue-i18n';

// import { CogShape } from 'bkui-vue/lib/icon';
// import { useProjectList } from '@/hooks';
// import AddProjectDialog from '@/components/AddProjectDialog';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  setup() {
    const NAV_WIDTH = 240;
    const NAV_TYPE = 'top-bottom';

    const { t } = useI18n();
    const route = useRoute();
    const router = useRouter();
    const userStore = useUserStore();

    let topMenuActiveItem = '';
    let menus: RouteRecordRaw[] = [];
    let openedKeys: string[] = [];
    let path = '';

    const changeMenus = (id: string, ...subPath: string[]) => {
      console.log('subPath', subPath);
      openedKeys = [`/${id}`];
      switch (id) {
        case 'business':
          topMenuActiveItem = 'business';
          menus = reactive(business);
          path = '/business/host';
          // openedKeys = [`/business${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
        case 'resource':
          topMenuActiveItem = 'resource';
          menus = reactive(resource);
          path = '/resource/account';
          // openedKeys = [`/resource${subPath[1] ? `/${subPath[0]}` : ''}`];
          // openedKeys = [`/resource${subPath[1] ? `/${subPath.join('/')}` : ''}`];
          break;
        case 'service':
          topMenuActiveItem = 'service';
          menus = reactive(service);
          path = '/service/service-apply';
          // openedKeys = [`/service${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
        case 'workbench':
          topMenuActiveItem = 'workbench';
          menus = reactive(workbench);
          path = '/workbench/auto';
          // openedKeys = [`/workbench${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
        default:
          topMenuActiveItem = 'resource';
          menus = reactive(resource);
          path = '/resource/account';
          // openedKeys = ['/resource'];
          // openedKeys = [`/resource${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
      }
    };

    watch(
      () => route,
      (val) => {
        const pathArr = val.path.slice(1, val.path.length).split('/');
        changeMenus(pathArr.shift(), ...pathArr);
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

    const logout = () => {
      deleteCookie('bk_token');
      deleteCookie('bk_ticket');
      const cUrl = window.location.href;
      if (window.PROJECT_CONFIG.LOGIN_FULL) {
        window.location.href = `${window.LOGIN_FULL}?c_url=${cUrl}`;
      } else {
        window.location.href = `//${window.PROJECT_CONFIG.BK_COMPONENT_API_URL || ''}/console/accounts/logout/`;
      }
    };

    onMounted(() => {
    });

    return () => (
      <main class="flex-column full-page">
        {/* <Header></Header> */}
          <div class="flex-1" >
            {
                <Navigation
                  navigationType={NAV_TYPE}
                  hoverWidth={NAV_WIDTH}
                  defaultOpen
                >
                  {{
                    'side-header': () => (
                      <div class="left-header flex-row justify-content-between align-items-center">
                        <div class="logo">
                          <img class="logo-icon" src={logo} />
                        </div>
                        <div class="title-text">{t('海垒2.0')}</div>
                      </div>
                    ),
                    header: () => (
                      <header class="bk-hcm-header">
                        <section class="flex-row justify-content-between header-width">
                          {headRouteConfig.map(({ id, route, name }) => (
                            <div
                              class={classes({
                                active: topMenuActiveItem === id,
                              }, 'header-title')}
                              key={id}
                              onClick={() => handleHeaderMenuClick(id, route)}
                            >
                              {t(name)}
                            </div>
                          ))}
                        </section>
                        <aside class="header-user">
                          <Dropdown
                            trigger='click'
                          >
                            {{
                              default: () => (
                                <span class="cursor-pointer flex-row align-items-center ">
                                  {userStore.username}
                                  <i class={'icon hcm-icon bkhcm-icon-down-shape pl5'}/>
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
                      <Menu class="menu-warp" style={`width: ${NAV_WIDTH}px`} uniqueOpen openedKeys={openedKeys} activeKey={route.meta.activeKey as string}>
                        {
                          menus.map(menuItem => (Array.isArray(menuItem.children) ? (
                            <Menu.Submenu
                              key={menuItem.path as string}
                              title={menuItem.name as string}>
                            {{
                              icon: () => <i class={'icon hcm-icon bkhcm-icon-automatic-typesetting menu-icon'}/>,
                              default: () => menuItem.children.map(child => (
                                <RouterLink to={`${child.path}`}>
                                    <Menu.Item key={child.meta.activeKey as string}>
                                      <p class="flex-row flex-1 justify-content-between align-items-center pr16">
                                        <span class="flex-1 text-ov">{child.name as string}</span>
                                      </p>
                                      {/* {route.meta.activeKey} */}
                                    </Menu.Item>
                                  </RouterLink>
                              )),
                            }}
                            </Menu.Submenu>
                          ) : (
                            <RouterLink to={`${menuItem.path}`}>
                              <Menu.Item
                              key={menuItem.meta.activeKey as string}>
                                {/* {menuItem.meta.activeKey} */}
                                {{
                                  // icon: () => <menuItem.icon/>,
                                  default: () => menuItem.name as string,
                                }}
                              </Menu.Item>
                            </RouterLink>
                          )))
                        }
                      </Menu>
                    ),
                    default: () => (
                      <div>
                        <div class="navigation-breadcrumb">
                            <Breadcrumb></Breadcrumb>
                        </div>
                        <div class="view-warp">
                          <RouterView></RouterView>
                        </div>
                      </div>
                    ),
                  }}
                </Navigation>
            }
          </div>
        {/* <AddProjectDialog isShow={showAddProjectDialog.value} onClose={toggleAddProjectDialog} /> */}
      </main>
    );
  },
});
