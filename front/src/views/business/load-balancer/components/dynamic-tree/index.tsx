import { defineComponent, onMounted, ref, inject, computed } from 'vue';
import { throttle } from 'lodash';
import axios from 'axios';
import './index.scss';
import clbIcon from '@/assets/image/clb.png';
import listenerIcon from '@/assets/image/listener.png';
import domainIcon from '@/assets/image/domain.png';
import LoadBalancerDropdownMenu from '../clb-dropdown-menu';

/**
 * 基于 bkui-vue Tree 的动态树，支持滚动加载数据。
 *
 * 注意点：
 * - 对数据格式有要求，详细见 src/api/db.json 文件
 * - 当前只支持三层树形结构的数据，如果要加层数，改 loadData 方法中的 _depth < num 即可。
 */
export default defineComponent({
  name: 'DynamicTree',
  props: {
    searchValue: {
      type: String,
      required: true,
    },
  },
  emits: ['handleTypeChange'],
  setup(props, ctx) {
    const treeData: any = inject('treeData');
    const baseUrl = 'http://localhost:3000';
    const loadingRef = ref();
    const rootPageNum = ref(1);
    const treeRef = inject('treeRef');
    const searchResultCount: any = inject('searchResultCount');

    const searchOption = computed(() => {
      return {
        value: props.searchValue,
        match: (searchValue: string, itemText: string) => {
          const result = new RegExp(searchValue, 'g').test(itemText);
          if (result) {
            searchResultCount.value = searchResultCount.value + 1;
          }
          return result;
        },
        showChildNodes: false,
      };
    });

    // Intersection Observer 监听器
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          // 触发 loadingRef 身上的 loadDataByScroll 自定义事件
          loadingRef.value.$emit('loadDataByScroll');
        }
      });
    });

    // _depth 与 type 的映射关系
    const depthTypeMap = ['clb', 'listener', 'domain'];
    // type 与 icon 的映射关系
    const typeIconMap = {
      clb: clbIcon,
      listener: listenerIcon,
      domain: domainIcon,
    };

    /**
     * 加载数据
     * @param {*} _item 需要加载数据的节点，值为 null 表示加载根节点的数据
     * @param {*} _depth 需要加载数据的节点的深度，取值为：0, 1, 2
     */
    const loadRemoteData = async (_item: any, _depth: number) => {
      const url = `${baseUrl}/${!_item ? depthTypeMap[_depth] : depthTypeMap[_depth + 1]}`;
      const params = {
        _page: !_item ? rootPageNum.value : _item.pageNum,
        _limit: 50,
        parentId: !_item ? null : _item.id, // 根节点没有 parentId
      };
      const [res1, res2] = await Promise.all([
        axios.get(url, { params }),
        axios.get(url, { params: { parentId: !_item ? null : _item.id } }),
      ]);

      // 组装新增的节点
      const _increamentNodes = res1.data.map((item: any) => {
        // 如果是加载根节点的数据，则 type 设置为当前 type；如果是加载子节点的数据，则 type 设置为下一级 type
        !_item ? (item.type = depthTypeMap[_depth]) : (item.type = depthTypeMap[_depth + 1]);
        // 如果是加载根节点或非叶子节点的数据，需要给每个 item 添加 async = true 用于异步加载，以及初始化 pageNum = 1
        if (_depth < 1 || !_item) {
          item.async = true;
          item.pageNum = 1;
        }
        return item;
      });

      if (!_item) {
        const _treeData = [...treeData.value, ..._increamentNodes];
        if (_treeData.length < res2.data.length) {
          treeData.value = [..._treeData, { type: 'loading' }];
        } else {
          treeData.value = _treeData;
        }
      } else {
        _item.children = [..._item.children, ..._increamentNodes];
        if (_item.children.length < res2.data.length) {
          _item.children.push({ type: 'loading', _parent: _item });
        }
      }
    };

    /**
     * 滚动加载数据
     * @param {*} data 当前可视区内的 loading 节点（Tree组件中的）
     * @param {*} attributes 当前可视区内的 loading 节点（Tree组件中的）相关的属性
     */
    const handleLoadDataByScroll = (data: any, attributes: any) => {
      // 有 _parent，加载非根节点的下一页数据；无 _parent，加载根节点的下一页数据
      if (data._parent) {
        // 1.移除loading节点
        data._parent.children.pop();
        // 2.更新分页参数
        data._parent.pageNum = data._parent.pageNum + 1;
        // 3.请求下一页数据
        loadRemoteData(data._parent, attributes.fullPath.split('-').length - 2);
      } else {
        treeData.value = treeData.value.slice(0, -1);
        rootPageNum.value = rootPageNum.value + 1;
        loadRemoteData(null, 0);
      }
    };

    const getTreeScrollFunc = () => {
      if (props.searchValue) return null;
      return throttle(() => {
        loadingRef.value && observer.observe(loadingRef.value.$el);
      }, 200);
    };

    const getTreeAsyncOption = () => {
      if (props.searchValue) return null;
      return {
        callback: (_item: any, _callback: Function, _schema: any) => {
          // 异步加载当前点击节点的 children node
          loadRemoteData(_item, _schema.fullPath.split('-').length - 1);
        },
        cache: true,
      };
    };

    const renderDefaultNode = (data: any, attributes: any) => {
      if (data.type === 'loading') {
        return (
          <bk-loading
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
          <div class='left-wrap'>
            {props.searchValue ? (
              <span
                v-html={data.name?.replace(
                  new RegExp(props.searchValue, 'g'),
                  `<font color='#3A84FF'>${props.searchValue}</font>`,
                )}></span>
            ) : (
              data.name
            )}
            {attributes.fullPath.split('-').length === 3 && (
              <bk-tag class='tag' theme='warning' radius='2px'>
                默认
              </bk-tag>
            )}
          </div>
          <div class='right-wrap'>
            <div class='count'>{data.id}</div>
            <LoadBalancerDropdownMenu class='more-action' type={data.type} />
          </div>
        </>
      );
    };

    const handleNodeClick = (node: any) => {
      ctx.emit('handleTypeChange', node.type);
    };

    onMounted(() => {
      // 组件挂载，加载 root node
      loadRemoteData(null, 0);
    });

    return () => (
      <div class='dynamic-tree-wrap'>
        <bk-tree
          ref={treeRef}
          data={treeData.value}
          label='name'
          children='children'
          level-line
          virtual-render
          line-height={36}
          node-content-action={['selected', 'click']}
          search={searchOption.value}
          onNodeClick={handleNodeClick}
          onScroll={getTreeScrollFunc()}
          async={getTreeAsyncOption()}>
          {{
            default: ({ data, attributes }: any) => renderDefaultNode(data, attributes),
            nodeType: (node: any) => {
              return <img src={typeIconMap[node.type]} alt='' style='padding-right: 8px;' />;
            },
          }}
        </bk-tree>
      </div>
    );
  },
});
