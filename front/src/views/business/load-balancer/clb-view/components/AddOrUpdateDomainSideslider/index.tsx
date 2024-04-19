import { PropType, defineComponent, onMounted, onUnmounted, ref } from 'vue';
// import components
import { Form } from 'bkui-vue';
import CommonSideslider from '@/components/common-sideslider';
// import stores
import { useLoadBalancerStore } from '@/store';
// import hooks
import useAddOrUpdateDomain, { OpAction } from './useAddOrUpdateDomain';
// import utils
import bus from '@/common/bus';
import { IOriginPage } from '@/typings';
import './index.scss';

const { FormItem } = Form;

export default defineComponent({
  name: 'AddOrUpdateDomainSideslider',
  props: {
    // listener: 具体的监听器下, domain: 具体的域名下
    originPage: String as PropType<IOriginPage>,
    getListData: Function as PropType<(...args: any) => any>,
  },
  setup(props) {
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    // use custom hooks
    const {
      isShow,
      action,
      formItemOptions,
      handleShow,
      handleSubmit,
      formData: formModel,
    } = useAddOrUpdateDomain(() => {
      typeof props.getListData === 'function' && props.getListData();
    }, props.originPage);

    const formInstance = ref();

    onMounted(() => {
      bus.$on('showAddDomainSideslider', handleShow);
    });

    onUnmounted(() => {
      bus.$off('showAddDomainSideslider');
    });

    return () => (
      <CommonSideslider
        class='domain-sideslider'
        title={`${action.value === OpAction.ADD ? '新增' : '编辑'}域名`}
        width={640}
        v-model:isShow={isShow.value}
        onHandleSubmit={() => {
          handleSubmit(formInstance);
        }}>
        <p class='readonly-info'>
          <span class='label'>负载均衡名称</span>:
          <span class='value'>{loadBalancerStore.currentSelectedTreeNode.lb.name}</span>
        </p>
        <p class='readonly-info'>
          <span class='label'>监听器名称</span>:
          <span class='value'>
            {props.originPage === 'listener'
              ? loadBalancerStore.currentSelectedTreeNode.name
              : loadBalancerStore.currentSelectedTreeNode.listener.lbl_name}
          </span>
        </p>
        <p class='readonly-info'>
          <span class='label'>协议端口</span>:
          <span class='value'>
            {props.originPage === 'listener'
              ? `${loadBalancerStore.currentSelectedTreeNode.protocol}:${loadBalancerStore.currentSelectedTreeNode.port}`
              : `${loadBalancerStore.currentSelectedTreeNode.listener.protocol}:${loadBalancerStore.currentSelectedTreeNode.listener.port}`}
          </span>
        </p>
        <Form formType='vertical' ref={formInstance} model={formModel}>
          {formItemOptions.value
            .filter(({ hidden }) => !hidden)
            .map(({ label, required, property, content }) => {
              return (
                <FormItem label={label} required={required} key={property}>
                  {content()}
                </FormItem>
              );
            })}
        </Form>
      </CommonSideslider>
    );
  },
});
