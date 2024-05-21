import { defineComponent, ref } from 'vue';
import { Tab } from 'bkui-vue';
import './index.scss';
import DomainList from './domain-list';
import ListenerDetail from './listener-detail';

const { TabPanel } = Tab;

export default defineComponent({
  name: 'SpecificListenerManager',
  setup() {
    const activeTab = ref('domain' as 'domain | info');
    const protocolType = ref('UDP' as 'HTTP' | 'HTTPS' | 'TCP' | 'UDP');
    const tabList = [
      { name: 'domain', label: '域名', component: <DomainList /> },
      { name: 'info', label: '基本信息', component: <ListenerDetail protocolType={protocolType.value} /> },
    ];

    return () => (
      <Tab class='manager-tab-wrap' v-model:active={activeTab.value} type='card-grid'>
        {tabList.map((tab) => (
          <TabPanel key={tab.name} name={tab.name} label={tab.label}>
            <div class='common-card-wrap'>{tab.component}</div>
          </TabPanel>
        ))}
      </Tab>
    );
  },
});
