import { defineComponent, onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import { Dialog, Form, Input, Message, Select } from 'bkui-vue';
// import components
import MemberSelect from '@/components/MemberSelect';
// import stores
import { useBusinessMapStore } from '@/store/useBusinessMap';
// import hooks
import { useI18n } from 'vue-i18n';
// import utils
import http from '@/http';
import bus from '@/common/bus';
// import constants
import { QueryRuleOPEnum } from '@/typings';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { FormItem } = Form;

export default (getListData: Function) => {
  // use hooks
  const route = useRoute();
  const { t } = useI18n();
  // use stores
  const businessMapStore = useBusinessMapStore();
  // define data
  const isUserDialogLoading = ref(false);
  const formRef = ref<InstanceType<typeof Form>>(null);
  // define data
  const isShowModifyUserDialog = ref(false);
  const userFormModel = reactive({
    bk_biz_ids: [],
    managers: [],
    memo: '',
    id: '',
  });
  // define function
  const clearUserFormParams = () => {
    Object.assign(userFormModel, {
      bk_biz_ids: [],
      managers: [],
      memo: '',
      id: '',
    });
  };
  const handleModifyAccount = (data: any) => {
    clearUserFormParams();
    isShowModifyUserDialog.value = true;
    Object.assign(userFormModel, {
      bk_biz_ids: data?.bk_biz_ids,
      managers: data?.managers,
      memo: data?.memo,
      id: data?.id,
    });
  };

  const handleModifyUserSubmit = async () => {
    await formRef.value.validate();
    try {
      isUserDialogLoading.value = true;
      await http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/sub_accounts/${userFormModel.id}`, userFormModel);
      Message({
        theme: 'success',
        message: t('编辑成功'),
      });
      isShowModifyUserDialog.value = false;
      getListData([{ op: QueryRuleOPEnum.EQ, field: 'account_id', value: route.query.accountId }]);
    } finally {
      isUserDialogLoading.value = false;
    }
  };

  const RenderForm = defineComponent({
    name: 'ModifyUserDialog',
    setup() {
      return () => (
        <Dialog
          isShow={isShowModifyUserDialog.value}
          width={680}
          title={t('编辑用户')}
          isLoading={isUserDialogLoading.value}
          onConfirm={handleModifyUserSubmit}
          onClosed={() => (isShowModifyUserDialog.value = false)}
          theme='primary'>
          <Form v-model={userFormModel} formType='vertical' ref={formRef}>
            <FormItem label={t('所属业务')} class={'api-secret-selector'} property='bk_biz_ids'>
              <Select v-model={userFormModel.bk_biz_ids} showSelectAll multiple multipleMode='tag' collapseTags>
                {businessMapStore.businessList.map((businessItem) => {
                  return (
                    <bk-option key={businessItem.id} value={businessItem.id} label={businessItem.name}></bk-option>
                  );
                })}
              </Select>
            </FormItem>
            <FormItem label={t('负责人')} class={'api-secret-selector'} property='managers'>
              <MemberSelect v-model={userFormModel.managers} />
            </FormItem>
            <FormItem label={t('备注')}>
              <Input type={'textarea'} v-model={userFormModel.memo} maxlength={256} resize={false} />
            </FormItem>
          </Form>
        </Dialog>
      );
    },
  });

  onMounted(() => {
    bus.$on('handleModifyAccount', handleModifyAccount);
  });

  onUnmounted(() => {
    bus.$off('handleModifyAccount');
  });

  return {
    RenderForm,
  };
};
