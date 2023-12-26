import { defineComponent, inject } from 'vue';
import { $bkPopover } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerDropdownMenu',
  props: {
    type: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const popInstance: any = inject('popInstance');
    const typeMenuMap = {
      all: [{ label: '购买负载均衡', url: 'add' }],
      clb: [
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

    const handleDropdownItemClick = () => {
      // dropdown item click event
    };

    const initPopInstance = () => {
      popInstance.value?.close();
      popInstance.value = $bkPopover({
        content: (
          <div class='dropdown-list'>
            {typeMenuMap[props.type].map(item => (
              <div class='dropdown-item' onClick={handleDropdownItemClick}>
                {item.label}
              </div>
            ))}
          </div>
        ),
        trigger: 'click',
        placement: 'bottom-start',
        theme: 'light',
        extCls: 'dropdown-popover-wrap',
      });
      popInstance.value.show();
    };

    const handleMoreActionClick = (e: Event) => {
      e.stopPropagation();
      initPopInstance();
      popInstance.value?.update(e);
      popInstance.value?.show();
    };

    return () => (
      <div onClick={handleMoreActionClick}>
        <i class='hcm-icon bkhcm-icon-more-fill'></i>
      </div>
    );
  },
});
