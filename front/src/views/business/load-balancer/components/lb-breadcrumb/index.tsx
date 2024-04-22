import { defineComponent, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerBreadcrumb',
  setup() {
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

    // 获取 lb 的 vip 信息
    const getLBVipText = (data: any) => {
      const { private_ipv4_addresses, private_ipv6_addresses, public_ipv4_addresses, public_ipv6_addresses } = data;
      if (public_ipv4_addresses.length > 0) return public_ipv4_addresses.join(',');
      if (public_ipv6_addresses.length > 0) return public_ipv6_addresses.join(',');
      if (private_ipv4_addresses.length > 0) return private_ipv4_addresses.join(',');
      if (private_ipv6_addresses.length > 0) return private_ipv6_addresses.join(',');
      return '--';
    };

    // 设置当前 listener 所归属的 lb 信息
    const getLBText = (lb: any) => {
      lbName.value = lb.name;
      lbExtension.value = getLBVipText(lb);
    };

    // 设置当前 domain 所归属的 listener 信息, 以及对应的 listener 所归属的 lb 信息
    const getFullText = (listener: any) => {
      listenerName.value = listener.name;
      listenerExtension.value = `${listener.protocol}:${listener.port}`;
      getLBText(listener.lb);
    };

    watch(
      () => route.name,
      (routeName) => {
        clearText();
        switch (routeName) {
          case 'specific-listener-manager':
            getFullText(loadBalancer.currentSelectedTreeNode);
            break;
          case 'specific-domain-manager':
            domain.value = route.params.id as string;
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
        <div class='text'>
          <span class='name'>
            <bk-overflow-title type='tips'>{lbName.value}</bk-overflow-title>
          </span>
          <span class='extension'>{`(${lbExtension.value})`}</span>
        </div>
        <div class='text'>
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
