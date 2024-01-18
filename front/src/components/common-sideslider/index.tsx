import { defineComponent } from 'vue';
import { Sideslider, Button } from 'bkui-vue';
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
  },
  emits: ['update:isShow', 'handleSubmit'],
  setup(props, ctx) {
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
        title={props.title}
        onClosed={() => triggerShow(false)}>
        {{
          default: () => ctx.slots.default?.(),
          footer: () => (
            <>
              <Button theme='primary' onClick={handleSubmit}>
                提交
              </Button>
              <Button onClick={() => triggerShow(false)}>取消</Button>
            </>
          ),
        }}
      </Sideslider>
    );
  },
});
