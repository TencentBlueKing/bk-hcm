import { defineComponent, ref } from 'vue';
import { Tab } from 'bkui-vue';
import './index.scss';
import DomainList from './domain-list';
import ListenerInfo from './listener-info';

const { TabPanel } = Tab;

export default defineComponent({
  name: 'ListenerManager',
  setup() {
    const activeTab = ref('domain' as 'domain | info');
    const protocolType = ref('UDP' as 'HTTP | HTTPS' | 'TCP' | 'UDP');
    const tabList = [
      { name: 'domain', label: '域名', component: <DomainList /> },
      { name: 'info', label: '基本信息', component: <ListenerInfo protocolType={protocolType.value} /> },
    ];

    return () => (
      <Tab class='manager-tab-wrap' v-model:active={activeTab.value} type='card-grid'>
        {tabList.map(tab => (
          <TabPanel key={tab.name} name={tab.name} label={tab.label}>
            <div class='common-card-wrap'>{tab.component}</div>
          </TabPanel>
        ))}
      </Tab>
    );
  },
});
