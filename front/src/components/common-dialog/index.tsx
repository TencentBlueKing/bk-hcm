import { PropType, defineComponent } from 'vue';
import { Dialog } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
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
    dialogType: String as PropType<'show' | 'operation' | 'confirm' | 'process'>,
  },
  emits: ['update:isShow', 'handleConfirm'],
  setup(props, { emit, slots }) {
    const { t } = useI18n();
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
        title={t(props.title)}
        width={props.width}
        dialogType={props.dialogType}
        onConfirm={handleConfirm}
        onClosed={() => triggerShow(false)}>
        {{
          default: () => slots.default?.(),
          tools: () => slots.tools?.(),
          footer: slots.footer ? () => slots.footer?.() : undefined,
        }}
      </Dialog>
    );
  },
});
