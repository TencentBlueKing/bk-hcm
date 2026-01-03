import { defineComponent, computed, onMounted } from 'vue';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { transferSimpleConditions } from '@/utils/scr/simple-query-builder';
import { Button, Form, TagInput } from 'bkui-vue';
import useFormModel from '@/hooks/useFormModel';
import { useBusinessGlobalStore } from '@/store/business-global';
import { timeFormatter, applicationTime, isEmpty } from '@/common/util';
import ExportToExcelBatchButton from '@/components/export-to-excel-batch-button/index.vue';
import RequirementTypeSelector from '@/components/scr/requirement-type-selector';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import HcmSearchBusiness from '@/components/search/business.vue';
import { serviceShareBizSelectedKey } from '@/constants/storage-symbols';
import rollRequest from '@blueking/roll-request';
import http from '@/http';

const { FormItem } = Form;
export default defineComponent({
  setup() {
    const { columns } = useColumns('hostApplyDevice');
    const { selections, handleSelectionChange } = useSelection();
    const clipHostIp = computed(() => {
      return selections.value.map((item) => item.ip).join('\n');
    });
    const clipHostAssetId = computed(() => {
      return selections.value.map((item) => item.asset_id).join('\n');
    });
    const businessGlobalStore = useBusinessGlobalStore();

    const { formModel, resetForm } = useFormModel({
      orderId: '',
      bkBizId: businessGlobalStore.getCacheSelected(serviceShareBizSelectedKey) ?? [0],
      bkUsername: [],
      ip: '',
      requireType: '',
      suborderId: '',
      dateRange: applicationTime(),
      assetId: [],
    });

    // 构建查询条件的函数
    const buildFilterPayload = () => ({
      filter: transferSimpleConditions([
        'AND',
        [
          'bk_biz_id',
          'in',
          formModel.bkBizId?.[0] === 0 || isEmpty(formModel.bkBizId)
            ? businessGlobalStore.businessAuthorizedList.map((item: any) => item.id)
            : formModel.bkBizId,
        ],
        ['require_type', '=', formModel.requireType],
        ['order_id', '=', formModel.orderId],
        ['suborder_id', '=', formModel.suborderId],
        ['bk_username', 'in', formModel.bkUsername],
        ['ip', 'in', formModel.ip],
        ['update_at', 'd>=', timeFormatter(formModel.dateRange[0], 'YYYY-MM-DD')],
        ['update_at', 'd<=', timeFormatter(formModel.dateRange[1], 'YYYY-MM-DD')],
        ['asset_id', 'in', formModel.assetId],
      ]),
    });

    const { CommonTable, getListData, isLoading, pagination } = useTable({
      tableOptions: {
        columns,
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
        sortOption: {
          sort: 'create_at',
          order: 'DESC',
        },
        immediate: false,
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/task/findmany/apply/device',
          payload: buildFilterPayload(),
        };
      },
    });

    const filterOrders = () => {
      pagination.start = 0;
      getListData();
    };

    // 导出全部的请求函数
    const exportAllRequest = async (signal: AbortSignal) => {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'enable_count',
      }).rollReqUseTotalCount(
        '/api/v1/woa/task/findmany/apply/device',
        {
          ...buildFilterPayload(),
        },
        {
          limit: 5000,
          total: pagination.count,
          listGetter: (res: { data: { info: any[] } }) => res.data.info,
          countGetter: (res: { data: { count: number } }) => res.data.count,
        },
        {
          signal,
        },
      );

      return list;
    };

    onMounted(() => {
      getListData();
    });

    return () => (
      <div class={'apply-list-container'}>
        <div class={'filter-container'}>
          <Form label-width='110' formType='vertical' class='scr-form-wrapper' model={formModel}>
            <FormItem label='业务'>
              <HcmSearchBusiness
                v-model={formModel.bkBizId}
                showAll
                {...{ scope: 'auth', emptySelectAll: true, cacheKey: serviceShareBizSelectedKey }}
              />
            </FormItem>
            <FormItem label='需求类型'>
              <RequirementTypeSelector v-model={formModel.requireType} />
            </FormItem>
            <FormItem label='单号'>
              <bk-input v-model={formModel.orderId} clearable type='number' placeholder='请输入单号'></bk-input>
            </FormItem>
            <FormItem label='申请人'>
              <hcm-form-user v-model={formModel.bkUsername} />
            </FormItem>
            <FormItem label='交付时间'>
              <bk-date-picker type='daterange' v-model={formModel.dateRange} clearable={false} />
            </FormItem>
            <FormItem label='内网 IP'>
              {/* <Input
                class={'filte-item'}
                type='textarea'
                clearable
                placeholder='请输入IP地址，多行换行分割'
                v-model={formModel.ip}
                autosize
                resize={false}
              /> */}
              <TagInput
                v-model={formModel.ip}
                allow-create
                collapse-tags
                allow-auto-match
                pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
                createTagValidator={(ip) =>
                  /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/.test(ip)
                }
                placeholder='输入合法的 IP 地址'
              />
            </FormItem>
            <FormItem label='固资号'>
              <TagInput
                v-model={formModel.assetId}
                allow-create
                collapse-tags
                allow-auto-match
                pasteFn={(v) => v.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }))}
                placeholder='请输入固资号'
              />
            </FormItem>
          </Form>
          <div class='btn-container'>
            <Button theme='primary' native-type='submit' loading={isLoading.value} onClick={filterOrders}>
              查询
            </Button>
            <Button
              onClick={() => {
                resetForm();
                filterOrders();
              }}>
              重置
            </Button>
          </div>
        </div>
        <div class='btn-container oper-btn-pad'>
          <ExportToExcelBatchButton
            showConfirmDialog
            data={selections.value}
            columns={columns}
            filename='设备列表'
            text='导出勾选'
            name='申领主机'
            disabled={selections.value.length === 0}
          />
          <ExportToExcelBatchButton
            showConfirmDialog
            request={exportAllRequest}
            columns={columns}
            filename='设备列表'
            text='导出全部'
            name='申领主机'
            pickNum={pagination.count}
            disabled={pagination.count === 0}
          />

          <Button v-clipboard={clipHostIp.value} disabled={selections.value.length === 0}>
            复制IP
          </Button>
          <Button v-clipboard={clipHostAssetId.value} disabled={selections.value.length === 0}>
            复制固单号
          </Button>
        </div>
        <CommonTable />
      </div>
    );
  },
});
