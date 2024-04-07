import { defineComponent, ref, watch } from 'vue';
// import components
import { Tab } from 'bkui-vue';
import ListenerList from './listener-list';
import TargetGroupDetail from './target-group-detail';
import HealthCheckupPage from './health-checkup';
import './index.scss';
import { useBusinessStore } from '@/store';
import { useRoute } from 'vue-router';

const { TabPanel } = Tab;

type TabType = 'listener' | 'info' | 'health';

export default defineComponent({
  name: 'SpecificTargetGroupManager',
  setup() {
    const businessStore = useBusinessStore();
    const route = useRoute();
    const tgDetail = ref({});
    const activeTab = ref<TabType>('listener');
    const tabList = [
      { name: 'listener', label: '绑定的监听器', component: ListenerList },
      { name: 'info', label: '基本信息', component: TargetGroupDetail },
      { name: 'health', label: '健康检查', component: HealthCheckupPage },
    ];

    const getTargetGroupDetail = async (id: string) => {
      const res = await businessStore.getTargetGroupDetail(id);
      tgDetail.value = res.data;
    };

    watch(
      () => route.query.tgId,
      (id) => {
        if (id) {
          getTargetGroupDetail(id as string);
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <Tab class='manager-tab-wrap' v-model:active={activeTab.value} type='card-grid'>
        {tabList.map((tab) => (
          <TabPanel key={tab.name} name={tab.name} label={tab.label}>
            <div class='common-card-wrap'>
              {<tab.component detail={tgDetail.value} getTargetGroupDetail={getTargetGroupDetail} />}
            </div>
          </TabPanel>
        ))}
      </Tab>
    );
  },
});
