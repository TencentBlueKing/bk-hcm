import { computed, defineComponent, ref, watch } from 'vue';
// import components
import { Button, Form, Input, Select, Slider } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { Plus } from 'bkui-vue/lib/icon';
import ZoneSelector from '@/components/zone-selector/index.vue';
import PrimaryStandZoneSelector from '../../components/common/primary-stand-zone-selector';
import VpcSelector from '../../components/common/vpc-selector';
import RegionVpcSelector from '../../components/common/region-vpc-selector';
import SubnetSelector from '../../components/common/subnet-selector';
import InputNumber from '@/components/input-number';
import ConditionOptions from '../../components/common/condition-options.vue';
import CommonCard from '@/components/CommonCard';
// import types
import { type ISubnetItem } from '../../cvm/children/SubnetPreviewDialog';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
// import constants
import { ResourceTypeEnum } from '@/common/constant';
import { LOAD_BALANCER_TYPE, ADDRESS_IP_VERSION, ZONE_TYPE, INTERNET_CHARGE_TYPE } from '@/constants/clb';
// import utils
import bus from '@/common/bus';
import { useI18n } from 'vue-i18n';
import { reqAccountNetworkType } from '@/api/load_balancers/apply-clb';
// import custom hooks
import useFilterResource from './useFilterResource';

const { Option } = Select;
const { FormItem } = Form;

// apply-clb, 渲染表单
export default (formModel: ApplyClbModel) => {
  // use hooks
  const { t } = useI18n();

  // define data
  const vpcId = ref('');
  const vpcData = ref(null); // 预览vpc
  const isVpcPreviewDialogShow = ref(false);
  const subnetData = ref(null); // 预览子网
  const isSubnetPreviewDialogShow = ref(false);
  const formRef = ref();
  // define computed properties
  const isIntranet = computed(() => formModel.load_balancer_type === 'INTERNAL');

  // define handler function
  const handleZoneChange = () => {
    vpcId.value = '';
    formModel.cloud_vpc_id = '';
    formModel.cloud_subnet_id = undefined;
  };
  const handleVpcChange = (vpc: any) => {
    vpcData.value = vpc;
    if (!vpc) return;
    if (vpcId.value !== vpc.id) {
      vpcId.value = vpc.id;
      formModel.cloud_subnet_id = undefined;
    }
  };
  const handleSubnetDataChange = (data: ISubnetItem) => {
    subnetData.value = data;
  };

  // use custom hooks
  const { ispList } = useFilterResource(formModel);

  // form item options
  const formItemOptions = computed(() => [
    {
      id: 'config',
      title: '配置信息',
      children: [
        [
          {
            label: '网络类型',
            required: true,
            property: 'load_balancer_type',
            description: '如需绑定弹性公网IP, 请切换到内网网络类型',
            content: () => (
              <BkRadioGroup v-model={formModel.load_balancer_type}>
                {LOAD_BALANCER_TYPE.map(({ label, value }) => (
                  <BkRadioButton label={value} class='w120'>
                    {t(label)}
                  </BkRadioButton>
                ))}
              </BkRadioGroup>
            ),
          },
          {
            label: 'IP版本',
            required: true,
            property: 'address_ip_version',
            hidden: isIntranet.value,
            content: () => (
              <BkRadioGroup v-model={formModel.address_ip_version}>
                {ADDRESS_IP_VERSION.map(({ label, value, isDisabled }) => {
                  const disabled = typeof isDisabled === 'function' ? isDisabled(formModel.region) : false;
                  return (
                    <BkRadioButton
                      label={value}
                      class='w120'
                      disabled={disabled}
                      v-bk-tooltips={{
                        content: t('当前地域不支持IPV6 NAT64'),
                        disabled: !disabled,
                      }}>
                      {t(label)}
                    </BkRadioButton>
                  );
                })}
              </BkRadioGroup>
            ),
          },
        ],
        [
          {
            label: '可用区类型',
            required: true,
            property: 'zoneType',
            hidden: isIntranet.value || formModel.address_ip_version !== 'IPV4',
            content: () => (
              <BkRadioGroup v-model={formModel.zoneType}>
                {ZONE_TYPE.map(({ label, value, isDisabled }) => {
                  const disabled = typeof isDisabled === 'function' ? isDisabled(formModel.region) : false;
                  return (
                    <BkRadioButton
                      label={value}
                      class='w120'
                      disabled={disabled}
                      v-bk-tooltips={{
                        content: t('当前地域不支持主备可用区'),
                        disabled: !disabled,
                      }}>
                      {t(label)}
                    </BkRadioButton>
                  );
                })}
              </BkRadioGroup>
            ),
          },
          {
            label: '可用区',
            required: true,
            property: 'zones',
            hidden: !isIntranet.value && formModel.address_ip_version !== 'IPV4',
            content: () => {
              let zoneSelectorVNode = null;
              if (isIntranet.value || formModel.zoneType === 'single') {
                zoneSelectorVNode = (
                  <ZoneSelector
                    v-model={formModel.zones}
                    vendor={formModel.vendor}
                    region={formModel.region}
                    onChange={handleZoneChange}
                    delayed={true}
                  />
                );
              } else {
                zoneSelectorVNode = (
                  <PrimaryStandZoneSelector
                    v-model:zones={formModel.zones}
                    v-model:backupZones={formModel.backup_zones}
                    vendor={formModel.vendor}
                    region={formModel.region}
                    onResetVipIsp={() => (formModel.vip_isp = '')}
                  />
                );
              }
              return zoneSelectorVNode;
            },
          },
        ],
        {
          label: 'VPC',
          required: true,
          property: 'cloud_vpc_id',
          content: () => (
            <div class='component-with-preview'>
              {isIntranet.value ? (
                <VpcSelector
                  class='base'
                  v-model={formModel.cloud_vpc_id}
                  bizId={formModel.bk_biz_id}
                  accountId={formModel.account_id}
                  vendor={formModel.vendor}
                  region={formModel.region}
                  zone={formModel.zones}
                  onChange={handleVpcChange}
                />
              ) : (
                <RegionVpcSelector
                  class='base'
                  v-model={formModel.cloud_vpc_id}
                  accountId={formModel.account_id}
                  region={formModel.region}
                  onChange={handleVpcChange}
                />
              )}

              <Button
                class='preview-btn'
                text
                theme='primary'
                disabled={!formModel.cloud_vpc_id}
                onClick={() => (isVpcPreviewDialogShow.value = true)}>
                {t('预览')}
              </Button>
            </div>
          ),
        },
        {
          label: '子网',
          required: true,
          property: 'cloud_subnet_id',
          hidden: !isIntranet.value,
          content: () => (
            <div class='component-with-preview'>
              <SubnetSelector
                class='base'
                v-model={formModel.cloud_subnet_id}
                bizId={formModel.bk_biz_id}
                vpcId={vpcId.value}
                vendor={formModel.vendor}
                region={formModel.region}
                accountId={formModel.account_id}
                zone={formModel.zones}
                clearable={false}
                handleChange={handleSubnetDataChange}
              />
              <Button
                class='preview-btn'
                text
                theme='primary'
                disabled={!formModel.cloud_subnet_id}
                onClick={() => {
                  isSubnetPreviewDialogShow.value = true;
                }}>
                {t('预览')}
              </Button>
            </div>
          ),
        },
        {
          label: '负载均衡规格类型',
          required: true,
          property: 'sla_type',
          hidden: isIntranet.value,
          content: () => (
            <BkButtonGroup>
              <Button
                selected={formModel.sla_type === 'shared'}
                onClick={() => (formModel.sla_type = 'shared')}
                class='w120'>
                {t('共享型')}
              </Button>
              <Button
                selected={formModel.sla_type !== 'shared'}
                onClick={() => bus.$emit('showSelectClbSpecTypeDialog')}
                class='w120'>
                {t('性能容量型')}
              </Button>
            </BkButtonGroup>
          ),
        },
        {
          label: '运营商类型',
          required: true,
          property: 'vip_isp',
          hidden: isIntranet.value,
          description: '运营商类型选择范围由主可用区, 备可用区, IP版本决定',
          content: () => (
            <Select v-model={formModel.vip_isp}>
              {ispList.value?.map((item) => (
                <Option key={item} id={item}>
                  {item}
                </Option>
              ))}
            </Select>
          ),
        },
        {
          label: '弹性公网 IP',
          // 弹性IP，仅内网可绑定。公网类型无法指定IP。绑定弹性IP后，内网CLB当做公网CLB使用
          hidden: !isIntranet.value,
          content: () => (
            <Button onClick={() => bus.$emit('showBindEipDialog')} theme={formModel.cloud_eip_id ? 'primary' : null}>
              <Plus class='f24' />
              {t('绑定弹性 IP')}
            </Button>
          ),
        },
      ],
    },
    {
      id: 'applyInfo',
      title: '购买信息',
      children: [
        {
          label: '实例计费模式',
          simpleShow: true,
          content: () => (
            <div class='simple-show-container'>
              <span class='label'>{t('实例计费模式')}</span>:<span class='value'>{t('按量计费')}</span>
              <i
                v-bk-tooltips={{ content: t('本期只支持按量计费'), placement: 'right' }}
                class='hcm-icon bkhcm-icon-prompt'></i>
            </div>
          ),
        },
        {
          label: '网络计费模式',
          required: true,
          property: 'internet_charge_type',
          hidden: (!isIntranet.value && formModel.account_type === 'LEGACY') || isIntranet.value,
          content: () => (
            <BkRadioGroup v-model={formModel.internet_charge_type}>
              {INTERNET_CHARGE_TYPE.map(({ label, value }) => (
                <BkRadioButton key={value} label={value} class='w88' disabled={!value}>
                  {t(label)}
                </BkRadioButton>
              ))}
            </BkRadioGroup>
          ),
        },
        {
          label: '带宽上限（Mbps）',
          required: true,
          property: 'internet_max_bandwidth_out',
          hidden: (!isIntranet.value && formModel.account_type === 'LEGACY') || isIntranet.value,
          content: () => (
            <div class='slider-wrap'>
              <Slider
                v-model={formModel.internet_max_bandwidth_out}
                maxValue={5120}
                customContent={{
                  0: { label: '0' },
                  256: { label: '256' },
                  512: { label: '512' },
                  1024: { label: '1024' },
                  2048: { label: '2048' },
                  5120: { label: '5120' },
                }}
                showInput
              />
              <div class='slider-unit-suffix'>Mbps</div>
            </div>
          ),
        },
        [
          {
            label: '购买数量',
            required: true,
            property: 'require_count',
            content: () => (
              <>
                <InputNumber v-model={formModel.require_count} min={1} />
                <div class='quota-info'>
                  {t('所在地域配额为')} <span class='quota-number'>{130}</span> / 500
                </div>
              </>
            ),
          },
          // {
          //   label: '购买时长',
          //   required: true,
          //   property: 'duration',
          //   content: () => (
          //     <div class='flex-row'>
          //       <Input
          //         v-model={formModel.duration}
          //         class='input-select-wrap'
          //         type='number'
          //         placeholder='0'
          //         min={1}
          //         max={unit.value === 'month' ? 11 : 5}>
          //         {{
          //           suffix: () => (
          //             <Select v-model={unit.value} clearable={false} class='input-suffix-select'>
          //               <Option label='月' value='month' />
          //               <Option label='年' value='year' />
          //             </Select>
          //           ),
          //         }}
          //       </Input>
          //       <Checkbox class='ml24' v-model={formModel.auto_renew}>
          //         自动续费
          //       </Checkbox>
          //     </div>
          //   ),
          // },
        ],
        {
          label: '实例名称',
          required: true,
          property: 'name',
          content: () => <Input class='w500' v-model={formModel.name}></Input>,
        },
        {
          label: '申请单备注',
          property: 'memo',
          content: () => (
            <Input type='textarea' v-model={formModel.memo} rows={3} maxlength={255} resize={false}></Input>
          ),
        },
      ],
    },
  ]);

  // define component
  const ApplyClbForm = defineComponent({
    setup() {
      return () => (
        <Form class='apply-clb-form-container' formType='vertical' model={formModel} ref={formRef}>
          <ConditionOptions
            type={ResourceTypeEnum.CLB}
            v-model:bizId={formModel.bk_biz_id}
            v-model:cloudAccountId={formModel.account_id}
            v-model:vendor={formModel.vendor}
            v-model:region={formModel.region}
          />
          {formItemOptions.value.map(({ id, title, children }) => (
            <CommonCard key={id} title={() => t(title)} class='form-card-container'>
              {children.map((item) => {
                let contentVNode = null;
                if (Array.isArray(item)) {
                  contentVNode = (
                    <div class='flex-row'>
                      {item.map(({ label, required, property, content, description, hidden }) => {
                        if (hidden) return null;
                        return (
                          <FormItem
                            key={property}
                            label={t(label)}
                            required={required}
                            property={property}
                            description={description}>
                            {content()}
                          </FormItem>
                        );
                      })}
                    </div>
                  );
                } else if (item.simpleShow) {
                  contentVNode = item.content();
                } else {
                  if (item.hidden) {
                    contentVNode = null;
                  } else {
                    contentVNode = (
                      <FormItem
                        key={item.property}
                        label={item.label}
                        required={item.required}
                        property={item.property}
                        description={item.description}>
                        {item.content()}
                      </FormItem>
                    );
                  }
                }
                return contentVNode;
              })}
            </CommonCard>
          ))}
        </Form>
      );
    },
  });

  watch(
    () => formModel.account_id,
    (val) => {
      // 当云账号变更时, 查询用户网络类型
      reqAccountNetworkType(val).then(({ data: { NetworkAccountType } }) => {
        formModel.account_type = NetworkAccountType;
      });
    },
  );

  return { vpcData, isVpcPreviewDialogShow, subnetData, isSubnetPreviewDialogShow, ApplyClbForm, formRef };
};
