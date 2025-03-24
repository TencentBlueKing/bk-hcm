import { PropType, computed, defineComponent, onMounted, onUnmounted, ref } from 'vue';
// import components
import { Form, Tag } from 'bkui-vue';
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

/**
 * * 用于新增或更新域名
 * * 页面loadBalancerStore.currentSelectedTreeNode为监听器
 */
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
    const isSniOnHTTPS = computed(
      () =>
        loadBalancerStore.currentSelectedTreeNode.sni_switch === 1 &&
        loadBalancerStore.currentSelectedTreeNode.protocol === 'HTTPS',
    );
    // use custom hooks
    const {
      isShow,
      action,
      formItemOptions,
      handleShow,
      handleSubmit,
      formData: formModel,
    } = useAddOrUpdateDomain(
      () => {
        typeof props.getListData === 'function' && props.getListData();
      },
      props.originPage,
      isSniOnHTTPS,
    );
    // <CommonSideslider>使用的loading
    const sideIsLoading = ref(false);
    const formInstance = ref();
    // CommonSideslider编辑点击提交触发
    const handleDomainSidesliderSubmit = async () => {
      sideIsLoading.value = true;
      try {
        await handleSubmit(formInstance);
        bus.$emit('resetLbTree');
      } finally {
        sideIsLoading.value = false;
      }
    };

    const rules = {
      url: [
        {
          validator: (value: string) => /^\/[\w\-/]*$/.test(value),
          message: 'URL路径不符合规范',
          trigger: 'change',
        },
      ],
    };

    onMounted(() => {
      bus.$on('showAddDomainSideslider', handleShow);
    });

    onUnmounted(() => {
      bus.$off('showAddDomainSideslider');
    });

    return () => (
      <CommonSideslider
        class='domain-sideslider'
        isSubmitLoading={sideIsLoading.value}
        title={`${action.value === OpAction.ADD ? '新增' : '编辑'}域名`}
        width={640}
        v-model:isShow={isShow.value}
        onHandleSubmit={() => {
          handleDomainSidesliderSubmit();
        }}>
        <p class='readonly-info'>
          <span class='label'>负载均衡名称</span>:
          <span class='value'>{loadBalancerStore.currentSelectedTreeNode.lb?.name}</span>
        </p>
        <p class='readonly-info'>
          <span class='label'>监听器名称</span>:
          <span class='value'>{loadBalancerStore.currentSelectedTreeNode.name}</span>
        </p>
        <p class='readonly-info'>
          <span class='label'>协议端口</span>:
          <span class='value'>
            {`${loadBalancerStore.currentSelectedTreeNode.protocol}:${loadBalancerStore.currentSelectedTreeNode.port}`}
          </span>
        </p>
        <p class='readonly-info'>
          <span class='label'>SNI</span>:
          <span class='value'>
            {!!loadBalancerStore.currentSelectedTreeNode.sni_switch ? (
              <Tag theme='success'>已开启</Tag>
            ) : (
              <Tag>未开启</Tag>
            )}
          </span>
        </p>
        <Form formType='vertical' ref={formInstance} model={formModel} rules={rules}>
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
