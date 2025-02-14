import { Dialog, Message } from 'bkui-vue';
import { PropType, defineComponent, ref, computed, h, reactive, watch } from 'vue';
import './index.scss';
import { useResourceStore, useAccountStore, useBusinessStore } from '@/store';

import ChargePersonSelector from '@/components/charge-person-selector/index.vue';
import { useI18n } from 'vue-i18n';
import {
  azureSourceAddressTypes,
  AzureSourceTypeArr,
  azureTargetAddressTypes,
  AzureTargetTypeArr,
} from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/azure';
import {
  awsSourceAddressTypes,
  AwsSourceTypeArr,
} from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/aws';
import {
  tcloudSourceAddressTypes,
  TcloudSourceTypeArr,
} from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/tcloud';
import { huaweiSourceAddressTypes } from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/huawei';
import { SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { useBusinessMapStore } from '@/store/useBusinessMap';

export const CloneSecurity = defineComponent({
  props: {
    filter: {
      type: Object as PropType<any>,
      default: { op: 'and', rules: [{ field: 'type', op: 'eq', value: 'ingress' }] },
    },
    data: {
      type: Object as PropType<any>,
    },
    isShow: {
      type: Boolean as PropType<any>,
      required: true,
    },
  },
  setup(props, { emit }) {
    const { t } = useI18n();
    const store = useAccountStore();
    const { getNameFromBusinessMap } = useBusinessMapStore();
    // tab 信息
    const types = [
      { name: 'ingress', label: t('入站规则') },
      { name: 'egress', label: t('出站规则') },
    ];

    const states = reactive<any>({
      datas: [],
      isLoading: true,
    });
    const filter = ref(props.filter);
    const personSelectorRef = ref(null);

    const vendor = computed(() => props?.data?.vendor);
    const business = computed(() => getNameFromBusinessMap(store.bizs));

    const inColumns: any = computed(() =>
      [
        {
          label: t('名称'),
          field: 'name',
          isShow: vendor.value === 'azure',
        },
        {
          label: t('优先级'),
          field: 'priority',
          isShow: vendor.value === 'huawei' || vendor.value === 'azure',
        },
        {
          label: t('源地址类型'),
          render({ data }: any) {
            const nowVendor = (vendor.value as VendorEnum) || VendorEnum.TCLOUD;
            const sourceMap: any = {
              [VendorEnum.AWS]: {
                types: awsSourceAddressTypes,
                arr: AwsSourceTypeArr,
              },
              [VendorEnum.AZURE]: {
                types: azureSourceAddressTypes,
                arr: AzureSourceTypeArr,
              },
              [VendorEnum.HUAWEI]: {
                types: huaweiSourceAddressTypes,
                arr: TcloudSourceTypeArr,
              },
              [VendorEnum.TCLOUD]: {
                types: tcloudSourceAddressTypes,
                arr: TcloudSourceTypeArr,
              },
            };
            const { types } = sourceMap[nowVendor];
            const { arr } = sourceMap[nowVendor];
            const map = new Map(types.map((item: { value: string; label: string }) => [item.value, item.label]));
            let k = '';
            arr.forEach((type: string) => data[type] && (k = type));
            return map.get(k) || '--';
          },
          isShow: true,
        },
        {
          label: t('源地址'),
          render({ data }: any) {
            return h('span', {}, [
              data.cloud_address_group_id ||
                data.cloud_address_id ||
                data.cloud_service_group_id ||
                data.cloud_target_security_group_id ||
                data.ipv4_cidr ||
                data.ipv6_cidr ||
                data.cloud_remote_group_id ||
                data.remote_ip_prefix ||
                (data.source_address_prefix === '*' ? t('ALL') : data.source_address_prefix) ||
                data.source_address_prefixes ||
                data.cloud_source_security_group_ids ||
                data.destination_address_prefix ||
                data.destination_address_prefixes ||
                data.cloud_destination_security_group_ids ||
                (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
            ]);
          },
          isShow: true,
        },
        {
          label: t('源端口'),
          render({ data }: any) {
            return (data.source_port_range === '*' ? 'ALL' : data.source_port_range) || '--';
          },
          isShow: vendor.value === 'azure',
        },
        {
          label: t('目标地址类型'),
          render({ data }: any) {
            const map = new Map(
              azureTargetAddressTypes.map((item: { value: string; label: string }) => [item.value, item.label]),
            );
            let k = '';
            AzureTargetTypeArr.forEach((type: string) => data[type] && (k = type));
            return map.get(k) || '--';
          },
          isShow: vendor.value === 'azure',
        },
        {
          label: t('类型'),
          field: 'ethertype',
          isShow: vendor.value === 'huawei',
        },

        {
          label: t('目标地址'),
          render({ data }: any) {
            return (
              (data.destination_address_prefix === '*' ? t('ALL') : data.destination_address_prefix) ||
              data.destination_address_prefixes ||
              data.cloud_destination_security_group_ids
            );
          },
          isShow: vendor.value === 'azure',
        },
        {
          label: vendor.value === 'azure' ? t('目标端口协议类型') : t('协议'),
          render({ data }: any) {
            return h('span', {}, [
              data.cloud_service_id ||
                (vendor.value === 'aws' && data.protocol === '-1'
                  ? t('ALL')
                  : vendor.value === 'huawei' && !data.protocol
                  ? t('ALL')
                  : vendor.value === 'azure' && data.protocol === '*'
                  ? t('ALL')
                  : `${data.protocol}`),
            ]);
          },
          isShow: true,
        },
        {
          label: vendor.value === 'azure' ? t('目标协议端口') : t('端口'),
          render({ data }: any) {
            return h('span', {}, [
              data.cloud_service_id ||
                (vendor.value === 'aws' && data.to_port === -1
                  ? t('ALL')
                  : vendor.value === 'huawei' && !data.port
                  ? t('ALL')
                  : vendor.value === 'azure' && data.destination_port_range === '*'
                  ? t('ALL')
                  : `${
                      data.port || data.to_port || data.destination_port_range || data.destination_port_ranges || '--'
                    }`),
            ]);
          },
          isShow: true,
        },
        {
          label: t('策略'),
          render({ data }: any) {
            return h('span', {}, [
              vendor.value === 'huawei'
                ? HuaweiSecurityRuleEnum[data.action]
                : vendor.value === 'azure'
                ? AzureSecurityRuleEnum[data.access]
                : vendor.value === 'aws'
                ? t('允许')
                : SecurityRuleEnum[data.action] || '--',
            ]);
          },
          isShow: vendor.value !== 'aws',
        },
        {
          label: t('备注'),
          field: 'memo',
          render: ({ data }) => data.memo || '--',
          isShow: true,
        },
      ].filter(({ isShow }) => !!isShow),
    );

    // 出站规则列字段
    const outColumns: any = computed(() =>
      [
        {
          label: t('名称'),
          field: 'name',
          isShow: vendor.value === 'azure',
        },
        {
          label: t('优先级'),
          field: 'priority',
          isShow: vendor.value === 'huawei' || vendor.value === 'azure',
        },
        {
          label: t('源地址类型'),
          render({ data }: any) {
            const map = new Map(
              azureSourceAddressTypes.map((item: { value: string; label: string }) => [item.value, item.label]),
            );
            let k = '';
            AzureSourceTypeArr.forEach((type: string) => data[type] && (k = type));
            return map.get(k) || '--';
          },
          isShow: vendor.value === 'azure',
        },
        {
          label: t('源地址'),
          render({ data }: any) {
            return h('span', {}, [
              data.cloud_address_group_id ||
                data.cloud_address_id ||
                data.cloud_service_group_id ||
                data.cloud_target_security_group_id ||
                data.ipv4_cidr ||
                data.ipv6_cidr ||
                data.cloud_remote_group_id ||
                data.remote_ip_prefix ||
                (data.source_address_prefix === '*' ? t('ALL') : data.source_address_prefix) ||
                data.source_address_prefixes ||
                data.cloud_source_security_group_ids ||
                data.destination_address_prefix ||
                data.destination_address_prefixes ||
                data.cloud_destination_security_group_ids ||
                (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
            ]);
          },
          isShow: vendor.value === 'azure',
        },
        {
          label: t('源端口'),
          render({ data }: any) {
            return (data.source_port_range === '*' ? 'ALL' : data.source_port_range) || '--';
          },
          isShow: vendor.value === 'azure',
        },
        {
          label: t('目标地址类型'),
          render({ data }: any) {
            const nowVendor = vendor.value as VendorEnum;
            const targetMap: any = {
              [VendorEnum.AWS]: {
                types: awsSourceAddressTypes,
                arr: AwsSourceTypeArr,
              },
              [VendorEnum.AZURE]: {
                types: azureTargetAddressTypes,
                arr: AzureTargetTypeArr,
              },
              [VendorEnum.HUAWEI]: {
                types: huaweiSourceAddressTypes,
                arr: TcloudSourceTypeArr,
              },
              [VendorEnum.TCLOUD]: {
                types: tcloudSourceAddressTypes,
                arr: TcloudSourceTypeArr,
              },
            };
            const { types } = targetMap[nowVendor];
            const { arr } = targetMap[nowVendor];
            const map = new Map(types.map((item: { value: string; label: string }) => [item.value, item.label]));
            let k = '';
            arr.forEach((type: string) => data[type] && (k = type));
            return map.get(k) || '--';
          },
          isShow: true,
        },
        {
          label: t('目标地址'),
          render({ data }: any) {
            return h('span', {}, [
              data.cloud_address_group_id ||
                data.cloud_address_id ||
                data.cloud_service_group_id ||
                data.cloud_target_security_group_id ||
                data.ipv4_cidr ||
                data.ipv6_cidr ||
                data.cloud_remote_group_id ||
                data.remote_ip_prefix ||
                data.cloud_source_security_group_ids ||
                (data.destination_address_prefix === '*' ? t('ALL') : data.destination_address_prefix) ||
                data.destination_address_prefixes ||
                data.cloud_destination_security_group_ids ||
                (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
            ]);
          },
          isShow: true,
        },
        {
          label: t('类型'),
          field: 'ethertype',
          isShow: vendor.value === 'huawei',
        },
        {
          label: vendor.value === 'azure' ? t('目标端口协议类型') : t('协议'),
          render({ data }: any) {
            return h('span', {}, [
              data.cloud_service_id ||
                (vendor.value === 'aws' && data.protocol === '-1'
                  ? t('ALL')
                  : vendor.value === 'huawei' && !data.protocol
                  ? t('ALL')
                  : vendor.value === 'azure' && data.protocol === '*'
                  ? t('ALL')
                  : `${data.protocol}`),
            ]);
          },
          isShow: true,
        },
        {
          label: vendor.value === 'azure' ? t('目标协议端口') : t('端口'),
          render({ data }: any) {
            return h('span', {}, [
              data.cloud_service_id ||
                (vendor.value === 'aws' && data.to_port === -1
                  ? t('ALL')
                  : vendor.value === 'huawei' && !data.port
                  ? t('ALL')
                  : vendor.value === 'azure' && data.destination_port_range === '*'
                  ? t('ALL')
                  : `${data.port || data.to_port || data.destination_port_range || '--'}`),
            ]);
          },
          isShow: true,
        },
        {
          label: t('策略'),
          render({ data }: any) {
            return h('span', {}, [
              vendor.value === 'huawei'
                ? HuaweiSecurityRuleEnum[data.action]
                : vendor.value === 'azure'
                ? AzureSecurityRuleEnum[data.access]
                : vendor.value === 'aws'
                ? t('允许')
                : SecurityRuleEnum[data.action] || '--',
            ]);
          },
          isShow: vendor.value !== 'aws',
        },
        {
          label: t('备注'),
          field: 'memo',
          render: ({ data }) => data.memo || '--',
          isShow: true,
        },
      ].filter(({ isShow }) => !!isShow),
    );

    const activeType = ref('ingress');
    const isLoading = ref(false);
    const useBusiness = useBusinessStore();
    const resourceStore = useResourceStore();

    const getList = async () => {
      try {
        const list = await resourceStore.getAllSort({
          id: props?.data?.id,
          vendor: vendor.value,
          filter: filter.value,
        });
        states.datas = list;
        return list;
      } catch {
        states.datas = [];
      } finally {
        states.isLoading = false;
      }
    };
    const handleClose = () => {
      emit('update:isShow', false);
    };
    const handleConfirm = async () => {
      const { id } = props.data;
      const { formData: personSelectorParams, validate } = personSelectorRef.value;
      const { bak_manager, manager } = personSelectorParams;
      await validate();
      isLoading.value = true;
      try {
        await useBusiness.cloneSecurity({
          id,
          bak_manager,
          manager,
        });
        Message({
          theme: 'success',
          message: t('克隆成功！'),
        });
        handleClose();
      } catch (error) {
        Message({
          theme: 'error',
          message: t('克隆失败！'),
        });
      } finally {
        isLoading.value = false;
      }
    };

    watch(
      () => props.isShow,
      (val: boolean) => {
        if (val) getList();
      },
      {
        immediate: true,
      },
    );
    watch(
      () => activeType.value,
      (val: string) => {
        states.isLoading = true;
        filter.value.rules[0].value = val;
        getList();
      },
    );
    return () => (
      <>
        <Dialog
          width={960}
          class={'clone-security-dialog'}
          isShow={props.isShow}
          title={t(`克隆安全组`)}
          theme={'primary'}
          onClosed={handleClose}
          onConfirm={handleConfirm}
          isLoading={isLoading.value}
          render-directive={'if'}>
          <div class={'security-info'}>
            <div>
              安全组名称：<span>{props.data.name}</span>
            </div>
            <div>
              管理业务：<span>{business.value}</span>
            </div>
            <div>
              使用业务：<span>{business.value}</span>
            </div>
          </div>
          <ChargePersonSelector
            ref={personSelectorRef}
            manager={props?.data?.manager}
            bakManager={props?.data?.bak_manager}></ChargePersonSelector>
          <div class={'security-rule'}>
            <div class={'title'}>安全组规则</div>
            <section class={'rule-main'}>
              <bk-radio-group v-model={activeType.value}>
                {types.map(({ name, label }) => (
                  <bk-radio-button key={name} label={name}>
                    {label}
                  </bk-radio-button>
                ))}
              </bk-radio-group>
            </section>
            <bk-table
              class={'mt20'}
              row-hover={'auto'}
              remote-pagination
              columns={activeType.value === 'ingress' ? inColumns.value : outColumns.value}
              data={states.datas}
              show-overflow-tooltip>
              {{
                empty: () => {
                  return (
                    <div class={'security-empty-container'}>
                      <bk-exception
                        class={'exception-wrap-item exception-part'}
                        type={'empty'}
                        scene={'part'}
                        description={'无规则，默认拒绝所有流量'}
                      />
                    </div>
                  );
                },
              }}
            </bk-table>
          </div>
        </Dialog>
      </>
    );
  },
});
