import { defineComponent, onMounted, onUnmounted, ref, watch } from 'vue';
// import components
import { Radio } from 'bkui-vue';
import CommonDialog from '@/components/common-dialog';
import LocalTable from '../../components/common/LocalTable';
// import types
import type { SpecAvailability, ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
// import utils
import { useI18n } from 'vue-i18n';
import bus from '@/common/bus';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';

// apply-clb, 性能容量型弹窗
export default (formModel: ApplyClbModel) => {
  // use hooks
  const { t } = useI18n();
  // define data
  const isClbSpecTypeDialogShow = ref(false);
  const selectedClbSpecType = ref('');
  const tableData = ref<Array<SpecAvailability>>([]);
  const columns = [
    {
      label: t('规格类型'),
      field: 'SpecType',
      render: ({ data }: any) => {
        return <Radio v-model={selectedClbSpecType.value} label={data.SpecType} />;
      },
    },
    {
      label: t('规格可用性'),
      field: 'Availability',
    },
  ];
  const searchData: Array<ISearchItem> = [
    { id: 'SpecType', name: '规格类型' },
    { id: 'Availability', name: '规格可用性' },
  ];

  // 点击表格row触发的钩子
  const handleRowClick = (row: SpecAvailability) => {
    selectedClbSpecType.value = row.SpecType;
  };

  // 选择机型
  const handleSelectClbSpecType = () => {
    formModel.sla_type = selectedClbSpecType.value;
  };

  const SelectClbSpecTypeDialog = defineComponent({
    setup() {
      return () => (
        <CommonDialog
          v-model:isShow={isClbSpecTypeDialogShow.value}
          title='选择实例规格'
          width={'60vw'}
          onHandleConfirm={handleSelectClbSpecType}>
          <LocalTable
            tableData={tableData.value}
            tableColumns={columns}
            searchOption={{
              filterable: true,
              data: searchData,
            }}
            onRowClick={(_: any, row: any) => handleRowClick(row)}
          />
        </CommonDialog>
      );
    },
  });

  watch(
    () => formModel.sla_type,
    (val) => {
      if (val === 'shared') {
        selectedClbSpecType.value = '';
      }
    },
  );

  onMounted(() => {
    bus.$on('showSelectClbSpecTypeDialog', () => {
      isClbSpecTypeDialogShow.value = true;
    });
    bus.$on('updateSpecAvailabilitySet', (data: any[]) => {
      tableData.value = data;
    });
  });

  onUnmounted(() => {
    bus.$off('showSelectClbSpecTypeDialog');
    bus.$off('updateSpecAvailabilitySet');
  });

  return {
    SelectClbSpecTypeDialog,
  };
};
