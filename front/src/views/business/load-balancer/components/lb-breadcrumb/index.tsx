import { defineComponent, ref, watch } from 'vue';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useResourceStore } from '@/store';
import './index.scss';
import { Loading } from 'bkui-vue';

export default defineComponent({
  name: 'LoadBalancerBreadcrumb',
  setup() {
    // use stores
    const loadBalancer = useLoadBalancerStore();
    const resourceStore = useResourceStore();

    const isLoading = ref(false);
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

    // 获取当前 listener 所归属的 lb 信息
    const getLBText = async (id: string) => {
      try {
        isLoading.value = true;
        const res = await resourceStore.detail('load_balancers', id);
        lbName.value = res.data.name;
        loadBalancer.setCurrentSelectedTreeNode({
          ...loadBalancer.currentSelectedTreeNode,
          lb: res.data,
        });
        lbExtension.value = getLBVipText(res.data);
      } finally {
        isLoading.value = false;
      }
    };

    // 获取当前 domain 所归属的 listener 信息, 以及对应的 listener 所归属的 lb 信息
    const getFullText = async (id: string) => {
      try {
        isLoading.value = true;
        const res = await resourceStore.detail('listeners', id);
        listenerName.value = res.data.name;
        loadBalancer.setCurrentSelectedTreeNode({
          ...loadBalancer.currentSelectedTreeNode,
          listener: res.data,
        });
        listenerExtension.value = `${res.data.protocol}:${res.data.port}`;
        await getLBText(res.data.lb_id);
      } finally {
        isLoading.value = false;
      }
    };

    watch(
      () => loadBalancer.currentSelectedTreeNode.id,
      () => {
        clearText();
        switch (loadBalancer.currentSelectedTreeNode.type) {
          case 'listener':
            listenerName.value = loadBalancer.currentSelectedTreeNode.name;
            listenerExtension.value = `${loadBalancer.currentSelectedTreeNode.protocol}:${loadBalancer.currentSelectedTreeNode.port}`;
            getLBText(loadBalancer.currentSelectedTreeNode.lb_id);
            break;
          case 'domain':
            domain.value = loadBalancer.currentSelectedTreeNode.domain;
            getFullText(loadBalancer.currentSelectedTreeNode.listener_id);
            break;
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <Loading loading={isLoading.value} opacity={1} color='#f5f7fb' class='lb-breadcrumb'>
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
      </Loading>
    );
  },
});
