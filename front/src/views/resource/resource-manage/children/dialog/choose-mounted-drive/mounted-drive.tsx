import { Table, Loading, Radio, Message } from 'bkui-vue';
import { defineComponent, h, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { InfoLine } from 'bkui-vue/lib/icon';
import StepDialog from '@/components/step-dialog/step-dialog';
import useQueryList from '../../../hooks/use-query-list';
import useColumns from '../../../hooks/use-columns';
import { useResourceStore } from '@/store/resource';

// 主机选硬盘挂载
export default defineComponent({
  components: {
    StepDialog,
    InfoLine,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
    detail: {
      type: Object,
    },
  },

  emits: ['update:isShow', 'success'],

  setup(props, { emit }) {
    const { t } = useI18n();

    const deviceName = ref();
    const cachingType = ref();

    const cacheTypes = ['None', 'ReadOnly', 'ReadWrite'];

    const rules = [
      {
        field: 'vendor',
        op: 'eq',
        value: props.detail.vendor,
      },
      {
        field: 'account_id',
        op: 'eq',
        value: props.detail.account_id,
      },
      {
        field: 'zone',
        op: 'eq',
        value: props.detail.zone,
      },
      {
        field: 'region',
        op: 'eq',
        value: props.detail.region,
      },
    ];

    // if (props.detail.vendor === 'azure') {
    //   rules.push({
    //     field: 'extension.resource_group_name',
    //     op: 'json_eq',
    //     value: props.detail.resource_group_name
    //   })
    // }

    const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
      {
        filter: {
          op: 'and',
          rules,
        },
      },
      'disks',
      null,
      'getUnbindCvmDisks',
    );

    const { columns } = useColumns('drive', true);

    const resourceStore = useResourceStore();

    const selectedId = ref<string>();

    const isConfirmLoading = ref(false);

    const renderColumns = [
      {
        label: '',
        field: 'radio',
        width: 40,
        minWidth: 40,
        render: ({ data }: any) => {
          const { id } = data;
          return h(
            'div',
            { class: 'flex-row align-items-center' },
            h(Radio, { label: id, key: id, modelValue: selectedId.value }, ' '),
          );
        },
      },
      { label: '硬盘ID', field: 'id' },
      ...columns.filter((column: any) => ['资源 ID', '云硬盘名称', '类型', '容量(GB)', '状态'].includes(column.label)),
    ];

    const handleRowClick = (_event: PointerEvent, row: any) => {
      selectedId.value = row.id;
    };

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      isConfirmLoading.value = true;
      const postData: any = {
        disk_id: selectedId.value,
        cvm_id: props.detail.id,
      };
      if (!selectedId.value) {
        Message({ theme: 'error', message: '请先选择云硬盘' });
        return;
      }
      if (props.detail.vendor === 'aws') {
        if (!deviceName.value) {
          Message({ theme: 'error', message: '请先输入设备名称' });
          return;
        }
        postData.device_name = deviceName.value;
      }
      if (props.detail.vendor === 'azure') {
        if (!cachingType.value) {
          Message({ theme: 'error', message: '请先选择缓存类型' });
          return;
        }
        postData.caching_type = cachingType.value;
      }
      resourceStore
        .attachDisk(postData)
        .then(() => {
          Message({ theme: 'success', message: '云硬盘挂载成功' });
          emit('success');
          handleClose();
        })
        .catch((err: any) => {
          Message({ theme: 'error', message: err.message || err });
        })
        .finally(() => {
          isConfirmLoading.value = false;
        });
    };

    return {
      deviceName,
      cachingType,
      cacheTypes,
      datas,
      pagination,
      isLoading,
      renderColumns,
      isConfirmLoading,
      selectedId,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
      handleRowClick,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const tooltipSlot = {
      content: () => (
        <>
          Linux设备名称参考：https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/UserGuide/device_naming.html
          <br />
          windows设备名称参考：https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/WindowsGuide/device_naming.html
        </>
      ),
    };
    const steps = [
      {
        isConfirmLoading: this.isConfirmLoading,
        component: () => (
          <Loading loading={this.isLoading}>
            {this.detail.vendor === 'aws' ? (
              <>
                <span class='mr10'>设备名称:</span>
                <bk-input v-model={this.deviceName} style='width: 200px;'></bk-input>
                <bk-popover placement='top' v-slots={tooltipSlot}>
                  <InfoLine />
                </bk-popover>
              </>
            ) : (
              ''
            )}
            {this.detail.vendor === 'azure' ? (
              <>
                <span class='mr10'>缓存类型:</span>
                <bk-select v-model={this.cachingType} style='width: 200px;display: inline-block;'>
                  {this.cacheTypes.map((type) => (
                    <bk-option key={type} value={type} label={type} />
                  ))}
                </bk-select>
              </>
            ) : (
              ''
            )}
            <Table
              class='mt20'
              row-hover='auto'
              remote-pagination
              pagination={this.pagination}
              columns={this.renderColumns}
              data={this.datas}
              onPageLimitChange={this.handlePageSizeChange}
              onPageValueChange={this.handlePageChange}
              onColumnSort={this.handleSort}
              onRowClick={this.handleRowClick}
            />
          </Loading>
        ),
      },
    ];

    return (
      <step-dialog
        title={this.t('挂载云硬盘')}
        isShow={this.isShow}
        steps={steps}
        confirmDisabled={!this.selectedId}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}></step-dialog>
    );
  },
});
