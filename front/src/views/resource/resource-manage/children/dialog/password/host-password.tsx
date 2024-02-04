import { defineComponent, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import './host-password.scss';
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
    const { t } = useI18n();

    // 状态
    const tableData = ref([]);
    const columns: any[] = [
      {
        label: '23',
      },
    ];
    const passwordForm = ref({
      username: '',
      password: '',
      confirmPassword: '',
      shutdown: false,
    });
    const passwordRule = {
      validator(val: string) {
        return /^(?=.*?[A-Za-z])(?=.*?[0-9])(?=.*?[^\w\s]).{8,20}$/.test(val);
      },
      message: t('密码长度8-20位，必须包含英文字母、数字和特殊字符'),
      trigger: 'blur',
    };
    const samePasswordRule = {
      validator(val: string) {
        return val === passwordForm.value.password;
      },
      message: t('确认密码必须和新密码一致'),
      trigger: 'blur',
    };
    const passwordRules = {
      password: [passwordRule],
      confirmPassword: [passwordRule, samePasswordRule],
    };
    const steps = [
      {
        component: () => (
          <>
            <span>{t('您已选择 {count} 台实例，进行重置密码操作', { count: 5 })}：</span>
            <bk-table class='mt20' row-hover='auto' columns={columns} data={tableData.value} show-overflow-tooltip />
            <bk-form class='mt20' label-width='100' model={passwordForm.value} rules={passwordRules}>
              <bk-form-item label={t('用户名')}>
                <span>{passwordForm.value.username}</span>
              </bk-form-item>
              <bk-form-item label={t('新密码')} property='password' required>
                <bk-input v-model={passwordForm.value.password} type='password' />
              </bk-form-item>
              <bk-form-item label={t('确认密码')} property='confirmPassword' required>
                <bk-input v-model={passwordForm.value.confirmPassword} type='password' />
              </bk-form-item>
              <bk-form-item label={t('强制关机')}>
                <bk-checkbox v-model={passwordForm.value.shutdown}></bk-checkbox>
              </bk-form-item>
            </bk-form>
            <h3 class='password-tips'>
              {t('提示信息：当前操作需要实例在关机状态下进行')}：<br />
              {t('1. 为了避免数据丢失，实例将关机中断您的业务，请仔细确认')}
              <br />
              {t('2. 强制关机可能会导致数据丢失或文件系统损坏，您也可以主动关机后再进行操作')}
              <br />
              {t('3. 强制关机可能需要您等待较长时间，请耐心等待')}
            </h3>
          </>
        ),
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
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return (
      <>
        <step-dialog
          title={this.title}
          isShow={this.isShow}
          steps={this.steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
