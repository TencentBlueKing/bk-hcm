import { Popover } from 'bkui-vue';

/**
 * lb-tree dropdown list render hooks
 */
export default () => {
  // type 与 dropdown menu 的映射关系
  const typeMenuMap = {
    all: [{ label: '购买负载均衡', url: 'add' }],
    load_balancers: [
      { label: '新增监听器', url: 'add' },
      { label: '查看详情', url: 'detail' },
      { label: '编辑', url: 'edit' },
      { label: '删除', url: 'delete' },
    ],
    listeners: [
      { label: '新增域名', url: 'add' },
      { label: '查看详情', url: 'detail' },
      { label: '编辑', url: 'edit' },
      { label: '删除', url: 'delete' },
    ],
    domains: [
      { label: '新增 URL 路径', url: 'add' },
      { label: '编辑', url: 'edit' },
      { label: '删除', url: 'delete' },
    ],
  };

  // define handler function
  const handleMoreActionClick = (e: Event, node: any) => {
    node.isDropdownListShow = !node.isDropdownListShow;
  };
  // define handler function
  const handleDropdownItemClick = () => {
    // dropdown item click event
  };

  // 渲染 dropdown menu
  const renderDropdownActionList = (node: any) => {
    return (
      <Popover
        trigger='click'
        theme='light'
        renderType='shown'
        placement='bottom-start'
        arrow={false}
        extCls='lb-tree-dropdown-popover-wrap'
        onAfterHidden={() => (node.isDropdownListShow = false)}>
        {{
          default: () => (
            <div class='more-action' onClick={(e) => handleMoreActionClick(e, node)}>
              <i class='hcm-icon bkhcm-icon-more-fill'></i>
            </div>
          ),
          content: () => (
            <div class='dropdown-list'>
              {typeMenuMap[node.type].map((item) => (
                <div class='dropdown-item' onClick={handleDropdownItemClick}>
                  {item.label}
                </div>
              ))}
            </div>
          ),
        }}
      </Popover>
    );
  };

  return {
    renderDropdownActionList,
  };
};
