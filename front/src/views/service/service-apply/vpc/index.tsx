import { computed, defineComponent, watch, ref } from 'vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import ConditionOptions from '../components/common/condition-options.vue';
import CloudAreaSelector from '../components/common/cloud-area-selector';
import ZoneSelector from '../components/common/zone-selector';
import { Form, Input, Checkbox, Button, Radio, Select } from 'bkui-vue';
import { Info } from 'bkui-vue/lib/icon';
import { useI18n } from 'vue-i18n';
import './index.scss';

import {
  ResourceTypeEnum,
  VendorEnum,
  CIDRLIST,
  CIDRDATARANGE,
  CIDRMASKRANGE,
  TCLOUDCIDRMASKRANGE,
} from '@/common/constant';
// import useVpcOptions from '../hooks/use-vpc-options';
import useCondtion from '../hooks/use-condtion';
import useVpcFormData from '../hooks/use-vpc-form-data';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useRouter } from 'vue-router';

const { FormItem, ComposeFormItem } = Form;
const { Group: RadioGroup } = Radio;
const { Option } = Select;
const ipv4CidrFir: any = ref('10');

export default defineComponent({
  props: {},
  setup() {
    const { cond, isEmptyCond } = useCondtion(ResourceTypeEnum.VPC);
    const { isResourcePage } = useWhereAmI();
    const { formData, formRef, handleFormSubmit, submitting } =      useVpcFormData(cond);
    // const __ = useVpcOptions(cond, formData);
    const { t } = useI18n();
    const router = useRouter();

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
            content: () => (
              <Input
                placeholder='填写VPC网络的名称'
                v-model={formData.name}></Input>
            ),
          },
          {
            label: 'IP来源类型',
            display: [
              VendorEnum.TCLOUD,
              VendorEnum.AZURE,
              VendorEnum.HUAWEI,
            ].includes(cond.vendor),
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
            property: 'ipv4Cidr1',
            content: () => (
              <>
                <div class='flex-row align-items-center'>
                  <ComposeFormItem class='mr5'>
                    <div class='flex-row'>
                      <Select
                        class='w110'
                        clearable={false}
                        v-model={formData.ipv4_cidr[0]}>
                        {CIDRLIST.map(item => (
                          <Option
                            key={item.id}
                            value={item.id}
                            label={item.name}>
                            {item.name}
                          </Option>
                        ))}
                      </Select>
                      <div>.</div>
                      <Input
                        type='number'
                        disabled={ipv4CidrFir.value === '192'}
                        placeholder={`${CIDRDATARANGE[ipv4CidrFir.value].min}-${
                          CIDRDATARANGE[ipv4CidrFir.value].max
                        }`}
                        min={CIDRDATARANGE[ipv4CidrFir.value].min}
                        max={CIDRDATARANGE[ipv4CidrFir.value].max}
                        v-model={formData.ipv4_cidr[1]}
                        class='w110'
                      />
                      <div>.</div>
                      <Input
                        type='number'
                        min={0}
                        max={255}
                        v-model={formData.ipv4_cidr[2]}
                        class='w110'
                      />
                      <div>.</div>
                      <Input
                        min={0}
                        max={255}
                        type='number'
                        v-model={formData.ipv4_cidr[3]}
                        class='w110'
                      />
                      <div>/</div>
                      <Input
                        type='number'
                        placeholder={`${
                          cond.vendor === 'tcloud'
                            ? TCLOUDCIDRMASKRANGE[ipv4CidrFir.value].min
                            : CIDRMASKRANGE[ipv4CidrFir.value].min
                        }-${CIDRMASKRANGE[ipv4CidrFir.value].max}`}
                        min={
                          cond.vendor === 'tcloud'
                            ? TCLOUDCIDRMASKRANGE[ipv4CidrFir.value].min
                            : CIDRMASKRANGE[ipv4CidrFir.value].min
                        }
                        max={CIDRMASKRANGE[ipv4CidrFir.value].max}
                        v-model={formData.ipv4_cidr[4]}
                        class='w110'
                      />
                    </div>
                  </ComposeFormItem>
                  <Info
                    v-BkTooltips={{
                      content: networkTips.value
                        ? networkTips.value
                        : '请先选择云厂商',
                    }}></Info>
                </div>
              </>
            ),
          },
          {
            label: '管控区域',
            description: '管控区是蓝鲸可以管控的Agent网络区域，以实现跨网管理。\n一个VPC，对应一个管控区。如VPC未绑定管控区，请到资源接入-VPC-绑定管控区操作。',
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
            content: () => (
              <Input
                placeholder='填写子网的名称'
                v-model={formData.subnet.name}
              />
            ),
          },
          {
            label: 'IPv4 CIDR',
            property: 'ipv4Cidr2',
            content: () => (
              <>
                <div class='flex-row align-items-center'>
                  <ComposeFormItem class='mr5'>
                    <div class='flex-row'>
                      {cond.vendor === VendorEnum.GCP ? (
                        <Select
                          class='w110'
                          clearable={false}
                          v-model={formData.subnet.ipv4_cidr[0]}>
                          {CIDRLIST.map(item => (
                            <Option
                              key={item.id}
                              value={item.id}
                              label={item.name}>
                              {item.name}
                            </Option>
                          ))}
                        </Select>
                      ) : (
                        <Input
                          type='number'
                          disabled
                          placeholder='1-255'
                          min={1}
                          max={255}
                          v-model={formData.subnet.ipv4_cidr[0]}
                          class='w110'
                        />
                      )}
                      <div>.</div>
                      <Input
                        type='number'
                        disabled={cond.vendor !== VendorEnum.GCP}
                        placeholder='0-255'
                        min={0}
                        max={255}
                        v-model={formData.subnet.ipv4_cidr[1]}
                        class='w110'
                      />
                      <div>.</div>
                      <Input
                        type='number'
                        placeholder='0-255'
                        min={0}
                        max={255}
                        v-model={formData.subnet.ipv4_cidr[2]}
                        class='w110'
                      />
                      <div>.</div>
                      <Input
                        type='number'
                        placeholder='0-255'
                        min={0}
                        max={255}
                        v-model={formData.subnet.ipv4_cidr[3]}
                        class='w110'
                      />
                      <div>/</div>
                      <Input
                        type='number'
                        placeholder='1-32'
                        min={1}
                        max={32}
                        v-model={formData.subnet.ipv4_cidr[4]}
                        class='w110'
                      />
                    </div>
                  </ComposeFormItem>
                  <Info v-BkTooltips={{ content: subnetTips.value }}></Info>
                </div>
              </>
            ),
          },
          {
            label: '可用区',
            display: [VendorEnum.TCLOUD].includes(cond.vendor),
            required: true,
            property: 'subnet.zone',
            description:
              '同一私有网络下可以有不同可用区的子网，同一私有网络下不同可用区的子网默认可以内网互通',
            content: () => (
              <ZoneSelector
                v-model={formData.subnet.zone}
                vendor={cond.vendor}
                region={cond.region}
              />
            ),
          },
          {
            label: '子网IPv6网段',
            display: cond.vendor === VendorEnum.HUAWEI,
            content: () => (
              <Checkbox v-model={formData.subnet.ipv6_enable}>
                开启IPv6
              </Checkbox>
            ),
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
              <span>
                默认防火墙规则是可以出，不允许进入。如需绑定防火墙规则，请在创建VPC后，进入VPC管理页面绑定。
              </span>
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
      ipv4Cidr1: [
        {
          trigger: 'change',
          message: 'IPv4 CIDR 不能为空',
          validator: () => {
            return (
              !!formData.ipv4_cidr[4]
            );
          },
        },
      ],
      ipv4Cidr2: [
        {
          trigger: 'change',
          message: 'IPv4 CIDR 前缀长度不能为空',
          validator: () => {
            return (
              !!formData.subnet.ipv4_cidr[4]
            );
          },
        },
        {
          trigger: 'change',
          message: '子网IPv4 CIDR不合法，或不在VPC的CIDR范围中',
          validator: () => {
            if (cond.vendor !== VendorEnum.GCP) {
              return (
                !formData.ipv4_cidr[4] || +formData.subnet.ipv4_cidr[4] >= +formData.ipv4_cidr[4]
              );
            }
            switch (+formData.subnet.ipv4_cidr[0]) {
              case 10:
                return +formData.subnet.ipv4_cidr[4] >= 8;
              case 172:
                return +formData.subnet.ipv4_cidr[4] >= 12;
              case 192:
                return +formData.subnet.ipv4_cidr[4] >= 16;
            }
          },
        },
      ],
    };

    watch(
      () => formData.ipv4_cidr[0],
      (val) => {
        ipv4CidrFir.value = val;
        const maskrang: any = cond.vendor === 'tcloud' ? TCLOUDCIDRMASKRANGE : CIDRMASKRANGE;
        if (val === '192') {
          formData.ipv4_cidr[1] = '168';
          if (formData.ipv4_cidr[4] < maskrang[val].min) {
            formData.ipv4_cidr[4] = maskrang[val].min;
          }
        } else if (val === '172') {
          if (formData.ipv4_cidr[4] < maskrang[val].min) {
            formData.ipv4_cidr[4] = maskrang[val].min;
          }
        } else {
          if (formData.ipv4_cidr[1] > CIDRDATARANGE[val].max) {
            formData.ipv4_cidr[1] = CIDRDATARANGE[val].max;
          }
          if (formData.ipv4_cidr[1] < CIDRDATARANGE[val].min) {
            formData.ipv4_cidr[1] = CIDRDATARANGE[val].min;
          }
        }
        // eslint-disable-next-line prefer-destructuring
        formData.subnet.ipv4_cidr[0] = val;
        // eslint-disable-next-line prefer-destructuring
        formData.subnet.ipv4_cidr[1] = formData.ipv4_cidr[1];
      },
      { immediate: true },
    );

    watch(
      () => formData.ipv4_cidr[1],
      () => {
        // eslint-disable-next-line prefer-destructuring
        formData.subnet.ipv4_cidr[1] = formData.ipv4_cidr[1];
      },
    );

    return () => (
      <div>
        <DetailHeader>
          <p class={'purchase-vpc-header-title'}>购买VPC</p>
        </DetailHeader>
        <div class="create-form-container" style={isResourcePage && { padding: 0 }}>
          <Form
            model={formData}
            rules={formRules}
            ref={formRef}
            onSubmit={handleFormSubmit}
            formType='vertical'>
            <ConditionOptions
              type={ResourceTypeEnum.VPC}
              v-model:bizId={cond.bizId}
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
                    .map(({
                      label,
                      description,
                      tips,
                      required,
                      property,
                      content,
                    }) => (
                        <FormItem
                          label={label}
                          required={required}
                          property={property}
                          description={description}>
                          {Array.isArray(content) ? (
                            <div class='flex-row'>
                              {content
                                .filter(sub => sub.display !== false)
                                .map(sub => (
                                  <FormItem
                                    label={sub.label}
                                    required={sub.required}
                                    property={sub.property}
                                    description={sub?.description}>
                                    {sub.content()}
                                    {sub.tips && (
                                      <div class='form-item-tips'>
                                        {sub.tips()}
                                      </div>
                                    )}
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
