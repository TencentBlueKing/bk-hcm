import { defineComponent, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
// import components
import TargetGroupList from './target-group-list';
import SpecificTargetGroupManager from './specific-target-group-manager';
import AllGroupsManager from './all-groups-manager';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import './index.scss';

type ActiveType = 'all' | 'specific';

export default defineComponent({
  name: 'TargetGroupView',
  setup() {
    // use hooks
    const route = useRoute();
    const componentMap = {
      all: <AllGroupsManager />,
      specific: <SpecificTargetGroupManager />,
    };
    const renderComponent = (type: ActiveType) => {
      return componentMap[type];
    };
    const activeType = ref<ActiveType>('all');

    watch(
      () => route.query,
      (val) => {
        const { tgId } = val;
        if (!tgId) return;
        // 如果url中存在tgId, 则表示当前已选中具体的目标组, 页面需要切换至对应目标组的监听器list页面, 并将tgId存入store
        activeType.value = 'specific';
        useLoadBalancerStore().setTargetGroupId(tgId as string);
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='group-view-page'>
        <div class='left-container'>
          <TargetGroupList onChangeActiveType={(type) => (activeType.value = type)} />
        </div>
        <div class='main-container'>{renderComponent(activeType.value)}</div>
      </div>
    );
  },
});
