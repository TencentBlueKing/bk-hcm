import { defineComponent, onMounted, onUnmounted, ref, PropType, reactive } from 'vue';
// import components
import { Form, Message } from 'bkui-vue';
import CommonSideslider from '@/components/common-sideslider';
// import stores
import { useAccountStore, useBusinessStore, useLoadBalancerStore } from '@/store';
// import custom hooks
import useAddOrUpdateTGForm from './useAddOrUpdateTGForm';
// import utils
import bus from '@/common/bus';

const { FormItem } = Form;

export default defineComponent({
  name: 'AddOrUpdateTGSideslider',
  props: {
    getListData: Function as PropType<(...args: any) => any>,
  },
  setup(props) {
    // use stores
    const accountStore = useAccountStore();
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();

    // 表单相关
    const getDefaultFormData = () => ({
      bk_biz_id: accountStore.bizs,
      account_id: '',
      name: '',
      protocol: '',
      port: 80,
      region: '',
      cloud_vpc_id: '',
      // rs_list: [] as any[],
    });
    const clearFormData = () => {
      Object.assign(formData, getDefaultFormData());
    };

    const isShow = ref(false);
    const isEdit = ref(false);
    const formData = reactive(getDefaultFormData());
    const { formItemOptions } = useAddOrUpdateTGForm(formData);

    // click-handler - 新建目标组
    const handleAddTargetGroup = () => {
      clearFormData();
      isEdit.value = false;
      loadBalancerStore.setCurrentScene('addTargetGroup');
      isShow.value = true;
    };

    // click-handler - 编辑目标组
    const handleEditTargetGroup = (data: any) => {
      Object.assign(formData, data);
      isEdit.value = true;
      loadBalancerStore.setCurrentScene('editTargetGroup');
      isShow.value = true;
    };

    // 更新选中的rs列表
    // const handleUpdateSelectedRsList = (data: any) => {
    //   formData.rs_list = data;
    // };

    // 处理参数
    const resolveFormData = () => ({ ...formData, port: +formData.port });
    // submit - 新增目标组
    const handleAddOrUpdateTargetGroupSubmit = async () => {
      const promise = isEdit.value
        ? businessStore.editTargetGroups(resolveFormData())
        : businessStore.createTargetGroups(resolveFormData());
      await promise;
      Message({
        message: isEdit.value ? '编辑成功' : '新建成功',
        theme: 'success',
      });
      isShow.value = false;
      props.getListData();
    };

    onMounted(() => {
      bus.$on('addTargetGroup', handleAddTargetGroup);
      bus.$on('editTargetGroup', handleEditTargetGroup);
      // bus.$on('updateSelectedRsList', handleUpdateSelectedRsList);
    });

    onUnmounted(() => {
      bus.$off('addTargetGroup');
      bus.$off('editTargetGroup');
      // bus.$off('updateSelectedRsList');
    });

    return () => (
      <CommonSideslider
        title={isEdit.value ? '编辑目标组' : '新建目标组'}
        width={960}
        v-model:isShow={isShow.value}
        onHandleSubmit={handleAddOrUpdateTargetGroupSubmit}>
        <bk-container margin={0}>
          <Form formType='vertical' model={formData}>
            {formItemOptions.value.map((item) => (
              <bk-row>
                {Array.isArray(item) ? (
                  item.map((subItem) => (
                    <bk-col span={subItem.span}>
                      <FormItem label={subItem.label} property={subItem.property} required={subItem.required}>
                        {subItem.content()}
                      </FormItem>
                    </bk-col>
                  ))
                ) : (
                  <bk-col span={item.span}>
                    <FormItem label={item.label} property={item.property} required={item.required}>
                      {item.content()}
                    </FormItem>
                  </bk-col>
                )}
              </bk-row>
            ))}
          </Form>
        </bk-container>
      </CommonSideslider>
    );
  },
});
