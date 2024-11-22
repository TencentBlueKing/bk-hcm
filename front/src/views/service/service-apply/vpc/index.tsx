import { computed, defineComponent, ref, watch } from 'vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import ConditionOptions from '../components/common/condition-options/index.vue';
import CloudAreaSelector from '../components/common/cloud-area-selector';
import ZoneSelector from '../components/common/zone-selector';
import { Form, Input, Checkbox, Button, Radio } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import './index.scss';

import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import useCondtion from '../hooks/use-condtion';
import useVpcFormData from '../hooks/use-vpc-form-data';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useRouter } from 'vue-router';
import { SubnetInput } from '@/components/subnet-input';
import { IP_RANGES } from './contansts';
import { Info } from 'bkui-vue/lib/icon';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';

const { FormItem } = Form;
const { Group: RadioGroup } = Radio;

export default defineComponent({
  props: {},
  setup() {
    const { cond, isEmptyCond } = useCondtion();
    const { isResourcePage, whereAmI } = useWhereAmI();
    const { formData, formRef, handleFormSubmit, submitting } = useVpcFormData(cond);
    const resourceAccountStore = useResourceAccountStore();
    const { t } = useI18n();
    const router = useRouter();

    const curIpRef = ref();
    const subIpRef = ref();

    const curCIDR = ref(IP_RANGES[VendorEnum.TCLOUD][0]);
    const subCIDR = ref(IP_RANGES[VendorEnum.TCLOUD][0]);

    const curRange = ref({
      idx: 0,
      range: IP_RANGES[VendorEnum.TCLOUD],
    });
    const subRange = ref({
      idx: 0,
      range: [curCIDR.value],
    });

    watch(
      () => curCIDR.value,
      (val) => {
        subRange.value = {
          idx: 0,
          range: [val],
        };
        subCIDR.value = val;
        formData.ipv4_cidr = `${val.ip}/${val.mask}`;
      },
      {
        deep: true,
      },
    );

    watch(
      () => subCIDR.value,
      (val) => {
        formData.subnet.ipv4_cidr = `${val.ip}/${val.mask}`;
      },
      {
        deep: true,
      },
    );

    watch(
      () => cond.vendor,
      () => {
        if (VendorEnum.GCP === cond.vendor) {
          subRange.value = {
            idx: 0,
            range: IP_RANGES[VendorEnum.GCP],
          };
          subCIDR.value = IP_RANGES[VendorEnum.GCP][0];
        }
      },
      {
        immediate: true,
      },
    );

    watch([() => resourceAccountStore.resourceAccount?.id, whereAmI.value], () => {
      if (whereAmI.value === Senarios.resource) {
        curIpRef.value?.reset();
        subIpRef.value?.reset();
      }
    });

    const submitDisabled = computed(() => isEmptyCond.value);

    const nameReg = /^(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?)$/;
    const nameRegMsg = '名称需要符合如下正则表达式: /(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?)/';

    const networkTips = computed(() => {
      const map = {
        [VendorEnum.TCLOUD]:
          'CIDR范围的有效范围为:\t\n10.0.0.0 - 10.255.255.255（掩码范围需在12 - 28之间）\t\n172.16.0.0 - 172.31.255.255（掩码范围需在12 - 28之间）\t\n192.168.0.0 - 192.168.255.255 （掩码范围需在16 - 28之间）\t\n更多信息请参考官方说明https://cloud.tencent.com/document/product/215/36515',
        [VendorEnum.AWS]:
          'CIDR范围的有效范围为:\t\n10.0.0.0 - 10.255.255.255（10/8 前缀）\t\n172.16.0.0 - 172.31.255.255（172.16/12 前缀）\t\n192.168.0.0 - 192.168.255.255（192.168/16 前缀）\t\n更多信息请参考官方说明https://docs.aws.amazon.com/zh_cn/vpc/latest/userguide/configure-your-vpc.html#add-cidr-block-restrictions',
        [VendorEnum.AZURE]:
          'CIDR范围的有效范围为:\t\n10.0.0.0 - 10.255.255.255（10/8 前缀）\t\n172.16.0.0 - 172.31.255.255（172.16/12 前缀）\t\n192.168.0.0 - 192.168.255.255（192.168/16 前缀）\t\n更多信息请参考官方说明https://learn.microsoft.com/zh-cn/azure/virtual-network/virtual-networks-faq#what-address-ranges-can-i-use-in-my-vnets',
        [VendorEnum.GCP]:
          'CIDR范围的有效范围为:\t\n10.0.0.0/8\t\n172.16.0.0/12\t\n192.168.0.0/16\t\n更多信息请参考官方说明https://cloud.google.com/vpc/docs/subnets?hl=zh-cn',
        [VendorEnum.HUAWEI]:
          'CIDR范围的有效范围为:\t\n10.0.0.0/8~28\t\n172.16.0.0/12~28\t\n192.168.0.0/16~28\t\n更多信息请参考官方说明https://support.huaweicloud.com/intl/zh-cn/usermanual-vpc/zh-cn_topic_0013935842.html',
      };
      return map[cond.vendor];
    });

    const subnetTips = computed(() => {
      const map = {
        [VendorEnum.GCP]:
          'CIDR范围的有效范围为:\t\n10.0.0.0/8\t\n172.16.0.0/12\t\n192.168.0.0/16\t\n更多信息请参考官方说明https://cloud.google.com/vpc/docs/subnets?hl=zh-cn',
      };
      return map[cond.vendor] || '请确保所填写的子网CIDR在VPC CIDR中';
    });

    const formConfig = computed(() => [
      {
        id: 'type',
        title: 'VPC类型',
        display: cond.vendor === VendorEnum.AWS,
        children: [
          {
            label: '类型',
            content: () => (
              <RadioGroup v-model={formData.type}>
                <Radio label={0}>基本配置</Radio>
              </RadioGroup>
            ),
          },
        ],
      },
      {
        id: 'network',
        title: 'VPC网络信息',
        children: [
          {
            label: '名称',
            required: true,
            property: 'name',
            maxlength: 60,
            description: nameRegMsg,
            content: () => <Input placeholder='填写VPC网络的名称' v-model={formData.name}></Input>,
          },
          {
            label: 'IP来源类型',
            display: [VendorEnum.TCLOUD, VendorEnum.AZURE, VendorEnum.HUAWEI].includes(cond.vendor),
            required: true,
            property: 'ip_source_type',
            content: () => (
              <RadioGroup v-model={formData.ip_source_type}>
                <Radio label={0}>业务私有</Radio>
                <Radio label={1} disabled={true}>
                  IP池
                </Radio>
              </RadioGroup>
            ),
          },
          {
            label: 'IPv4 CIDR',
            display: cond.vendor !== VendorEnum.GCP,
            content: () => (
              <div class={'flex-row align-item-center'}>
                <SubnetInput
                  disabled={!cond.vendor}
                  ref={curIpRef}
                  ips={curRange.value}
                  v-model={curCIDR.value}
                  onChangeIdx={(idx) => {
                    curCIDR.value = IP_RANGES[cond.vendor][idx];
                    curRange.value.idx = idx;
                  }}
                />
                <Info
                  class={'ml6'}
                  v-BkTooltips={{
                    content: networkTips.value ? networkTips.value : '请先选择云厂商',
                  }}></Info>
              </div>
            ),
          },
          {
            label: '管控区域',
            description:
              '管控区是蓝鲸可以管控的Agent网络区域，以实现跨网管理。\n一个VPC，对应一个管控区。如VPC未绑定管控区，请到资源接入-VPC-绑定管控区操作。',
            required: true,
            property: 'bk_cloud_id',
            content: () => <CloudAreaSelector v-model={formData.bk_cloud_id} />,
          },
          {
            label: 'BastionHost',
            display: cond.vendor === VendorEnum.AZURE,
            property: 'bastion_host_enable',
            content: () => (
              <RadioGroup v-model={formData.bastion_host_enable}>
                <Radio label={false}>禁用</Radio>
                <Radio label={true} disabled={true}>
                  暂不支持启用
                </Radio>
              </RadioGroup>
            ),
          },
          {
            label: 'DDoS 保护标准',
            display: cond.vendor === VendorEnum.AZURE,
            property: 'ddos_enable',
            content: () => (
              <RadioGroup v-model={formData.ddos_enable}>
                <Radio label={false}>禁用</Radio>
                <Radio label={true} disabled={true}>
                  暂不支持启用
                </Radio>
              </RadioGroup>
            ),
          },
          {
            label: '防火墙',
            display: cond.vendor === VendorEnum.AZURE,
            property: 'firewall_enable',
            content: () => (
              <RadioGroup v-model={formData.firewall_enable}>
                <Radio label={false}>禁用</Radio>
                <Radio label={true} disabled={true}>
                  暂不支持启用
                </Radio>
              </RadioGroup>
            ),
          },
          {
            label: '租期',
            display: cond.vendor === VendorEnum.AWS,
            required: true,
            property: 'instance_tenancy',
            content: () => (
              <RadioGroup v-model={formData.instance_tenancy}>
                <Radio label={'default'}>默认</Radio>
                <Radio label={'dedicated'}>专用</Radio>
              </RadioGroup>
            ),
          },
          {
            label: '企业项目',
            display: cond.vendor === VendorEnum.HUAWEI,
            content: () => <span>default</span>,
          },
          {
            label: '动态路由模式',
            display: cond.vendor === VendorEnum.GCP,
            required: true,
            content: () => (
              <RadioGroup v-model={formData.routing_mode}>
                <Radio label={'REGIONAL'}>区域</Radio>
                <Radio label={'GLOBAL'}>全局</Radio>
              </RadioGroup>
            ),
          },
        ],
      },
      {
        id: 'subnet',
        title: '初始子网信息',
        display: cond.vendor !== VendorEnum.AWS,
        children: [
          {
            label: '名称',
            required: true,
            property: 'subnet.name',
            maxlength: 60,
            description: nameRegMsg,
            content: () => <Input placeholder='填写子网的名称' v-model={formData.subnet.name} />,
          },
          {
            label: 'IPv4 CIDR',
            property: 'ipv4Cidr2',
            content: () => (
              <>
                <div class='flex-row align-items-center'>
                  <SubnetInput
                    ips={subRange.value}
                    ref={subIpRef}
                    v-model={subCIDR.value}
                    disabled={!cond.vendor}
                    isSub
                    onChangeIdx={(idx) => {
                      subCIDR.value = IP_RANGES[cond.vendor][idx];
                      subRange.value.idx = idx;
                    }}
                  />
                  <Info v-BkTooltips={{ content: subnetTips.value }} class={'ml6'}></Info>
                </div>
              </>
            ),
          },
          {
            label: '可用区',
            display: [VendorEnum.TCLOUD].includes(cond.vendor),
            required: true,
            property: 'subnet.zone',
            description: '同一私有网络下可以有不同可用区的子网，同一私有网络下不同可用区的子网默认可以内网互通',
            content: () => <ZoneSelector v-model={formData.subnet.zone} vendor={cond.vendor} region={cond.region} />,
          },
          {
            label: '子网IPv6网段',
            display: cond.vendor === VendorEnum.HUAWEI,
            content: () => <Checkbox v-model={formData.subnet.ipv6_enable}>开启IPv6</Checkbox>,
          },
          {
            label: '关联路由表',
            display: [VendorEnum.TCLOUD, VendorEnum.HUAWEI].includes(cond.vendor),
            content: () => <span>默认</span>,
          },
          {
            label: '专用访问通道',
            display: cond.vendor === VendorEnum.GCP,
            content: () => (
              <RadioGroup v-model={formData.subnet.private_ip_google_access}>
                <Radio label={false}>禁用</Radio>
                <Radio label={true}>启用</Radio>
              </RadioGroup>
            ),
          },
          {
            label: '流日志',
            display: cond.vendor === VendorEnum.GCP,
            content: () => (
              <RadioGroup v-model={formData.subnet.enable_flow_logs}>
                <Radio label={false}>禁用</Radio>
                <Radio label={true}>启用</Radio>
              </RadioGroup>
            ),
          },
          {
            label: '防火墙规则',
            display: cond.vendor === VendorEnum.GCP,
            content: () => (
              <span>默认防火墙规则是可以出，不允许进入。如需绑定防火墙规则，请在创建VPC后，进入VPC管理页面绑定。</span>
            ),
          },
        ],
      },
    ]);

    const formRules = {
      name: [
        {
          pattern: nameReg,
          message: nameRegMsg,
          trigger: 'change',
        },
      ],
      'subnet.name': [
        {
          pattern: nameReg,
          message: nameRegMsg,
          trigger: 'change',
        },
      ],
    };

    return () => (
      <div>
        <DetailHeader>
          <p class={'purchase-vpc-header-title'}>购买VPC</p>
        </DetailHeader>
        <div class='create-form-container' style={isResourcePage && { padding: 0 }}>
          <Form model={formData} rules={formRules} ref={formRef} onSubmit={handleFormSubmit} formType='vertical'>
            <ConditionOptions
              type={ResourceTypeEnum.VPC}
              bizs={cond.bizId}
              v-model:cloudAccountId={cond.cloudAccountId}
              v-model:vendor={cond.vendor}
              v-model:region={cond.region}
              v-model:resourceGroup={cond.resourceGroup}
            />
            {formConfig.value
              .filter(({ display }) => display !== false)
              .map(({ title, children }) => (
                <CommonCard title={() => title} class={'mb16'}>
                  {children
                    .filter(({ display }) => display !== false)
                    .map(({ label, description, tips, required, property, content }) => (
                      <FormItem label={label} required={required} property={property} description={description}>
                        {Array.isArray(content) ? (
                          <div class='flex-row'>
                            {content
                              .filter((sub) => sub.display !== false)
                              .map((sub) => (
                                <FormItem
                                  label={sub.label}
                                  required={sub.required}
                                  property={sub.property}
                                  description={sub?.description}>
                                  {sub.content()}
                                  {sub.tips && <div class='form-item-tips'>{sub.tips()}</div>}
                                </FormItem>
                              ))}
                          </div>
                        ) : (
                          content()
                        )}
                        {tips && <div class='form-item-tips'>{tips()}</div>}
                      </FormItem>
                    ))}
                </CommonCard>
              ))}
          </Form>
        </div>
        <div class='action-bar' style={{ paddingLeft: isResourcePage && 'calc(15% + 24px)' }}>
          <Button
            theme='primary'
            loading={submitting.value}
            disabled={submitDisabled.value}
            class={'mr8'}
            onClick={handleFormSubmit}>
            {isResourcePage ? t('提交') : t('提交审批')}
          </Button>
          <Button onClick={() => router.back()}>{t('取消')}</Button>
        </div>
      </div>
    );
  },
});
