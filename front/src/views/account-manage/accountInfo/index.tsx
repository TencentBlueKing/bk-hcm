import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { defineComponent, ref, watch } from 'vue';
import { RouterView, useRoute, useRouter } from 'vue-router';
import './index.scss';
import {
  MENU_SERVICE_ACCOUNT_BASIC,
  MENU_SERVICE_ACCOUNT_RESOURCE,
  MENU_SERVICE_ACCOUNT_USERS,
} from '@/constants/menu-symbol';

// 账号详情子路由对应的 tab 配置，key 为字符串用于 Tab 组件，name 为 Symbol 用于路由导航
const ACCOUNT_DETAIL_TABS = [
  { key: 'basic', label: '基本信息', name: MENU_SERVICE_ACCOUNT_BASIC },
  { key: 'resource', label: '资源状态', name: MENU_SERVICE_ACCOUNT_RESOURCE },
  { key: 'user', label: '用户列表', name: MENU_SERVICE_ACCOUNT_USERS },
];

// 根据路由 name(Symbol) 找到对应的 tab key(string)
const getTabKeyByRouteName = (routeName: symbol | string | undefined | null) => {
  const tab = ACCOUNT_DETAIL_TABS.find((t) => t.name === routeName);
  return tab?.key || ACCOUNT_DETAIL_TABS[0].key;
};

export default defineComponent({
  setup() {
    const router = useRouter();
    const route = useRoute();

    const activeTab = ref(getTabKeyByRouteName(route.name));
    watch(
      () => route.name,
      (name) => {
        if (name) {
          activeTab.value = getTabKeyByRouteName(name);
        }
      },
    );

    const handleTabChange = (key: string) => {
      const tab = ACCOUNT_DETAIL_TABS.find((t) => t.key === key);
      if (tab) {
        router.push({ name: tab.name, params: route.params });
      }
    };

    return () => (
      <div class={'account-info-container'}>
        <Tab v-model:active={activeTab.value} type='card-grid' onChange={handleTabChange}>
          {ACCOUNT_DETAIL_TABS.map(({ key, label }) => (
            <BkTabPanel key={key} label={label} name={key as any} renderDirective='if'>
              <RouterView />
            </BkTabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
