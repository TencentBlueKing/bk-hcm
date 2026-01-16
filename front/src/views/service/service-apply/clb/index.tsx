import { ref, computed, defineComponent, reactive } from 'vue';
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
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { MENU_BUSINESS_LOAD_BALANCER, MENU_RESOURCE_RESOURCE_MANAGEMENT } from '@/constants/menu-symbol';
import { applyClbSuccessHandler } from './apply-clb.plugin';

export default defineComponent({
  name: 'ApplyLoadBalancer',
  setup() {
    const route = useRoute();
    const router = useRouter();
    // use hooks
    const { t } = useI18n();
    const { getBizsId, isBusinessPage, whereAmI } = useWhereAmI();
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

    const applyLoading = ref(false);

    // use custom hooks
    const { subnetData, isSubnetPreviewDialogShow, ApplyClbForm, configureList } = useRenderForm(formModel);
    const { BindEipDialog } = useBindEip(formModel);

    const fromConfig = computed<Partial<RouteLocationRaw>>(() => {
      return { query: { ...route.query } };
    });

    const goBack = () => {
      const to = isBusinessPage
        ? { name: MENU_BUSINESS_LOAD_BALANCER, query: { [GLOBAL_BIZS_KEY]: route.query[GLOBAL_BIZS_KEY] } }
        : { name: MENU_RESOURCE_RESOURCE_MANAGEMENT, query: { type: 'clb' } };

      router.replace(to);
    };

    const handleApplyClb = async (params: ApplyClbModel[]) => {
      try {
        applyLoading.value = true;
        const { vendor } = formModel;
        const url = isBusinessPage
          ? `/api/v1/cloud/vendors/${vendor}/applications/types/create_load_balancer`
          : `/api/v1/cloud/load_balancers/create`;

        const allApi = params.map((item: ApplyClbModel) => http.post(url, item));
        await Promise.any(allApi);
        applyClbSuccessHandler(isBusinessPage, goBack, { bizId: formModel.bk_biz_id });
      } finally {
        applyLoading.value = false;
      }
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
        <BottomBar list={configureList} loading={applyLoading.value} onConfirm={handleApplyClb} onCancel={goBack} />

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
