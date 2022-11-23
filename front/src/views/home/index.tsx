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
import { classes } from '@/common/util';
import logo from '@/assets/image/logo.png';
import './index.scss';
import { useUser } from '@/store';

// import { CogShape } from 'bkui-vue/lib/icon';
// import { useProjectList } from '@/hooks';
// import AddProjectDialog from '@/components/AddProjectDialog';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  setup() {
    const NAV_WIDTH = 240;
    const NAV_TYPE = 'top-bottom';

    const route = useRoute();
    const router = useRouter();
    const userStore = useUser();

    let topMenuActiveItem = '';
    let menus: RouteRecordRaw[] = [];
    let openedKeys: string[] = [];
    let path = '';

    const changeMenus = (id: string, subPath: string[] = []) => {
      switch (id) {
        case 'business':
          topMenuActiveItem = 'business';
          menus = reactive(business);
          path = '/business/host';
          openedKeys = [`/business${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
        case 'resource':
          topMenuActiveItem = 'resource';
          menus = reactive(resource);
          path = '/resource/account';
          openedKeys = [`/resource${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
        case 'service':
          topMenuActiveItem = 'service';
          menus = reactive(service);
          path = '/service/serviceApply';
          openedKeys = [`/service${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
        case 'workbench':
          topMenuActiveItem = 'workbench';
          menus = reactive(workbench);
          path = '/workbench/auto';
          openedKeys = [`/workbench${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
        default:
          topMenuActiveItem = 'resource';
          menus = reactive(resource);
          path = '/resource/account';
          openedKeys = [`/resource${subPath[1] ? `/${subPath[0]}` : ''}`];
          break;
      }
    };

    watch(
      () => route,
      (val) => {
        const pathArr = val.path.slice(1, val.path.length).split('/');
        changeMenus(pathArr[0], [pathArr[1], pathArr[2]]);
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
      console.log('退出');
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
                        <div class="title-text">海垒2.0</div>
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
                              {name}
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
                                  退出登陆
                                  </DropdownItem>
                                </DropdownMenu>
                              ),
                            }}
                          </Dropdown>
                        </aside>
                      </header>
                    ),
                    menu: () => (
                      <Menu style={`width: ${NAV_WIDTH}px`} uniqueOpen openedKeys={openedKeys} activeKey={route.meta.activeKey as string}>
                        {
                          menus.map(menuItem => (Array.isArray(menuItem.children) ? (
                            <Menu.Submenu
                              key={menuItem.path as string}
                              title={menuItem.name as string}>
                            {{
                              // icon: () => <menuItem.icon/>,
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
                              <Menu.Item key={menuItem.meta.activeKey as string}>
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
                        <RouterView></RouterView>
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
