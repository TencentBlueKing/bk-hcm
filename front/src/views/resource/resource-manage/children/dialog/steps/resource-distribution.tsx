import './resource-distribution.scss';
import { CloudType } from '@/typings';
import {
  defineComponent,
  ref,
  h,
  watch,
  computed,
} from 'vue';
import {
  Select,
  Message,
} from 'bkui-vue';
import {
  RESOURCE_TYPES,
} from '@/common/constant';
import StepDialog from '@/components/step-dialog/step-dialog';
import AccountSelector from '@/components/account-selector/index.vue';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';
import {
  useI18n,
} from 'vue-i18n';
import {
  useResourceStore,
  useAccountStore,
} from '@/store';

import type {
  FilterType,
} from '@/typings/resource';

export default defineComponent({
  components: {
    StepDialog,
    AccountSelector,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
    chooseResourceType: {
      type: Boolean,
    },
    data: {
      type: Array,
      default() {
        return [];
      },
    },
  },

  emits: ['update:isShow', 'handlePageSizeChange', 'handlePageChange'],

  setup(props, { emit }) {
    // use hooks
    const {
      t,
    } = useI18n();

    const resourceStore = useResourceStore();
    const accountStore = useAccountStore();

    // 状态
    const validateMap = ref({});
    const business = ref('');
    const accountId = ref('');
    const businessList = ref([]);
    const isBusinessError = ref(false);
    const cloudAreas = ref([]);
    const isLoadingCloudAreas = ref(false);
    const cloudAreaPage = ref(0);
    const isBindVPC = ref(false);
    const isBingdingVPC = ref(false);
    const hasBindVPC = ref(false);
    const disableNext = ref(true);
    const vpcFilter: any = ref({ filter: { op: 'and', rules: [] } });
    const accountFilter = ref<FilterType>({ op: 'and', rules: [{ field: 'type', op: 'eq', value: 'resource' }] });
    const vpcTableData = ref<any>([]);
    const resourceTypes = ref([
      'host',
      'vpc',
      'subnet',
      'security',
      'drive',
      'network-interface',
      'ip',
      'routing',
      'image',
    ]);
    const errorList = ref([]);
    const isConfirmLoading = ref(false);
    const VPCColumns = [
      {
        label: 'ID',
        field: 'id',
      },
      {
        label: '账号 ID',
        field: 'account_id',
      },
      {
        label: '资源 ID',
        field: 'cloud_id',
      },
      {
        label: '名称',
        field: 'name',
      },
      {
        label: '云厂商',
        field: 'vendor',
        render({ cell }: { cell: string }) {
          return h(
            'span',
            [
              CloudType[cell] || '--',
            ],
          );
        },
      },
      {
        label: '云区域',
        render({ data }: any) {
          if (data.bk_cloud_id > -1) {
            return data.bk_cloud_id;
          }
          // 校验
          const validate = () => {
            if (!data.temp_bk_cloud_id) {
              errorList.value.push(data.id);
              return false;
            }
            const index = errorList.value.findIndex(item => item === data.id);
            errorList.value.splice(index, 1);
            return true;
          };
          if (!validateMap.value[data.id]) {
            validateMap.value[data.id] = validate;
          }
          // 未绑定需要先绑定云区域
          return () => h(
            Select,
            {
              class: {
                'resource-is-error': errorList.value.includes(data.id),
              },
              displayKey: 'name',
              idKey: 'id',
              list: cloudAreas.value,
              modelValue: data.temp_bk_cloud_id,
              scrollLoading: isLoadingCloudAreas.value,
              filterable: true,
              onScrollEnd() {
                getCloudAreas();
              },
              onChange(val: string) {
                data.temp_bk_cloud_id = val;
                validate();
              },
            },
          );
        },
      },
    ];
    const businessColumns = [
      {
        label: '云厂商',
        field: 'vendor',
        render({ cell }: { cell: string }) {
          return h(
            'span',
            [
              CloudType[cell] || '--',
            ],
          );
        },
      },
      {
        label: '数量',
        field: 'num',
      },
      {
        label: '云区域（ID）',
        field: 'bk_cloud_id',
      },
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      let type = 'vpcs';
      let params: any = {
        vpc_ids: props.data.map((item: any) => item.vpc_id),
        bk_biz_id: business.value,
      };

      // 快速分配参数
      if (props.chooseResourceType) {
        type = 'resources';
        params = {
          account_id: accountId.value,
          bk_biz_id: business.value,
          is_all_res_type: true,
        };
      }

      isConfirmLoading.value = true;
      resourceStore
        .assignBusiness(
          type,
          params,
        )
        .then(() => {
          Message({
            theme: 'success',
            message: '已分配',
          });
          handleClose();
        })
        .finally(() => {
          isConfirmLoading.value = false;
        });
    };

    const getCloudAreas = () => {
      if (isLoadingCloudAreas.value) return;
      isLoadingCloudAreas.value = true;
      resourceStore
        .getCloudAreas({
          page: {
            start: cloudAreaPage.value,
            limit: 100,
          },
        })
        .then((res: any) => {
          cloudAreaPage.value += 1;
          cloudAreas.value.push(...res?.data?.info || []);
        })
        .finally(() => {
          isLoadingCloudAreas.value = false;
        });
    };

    const getBusinessList = async () => {
      try {
        const res = await accountStore.getBizList();
        businessList.value = res?.data;
      } catch (error) {
        console.log(error);
      }
    };

    const getAccountBusinessID = async () => {
      try {
        const res = await accountStore.getBizIdWithAccountId(accountId.value);
        business.value = res?.data?.bk_biz_ids[0];
      } catch (error) {
        console.log(error);
      }
    };


    const validateCloudArea = () => {
      let hasError = false;
      Object.values(validateMap.value).forEach((validate) => {
        if (!(validate as Function)()) {
          hasError = true;
        }
      });
      return hasError;
    };

    // 绑定vpc到云区域
    const handleBindVPC = () => {
      // 校验不通过
      if (validateCloudArea()) {
        Message({
          theme: 'error',
          message: '请先选择云区域',
        });
        return;
      }

      isBingdingVPC.value = true;
      const dataList = props.chooseResourceType ? vpcTableData.value : props.data;
      const bindCloudAreaData = dataList
        .filter((item: any) => item.bk_cloud_id <= -1 || !item.bk_cloud_id)
        .map((item: any) => ({ vpc_id: item.vpc_id || item.id, bk_cloud_id: item.temp_bk_cloud_id }));
      resourceStore
        .bindVPCWithCloudArea(bindCloudAreaData)
        .then(() => {
          dataList
            .filter((item: any) => item.bk_cloud_id <= -1 || !item.bk_cloud_id)
            .forEach((item: any) => {
              item.bk_cloud_id = item.temp_bk_cloud_id;
            });
          disableNext.value = false;
        })
        .finally(() => {
          isBingdingVPC.value = false;
        });
    };

    // 聚合分配确认数据
    const computedCloudData = computed(() => {
      const computedData = props.chooseResourceType ? vpcTableData.value : props.data;
      return computedData.reduce((acc: any[], cur: any) => {
        const cloudData = acc.find((item: any) => item.bk_cloud_id === cur.bk_cloud_id && item.vendor === cur.vendor);
        if (cloudData) {
          cloudData.num += 1;
        } else {
          acc.push({
            bk_cloud_id: cur.bk_cloud_id,
            vendor: cur.vendor,
            num: 1,
          });
        }
        return acc;
      }, []);
    });

    watch(
      () => props.isShow,
      () => {
        if (props.isShow) {
          // 重置状态
          cloudAreas.value = [];
          cloudAreaPage.value = 0;
          disableNext.value = true;
          hasBindVPC.value = false;
          business.value = '';
          validateMap.value = {};
          errorList.value = [];
          // 获取数据
          getCloudAreas();
          getBusinessList();
          // 判断是否需要绑定云区域
          if (!props.chooseResourceType) {
            if (props.data.every((item: any) => item.bk_cloud_id > -1)) {
              disableNext.value = false;
              hasBindVPC.value = true;
            }
          }
        }
      },
    );


    // 根据选择账号获取vpc列表
    watch(
      accountId,
      (value) => {
        Object.assign(vpcFilter.value, { filter: {
          op: 'and',
          rules: [{
            field: 'account_id',
            op: 'eq',
            value,
          }] } });

        getAccountBusinessID();
      },
    );


    const {
      datas,
      isLoading,
      pagination,
      handlePageSizeChange,
      handlePageChange,
    } = useQueryList(vpcFilter.value, 'vpcs');


    watch(
      // 监听数据
      () => datas,
      (vpcData) => {
        if (props.isShow) {
          vpcTableData.value = vpcData.value;
          disableNext.value = true;
          hasBindVPC.value = false;
          validateMap.value = {};
          errorList.value = [];
          // 判断是否需要绑定云区域
          if (vpcTableData.value.every((item: any) => item.bk_cloud_id > -1)) {
            disableNext.value = false;
            hasBindVPC.value = true;
          }
        }
      },
      { deep: true, immediate: true },
    );


    return {
      business,
      accountId,
      businessList,
      isBindVPC,
      isBingdingVPC,
      hasBindVPC,
      disableNext,
      resourceTypes,
      VPCColumns,
      businessColumns,
      computedCloudData,
      isConfirmLoading,
      isBusinessError,
      t,
      handleClose,
      handleConfirm,
      handleBindVPC,
      handlePageSizeChange,
      handlePageChange,
      pagination,
      isLoading,
      vpcTableData,
      datas,
      accountFilter,
    };
  },

  render() {
    // 渲染每一步
    const steps: any[] = [
      {
        title: '前置检查',
        disableNext: this.disableNext,
        component: () => <>
        {this.chooseResourceType
          ? <bk-loading loading={this.isLoading}>
        <div class="flex-row align-items-center mr20">
          <span class="pr10">{this.t('云账号')}</span>
          <AccountSelector filter={this.accountFilter} v-model={this.accountId}></AccountSelector>
          <div class="flex-row align-items-center">
          <section class="resource-head ml20">
            { this.t('目标业务') }
            <bk-select
              v-model={this.business}
              disabled
              filterable
              class={{
                ml10: true,
                'resource-is-error': this.isBusinessError,
              }}
              onChange={(val: any) => this.isBusinessError = !val}
            >
              {
                this.businessList.map(business => <bk-option
                  value={business.id}
                  label={business.name}
                />)
              }
            </bk-select>
          </section>
          </div>
        </div>
        <bk-table
          class="mt20"
          row-hover="auto"
          remote-pagination
          pagination={this.pagination}
          onPageLimitChange={this.handlePageSizeChange}
          onPageValueChange={this.handlePageChange}
          columns={this.VPCColumns}
          data={this.vpcTableData.length ? this.vpcTableData : this.datas}
        />
        </bk-loading>
          : <bk-table
              class="mt20"
              row-hover="auto"
              columns={this.VPCColumns}
              data={this.data}
            />
        }
          {
            !this.hasBindVPC
              ? <bk-checkbox class="mt10" v-model={this.isBindVPC}>
                  注：VPC绑定云区域信息无法修改，请提前确认
                </bk-checkbox>
              : ''
          }
        </>,
        footer: () => <>{
          !this.hasBindVPC
            ? <><bk-button
                class="mr10"
                loading={this.isBingdingVPC}
                disabled={!this.isBindVPC}
                onClick={this.handleBindVPC}
              >
                VPC 绑定云区域
              </bk-button>
              </>
            : ''
        }</>
        ,
      },
      {
        title: '分配确认',
        isConfirmLoading: this.isConfirmLoading,
        component: () => <>
          <bk-table
            class="mt20"
            row-hover="auto"
            columns={this.businessColumns}
            data={this.computedCloudData}
          />
        </>,
      },
    ];

    // 快速分配的时候需要选择资源类型
    if (this.chooseResourceType) {
      steps.unshift({
        title: this.t('资源类型'),
        component: () => <>
          <section>
            <span>{this.t('分配资源类型')}</span>
            <bk-checkbox-group
              class="resource-types"
              v-model={this.resourceTypes}
            >
              {
                RESOURCE_TYPES.map((type) => {
                  return <bk-checkbox disabled={true} label={type.type}>{ this.t(type.name) }</bk-checkbox>;
                })
              }
            </bk-checkbox-group>
          </section>
        </>,
      });
    }

    return <>
      <step-dialog
        business={this.business}
        title={this.title}
        isShow={this.isShow}
        dialogHeight={this.chooseResourceType ? '800' : '720'}
        steps={steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});
