import { defineComponent, ref } from 'vue';

import { useI18n } from 'vue-i18n';
import { useResourceStore } from '@/store/resource';

import StepDialog from '@/components/step-dialog/step-dialog';

export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
  },

  emits: ['update:isShow'],

  setup(_, { emit }) {
    const { t } = useI18n();

    const resourceStore = useResourceStore();

    // 状态
    const isDeleting = ref(false);
    const tableData = ref([]);
    const columns = [
      {
        label: 'ID',
        field: 'id',
      },
      {
        label: '资源ID',
        field: 'cid',
      },
      {
        label: '名称',
        field: 'name',
      },
      {
        label: '云厂商',
        field: 'vendor',
      },
      {
        label: '业务',
        field: 'id',
      },
      {
        label: '业务拓扑',
        field: 'id',
      },
      {
        label: '云区域',
        field: 'id',
      },
      {
        label: '地域',
        field: 'region',
      },
      {
        label: 'IPv4 CIDR',
        field: 'ipv4_cidr',
      },
      {
        label: 'IPv6 CIDR',
        field: 'ipv6_cidr',
      },
      {
        label: '状态',
        field: 'status',
      },
      {
        label: '默认VPC',
        field: 'is_default',
      },
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      isDeleting.value = true;
      resourceStore
        .deleteBatch('security_groups', { ids: '[111]' })
        .then(() => {
          handleClose();
        })
        .finally(() => {
          isDeleting.value = false;
        });
    };

    return {
      isDeleting,
      columns,
      tableData,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const steps = [
      {
        component: () => (
          <>
            <bk-table
              class='mb20'
              row-hover='auto'
              columns={this.columns}
              data={this.tableData}
              show-overflow-tooltip
            />
            <h3 class='g-resource-tips'>
              {this.t('安全组被实例关联或者被其他安全组规则关联时不能直接删除，请删除关联关系后再进行删除')}：
              <bk-button text theme='primary'>
                {this.t('查看关联实例')}
              </bk-button>
              <br />
            </h3>
          </>
        ),
        isConfirmLoading: this.isDeleting,
      },
    ];

    return (
      <>
        <step-dialog
          title={this.t('删除 安全组')}
          isShow={this.isShow}
          steps={steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
