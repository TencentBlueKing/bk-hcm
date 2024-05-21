import { defineComponent, ref } from 'vue';
import { Tab } from 'bkui-vue';
import './index.scss';
import ListenerList from './listener-list';
import TargetGroupDetail from './target-group-detail';
import HealthCheckupPage from './health-checkup';

const { TabPanel } = Tab;

export default defineComponent({
  name: 'SpecificTargetGroupManager',
  setup() {
    const activeTab = ref('listener' as 'listener' | 'info' | 'health');
    const tabList = [
      { name: 'listener', label: '绑定的监听器', component: <ListenerList /> },
      { name: 'info', label: '基本信息', component: <TargetGroupDetail /> },
      { name: 'health', label: '健康检查', component: <HealthCheckupPage /> },
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
