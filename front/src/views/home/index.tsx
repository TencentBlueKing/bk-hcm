import { defineComponent, onMounted, ref, reactive } from 'vue';
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { Menu, Navigation, Dropdown } from 'bkui-vue';
import { headRouteConfig } from '@/router/header-config';
import Breadcrumb from './breadcrumb';
import work from '@/router/module/work';
import cost from '@/router/module/cost';
import resources from '@/router/module/resources';
import services from '@/router/module/services';
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
    const route = useRoute();
    const router = useRouter();
    const userStore = useUser();
    // const { projects, currentProjectId, handleProjectChange } = useProjectList();
    const activeItem = ref('resources');
    let menus = reactive(resources);
    let openedKeys = reactive(['/resource']);
    let path = '/resource/vm';
    const NAV_WIDTH = 240;
    const NAV_TYPE = 'top-bottom';

    const handleHeaderMenuClick = (id: string, name: string): void => {
      if (route.name !== name) {
        activeItem.value = id;
        changeMenus(activeItem.value);
        console.log('name', name);
      }
    };

    const logout = () => {
      console.log('退出');
    };

    const changeMenus = (id: string) => {
      switch (id) {
        case 'resources': {
          menus = reactive(resources);
          openedKeys = reactive(['/resource']);
          path = '/resource/vm';
          break;
        }
        case 'services': {
          menus = reactive(services);
          path = '/service/serviceApply';
          break;
        }
        case 'cost': {
          menus = reactive(cost);
          path = '/cost/resourceAnalyze';
          break;
        }
        case 'work': {
          menus = reactive(work);
          path = '/workbench/projectManage';
          break;
        }
        default: {
          menus = reactive(resources);
          path = '/resource/vm';
          break;
        }
      }
      router.push({
        path,
      });
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
                                active: activeItem.value === id,
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
                              key={menuItem.path}
                              title={menuItem.name as string}>
                            {{
                              // icon: () => <menuItem.icon/>,
                              default: () => menuItem.children.map(child => (
                                <RouterLink to={`${child.path}`}>
                                    <Menu.Item key={child.meta.activeKey as string}>
                                      <p class="flex-row flex-1 justify-content-between align-items-center pr16">
                                        <span class="flex-1 text-ov">{child.name as string}</span>
                                      </p>
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
