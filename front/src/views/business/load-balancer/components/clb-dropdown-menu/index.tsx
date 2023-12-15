import { defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  name: 'LoadBalancerDropdownMenu',
  props: {
    uuid: {
      type: String,
      required: true,
    },
    type: {
      type: String,
      required: true,
    },
  },
  setup(props) {
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

    return () => (
      <bk-dropdown key={props.uuid} class='more-action' trigger='click' placement='right-start' onClick={(e: MouseEvent) => {
        e.stopPropagation();
      }}>
        {{
          default: () => <i class='hcm-icon bkhcm-icon-more-fill'></i>,
          content: () => (
            <bk-dropdown-menu>
              {typeMenuMap[props.type].map((item: any) => {
                return <bk-dropdown-item key={item.label}>{item.label}</bk-dropdown-item>;
              })}
            </bk-dropdown-menu>
          ),
        }}
      </bk-dropdown>
    );
  },
});
