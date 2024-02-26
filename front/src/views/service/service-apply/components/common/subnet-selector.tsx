import http from '@/http';
import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Button, Select } from 'bkui-vue';

import { QueryRuleOPEnum } from '@/typings/common';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ISubnetItem } from '../../cvm/children/SubnetPreviewDialog';
import RightTurnLine from 'bkui-vue/lib/icon/right-turn-line';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { Option } = Select;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    bizId: Number as PropType<number | string>,
    vpcId: String as PropType<string>,
    vendor: String as PropType<string>,
    region: String as PropType<string>,
    accountId: String as PropType<string>,
    zone: String as PropType<string>,
    resourceGroup: String as PropType<string>,
    handleChange: Function as PropType<(data: ISubnetItem) => void>,
    clearable: {
      type: Boolean,
      default: true,
    },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs, expose }) {
    const list = ref([]);
    const loading = ref(false);
    const { isResourcePage, isServicePage } = useWhereAmI();

    expose({ subnetList: list });

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    const getSubnetsData = async (
      bizId: string | number,
      region: string,
      vendor: string,
      vpcId: string,
      accountId: string,
      zone: string,
    ) => {
      if ((!bizId && isServicePage) || !vpcId) {
        list.value = [];
        return;
      }

      loading.value = true;

      const filter = {
        op: 'and',
        rules: [
          {
            field: 'vpc_id',
            op: QueryRuleOPEnum.EQ,
            value: vpcId,
          },
          {
            field: 'account_id',
            op: QueryRuleOPEnum.EQ,
            value: accountId,
          },
          {
            field: 'region',
            op: QueryRuleOPEnum.EQ,
            value: region,
          },
        ],
      };

      if ([VendorEnum.TCLOUD, VendorEnum.AWS].includes(vendor)) {
        filter.rules.push({
          field: 'zone',
          op: QueryRuleOPEnum.EQ,
          value: zone,
        });
      }

      // if (vendor === VendorEnum.AZURE) {
      //   filter.rules.push({
      //     field: 'extension.resource_group_name',
      //     op: QueryRuleOPEnum.JSON_EQ,
      //     value: resourceGroup,
      //   });
      // }

      const result = await http.post(isResourcePage
        ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/subnets/with/ip_count/list`
        : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/bizs/${bizId}/subnets/with/ip_count/list`, {
        // const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/subnets/list`, {
        filter,
        page: {
          count: false,
          start: 0,
          limit: 50,
        },
      });
      list.value = result?.data?.details ?? [];
      loading.value = false;
    };

    watch([
      () => props.bizId,
      () => props.region,
      () => props.vendor,
      () => props.vpcId,
      () => props.accountId,
      () => props.zone,
      () => props.resourceGroup,
    ], async ([bizId, region, vendor, vpcId, accountId, zone]) => {
      await getSubnetsData(bizId, region, vendor, vpcId, accountId, zone);
    });

    return () => (
      <div>
        <Select
          filterable={true}
          modelValue={selected.value}
          onUpdate:modelValue={val => selected.value = val}
          loading={loading.value}
          clearable={props.clearable}
          {...{ attrs }}
          onChange={(cloud_id: string) => {
            console.log(cloud_id);
            const data = list.value.find(item => item.cloud_id === cloud_id);
            props.handleChange(data);
          }}
        >
          {
            list.value.map(({ cloud_id, name, ipv4_cidr, available_ip_count }) => (
              <Option key={cloud_id} value={cloud_id} label={`${name} ${ipv4_cidr} ${props.vendor !== VendorEnum.GCP ? `剩余IP ${available_ip_count}` : ''}`}></Option>
            ))
          }
        </Select>
        {props.vpcId && !list.value.length ? (
          <div class={'subnet-selector-tips'}>
            {/* {whereAmI.value === Senarios.resource ? ( */}
              <>
                <span class={'subnet-create-tips'}>
                  {'所选的VPC，在当前区无可用的子网，可切换VPC或'}
                </span>
                <Button
                  text
                  theme='primary'
                  class={'mr6'}
                  onClick={() => {
                    const url = '/#/resource/resource?type=subnet';
                    window.open(url, '_blank');
                  }}>
                  新建子网
                </Button>
              </>
            {/* ) : (
              <>
                <span class={'subnet-create-tips mr6'}>
                  该VPC下在本可用区存在子网，但未分配给本业务。
                  <Button
                    text
                    theme='primary'
                    onClick={() => {
                      const url = '/#/business/subnet';
                      window.open(url, '_blank');
                    }}>
                    新建子网
                  </Button>
                  ,或者在资源接入-子网中分配给本业务
                </span>
              </>
            )} */}
            <Button
              text
              onClick={() => {
                getSubnetsData(
                  props.bizId,
                  props.region,
                  props.vendor,
                  props.vpcId,
                  props.accountId,
                  props.zone,
                );
              }}>
              <RightTurnLine fill='#3A84FF' />
            </Button>
          </div>
        ) : null}
      </div>
    );
  },
});
