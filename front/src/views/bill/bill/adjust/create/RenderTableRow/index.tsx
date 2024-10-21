import { PropType, defineComponent, ref, watch } from 'vue';
import { InputColumn, OperationColumn, TextPlainColumn } from '@blueking/ediatable';
import AdjustTypeSelector, { AdjustTypeEnum } from './components/AdjustTypeSelector';
import SubAccountSelector from '../../../components/search/sub-account-selector';
import { VendorEnum } from '@/common/constant';
import useFormModel from '@/hooks/useFormModel';
import { diffModelKey, getDiffSelectorComp } from '../adjust-create-table.plugin';

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
  emits: ['add', 'remove', 'copy', 'change'],
  setup(props, { emit, expose }) {
    const { formModel, resetForm, setFormValues } = useFormModel({
      type: AdjustTypeEnum.Increase,
      [diffModelKey]: '',
      main_account_id: '',
      cost: '',
      memo: '',
    });

    const costRef = ref();
    const memoRef = ref();
    const selectorRef = ref();
    const mainAccountRef = ref();

    const handleAdd = () => {
      emit('add');
    };

    const handleRemove = () => {
      emit('remove');
    };

    const handleCopy = () => {
      emit('copy', formModel);
    };

    watch(
      () => props.editData,
      (data) => {
        setFormValues(data);
      },
      {
        deep: true,
        immediate: true,
      },
    );

    watch(
      () => formModel,
      (val) => {
        emit('change', val);
      },
      {
        deep: true,
      },
    );

    watch(
      () => props.rootAccountId,
      () => {
        formModel.main_account_id = '';
      },
    );

    expose({
      getValue: async () => {
        return await Promise.all([
          costRef.value!.getValue(),
          memoRef.value!.getValue(),
          mainAccountRef.value!.getValue(),
        ]).then(() => {
          return formModel;
        });
      },
      reset: resetForm,
      getRowValue: () => {
        return formModel;
      },
    });

    return () => (
      <>
        <tr>
          <td>
            <AdjustTypeSelector v-model={formModel.type} />
          </td>
          <td>{getDiffSelectorComp(formModel, selectorRef)}</td>
          <td>
            <SubAccountSelector
              isEditable={true}
              v-model={formModel.main_account_id}
              ref={mainAccountRef}
              vendor={[props.vendor]}
              rootAccountId={props.rootAccountId ? [props.rootAccountId] : []}
            />
          </td>
          <td>
            <TextPlainColumn>人工调账</TextPlainColumn>
          </td>
          <td>
            <InputColumn
              type='number'
              min={0}
              precision={3}
              ref={costRef}
              v-model={formModel.cost}
              rules={[
                {
                  validator: (value: string) => Boolean(value),
                  message: '金额不能为空',
                },
              ]}
            />
          </td>
          <td>
            <InputColumn ref={memoRef} v-model={formModel.memo} />
          </td>
          {!props.edit && (
            <OperationColumn
              removeable={props.removeable}
              onAdd={handleAdd}
              onRemove={handleRemove}
              showCopy
              onCopy={handleCopy}
            />
          )}
        </tr>
      </>
    );
  },
});
