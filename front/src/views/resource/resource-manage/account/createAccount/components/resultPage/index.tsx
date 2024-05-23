import { Button, Card, Input } from 'bkui-vue';
import { PropType, defineComponent } from 'vue';
import './index.scss';
import SuccessIcon from '@/assets/image/success.png';
import FailureIcon from '@/assets/image/failure.png';
import { useRoute, useRouter } from 'vue-router';

type ResultType = 'success' | 'failure';

export default defineComponent({
  props: {
    type: {
      type: String as PropType<ResultType>,
      default: 'success',
      required: true,
    },
    errMsg: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const router = useRouter();
    const route = useRoute();
    return () => (
      <div class={'result-page'}>
        <Card class={'result-page-card'} showHeader={false}>
          <div class={'result-page-content'}>
            {props.type === 'success' ? (
              <img src={SuccessIcon} alt='success' class={'result-page-icon-success'} />
            ) : (
              <img src={FailureIcon} alt='failure' class={'result-page-icon-success'} />
            )}
            <p class={'result-page-title'}>{props.type === 'success' ? '账号信息填写成功' : '账号接入失败'}</p>
            <p class={'result-page-text'}>
              {props.type === 'success' ? '账号审批通过后，可以进行资源管理' : '错误详情如下所示'}
            </p>
            <div>
              {props.type === 'success' ? (
                <>
                  <Button
                    theme='primary'
                    class={'result-page-success-btn'}
                    onClick={() => {
                      router.replace({
                        path: '/service/my-apply',
                        query: route.query,
                      });
                    }}>
                    查看单据
                  </Button>
                  {/* <Button>资源管理</Button> */}
                </>
              ) : (
                <Input
                  type='textarea'
                  readonly
                  class={'result-page-failure-box'}
                  v-model={props.errMsg}
                  resize={false}
                  placeholder=' '></Input>
              )}
            </div>
          </div>
        </Card>
      </div>
    );
  },
});
