import { defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
// import components
import { Message, Tab } from 'bkui-vue';
import ListenerList from './listener-list';
import ClbDetail from './clb-detail';
import SecurityGroup from './security-group';
// import stores
import { useBusinessStore, useLoadBalancerStore } from '@/store';
// import hooks and utils
import useActiveTab from '@/hooks/useActiveTab';
import bus from '@/common/bus';
import './index.scss';

export enum TypeEnum {
  list = 'list',
  detail = 'detail',
  security = 'security',
}

const { TabPanel } = Tab;

export default defineComponent({
  // 路由导航完成前, 预加载负载均衡详情数据, 并存入store中
  async beforeRouteEnter(to, _, next) {
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();
    const { data } = await businessStore.getLbDetail(to.params.id as string);
    loadBalancerStore.setCurrentSelectedTreeNode(data);
    next();
  },
  props: { id: String, type: String },
  setup(props) {
    // use stores
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();

    const { activeTab, handleActiveTabChange } = useActiveTab(TypeEnum.list);
    const tabList = [
      {
        name: TypeEnum.list,
        label: '监听器',
        component: ListenerList,
      },
      {
        name: TypeEnum.detail,
        label: '基本信息',
        component: ClbDetail,
      },
      {
        name: TypeEnum.security,
        label: '安全组',
        component: SecurityGroup,
      },
    ];

    const detail: { [key: string]: any } = ref(loadBalancerStore.currentSelectedTreeNode);
    const getDetails = async (id: string) => {
      const res = await businessStore.getLbDetail(id);
      detail.value = res.data;
      // 更新一下store
      loadBalancerStore.setCurrentSelectedTreeNode(detail.value);
    };
    const updateLb = async (payload: Record<string, any>) => {
      await businessStore.updateLbDetail(detail.value.vendor, {
        id: detail.value.id,
        ...payload,
      });
      await getDetails(detail.value.id);
      Message({
        message: '更新成功',
        theme: 'success',
      });
    };

    watch(
      () => props.id,
      async (id) => {
        id && (await getDetails(id));
      },
    );

    onMounted(() => {
      bus.$on('changeSpecificClbActiveTab', handleActiveTabChange);
    });

    onUnmounted(() => {
      bus.$off('changeSpecificClbActiveTab');
    });

    return () => (
      <Tab
        v-model:active={activeTab.value}
        type={'card-grid'}
        onChange={handleActiveTabChange}
        class='manager-tab-wrap has-breadcrumb'>
        {tabList.map((tab) => (
          <TabPanel
            key={tab.name}
            name={tab.name}
            label={tab.label}
            class={'clb-list-tab-content-container'}
            renderDirective='if'>
            <div class='common-card-wrap'>
              <tab.component detail={detail.value} getDetails={getDetails} updateLb={updateLb} {...props} />
            </div>
          </TabPanel>
        ))}
      </Tab>
    );
  },
});
