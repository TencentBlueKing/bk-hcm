import { PropType, computed, defineComponent, onMounted, ref } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useStatus } from './useStatus';
import { Close, Spinner, Success } from 'bkui-vue/lib/icon';
import { Button, Exception, Form, Input, Message, Select } from 'bkui-vue';
import CommonDialog from '@/components/common-dialog';
import useBillStore from '@/store/useBillStore';
import useFormModel from '@/hooks/useFormModel';

const { FormItem } = Form;
const { Option } = Select;

export interface IDetail {
  id: string;
  source: string;
  sn: string;
  type: string;
  status: string;
  applicant: string;
  // content: {
  //   bk_biz_id: number; // 业务
  //   bak_managers: string[]; // 备份负责人
  //   op_product_id: number; // 运营产品
  //   id: string; // 一级账号ID
  //   vendor: string; // 云厂商
  //   dept_id: number;
  //   managers: string[]; // 主负责人
  // };
  content: string;
  delivery_detail: {
    complete: string;
  };
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  ticket_url: string;
}

export default defineComponent({
  props: {
    detail: {
      required: true,
      type: Object as PropType<IDetail>,
    },
  },
  setup(props) {
    const businessMapStore = useBusinessMapStore();
    const info = computed(() => JSON.parse(props.detail.content));
    const statusMap = useStatus(props.detail.delivery_detail);
    const isDialogShow = ref(false);
    const rootAccountList = ref([]);
    const billStore = useBillStore();
    const computedExtension = computed(() => {
      let extension = {}; // aws\gcp
      switch (info.value.vendor) {
        case 'azure':
          extension = {
            cloud_subscription_name: '', // 订阅名
            cloud_subscription_id: '', // 订阅ID
            cloud_init_password: '', // 初始密码
          };
          break;
        case 'huawei':
          extension = {
            cloud_main_account_name: '', // 二级账号名
            cloud_main_account_id: '', // 二级账号ID
            cloud_init_password: '', // 初始密码
          };
          break;
        case 'zenlayer':
        case 'kaopu':
          extension = {
            cloud_main_account_name: '', // 二级账号名
            cloud_main_account_id: '', // 二级账号ID
            cloud_init_password: '', // 初始密码
          };
          break;
      }
      return extension;
    });
    const { formModel } = useFormModel({
      root_account_id: '', // 一级账号ID
      extension: computedExtension.value, // 不同云厂商的信息
    });

    onMounted(async () => {
      const { data } = await billStore.root_accounts_list({
        filter: {
          op: 'and',
          rules: [],
        },
        page: {
          limit: 500,
          start: 0,
          count: false,
        },
      });
      rootAccountList.value = data.details.map((v: any) => ({
        name: v.name,
        id: v.id,
        key: v.id,
      }));
    });

    return () => (
      <div class={'account-apply-detail-container'}>
        <div class={'card-wrapper'}>
          <div class={'flex-row align-item-center'}>
            <div class={'flex-row align-item-center'}>
              {statusMap[props.detail.status].tag === 'success' && <Success height={21} width={21} fill='#2DCB56' />}
              {statusMap[props.detail.status].tag === 'abort' && <Close height={21} width={21} fill='#EA3636' />}
              {statusMap[props.detail.status].tag === 'pending' && <Spinner height={21} width={21} fill='#3A84FF' />}
              <div class={'ml4'}>{statusMap[props.detail.status].label}</div>
            </div>
            <div class='approval-process-wrapper' onClick={() => window.open(props.detail.ticket_url, '_blank')}>
              审批单详情
              <i class='hcm-icon bkhcm-icon-jump-fill'></i>
            </div>
          </div>
        </div>
        <div class={'card-wrapper mt24'}>
          <p class={'title'}>申请单信息</p>
          <DetailInfo
            detail={info.value}
            fields={[
              {
                prop: 'vendor',
                name: '云厂商',
              },
              {
                prop: 'bk_biz_id',
                name: '业务',
                render: () => businessMapStore.businessMap.get(info.value.bk_biz_id) || info.value.bk_biz_id || '--',
              },
              {
                prop: 'id',
                name: '一级账号ID',
              },
              {
                prop: 'managers',
                name: '主负责人',
              },
              {
                prop: 'bak_managers',
                name: '备份负责人',
              },
              {
                prop: 'op_product_id',
                name: '运营产品',
              },
            ]}
          />
        </div>

        <div class={'card-wrapper mt24'}>
          <p class={'title'}>账号信息</p>
          {statusMap[props.detail.status].tag === 'success' ? (
            <DetailInfo
              wide
              detail={info.value}
              fields={[
                {
                  prop: 'id',
                  name: '二级账号ID',
                },
                {
                  prop: 'name',
                  name: '账号名称',
                },
              ]}
            />
          ) : (
            <Exception scene='part' type='empty' description='当前账号未创建'>
              <Button text theme='primary' class={'create-account-btn'} onClick={() => (isDialogShow.value = true)}>
                录入账号
              </Button>
            </Exception>
          )}
        </div>

        <CommonDialog
          v-model:isShow={isDialogShow.value}
          title='录入已创建账号'
          onHandleConfirm={async () => {
            await billStore.complete_main_account({
              ...formModel,
              sn: props.detail.sn,
              id: props.detail.id,
              vendor: info.value.vendor,
            });
            Message({
              message: '录入成功',
              theme: 'success',
            });
          }}>
          <Form formType='vertical'>
            <FormItem label='所属一级账号'>
              <Select v-model={formModel.root_account_id} clearable={false}>
                {rootAccountList.value.map(({ name, id, key }) => (
                  <Option name={name} id={id} key={key}></Option>
                ))}
              </Select>
            </FormItem>
            {info.value.vendor === 'azure' && (
              <>
                <FormItem label='订阅名称' required>
                  <Input v-model={formModel.extension.cloud_subscription_name} />
                </FormItem>
                <FormItem label='订阅ID' required>
                  <Input v-model={formModel.extension.cloud_subscription_id} />
                </FormItem>
                <FormItem label='初始密码' required>
                  <Input v-model={formModel.extension.cloud_init_password} />
                </FormItem>
              </>
            )}

            {info.value.vendor === 'huawei' && (
              <>
                <FormItem label='二级账号名称' required>
                  <Input v-model={formModel.extension.cloud_main_account_name} />
                </FormItem>
                <FormItem label='二级账号ID' required>
                  <Input v-model={formModel.extension.cloud_main_account_id} />
                </FormItem>
                <FormItem label='初始密码' required>
                  <Input v-model={formModel.extension.cloud_init_password} />
                </FormItem>
              </>
            )}

            {['zenlayer', 'kaopu'].includes(info.value.vendor) && (
              <>
                <FormItem label='二级账号名称' required>
                  <Input v-model={formModel.extension.cloud_main_account_name} />
                </FormItem>
                <FormItem label='二级账号ID' required>
                  <Input v-model={formModel.extension.cloud_main_account_id} />
                </FormItem>
                <FormItem label='初始密码' required>
                  <Input v-model={formModel.extension.cloud_init_password} />
                </FormItem>
              </>
            )}
          </Form>
        </CommonDialog>
      </div>
    );
  },
});
