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
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
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
const fetchUrl = ref<string>(`vendors/${props.data.vendor}/security_groups/${securityId.value}/rules/list`);
const fetchFilter = ref<any>();
const isListLoading = ref(false);

const state = reactive<any>({
  datas: [],
  pagination: {
    current: 1,
    limit: 10,
    count: 0,
  },
  isLoading: false,
  handlePageChange: () => {},
  handlePageSizeChange: () => {},
  handleSort: () => {},
  columns: useColumns('securityCommon', false, props.data.vendor),
});

// use hook
const {
  t,
} = useI18n();
const resourceStore = useResourceStore();


watch(() => activeType.value, (val) => {
  fetchFilter.value.filter.rules[0].value = val;
  state.columns.forEach((e: any) => {
    if (e.field === 'resource') {
      e.label = val === 'ingress' ? t('来源') : t('目标');
    }
  });
});


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
  isListLoading.value = true;
  try {
    const res = await resourceStore.getSecurityGroupsListByCvmId(props.data.id);
    tableData.value = res.data;
  } catch (error) {
    console.log(error);
  } finally {
    isListLoading.value = false;
  }
};

const showRuleDialog = async () => {
  isShow.value = true;
  // 获取列表数据
  fetchUrl.value = `vendors/${props.data.vendor}/security_groups/${securityId.value}/rules/list`;
  fetchFilter.value = { filter: { op: 'and', rules: [{ field: 'type', op: 'eq', value: activeType.value }] } };
  const {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
    getList,
  } = useQueryCommonList(fetchFilter.value, fetchUrl);

  state.datas = datas;
  state.isLoading = isLoading;
  state.pagination = pagination;
  state.handlePageChange = handlePageChange;
  state.handlePageSizeChange = handlePageSizeChange;
  state.handleSort = handleSort;
  state.getList = getList;
  state.columns = useColumns('securityCommon', false, props.data.vendor);

  if (props.data.vendor === 'huawei') {
    const huaweiColummns = [{
      label: t('优先级'),
      field: 'priority',
    }, {
      label: t('类型'),
      field: 'ethertype',
    }];
    state.columns.unshift(...huaweiColummns);
  } else if (props.data.vendor === 'azure') {
    const awsColummns = [{
      label: t('优先级'),
      field: 'priority',
    }, {
      label: t('名称'),
      field: 'name',
    }];
    state.columns.unshift(...awsColummns);
  }
  console.log('state.columns', state.columns);
};

getSecurityGroupsList();
</script>

<template>
  <bk-loading
    :loading="isListLoading"
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
