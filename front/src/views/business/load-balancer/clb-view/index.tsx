import { defineComponent, ref } from 'vue';
// import components
import LbTree from './lb-tree';
import AllClbsManager from './all-clbs-manager';
import SpecificListenerManager from './specific-listener-manager';
import SpecificClbManager from './specific-clb-manager';
import SpecificDomainManager from './specific-domain-manager';
import './index.scss';

type NodeType = 'all' | 'lb' | 'listener' | 'domain';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const isAdvancedSearchShow = ref(false);

    const componentMap = {
      all: <AllClbsManager />,
      lb: <SpecificClbManager />,
      listener: <SpecificListenerManager />,
      domain: <SpecificDomainManager />,
    };
    const renderComponent = (type: NodeType) => {
      return componentMap[type];
    };

    const activeType = ref<NodeType>('all');

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <LbTree v-model:activeType={activeType.value} />
        </div>
        {isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>}
        <div class='main-container'>
          {/* 
            面包屑技术设计:
              1. 这里引入面包屑组件
              2. 当 lb-tree 触发 node-click 事件时, 如果 type 不为 'lb', 则组装面包屑内容. 使用 bus 进行通信
                面包屑组件    bus.$on('showBreadcrumb', (...names) => { 组装面包屑内容 })
                lb-tree组件  emit('showBreadcrumb', names)
          */}
          {renderComponent(activeType.value)}
        </div>
      </div>
    );
  },
});
