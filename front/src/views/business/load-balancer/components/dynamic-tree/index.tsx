import { defineComponent, onMounted, ref } from "vue";
import { throttle } from "lodash";
import axios from "axios";
import './index.scss';

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
    },
    rootType: {
      type: String,
      required: true,
    },
    typeIconMap: {
      type: Object,
      required: true,
    }
  },
  emits: ["update:treeData"],
  setup(props, ctx) {
    const loadingRef = ref();
    const rootPageNum = ref(1);

    // Intersection Observer 监听器
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          // 触发 loadingRef 身上的 loadDataByScroll 自定义事件
          loadingRef.value.$emit("loadDataByScroll");
        }
      });
    });

    /**
     * 加载数据
     * @param {*} _item 需要加载数据的节点
     * @param {*} _depth 需要加载数据的节点的深度
     * @param {*} isLoadRoot 值为 true 时，加载根节点；建议为 true 时，前两个参数设置为 null 和 -1。
     */
    const loadRemoteData = async(_item: any, _depth: number, isLoadRoot?: boolean) => {
      const url = props.baseUrl + `/${isLoadRoot ? props.rootType : _item.subType}`;
      const params = { 
        _page: isLoadRoot ? rootPageNum.value : _item.pageNum, 
        _limit: 50, 
        parentId: isLoadRoot ? null : _item.id // 根节点没有 parentId，或者后端给个 null 也行，这样前端这里就不需要判断了
      };
      const [res1, res2] = await Promise.all([ axios.get(url, {params}), axios.get(url, {params: { parentId: isLoadRoot ? null : _item.id }}) ]);

      // 组装新增的节点
      const _increamentNodes = res1.data.map((item: any) => {
        // 如果是加载根节点的数据，就不用设置 item.type
        !isLoadRoot && (item.type = _item.subType);
        // 如果是加载根节点或非叶子节点的数据，需要给每个 item 添加 async = true 以及初始化 pageNum = 1
        if (_depth < 3 || isLoadRoot) {
          item.async = true;
          item.pageNum = 1;
        }
        return item;
      })
      
      if (isLoadRoot) {
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
        loadRemoteData(data._parent, attributes.fullPath.split("-").length);
      } else {
        ctx.emit('update:treeData', props.treeData.slice(0, -1));
        rootPageNum.value++;
        loadRemoteData(null, -1, true);
      }
    }

    onMounted(() => {
      // 组件挂载，加载 root node
      loadRemoteData(null, -1, true);
    })

    return () => (
      <div>
        <bk-tree data={props.treeData} label="name" children="children" level-line virtual-render line-height={36} offset-left={16}
          onScroll={throttle(() => { loadingRef.value && observer.observe(loadingRef.value.$el); }, 300)}
          async={{ callback: (_item: any, _callback: Function, _schema: any) => { loadRemoteData(_item, _schema.fullPath.split("-").length + 1) }, cache: true }}>
          {{
            default: ({ data, attributes }: any) => {
              if (data?.type === 'loading') {
                return (
                  <bk-loading ref={loadingRef} loading size="small" onLoadDataByScroll={() => { 
                    // 因为在标签上使用 data-xxx 会丢失引用，但我需要 data._parent 的引用（因为加载数据时会直接操作该对象），所以这里借用了闭包的特性。
                    handleLoadDataByScroll(data, attributes)
                  }}>
                    <div style={{height: "36px"}}></div>
                  </bk-loading>
                )
              }
              return <div class='i-tree-node-item-wrap'>
                <span class='node-name'>{data.name}</span>
                <div class='right-wrap'>5</div>
              </div>
            },
            nodeType: (node: any) => {
              return <img src={props.typeIconMap[node?.type]} alt="" style="padding-right: 8px;"/>
            }
          }}
        </bk-tree>
      </div>
    )
  }
})
