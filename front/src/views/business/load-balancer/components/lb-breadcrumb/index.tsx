import { defineComponent, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { getLbVip } from '@/utils';
import { LBRouteName } from '@/constants';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerBreadcrumb',
  setup() {
    const router = useRouter();
    const route = useRoute();
    // use stores
    const loadBalancer = useLoadBalancerStore();

    const lbName = ref(''); // 负载均衡器名称
    const lbExtension = ref(''); // 负载均衡器 vip 信息
    const listenerName = ref(''); // 监听器名称
    const listenerExtension = ref(''); // 监听器协议端口
    const domain = ref(''); // 域名

    // 清空文本
    const clearText = () => {
      lbName.value = '';
      lbExtension.value = '';
      listenerName.value = '';
      domain.value = '';
    };

    // 设置当前 listener 所归属的 lb 信息
    const getLBText = (lb: any) => {
      lbName.value = lb.name;
      lbExtension.value = getLbVip(lb);
    };

    // 设置当前 domain 所归属的 listener 信息, 以及对应的 listener 所归属的 lb 信息
    const getFullText = (listener: any) => {
      listenerName.value = listener.name;
      listenerExtension.value = `${listener.protocol}:${listener.port}`;
      getLBText(listener.lb);
    };

    // 跳转至上级页面
    const goListPage = (to: LBRouteName) => {
      return () => {
        // loadBalancer.currentSelectedTreeNode 为监听器详情信息
        const { lb_id, lbl_id, protocol, bk_biz_id } = loadBalancer.currentSelectedTreeNode;
        switch (to) {
          case LBRouteName.lb:
            router.push({ name: LBRouteName.lb, params: { id: lb_id }, query: { bizs: bk_biz_id } });
            break;
          case LBRouteName.listener:
            router.push({ name: LBRouteName.listener, params: { id: lbl_id }, query: { bizs: bk_biz_id, protocol } });
            break;
          default:
            break;
        }
      };
    };

    watch(
      [() => route.name, () => route.params.id],
      ([routeName, id]) => {
        clearText();
        switch (routeName) {
          case 'specific-listener-manager':
            getFullText(loadBalancer.currentSelectedTreeNode);
            break;
          case 'specific-domain-manager':
            domain.value = id as string;
            getFullText(loadBalancer.currentSelectedTreeNode);
            break;
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class='lb-breadcrumb'>
        <div class='text' onClick={goListPage(LBRouteName.lb)}>
          <span class='name'>
            <bk-overflow-title type='tips'>{lbName.value}</bk-overflow-title>
          </span>
          <span class='extension'>{`(${lbExtension.value})`}</span>
        </div>
        <div class='text' onClick={domain.value ? goListPage(LBRouteName.listener) : null}>
          <span class='name'>
            <bk-overflow-title type='tips'>{listenerName.value}</bk-overflow-title>
          </span>
          <span class='extension'>{`(${listenerExtension.value})`}</span>
        </div>
        {domain.value && <div class='text'>{domain.value}</div>}
      </div>
    );
  },
});
