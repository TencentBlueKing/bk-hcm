import { PropType, Ref, defineComponent, inject, nextTick, provide, reactive, ref, watch } from 'vue';
import { Form, Message, Select } from 'bkui-vue';
import PrimaryAccountSelector from '../../components/search/primary-account-selector';
import VendorRadioGroup from '@/components/vendor-radio-group';
import CommonSideslider from '@/components/common-sideslider';

import { useI18n } from 'vue-i18n';
import AdjustTable from './AdjustTable';
import useFormModel from '@/hooks/useFormModel';
import { VendorEnum } from '@/common/constant';
import { BILLS_CURRENCY } from '@/constants/bill';
import useBillStore, { UpdateAdjustmentItemParams } from '@/store/useBillStore';
import AdjustAmount from './AdjustAmount';

const { Option } = Select;

export default defineComponent({
  props: {
    edit: {
      required: true,
      type: Boolean,
    },
    editData: {
      required: true,
      type: Object as PropType<UpdateAdjustmentItemParams>,
    },
  },
  emits: ['update', 'clearEdit'],
  setup(props, { expose, emit }) {
    const { t } = useI18n();
    const billStore = useBillStore();
    const isShow = ref(false);
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');
    const { formModel, resetForm } = useFormModel({
      vendor: VendorEnum.AZURE,
      root_account_id: '',
      currency: 'USD',
    });
    const editData = ref<UpdateAdjustmentItemParams>({});
    const isLoading = ref(false);
    const formRef = ref();
    const tableRef = ref();

    const triggerShow = (v: boolean) => {
      isShow.value = v;
      nextTick(() => {
        formRef.value?.clearValidate();
      });
    };
    const costSum = reactive({
      increaseSum: 0,
      decreaseSum: 0,
    });
    const setCostSum = (increaseVal: number, decreaseVal: number) => {
      costSum.increaseSum = increaseVal;
      costSum.decreaseSum = decreaseVal;
    };
    provide('adjust_bill_set_costSum', setCostSum);

    expose({ triggerShow });

    const handleSubmit = async () => {
      const [tableData] = await Promise.all([tableRef.value.getValue(), formRef.value.validate()]);
      try {
        isLoading.value = true;
        const paylaod = {
          ...formModel,
          items: tableData.map((row: any) => {
            return {
              ...row,
              bill_year: bill_year.value,
              bill_month: bill_month.value,
              ...formModel,
              product_id: !!row.product_id ? row.product_id : undefined,
              bk_biz_id: !!row.bk_biz_id ? row.bk_biz_id : undefined,
            };
          }),
        };
        if (props.edit)
          await billStore.update_adjustment_item(editData.value.id, {
            ...editData.value,
            ...tableData?.[0],
            ...formModel,
          });
        else await billStore.create_adjustment_items(paylaod);
        isShow.value = false;
        Message({
          message: props.edit ? t('编辑成功') : t('创建成功'),
          theme: 'success',
        });
        emit('update');
      } finally {
        isLoading.value = false;
      }
    };

    const reset = () => {
      resetForm();
      // emit('clearEdit');
      tableRef.value?.reset();
    };

    watch(
      [() => props.edit, () => props.editData],
      ([isEdit, data]) => {
        if (!isEdit) reset();
        else {
          editData.value = data;
          formModel.currency = data.currency;
          formModel.vendor = data.vendor;
          formModel.root_account_id = data.root_account_id;
        }
      },
      {
        deep: true,
      },
    );

    return () => (
      <CommonSideslider
        renderType='if'
        v-model:isShow={isShow.value}
        width={1280}
        title='新增调账'
        onHandleSubmit={handleSubmit}
        isSubmitLoading={isLoading.value}>
        {{
          default: () => (
            <Form formType='vertical' ref={formRef} model={formModel}>
              <Form.FormItem label={t('云厂商')} required property='vendor'>
                <VendorRadioGroup
                  disabled={props.edit}
                  v-model={formModel.vendor}
                  onUpdate:modelValue={() => {
                    if (!props.edit) formModel.root_account_id = '';
                  }}
                />
              </Form.FormItem>
              <Form.FormItem label={t('一级账号')} required property='root_account_id'>
                <PrimaryAccountSelector
                  disabled={props.edit}
                  v-model={formModel.root_account_id}
                  multiple={false}
                  vendor={formModel.vendor}
                  autoSelect={!props.edit}
                />
              </Form.FormItem>
              <Form.FormItem label={'币种'} required property='currency'>
                <Select v-model={formModel.currency}>
                  {BILLS_CURRENCY.map(({ name, id }) => (
                    <Option name={name} id={id} key={id} />
                  ))}
                </Select>
              </Form.FormItem>
              <Form.FormItem label={t('调账配置')} required>
                <div>
                  <AdjustTable
                    ref={tableRef}
                    vendor={formModel.vendor}
                    edit={props.edit}
                    editData={editData.value}
                    rootAccountId={formModel.root_account_id}
                  />
                </div>
              </Form.FormItem>
              <Form.FormItem label={t('结果预览')}>
                <AdjustAmount adjustData={costSum} currency={formModel.currency as 'RMB' | 'USD'} />
              </Form.FormItem>
            </Form>
          ),
        }}
      </CommonSideslider>
    );
  },
});
