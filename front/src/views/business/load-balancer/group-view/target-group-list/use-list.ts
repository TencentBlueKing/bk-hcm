import { Ref, ref } from 'vue';
import http from '@/http';
import rollRequest from '@blueking/roll-request';
import { QueryRuleOPEnum } from '@/typings';

export const useList = (url: string, rules: Ref<any[]>) => {
  const isDataLoad = ref(false);
  const isScrollLoading = ref(false);

  const totalCount = ref(0);
  const dataList = ref([]);

  // 生成器的引用
  let dataListGen: any = null;

  const getData = async () => {
    isDataLoad.value = true;

    // 获取总数
    const countRes = await http.post(url, {
      filter: {
        op: QueryRuleOPEnum.AND,
        rules: rules.value,
      },
      page: {
        count: true,
      },
    });
    totalCount.value = countRes.data.count;

    // 获取数据的生成器，默认获取第1页数据，后续通过getNextData获取下一页数据
    dataListGen = await rollRequest({
      httpClient: http,
      pageEnableCountKey: 'count',
    }).rollReqUseCount(
      url,
      {
        filter: {
          op: QueryRuleOPEnum.AND,
          rules: rules.value,
        },
      },
      { limit: 500, countGetter: (res) => res.data.count, listGetter: (res) => res.data.details },
      true,
    );
    const { value, done } = dataListGen.next();
    if (!done) {
      dataList.value = (await value)?.data?.details;
    }

    isDataLoad.value = false;
  };

  const getNextData = async () => {
    if (!dataListGen) return;
    isScrollLoading.value = true;
    const { value, done } = dataListGen.next();
    if (!done) {
      dataList.value = [...dataList.value, ...(await value)?.data?.details];
    }
    isScrollLoading.value = false;
  };

  const handleRefresh = () => {
    getData();
  };

  // 默认执行
  getData();

  return {
    dataList,
    totalCount,
    isDataLoad,
    isScrollLoading,
    handleRefresh,
    getNextData,
  };
};
