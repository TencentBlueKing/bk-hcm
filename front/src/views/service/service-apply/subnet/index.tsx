import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import RouteTableSelector from '@/components/route-table-selector/index.vue';
import ConditionOptions from '../components/common/condition-options/index.vue';
import VpcSelector from '@/views/service/service-apply/components/common/vpc-selector';
import ZoneSelector from '@/components/zone-selector/index.vue';
import { defineComponent, reactive, ref, watch } from 'vue';
import { Button, Form, Input, Select, Message } from 'bkui-vue';
import './index.scss';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import { useAccountStore, useBusinessStore, useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useRouter } from 'vue-router';
import CommonCard from '@/components/CommonCard';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  props: {},
  setup() {
    const { getBizsId, whereAmI, isResourcePage } = useWhereAmI();
    const formModel = reactive({
      biz_id: whereAmI.value === Senarios.business ? getBizsId() : 0,
      account_id: '' as string, // 云账号
      vendor: null as VendorEnum, // 云厂商
      resource_group: '' as string, // 资源组
      region: '' as string, // 云地域
      zone: '' as string, // 可用区
      cloud_vpc_id: '', // 所属的VPC网络
      name: '' as string, // 子网名称
      ipv4_cidr: '' as string | string[], // IPV4 CIDR
      cloud_route_table_id: '', // 关联的路由表,
      gateway_ip: '' as string, // 网关地址
    });
    const formRef = ref(null);
    const formRules = {
      cloud_vpc_id: [
        {
          message: '请选择新子网所属VPC网络',
        },
      ],
      name: [
        {
          message: "名字应不超过60个字符，允许字母、数字、中文字符，'-'、'_'、'.'",
          trigger: 'blur',
          validator: (value: string) => /^[\u4e00-\u9fa5\w.-]{1,60}$/.test(value),
        },
      ],
      ipv4_cidr: [
        {
          message: '请正确填写IPv4 CIDR',
          validator: (value: any) => {
            const [, , cidr_host1, cidr_host2, cidr_mask] = value.split(/[./]/);
            if (
              isNaN(cidr_host1) ||
              isNaN(cidr_host2) ||
              cidr_host1 < 0 ||
              cidr_host2 < 0 ||
              cidr_mask < subIpv4cidr.value[2] ||
              cidr_mask > 31
            )
              return false;
            return true;
          },
        },
      ],
    };
    const cidr_host1 = ref('');
    const cidr_host2 = ref('');
    const cidr_mask = ref('');
    const submitLoading = ref(false);
    const router = useRouter();

    const resourceStore = useResourceStore();
    const businessStore = useBusinessStore();
    const accountStore = useAccountStore();

    const subIpv4cidr = ref([10, 0, 28]);

    // const handleChange = (data: BusinessFormFilter) => {
    //   formModel.account_id = data.account_id as string;
    //   formModel.vendor = data.vendor as VendorEnum;
    //   formModel.region = data.region as string;
    // };

    const getVpcDetail = async (vpc: { id: string }) => {
      const vpcId = vpc.id;
      if (!vpcId) return;
      const res = await resourceStore.detail('vpcs', vpcId);
      const arr = res.data?.extension?.cidr || [];
      const idx = arr.findIndex(({ type }: { type: string }) => type === 'ipv4');
      if (idx !== -1) {
        const [ip, mask] = arr[idx].cidr.split('/');
        const ipArr = ip.split('.');
        subIpv4cidr.value = [ipArr[0], ipArr[1], mask];
      }
    };

    const handleSubmit = async () => {
      await formRef.value.validate();
      if (formModel.vendor === VendorEnum.AZURE) {
        formModel.ipv4_cidr = [formModel.ipv4_cidr] as string[];
      }
      submitLoading.value = true;
      try {
        await businessStore.createSubnet(accountStore.bizs, formModel, whereAmI.value === Senarios.resource);
        Message({
          theme: 'success',
          message: '创建成功',
        });
        handleCancel();
      } catch (error) {
        console.error(error);
      } finally {
        submitLoading.value = false;
      }
    };

    const handleCancel = () => {
      if (window.history.state.back) {
        router.back();
      } else {
        router.replace({
          path: '/resource/resource',
          query: {
            type: 'subnet',
          },
        });
      }
    };

    watch(
      () => [subIpv4cidr.value, cidr_host1.value, cidr_host2.value, cidr_mask.value],
      () => {
        formModel.ipv4_cidr = `${subIpv4cidr.value[0]}.${subIpv4cidr.value[1]}.${cidr_host1.value}.${cidr_host2.value}/${cidr_mask.value}`;
      },
      {
        deep: true,
      },
    );

    // 当云账号、云地域、可用区变化时, 清空表单
    watch(
      () => [formModel.account_id, formModel.region, formModel.zone],
      () => {
        Object.assign(formModel, {
          cloud_vpc_id: '', // 所属的VPC网络
          name: '', // 子网名称
          cloud_route_table_id: '', // 关联的路由表,
          gateway_ip: '', // 网关地址
        });
        cidr_host1.value = '';
        cidr_host2.value = '';
        cidr_mask.value = '';
      },
    );

    return () => (
      <div>
        <DetailHeader>
          <span class={'subnet-title'}>新建子网</span>
        </DetailHeader>
        <div
          class='create-form-container subnet-wrap'
          style={whereAmI.value === Senarios.resource && { padding: 0, marginBottom: '68px' }}>
          <Form formType='vertical' model={formModel} ref={formRef} rules={formRules}>
            <ConditionOptions
              type={ResourceTypeEnum.SUBNET}
              bizs={formModel.biz_id}
              v-model:cloudAccountId={formModel.account_id}
              v-model:vendor={formModel.vendor}
              v-model:region={formModel.region}
              v-model:resourceGroup={formModel.resource_group}>
              {{
                default: () => (
                  <FormItem label={'可用区'} required property='zone'>
                    <ZoneSelector v-model={formModel.zone} vendor={formModel.vendor} region={formModel.region} />
                  </FormItem>
                ),
              }}
            </ConditionOptions>
            <CommonCard title={() => '子网信息'}>
              <FormItem
                label='所属VPC网络'
                property='cloud_vpc_id'
                required
                style={{
                  width: '590px',
                }}>
                <VpcSelector
                  isSubnet={true}
                  zone={formModel.zone}
                  bizId={formModel.biz_id}
                  vendor={formModel.vendor}
                  region={formModel.region}
                  v-model={formModel.cloud_vpc_id}
                  accountId={formModel.account_id}
                  clearable={false}
                  onChange={getVpcDetail}
                />
              </FormItem>
              <FormItem
                label='子网名称'
                property='name'
                required
                style={{
                  width: '880px',
                }}>
                <Input
                  maxlength={60}
                  placeholder="不超过60个字符，允许字母、数字、中文字符，'-'、'_'、'.'"
                  v-model={formModel.name}></Input>
              </FormItem>
              <FormItem label='IPv4 CIDR' property='ipv4_cidr' required>
                <div class={'cidr-selector-container'}>
                  {`${subIpv4cidr.value[0]}.${subIpv4cidr.value[1]}.`}
                  <Input class={'cidr-selector'} placeholder='16' v-model={cidr_host1.value} />.
                  <Input class={'cidr-selector'} placeholder='16' v-model={cidr_host2.value} />
                  <p>/</p>
                  <Select class={'cidr-selector'} placeholder={`${subIpv4cidr.value[2]}-31`} v-model={cidr_mask.value}>
                    {new Array(31 - subIpv4cidr.value[2] + 1)
                      .fill(0)
                      .map((_, idx) => idx + +subIpv4cidr.value[2])
                      .map((num) => (
                        <Option key={num} label={num} value={num}>
                          {num}
                        </Option>
                      ))}
                  </Select>
                </div>
              </FormItem>
              <FormItem
                label='关联路由表'
                style={{
                  width: '590px',
                }}>
                <RouteTableSelector cloud-vpc-id={formModel.cloud_vpc_id} v-model={formModel.cloud_route_table_id} />
              </FormItem>
              {formModel.vendor === 'huawei' && (
                <FormItem
                  label='网关地址'
                  property='gateway_ip'
                  required
                  description={'子网的网关地址，默认建议填写子网中的第1个IP'}
                  style={{ width: '880px' }}>
                  <Input v-model={formModel.gateway_ip} placeholder='请输入网关地址'></Input>
                </FormItem>
              )}
            </CommonCard>
          </Form>
        </div>
        <div class={'button-group'} style={{ paddingLeft: isResourcePage && 'calc(15% + 24px)' }}>
          <Button theme={'primary'} class={'button-submit'} onClick={handleSubmit} loading={submitLoading.value}>
            提交
          </Button>
          <Button class={'button-cancel'} loading={submitLoading.value} onClick={handleCancel}>
            取消
          </Button>
        </div>
      </div>
    );
  },
});
