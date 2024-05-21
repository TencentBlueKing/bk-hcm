import { defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import { Button, Dialog, Form, Input, Message } from 'bkui-vue';
// @ts-ignore
// import AppSelect from '@blueking/app-select';
// import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useSchemeStore } from '@/store';
import { debounce } from 'lodash-es';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';

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
    // const businessMapStore = useBusinessMapStore();
    const schemeStore = useSchemeStore();
    const formData = reactive({
      name: schemeStore.recommendationSchemes[props.idx].name,
      bk_biz_id: -1,
    });
    const formInstance = ref(null);
    const isNameDuplicate = ref(false);

    const handleConfirm = async () => {
      await checkNameIsDuplicate();
      if (isNameDuplicate.value) return;
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
      schemeStore.setRecommendationSchemes(
        schemeStore.recommendationSchemes.map((scheme, idx) => {
          if (idx === props.idx) {
            scheme.name = formData.name;
            scheme.isSaved = true;
          }
          return scheme;
        }),
      );
      schemeStore.setSchemeData({
        ...schemeStore.schemeData,
        name: formData.name,
      });
      isDialogShow.value = false;
    };

    watch(
      () => schemeStore.selectedSchemeIdx,
      (idx) => (formData.name = schemeStore.recommendationSchemes[idx].name),
    );

    const checkNameIsDuplicate = async () => {
      const filterQuery: QueryFilterType = {
        op: QueryRuleOPEnum.AND,
        rules: [
          {
            field: 'name',
            op: QueryRuleOPEnum.EQ,
            value: formData.name,
          },
          {
            field: 'bk_biz_id',
            op: QueryRuleOPEnum.EQ,
            value: formData.bk_biz_id,
          },
        ],
      };
      const pageQuery = {
        start: 0,
        limit: 1,
      };
      const res = await schemeStore.listCloudSelectionScheme(filterQuery, pageQuery);
      isNameDuplicate.value = !!res.data.details.length;
    };

    return () => (
      <>
        <Button
          theme='primary'
          onClick={() => (isDialogShow.value = true)}
          disabled={schemeStore.recommendationSchemes[props.idx].isSaved}>
          {schemeStore.recommendationSchemes[props.idx].isSaved ? '已保存' : '保存'}
        </Button>

        <Dialog
          title='保存该方案'
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={handleConfirm}>
          <Form
            formType='vertical'
            model={formData}
            ref={formInstance}
            rules={{
              name: [
                {
                  trigger: 'change',
                  message: '方案名称不能为空',
                  validator: (val: string) => val.trim().length,
                },
              ],
            }}>
            <FormItem label='方案名称' required property='name'>
              <Input v-model={formData.name} maxlength={28} onInput={debounce(checkNameIsDuplicate, 300)} />
              {isNameDuplicate.value ? (
                <span
                  style={{
                    color: '#ea3636',
                    fontSize: '12px',
                  }}>
                  方案名称与已存在的方案名重复
                </span>
              ) : null}
            </FormItem>
            {/* <FormItem label='标签' property='bk_biz_id'>
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
            </FormItem> */}
          </Form>
        </Dialog>
      </>
    );
  },
});
