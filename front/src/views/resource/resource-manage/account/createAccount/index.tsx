import { Button, Dialog, Steps } from 'bkui-vue';
import { PropType, defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  props: {
    isShow: {
      type: Boolean,
      required: true,
    },
    onSubmit: {
      type: Function as PropType<() => void>,
      required: true,
    },
    onCancel: {
      type: Function as PropType<() => void>,
      required: true,
    },
  },
  setup(props) {
    return () => (
      <Dialog
        fullscreen
        isShow={props.isShow}
        class={'create-account-dialog-container'}
      >
        {{
          tools: () => (
            <div class={'create-account-dialog-tools'}>
              云账号接入
            </div>
          ),
          default: () => (
            <div class={'create-account-dialog-content'}>
              <Steps
                class={'create-account-dialog-steps'}
                steps={[
                  {
                    title: '录入账号',
                    component: () => 123,
                  },
                  {
                    title: '资源同步',
                    component: () => 456,
                  }
                ]}
              />
            </div>
          ),
          footer: () => (
            <div class={'create-account-dialog-footer'}>
              <Button
                theme={'primary'}
                class={'mr8'}
                onClick={props.onSubmit}
              >
                下一步
              </Button>
              <Button onClick={props.onCancel}>
                取消
              </Button>
            </div>
          ),
        }}
      </Dialog>
    );
  },
});
