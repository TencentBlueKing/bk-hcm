import { defineComponent, onMounted, ref, inject } from "vue";
import { throttle } from "lodash";
import axios from "axios";
import './index.scss';
import clbIcon from "@/assets/image/clb.png";
import listenerIcon from "@/assets/image/listener.png";
import domainIcon from "@/assets/image/domain.png";

/**
 * 基于 bkui-vue Tree 的动态树，支持滚动加载数据。
 * 
 * 注意点：
 * - 对数据格式有要求，详细见 src/api/db.json 文件
 * - 当前只支持三层树形结构的数据，如果要加层数，改 loadData 方法中的 _depth < num 即可。
 */
export default defineComponent({
  name: "DynamicTree",
  props: {
    baseUrl: {
      type: String,
      required: true,
    },
    treeData: {
      type: Object,
      required: true,
    }
  },
  emits: ["update:treeData"],
  setup(props, ctx) {
    const loadingRef = ref();
    const rootPageNum = ref(1);
    const treeRef = inject('treeRef');
    const currentExpandItems: any = inject('currentExpandItems');

    // Intersection Observer 监听器
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          // 触发 loadingRef 身上的 loadDataByScroll 自定义事件
          loadingRef.value.$emit("loadDataByScroll");
        }
      });
    });

    // _depth 与 type 的映射关系
    const depthTypeMap = ['clb', 'listener', 'domain'];
    // type 与 icon 的映射关系
    const typeIconMap = {
      clb: clbIcon,
      listener: listenerIcon,
      domain: domainIcon
    };

    /**
     * 加载数据
     * @param {*} _item 需要加载数据的节点，值为 null 表示加载根节点的数据
     * @param {*} _depth 需要加载数据的节点的深度，取值为：0, 1, 2
     */
    const loadRemoteData = async(_item: any, _depth: number) => {
      const url = props.baseUrl + `/${!_item ? depthTypeMap[_depth] : depthTypeMap[_depth+1]}`;
      const params = { 
        _page: !_item ? rootPageNum.value : _item.pageNum,
        _limit: 50, 
        parentId: !_item ? null : _item.id // 根节点没有 parentId
      };
      const [res1, res2] = await Promise.all([ axios.get(url, {params}), axios.get(url, {params: { parentId: !_item ? null : _item.id }}) ]);

      // 组装新增的节点
      const _increamentNodes = res1.data.map((item: any) => {
        // 如果是加载根节点的数据，则 type 设置为当前 type；如果是加载子节点的数据，则 type 设置为下一级 type
        !_item ? (item.type = depthTypeMap[_depth]) : (item.type = depthTypeMap[_depth+1]);
        // 如果是加载根节点或非叶子节点的数据，需要给每个 item 添加 async = true 用于异步加载，以及初始化 pageNum = 1
        if (_depth < 1 || !_item) {
          item.async = true;
          item.pageNum = 1;
        }
        return item;
      })
      
      if (!_item) {
        const _treeData = [...props.treeData, ..._increamentNodes];
        if (_treeData.length < res2.data.length) {
          ctx.emit('update:treeData',  [..._treeData, {type: "loading"}]);
        } else {
          ctx.emit('update:treeData', _treeData);
        }
      } else {
        _item.children = [..._item.children, ..._increamentNodes];
        if (_item.children.length < res2.data.length) {
          _item.children.push({type: "loading", _parent: _item});
        }
      }
    }

    /**
     * 滚动加载数据
     * @param {*} data 当前可视区内的 loading 节点（Tree组件中的）
     * @param {*} attributes 当前可视区内的 loading 节点（Tree组件中的）相关的属性
     */
    const handleLoadDataByScroll = (data: any, attributes: any) => {
      // 有 _parent，加载非根节点的下一页数据；无 _parent，加载根节点的下一页数据
      if (data._parent) { 
        //1.移除loading节点
        data._parent.children.pop();
        //2.更新分页参数
        data._parent.pageNum++;
        //3.请求下一页数据
        loadRemoteData(data._parent, attributes.fullPath.split("-").length-2);
      } else {
        ctx.emit('update:treeData', props.treeData.slice(0, -1));
        rootPageNum.value++;
        loadRemoteData(null, 0);
      }
    }
    
    /**
     * 节点展开时触发的事件
     * @param _item 触发事件的节点
     */
    const handleNodeExpand = (_item: any) => {
      currentExpandItems.value.push(_item);
    }

    /**
     * 节点收起时触发的事件
     * @param _item 触发事件的节点
     */
    const handleNodeCollapse = (_item: any) => {
      currentExpandItems.value = currentExpandItems.value.filter((item: any) => item !== _item);
    }

    onMounted(() => {
      // 组件挂载，加载 root node
      loadRemoteData(null, 0);
    })

    return () => (
      <div class='dynamic-tree-wrap'>
        <bk-tree ref={treeRef} data={props.treeData} label="name" children="children" level-line virtual-render line-height={36}
          node-content-action={['selected', 'click']}
          onScroll={throttle(() => { loadingRef.value && observer.observe(loadingRef.value.$el); }, 300)}
          onNodeExpand={handleNodeExpand} onNodeCollapse={handleNodeCollapse}
          async={{ 
            callback: (_item: any, _callback: Function, _schema: any) => { 
              // 异步加载当前点击节点的 children node
              loadRemoteData(_item, _schema.fullPath.split("-").length-1); 
            }, 
            cache: true 
          }}>
          {{
            default: ({ data, attributes }: any) => {
              if (data.type === 'loading') {
                return (
                  <bk-loading ref={loadingRef} loading size="small" onLoadDataByScroll={() => { 
                    // 因为在标签上使用 data-xxx 会丢失引用，但我需要 data._parent 的引用（因为加载数据时会直接操作该对象），所以这里借用了闭包的特性。
                    handleLoadDataByScroll(data, attributes)
                  }}>
                    <div style={{ height: "36px" }}></div>
                  </bk-loading>
                )
              }
              return <div class='i-tree-node-item-wrap'>
                <div class='left-wrap'>
                  {data.name}
                  {
                    attributes.fullPath.split("-").length === 3 && <bk-tag class='tag' theme="warning" radius="2px">默认</bk-tag>
                  }
                </div>
                <div class='right-wrap'>
                  <div class='count'>{data.id}</div>
                  <div class='more-action'>
                    <i class='hcm-icon bkhcm-icon-more-fill'></i>
                  </div>
                </div>
              </div>
            },
            nodeType: (node: any) => {
              return <img src={typeIconMap[node.type]} alt="" style="padding-right: 8px;"/>
            }
          }}
        </bk-tree>
      </div>
    )
  }
})
