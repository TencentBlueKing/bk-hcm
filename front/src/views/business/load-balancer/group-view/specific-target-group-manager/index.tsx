import { defineComponent, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ITargetGroupDetail, useBusinessStore, useLoadBalancerStore } from '@/store';
import useActiveTab from '@/hooks/useActiveTab';

import { Tab } from 'bkui-vue';
import ListenerList from './listener-list';
import TargetGroupDetail from './target-group-detail';
import HealthCheckupPage from './health-checkup';
import './index.scss';

const { TabPanel } = Tab;

enum TabType {
  list = 'list',
  detail = 'detail',
  health = 'health',
}

export default defineComponent({
  name: 'SpecificTargetGroupManager',
  setup() {
    const route = useRoute();

    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();
    const tgDetail = ref<Partial<ITargetGroupDetail>>({});
    const { activeTab, handleActiveTabChange } = useActiveTab(TabType.list);
    const tabList = [
      { name: TabType.list, label: '绑定的监听器', component: ListenerList },
      { name: TabType.detail, label: '基本信息', component: TargetGroupDetail },
      { name: TabType.health, label: '健康检查', component: HealthCheckupPage },
    ];

    const getTargetGroupDetail = async (id: string) => {
      const res = await businessStore.getTargetGroupDetail(id);
      res.data.target_list = res.data.target_list.map((item: any) => {
        item.region = item.zone.slice(0, item.zone.lastIndexOf('-'));
        return item;
      });
      tgDetail.value = res.data;
    };

    const getListenerDetail = async () => {
      const { id, vendor } = tgDetail.value;
      // 请求绑定的监听器规则
      const rulesRes = await businessStore.list(
        {
          page: { limit: 1, start: 0, count: false },
          filter: { op: 'and', rules: [] },
        },
        `vendors/${vendor}/target_groups/${id}/rules`,
      );
      const listenerItem = rulesRes.data.details[0];
      if (!listenerItem) return;
      // 请求监听器详情, 获取端口段信息
      const detailRes = await businessStore.detail('listeners', listenerItem.lbl_id);
      loadBalancerStore.setListenerDetailWithTargetGroup(detailRes.data);
    };

    watch(
      () => route.params.id,
      async (id) => {
        if (!id) return;
        // 目标组id状态保持
        const targetGroupId = id as string;
        loadBalancerStore.setTargetGroupId(targetGroupId);
        await getTargetGroupDetail(targetGroupId);
        await getListenerDetail();
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='specific-target-group-manager'>
        <Tab
          class='manager-tab-wrap'
          v-model:active={activeTab.value}
          type='card-grid'
          onChange={handleActiveTabChange}>
          {tabList.map((tab) => (
            <TabPanel key={tab.name} name={tab.name} label={tab.label}>
              <tab.component
                detail={tgDetail.value}
                getTargetGroupDetail={getTargetGroupDetail}
                id={route.params.id as string}
              />
            </TabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
