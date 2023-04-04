import { computed, defineComponent } from 'vue';
import ContentContainer from '../components/common/content-container.vue';
import ConditionOptions from '../components/common/condition-options.vue';
import FormGroup from '../components/common/form-group.vue';
import CloudAreaSelector from '../components/common/cloud-area-selector';
import ZoneSelector from '../components/common/zone-selector';
import { Form, Input, Checkbox, Button, Radio  } from 'bkui-vue';
import { Info } from 'bkui-vue/lib/icon';

import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import useVpcOptions from '../hooks/use-vpc-options';
import useCondtion from '../hooks/use-condtion';
import useVpcFormData from '../hooks/use-vpc-form-data';

const { FormItem, ComposeFormItem } = Form;
const { Group: RadioGroup } = Radio;

export default defineComponent({
  props: {},
  setup(props, ctx) {
    const { cond, isEmptyCond } = useCondtion(ResourceTypeEnum.VPC);
    const { formData, formRef, handleFormSubmit, submitting } = useVpcFormData(cond);
    const __ = useVpcOptions(cond, formData);

    const submitDisabled = computed(() => isEmptyCond.value);

    const nameReg = /^[\w\u4e00-\u9fa5-.]{0,59}$/;
    const nameRegMsg = '不超过60个字符，允许字母、数字、中文字符，\'-\'、\'_\'、\'.\'';

    const networkTips = computed(() => {
      const map = {
        [VendorEnum.TCLOUD]: 'CIDR范围的有效范围为:<br />10.0.0.0 - 10.255.255.255（掩码范围需在12 - 28之间）<br />172.16.0.0 - 172.31.255.255（掩码范围需在12 - 28之间）<br />192.168.0.0 - 192.168.255.255 （掩码范围需在16 - 28之间）<br />更多信息请参考官方说明https://cloud.tencent.com/document/product/215/36515',
        [VendorEnum.AWS]: 'CIDR范围的有效范围为:<br />10.0.0.0 - 10.255.255.255（10/8 前缀）<br />172.16.0.0 - 172.31.255.255（172.16/12 前缀）<br />192.168.0.0 - 192.168.255.255（192.168/16 前缀）<br />更多信息请参考官方说明https://docs.aws.amazon.com/zh_cn/vpc/latest/userguide/configure-your-vpc.html#add-cidr-block-restrictions',
        [VendorEnum.AZURE]: 'CIDR范围的有效范围为:<br />10.0.0.0 - 10.255.255.255（10/8 前缀）<br />172.16.0.0 - 172.31.255.255（172.16/12 前缀）<br />192.168.0.0 - 192.168.255.255（192.168/16 前缀）<br />更多信息请参考官方说明https://learn.microsoft.com/zh-cn/azure/virtual-network/virtual-networks-faq#what-address-ranges-can-i-use-in-my-vnets',
        [VendorEnum.GCP]: 'CIDR范围的有效范围为:<br />10.0.0.0/8<br />172.16.0.0/12<br />192.168.0.0/16<br />更多信息请参考官方说明https://cloud.google.com/vpc/docs/subnets?hl=zh-cn',
        [VendorEnum.HUAWEI]: 'CIDR范围的有效范围为:<br />10.0.0.0/8~24<br />172.16.0.0/12~24<br />192.168.0.0/16~24<br />更多信息请参考官方说明https://support.huaweicloud.com/intl/zh-cn/usermanual-vpc/zh-cn_topic_0013935842.html'
      }
      return map[cond.vendor]
    })

    const formConfig = computed(() => [
      {
        id: 'type',
        title: 'VPC类型',
        display: cond.vendor === VendorEnum.AWS,
        children: [
          {
            label: '类型',
            content: () => <RadioGroup v-model={formData.type}>
              <Radio label={0}>基本配置</Radio>
            </RadioGroup>,
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
            content: () => <RadioGroup v-model={formData.ip_source_type}>
              <Radio label={0}>业务私有</Radio>
              <Radio label={1} disabled={true}>IP池</Radio>
            </RadioGroup>,
          },
          {
            label: 'IPv4 CIDR',
            display: cond.vendor !== VendorEnum.GCP,
            required: true,
            property: 'ipv4_cidr',
            content: () => <>
              <ComposeFormItem class="mr5">
                <Input type='number' placeholder='1-255' min={1} max={255} v-model={formData.ipv4_cidr[0]} class="w110" suffix="." />
                <Input type='number' placeholder='0-255' min={0} max={255} v-model={formData.ipv4_cidr[1]} class="w110" suffix="." />
                <Input type='number' placeholder='0-255' min={0} max={255} v-model={formData.ipv4_cidr[2]} class="w110" suffix="." />
                <Input type='number' placeholder='0-255' min={0} max={255} v-model={formData.ipv4_cidr[3]} class="w110" suffix="/" />
                <Input type='number' placeholder='1-32' min={1} max={32} v-model={formData.ipv4_cidr[4]} class="w110" />
              </ComposeFormItem>
              <Info v-BkTooltips={{ content: networkTips.value }}></Info>
            </>,
          },
          {
            label: '所属的蓝鲸云区域',
            required: true,
            property: 'bk_cloud_id',
            content: () => <CloudAreaSelector v-model={formData.bk_cloud_id} />,
          },
          {
            label: 'BastionHost',
            display: cond.vendor === VendorEnum.AZURE,
            property: 'bastion_host_enable',
            content: () => <RadioGroup v-model={formData.bastion_host_enable}>
              <Radio label={false}>禁用</Radio>
              <Radio label={true} disabled={true}>暂不支持启用</Radio>
            </RadioGroup>,
          },
          {
            label: 'DDoS 保护标准',
            display: cond.vendor === VendorEnum.AZURE,
            property: 'ddos_enable',
            content: () => <RadioGroup v-model={formData.ddos_enable}>
              <Radio label={false}>禁用</Radio>
              <Radio label={true} disabled={true}>暂不支持启用</Radio>
            </RadioGroup>,
          },
          {
            label: '防火墙',
            display: cond.vendor === VendorEnum.AZURE,
            property: 'firewall_enable',
            content: () => <RadioGroup v-model={formData.firewall_enable}>
              <Radio label={false}>禁用</Radio>
              <Radio label={true} disabled={true}>暂不支持启用</Radio>
            </RadioGroup>,
          },
          {
            label: '租期',
            display: cond.vendor === VendorEnum.AWS,
            required: true,
            property: 'instance_tenancy',
            content: () => <RadioGroup v-model={formData.instance_tenancy}>
              <Radio label={'default'}>默认</Radio>
              <Radio label={'dedicated'}>专用</Radio>
            </RadioGroup>,
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
            content: () => <RadioGroup v-model={formData.routing_mode}>
              <Radio label={'REGIONAL'}>区域</Radio>
              <Radio label={'GLOBAL'}>全局</Radio>
            </RadioGroup>,
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
            required: true,
            content: () => <>
              <ComposeFormItem class="mr5">
                <Input type='number' placeholder='1-255' min={1} max={255} v-model={formData.subnet.ipv4_cidr[0]} class="w110" suffix="." />
                <Input type='number' placeholder='0-255' min={0} max={255} v-model={formData.subnet.ipv4_cidr[1]} class="w110" suffix="." />
                <Input type='number' placeholder='0-255' min={0} max={255} v-model={formData.subnet.ipv4_cidr[2]} class="w110" suffix="." />
                <Input type='number' placeholder='0-255' min={0} max={255} v-model={formData.subnet.ipv4_cidr[3]} class="w110" suffix="/" />
                <Input type='number' placeholder='1-32' min={1} max={32} v-model={formData.subnet.ipv4_cidr[4]} class="w110" />
              </ComposeFormItem>
              <Info v-BkTooltips={{ content: '请确保所填写的子网CIDR在VPC CIDR中' }}></Info>
            </>,
          },
          {
            label: '可用区',
            display: [VendorEnum.TCLOUD].includes(cond.vendor),
            required: true,
            property: 'subnet.zone',
            description: '同一私有网络下可以有不同可用区的子网，同一私有网络下不同可用区的子网默认可以内网互通',
            content: () => <ZoneSelector
              v-model={formData.subnet.zone}
              vendor={cond.vendor}
              region={cond.region} />,
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
            content: () => <RadioGroup v-model={formData.subnet.private_ip_google_access}>
              <Radio label={false}>禁用</Radio>
              <Radio label={true}>启用</Radio>
            </RadioGroup>,
          },
          {
            label: '流日志',
            display: cond.vendor === VendorEnum.GCP,
            content: () => <RadioGroup v-model={formData.subnet.enable_flow_logs}>
              <Radio label={false}>禁用</Radio>
              <Radio label={true}>启用</Radio>
            </RadioGroup>,
          },
          {
            label: '防火墙规则',
            display: cond.vendor === VendorEnum.GCP,
            content: () => <span>默认防火墙规则是可以出，不允许进入。如需绑定防火墙规则，请在创建VPC后，进入VPC管理页面绑定。</span>,
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
          trigger: 'input',
        },
      ],
    };

    return () => <ContentContainer>
      <ConditionOptions
        type={ResourceTypeEnum.VPC}
        v-model:bizId={cond.bizId}
        v-model:cloudAccountId={cond.cloudAccountId}
        v-model:vendor={cond.vendor}
        v-model:region={cond.region}
        v-model:resourceGroup={cond.resourceGroup}
      />
      <Form model={formData} rules={formRules} ref={formRef} onSubmit={handleFormSubmit}>
        {
          formConfig.value
            .filter(({ display }) => display !== false)
            .map(({ title, children }) => (
              <FormGroup title={title}>
                {
                  children
                    .filter(({ display }) => display !== false)
                    .map(({ label, description, tips, required, property, content }) => (
                    <FormItem
                      label={label}
                      required={required}
                      property={property}
                      description={description}
                    >
                      {
                        Array.isArray(content)
                          ? <div class="flex-row">
                            {
                              content
                                .filter(sub => sub.display !== false)
                                .map(sub => (
                                  <FormItem
                                    label={sub.label}
                                    required={sub.required}
                                    property={sub.property}
                                    description={sub?.description}
                                  >
                                    {sub.content()}
                                    { sub.tips && <div class="form-item-tips">{sub.tips()}</div> }
                                  </FormItem>
                                ))
                            }
                          </div>
                          : content()
                      }
                      { tips && <div class="form-item-tips">{tips()}</div> }
                    </FormItem>
                    ))
                }
              </FormGroup>
            ))
        }
        <div class="action-bar">
          <Button theme='primary' loading={submitting.value} disabled={submitDisabled.value} onClick={handleFormSubmit}>提交审批</Button>
          <Button>取消</Button>
        </div>
      </Form>
    </ContentContainer>;
  },
});
