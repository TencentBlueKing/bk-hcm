import { Ref, ref } from 'vue';
// import types
import { QueryRuleOPEnum } from '@/typings';
// import utils
import { localStorageActions } from '@/common/util';
import http from '@/http';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

/**
 * 加载 lb-tree 数据
 */
export default (treeData: Ref) => {
  // _depth 与 type 的映射关系
  const depthTypeMap = ['load_balancers', 'listeners', 'domains'];

  // define data
  const rootStart = ref(0);

  /**
   * 加载数据
   * @param {*} _item 需要加载数据的节点，值为 null 表示加载根节点的数据
   * @param {*} _depth 需要加载数据的节点的深度，取值为：0, 1, 2
   */
  const loadRemoteData = async (_item: any, _depth: number) => {
    const url = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${localStorageActions.get('bizs')}/${
      !_item ? depthTypeMap[_depth] : `${depthTypeMap[_depth]}/${_item.id}/${depthTypeMap[_depth + 1]}`
    }/list`;
    const startIdx = !_item ? rootStart.value : _item.start;
    const [detailsRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        http.post(url, {
          filter: {
            op: QueryRuleOPEnum.AND,
            rules: [],
          },
          page: {
            count: isCount,
            start: isCount ? 0 : startIdx,
            limit: isCount ? 0 : 50,
          },
        }),
      ),
    );

    // 组装新增的节点
    const _incrementNodes = detailsRes.data.details.map((item: any) => {
      // 如果是加载根节点的数据，则 type 设置为当前 type；如果是加载子节点的数据，则 type 设置为下一级 type
      !_item ? (item.type = depthTypeMap[_depth]) : (item.type = depthTypeMap[_depth + 1]);
      // 如果是加载根节点或非叶子节点的数据，需要给每个 item 添加 async = true 用于异步加载，以及初始化 start = 0
      if (_depth < 1 || !_item) {
        item.async = true;
        item.start = 0;
      }
      // dropdown 是否显示的标识
      item.isDropdownListShow = false;
      return item;
    });

    if (!_item) {
      const _treeData = [...treeData.value, ..._incrementNodes];
      if (_treeData.length < countRes.data.count) {
        treeData.value = [..._treeData, { type: 'loading' }];
      } else {
        treeData.value = _treeData;
      }
    } else {
      _item.children = [..._item.children, ..._incrementNodes];
      if (_item.children.length < countRes.data.count) {
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
      data._parent.start = data._parent.start + 50;
      // 3.请求下一页数据
      loadRemoteData(data._parent, attributes.fullPath.split('-').length - 2);
    } else {
      treeData.value = treeData.value.slice(0, -1);
      rootStart.value = rootStart.value + 50;
      loadRemoteData(null, 0);
    }
  };

  return {
    loadRemoteData,
    handleLoadDataByScroll,
  };
};
