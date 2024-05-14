import { PropType, defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import { Select, Tag } from 'bkui-vue';
import { useResourceStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';

const { Option } = Select;

export default defineComponent({
  props: {
    accountId: {
      required: true,
      type: Number,
    },
    type: {
      required: true,
      type: String as PropType<'CA' | 'SVR'>,
    },
    modelValue: String,
  },
  setup(props, { emit }) {
    const resourceStore = useResourceStore();
    const certIds = ref([]);
    const list = ref([]);
    const pagination = reactive({
      start: 0,
      limit: 50,
      hasNext: true,
    });
    const isLoading = ref(false);

    const getList = async () => {
      isLoading.value = true;
      try {
        const [detailRes, countRes] = await Promise.all(
          [false, true].map((isCount) =>
            resourceStore.list(
              {
                filter: {
                  op: QueryRuleOPEnum.AND,
                  rules: [
                    { field: 'cert_type', op: QueryRuleOPEnum.EQ, value: props.type },
                    {
                      field: 'account_id',
                      op: QueryRuleOPEnum.EQ,
                      value: props.accountId,
                    },
                  ],
                },
                page: {
                  count: isCount,
                  start: isCount ? 0 : pagination.start,
                  limit: isCount ? 0 : pagination.limit,
                  sort: isCount ? undefined : 'created_at',
                  order: isCount ? undefined : 'DESC',
                },
              },
              'certs',
            ),
          ),
        );
        list.value = [...list.value, ...detailRes.data.details];
        if (list.value.length >= countRes.data.count) {
          pagination.hasNext = false;
        } else {
          pagination.start += pagination.limit;
        }
      } finally {
        isLoading.value = false;
      }
    };

    const scrollToEnd = () => {
      if (pagination.hasNext) getList();
    };

    watch(
      () => props.accountId,
      () => getList(),
      {
        immediate: true,
      },
    );

    watch(
      () => certIds.value,
      (val) => emit('update:modelValue', val),
      {
        immediate: true,
      },
    );

    return () => (
      <Select
        v-model={certIds.value}
        multiple={props.type === 'SVR'}
        scrollLoading={isLoading.value}
        onScroll-end={scrollToEnd}>
        {list.value
          .sort((a, b) => a.cert_status - b.cert_status)
          .map(({ cloud_id, name, cert_status }) => (
            <Option key={cloud_id} id={cloud_id} name={name} disabled={cert_status === '3'}>
              {name}
              {cert_status === '3' && (
                <Tag theme='danger' style={{ marginLeft: '12px' }}>
                  已过期
                </Tag>
              )}
            </Option>
          ))}
      </Select>
    );
  },
});
