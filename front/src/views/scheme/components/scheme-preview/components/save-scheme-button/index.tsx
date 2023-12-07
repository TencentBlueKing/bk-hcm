import { defineComponent, reactive, ref } from 'vue';
import './index.scss';
import { Button, Dialog, Form, Input, Message } from 'bkui-vue';
// @ts-ignore
import AppSelect from '@blueking/app-select';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useSchemeStore } from '@/store';

const { FormItem } = Form;

export default defineComponent({
  props: {
    idx: {
      required: true,
      type: Number,
    },
  },
  setup(props) {
    const isDialogShow = ref(false);
    const isSaved = ref(false);
    const businessMapStore = useBusinessMapStore();
    const schemeStore = useSchemeStore();
    const formData = reactive({
      name: schemeStore.recommendationSchemes[props.idx].name,
      bk_biz_id: 0,
    });
    const formInstance = ref(null);

    const handleConfirm = async () => {
      await formInstance.value.validate();
      const saveData = {
        ...formData,
        user_distribution: schemeStore.userDistribution,
        cover_rate: schemeStore.recommendationSchemes[props.idx].cover_rate,
        composite_score: schemeStore.recommendationSchemes[props.idx].composite_score,
        net_score: schemeStore.recommendationSchemes[props.idx].net_score,
        cost_score: schemeStore.recommendationSchemes[props.idx].cost_score,
        result_idc_ids: schemeStore.recommendationSchemes[props.idx].result_idc_ids,
        cover_ping: schemeStore.schemeConfig.cover_ping,
        biz_type: schemeStore.schemeConfig.biz_type,
        deployment_architecture: schemeStore.schemeConfig.deployment_architecture,
      };
      await schemeStore.createScheme(saveData);
      Message({
        theme: 'success',
        message: '保存成功',
      });
      schemeStore.setRecommendationSchemes(schemeStore.recommendationSchemes.map((scheme, idx) => {
        if (idx === props.idx) scheme.name = formData.name;
        return scheme;
      }));
      isDialogShow.value = false;
      isSaved.value = true;
    };
    return () => (
      <>
        <Button
          theme='primary'
          onClick={() => (isDialogShow.value = true)}
          disabled={isSaved.value}>
          {isSaved.value ? '已保存' : '保存'}
        </Button>

        <Dialog
          title='保存该方案'
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={handleConfirm}>
          <Form formType='vertical' model={formData} ref={formInstance}>
            <FormItem label='方案名称' required property='name'>
              <Input v-model={formData.name} maxlength={28}/>
            </FormItem>
            <FormItem label='标签' property='bk_biz_id'>
              <AppSelect
                data={businessMapStore.businessList}
                value={{
                  id: formData.bk_biz_id,
                }}
                onChange={
                  (val: {id: number, val: string}) => {
                    formData.bk_biz_id = val.id;
                  }
                }
              />
            </FormItem>
          </Form>
        </Dialog>
      </>
    );
  },
});
