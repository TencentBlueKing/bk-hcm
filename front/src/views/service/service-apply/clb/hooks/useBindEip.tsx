import { Radio } from 'bkui-vue';
// import components
import { defineComponent, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import CommonDialog from '@/components/common-dialog';
// import hooks
import { useTable } from '@/hooks/useTable/useTable';
// import utils
import bus from '@/common/bus';
// import types
import { type ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { QueryRuleOPEnum } from '@/typings';

// apply-clb, 绑定弹性IP弹窗
export default (formModel: ApplyClbModel) => {
  // define data
  const isBindEipDialogShow = ref(false);
  const selectedEipData = reactive({ id: '', name: '', public_ip: '' });
  const eipListSearchRules = reactive([]);

  // use hooks
  const { CommonTable } = useTable({
    searchOptions: {
      disabled: true,
    },
    tableOptions: {
      columns: [
        {
          label: 'ID',
          field: 'cloud_id',
          render: ({ data }: any) => {
            return <Radio v-model={selectedEipData.id} label={data.cloud_id} />;
          },
        },
        {
          label: '名称',
          field: 'name',
        },
        {
          label: '弹性公网IP',
          field: 'public_ip',
        },
      ],
      extra: {
        onRowClick: (_event: Event, _row: any) => {
          Object.assign(selectedEipData, _row);
        },
      },
    },
    requestOption: {
      type: 'eips',
      filterOption: {
        rules: eipListSearchRules,
      },
    },
  });

  // define handler function
  const handleBindEip = () => {
    formModel.cloud_eip_id = selectedEipData.id;
  };

  // 清除绑定的eip
  const handleClearBind = () => {
    Object.assign(selectedEipData, { id: '', name: '', public_ip: '' });
    handleBindEip();
  };

  // define component
  const BindEipDialog = defineComponent({
    setup() {
      return () => (
        <CommonDialog
          v-model:isShow={isBindEipDialogShow.value}
          title='绑定弹性 IP'
          width={620}
          onHandleConfirm={handleBindEip}>
          <div>选择 EIP</div>
          <CommonTable />
        </CommonDialog>
      );
    },
  });

  watch(
    () => formModel.region,
    (newV) => {
      handleClearBind();
      // 当 region 改变时, 重新获取 EIP 列表
      Object.assign(eipListSearchRules, [
        { op: QueryRuleOPEnum.EQ, field: 'account_id', value: formModel.account_id },
        { op: QueryRuleOPEnum.EQ, field: 'region', value: newV },
        { op: QueryRuleOPEnum.NEQ, field: 'recycle_status', value: 'recycling' },
      ]);
    },
  );

  onMounted(() => {
    bus.$on('showBindEipDialog', () => {
      isBindEipDialogShow.value = true;
    });
  });

  onUnmounted(() => {
    bus.$off('showBindEipDialog');
  });

  return {
    BindEipDialog,
  };
};
