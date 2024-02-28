import { defineComponent, PropType } from 'vue';
import { Dialog } from 'bkui-vue';
import { useLocalTable } from '@/hooks/useLocalTable';
import type { IProp } from '@/hooks/useLocalTable';
import './index.scss';

export default defineComponent({
  name: 'BatchOperationDialog',
  props: {
    isShow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    title: {
      type: String as PropType<string>,
      required: true,
    },
    theme: {
      type: String as PropType<'primary' | 'warning' | 'success' | 'danger'>,
      default: 'primary',
    },
    confirmText: {
      type: String as PropType<string>,
      default: '确定',
    },
    tableProps: {
      type: Object as PropType<IProp>,
    },
    custom: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:isShow', 'handleConfirm'],
  setup(props, { emit, slots }) {
    const triggerShow = (isShow: boolean) => {
      emit('update:isShow', isShow);
    };
    const handleConfirm = () => {
      emit('handleConfirm');
      triggerShow(false);
    };
    // 默认渲染
    const renderDefaultSlot = () => {
      const { CommonLocalTable } = useLocalTable(props.tableProps);
      return (
        <div class='batch-operation-dialog-content'>
          <div class='tips'>{slots.tips?.()}</div>
          <CommonLocalTable>
            {{
              tab: () => slots.tab?.(),
            }}
          </CommonLocalTable>
        </div>
      );
    };
    // 自定义渲染
    const renderCustomDefaultSlot = () => {
      return (<div class='batch-operation-dialog-content'>自定义内容</div>);
    };
    return () => (
      <Dialog
        class='batch-operation-dialog'
        width={960}
        isShow={props.isShow}
        title={props.title}
        theme={props.theme}
        confirmText={props.confirmText}
        onConfirm={handleConfirm}
        onClosed={() => triggerShow(false)}>
        {{
          default: props.custom ? renderCustomDefaultSlot : renderDefaultSlot,
        }}
      </Dialog>
    );
  },
});
