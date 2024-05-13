import { defineComponent, ref, watch } from 'vue';
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

    // 面包屑信息
    const breadcrumbs = ref([]);

    // 设置面包屑信息
    const setBreadcrumbs = (routeName: LBRouteName) => {
      // 1. 清空之前的状态
      breadcrumbs.value = [];

      // 2. 根据路由名称, 设置面包屑信息
      const currentNode = loadBalancerStore.currentSelectedTreeNode;
      const { name, lb, lb_id, lbl_id, protocol, port } = currentNode;

      switch (routeName) {
        // 具体的负载均衡
        case LBRouteName.lb:
          breadcrumbs.value.push({ name, extension: getInstVip(currentNode) });
          break;
        // 具体的监听器
        case LBRouteName.listener:
          breadcrumbs.value.push(
            { name: lb.name, extension: getInstVip(lb), linkHandler: getLinkHandler(LBRouteName.lb, lb_id) },
            { name, extension: `${protocol}:${port}` },
          );
          break;
        // 具体的域名
        case LBRouteName.domain:
          breadcrumbs.value.push(
            { name: lb.name, extension: getInstVip(lb), linkHandler: getLinkHandler(LBRouteName.lb, lb_id) },
            {
              name,
              extension: `${protocol}:${port}`,
              linkHandler: getLinkHandler(LBRouteName.listener, lbl_id, { protocol }),
            },
            { name: route.params.id },
          );
          break;
      }

      // 获取链接跳转函数
      function getLinkHandler(routeName: LBRouteName, id: string, extQueryParam = {}) {
        return () => {
          router.push({ name: routeName, params: { id }, query: { ...route.query, ...extQueryParam } });
        };
      }
    };

    watch(
      () => route.params.id,
      (val) => {
        if (!val) return;
        setBreadcrumbs(route.name as LBRouteName);
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='lb-breadcrumb'>
        {breadcrumbs.value.map(({ name, extension, linkHandler }) => (
          <div class='text' onClick={linkHandler || (() => {})}>
            <span class='name'>
              <bk-overflow-title type='tips'>{name}</bk-overflow-title>
            </span>
            {extension && <span class='extension'>{`(${extension})`}</span>}
          </div>
        ))}
      </div>
    );
  },
});
