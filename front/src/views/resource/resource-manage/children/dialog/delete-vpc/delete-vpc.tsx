import {
  defineComponent,
  ref,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
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

  setup(props, { emit }) {
    const {
      t,
    } = useI18n();

    // 状态
    const tableData = ref([]);
    const columns: any[] = [{ label: '23' }];
    const steps = [
      {
        component: () => <>
          <bk-table
            class="mb20"
            row-hover="auto"
            columns={columns}
            data={tableData.value}
          />
          <h3 class="g-resource-tips">
            { t('请注意该VPC包含一个或多个资源，在释放这些资源前，无法删除VPC') }：<br />
            { t('子网：{count} 个', { count: 5 }) }<br />
            { t('CVM：{count} 个', { count: 5 }) }
          </h3>
        </>,
      },
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      handleClose();
    };

    return {
      steps,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return <>
      <step-dialog
        title={this.t('删除 VPC')}
        isShow={this.isShow}
        steps={this.steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});
