import http from '@/http';
import { computed, defineComponent, PropType, reactive, ref, watch } from 'vue';
import { Button, Dialog, Form, Table } from 'bkui-vue';
import './machine-type-selector.scss';

// import { formatStorageSize } from '@/common/util';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { Plus } from 'bkui-vue/lib/icon';
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
    const columns = [{
      label: '类型',
      field: 'instance_family',
      filter: true,
    },
    {
      label: '规格',
      field: 'instance_type',
    },
    {
      label: 'CPU',
      field: 'cpu',
    },
    {
      label: '内存',
      field: 'memory',
    },
    {
      label: '处理器型号',
      field: 'cpu_type',
    },
    {
      label: '网络收发包',
      field: 'instance_pps',
    },
    {
      label: '参考费用',
      field: 'price',
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

    // const handleChange = (val: string) => {
    //   const data = list.value.find(item => item.instance_type === val);
    //   emit('change', data);
    // };

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
        {selected.value ? (
          <div class={'selected-block-container'}>
            <div class={'selected-block'}>
              Amazon Linux 2 AMI (HVM) - Kernel 5.10, SSD Volume Type
            </div>
          </div>
        ) : (
          <Button onClick={() => (isDialogShow.value = true)}>
            <Plus class='f20' />
            选择机型
          </Button>
        )}
        <Dialog
          isShow={isDialogShow.value}
          onClosed={() => (isDialogShow.value = false)}
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
              <div class={'selected-block-container'}>
                <div class={'selected-block'}>
                  Amazon Linux 2 AMI (HVM) - Kernel 5.10, SSD Volume Type
                </div>
              </div>
            </FormItem>
          </Form>
          <Table
            data={list.value}
            columns={columns}
            pagination={pagination}
          >

          </Table>
        </Dialog>
      </div>
    );
  },
});
