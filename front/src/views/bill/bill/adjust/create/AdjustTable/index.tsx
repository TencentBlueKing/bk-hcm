import { PropType, defineComponent, inject, ref } from 'vue';
import './index.scss';
import RenderTable from '../RenderTable';
import RenderTableRow from '../RenderTableRow';
import { AdjustmentItem } from '@/typings/bill';
import { VendorEnum } from '@/common/constant';
import { AdjustTypeEnum } from '../RenderTableRow/components/AdjustTypeSelector';

export default defineComponent({
  props: {
    edit: {
      default: false,
      type: Boolean,
    },
    editData: {
      default: [],
      type: Object,
    },
    vendor: {
      required: true,
      type: String as PropType<VendorEnum>,
    },
    rootAccountId: {
      required: true,
      type: String,
    },
  },
  setup(props, { expose }) {
    const Record = (): Partial<AdjustmentItem> => {
      return {
        product_id: -1,
        bk_biz_id: -1,
        type: AdjustTypeEnum.Increase,
        cost: '',
        main_account_id: '',
        memo: '',
      };
    };
    const rowRefs = [ref(null)];
    const tableData = ref([Record()]);
    const setCostSum = inject<Function>('adjust_bill_set_costSum');

    const handleCostChange = async () => {
      const tableData = rowRefs.map((row) => row.value.getRowValue());
      let increaseSum = 0;
      let decreaseSum = 0;
      for (const { cost, type } of tableData) {
        if (type === AdjustTypeEnum.Increase) increaseSum += +cost;
        else decreaseSum += +cost;
      }
      setCostSum(increaseSum, decreaseSum);
    };

    expose({
      getValue: async () => {
        return await Promise.all(rowRefs.map((row) => row.value.getValue()));
      },
      reset: () => {
        tableData.value = [Record()];
        rowRefs.map((row) => row.value.reset());
      },
    });

    return () => (
      <div class={'adjust-table-container'}>
        <RenderTable edit={props.edit}>
          {tableData.value.map((item, idx) => (
            <RenderTableRow
              edit={props.edit}
              editData={props.edit ? props.editData : item}
              vendor={props.vendor}
              rootAccountId={props.rootAccountId}
              onCostChange={handleCostChange}
              onAdd={() => {
                tableData.value.push(Record());
                rowRefs.push(ref(null));
              }}
              onRemove={() => {
                tableData.value.splice(idx, 1);
                rowRefs.splice(idx, 1);
              }}
              onCopy={() => {
                tableData.value.splice(idx, 0, tableData.value[idx]);
              }}
              removeable={tableData.value.length < 2}
              ref={rowRefs[idx]}
            />
          ))}
        </RenderTable>
      </div>
    );
  },
});
