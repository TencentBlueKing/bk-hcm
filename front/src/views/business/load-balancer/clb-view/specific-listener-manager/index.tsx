import { defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
// import components
import { Tab } from 'bkui-vue';
import DomainList from './domain-list';
import ListenerDetail from './listener-detail';
// import stores
import { useLoadBalancerStore } from '@/store';
// import constants
import { TRANSPORT_LAYER_LIST } from '@/constants';
// import utils
import bus from '@/common/bus';
import './index.scss';

const { TabPanel } = Tab;

export default defineComponent({
  name: 'SpecificListenerManager',
  setup() {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    const activeTab = ref('domain');
    const tabList = ref([]);

    watch(
      () => loadBalancerStore.currentSelectedTreeNode.id,
      () => {
        const { type } = loadBalancerStore.currentSelectedTreeNode;
        if (type !== 'listener') return;
        const { protocol } = loadBalancerStore.currentSelectedTreeNode;
        if (TRANSPORT_LAYER_LIST.includes(protocol)) {
          // 4层监听器没有下级资源，直接显示基本信息
          activeTab.value = 'detail';
          tabList.value = [{ name: 'detail', label: '基本信息', component: <ListenerDetail /> }];
        } else {
          tabList.value = [
            { name: 'domain', label: '域名', component: <DomainList /> },
            { name: 'detail', label: '基本信息', component: <ListenerDetail /> },
          ];
        }
      },
      { immediate: true },
    );

    onMounted(() => {
      bus.$on('changeSpecificListenerActiveTab', (v: any) => (activeTab.value = v));
    });

    onUnmounted(() => {
      bus.$off('changeSpecificListenerActiveTab');
    });

    return () => (
      <Tab class='manager-tab-wrap has-breadcrumb' v-model:active={activeTab.value} type='card-grid'>
        {tabList.value.map((tab) => (
          <TabPanel key={tab.name} name={tab.name} label={tab.label}>
            <div class='common-card-wrap'>{tab.component}</div>
          </TabPanel>
        ))}
      </Tab>
    );
  },
});
