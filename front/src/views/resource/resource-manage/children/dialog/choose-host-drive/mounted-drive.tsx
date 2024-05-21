import { Table, Loading, Radio, Message } from 'bkui-vue';
import { InfoLine } from 'bkui-vue/lib/icon';
import { defineComponent, h, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import useQueryList from '../../../hooks/use-query-list';
import useColumns from '../../../hooks/use-columns';
import { useResourceStore } from '@/store/resource';

// 硬盘选主机挂载
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

  emits: ['update:isShow', 'success-attach'],

  setup(props, { emit }) {
    const { t } = useI18n();

    const deviceName = ref();
    const cachingType = ref();

    const cacheTypes = ['None', 'ReadOnly', 'ReadWrite'];

    const rules: any[] = [
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

    if (!location.href.includes('business')) {
      rules.push({
        field: 'bk_biz_id',
        op: 'eq',
        value: -1,
      });
    }

    if (props.detail.vendor === 'azure') {
      // rules.push({
      //   field: 'extension.resource_group_name',
      //   op: 'json_eq',
      //   value: props.detail.resource_group_name
      // })
      rules.splice(2, 1);
      if (!props.detail.zones) {
        rules.push({
          field: 'extension',
          op: 'json_not_contains_path',
          value: 'zones',
        });
      }
      if (Array.isArray(props.detail.zones) && props.detail.zones.length > 0) {
        rules.push({
          op: 'or',
          rules: [
            {
              field: 'extension',
              op: 'json_not_contains_path',
              value: 'zones',
            },
            {
              field: 'extension.zones',
              op: 'json_overlaps',
              value: props.detail.zones,
            },
          ],
        });
      }
    }

    const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
      {
        filter: {
          op: 'and',
          rules,
        },
      },
      'cvms',
      null,
      'getUnbindDiskCvms',
      {
        not_equal_disk_id: props.detail.id,
      },
    );

    const { columns } = useColumns('cvms', true);

    const resourceStore = useResourceStore();

    const selection = ref<any>({});

    const isConfirmLoading = ref(false);

    const renderColumns = [
      {
        label: 'ID',
        field: 'id',
        render({ data }: any) {
          return h(Radio, {
            'model-value': selection.value.id,
            label: data.id,
            key: data.id,
            onChange() {
              selection.value = data;
            },
          });
        },
      },
      ...columns,
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      const postData: any = {
        disk_id: props.detail.id,
        cvm_id: selection.value.id,
      };
      if (!selection.value.id) {
        Message({
          theme: 'error',
          message: '请先选择主机',
        });
        return;
      }
      if (props.detail.vendor === 'aws') {
        if (!deviceName.value) {
          Message({
            theme: 'error',
            message: '请先输入设备名称',
          });
          return;
        }
        postData.device_name = deviceName.value;
      }
      if (props.detail.vendor === 'azure') {
        if (!cachingType.value) {
          Message({
            theme: 'error',
            message: '请先选择缓存类型',
          });
          return;
        }
        postData.caching_type = cachingType.value;
      }
      isConfirmLoading.value = true;
      resourceStore
        .attachDisk(postData)
        .then(() => {
          emit('success-attach');
          handleClose();
        })
        .catch((err: any) => {
          Message({
            theme: 'error',
            message: err.message || err,
          });
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
      handlePageChange,
      handlePageSizeChange,
      handleSort,
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
                <bk-input
                  v-model={this.deviceName}
                  placeholder='/dev/sdb'
                  style='width: 200px;margin-right: 5px'></bk-input>
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
            />
          </Loading>
        ),
      },
    ];

    return (
      <>
        <step-dialog
          title={this.t('挂载云硬盘')}
          isShow={this.isShow}
          steps={steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
