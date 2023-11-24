import { PropType, defineComponent, ref } from 'vue';
import { Dialog, SearchSelect, Tree } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'SchemeUserProportionShowDialog',
  props: {
    isShow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  emits: ['update:isShow'],
  setup(props, ctx) {
    const toggleShow = (isShow: boolean) => {
      ctx.emit('update:isShow', isShow);
    };
    const treeData = ref([]);

    return () => (
      <Dialog
        dialogType='show'
        class='user-proportion-detail-dialog'
        isShow={props.isShow}
        title='分布权重占比'
        onClosed={() => toggleShow(false)}>
        <div class='tips-wrap mb16'>
          <i class='hcm-icon bkhcm-icon-info-line'></i>
          <div class='tips-text'>占比权重说明说明说明</div>
        </div>
        <SearchSelect
          class='mb16'
          modelValue={[]}
          data={[]}
          placeholder='请输入'
        />
        <Tree
          data={treeData.value}
          label='name'
          children='children'
          search=''
          show-node-type-icon={false}
          prefixIcon={(params: any, renderType: any) => {
            if (params.children.length === 0) return null;
            console.log(params, renderType);
            return params.isOpen ? (
              <i class='hcm-icon bkhcm-icon-minus-circle'></i>
            ) : (
              <i class='hcm-icon bkhcm-icon-plus-circle'></i>
            );
          }}>
          {{
            nodeAppend: () => <span class='proportion-num'>10</span>,
          }}
        </Tree>
      </Dialog>
    );
  },
});
