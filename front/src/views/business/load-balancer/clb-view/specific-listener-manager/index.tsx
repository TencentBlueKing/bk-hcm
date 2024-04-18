import { defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
// import components
import { Tab } from 'bkui-vue';
import DomainList from './domain-list';
import ListenerDetail from './listener-detail';
// import stores
import { useLoadBalancerStore } from '@/store';
// import hooks
import useActiveTab from '@/hooks/useActiveTab';
// import constants
import { TRANSPORT_LAYER_LIST } from '@/constants';
// import utils
import bus from '@/common/bus';
import './index.scss';

const { TabPanel } = Tab;

enum TabTypeEnum {
  list = 'list',
  detail = 'detail',
}

export default defineComponent({
  name: 'SpecificListenerManager',
  setup() {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    const { activeTab, handleActiveTabChange } = useActiveTab(TabTypeEnum.list);
    const tabList = ref([]);

    watch(
      () => loadBalancerStore.currentSelectedTreeNode.id,
      () => {
        const { type } = loadBalancerStore.currentSelectedTreeNode;
        if (type !== 'listener') return;
        const { protocol } = loadBalancerStore.currentSelectedTreeNode;
        if (TRANSPORT_LAYER_LIST.includes(protocol)) {
          // 4层监听器没有下级资源，直接显示基本信息
          handleActiveTabChange(TabTypeEnum.detail);
          tabList.value = [{ name: TabTypeEnum.detail, label: '基本信息', component: <ListenerDetail /> }];
        } else {
          handleActiveTabChange(TabTypeEnum.list);
          tabList.value = [
            { name: TabTypeEnum.list, label: '域名', component: <DomainList /> },
            { name: TabTypeEnum.detail, label: '基本信息', component: <ListenerDetail /> },
          ];
        }
      },
      { immediate: true },
    );

    onMounted(() => {
      bus.$on('changeSpecificListenerActiveTab', handleActiveTabChange);
    });

    onUnmounted(() => {
      bus.$off('changeSpecificListenerActiveTab');
    });

    return () => (
      <Tab
        class='manager-tab-wrap has-breadcrumb'
        v-model:active={activeTab.value}
        type='card-grid'
        onChange={handleActiveTabChange}>
        {tabList.value.map((tab) => (
          <TabPanel key={tab.name} name={tab.name} label={tab.label}>
            <div class='common-card-wrap'>{tab.component}</div>
          </TabPanel>
        ))}
      </Tab>
    );
  },
});
