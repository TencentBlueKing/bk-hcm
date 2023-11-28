import { PropType, defineComponent, ref, watch } from 'vue';
import { Dialog, Input, Tree } from 'bkui-vue';
import './index.scss';
import { IAreaInfo } from '@/typings/scheme';

export default defineComponent({
  name: 'SchemeUserProportionShowDialog',
  props: {
    isShow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    treeData: {
      type: Array<IAreaInfo> as PropType<Array<IAreaInfo>>,
    }
  },
  emits: ['update:isShow'],
  setup(props, ctx) {
    const toggleShow = (isShow: boolean) => {
      ctx.emit('update:isShow', isShow);
    };

    const searchVal = ref('');
    const treeData = ref<Array<IAreaInfo>>([]);

    watch(() => props.treeData, (val) => {
      treeData.value = val;
    })

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
        <Input
          v-model={searchVal.value}
          class='mb16'
          type="search"
          placeholder='请输入'
        />
        <Tree
          data={props.treeData}
          label='name'
          children='children'
          search={searchVal.value}
          show-node-type-icon={false}
          prefixIcon={(params: any, renderType: any) => {
            if (params.children?.length === 0) return null;
            return params.isOpen ? (
              <i class='hcm-icon bkhcm-icon-minus-circle'></i>
            ) : (
              <i class='hcm-icon bkhcm-icon-plus-circle'></i>
            );
          }}>
          {{
            nodeAppend: (node: any) => {
              const {hasChild, isOpen} = node;
              if (hasChild) {
                if (isOpen) {
                  return null;
                } else {
                  // 计算子节点的权重之和, 展示到后面
                  <span class='proportion-num'>{10}</span>
                }
              } else {
                return <span class='proportion-num'>{node.value}</span>
              }
           },
          }}
        </Tree>
      </Dialog>
    );
  },
});
