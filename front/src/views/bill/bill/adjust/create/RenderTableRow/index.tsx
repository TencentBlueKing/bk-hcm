import { PropType, defineComponent, ref, watch } from 'vue';
import { InputColumn, OperationColumn, TextPlainColumn } from '@blueking/ediatable';
import AdjustTypeSelector, { AdjustTypeEnum } from './components/AdjustTypeSelector';
import BusinessSelector from '@/components/business-selector/index.vue';
import SubAccountSelector from '../../../components/search/sub-account-selector';
import { VendorEnum } from '@/common/constant';

export default defineComponent({
  props: {
    removeable: {
      required: true,
      type: Boolean,
      default: false,
    },
    vendor: {
      required: true,
      type: String as PropType<VendorEnum>,
    },
    rootAccountId: {
      required: true,
      type: String,
    },
    editData: {
      required: true,
      type: Object,
      default: {},
    },
    edit: {
      required: true,
      type: Boolean,
    },
  },
  emits: ['add', 'remove', 'costChange', 'copy'],
  setup(props, { emit, expose }) {
    const type = ref(AdjustTypeEnum.Increase);
    const bk_biz_id = ref('');
    const main_account_id = ref('');
    const cost = ref('');
    const memo = ref('');

    const costRef = ref();
    const memoRef = ref();
    const bizIdRef = ref();
    const mainAccountRef = ref();

    const handleAdd = () => {
      emit('add');
    };

    const handleRemove = () => {
      emit('remove');
    };

    const reset = () => {
      type.value = AdjustTypeEnum.Increase;
      bk_biz_id.value = '';
      main_account_id.value = '';
      cost.value = '';
      memo.value = '';
    };

    watch(
      [() => props.edit, () => props.editData],
      ([isEdit, data]) => {
        if (isEdit) {
          type.value = data.type;
          main_account_id.value = data.main_account_id;
          bk_biz_id.value = data.bk_biz_id;
          cost.value = data.cost;
          memo.value = data.memo;
        } else {
          reset();
        }
      },
      {
        deep: true,
        immediate: true,
      },
    );

    watch([() => cost.value, () => type.value], ([val]) => {
      emit('costChange', val ? val : 0);
    });

    expose({
      getValue: async () => {
        return await Promise.all([
          costRef.value!.getValue(),
          memoRef.value!.getValue(),
          // bizIdRef.value!.getValue(),
          mainAccountRef.value!.getValue(),
        ]).then((data) => {
          const [cost, memo] = data;
          return {
            type: type.value,
            bk_biz_id: bk_biz_id.value,
            main_account_id: main_account_id.value,
            cost,
            memo,
          };
        });
      },
      reset,
      getRowValue: () => {
        return {
          type: type.value,
          bk_biz_id: bk_biz_id.value,
          main_account_id: main_account_id.value,
          cost: cost.value,
          memo: memo.value,
        };
      },
    });

    return () => (
      <>
        <tr>
          <td>
            <AdjustTypeSelector v-model={type.value} />
          </td>
          <td>
            <BusinessSelector v-model={bk_biz_id.value} ref={bizIdRef} isEditable />
          </td>
          <td>
            <SubAccountSelector
              isEditable={true}
              v-model={main_account_id.value}
              ref={mainAccountRef}
              vendor={[props.vendor]}
              rootAccountId={[props.rootAccountId]}
            />
          </td>
          <td>
            <TextPlainColumn>人工调账</TextPlainColumn>
          </td>
          <td>
            <InputColumn
              type='number'
              precision={3}
              ref={costRef}
              v-model={cost.value}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '金额不能为空',
                },
              ]}
            />
          </td>
          <td>
            <InputColumn ref={memoRef} v-model={memo.value} />
          </td>
          {!props.edit && <OperationColumn removeable={props.removeable} onAdd={handleAdd} onRemove={handleRemove} />}
        </tr>
      </>
    );
  },
});
