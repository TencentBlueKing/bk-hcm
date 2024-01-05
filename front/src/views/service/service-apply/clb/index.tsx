import { defineComponent, reactive } from 'vue';
import { Container, Form } from 'bkui-vue';
import './index.scss';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import ConditionOptions from '../components/common/condition-options.vue';

export default defineComponent({
  name: 'ApplyLoadBalancer',
  setup() {
    const formModel = reactive({
      bizId: '' as string,
      cloudAccountId: '' as string, // 云账号
      vendor: null as VendorEnum, // 云厂商
      region: '' as string, // 云地域
    });
    return () => (
      <div class='apply-clb-page'>
        <DetailHeader>
          <p class='apply-clb-header-title'>购买负载均衡</p>
        </DetailHeader>
        <Form class='apply-clb-form-container' formType='vertical' model={formModel}>
          <Container margin={0}>
            <ConditionOptions
              type={ResourceTypeEnum.CLB}
              v-model:bizId={formModel.bizId}
              v-model:cloudAccountId={formModel.cloudAccountId}
              v-model:vendor={formModel.vendor}
              v-model:region={formModel.region}
            />
          </Container>
        </Form>
      </div>
    );
  },
});
