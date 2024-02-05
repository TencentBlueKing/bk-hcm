import { PropType, computed, defineComponent, ref } from 'vue';
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
    },
  },
  emits: ['update:isShow'],
  setup(props, ctx) {
    const toggleShow = (isShow: boolean) => {
      ctx.emit('update:isShow', isShow);
      searchVal.value = '';
    };

    const searchVal = ref('');
    const treeData = computed(() =>
      props.treeData.map((item) => {
        item.value = +item.children.reduce((prev, curr) => prev + curr.value, 0).toFixed(2);
        item.children = item.children.map((child) => {
          child.value = +child.value.toFixed(2);
          return child;
        });
        return item;
      }),
    );
    const searchOption = computed(() => {
      return {
        value: searchVal.value,
        showChildNodes: false,
      };
    });

    const getPrefixIcon = (params: any) => {
      const {
        __attr__: { isRoot, isOpen },
      } = params;
      // 非根节点
      if (!isRoot) return null;
      // 根节点下钻
      if (isOpen) return <i class='hcm-icon bkhcm-icon-minus-circle'></i>;
      // 根节点未下钻
      return <i class='hcm-icon bkhcm-icon-plus-circle'></i>;
    };

    const getNodeAppend = (node: any) => {
      const {
        __attr__: { isRoot, isOpen, hasChild },
      } = node;
      // 非根节点 或 根节点无子节点
      if (!isRoot || !hasChild) return <span class='proportion-num'>{node.value}</span>;
      // 根节点下钻
      if (isOpen) return null;
      // 根节点未下钻
      return <span class='proportion-num'>{node.value}</span>;
    };

    return () => (
      <Dialog
        dialogType='show'
        class='user-proportion-detail-dialog'
        isShow={props.isShow}
        title='分布权重占比'
        onClosed={() => toggleShow(false)}>
        <div class='tips-wrap mb16'>
          <i class='hcm-icon bkhcm-icon-info-line'></i>
          <div class='tips-text'>
            当前采用各地区玩家的活跃程度（如登录行为）作为用户分布占比权重的衡量依据，暂不可更改。用户在各地区的分布占比会影响到最终的推荐结果。
          </div>
        </div>
        <Input v-model={searchVal.value} class='mb16' type='search' placeholder='请输入' />
        <Tree
          data={treeData.value}
          label='name'
          children='children'
          search={searchOption.value}
          show-node-type-icon={false}
          selectable={false}
          indent={30}
          line-height={36}
          prefixIcon={getPrefixIcon}>
          {{
            nodeAppend: getNodeAppend,
          }}
        </Tree>
      </Dialog>
    );
  },
});
