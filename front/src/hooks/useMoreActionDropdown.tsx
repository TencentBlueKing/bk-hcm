import { ref } from 'vue';
import { $bkPopover } from 'bkui-vue';
import { TRANSPORT_LAYER_LIST } from '@/constants';

/**
 * more-action dropdown list render hooks
 */
export default (typeMenuMap: any) => {
  const popInstance = ref();
  const isDropdownShow = ref(false);
  const currentPopBoundaryNodeKey = ref(''); // 当前弹出层所在节点key

  // 显示popover时, 记录当前显示的节点key
  const handlePopShow = (node: any) => {
    isDropdownShow.value = true;
    currentPopBoundaryNodeKey.value = node.nodeKey;
  };

  // 隐藏popover时, 清空key
  const handlePopHide = () => {
    isDropdownShow.value = false;
    currentPopBoundaryNodeKey.value = '';
  };

  // 初始化popover, 并显示
  const showDropdownList = (e: any, node: any) => {
    popInstance.value?.close();
    popInstance.value = $bkPopover({
      isShow: isDropdownShow.value,
      // todo: 暂不支持 manual
      trigger: 'manual',
      forceClickoutside: true,
      theme: 'light',
      renderType: 'shown',
      placement: 'bottom-start',
      arrow: false,
      extCls: 'more-action-dropdown-menu',
      allowHtml: true,
      target: e,
      content: (
        <div class='dropdown-list'>
          {typeMenuMap[node.type].map((item: any, index: number) => {
            if (node.type === 'listener' && TRANSPORT_LAYER_LIST.includes(node.protocol) && index === 0) return null;
            const disabled = typeof item.isDisabled === 'function' ? item.isDisabled(node) : false;
            const tooltips = typeof item.tooltips === 'function' ? item.tooltips(node) : { disabled: true };
            const hasPermission = typeof item.preAuth === 'function' ? item.preAuth() : true;
            return (
              <div
                class={['dropdown-item', { disabled, 'hcm-no-permision-text-btn': !hasPermission }]}
                onClick={() => {
                  !disabled && item.handler(node);
                  handlePopHide();
                }}
                v-bk-tooltips={tooltips}>
                {item.label}
              </div>
            );
          })}
        </div>
      ),
      onShow: () => handlePopShow(node),
      onHide: handlePopHide,
    });
    popInstance.value?.show();
    popInstance.value?.update(e.target);
    popInstance.value?.show();
  };

  return {
    showDropdownList,
    isDropdownShow,
    currentPopBoundaryNodeKey,
  };
};
