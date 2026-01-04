import { computed, defineComponent, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { Button, Dropdown, Message } from 'bkui-vue';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import ExportToExcelButton from '@/components/export-to-excel-button';
import { useI18n } from 'vue-i18n';
import { useTable } from '@/hooks/useTable/useTable';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import http from '@/http';
import { getEntirePath } from '@/utils';
import { getRecycleStageOpts } from '@/api/host/recycle';
import { useUserStore } from '@/store';
import { getDateRange, transformFlatCondition } from '@/utils/search';
import type { ModelProperty } from '@/model/typings';
import { getModel } from '@/model/manager';
import HocSearch from '@/model/hoc-search.vue';
import { HostRecycleSearch } from '@/model/order/host-recycle-search';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

export default defineComponent({
  setup() {
    const router = useRouter();
    const userStore = useUserStore();
    const { getBusinessApiPath } = useWhereAmI();
    const { t } = useI18n();
    const route = useRoute();

    const currentOperateRowIndex = ref(-1);

    const searchFields = getModel(HostRecycleSearch).getProperties();
    const searchQs = useSearchQs({ key: 'filter', properties: searchFields });

    const condition = ref<Record<string, any>>({});
    const searchValues = ref<Record<string, any>>({});

    const opBtnDisabled = computed(() => {
      return (status: any) =>
        ['UNCOMMIT', 'COMMITTED', 'DETECTING', 'FOR_AUDIT', 'TRANSITING', 'RETURNING', 'DONE', 'TERMINATE'].includes(
          status,
        );
    });

    const { columns } = useScrColumns('hostRecycleApplication');
    columns.splice(1, 0, {
      label: t('单号/子单号'),
      width: 100,
      render: ({ row }: any) => {
        return (
          <>
            <p>{row.order_id}</p>
            <div>
              <Button theme='primary' text onClick={() => enterDetail(row)}>
                {row.suborder_id}
              </Button>
            </div>
          </>
        );
      },
      exportFormatter: (data: any) => `${data.order_id}/${data.suborder_id}`,
    });
    const { selections, handleSelectionChange } = useSelection();
    const { CommonTable, getListData, dataList } = useTable({
      tableOptions: {
        columns: [
          ...columns,
          {
            label: t('操作'),
            width: 120,
            fixed: 'right',
            render: ({ row, index }: any) => (
              <div class={cssModule['operation-column']}>
                <Button text theme='primary' onClick={() => returnPreDetails(row)}>
                  {t('预检详情')}
                </Button>
                <Dropdown
                  trigger='click'
                  popoverOptions={{
                    renderType: 'shown',
                    onAfterShow: () => (currentOperateRowIndex.value = index),
                    onAfterHidden: () => (currentOperateRowIndex.value = -1),
                  }}>
                  {{
                    default: () => (
                      <div
                        class={[
                          cssModule['more-action'],
                          { [cssModule['current-operate-row']]: currentOperateRowIndex.value === index },
                        ]}>
                        <i class='hcm-icon bkhcm-icon-more-fill'></i>
                      </div>
                    ),
                    content: () => (
                      <Dropdown.DropdownMenu>
                        <Dropdown.DropdownItem
                          key='retry'
                          class={[
                            cssModule['more-action-item'],
                            { [cssModule.disabled]: opBtnDisabled.value(row.status) },
                          ]}
                          onClick={() => retryOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}>
                          {t('全部重试')}
                        </Dropdown.DropdownItem>
                        <Dropdown.DropdownItem
                          key='stop'
                          class={[
                            cssModule['more-action-item'],
                            { [cssModule.disabled]: opBtnDisabled.value(row.status) },
                          ]}
                          onClick={() => stopOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}>
                          {t('全部终止')}
                        </Dropdown.DropdownItem>
                        <Dropdown.DropdownItem
                          key='submit'
                          class={[
                            cssModule['more-action-item'],
                            { [cssModule.disabled]: opBtnDisabled.value(row.status) },
                          ]}
                          onClick={() => submitOrderFunc(row.suborder_id, opBtnDisabled.value(row.status))}>
                          {t('剔除预检失败IP重试')}
                        </Dropdown.DropdownItem>
                      </Dropdown.DropdownMenu>
                    ),
                  }}
                </Dropdown>
              </div>
            ),
          },
        ],
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: { sort: 'create_at', order: 'DESC' },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/recycle/order`,
          payload: transformFlatCondition(condition.value, searchFields),
        };
      },
    });

    const getSearchCompProps = (field: ModelProperty) => {
      if (field.id === 'create_at') {
        return {
          type: 'daterange',
          format: 'yyyy-MM-dd',
        };
      }
      if (field.id === 'order_id') {
        return {
          pasteFn: (value: string) =>
            value
              .split(/\r\n|\n|\r/)
              .filter((tag) => /^\d+$/.test(tag))
              .map((tag) => ({ id: tag, name: tag })),
        };
      }
      if (field.id === 'suborder_id') {
        return {
          pasteFn: (value: string) => value.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag })),
        };
      }
      if (field.id === 'stage') {
        const stages = stageList.value.reduce((acc, { stage, description }) => {
          acc[stage] = description;
          return acc;
        }, {});
        return {
          option: stages,
        };
      }
      if (field.id === 'recycle_type') {
        return {
          useNameValue: true,
        };
      }
      return {
        option: field.option,
      };
    };

    const handleSearch = () => {
      searchQs.set(searchValues.value);
    };

    const handleReset = () => {
      searchQs.clear();
    };

    watch(
      () => route.query,
      async (query) => {
        condition.value = searchQs.get(query, {
          create_at: getDateRange('last30d', true),
          bk_username: [userStore.username],
        });

        searchValues.value = condition.value;

        getListData();
      },
      { immediate: true },
    );

    const enterDetail = (row: any) => {
      router.push({
        name: 'HostRecycleDocDetail',
        query: { [GLOBAL_BIZS_KEY]: row.bk_biz_id, suborderId: row.suborder_id, bkBizId: row.bk_biz_id },
      });
    };
    const returnPreDetails = (row: any) => {
      router.push({ name: 'HostRecyclePreDetail', query: { ...route.query, suborder_id: row.suborder_id } });
    };
    const getBatchSuborderId = () => selections.value.map((item) => item.suborder_id);
    const goToPrecheck = () => {
      router.push({
        name: 'HostRecyclePreDetail',
        query: { ...route.query, suborder_id: getBatchSuborderId().join('\n') },
      });
    };
    const textTip = (text: string, theme: 'error' | 'success') => {
      const themeDes = { error: t('失败'), success: t('成功') };
      Message({ message: `${text}${themeDes[theme]}`, theme, duration: 1500 });
    };
    const retryOrderFunc = async (id: string, disabled: boolean) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      const res = await http.post(getEntirePath(`${getBusinessApiPath()}task/start/recycle/order`), {
        suborder_id: suborderId,
      });

      if (res.code === 0) {
        textTip(t('重试'), 'success');
        getListData();
      }
    };
    const stopOrderFunc = async (id: string, disabled: boolean) => {
      if (disabled) return;
      const res = await http.post(getEntirePath(`${getBusinessApiPath()}task/terminate/recycle/order`), {
        suborder_id: [id],
      });

      if (res.code === 0) {
        textTip(t('终止'), 'success');
        getListData();
      }
    };
    const submitOrderFunc = async (id: string, disabled: boolean) => {
      if (disabled) return;
      const suborderId = id === 'isBatch' ? getBatchSuborderId() : [id];
      const res = await http.post(getEntirePath(`${getBusinessApiPath()}task/revise/recycle/order`), {
        suborder_id: suborderId,
      });

      if (res.code === 0) {
        textTip(t('去除预检失败IP提交'), 'success');
        getListData();
      }
    };

    const stageList = ref([]);
    const fetchStageList = async () => {
      const data = await getRecycleStageOpts();
      stageList.value = data?.info || [];
    };

    onMounted(() => {
      fetchStageList();
    });

    return () => (
      <>
        <div style={{ padding: '24px 24px 0 24px' }}>
          <GridContainer layout='vertical' column={4} content-min-width={'1fr'} gap={[16, 60]}>
            {searchFields.map((field) => (
              <GridItemFormElement key={field.id} label={field.name}>
                <HocSearch
                  is={field.type}
                  display={field.meta?.display}
                  v-model={searchValues.value[field.id]}
                  {...getSearchCompProps(field)}
                />
              </GridItemFormElement>
            ))}
            <GridItem span={4}>
              <div style={{ display: 'flex', gap: '8px' }}>
                <bk-button theme='primary' style={{ minWidth: '86px' }} onClick={handleSearch}>
                  查询
                </bk-button>
                <bk-button style={{ minWidth: '86px' }} onClick={handleReset}>
                  重置
                </bk-button>
              </div>
            </GridItem>
          </GridContainer>
        </div>
        <section class={cssModule['table-wrapper']}>
          <div class={[cssModule.buttons, cssModule.mb16]}>
            <ExportToExcelButton
              class={cssModule.button}
              data={dataList.value}
              columns={columns}
              text={t('导出当页')}
              filename={t('回收单据列表')}
            />
            <Button class={cssModule.button} disabled={!selections.value.length} onClick={goToPrecheck}>
              {t('批量查看预检详情')}
            </Button>
            <Button
              class={cssModule.button}
              disabled={!selections.value.length}
              onClick={() => retryOrderFunc('isBatch', false)}>
              {t('批量重试')}
            </Button>
            <Button
              class={cssModule.button}
              disabled={!selections.value.length}
              onClick={() => submitOrderFunc('isBatch', false)}>
              {t('批量去除预检失败IP提交')}
            </Button>
          </div>
          <CommonTable style={{ height: 'calc(100% - 48px)' }} />
        </section>
      </>
    );
  },
});
