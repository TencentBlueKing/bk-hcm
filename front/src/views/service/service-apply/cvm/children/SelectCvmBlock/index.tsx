import { Button, Dialog } from 'bkui-vue';
import { defineComponent, ref } from 'vue';
import './index.scss';
import { Plus } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useTable/useTable';
import { QueryRuleOPEnum } from '@/typings';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    vendor: {
      type: String,
    },
    region: {
      type: String,
    },
    zone: {
      type: String,
    },
    accountId: {
      type: String,
    },
  },
  setup(props) {
    const isSelected = ref(false);
    const isDialogShow = ref(false);
    const columns = [
      {
        label: '类型',
        field: 'instance_family',
      },
      {
        label: '规格',
        field: 'instance_type',
      },
      {
        label: 'CPU',
        field: 'cpu',
      },
      {
        label: '内存',
        field: 'memory',
      },
      {
        label: '处理器型号',
        field: 'cpu_type',
      },
      {
        label: '网络收发包',
        field: 'instance_pps',
      },
      {
        label: '参考费用',
        field: 'price',
      },
    ];
    const { CommonTable } = useTable({
      columns,
      searchData: [],
      searchUrl: `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/instance_types/list`,
      defaultFilterRules: [
        {
          op: QueryRuleOPEnum.EQ,
          field: 'vendor',
          value: props.vendor,
        },
        {
          op: QueryRuleOPEnum.EQ,
          field: 'region',
          value: props.region,
        },
        {
          op: QueryRuleOPEnum.EQ,
          field: 'zone',
          value: props.zone,
        },
        {
          op: QueryRuleOPEnum.EQ,
          field: 'account_id',
          value: props.accountId,
        },
      ],
    });
    return () => (
      <>
        <div>
          {isSelected.value ? (
            <div class={'selected-block-container'}>
              <div class={'selected-block'}>Amazon Linux 2 AMI (HVM) - Kernel 5.10, SSD Volume Type</div>
            </div>
          ) : (
            <Button onClick={() => (isDialogShow.value = true)}>
              <Plus class='f20' />
              选择机型
            </Button>
          )}
        </div>
        <Dialog
          isShow={isDialogShow.value}
          title={'选择机型'}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={() => (isDialogShow.value = true)}>
          <CommonTable />
        </Dialog>
      </>
    );
  },
});
