<script lang="ts" setup>
import {
  ref,
  PropType,
  reactive,
  h,
  watch,
} from 'vue';
import {
  Button,
} from 'bkui-vue';
import {
  useI18n,
} from 'vue-i18n';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import {
  useResourceStore,
} from '@/store/resource';

const props = defineProps({
  data: {
    type: Object as PropType<any>,
  },
});

const activeType = ref('ingress');
const tableData = ref([]);
const isShow = ref(false);
const securityId = ref(0);
const isLoading = ref(false);

const state = reactive<any>({
  datas: [],
  pagination: {
    current: 1,
    limit: 10,
    count: 0,
  },
  isLoading: true,
  handlePageChange: () => {},
  handlePageSizeChange: () => {},
  handleSort: () => {},
  columns: useColumns('securityCommon'),
});

// use hook
const {
  t,
} = useI18n();
const resourceStore = useResourceStore();

// 获取列表数据
const fetchList = (fetchType: string) => {
  console.log(111234);
  const {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  } = useQueryList({ filter: { op: 'and', rules: [{ field: 'type', op: 'eq', value: activeType.value }] } }, fetchType);
  console.log('datas', datas);
  return {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
};

// 切换tab
const handleSwtichType = async () => {
  const params = {
    fetchUrl: `vendors/${props.data.vendor}/security_groups/${securityId.value}/rules`,
  };
  console.log('params', params);
  // eslint-disable-next-line max-len
  const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = fetchList(params.fetchUrl);
  state.datas = datas;
  state.isLoading = isLoading;
  state.pagination = pagination;
  state.handlePageChange = handlePageChange;
  state.handlePageSizeChange = handlePageSizeChange;
  state.handleSort = handleSort;
};

watch(
  () => activeType.value,
  handleSwtichType,
);

const columns = [
  {
    label: 'ID',
    field: 'id',
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            console.log(233);
            securityId.value = data.id;
            showRuleDialog();
          },
        },
        [
          data.id || '--',
        ],
      );
    },
  },
  {
    label: '名称',
    field: 'name',
  },
];

// tab 信息
const types = [
  { name: 'ingress', label: t('入站规则') },
  { name: 'egress', label: t('出站规则') },
];

const getSecurityGroupsList = async () => {
  isLoading.value = true;
  try {
    const res = await resourceStore.getSecurityGroupsListByCvmId(props.data.id);
    tableData.value = res.data;
  } catch (error) {
    console.log(error);
  } finally {
    isLoading.value = false;
  }
};

const showRuleDialog = async () => {
  handleSwtichType();
  isShow.value = true;
};

getSecurityGroupsList();
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <bk-table
      class="mt20"
      row-hover="auto"
      :columns="columns"
      :data="tableData"
    />
  </bk-loading>
  <bk-dialog
    v-model:isShow="isShow"
    :title="activeType === 'ingress' ? '入站规则' : '出站规则'"
    width="1200"
    :theme="'primary'"
    :quick-close="false"
    :dialog-type="'show'">

    <bk-loading
      :loading="state.isLoading"
    >
      <section class="mt20">
        <bk-radio-group
          v-model="activeType"
        >
          <bk-radio-button
            v-for="item in types"
            :key="item.name"
            :label="item.name"
          >
            {{ item.label }}
          </bk-radio-button>
        </bk-radio-group>

      </section>
      <bk-table
        class="mt20"
        row-hover="auto"
        :columns="state.columns"
        :data="state.datas"
        remote-pagination
        :pagination="state.pagination"
        @page-limit-change="state.handlePageSizeChange"
        @page-value-change="state.handlePageChange"
        @column-sort="state.handleSort"
      />
    </bk-loading>
  </bk-dialog>
</template>

<style lang="scss" scoped>
  .security-head {
    display: flex;
    align-items: center;
  }
</style>
