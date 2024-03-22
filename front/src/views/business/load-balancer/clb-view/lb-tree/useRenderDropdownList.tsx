import { ref } from 'vue';
import { $bkPopover } from 'bkui-vue';

/**
 * lb-tree dropdown list render hooks
 */
export default () => {
  // type 与 dropdown menu 的映射关系
  const typeMenuMap = {
    all: [{ label: '购买负载均衡', url: 'add' }],
    lb: [
      { label: '新增监听器', url: 'add' },
      { label: '查看详情', url: 'detail' },
      { label: '编辑', url: 'edit' },
      { label: '删除', url: 'delete' },
    ],
    listener: [
      { label: '新增域名', url: 'add' },
      { label: '查看详情', url: 'detail' },
      { label: '编辑', url: 'edit' },
      { label: '删除', url: 'delete' },
    ],
    domain: [
      { label: '新增 URL 路径', url: 'add' },
      { label: '编辑', url: 'edit' },
      { label: '删除', url: 'delete' },
    ],
  };

  const popInstance = ref();
  const currentPopBoundaryNodeKey = ref(''); // 当前弹出层所在节点key

  // define handler function
  const handleDropdownItemClick = () => {
    // dropdown item click event
  };

  // 显示popover时, 记录当前显示的节点key
  const handlePopShow = (node: any) => {
    currentPopBoundaryNodeKey.value = node.id;
  };

  // 隐藏popover时, 清空key
  const handlePopHide = () => {
    currentPopBoundaryNodeKey.value = '';
  };

  // 初始化popover, 并显示
  const showDropdownList = (e: Event, node: any) => {
    popInstance.value?.close();
    popInstance.value = $bkPopover({
      trigger: 'click',
      theme: 'light',
      renderType: 'shown',
      placement: 'bottom-start',
      arrow: false,
      extCls: 'lb-tree-dropdown-popover-wrap',
      allowHtml: true,
      content: (
        <div class='dropdown-list'>
          {typeMenuMap[node.type].map((item) => (
            <div class='dropdown-item' onClick={handleDropdownItemClick}>
              {item.label}
            </div>
          ))}
        </div>
      ),
      onShow: () => handlePopShow(node),
      onHide: handlePopHide,
    });
    popInstance.value?.show();
    popInstance.value?.update(e);
    popInstance.value?.show();
  };

  return {
    showDropdownList,
    currentPopBoundaryNodeKey,
  };
};
