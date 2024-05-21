import { computed, defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
// import components
import { PlayShape } from 'bkui-vue/lib/icon';
import { Loading, Message, OverflowTitle, Tree } from 'bkui-vue';
import SimpleSearchSelect from '../../components/simple-search-select';
import Confirm from '@/components/confirm';
// import stores
import { useBusinessStore, useLoadBalancerStore } from '@/store';
// import custom hooks
import { useI18n } from 'vue-i18n';
import useLoadTreeData from './useLoadTreeData';
import useMoreActionDropdown from '@/hooks/useMoreActionDropdown';
// import utils
import { throttle } from 'lodash';
import bus from '@/common/bus';
import { getInstVip } from '@/utils';
// import static resources
import allLBIcon from '@/assets/image/all-lb.svg';
import lbIcon from '@/assets/image/loadbalancer.svg';
import listenerIcon from '@/assets/image/listener.svg';
import domainIcon from '@/assets/image/domain.svg';
// import constants
import { LBRouteName, LB_ROUTE_NAME_MAP, TRANSPORT_LAYER_LIST } from '@/constants';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerTree',
  setup() {
    // use hooks
    const { t } = useI18n();
    const router = useRouter();
    const route = useRoute();
    // use stores
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();

    // 搜索相关
    const searchValue = ref('');
    const searchDataList = [
      { id: 'lb_name', name: '负载均衡名称' },
      { id: 'lb_vip', name: '负载均衡VIP' },
      { id: 'listener_name', name: '监听器名称' },
      { id: 'protocol', name: '协议' },
      { id: 'port', name: '端口' },
      { id: 'domain', name: '域名' },
    ];

    // lb-tree相关
    const treeData = ref([]);
    const treeRef = ref();
    const allLBNode = { type: 'all', isDropdownListShow: false, id: '-1' };
    const lastSelectedNode = ref(); // 记录上一次选中的tree-node, 不包括全部负载均衡
    const loadingRef = ref();
    const expandedNodeArr = ref([]);

    // use custom hooks
    const { loadRemoteData, handleLoadDataByScroll, reset, isLoading } = useLoadTreeData(treeData);

    // 删除负载均衡
    const handleDeleteLB = (node: any) => {
      const { id, name } = node;
      Confirm('请确定删除负载均衡', `将删除负载均衡【${name}】`, () => {
        businessStore.deleteBatch('load_balancers', { ids: [id] }).then(() => {
          Message({ theme: 'success', message: '删除成功' });
          // 本期暂时先重新拉取lb列表
          reset();
          // 导航至全部负载均衡
          router.push({ name: LBRouteName.allLbs, query: { bizs: route.query.bizs } });
        });
      });
    };

    // 删除监听器
    const handleDeleteListener = (node: any) => {
      const { id, name } = node;
      Confirm('请确定删除监听器', `将删除监听器【${name}】`, () => {
        businessStore.deleteBatch('listeners', { ids: [id] }).then(() => {
          Message({ theme: 'success', message: '删除成功' });
          // 本期暂时先重新拉取lb列表
          reset();
          // 导航至全部负载均衡
          router.push({ name: LBRouteName.allLbs, query: { bizs: route.query.bizs } });
        });
      });
    };

    // 删除域名
    const handleDeleteDomain = (node: any) => {
      const { listener_id, domain } = node;
      Confirm('请确定删除域名', `将删除域名【${domain}】`, async () => {
        await businessStore.batchDeleteDomains({ lbl_id: listener_id, domains: [domain] });
        Message({ theme: 'success', message: '删除成功' });
        // 本期暂时先重新拉取lb列表
        reset();
        // 导航至全部负载均衡
        router.push({ name: LBRouteName.allLbs, query: { bizs: route.query.bizs } });
      });
    };

    // type 与 dropdown menu 的映射关系
    const typeMenuMap = {
      all: [
        {
          label: '购买负载均衡',
          handler: () => router.push({ path: '/business/service/service-apply/clb' }),
        },
      ],
      lb: [
        { label: '新增监听器', handler: () => bus.$emit('showAddListenerSideslider') },
        {
          label: '删除',
          handler: handleDeleteLB,
          isDisabled: (item: any) => item.listenerNum > 0 || item.delete_protect,
          tooltips: (item: any) => {
            if (item.listenerNum > 0) {
              return { content: t('该负载均衡已绑定监听器, 不可删除'), disabled: !(item.listenerNum > 0) };
            }
            if (item.delete_protect) {
              return { content: t('该负载均衡已开启删除保护, 不可删除'), disabled: !item.delete_protect };
            }
          },
        },
      ],
      listener: [
        { label: '新增域名', handler: () => bus.$emit('showAddDomainSideslider') },
        { label: '编辑', handler: ({ id }: any) => bus.$emit('showEditListenerSideslider', id) },
        { label: '删除', handler: handleDeleteListener },
      ],
      domain: [
        { label: '新增 URL 路径', handler: () => bus.$emit('showAddUrlSideslider') },
        { label: '编辑', handler: (node: any) => bus.$emit('showAddDomainSideslider', node) },
        { label: '删除', handler: handleDeleteDomain },
      ],
    };
    const { showDropdownList, currentPopBoundaryNodeKey } = useMoreActionDropdown(typeMenuMap);

    const searchK = ref('');
    const searchV = ref('');
    const selectedNode = ref(null);
    const searchResultCount = ref(0);
    const searchOption = computed(() => {
      // searchOption 重新计算时, 先恢复初始状态
      searchResultCount.value = 0;
      isLoading.value = true;

      return {
        value: searchValue.value,
        match: (searchValue: string, itemText: string, item: any) => {
          [searchK.value, searchV.value] = searchValue.split('：');
          let result = false;
          switch (searchK.value) {
            case 'lb_name':
              item.type === 'lb' && (result = new RegExp(searchV.value, 'g').test(itemText));
              break;
            case 'lb_vip':
              item.type === 'lb' && (result = new RegExp(searchV.value, 'g').test(getInstVip(item)));
              break;
            case 'listener_name':
              item.type === 'listener' && (result = new RegExp(searchV.value, 'g').test(itemText));
              break;
            case 'protocol':
              item.type === 'listener' && (result = new RegExp(searchV.value, 'g').test(item.protocol));
              break;
            case 'port':
              item.type === 'listener' && (result = new RegExp(searchV.value, 'g').test(item.port));
              break;
            case 'domain':
              item.type === 'domain' && (result = new RegExp(`${searchV.value}`, 'g').test(itemText));
              break;
            default:
              break;
          }
          result && (searchResultCount.value += 1);
          // 关闭 loading
          isLoading.value = false;
          return result;
        },
        showChildNodes: true,
      };
    });

    watch(
      () => loadBalancerStore.lbTreeSearchTarget,
      async (val) => {
        // 用搜索结果替换treeData, data为组装后的treeData, targetNode为要选中的节点
        const updateTreeData = (data: any, targetNode: any, searchK: string, searchV: string) => {
          treeData.value = [data];
          selectedNode.value = targetNode;
          searchValue.value = `${searchK}：${searchV}`;
        };

        if (val) {
          const { searchK, searchV, type } = val;
          // 如果点击的是负载均衡, 则直接将搜索结果作为treeData
          if (type === 'lb') {
            const lbNode = { ...loadBalancerStore.lbTreeSearchTarget, type: 'lb', async: true };
            updateTreeData(lbNode, lbNode, searchK, searchV);
          }
          // 如果点击的是监听器, 则需要先构建监听器的负载均衡节点, 再将监听器节点作为 children 添加到负载均衡节点
          else if (type === 'listener') {
            const lbRes = await businessStore.detail('load_balancers', loadBalancerStore.lbTreeSearchTarget.lb_id);
            const listenerNode = { ...loadBalancerStore.lbTreeSearchTarget, async: true };
            updateTreeData(
              { ...lbRes.data, type: 'lb', listenerNum: 1, async: true, children: [listenerNode] },
              listenerNode,
              searchK,
              searchV,
            );
          }
          // 如果点击的是域名, 则需要先构建域名的负载均衡以及监听器节点, 再将监听器节点作为 children 添加上去
          else {
            const { domain, lbl_id } = loadBalancerStore.lbTreeSearchTarget;
            const listenerRes = await businessStore.detail('listeners', lbl_id);
            const lbRes = await businessStore.detail('load_balancers', listenerRes.data.lb_id);
            const domainNode = { ...loadBalancerStore.lbTreeSearchTarget, id: domain, name: domain };
            updateTreeData(
              [
                {
                  ...lbRes.data,
                  type: 'lb',
                  listenerNum: 1,
                  async: true,
                  children: [{ ...listenerRes.data, type: 'listener', async: true, children: [domainNode] }],
                },
              ],
              domainNode,
              searchK,
              searchV,
            );
          }
        } else {
          searchValue.value = '';
        }
      },
      {
        immediate: true,
      },
    );

    // Intersection Observer 监听器
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          // 触发 loadingRef 身上的 loadDataByScroll 自定义事件
          loadingRef.value.$emit('loadDataByScroll');
        }
      });
    });

    // type 与 icon 的映射关系
    const typeIconMap = {
      lb: lbIcon,
      listener: listenerIcon,
      domain: domainIcon,
    };

    // generator函数 - 滚动加载函数
    const getTreeScrollFunc = () => {
      return throttle(() => {
        loadingRef.value && observer.observe(loadingRef.value.$el);

        // // 记录当前是否滚动了一屏的高度
        // const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
        // if (treeRef.value.$el.scrollTop >= viewportHeight) {
        //   isScrollOnePageHeight.value = true;
        // } else {
        //   isScrollOnePageHeight.value = false;
        // }
      }, 200);
    };

    // generator函数 - lb-tree 懒加载配置对象
    const getTreeAsyncOption = () => {
      return {
        callback: (_item: any, _callback: Function, _schema: any) => {
          // 如果是4层监听器, 无需加载其下级资源
          if (_item.type === 'listener' && TRANSPORT_LAYER_LIST.includes(_item.protocol)) return;
          // 异步加载当前点击节点的 children node
          loadRemoteData(_item, _schema.fullPath.split('-').length - 1);
        },
        cache: true,
      };
    };

    // util-路由切换
    const pushState = (node: any) => {
      // util-计算tab类型
      const getTabType = (nodeType: string, protocol: string | undefined) => {
        // 节点类型为lb, listener时, 需要设置query参数(type)
        if (['lb', 'listener'].includes(nodeType)) {
          // 记录当前url上的query参数(type)
          const tabType = route.query.type;
          const lastNodeType = lastSelectedNode.value?.type;
          // 1. tabType无值或者当前点击节点的类型与上一次不一样, 则赋初始值
          if (!tabType || lastNodeType !== nodeType) return 'list';
          // 2. 如果当前节点类型为listener, 且为四层协议, 则直接显示详情
          if (nodeType === 'listener' && TRANSPORT_LAYER_LIST.includes(protocol)) return 'detail';
          // 3. 如果当前点击节点的类型与上一次一样, 则返回上一次的tab类型
          if (lastNodeType === nodeType) return tabType;
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
        },
      });
    };

    // define handler function - 节点点击
    const handleNodeClick = (node: any) => {
      // 防止重复点击
      if (route.params.id === node.id) return;
      // 切换四级路由组件
      pushState(node);
      // 交互 - 高亮切换效果
      if (node.type !== 'all') {
        lastSelectedNode.value = node;
      } else {
        treeRef.value.setSelect(lastSelectedNode.value, false);
      }
    };

    // define handler function - 节点展开
    const handleNodeExpand = (node: any) => {
      expandedNodeArr.value.push(node);
    };

    // define handler function - 节点折叠
    const handleNodeCollapse = (node: any) => {
      const idx = expandedNodeArr.value.findIndex((item) => item === node);
      expandedNodeArr.value.splice(idx, 1);
    };

    onMounted(() => {
      // 组件挂载，加载 root node
      loadRemoteData(null, 0);
    });

    onMounted(() => {
      // 重新加载lb-tree数据
      bus.$on('resetLbTree', reset);
    });

    onUnmounted(() => {
      bus.$off('resetLbTree');
      loadBalancerStore.setLbTreeSearchTarget(null);
    });

    return () => (
      <div class='load-balancer-tree'>
        {/* 搜索 */}
        <SimpleSearchSelect
          v-model={searchValue.value}
          dataList={searchDataList}
          clearHandler={() => {
            loadBalancerStore.setLbTreeSearchTarget(null);
            reset();
          }}
        />
        <Loading class='lb-tree-container' loading={isLoading.value} opacity={1}>
          {/* 全部负载均衡 / 搜索结果 */}
          {(function () {
            if (searchValue.value) {
              if (searchResultCount.value) {
                return <div class='search-result-wrap'>共 {searchResultCount.value} 条搜索结果</div>;
              }
            } else {
              return (
                <div
                  class={[
                    'all-lb-item',
                    `${route.meta.type === 'all' ? ' selected' : ''}`,
                    `${currentPopBoundaryNodeKey.value === '-1' ? ' show-dropdown' : ''}`,
                  ]}
                  onClick={() => handleNodeClick(allLBNode)}>
                  <div class='base-info'>
                    <img src={allLBIcon} alt='' class='prefix-icon' />
                    <span class='text'>全部负载均衡</span>
                  </div>
                  <div class='ext-info'>
                    <div class='count'>{treeData.value.length}</div>
                    <div class='more-action' onClick={(e) => showDropdownList(e, allLBNode)}>
                      <i class='hcm-icon bkhcm-icon-more-fill'></i>
                    </div>
                  </div>
                </div>
              );
            }
          })()}
          {/* lb-tree */}
          <Tree
            class='lb-tree'
            node-key='id'
            ref={treeRef}
            data={treeData.value}
            label='name'
            children='children'
            level-line
            // virtual-render
            indent={16}
            line-height={36}
            onNodeClick={handleNodeClick}
            onScroll={getTreeScrollFunc()}
            async={getTreeAsyncOption()}
            onNodeExpand={handleNodeExpand}
            onNodeCollapse={handleNodeCollapse}
            search={searchOption.value}
            selected={selectedNode.value}>
            {{
              default: ({ data, attributes }: any) => {
                if (data.type === 'loading') {
                  return (
                    <bk-loading
                      class='tree-loading-node'
                      ref={loadingRef}
                      loading
                      size='small'
                      onLoadDataByScroll={() => {
                        // 因为在标签上使用 data-xxx 会丢失引用，但我需要 data._parent 的引用（因为加载数据时会直接操作该对象），所以这里借用了闭包的特性。
                        handleLoadDataByScroll(data, attributes);
                      }}>
                      <div style={{ height: '36px' }}></div>
                    </bk-loading>
                  );
                }
                const { type, id, name, protocol, port, isDefault, listenerNum, domain_num, url_count } = data;
                const extension =
                  // eslint-disable-next-line no-nested-ternary
                  type === 'lb' ? ` (${getInstVip(data)})` : type === 'listener' ? `(${protocol}:${port})` : '';
                return (
                  <>
                    <OverflowTitle type='tips' class='base-info'>
                      {searchValue.value ? (
                        <span
                          v-html={
                            ['lb_name', 'listener_name', 'domain'].includes(searchK.value)
                              ? `${name?.replace(
                                  new RegExp(searchV.value, 'g'),
                                  `<font color='#3A84FF'>${searchV.value}</font>`,
                                )} ${extension}`
                              : `${name} ${extension?.replace(
                                  new RegExp(searchV.value, 'g'),
                                  `<font color='#3A84FF'>${searchV.value}</font>`,
                                )}`
                          }></span>
                      ) : (
                        `${name} ${extension}`
                      )}
                      {attributes.fullPath.split('-').length === 3 && isDefault && (
                        <bk-tag class='tag ml5' theme='warning'>
                          默认
                        </bk-tag>
                      )}
                    </OverflowTitle>
                    <div class={`ext-info${currentPopBoundaryNodeKey.value === id ? ' show-dropdown' : ''}`}>
                      <div class='count'>
                        {(function () {
                          switch (type) {
                            case 'lb':
                              return listenerNum || 0;
                            case 'listener':
                              if (TRANSPORT_LAYER_LIST.includes(protocol)) return null;
                              return domain_num || 0;
                            case 'domain':
                              return url_count || 0;
                            default:
                              break;
                          }
                        })()}
                      </div>
                      <div class='more-action' onClick={(e) => showDropdownList(e, data)}>
                        <i class='hcm-icon bkhcm-icon-more-fill'></i>
                      </div>
                    </div>
                  </>
                );
              },
              nodeType: (node: any) => {
                if (node.type === 'loading') {
                  return null;
                }
                return <img src={typeIconMap[node.type]} alt='' class='prefix-icon' />;
              },
              nodeAction: (node: any) => {
                const { type, listenerNum, domain_num } = node;
                let isVisible = true;
                if ((type === 'lb' && !listenerNum) || (type === 'listener' && domain_num === 0) || type === 'domain') {
                  isVisible = false;
                }
                return (
                  <PlayShape
                    style={{
                      width: '10px',
                      color: !isVisible ? 'transparent' : '#979ba5',
                      transform: `${node.__attr__.isOpen ? 'rotate(90deg)' : 'rotate(0)'}`,
                    }}
                  />
                );
              },
            }}
          </Tree>
        </Loading>
      </div>
    );
  },
});
