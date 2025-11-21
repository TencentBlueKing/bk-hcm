<script setup lang="ts">
import { ref, useAttrs, watch } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useWhereAmI } from '@/hooks/useWhereAmI';

interface IProps {
  title?: string;
  dataInfo?: any;
}

const model = defineModel<boolean>();
const props = withDefaults(defineProps<IProps>(), {
  title: '回收预检详情',
  dataInfo: {},
});
const attrs = useAttrs();

const { getBusinessApiPath } = useWhereAmI();
const { columns } = useColumns('ExecutionRecords');
const requestParams = ref<any>({});
const { CommonTable, getListData } = useTable({
  tableOptions: {
    columns,
  },
  requestOption: {
    dataPath: 'data.info',
    immediate: false,
  },
  scrConfig: () => {
    return {
      payload: {
        ...requestParams.value,
      },
      url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/recycle/detect/step`,
    };
  },
});

watch(
  () => props.dataInfo,
  () => {
    requestParams.value = {
      bk_biz_id: props.dataInfo.bk_biz_id,
      ip: [props.dataInfo.ip],
      page: props.dataInfo.page,
    };
    if (props.dataInfo.suborderId) {
      requestParams.value.suborder_id = [props.dataInfo.suborderId];
    }
    getListData();
  },
  { deep: true, immediate: true },
);
</script>

<template>
  <bk-sideslider class="common-sideslider" v-bind="attrs" width="1000" v-model:is-show="model" :title="props.title">
    <div class="common-sideslider-content" style="padding: 24px">
      <div class="execute-record-top">IP : {{ props.dataInfo.ip }}</div>
      <CommonTable />
    </div>
  </bk-sideslider>
</template>
