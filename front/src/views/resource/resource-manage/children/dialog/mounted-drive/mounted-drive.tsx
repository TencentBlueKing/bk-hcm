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
    const chooseDrive = ref([]);
    const sources = ref([
      { service_code: 'pipeline', service_name: '流水线' },
      { service_code: 'codecc', service_name: '代码检查' },
      { service_code: 'bcs', service_name: '容器服务' },
      { service_code: 'artifactory', service_name: '版本仓库' },
      { service_code: 'ticket', service_name: '凭证管理' },
      { service_code: 'code', service_name: '代码库' },
      { service_code: 'experience', service_name: '版本体验' },
      { service_code: 'environment', service_name: '环境管理' },
    ]);
    const steps = [
      {
        component: () => <>
          <bk-transfer
            target-list={chooseDrive.value}
            source-list={sources.value}
            title={[t('选择云硬盘'), t('已选择')]}
            empty-content={[t('暂无云硬盘'), t('未选择任何云硬盘')]}
            display-key="service_name"
            setting-key="service_code"
          />
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
        title={this.t('挂载云硬盘')}
        isShow={this.isShow}
        steps={this.steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});
