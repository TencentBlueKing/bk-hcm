import { defineComponent, ref, computed, onMounted, watch } from 'vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { getRequireTypes } from '@/api/host/task';
import { getCvmProduceOrderList, getCvmProducedResources } from '@/api/host/cvm';
import MemberSelect from '@/components/MemberSelect';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import FastCvmProduce from './component/fast-cvm-produce';
import CreateOrderDialog from './component/create-order/index.vue';
import { useUserStore } from '@/store';
import SuccessProduceDetail from './component/success-produce-detail';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { Button, Form, Select } from 'bkui-vue';
import { Copy, Plus, Search } from 'bkui-vue/lib/icon';
import FloatInput from '@/components/float-input';
import { statusList } from './transform';
import { merge, throttle } from 'lodash';
import dayjs from 'dayjs';
import './index.scss';
import { ICvmDeviceDetailItem } from '@/typings/ziyanScr';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import PropertyList from '@/views/ziyanScr/cvm-produce/component/property-display/property-list';

const { FormItem } = Form;

export default defineComponent({
  components: {
    MemberSelect,
    AreaSelector,
    ZoneSelector,
    FastCvmProduce,
    CreateOrderDialog,
    SuccessProduceDetail,
    FloatInput,
  },
  setup() {
    const userStore = useUserStore();
    const defaultCvmProduceForm = () => ({
      require_type: [],
      region: [],
      zone: [],
      device_type: [],
      order_id: [],
      task_id: [],
      status: [],
      bk_username: [userStore.username],
    });
    const defaultTime = () => [new Date(dayjs().subtract(1, 'week').format('YYYY-MM-DD')), new Date()];
    const cvmProduceForm = ref(defaultCvmProduceForm());
    const timeForm = ref(defaultTime());
    const handleTime = (time) => (!time ? '' : dayjs(time).format('YYYY-MM-DD'));

    const timeObj = computed(() => {
      return {
        start: handleTime(timeForm.value[0]),
        end: handleTime(timeForm.value[1]),
      };
    });
    const pageInfo = ref({
      start: 0,
      limit: 10,
    });
    const requestListParams = ref({
      ...timeObj.value,
      bk_username: [userStore.username],
      page: pageInfo.value,
    });

    watch(
      () => userStore.username,
      (username) => {
        cvmProduceForm.value.bk_username = [username];
        requestListParams.value.bk_username = [username];
      },
    );

    const loadOrders = () => {
      const params = {
        ...cvmProduceForm.value,
        ...timeObj.value,
        page: pageInfo.value,
      };
      params.order_id = params.order_id.map((item) => Number(item));
      requestListParams.value = { ...params };
      getListData();
    };
    const filterOrders = () => {
      pageInfo.value.start = 0;
      loadOrders();
    };
    const clearFilter = () => {
      cvmProduceForm.value = defaultCvmProduceForm();
      timeForm.value = defaultTime();
      filterOrders();
    };
    const orderClipboard = ref({});
    const isShowProduceDetail = ref(false);
    const handleCheckSuccessNum = () => {
      isShowProduceDetail.value = true;
    };
    const { columns } = useColumns('cvmProduceQuery');
    columns.splice(14, 0, {
      label: '生产情况-成功',
      field: 'success_num',
      width: 120,
      fixed: 'right',
      render: ({ row }) => {
        if (row.success_num > 0) {
          const { order_id } = row;
          const ips = orderClipboard.value[order_id]?.ips || [];
          const assetIds = orderClipboard.value[order_id]?.assetIds || [];

          return (
            <div class='success-container'>
              <div>
                <Button text theme='primary' onClick={handleCheckSuccessNum}>
                  {row.success_num}
                </Button>
              </div>
              <div v-bk-tooltips={{ placement: 'top', content: '复制 IP' }}>
                <Button text theme='primary' v-clipboard={ips.join('\n')}>
                  <Copy />
                </Button>
              </div>
              <div v-bk-tooltips={{ placement: 'top', content: '复制固资号' }}>
                <Button text theme='primary' v-clipboard={assetIds.join('\n')}>
                  <Copy />
                </Button>
              </div>
            </div>
          );
        }
        return <span>{row.success_num}</span>;
      },
    });
    const tableColumns = [...columns];
    const { CommonTable, getListData, dataList, pagination, sort, order } = useTable({
      tableOptions: {
        columns: tableColumns,
        extra: {
          onRowMouseEnter: (e, row) => {
            handleCellMouseEnter(row);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          sort: 'create_at',
          order: 'DESC',
        },
      },
      scrConfig: () => {
        return {
          url: '/api/v1/woa/cvm/findmany/apply/order',
          payload: {
            ...requestListParams.value,
          },
        };
      },
    });

    const isCreateCvmDialogShow = ref(false);
    const handleCreateCvmConfirmSuccess = () => {
      pageInfo.value.start = 0;
      loadOrders();
    };
    const handleCreateCvmClosed = () => {
      fastProduceData.value = null;
    };

    const fastProduceData = ref(null);
    const handleOrderCreate = (resource: ICvmDeviceDetailItem) => {
      isCreateCvmDialogShow.value = true;
      if (resource) {
        fastProduceData.value = resource;
      }
    };
    const isFastProVisible = ref(false);
    const handleFastCvmProduce = () => {
      isFastProVisible.value = true;
    };
    // 需求类型
    const requireTypeList = ref([]);
    const fetchRequireType = async () => {
      const res = await getRequireTypes();
      requireTypeList.value = res.data.info.map((item) => ({
        label: item.require_name,
        value: item.require_type,
      }));
    };
    const throttleLoadHostInfo = ref(null);
    const loadProducedResources = (orderId) => {
      return getCvmProducedResources({ order_id: orderId });
    };
    const producedDetail = ref([]);
    const loadCvmProduceDetail = () => {
      throttleLoadHostInfo.value = throttle(
        async (row) => {
          const res = await loadProducedResources(row.order_id);
          const ips = res.data.info.map((item) => item.ip);
          const assetIds = res.data.info.map((item) => item.asset_id);
          orderClipboard.value[row.order_id] = {
            ips,
            assetIds,
          };
          producedDetail.value = res?.data?.info || [];
        },
        500,
        { trailing: true },
      );
    };
    const handleCellMouseEnter = (row) => {
      if (row.success_num > 0) {
        throttleLoadHostInfo.value(row);
      }
    };
    const pollProduceOrderList = () => {
      const newPage = {
        start: pagination.start,
        limit: pagination.limit,
        sort: `${sort.value}:${order.value === 'ASC' ? 1 : -1}`,
      };
      const params = Object.assign(requestListParams.value, { page: newPage });
      getCvmProduceOrderList(params).then((res) => {
        dataList.value.forEach((currentOrder) => {
          const newOrder = res?.data?.info?.find((item) => item.order_id === currentOrder.order_id) || null;
          if (newOrder) {
            merge(currentOrder, newOrder);
          }
        });
      });
    };

    const cvmDevicetypeParams = computed(() => {
      const { region, zone } = cvmProduceForm.value;
      return { region, zone };
    });

    const taskPoll = useTimeoutPoll(pollProduceOrderList, 30000, { max: 60 });

    onMounted(() => {
      fetchRequireType();
      loadCvmProduceDetail();
      taskPoll.resume();
    });

    return () => (
      <div class='apply-list-container cvm-produce-wrapper'>
        <div class={'filter-container search-container'}>
          <Form formType='vertical' class='scr-form-wrapper' model={cvmProduceForm}>
            <FormItem label='需求类型'>
              <Select v-model={cvmProduceForm.value.require_type} multiple clearable placeholder='请选择'>
                {requireTypeList.value.map(({ label, value }) => {
                  return <Select.Option key={value} name={label} id={value} />;
                })}
              </Select>
            </FormItem>
            <FormItem label='地域'>
              <area-selector multiple v-model={cvmProduceForm.value.region} params={{ resourceType: 'QCLOUDCVM' }} />
            </FormItem>
            <FormItem label='园区'>
              <zone-selector
                multiple
                v-model={cvmProduceForm.value.zone}
                params={{ resourceType: 'QCLOUDCVM', region: cvmProduceForm.value.region }}
              />
            </FormItem>
            <FormItem label='机型'>
              <DevicetypeSelector
                v-model={cvmProduceForm.value.device_type}
                resourceType='cvm'
                params={cvmDevicetypeParams.value}
                multiple
              />
            </FormItem>
            <FormItem label='单号'>
              <FloatInput v-model={cvmProduceForm.value.order_id} placeholder='请输入单号，多个换行分割' />
            </FormItem>
            <FormItem label='云梯单号'>
              <FloatInput v-model={cvmProduceForm.value.task_id} placeholder='请输入云梯单号，多个换行分割' />
            </FormItem>
            <FormItem label='状态'>
              <Select v-model={cvmProduceForm.value.status} multiple clearable placeholder='请选择状态'>
                {statusList.value.map(({ status, description }) => {
                  return <Select.Option key={status} name={description} id={status} />;
                })}
              </Select>
            </FormItem>
            <FormItem label='创建人'>
              <member-select
                v-model={cvmProduceForm.value.bk_username}
                multiple
                clearable
                defaultUserlist={[
                  {
                    username: userStore.username,
                    display_name: userStore.username,
                  },
                ]}
                placeholder='请输入企业微信名'
              />
            </FormItem>
            <FormItem label='回收时间'>
              <bk-date-picker v-model={timeForm.value} type='daterange' />
            </FormItem>
          </Form>
          <div class='button-container'>
            <Button theme='primary' onClick={filterOrders}>
              <Search />
              查询
            </Button>
            <Button onClick={() => clearFilter()}>重置</Button>
          </div>
        </div>
        <div class='cvm-produce-table-panel'>
          <div class='toolbar'>
            <Button theme='primary' onClick={() => handleOrderCreate(null)}>
              <Plus class='f22' />
              创建单据
            </Button>
            <Button theme='primary' onClick={handleFastCvmProduce}>
              快速生产
            </Button>
          </div>
          <CommonTable class={'filter-common-table cvm-produce-table'}>
            {{
              expandRow: (row: any) => {
                return (
                  <PropertyList
                    properties={{
                      imageId: row.spec.image_id,
                      systemDisk: row.spec.system_disk,
                      dataDisk: row.spec.data_disk,
                      bkBizId: row.bk_biz_id,
                      module: 'SA云化池',
                      vpc: row.spec.vpc,
                      subnet: row.spec.subnet,
                    }}
                    row={row}
                  />
                );
              },
            }}
          </CommonTable>
        </div>
        <CreateOrderDialog
          v-model={isCreateCvmDialogShow.value}
          cvmDeviceDetail={fastProduceData.value}
          onConfirm-success={handleCreateCvmConfirmSuccess}
          onClosed={handleCreateCvmClosed}
        />
        <fast-cvm-produce v-model={isFastProVisible.value} />
        <success-produce-detail v-model={isShowProduceDetail.value} tableData={producedDetail.value} />
      </div>
    );
  },
});
