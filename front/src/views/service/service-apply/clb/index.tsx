import { computed, defineComponent, reactive } from 'vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import SubnetPreviewDialog from '../cvm/children/SubnetPreviewDialog';
import useBindEip from './hooks/useBindEip';
import useRenderForm from './hooks/useRenderForm';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { useI18n } from 'vue-i18n';
import './index.scss';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import BottomBar from './children/bottom-bar';
import http from '@/http';

export default defineComponent({
  name: 'ApplyLoadBalancer',
  setup() {
    const route = useRoute();
    const router = useRouter();
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
      load_balancer_pass_to_target: undefined,
      vip_isp: '',
      sla_type: 'shared',
      internet_charge_type: 'TRAFFIC_POSTPAID_BY_HOUR',
      require_count: 1,
      name: '',
      vendor: null,
      account_type: 'STANDARD',
      slaType: '0',
      egress: undefined,
    });

    // use custom hooks
    const { subnetData, isSubnetPreviewDialogShow, ApplyClbForm, configureList } = useRenderForm(formModel);
    const { BindEipDialog } = useBindEip(formModel);

    const fromConfig = computed<Partial<RouteLocationRaw>>(() => {
      return { query: { ...route.query } };
    });

    const goBack = () => {
      router.go(-1);
    };

    const handleApplyClb = async (params: ApplyClbModel[], url: string) => {
      const allApi = params.map((item: ApplyClbModel) => http.post(url, item));
      await Promise.any(allApi);
    };

    return () => (
      <div class='apply-clb-page'>
        {/* header */}
        <DetailHeader fromConfig={fromConfig.value}>
          <p class='apply-clb-header-title'>{t('购买负载均衡')}</p>
        </DetailHeader>

        {/* form */}
        <ApplyClbForm />

        {/* bottom */}
        <BottomBar list={configureList} onConfirm={handleApplyClb} onCancel={goBack} />

        <SubnetPreviewDialog
          isShow={isSubnetPreviewDialogShow.value}
          data={subnetData.value}
          handleClose={() => (isSubnetPreviewDialogShow.value = false)}
        />
        <BindEipDialog />
      </div>
    );
  },
});
