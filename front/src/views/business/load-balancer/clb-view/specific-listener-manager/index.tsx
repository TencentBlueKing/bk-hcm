import { defineComponent, onMounted, onUnmounted, ref, watchEffect } from 'vue';
// import components
import { Tab } from 'bkui-vue';
import DomainList from './domain-list';
import ListenerDetail from './listener-detail';
import AddOrUpdateListenerSideslider from '../components/AddOrUpdateListenerSideslider';
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
  props: { id: String, type: String, protocol: String },
  setup(props) {
    const { activeTab, handleActiveTabChange } = useActiveTab(props.type);
    const tabList = ref([]);

    watchEffect(() => {
      const { protocol } = props;
      const isTransportLayer = TRANSPORT_LAYER_LIST.includes(protocol);
      if (isTransportLayer) {
        // 4层监听器没有下级资源，不显示域名信息
        tabList.value = [{ name: TabTypeEnum.detail, label: '基本信息', component: ListenerDetail }];
      } else {
        tabList.value = [
          { name: TabTypeEnum.list, label: '域名', component: DomainList },
          { name: TabTypeEnum.detail, label: '基本信息', component: ListenerDetail },
        ];
      }
    });

    onMounted(() => {
      // 切换至指定tab
      bus.$on('changeSpecificListenerActiveTab', handleActiveTabChange);
    });

    onUnmounted(() => {
      bus.$off('changeSpecificListenerActiveTab');
    });

    return () => (
      <>
        <Tab
          class='manager-tab-wrap has-breadcrumb'
          v-model:active={activeTab.value}
          type='card-grid'
          onChange={handleActiveTabChange}>
          {tabList.value.map((tab) => (
            <TabPanel key={tab.name} name={tab.name} label={tab.label}>
              <div class='common-card-wrap'>
                <tab.component {...props} />
              </div>
            </TabPanel>
          ))}
        </Tab>
        {/* 编辑监听器 */}
        <AddOrUpdateListenerSideslider />
      </>
    );
  },
});
