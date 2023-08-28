import { Button, Card, Input } from 'bkui-vue';
import { PropType, defineComponent, ref } from 'vue';
import './index.scss';
import SuccessIcon from '@/assets/image/success.png';
import FailureIcon from '@/assets/image/failure.png';

type ResultType = 'success' | 'failure';

export default defineComponent({
  props: {
    type: {
      type: String as PropType<ResultType>,
      default: 'success',
      required: true,
    },
  },
  setup(props) {
    const failureMsg = ref('lkjl;kjalsdkjflaskjdflaksjdflkasmvlaksjdfopqiweuroiqwjefjaskdfj;lsadjflsjdlgwoqhiopfisdfjlaskjdf;lasjdglakhgdkashdfkshklfdajsdlfkjasldjglaskdhakshdflkjasdflkjasl;dfjlaskdfjlsjdflsdcmlsx,cna,msgdnkajgkasejlfkjsaldfj;lsjdflsajdlkfjslkvnmfn,mxnvlk');
    return () => (
      <div class={'result-page'}>
        <Card class={'result-page-card'} showHeader={false}>
          <div class={'result-page-content'}>
            {props.type === 'success' ? (
              <img
                src={SuccessIcon}
                alt='success'
                class={'result-page-icon-success'}
              />
            ) : (
              <img
                src={FailureIcon}
                alt='success'
                class={'result-page-icon-success'}
              />
            )}
            <p class={'result-page-title'}>
              {props.type === 'success' ? '任务接入成功' : '任务接入失败'}
            </p>
            <p class={'result-page-text'}>
              {props.type === 'success'
                ? '可以进行同步任务查看，或进行资源管理'
                : '云API权限不足，错误详情如下所示'}
            </p>
            <div>
              {props.type === 'success' ? (
                <>
                  <Button theme='primary' class={'result-page-success-btn'}>
                    任务详情
                  </Button>
                  <Button>资源管理</Button>
                </>
              ) : (
                <Input
                  type='textarea'
                  disabled={true}
                  class={'result-page-failure-box'}
                  v-model={failureMsg.value}
                  >
                </Input>
              )}
            </div>
          </div>
        </Card>
      </div>
    );
  },
});
