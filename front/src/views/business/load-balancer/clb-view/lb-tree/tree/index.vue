<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, useTemplateRef, inject, ComputedRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Loading, Tree } from 'bkui-vue';
import ITreeNode from './node/index.vue';
import ITreeNodeAction from './node-action/index.vue';

import allLbIcon from '@/assets/image/all-lb.svg';
import lbIcon from '@/assets/image/loadbalancer.svg';
import listenerIcon from '@/assets/image/listener.svg';
import domainIcon from '@/assets/image/domain.svg';

import { useBusinessStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useMoreActionDropdown from '@/hooks/useMoreActionDropdown';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';

import { asyncGetListenerCount, getInstVip } from '@/utils';
import { LB_ROUTE_NAME_MAP, LBRouteName, ListenerPanelEnum, TRANSPORT_LAYER_LIST } from '@/constants';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import type {
  Domain,
  DomainListResData,
  Listener,
  ListenerListResData,
  LoadBalancer,
  LoadBalancerListResData,
  LoadBalancers,
  ResourceType,
} from '../types';
import http from '@/http';
import bus from '@/common/bus';

const prefixIconMap = {
  all: allLbIcon,
  lb: lbIcon,
  listener: listenerIcon,
  domain: domainIcon,
};

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const businessStore = useBusinessStore();
const { getBusinessApiPath } = useWhereAmI();
const globalPermissionDialogStore = useGlobalPermissionDialog();
const { authVerifyData, handleAuth } = useVerify();
const createClbActionName: ComputedRef<'biz_clb_resource_create' | 'clb_resource_create'> =
  inject('createClbActionName');

const treeData = ref([]);
const pagination = reactive({ start: 0, count: 0, loading: false });

/**
 * 获取负载均衡列表
 * @param rules 过滤条件
 */
const getLoadBalancerList = async (rules: RulesItem[] = []): Promise<[LoadBalancers, number]> => {
  // 请求接口：负载均衡列表
  const [detailsRes, countRes] = await Promise.all<LoadBalancerListResData>(
    [false, true].map((isCount) =>
      http.post(`/api/v1/cloud/${getBusinessApiPath()}load_balancers/with/delete_protection/list`, {
        filter: { op: QueryRuleOPEnum.AND, rules },
        page: { count: isCount, start: isCount ? 0 : pagination.start, limit: isCount ? 0 : 50 },
      }),
    ),
  );

  if (!detailsRes.data.details) return [[], 0];

  // 请求接口：各负载均衡下的监听器数量
  detailsRes.data.details = await asyncGetListenerCount(businessStore.asyncGetListenerCount, detailsRes.data.details);

  // 组装数据
  const increment = detailsRes.data.details.map((item: LoadBalancer) => {
    item.nodeKey = item.cloud_id;
    item.type = 'lb';
    item.displayValue = `${item.name} (${getInstVip(item)})`;
    item.async = true;
    item.start = 0;
    item.count = item.listenerNum ?? 0;
    return item;
  });

  return [increment, countRes.data.count];
};
const loadLoadBalancerList = async (rules: RulesItem[] = []) => {
  pagination.loading = true;
  try {
    const [details, count] = await getLoadBalancerList(rules);
    // 更新数据
    treeData.value = [...treeData.value, ...details];
    pagination.count = count;
  } catch (error) {
    reset();
  } finally {
    pagination.loading = false;
  }
};
// 重置数据
const reset = () => {
  treeData.value = [];
  Object.assign(pagination, { start: 0, count: 0 });
};

/**
 * 根据负载均衡ID获取负载均衡下的监听器列表
 *
 * *TCP、UDP无下级资源, 不需要请求，可通过 async 配置进行限制。
 * @param lb 负载均衡节点数据
 */
const getListenerList = async (lb: LoadBalancer) => {
  const { id, start, count } = lb;
  if (count === 0) return;

  // 请求接口：负载均衡下的监听器列表（监听器总数由异步接口返回，此处不再更新count）
  const detailsRes: ListenerListResData = await http.post(
    `/api/v1/cloud/${getBusinessApiPath()}load_balancers/${id}/listeners/list`,
    {
      filter: { op: QueryRuleOPEnum.AND, rules: [] },
      page: { count: false, start, limit: 50 },
    },
  );

  if (!detailsRes.data.details) return [];

  // 组装数据
  const increment = detailsRes.data.details.map((item: Listener) => {
    const { cloud_id, name, protocol, port, end_port, domain_num } = item;
    item.nodeKey = cloud_id;
    item.type = 'listener';
    item.displayValue = `${name} (${protocol}:${port}${end_port ? `-${end_port}` : ''})`;
    // 无需异步加载：4层监听器没有下级资源；域名数量为0；
    item.async = !(TRANSPORT_LAYER_LIST.includes(protocol) || domain_num === 0);
    item.count = domain_num;
    return item;
  });

  // 更新数据
  return increment;
};

/**
 * 根据监听器ID获取监听器下的域名列表
 * @param listener 监听器节点数据
 */
const getDomainList = async (listener: Listener) => {
  const { id, vendor } = listener;

  // 请求接口：域名列表
  const res: DomainListResData = await http.post(
    `/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/listeners/${id}/domains/list`,
  );

  // 组装数据
  const { default_domain, domain_list } = res.data;
  const increment = domain_list.map((item: Domain) => {
    const { domain, url_count } = item;
    item.nodeKey = listener.cloud_id + domain; // 接口未提供cloud_id，此处手动拼接监听器的cloud_id与域名
    item.type = 'domain';
    item.displayValue = domain;
    item.count = url_count;
    item.id = domain;
    item.listener_id = id;
    item.isDefault = default_domain === domain;
    item.vendor = vendor;
    return item;
  });

  // 更新数据
  return increment;
};

const asyncCallback = async (node: any) => {
  const { type } = node as ResourceType;
  if (type === 'lb') return getListenerList(node as LoadBalancer) as Promise<any>;
  if (type === 'listener') return getDomainList(node as Listener) as Promise<any>;
};

const treeRef = useTemplateRef<typeof Tree>('tree');
const allLBNode = { type: 'all', isDropdownListShow: false, id: '-1' };
const lastSelectedNode = ref(); // 记录上一次选中的tree-node, 不包括全部负载均衡

// util-路由切换
const pushState = (node: any) => {
  // util-计算tab类型
  const getTabType = (nodeType: string, protocol: string | undefined) => {
    // 节点类型为lb, listener时, 需要设置query参数(type)
    if (['lb', 'listener'].includes(nodeType)) {
      // 记录当前url上的query参数(type)
      const tabType = route.query.type;
      const lastNodeType = lastSelectedNode.value?.type;
      // 1. 如果当前点击节点的类型与上一次一样, 则返回上一次的tab类型；否则，默认为list
      let resultType = lastNodeType === nodeType ? tabType : ListenerPanelEnum.LIST;
      // 2. 如果当前节点类型为listener, 且为四层协议, 返回target_group
      if (nodeType === 'listener' && TRANSPORT_LAYER_LIST.includes(protocol)) {
        resultType = ListenerPanelEnum.TARGET_GROUP;
      }
      return resultType;
    }
    // 其他情况, 不需要设置tab类型
    return undefined;
  };
  router.push({
    name: LB_ROUTE_NAME_MAP[node.type],
    params: { id: node.id },
    query: {
      ...route.query,
      // 设置tab类型标识(node.protocol只有listener有值)
      type: getTabType(node.type, node.protocol),
      // 如果节点类型为listener, 则设置protocol标识
      protocol: node.type === 'listener' ? node.protocol : undefined,
      // 如果节点类型为domain, 则设置listener_id
      listener_id: node.type === 'domain' ? node.listener_id : undefined,
      vendor: node.vendor,
    },
  });
};
// define handler function - 节点点击
const handleNodeClick = (node: any) => {
  // 切换四级路由组件
  pushState(node);
  // 交互 - 高亮切换效果
  if (node.type !== 'all') {
    lastSelectedNode.value = node;
  } else {
    treeRef.value.setSelect(lastSelectedNode.value, false);
  }
};

// more-action - type 与 dropdown menu 的映射关系
const typeMenuMap = {
  all: [
    {
      label: '购买负载均衡',
      handler: () => {
        if (!authVerifyData?.value?.permissionAction?.[createClbActionName.value]) {
          handleAuth(createClbActionName.value);
          globalPermissionDialogStore.setShow(true);
          return;
        }
        router.push({ path: '/business/service/service-apply/clb' });
      },
      preAuth: () => authVerifyData?.value?.permissionAction?.[createClbActionName.value],
    },
  ],
  lb: [{ label: '新增监听器', handler: () => bus.$emit('showAddListenerSideslider') }],
  listener: [
    { label: '新增域名', handler: () => bus.$emit('showAddDomainSideslider') },
    { label: '编辑', handler: ({ id }: any) => bus.$emit('showEditListenerSideslider', id) },
  ],
  domain: [
    { label: '新增 URL 路径', handler: () => bus.$emit('showAddUrlSideslider') },
    { label: '编辑', handler: (node: any) => bus.$emit('showAddDomainSideslider', node) },
  ],
};
const { showDropdownList, currentPopBoundaryNodeKey } = useMoreActionDropdown(typeMenuMap);

// 滚动加载 - 节点进入父容器可视区域时执行的回调
const intersectionObserverCallback = async (args: any) => {
  if (!args) return;
  const { index, node, parent } = args;
  const { type } = node as ResourceType;
  if (
    type === 'lb' &&
    index === treeData.value.length - 1 &&
    treeData.value.length < pagination.count &&
    !pagination.loading
  ) {
    try {
      // 加载下一页的负载均衡列表
      pagination.start += 50;
      treeData.value.push({ type: 'loading' });
      await loadLoadBalancerList();
      treeData.value = treeData.value.filter((item: any) => item.type !== 'loading');
    } finally {
      pagination.loading = false;
    }
  } else if (
    type === 'listener' &&
    index === parent.children.length - 1 &&
    parent.children.length < parent.count &&
    !parent.loading
  ) {
    // 加载下一页的监听器列表
    try {
      parent.start += 50;
      parent.loading = true;
      parent.children.push({ type: 'loading' });
      const increment = await getListenerList(parent);
      parent.children.push(...increment);
      parent.children = parent.children.filter((item: any) => item.type !== 'loading');
    } finally {
      parent.loading = false;
    }
  }
};

// 搜索高亮
const searchValue = ref('');
const search = (rules: RulesItem[], value: string) => {
  searchValue.value = value;
  reset();
  loadLoadBalancerList(rules);
};

onMounted(() => {
  // 组件挂载，加载第一页负载均衡列表
  loadLoadBalancerList();

  bus.$on('resetLbTree', (rules: RulesItem[]) => {
    reset();
    loadLoadBalancerList(rules);
  });
});

onUnmounted(() => {
  bus.$off('resetLbTree');
});

defineExpose({ search });
</script>

<template>
  <div class="lb-tree-container">
    <!-- 全部负载均衡 -->
    <ITreeNode
      :display-value="t('全部负载均衡')"
      :count="pagination.count"
      class="all-lbs-node"
      :class="{ 'is-selected': route.name === LBRouteName.allLbs, 'show-dropdown': currentPopBoundaryNodeKey === '-1' }"
      :handle-more-action-click="(e) => showDropdownList(e, allLBNode)"
      @click="handleNodeClick(allLBNode)"
    >
      <template #prefix-icon>
        <img :src="allLbIcon" alt="" style="height: 20px; width: 20px; margin-right: 8px" />
      </template>
    </ITreeNode>
    <!-- 负载均衡树 -->
    <Loading :loading="pagination.loading" class="lb-tree">
      <Tree
        ref="tree"
        node-key="nodeKey"
        label="name"
        :indent="16"
        :offset-left="16"
        :line-height="36"
        :level-line="(node: ResourceType) => node?.type === 'loading' ? 'none' : '1px dashed #c3cdd7'"
        :data="treeData"
        :async="{ cache: true, callback: asyncCallback }"
        :intersection-observer="{ enabled: true, callback: intersectionObserverCallback }"
        @node-click="handleNodeClick"
      >
        <template #nodeAction="node">
          <ITreeNodeAction :node="node" />
        </template>
        <template #nodeType="node">
          <img v-if="node.type !== 'loading'" :src="prefixIconMap[node.type as keyof typeof prefixIconMap]" alt="" />
        </template>
        <template #default="{ data }">
          <!-- loading节点 -->
          <Loading v-if="data.type === 'loading'" class="loading-node" loading size="small">
            <div style="height: 36px"></div>
          </Loading>
          <!-- 资源节点 -->
          <ITreeNode
            v-else
            :class="{ 'show-dropdown': currentPopBoundaryNodeKey === data.nodeKey }"
            :display-value="data.displayValue"
            :count="data.count"
            :handle-more-action-click="(e) => showDropdownList(e, data)"
            :no-count="data.type === 'listener' && TRANSPORT_LAYER_LIST.includes(data.protocol)"
            :show-default-domain-tag="data.type === 'domain' && data.isDefault"
            :search-value="searchValue"
          />
        </template>
      </Tree>
    </Loading>
  </div>
</template>

<style scoped lang="scss">
.lb-tree-container {
  height: calc(100% - 56px);

  .all-lbs-node {
    padding-left: 16px;
    box-shadow: 0 -1px 0 0 #eaebf0, 0 1px 0 0 #eaebf0;

    &.is-selected {
      background-color: #e1ecff !important;
    }

    &:hover {
      background-color: #f0f1f5;
    }
  }

  :deep(.bk-node-row) {
    .bk-node-content img {
      width: 20px;
      height: 20px;
      margin-right: 8px;
    }

    &:hover {
      background-color: #f0f1f5;
    }

    &.is-selected {
      background-color: #e1ecff;
    }
  }

  .lb-tree {
    height: calc(100% - 36px) !important;

    .loading-node {
      position: relative;
      left: -16px;
    }
  }

  .show-dropdown {
    :deep(.suffix) {
      .count {
        opacity: 0;
      }

      .more-action {
        background-color: #dcdee5;
        opacity: 1;
      }
    }
  }
}
</style>
