import { computed, defineComponent, ref, watch } from 'vue';
// import components
import { Button, Form, Input, Select, Slider } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { EditLine, Plus } from 'bkui-vue/lib/icon';
import ZoneSelector from '@/components/zone-selector/index.vue';
import PrimaryStandZoneSelector from '../../components/common/PrimaryStandZoneSelector';
import RegionVpcSelector from '../../components/common/RegionVpcSelector';
import SubnetSelector from '../../components/common/subnet-selector';
import InputNumber from '@/components/input-number';
import ConditionOptions from '../../components/common/condition-options.vue';
import CommonCard from '@/components/CommonCard';
import VpcReviewPopover from '../../components/common/VpcReviewPopover';
import SelectedItemPreviewComp from '@/components/SelectedItemPreviewComp';
import BandwidthPackageSelector from '../../components/common/BandwidthPackageSelector';
// import types
import { type ISubnetItem } from '../../cvm/children/SubnetPreviewDialog';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
// import constants
import { CLB_SPECS, LB_ISP, ResourceTypeEnum } from '@/common/constant';
import { LOAD_BALANCER_TYPE, ADDRESS_IP_VERSION, ZONE_TYPE, INTERNET_CHARGE_TYPE, ISP_TYPES } from '@/constants/clb';
// import utils
import bus from '@/common/bus';
import { useI18n } from 'vue-i18n';
import { reqAccountNetworkType } from '@/api/load_balancers/apply-clb';
// import custom hooks
import useFilterResource from './useFilterResource';
import { CLB_QUOTA_NAME } from '@/typings';
import { useBusinessStore, useResourceStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const { Option } = Select;
const { FormItem } = Form;

// apply-clb, 渲染表单
export default (formModel: ApplyClbModel) => {
  // use hooks
  const { t } = useI18n();
  const { isBusinessPage } = useWhereAmI();
  const resourceStore = useResourceStore();
  const businessStore = useBusinessStore();

  // define data
  const vpcId = ref('');
  const vpcData = ref(null); // 预览vpc
  const subnetData = ref(null); // 预览子网
  const isSubnetPreviewDialogShow = ref(false);
  const formRef = ref();
  // define computed properties
  const isIntranet = computed(() => formModel.load_balancer_type === 'INTERNAL');

  // define handler function
  const handleVpcChange = async (vpc: any) => {
    if (vpc) {
      // 获取 vpc 详情用于预览
      const detailApi = isBusinessPage ? businessStore.detail : resourceStore.detail;
      detailApi('vpcs', vpc.id).then(({ data }: any) => (vpcData.value = data));
      if (vpcId.value !== vpc.id) {
        vpcId.value = vpc.id;
        formModel.cloud_subnet_id = undefined;
      }
    } else {
      vpcData.value = null;
    }

    if (!vpc) return;
  };
  const handleSubnetDataChange = (data: ISubnetItem) => {
    subnetData.value = data;
  };

  // use custom hooks
  const { ispList, isResourceListLoading, quotas, isInquiryPricesLoading } = useFilterResource(formModel);

  // 当前地域下负载均衡的配额
  const currentLbQuota = computed(() => {
    const quotaName =
      formModel.load_balancer_type === 'OPEN'
        ? CLB_QUOTA_NAME.TOTAL_OPEN_CLB_QUOTA
        : CLB_QUOTA_NAME.TOTAL_INTERNAL_CLB_QUOTA;
    return quotas.value.find(({ quota_id }) => quotaName === quota_id);
  });
  // 购买数量的最大值
  const requireCountMax = computed(() => currentLbQuota.value?.quota_limit - currentLbQuota.value?.quota_current || 1);
  // 配额余量
  const quotaRemaining = computed(() =>
    currentLbQuota.value?.quota_limit ? requireCountMax.value - formModel.require_count : 0,
  );

  const rules = {
    name: [
      {
        validator: (value: string) => /^[a-zA-Z0-9]([-a-zA-Z0-9]{0,58})[a-zA-Z0-9]$/.test(value),
        message: '60个字符，字母、数字、“-”，且必须以字母、数字开头和结尾。',
        trigger: 'change',
      },
    ],
  };

  // change-handle - 更新 sla_type
  const handleSlaTypeChange = (v: '0' | '1') => {
    if (v === '0') formModel.sla_type = 'shared';
  };

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
            description: '公网：面向公网使用的负载均衡。\n内网：面向内网使用的负载均衡。',
            content: () => (
              <BkRadioGroup v-model={formModel.load_balancer_type}>
                {LOAD_BALANCER_TYPE.map(({ label, value }) => (
                  <BkRadioButton label={value} class='w110'>
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
            description: '支持IPv4, IPv6, 以及IPv6 NAT64（负载均衡通过IPv6地址，将用户请求转发给后端IPv4地址的服务器）',
            hidden: isIntranet.value,
            content: () => (
              <BkRadioGroup v-model={formModel.address_ip_version}>
                {ADDRESS_IP_VERSION.map(({ label, value, isDisabled }) => {
                  const disabled = typeof isDisabled === 'function' ? isDisabled(formModel.region) : false;
                  return (
                    <BkRadioButton
                      label={value}
                      class='w110'
                      disabled={disabled}
                      v-bk-tooltips={{
                        content: t('当前地域不支持IPv6 NAT64'),
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
        {
          label: 'VPC',
          required: true,
          property: 'cloud_vpc_id',
          content: () => (
            <div class='component-with-preview'>
              <RegionVpcSelector
                class='flex-1'
                v-model={formModel.cloud_vpc_id}
                accountId={formModel.account_id}
                vendor={formModel.vendor}
                region={formModel.region}
                onChange={handleVpcChange}
              />
              <VpcReviewPopover data={vpcData.value} />
            </div>
          ),
        },
        {
          label: '可用区',
          description:
            '单可用区：仅支持一个可用区。\n主备可用区：主可用区是当前承载流量的可用区。备可用区默认不承载流量，主可用区不可用时才使用备可用区。',
          hidden: !isIntranet.value && formModel.address_ip_version !== 'IPV4',
          content: () => (
            <div class='flex-row'>
              {!isIntranet.value && (
                <Select v-model={formModel.zoneType} clearable={false} filterable={false} class='w220'>
                  {ZONE_TYPE.map(({ label, value, isDisabled }) => {
                    const disabled =
                      typeof isDisabled === 'function' ? isDisabled(formModel.region, formModel.account_type) : false;
                    return (
                      <Option
                        id={value}
                        name={label}
                        disabled={disabled}
                        v-bk-tooltips={{
                          boundary: 'parent',
                          placement: 'right',
                          content:
                            formModel.account_type === 'LEGACY' ? (
                              <span>
                                {t('仅标准型账号支持主备可用区')}账号类型说明，参考{' '}
                                <a
                                  href='https://cloud.tencent.com/document/product/1199/49090#judge'
                                  target='_blank'
                                  style={{ color: '#3A84FF' }}>
                                  https://cloud.tencent.com/document/product/1199/49090#judge
                                </a>
                              </span>
                            ) : (
                              t('仅广州、上海、南京、北京、中国香港、首尔地域的 IPv4 版本的 CLB 支持主备可用区')
                            ),
                          disabled: !disabled,
                        }}>
                        {t(label)}
                      </Option>
                    );
                  })}
                </Select>
              )}
              {(function () {
                let zoneSelectorVNode = null;
                if (isIntranet.value || formModel.zoneType === '0') {
                  zoneSelectorVNode = (
                    <ZoneSelector
                      class='flex-1'
                      v-model={formModel.zones}
                      vendor={formModel.vendor}
                      region={formModel.region}
                      delayed={true}
                      isLoading={isResourceListLoading.value}
                    />
                  );
                } else {
                  zoneSelectorVNode = (
                    <PrimaryStandZoneSelector
                      class='flex-1'
                      v-model:zones={formModel.zones}
                      v-model:backupZones={formModel.backup_zones}
                      vendor={formModel.vendor}
                      region={formModel.region}
                    />
                  );
                }
                return zoneSelectorVNode;
              })()}
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
                class='flex-1'
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
          label: '运营商类型',
          required: true,
          property: 'vip_isp',
          hidden: isIntranet.value,
          content: () => {
            return (
              <BkRadioGroup v-model={formModel.vip_isp}>
                {ISP_TYPES.map((item) => (
                  <BkRadioButton
                    key={item}
                    label={item}
                    class='w110'
                    disabled={!ispList.value?.includes(item)}
                    v-bk-tooltips={(function () {
                      if (!formModel.zones) {
                        return { content: '请选择可用区', disabled: formModel.zones };
                      }
                      return { content: '当前地域不支持', disabled: ispList.value?.includes(item) };
                    })()}>
                    {LB_ISP[item]}
                  </BkRadioButton>
                ))}
              </BkRadioGroup>
            );
          },
        },
        [
          {
            label: '负载均衡规格类型',
            required: true,
            property: 'slaType',
            description:
              '共享型实例：按照规格提供性能保障，单实例最大支持并发连接数5万、每秒新建连接数5000、每秒查询数（QPS）5000。\n性能容量型实例：按照规格提供性能保障，单实例最大可支持并发连接数1000万、每秒新建连接数100万、每秒查询数（QPS）30万。',
            hidden: isIntranet.value,
            content: () => (
              <Select
                v-model={formModel.slaType}
                filterable={false}
                clearable={false}
                class='w220'
                onChange={handleSlaTypeChange}>
                <Option id='0' name={t('共享型')} />
                <Option
                  id='1'
                  name={t('性能容量型')}
                  disabled={!formModel.vip_isp}
                  v-bk-tooltips={{ content: '请选择运营商类型', disabled: !!formModel.vip_isp, boundary: 'parent' }}
                />
              </Select>
            ),
          },
          {
            label: '实例规格',
            required: true,
            property: 'sla_type',
            hidden: formModel.slaType !== '1',
            content: () => {
              if (formModel.sla_type !== 'shared') {
                return (
                  <SelectedItemPreviewComp
                    content={CLB_SPECS[formModel.sla_type]}
                    onClick={() => bus.$emit('showLbSpecTypeSelectDialog')}
                  />
                );
              }
              return (
                <Button
                  onClick={() => bus.$emit('showLbSpecTypeSelectDialog')}
                  disabled={!formModel.vip_isp}
                  v-bk-tooltips={{ content: '请选择运营商类型', disabled: !!formModel.vip_isp }}>
                  <Plus class='f24' />
                  {t('选择实例规格')}
                </Button>
              );
            },
          },
        ],
        {
          label: '弹性公网 IP',
          // 弹性IP，仅内网可绑定。公网类型无法指定IP。绑定弹性IP后，内网CLB当做公网CLB使用
          hidden: !isIntranet.value,
          content: () => {
            if (formModel.cloud_eip_id) {
              return (
                <div style=''>
                  <div class={'image-selector-selected-block-container'}>
                    <div class={'selected-block mr8'}>{formModel.cloud_eip_id} </div>
                    <EditLine
                      fill='#3A84FF'
                      width={13.5}
                      height={13.5}
                      onClick={() => bus.$emit('showBindEipDialog')}
                    />
                  </div>
                </div>
              );
            }
            return (
              <Button theme='primary' onClick={() => bus.$emit('showBindEipDialog')}>
                <Plus class='f24' />
                {t('绑定弹性 IP')}
              </Button>
            );
          },
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
            <BkRadioGroup
              v-model={formModel.internet_charge_type}
              onChange={(val) => {
                if (val !== 'BANDWIDTH_PACKAGE') formModel.bandwidth_package_id = undefined;
              }}>
              {INTERNET_CHARGE_TYPE.map(({ label, value }) => (
                <BkRadioButton
                  key={value}
                  label={value}
                  class='w88'
                  disabled={!value}
                  v-bk-tooltips={{
                    content: '云平台当前API接口暂不支持包月参数',
                    disabled: value,
                  }}>
                  {t(label)}
                </BkRadioButton>
              ))}
            </BkRadioGroup>
          ),
        },
        {
          label: '共享带宽包',
          required: true,
          property: 'bandwidth_package_id',
          hidden: formModel.internet_charge_type !== 'BANDWIDTH_PACKAGE',
          content: () => (
            <BandwidthPackageSelector
              v-model={formModel.bandwidth_package_id}
              accountId={formModel.account_id}
              region={formModel.region}
            />
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
                minValue={1}
                maxValue={5120}
                customContent={{
                  1: { label: '1' },
                  256: { label: '256' },
                  512: { label: '512' },
                  1024: { label: '1024' },
                  2048: { label: '2048' },
                  5120: { label: '5120' },
                }}
                showInput
                labelClick>
                {{
                  end: () => <div class='slider-unit-suffix'>Mbps</div>,
                }}
              </Slider>
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
                <InputNumber v-model={formModel.require_count} min={1} max={requireCountMax.value} />
                <div class='quota-info'>
                  {t('所在地域配额为')}
                  <span class='quota-number ml5'>{quotaRemaining.value}</span>
                  <span class='ml5 mr5'>/</span>
                  {currentLbQuota.value?.quota_limit || 0}
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
          description: '单个实例：以填写的名称命名。\n多个实例：以填写的名称为前缀，由系统自动补充随机的后缀。',
          content: () => <Input class='w500' v-model={formModel.name} placeholder='请输入实例名称'></Input>,
        },
        {
          label: '申请单备注',
          property: 'memo',
          content: () => (
            <Input
              type='textarea'
              v-model={formModel.memo}
              rows={3}
              maxlength={255}
              resize={false}
              placeholder='请输入申请单备注'></Input>
          ),
        },
      ],
    },
  ]);

  // define component
  const ApplyClbForm = defineComponent({
    setup() {
      return () => (
        <Form class='apply-clb-form-container' formType='vertical' model={formModel} ref={formRef} rules={rules}>
          <ConditionOptions
            type={ResourceTypeEnum.CLB}
            bizs={formModel.bk_biz_id}
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

  return {
    vpcData,
    subnetData,
    isSubnetPreviewDialogShow,
    ApplyClbForm,
    formRef,
    isInquiryPricesLoading,
  };
};
