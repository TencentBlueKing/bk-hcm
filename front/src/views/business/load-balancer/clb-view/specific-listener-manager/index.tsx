import { defineComponent, h, ref, watch } from 'vue';
// import components
import DomainList from './domain-list';
import ListenerDetail from './listener-detail';
import TargetGroupView from './target-group/index.vue';
import AddOrUpdateListenerSideslider from '../components/AddOrUpdateListenerSideslider';
// import stores
import { useBusinessStore, useLoadBalancerStore } from '@/store';
// import hooks
import useActiveTab from '@/hooks/useActiveTab';
// import constants
import { ListenerPanelEnum, TRANSPORT_LAYER_LIST } from '@/constants';
import './index.scss';

export default defineComponent({
  name: 'SpecificListenerManager',
  // 导航完成前, 预加载监听器详情以及对应负载均衡详情数据, 并存入store中
  async beforeRouteEnter(to, _, next) {
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();
    // 监听器详情
    const { data: listenerDetail } = await businessStore.detail('listeners', to.params.id as string);
    // 负载均衡详情
    const { data: lbDetail } = await businessStore.detail('load_balancers', listenerDetail.lb_id);
    loadBalancerStore.setCurrentSelectedTreeNode({ ...listenerDetail, lb: lbDetail });
    next();
  },
  props: { id: String, type: String, protocol: String },
  setup(props) {
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();

    const { activeTab, handleActiveTabChange } = useActiveTab(props.type);
    const tabList = ref([]);

    watch(
      () => props.protocol,
      (val) => {
        const isTransportLayer = TRANSPORT_LAYER_LIST.includes(val);
        if (isTransportLayer) {
          // 4层监听器没有下级资源，不显示域名信息
          tabList.value = [
            { name: ListenerPanelEnum.TARGET_GROUP, label: '目标组', component: TargetGroupView },
            { name: ListenerPanelEnum.DETAIL, label: '基本信息', component: ListenerDetail },
          ];
        } else {
          tabList.value = [
            { name: ListenerPanelEnum.LIST, label: '域名', component: DomainList },
            { name: ListenerPanelEnum.DETAIL, label: '基本信息', component: ListenerDetail },
          ];
        }
      },
      { immediate: true },
    );

    const getListenerDetail = async (id: String) => {
      // 监听器详情
      const { data: listenerDetail } = await businessStore.detail('listeners', id as string);
      // 负载均衡详情
      const { data: lbDetail } = await businessStore.detail('load_balancers', listenerDetail.lb_id);
      loadBalancerStore.setCurrentSelectedTreeNode({ ...listenerDetail, lb: lbDetail });
    };

    watch(
      () => props.id,
      (id) => {
        id && getListenerDetail(id);
      },
    );

    return () => (
      <>
        <bk-tab
          class='manager-tab-wrap has-breadcrumb'
          v-model:active={activeTab.value}
          // 这里使用key解决切换监听器后，协议变更而页面不更新问题
          key={props.protocol}
          type='card-grid'
          onChange={handleActiveTabChange}
        >
          {tabList.value.map((tab) => (
            <bk-tab-panel key={tab.name} name={tab.name} label={tab.label}>
              <div class='common-card-wrap'>{h(tab.component, props)}</div>
            </bk-tab-panel>
          ))}
        </bk-tab>
        {/* 编辑监听器 */}
        <AddOrUpdateListenerSideslider originPage='listener' />
      </>
    );
  },
});
