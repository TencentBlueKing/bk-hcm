import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useBusinessStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { AUTH_BIZ_CREATE_CLB, AUTH_CREATE_CLB } from '@/constants/auth-symbols';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import type { IQueryResData, LbPrice } from '@/typings';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { debounce } from 'lodash';
import http from '@/http';
import { applyClbSuccessHandler } from '../apply-clb.plugin';
import { BGP_VIP_ISP_TYPES } from '@/constants';

import { Form } from 'bkui-vue';

export default defineComponent({
  props: {
    formModel: Object as PropType<ApplyClbModel>,
    formRef: Object as PropType<InstanceType<typeof Form>>,
  },
  setup(props) {
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();
    const { whereAmI, isBusinessPage, getBizsId } = useWhereAmI();
    const businessStore = useBusinessStore();

    // 权限校验
    const computedBizId = computed(() => (whereAmI.value === Senarios.business ? getBizsId() : undefined));
    const createClbAuthSymbol = computed(() => {
      return whereAmI.value === Senarios.business ? AUTH_BIZ_CREATE_CLB : AUTH_CREATE_CLB;
    });

    // 询价
    const prices = ref<LbPrice>();
    const priceTableData = computed(() => {
      return [
        {
          billingItem: t('实例费用'),
          billingMode: t('包年包月'),
          price: prices.value?.instance_price?.unit_price_discount
            ? `${prices.value.instance_price.unit_price_discount} ${t('元')}`
            : '--',
        },
        {
          billingItem: t('网络费用'),
          billingMode: t('包月'),
          price: prices.value?.bandwidth_price?.unit_price_discount
            ? `${prices.value.bandwidth_price.unit_price_discount} ${t('元')}`
            : '--',
        },
      ];
    });
    const totalPrice = computed(() => {
      const instancePrice = prices.value?.instance_price?.unit_price_discount || 0;
      const bandwidthPrice = prices.value?.bandwidth_price?.unit_price_discount || 0;
      return (instancePrice + bandwidthPrice).toFixed(2);
    });

    const isInquiryPricesLoading = ref(false);
    const isInquiryPrices = computed(() => {
      const {
        account_id,
        region,
        cloud_vpc_id,
        cloud_subnet_id,
        require_count,
        name,
        load_balancer_type,
        account_type,
        address_ip_version,
        vip_isp,
        sla_type,
        internet_charge_type,
        internet_max_bandwidth_out,
        load_balancer_pass_to_target,
      } = props.formModel;

      // 基本验证
      const hasRequiredFields =
        account_id &&
        region &&
        load_balancer_pass_to_target !== undefined &&
        require_count !== 0 &&
        name &&
        /^[a-zA-Z0-9]([-a-zA-Z0-9]{0,58})[a-zA-Z0-9]$/.test(name);

      if (!hasRequiredFields) return false;

      // 内网负载均衡
      if (load_balancer_type === 'INTERNAL') {
        return Boolean(cloud_vpc_id && cloud_subnet_id);
      }

      // 公网负载均衡 - 传统账号
      if (account_type === 'LEGACY') {
        return Boolean(address_ip_version && cloud_vpc_id && vip_isp && sla_type);
      }

      // 公网负载均衡 - 标准账号
      return Boolean(
        address_ip_version && cloud_vpc_id && vip_isp && sla_type && internet_charge_type && internet_max_bandwidth_out,
      );
    });
    const inquiryPrices = async () => {
      isInquiryPricesLoading.value = true;
      const { formModel } = props;
      try {
        // eslint-disable-next-line prefer-const
        let zones = formModel.zones ? [formModel.zones] : [];
        const backup_zones = formModel.backup_zones ? [formModel.backup_zones] : undefined;
        const bandwidthpkg_sub_type = BGP_VIP_ISP_TYPES.includes(formModel.vip_isp) ? 'BGP' : 'SINGLEISP';

        const { data } = await businessStore.lbPricesInquiry({
          ...formModel,
          bk_biz_id: isBusinessPage ? formModel.bk_biz_id : undefined,
          zones,
          backup_zones,
          bandwidthpkg_sub_type,
          bandwidth_package_id: undefined,
        });

        prices.value = data;
      } catch (error) {
        console.error(error);
        return Promise.reject(error);
      } finally {
        isInquiryPricesLoading.value = false;
      }
    };
    watch(
      () => props.formModel,
      debounce(() => {
        if (isInquiryPrices.value) {
          inquiryPrices();
        } else {
          prices.value = { bandwidth_price: null, instance_price: null, lcu_price: null };
        }
      }, 500),
      { deep: true },
    );

    const applyLoading = ref(false);
    const isOpen = computed(() => props.formModel.load_balancer_type === 'OPEN');
    const isIpv4 = computed(() => props.formModel.address_ip_version === 'IPV4');
    const hasZonesConfig = computed(() => (isOpen.value && isIpv4.value) || !isOpen.value);
    const hasBackupZonesConfig = computed(() => isOpen.value && isIpv4.value);
    const hasInternetChargeTypeConfig = computed(() => isOpen.value && props.formModel.account_type !== 'LEGACY');

    // 提交
    const handleParams = () => {
      const { formModel } = props;
      const zones = hasZonesConfig.value ? (formModel.zones ? [formModel.zones] : []) : undefined;
      const vipIsp = isOpen.value ? formModel.vip_isp : undefined;

      return {
        ...formModel,
        bk_biz_id: isBusinessPage ? formModel.bk_biz_id : undefined,
        sla_type: formModel.sla_type === 'shared' ? '' : formModel.sla_type,
        // 只有公网下可以配置
        address_ip_version: isOpen.value ? formModel.address_ip_version : undefined,
        vip_isp: vipIsp,
        // 只有公网下的标准账号可以配置（内网支持配置带宽上限）
        internet_charge_type: hasInternetChargeTypeConfig.value ? formModel.internet_charge_type : undefined,
        internet_max_bandwidth_out:
          hasInternetChargeTypeConfig.value || !isOpen.value ? formModel.internet_max_bandwidth_out : undefined,
        // 只有公网下ipv4以及内网下可以配置
        zones,
        // 只有公网下ipv4可以配置
        // eslint-disable-next-line no-nested-ternary
        backup_zones: hasBackupZonesConfig.value ? (formModel.backup_zones ? [formModel.backup_zones] : []) : undefined,
        // 内网/公网IPv6需要选择子网
        cloud_subnet_id:
          !isOpen.value || (isOpen.value && formModel.address_ip_version === 'IPv6FullChain')
            ? formModel.cloud_subnet_id
            : undefined,
        // 内网下支持EIP
        cloud_eip_id: !isOpen.value ? formModel.cloud_eip_id ?? undefined : undefined,
        // 后端无用字段
        account_type: undefined as undefined,
        zoneType: undefined as undefined,
        slaType: undefined as undefined,
      };
    };
    const handleApplyClb = async () => {
      await props.formRef.validate();
      applyLoading.value = true;
      try {
        const { formModel } = props;
        const url = isBusinessPage
          ? `/api/v1/cloud/vendors/${formModel.vendor}/applications/types/create_load_balancer`
          : `/api/v1/cloud/load_balancers/create`;

        const res: IQueryResData<{ id: string }> = await http.post(url, handleParams());
        const { id } = res.data || {};

        applyClbSuccessHandler(isBusinessPage, goBack, { ...formModel, id });
      } catch (error) {
        console.error(error);
        return Promise.reject(error);
      } finally {
        applyLoading.value = false;
      }
    };
    const goBack = () => {
      const path =
        whereAmI.value === Senarios.business
          ? `/business/loadbalancer/clb-view?bizs=${route.query[GLOBAL_BIZS_KEY]}`
          : '/resource/resource?type=clb';

      router.replace({ path, query: { ...route.query } });
    };

    return () => (
      <div class='apply-clb-bottom-bar'>
        <div class='info-wrap'>
          <bk-popover theme='light' width={362} placement='top' offset={12}>
            {{
              default: () => <span class='label has-tips'>{t('配置费用')}</span>,
              content: () => (
                <bk-table data={priceTableData.value}>
                  <bk-table-column field='billingItem' label={t('计费项')}></bk-table-column>
                  <bk-table-column field='billingMode' label={t('计费模式')}></bk-table-column>
                  <bk-table-column field='price' label={t('价格')} align='right'></bk-table-column>
                </bk-table>
              ),
            }}
          </bk-popover>
          :
          <bk-loading loading={isInquiryPricesLoading.value} size='small' opacity={1} color='#fafbfd' class='value'>
            <span class='number'>{totalPrice.value}</span>
            {/* 本期只支持按量计费, 按照按量计费的模式进行单位显示 */}
            <span class='unit'>{t('元/小时')}</span>
          </bk-loading>
        </div>
        <div class='operation-btn-wrap'>
          <hcm-auth sign={{ type: createClbAuthSymbol.value, relation: [computedBizId.value] }} class='mr8'>
            {{
              default: ({ noPerm }: { noPerm: boolean }) => (
                <bk-button
                  theme='primary'
                  onClick={handleApplyClb}
                  loading={applyLoading.value}
                  disabled={isInquiryPricesLoading.value || noPerm}>
                  {t('立即购买')}
                </bk-button>
              ),
            }}
          </hcm-auth>
          <bk-button loading={applyLoading.value} onClick={goBack}>
            {t('取消')}
          </bk-button>
        </div>
      </div>
    );
  },
});
