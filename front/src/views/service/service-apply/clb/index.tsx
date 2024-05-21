import { defineComponent, reactive } from 'vue';
// import components
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import SubnetPreviewDialog from '../cvm/children/SubnetPreviewDialog';
import VpcPreviewDialog from '../cvm/children/VpcPreviewDialog';
import LbSpecTypeSelectDialog from '@/views/business/load-balancer/components/LbSpecTypeDialog';
// import custom hooks
import useBindEip from './hooks/useBindEip';
import useRenderForm from './hooks/useRenderForm';
import useBottomBar from './hooks/useBottomBar';
import useHandleParams from './hooks/useHandleParams';
// import types
import { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
// import utils
import { useI18n } from 'vue-i18n';
import './index.scss';

export default defineComponent({
  name: 'ApplyLoadBalancer',
  setup() {
    // use hooks
    const { t } = useI18n();
    // define data
    const formModel = reactive<ApplyClbModel>({
      bk_biz_id: 0,
      account_id: '',
      region: '',
      load_balancer_type: 'OPEN',
      address_ip_version: 'IPV4',
      zoneType: 'single',
      zones: '',
      cloud_vpc_id: '',
      internet_charge_type: 'TRAFFIC_POSTPAID_BY_HOUR',
      sla_type: 'shared',
      require_count: 1,
      name: '',
      vendor: null,
      account_type: 'STANDARD',
    });

    // use custom hooks
    const { vpcData, isVpcPreviewDialogShow, subnetData, isSubnetPreviewDialogShow, ApplyClbForm, formRef } =
      useRenderForm(formModel);
    const { BindEipDialog } = useBindEip(formModel);
    const { ApplyClbBottomBar } = useBottomBar(formModel, formRef);
    useHandleParams(formModel, formRef);

    return () => (
      <div class='apply-clb-page'>
        {/* header */}
        <DetailHeader>
          <p class='apply-clb-header-title'>{t('购买负载均衡')}</p>
        </DetailHeader>

        {/* form */}
        <ApplyClbForm />

        {/* bottom */}
        <ApplyClbBottomBar />

        <VpcPreviewDialog
          isShow={isVpcPreviewDialogShow.value}
          data={vpcData.value}
          handleClose={() => (isVpcPreviewDialogShow.value = false)}
        />
        <SubnetPreviewDialog
          isShow={isSubnetPreviewDialogShow.value}
          data={subnetData.value}
          handleClose={() => (isSubnetPreviewDialogShow.value = false)}
        />
        <BindEipDialog />
        {/* 负载均衡规格类型选择弹框 */}
        <LbSpecTypeSelectDialog v-model={formModel.sla_type} />
      </div>
    );
  },
});
