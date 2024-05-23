import { defineComponent } from 'vue';
import { Sideslider, Button } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import './index.scss';

export default defineComponent({
  name: 'CommonSideslider',
  props: {
    isShow: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      required: true,
    },
    width: {
      type: [Number, String],
      default: 400,
    },
    isSubmitDisabled: {
      type: Boolean,
      default: false,
    },
    isSubmitLoading: {
      type: Boolean,
      default: false,
    },
    handleClose: Function,
  },
  emits: ['update:isShow', 'handleSubmit'],
  setup(props, ctx) {
    // use hooks
    const { t } = useI18n();

    const triggerShow = (isShow: boolean) => {
      ctx.emit('update:isShow', isShow);
    };

    const handleSubmit = () => {
      ctx.emit('handleSubmit');
    };

    return () => (
      <Sideslider
        class='common-sideslider'
        width={props.width}
        isShow={props.isShow}
        title={t(props.title)}
        onClosed={() => {
          triggerShow(false);
          props.handleClose?.();
        }}>
        {{
          default: () => <div class='common-sideslider-content'>{ctx.slots.default?.()}</div>,
          footer: () => (
            <>
              <Button
                theme='primary'
                onClick={handleSubmit}
                disabled={props.isSubmitDisabled}
                loading={props.isSubmitLoading}>
                {t('提交')}
              </Button>
              <Button onClick={() => triggerShow(false)}>{t('取消')}</Button>
            </>
          ),
        }}
      </Sideslider>
    );
  },
});
