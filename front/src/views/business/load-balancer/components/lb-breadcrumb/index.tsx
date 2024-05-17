import { defineComponent } from 'vue';
import { useRoute, useRouter } from 'vue-router';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { getInstVip } from '@/utils';
import { LBRouteName } from '@/constants';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerBreadcrumb',
  setup() {
    const router = useRouter();
    const route = useRoute();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    // 获取链接跳转函数
    function getLinkHandler(routeName: LBRouteName, id: string, extQueryParam = {}) {
      return () => {
        router.push({ name: routeName, params: { id }, query: { ...route.query, ...extQueryParam } });
      };
    }

    // click-handler - 跳转至具体的负载均衡
    const handleClickLb = () => {
      return route.name !== LBRouteName.lb
        ? getLinkHandler(LBRouteName.lb, loadBalancerStore.currentSelectedTreeNode.lb_id)
        : null;
    };

    // click-handler - 跳转至具体的监听器
    const handleClickListener = () => {
      return route.name === LBRouteName.domain
        ? getLinkHandler(LBRouteName.listener, loadBalancerStore.currentSelectedTreeNode.lbl_id, {
            protocol: loadBalancerStore.currentSelectedTreeNode.protocol,
          })
        : null;
    };

    return () => (
      <div class='lb-breadcrumb'>
        {/* 负载均衡 */}
        <div class='text' onClick={handleClickLb()}>
          <span class='name'>
            <bk-overflow-title type='tips'>
              {route.name === LBRouteName.lb
                ? loadBalancerStore.currentSelectedTreeNode.name
                : loadBalancerStore.currentSelectedTreeNode.lb?.name}
            </bk-overflow-title>
          </span>
          <span class='extension'>{`(${getInstVip(loadBalancerStore.currentSelectedTreeNode)})`}</span>
        </div>
        {/* 监听器 */}
        {route.name !== LBRouteName.lb && (
          <div class='text' onClick={handleClickListener()}>
            <span class='name'>
              <bk-overflow-title type='tips'>{loadBalancerStore.currentSelectedTreeNode.name}</bk-overflow-title>
            </span>
            <span class='extension'>{`(${loadBalancerStore.currentSelectedTreeNode.protocol}:${loadBalancerStore.currentSelectedTreeNode.port})`}</span>
          </div>
        )}
        {/* 域名 */}
        {route.name === LBRouteName.domain && (
          <div class='text'>
            <span class='name'>
              <bk-overflow-title type='tips'>{route.params.id}</bk-overflow-title>
            </span>
          </div>
        )}
      </div>
    );
  },
});
