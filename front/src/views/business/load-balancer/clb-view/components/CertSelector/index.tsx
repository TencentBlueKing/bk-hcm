import { PropType, computed, defineComponent, ref, watch } from 'vue';
import { Select, Tag } from 'bkui-vue';
import { useSingleList } from '@/hooks/useSingleList';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { QueryRuleOPEnum } from '@/typings';

const { Option } = Select;

/**
 * 证书选择器
 * @prop modelValue - 证书 ID
 * @prop type - 证书类型 (SVR | CA)
 * @prop accountId - 云账户 ID
 */
export default defineComponent({
  name: 'CertSelector',
  props: {
    modelValue: [String, Array<String>] as PropType<String | Array<String>>,
    type: String as PropType<'SVR' | 'CA'>,
    accountId: String,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, expose }) {
    const { getBusinessApiPath } = useWhereAmI();

    const selected = ref<String | Array<String>>(props.modelValue);
    const isSVRCert = computed(() => props.type === 'SVR');
    const { dataList, isDataLoad, handleScrollEnd, handleRefresh } = useSingleList({
      url: `/api/v1/cloud/${getBusinessApiPath()}certs/list`,
      rules: [
        { field: 'cert_type', op: QueryRuleOPEnum.EQ, value: props.type },
        { field: 'account_id', op: QueryRuleOPEnum.EQ, value: props.accountId },
      ],
      immediate: true,
    });

    // select-handler - 选择证书后更新 props.modelValue
    watch(selected, (val) => emit('update:modelValue', val));

    // 对外暴露刷新列表的方法
    expose({ dataList, handleRefresh });

    return () => (
      <Select
        v-model={selected.value}
        multiple={isSVRCert.value}
        onScroll-end={handleScrollEnd}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {dataList.value
          .sort((a, b) => a.cert_status - b.cert_status)
          .map(({ cloud_id, name, cert_status, domain, encrypt_algorithm }) => (
            <Option key={cloud_id} id={cloud_id} name={name} disabled={cert_status === '3'}>
              {name}&nbsp;(主域名 : {domain ? domain[0] : '--'}, 备用域名：{domain ? domain[1] : '--'})
              {cert_status === '3' ? (
                <Tag theme='danger' style={{ marginLeft: '12px' }}>
                  已过期
                </Tag>
              ) : (
                <Tag theme='info' style={{ marginLeft: '12px' }}>
                  {encrypt_algorithm}
                </Tag>
              )}
            </Option>
          ))}
      </Select>
    );
  },
});
