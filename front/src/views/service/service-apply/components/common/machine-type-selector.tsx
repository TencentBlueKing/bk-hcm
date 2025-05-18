import http from '@/http';
import { computed, defineComponent, PropType, reactive, ref, watch, watchEffect } from 'vue';
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
    const resList = ref([]);
    const pagination = reactive({
      start: 0,
      limit: 10,
      count: 100,
    });
    const searchVal = ref([]);
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
    const checkedInstance = reactive({
      instanceType: '',
      typeName: '',
      cpu: '',
      memory: '',
    });

    const columns = ref([
      {
        label: '类型',
        field: 'type_name',
        render: ({ cell, row }: any) => {
          return (
            <div class={'flex-row'}>
              <Radio
                v-model={checkedInstance.instanceType}
                checked={checkedInstance.instanceType === row.instance_type}
                label={row.instance_type}
                // onChange={() => handleChangeCheckedInstance(row)}
              >
                {props.vendor === VendorEnum.TCLOUD ? cell : row.instance_type}
              </Radio>
            </div>
          );
        },
      },
      {
        label: '规格',
        field: 'instance_type',
      },
      {
        label: 'CPU',
        field: 'cpu',
        render: ({ cell }: { cell: string }) => `${cell}核`,
      },
      {
        label: '内存',
        field: 'memory',
        render: ({ cell }: { cell: string }) => `${Math.floor(+cell / 1024)}GB`,
      },
      {
        label: '处理器型号',
        field: 'cpu_type',
      },
      {
        label: '内网带宽',
        field: 'instance_bandwidth',
        render: ({ row }: { row: any }) =>
          `${props.vendor === VendorEnum.TCLOUD ? `${row.instance_bandwidth}Gbps` : row.network_performance}`,
      },
      {
        label: '网络收发包',
        field: 'instance_pps',
        render: ({ cell }: { cell: string }) => `${cell}万PPS`,
      },
      {
        label: '参考费用',
        field: 'price',
        fixed: 'right',
        render: ({ row }: any) => (
          <span class={'instance-price'}>{`${
            row?.Price?.DiscountPrice || row?.Price?.DiscountPriceOneYear
          }元/月`}</span>
        ),
      },
    ]);

    const selected = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    watch(
      [() => props.vendor, () => props.accountId, () => props.region, () => props.zone, () => props.instanceChargeType],
      async ([vendor, accountId, region, zone, instanceChargeType], [, , , oldZone]) => {
        if (
          !vendor ||
          !accountId ||
          !region ||
          (vendor !== VendorEnum.AZURE && !zone) ||
          (vendor === VendorEnum.TCLOUD && !instanceChargeType)
        ) {
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
            : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${props.bizId}/instance_types/list`,
          {
            account_id: accountId,
            vendor,
            region,
            zone,
            instance_charge_type: instanceChargeType,
          },
        );
        list.value = result.data?.instance_types || [];
        instanceFamilyTypesList.value =
          props.vendor === VendorEnum.TCLOUD ? result.data?.instance_family_type_names : result.data?.instance_families;

        loading.value = false;
      },
    );

    const computedColumns = computed(() => {
      return columns.value
        .filter(
          ({ field }) => !['cpu_type', 'instance_pps', 'price'].includes(field) || props.vendor === VendorEnum.TCLOUD,
        )
        .filter(
          ({ field }) =>
            !['instance_bandwidth'].includes(field) ||
            [VendorEnum.TCLOUD, VendorEnum.AWS].includes(props.vendor as VendorEnum),
        );
    });

    watchEffect(() => {
      if (!selectedFamilyType.value) resList.value = list.value;
      const reg = new RegExp(selectedFamilyType.value);
      switch (props.vendor) {
        case VendorEnum.TCLOUD:
          resList.value = list.value.filter(({ type_name }) => reg.test(type_name));
          break;
        case VendorEnum.GCP:
        case VendorEnum.AWS:
        case VendorEnum.AZURE:
          resList.value = list.value.filter(({ instance_family }) => reg.test(instance_family));
          break;
        case VendorEnum.HUAWEI:
          if (!selectedFamilyType.value) break;
          resList.value = list.value.filter(({ instance_family }) => selectedFamilyType.value === instance_family);
          break;
      }
      for (const { id, values } of searchVal.value) {
        let val = values?.[0]?.id;
        if (id === 'memory' && !isNaN(values?.[0]?.id)) val = +val * 1024;
        resList.value = resList.value.filter((item) => item[id] === +val);
      }
    });

    watch(
      () => props.vendor,
      () => (selectedFamilyType.value = ''),
    );

    watch(
      () => props.accountId,
      (newVal, oldVal) => {
        if (newVal !== oldVal) {
          checkedInstance.instanceType = '';
          checkedInstance.typeName = '';
          checkedInstance.cpu = '';
          checkedInstance.memory = '';
        }
      },
    );

    const computedDisabled = computed(() => {
      return !(props.accountId && props.region && props.vendor && props.zone);
    });

    const handleChange = () => {
      selected.value = checkedInstance.instanceType;
      const data = list.value.find((item) => item.instance_type === checkedInstance.instanceType);
      emit('change', data);
      isDialogShow.value = false;
    };

    const handleChangeCheckedInstance = (data: any) => {
      checkedInstance.instanceType = data.instance_type;
      checkedInstance.typeName = VendorEnum.TCLOUD === props.vendor ? data.type_name : data.instance_type;
      checkedInstance.cpu = `${data.cpu}核`;
      checkedInstance.memory = `${data.memory / 1024}GB`;
    };

    const handleOnRowClick = (_: any, row: any) => {
      handleChangeCheckedInstance(row);
    };

    const bkTooltipsOptions = computed(() => {
      if (checkedInstance.instanceType)
        return {
          content: `${checkedInstance.instanceType} (${checkedInstance.typeName}, ${checkedInstance.cpu}${checkedInstance.memory})`,
        };
      return {
        content: '--',
      };
    });

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
          {selected.value ? (
            <div class={'selected-block mr8'} v-BkTooltips={bkTooltipsOptions.value}>
              {`${checkedInstance.instanceType} (${checkedInstance.typeName}, ${checkedInstance.cpu}${checkedInstance.memory})`}
            </div>
          ) : null}
          {selected.value ? (
            <EditLine fill='#3A84FF' width={13.5} height={13.5} onClick={() => (isDialogShow.value = true)} />
          ) : (
            <Button onClick={() => (isDialogShow.value = true)} disabled={computedDisabled.value}>
              <Plus class='f20' />
              选择机型
            </Button>
          )}
        </div>
        <Dialog
          isShow={isDialogShow.value}
          onConfirm={handleChange}
          title='选择机型'
          closeIcon={false}
          quick-close={false}
          width={'60vw'}
          height={'80vh'}>
          {{
            default: () => (
              <>
                <Form class='selected-block-dialog-form' labelWidth={100} labelPosition='right'>
                  <FormItem label='机型族'>
                    <BkButtonGroup>
                      <Button
                        selected={selectedFamilyType.value === ''}
                        onClick={() => {
                          selectedFamilyType.value = '';
                        }}>
                        全部
                      </Button>
                      {instanceFamilyTypesList.value.map((name) => (
                        <Button
                          selected={selectedFamilyType.value === name}
                          onClick={() => {
                            selectedFamilyType.value = name;
                          }}>
                          {name}
                        </Button>
                      ))}
                    </BkButtonGroup>
                  </FormItem>
                  <FormItem label='已选'>
                    <div class={'instance-type-search-seletor-container'}>
                      <div class={'selected-block-container'}>
                        <div class={'selected-block'} v-BkTooltips={bkTooltipsOptions.value}>
                          {checkedInstance.instanceType
                            ? `${checkedInstance.instanceType}  (${checkedInstance.typeName}, ${checkedInstance.cpu}${checkedInstance.memory})`
                            : '--'}
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
                    data={resList.value}
                    columns={computedColumns.value}
                    pagination={pagination}
                    onRowClick={handleOnRowClick}
                    rowKey={'instance_type'}
                    showOverflowTooltip
                  />
                </Loading>
              </>
            ),
            footer: () => (
              <div>
                <Button theme='primary' class={'mr6'} disabled={!checkedInstance.instanceType} onClick={handleChange}>
                  {' '}
                  确认{' '}
                </Button>
                <Button
                  onClick={() => {
                    isDialogShow.value = false;
                    if (!selected.value) {
                      Object.assign(checkedInstance, {
                        instanceType: '',
                        typeName: '',
                        cpu: '',
                        memory: '',
                      });
                    }
                  }}>
                  {' '}
                  取消{' '}
                </Button>
              </div>
            ),
          }}
        </Dialog>
      </div>
    );
  },
});
