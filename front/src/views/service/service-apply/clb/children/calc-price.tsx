import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useBusinessStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { LbPrice } from '@/typings';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { debounce } from 'lodash';
import { BGP_VIP_ISP_TYPES } from '@/constants';
import cssModule from '../index.module.scss';

import { Form } from 'bkui-vue';

export default defineComponent({
  props: {
    formModel: Object as PropType<ApplyClbModel>,
    formRef: Object as PropType<InstanceType<typeof Form>>,
    showBuy: Boolean,
  },
  setup(props) {
    const { t } = useI18n();
    const { isBusinessPage } = useWhereAmI();
    const businessStore = useBusinessStore();

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

    return () => (
      <div class={cssModule.applyClbBottomBar}>
        <div class={cssModule.infoWrap}>
          <bk-popover theme='light' width={362} placement='top' offset={12}>
            {{
              default: () => <span class={[cssModule.label, cssModule.hasTips]}>{t('配置费用')}</span>,
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
          <bk-loading
            loading={isInquiryPricesLoading.value}
            size='small'
            opacity={1}
            color='#fafbfd'
            class={cssModule.value}>
            <span class={cssModule.number}>{totalPrice.value}</span>
            {/* 本期只支持按量计费, 按照按量计费的模式进行单位显示 */}
            <span class={cssModule.unit}>{t('元/小时')}</span>
          </bk-loading>
        </div>
      </div>
    );
  },
});
