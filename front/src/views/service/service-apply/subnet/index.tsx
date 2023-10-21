import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import RouteTableSelector from '@/components/route-table-selector/index.vue';
import ConditionOptions from '../components/common/condition-options.vue';
import VpcSelector from '@/components/vpc-selector/index.vue';
import ZoneSelector from '@/components/zone-selector/index.vue';
import { defineComponent, reactive, ref, watch } from 'vue';
import { Button, Card, Form, Input, Select, Message } from 'bkui-vue';
import './index.scss';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import { useAccountStore, useBusinessStore, useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useRouter } from 'vue-router';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  props: {},
  setup() {
    const formModel = reactive({
      biz_id: '' as string,
      account_id: '' as string, // 云账号
      vendor: null as VendorEnum, // 云厂商
      region: '' as string, // 云地域
      zone: '' as string, // 可用区
      cloud_vpc_id: 0 as number, // 所属的VPC网络
      name: '' as string, // 子网名称
      ipv4_cidr: '' as string, // IPV4 CIDR
      cloud_route_table_id: 0 as number, // 关联的路由表
    });
    const cidr_host = ref('');
    const cidr_mask = ref('');
    const submitLoading = ref(false);
    const { whereAmI } = useWhereAmI();
    const router = useRouter();

    const resourceStore = useResourceStore();
    const businessStore = useBusinessStore();
    const accountStore = useAccountStore();

    // const handleChange = (data: BusinessFormFilter) => {
    //   formModel.account_id = data.account_id as string;
    //   formModel.vendor = data.vendor as VendorEnum;
    //   formModel.region = data.region as string;
    // };

    const getVpcDetail = async (vpcId: string) => {
      console.log('vpcId', vpcId);
      if (!vpcId) return;
      await resourceStore.detail('vpcs', vpcId);
    };

    const handleSubmit = async () => {
      submitLoading.value = true;
      try {
        await businessStore.createSubnet(
          accountStore.bizs,
          formModel,
          whereAmI.value === Senarios.resource,
        );
        Message({
          theme: 'success',
          message: '创建成功',
        });
        handleCancel();
      } catch (error) {
        console.log(error);
      } finally {
        submitLoading.value = false;
      }
    };

    const handleCancel = () => {
      router.back();
    };

    watch(
      () => [cidr_host.value, cidr_mask.value],
      (vals) => {
        const [host, mask] = vals;
        formModel.ipv4_cidr = `10.10.10.${host}/${mask}`;
        console.log(host, mask, formModel.ipv4_cidr);
      },
    );

    return () => (
      <div>
        <DetailHeader>
          <span class={'subnet-title'}>新建子网</span>
        </DetailHeader>

        <div class={'create-subnet-form-contianer'}>
          <ConditionOptions
            type={ResourceTypeEnum.CVM}
            v-model:bizId={formModel.biz_id}
            v-model:cloudAccountId={formModel.account_id}
            v-model:vendor={formModel.vendor}
            v-model:region={formModel.region}
            // v-model:resourceGroup={formModel.resourceGroup}
            class={'mb16 mt24'}>
            {{
              default: () => (
                <Form formType='vertical'>
                  <FormItem label={'可用区'}>
                    <ZoneSelector
                      v-model={formModel.zone}
                      vendor={formModel.vendor}
                      region={formModel.region}
                    />
                  </FormItem>
                </Form>
              ),
            }}
          </ConditionOptions>
          <Card showHeader={false} class={'subnet-basic-info'}>
            <p class={'info-title'}>子网信息</p>
            <Form formType='vertical' model={formModel} class={'ml30'}>
              <FormItem
                label='所属VPC网络'
                style={{
                  width: '590px',
                }}>
                <VpcSelector
                  vendor={formModel.vendor}
                  region={formModel.region}
                  v-model={formModel.cloud_vpc_id}
                  onHandleVpcDetail={getVpcDetail}
                />
              </FormItem>
              <FormItem
                label='子网名称'
                style={{
                  width: '880px',
                }}>
                <Input
                  maxlength={60}
                  placeholder="不超过60个字符，允许字母、数字、中文字符，'-'、'_'、'.'"
                  v-model={formModel.name}></Input>
              </FormItem>
              <FormItem label='IPv4 CIDR'>
                <div class={'cidr-selector-container'}>
                  10.10.10.
                  <Input
                    class={'cidr-selector'}
                    placeholder='16'
                    v-model={cidr_host.value}
                  />
                  <p>/</p>
                  <Select
                    class={'cidr-selector'}
                    placeholder='28'
                    v-model={cidr_mask.value}>
                    {new Array(32)
                      .fill(0)
                      .map((_, idx) => idx + 1)
                      .map(num => (
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
                <RouteTableSelector
                  cloud-vpc-id={formModel.cloud_vpc_id}
                  v-model={formModel.cloud_route_table_id}
                />
              </FormItem>
            </Form>
          </Card>
          <div class={'button-group'}>
            <Button
              theme={'primary'}
              class={'button-submit'}
              onClick={handleSubmit}
              loading={submitLoading.value}>
              提交
            </Button>
            <Button
              class={'button-cancel'}
              loading={submitLoading.value}
              onClick={handleCancel}>
              取消
            </Button>
          </div>
        </div>
      </div>
    );
  },
});
