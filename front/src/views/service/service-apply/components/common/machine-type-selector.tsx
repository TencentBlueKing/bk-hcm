import http from '@/http';
import { computed, defineComponent, PropType, reactive, ref, watch } from 'vue';
import { Button, Dialog, Form, Loading, Radio, SearchSelect, Table } from 'bkui-vue';
import './machine-type-selector.scss';

// import { formatStorageSize } from '@/common/util';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { EditLine, Plus } from 'bkui-vue/lib/icon';
import { BkButtonGroup } from 'bkui-vue/lib/button';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

// const { Option } = Select;
const { FormItem } = Form;

export default defineComponent({
  props: {
    modelValue: String as PropType<string>,
    vendor: String as PropType<string>,
    accountId: String as PropType<string>,
    region: String as PropType<string>,
    zone: String as PropType<string>,
    bizId: Number as PropType<number | string>,
    instanceChargeType: String as PropType<string>,
  },
  emits: ['update:modelValue', 'change'],
  // setup(props, { emit, attrs }) {
  setup(props, { emit }) {
    const list = ref([]);
    const loading = ref(false);
    const { isResourcePage } = useWhereAmI();
    const isDialogShow = ref(false);
    const instanceFamilyTypesList = ref([]);
    const selectedFamilyType = ref('');
    const pagination = reactive({
      start: 0,
      limit: 10,
      count: 100,
    });
    const searchVal = ref('');
    const searchData = [
      {
        name: '内存',
        id: 'memory',
      },
      {
        name: 'CPU',
        id: 'cpu',
      },
    ];
    const checkedInstanceType = ref('');
    const columns = [{
      label: '类型',
      field: 'type_name',
      render: ({ cell, data }: any) => {
        return (<div class={'flex-row'}>
        <Radio
          v-model={checkedInstanceType.value}
          checked={checkedInstanceType.value === data.instance_type}
          label={data.instance_type}
        >
          { cell }
        </Radio>
      </div>);
      },
    },
    {
      label: '规格',
      field: 'instance_type',
    },
    {
      label: 'CPU',
      field: 'cpu',
      render: ({ cell }: {cell: string}) => `${cell}核`,
    },
    {
      label: '内存',
      field: 'memory',
      render: ({ cell }: {cell: string}) => `${Math.floor(+cell / 1024)}GB`,
    },
    {
      label: '处理器型号',
      field: 'cpu_type',
    },
    {
      label: '内网带宽',
      field: 'instance_bandwidth',
      render: ({ cell }: {cell: string}) => `${cell}Gbps`,
    },
    {
      label: '网络收发包',
      field: 'instance_pps',
      render: ({ cell }: {cell: string}) => `${cell}万PPS`,
    },
    {
      label: '参考费用',
      field: 'price',
      fixed: 'right',
      render: ({ data }: any) => <span class={'instance-price'}>{`${data?.Price?.DiscountPrice}元/月`}</span>,
    }];

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    watch([
      () => props.vendor,
      () => props.accountId,
      () => props.region,
      () => props.zone,
      () => props.instanceChargeType,
    ], async ([vendor, accountId, region, zone, instanceChargeType], [,,,oldZone]) => {
      if (!vendor || !accountId || !region || (vendor !== VendorEnum.AZURE && !zone)
      || (vendor === VendorEnum.TCLOUD && !instanceChargeType)) {
        list.value = [];
        return;
      }

      // AZURE时与zone无关，只需要满足其它条件时请求一次
      if (vendor === VendorEnum.AZURE && zone !== oldZone) {
        return;
      }

      loading.value = true;
      const result = await http.post(
        isResourcePage
          ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/instance_types/list`
          : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${props.bizId}/instance_types/list`
        , {
          account_id: accountId,
          vendor,
          region,
          zone,
          instance_charge_type: instanceChargeType,
        },
      );
      list.value = result.data?.instance_types || [];
      instanceFamilyTypesList.value = result.data?.instance_family_type_names;

      loading.value = false;
    });

    const computedList = computed(() => {
      if (!selectedFamilyType.value) return list.value;
      const reg = new RegExp(selectedFamilyType.value);
      return list.value.filter(({ type_name }) => reg.test(type_name));
    });

    const computedDisabled = computed(() => {
      return !(props.accountId && props.region && props.vendor && props.zone);
    });
    // watch(
    //   () => searchVal.value,
    //   (val) => {
    //     const arr = val?.values;
    //     const map = new Map();
    //     for(let )
    //   },
    // );

    const handleChange = () => {
      selected.value = checkedInstanceType.value;
      const data = list.value.find(item => item.instance_type === checkedInstanceType.value);
      emit('change', data);
      isDialogShow.value = false;
    };

    return () => (
      // <Select
      //   filterable={true}
      //   modelValue={selected.value}
      //   onUpdate:modelValue={val => selected.value = val}
      //   loading={loading.value}
      //   onChange={handleChange}
      //   {...{ attrs }}
      // >
      //   {
      //     list.value.map(({ instance_type, cpu, memory, status }, index) => (
      //       <Option
      //         key={index}
      //         value={instance_type}
      //         disabled={status === 'SOLD_OUT'}
      //         // eslint-disable-next-line no-nested-ternary
      //         label={`${instance_type} (${cpu}核CPU，${formatStorageSize(memory * 1024 ** 2)}内存)
      // ${props.vendor === VendorEnum.TCLOUD ? (status === 'SELL' ? '可购买' : '已售罄') : ''}`}
      //       >
      //       </Option>
      //     ))
      //   }
      // </Select>
      <div>
        <div class={'selected-block-container'}>
          {
            selected.value ? (
              <div class={'selected-block mr8'}>
                Amazon Linux 2 AMI (HVM) - Kernel 5.10, SSD Volume Type
              </div>
            ) : null
          }
          {selected.value ? (
            <EditLine fill='#3A84FF' width={13.5} height={13.5}/>
          ) : (
            <Button onClick={() => (isDialogShow.value = true)} disabled={computedDisabled.value}>
              <Plus class='f20' />
              选择机型
            </Button>
          )}
        </div>
        <Dialog
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
          onConfirm={handleChange}
          title='选择机型'
          width={1500}>
          <Form>
            <FormItem label='机型族' labelPosition='left'>
              <BkButtonGroup>
                {instanceFamilyTypesList.value.map(name => (
                  <Button
                    selected={selectedFamilyType.value === name}
                    onClick={() => {
                      selectedFamilyType.value = name;
                    }}
                  >{name}</Button>
                ))}
              </BkButtonGroup>
            </FormItem>
            <FormItem label='已选' labelPosition='left'>
              <div class={'instance-type-search-seletor-container'}>
                <div class={'selected-block-container'}>
                  <div class={'selected-block'}>
                    { checkedInstanceType.value || '--' }
                  </div>
                </div>
                <SearchSelect
                  class='w500 instance-type-search-seletor'
                  v-model={searchVal.value}
                  data={searchData}
                />
              </div>
            </FormItem>
          </Form>
          <Loading loading={loading.value}>
            <Table
              data={computedList.value}
              columns={columns}
              pagination={pagination}
            />
          </Loading>
        </Dialog>
      </div>
    );
  },
});
