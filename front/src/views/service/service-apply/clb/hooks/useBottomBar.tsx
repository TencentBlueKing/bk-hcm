import { Ref, computed, defineComponent, onMounted, onUnmounted, ref, shallowRef } from 'vue';
import { useRouter, useRoute } from 'vue-router';
// import components
import { Button, Loading, Popover, Table } from 'bkui-vue';
// import types
import { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { IQueryResData, LbPrice } from '@/typings';
// import utils
import { useI18n } from 'vue-i18n';
import bus from '@/common/bus';
import http from '@/http';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { applyClbSuccessHandler } from '../apply-clb.plugin';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Column } = Table;

// apply-clb, 底栏
export default (
  formModel: ApplyClbModel,
  formRef: any,
  isInquiryPrices: Ref<boolean>,
  isInquiryPricesLoading: Ref<boolean>,
) => {
  // use hooks
  const router = useRouter();
  const route = useRoute();
  const { whereAmI, isBusinessPage } = useWhereAmI();
  const { t } = useI18n();

  // 权限校验
  const { handleAuth, authVerifyData } = useVerify();
  const globalPermissionDialogStore = useGlobalPermissionDialog();
  const createClbActionName = computed(() => {
    if (whereAmI.value === Senarios.business) {
      return 'biz_clb_resource_create';
    }
    return 'clb_resource_create';
  });

  // use stores
  // define data
  const applyLoading = ref(false);
  const prices = shallowRef<LbPrice>();
  // 价格表格数据
  const priceTableData = computed(() => [
    {
      billingItem: '实例费用',
      billingMode: '包年包月',
      price: prices.value.instance_price?.unit_price_discount
        ? `${prices.value.instance_price?.unit_price_discount} 元`
        : '--',
    },
    {
      billingItem: '网络费用',
      billingMode: '包月',
      price: prices.value.bandwidth_price?.unit_price_discount
        ? `${prices.value.bandwidth_price?.unit_price_discount} 元`
        : '--',
    },
  ]);
  // 总价格
  const totalPrice = computed(() => {
    const instancePrice = prices.value?.instance_price?.unit_price_discount || 0;
    const bandwidthPrice = prices.value?.bandwidth_price?.unit_price_discount || 0;
    return (instancePrice + bandwidthPrice).toFixed(2);
  });

  const isOpen = computed(() => formModel.load_balancer_type === 'OPEN');
  const isIpv4 = computed(() => formModel.address_ip_version === 'IPV4');
  const hasZonesConfig = computed(() => (isOpen.value && isIpv4.value) || !isOpen.value);
  const hasBackupZonesConfig = computed(() => isOpen.value && isIpv4.value);
  const hasInternetChargeTypeConfig = computed(() => isOpen.value && formModel.account_type !== 'LEGACY');

  const handleParams = () => {
    // eslint-disable-next-line no-nested-ternary, prefer-const
    let zones = hasZonesConfig.value ? (formModel.zones ? [formModel.zones] : []) : undefined;

    return {
      ...formModel,
      bk_biz_id: isBusinessPage ? formModel.bk_biz_id : undefined,
      // 只有公网下可以配置
      address_ip_version: isOpen.value ? formModel.address_ip_version : undefined,
      vip_isp: isOpen.value ? formModel.vip_isp : undefined,
      // eslint-disable-next-line no-nested-ternary
      sla_type: isOpen.value ? (formModel.sla_type === 'shared' ? '' : formModel.sla_type) : undefined,
      // 只有公网下的标准账号可以配置
      internet_charge_type: hasInternetChargeTypeConfig.value ? formModel.internet_charge_type : undefined,
      internet_max_bandwidth_out: hasInternetChargeTypeConfig.value ? formModel.internet_max_bandwidth_out : undefined,
      // 只有公网下ipv4以及内网下可以配置
      zones,
      // 只有公网下ipv4可以配置
      // eslint-disable-next-line no-nested-ternary
      backup_zones: hasBackupZonesConfig.value ? (formModel.backup_zones ? [formModel.backup_zones] : []) : undefined,
      // 只有内网下可以配置
      cloud_subnet_id: !isOpen.value ? formModel.cloud_subnet_id : undefined,
      cloud_eip_id: !isOpen.value ? formModel.cloud_eip_id ?? undefined : undefined,
      // 后端无用字段
      account_type: undefined as undefined,
      zoneType: undefined as undefined,
      slaType: undefined as undefined,
    };
  };

  // define handler function
  const handleApplyClb = async () => {
    if (!authVerifyData.value?.permissionAction?.[createClbActionName.value]) {
      handleAuth(createClbActionName.value);
      globalPermissionDialogStore.setShow(true);
      return;
    }
    try {
      await formRef.value.validate();
      // 整理参数
      applyLoading.value = true;
      const url = isBusinessPage
        ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${formModel.vendor}/applications/types/create_load_balancer`
        : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/load_balancers/create`;
      const res: IQueryResData<{ id: string }> = await http.post(url, handleParams());
      const { id } = res.data || {};
      applyClbSuccessHandler(isBusinessPage, goBack, { ...formModel, id });
    } finally {
      applyLoading.value = false;
    }
  };

  // 返回上一级, 并且不保留历史记录
  const goBack = () => {
    router.replace({
      path:
        whereAmI.value === Senarios.business
          ? `/business/loadbalancer/clb-view?bizs=${route.query[GLOBAL_BIZS_KEY]}`
          : '/resource/resource?type=clb',
      query: { ...route.query },
    });
  };

  // define component
  const ApplyClbBottomBar = defineComponent({
    setup() {
      return () => (
        <div class='apply-clb-bottom-bar'>
          <div class='info-wrap'>
            <Popover theme='light' width={362} placement='top' offset={12}>
              {{
                default: () => <span class='label has-tips'>{t('配置费用')}</span>,
                content: () => (
                  <Table data={priceTableData.value}>
                    <Column field='billingItem' label={t('计费项')}></Column>
                    <Column field='billingMode' label={t('计费模式')}></Column>
                    <Column field='price' label={t('价格')} align='right'></Column>
                  </Table>
                ),
              }}
            </Popover>
            :
            <Loading loading={isInquiryPricesLoading.value} size='small' opacity={1} color='#fafbfd' class='value'>
              <span class='number'>{totalPrice.value}</span>
              {/* 本期只支持按量计费, 按照按量计费的模式进行单位显示 */}
              <span class='unit'>{t('元/小时')}</span>
            </Loading>
          </div>
          <div class='operation-btn-wrap'>
            <Button
              class={{ 'hcm-no-permision-btn': !authVerifyData.value?.permissionAction?.[createClbActionName.value] }}
              theme='primary'
              onClick={handleApplyClb}
              loading={applyLoading.value}
              disabled={!isInquiryPrices.value || isInquiryPricesLoading.value}>
              {t('立即购买')}
            </Button>
            <Button loading={applyLoading.value} onClick={goBack}>
              {t('取消')}
            </Button>
          </div>
        </div>
      );
    },
  });

  onMounted(() => {
    bus.$on('changeLbPrice', (v: LbPrice) => (prices.value = v));
  });

  onUnmounted(() => {
    bus.$off('changeLbPrice');
  });

  return {
    ApplyClbBottomBar,
  };
};
