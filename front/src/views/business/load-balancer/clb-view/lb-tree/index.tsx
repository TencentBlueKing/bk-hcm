import { PropType, defineComponent, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
// import components
import SimpleSearchSelect from '../../components/simple-search-select';
import { Message, Tree } from 'bkui-vue';
import Confirm from '@/components/confirm';
// import stores
import { useBusinessStore, useLoadBalancerStore, useResourceStore } from '@/store';
// import custom hooks
import useLoadTreeData from './useLoadTreeData';
import useMoreActionDropdown from '@/hooks/useMoreActionDropdown';
// import utils
import { throttle } from 'lodash';
import bus from '@/common/bus';
// import static resources
import allLBIcon from '@/assets/image/all-lb.svg';
import lbIcon from '@/assets/image/loadbalancer.svg';
import listenerIcon from '@/assets/image/listener.svg';
import domainIcon from '@/assets/image/domain.svg';
// import constants
import { TRANSPORT_LAYER_LIST } from '@/constants';
import './index.scss';

type NodeType = 'all' | 'lb' | 'listener' | 'domain';

export default defineComponent({
  name: 'LoadBalancerTree',
  props: {
    activeType: String as PropType<NodeType>,
  },
  emits: ['update:activeType'],
  setup(props, { emit }) {
    // use hooks
    const router = useRouter();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const resourceStore = useResourceStore();
    const businessStore = useBusinessStore();

    // 搜索相关
    const searchValue = ref('');
    const searchDataList = [
      { id: 'clb_name', name: '负载均衡名称' },
      { id: 'clb_vip', name: '负载均衡VIP' },
      { id: 'listener_name', name: '监听器名称' },
      { id: 'protocol', name: '协议' },
      { id: 'port', name: '端口' },
      { id: 'domain', name: '域名' },
    ];

    // lb-tree相关
    const treeData = ref([]);
    const treeRef = ref();
    const allLBNode = ref({ type: 'all', isDropdownListShow: false, id: '-1' });
    const lastSelectedNode = ref(); // 记录上一次选中的tree-node, 不包括全部负载均衡
    const loadingRef = ref();
    const expandedNodeArr = ref([]);

    // use custom hooks
    const { loadRemoteData, handleLoadDataByScroll } = useLoadTreeData(treeData);

    // 删除监听器
    const handleDeleteListener = () => {
      const { id, name } = loadBalancerStore.currentSelectedTreeNode;
      Confirm('请确定删除监听器', `将删除监听器【${name}】`, () => {
        resourceStore.deleteBatch('listeners', { ids: [id] }).then(() => {
          Message({ theme: 'success', message: '删除成功' });
          // todo: 重新请求对应lb下的listener列表
        });
      });
    };

    // 删除域名
    const handleDeleteDomain = () => {
      const { listener_id, domain } = loadBalancerStore.currentSelectedTreeNode;
      Confirm('请确定删除域名', `将删除域名【${domain}】`, async () => {
        // todo: 这里的接口需要再联调
        await businessStore.deleteRules(listener_id, { lbl_id: listener_id, domain });
        Message({
          message: '删除成功',
          theme: 'success',
        });
        // todo: 重新请求对应listener下的domain列表
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
        { label: '查看详情', handler: () => bus.$emit('changeSpecificClbActiveTab', 'detail') },
        { label: '删除', handler: () => {} }, // todo: 等批量删除CLB接口联调完后进行完善
      ],
      listener: [
        { label: '新增域名', handler: () => bus.$emit('showAddDomainSideslider') },
        { label: '查看详情', handler: () => bus.$emit('changeSpecificListenerActiveTab', 'detail') },
        { label: '编辑', handler: () => {} }, // todo: 待产品确定编辑位置
        { label: '删除', handler: handleDeleteListener },
      ],
      domain: [
        { label: '新增 URL 路径', handler: () => bus.$emit('showAddUrlSideslider') },
        { label: '编辑', handler: () => {} }, // todo: 待产品确定编辑位置
        { label: '删除', handler: handleDeleteDomain },
      ],
    };
    const { showDropdownList, currentPopBoundaryNodeKey } = useMoreActionDropdown(typeMenuMap);

    // const searchOption = computed(() => {
    //   return {
    //     value: searchValue.value,
    //     match: (searchValue: string, itemText: string, item: any) => {
    //       // todo: 需要补充搜索关键词的映射，如 key=clb_name，则需要匹配 type=clb 且 name=searchValue 的项
    //       const v = searchValue.split(':')[1];
    //       let result = false;
    //       if (item.type === 'clb') {
    //         result = new RegExp(v, 'g').test(itemText);
    //         if (result) {
    //           searchResultCount.value = searchResultCount.value + 1;
    //         }
    //       }
    //       return result;
    //     },
    //     showChildNodes: false,
    //   };
    // });

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
      if (searchValue.value) return null;
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
      if (searchValue.value) return null;
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

    // render - lb-tree 的节点
    const renderDefaultNode = (data: any, attributes: any) => {
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
      return (
        <>
          <div class='base-info'>
            {searchValue.value ? (
              <span
                v-html={data.name?.replace(
                  new RegExp(searchValue.value.split(':')[1], 'g'),
                  `<font color='#3A84FF'>${searchValue.value.split(':')[1]}</font>`,
                )}></span>
            ) : (
              data.name
            )}
            {attributes.fullPath.split('-').length === 3 && data.isDefault && (
              <bk-tag class='tag' theme='warning' radius='2px'>
                默认
              </bk-tag>
            )}
          </div>
          <div class={`ext-info${currentPopBoundaryNodeKey.value === data.id ? ' show-dropdown' : ''}`}>
            <div class='count'>
              {/* eslint-disable-next-line no-nested-ternary */}
              {data.type === 'lb' ? data.id : data.type === 'listener' ? data.domain_num : data.url_count}
            </div>
            <div class='more-action' onClick={(e) => showDropdownList(e, data)}>
              <i class='hcm-icon bkhcm-icon-more-fill'></i>
            </div>
          </div>
        </>
      );
    };

    // render - prefix icon
    const renderPrefixIcon = (node: any) => {
      if (node.type === 'loading') {
        return null;
      }
      return <img src={typeIconMap[node.type]} alt='' class='prefix-icon' />;
    };

    // define handler function - 节点点击
    const handleNodeClick = (node: any) => {
      // 更新 store 中当前选中的节点
      loadBalancerStore.setCurrentSelectedTreeNode(node);
      // 交互 - 高亮切换效果
      if (node.type !== 'all') {
        lastSelectedNode.value = node;
      } else {
        treeRef.value.setSelect(lastSelectedNode.value, false);
      }
      // 切换右侧组件
      emit('update:activeType', node.type);
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

    return () => (
      <div class='load-balancer-tree'>
        {/* 搜索 */}
        <SimpleSearchSelect v-model={searchValue.value} dataList={searchDataList} />
        {/* 全部负载均衡 */}
        <div
          class={[
            'all-lb-item',
            `${props.activeType === 'all' ? ' selected' : ''}`,
            `${currentPopBoundaryNodeKey.value === '-1' ? ' show-dropdown' : ''}`,
          ]}
          onClick={() => handleNodeClick(allLBNode.value)}>
          <div class='base-info'>
            <img src={allLBIcon} alt='' class='prefix-icon' />
            <span class='text'>全部负载均衡</span>
          </div>
          <div class='ext-info'>
            <div class='count'>{6654}</div>
            <div class='more-action' onClick={(e) => showDropdownList(e, allLBNode.value)}>
              <i class='hcm-icon bkhcm-icon-more-fill'></i>
            </div>
          </div>
        </div>
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
          line-height={36}
          onNodeClick={handleNodeClick}
          onScroll={getTreeScrollFunc()}
          async={getTreeAsyncOption()}
          onNodeExpand={handleNodeExpand}
          onNodeCollapse={handleNodeCollapse}>
          {{
            default: ({ data, attributes }: any) => renderDefaultNode(data, attributes),
            nodeType: (node: any) => renderPrefixIcon(node),
          }}
        </Tree>
      </div>
    );
  },
});
