import { defineComponent } from 'vue';
import { Dialog } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'CommonDialog',
  props: {
    isShow: {
      type: Boolean,
      default: false,
    },
    title: String,
    width: [String, Number],
  },
  emits: ['update:isShow', 'handleConfirm'],
  setup(props, { emit, slots }) {
    const triggerShow = (isShow: boolean) => {
      emit('update:isShow', isShow);
    };
    const handleConfirm = () => {
      emit('update:isShow', false);
      emit('handleConfirm');
    };
    return () => (
      <Dialog
        class='common-dialog'
        isShow={props.isShow}
        title={props.title}
        width={props.width}
        onConfirm={handleConfirm}
        onClosed={() => triggerShow(false)}>
        {{
          default: () => slots.default?.(),
          tools: () => slots.tools?.(),
        }}
      </Dialog>
    );
  },
});
