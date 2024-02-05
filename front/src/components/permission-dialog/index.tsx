/* eslint-disable no-nested-ternary */
import { defineComponent, PropType, watch, ref, h } from 'vue';
import permissions from '@/assets/image/403.png';
import { useVerify } from '@/hooks';
import './index.scss';
import { useI18n } from 'vue-i18n';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useAccountStore } from '@/store';
import { useBusinessMapStore } from '@/store/useBusinessMap';

type permissionType = {
  system_id: string;
  actions: any;
};

export default defineComponent({
  name: 'PermissionDialog',
  props: {
    title: {
      type: String,
      default: '',
    },
    isShow: {
      type: Boolean,
      default: false,
    },
    params: {
      type: Object as PropType<permissionType>,
    },
    size: {
      type: String,
      default: 'medium',
    },
    loading: {
      type: Boolean,
    },
  },

  emits: ['confirm', 'cancel'],

  setup(_, { emit }) {
    const { t } = useI18n();
    const { whereAmI } = useWhereAmI();
    const resourceAccountStore = useResourceAccountStore();
    const accountStore = useAccountStore();
    const businessMapStore = useBusinessMapStore();

    const columns = [
      {
        label: '需要申请的权限',
        field: 'name',
      },
      {
        label: '关联的资源实例',
        field: 'memo',
        render({ data }: any) {
          return h('span', {}, [
            `${data?.related_resource_types[0]?.type_name || '--'}${
              whereAmI.value === Senarios.resource
                ? resourceAccountStore.resourceAccount?.name
                  ? `: ${resourceAccountStore.resourceAccount?.name}`
                  : ''
                : ` ${businessMapStore.getNameFromBusinessMap(accountStore.bizs)}`
            }`,
          ]);
        },
      },
    ];

    const tableData = ref([]);
    const url = ref('');

    const handleClose = () => {
      emit('cancel');
    };

    const handleConfirm = () => {
      emit('confirm', url.value);
    };

    // hook
    const { getActionPermission } = useVerify();

    watch(
      () => _.isShow,
      async (val) => {
        if (val) {
          tableData.value = _.params.actions;
          url.value = await getActionPermission(_.params);
        }
      },
    );

    return {
      t,
      columns,
      tableData,
      handleClose,
      handleConfirm,
      resourceAccountStore,
      whereAmI,
      accountStore,
    };
  },

  render() {
    return (
      <>
        <bk-dialog
          class='permissions-dialog-cls'
          theme='primary'
          width={740}
          height={450}
          title={this.title}
          size={this.size}
          isShow={this.isShow}
          onClosed={this.handleClose}>
          {{
            default: () => {
              return (
                <>
                  <img class='no-permission-img' src={permissions} alt='403'></img>
                  <div class='no-permission-text'>{this.t('没有权限访问或操作此资源')}</div>
                  <bk-table
                    align='left'
                    class='mt20 no-permission-table'
                    row-hover='auto'
                    columns={this.columns}
                    data={this.tableData}
                    show-overflow-tooltip
                  />
                </>
              );
            },
            footer: () => {
              return (
                <>
                  <bk-button
                    class='mr10 dialog-button'
                    theme='primary'
                    loading={this.loading}
                    onClick={this.handleConfirm}>
                    {this.t('去申请')}
                  </bk-button>
                  <bk-button class='dialog-button' onClick={this.handleClose}>
                    {this.t('取消')}
                  </bk-button>
                </>
              );
            },
          }}
        </bk-dialog>
      </>
    );
  },
});
