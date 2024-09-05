import { defineComponent, reactive } from 'vue';
// import components
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import SubnetPreviewDialog from '../cvm/children/SubnetPreviewDialog';
import LbSpecTypeSelectDialog from '@/views/business/load-balancer/components/LbSpecTypeDialog';
// import custom hooks
import useBindEip from './hooks/useBindEip';
import useRenderForm from './hooks/useRenderForm';
import useBottomBar from './hooks/useBottomBar';
import useHandleParams from './hooks/useHandleParams';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
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
    const { getBizsId, whereAmI } = useWhereAmI();
    // define data
    const formModel = reactive<ApplyClbModel>({
      bk_biz_id: whereAmI.value === Senarios.business ? getBizsId() : 0,
      account_id: '',
      region: '',
      load_balancer_type: 'OPEN',
      address_ip_version: 'IPV4',
      cloud_vpc_id: '',
      zoneType: '0',
      zones: '',
      backup_zones: '',
      vip_isp: '',
      sla_type: 'shared',
      internet_charge_type: 'TRAFFIC_POSTPAID_BY_HOUR',
      require_count: 1,
      name: '',
      vendor: null,
      account_type: 'STANDARD',
      slaType: '0',
    });

    // use custom hooks
    const { subnetData, isSubnetPreviewDialogShow, ApplyClbForm, formRef, isInquiryPricesLoading } =
      useRenderForm(formModel);
    const { BindEipDialog } = useBindEip(formModel);
    const { ApplyClbBottomBar } = useBottomBar(formModel, formRef, isInquiryPricesLoading);
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
