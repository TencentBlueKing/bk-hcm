<template>
  <Button class="mb10" theme="primary" outline :disabled="!vendor" @click="showConfigureSlider">
    <Plus />
    {{ t('添加') }}
  </Button>
  <bk-table
    class="config-list-table"
    :data="list"
    :columns="configListColumns"
    :max-height="300"
    row-hover="auto"
    :stripe="true"
    show-overflow-tooltip
    :settings="settings"
  />
</template>

<script setup lang="ts">
import { h, ref } from 'vue';
import type { Ref } from 'vue';
import { Button } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import { useI18n } from 'vue-i18n';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { useRegionsStore } from '@/store/useRegionsStore';
import { LB_ISP, VendorEnum, CLB_SPECS, NET_CHARGE_MAP } from '@/common/constant';
import { LB_NETWORK_TYPE_MAP } from '@/constants';
import { IP_VERSION_DISPLAY_NAME, IpVersionType } from '@/views/load-balancer/constants';
import { type Settings } from 'bkui-vue/lib/table/props';

withDefaults(defineProps<IConfigurationListProps>(), {
  list: () => [],
  vendor: '',
});

const emit = defineEmits(['showConfigureSlider', 'cloneData', 'removeData']);

const { t } = useI18n();
const { getRegionName } = useRegionsStore();

export interface IConfigurationListProps {
  list: ApplyClbModel[];
  vendor: string;
}
const configListColumns = [
  {
    label: '地域',
    field: 'region',
    width: 120,
    isDefaultShow: true,
    render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
  },
  {
    label: '可用区',
    field: 'zones',
    isDefaultShow: true,
    render({ cell }: { cell: string }) {
      return h('span', [cell || '--']);
    },
  },
  {
    label: '网络类型',
    field: 'load_balancer_type',
    isDefaultShow: true,
    render: ({ cell }: { cell: string }) => LB_NETWORK_TYPE_MAP[cell] || '--',
  },
  {
    label: '运营商类型',
    field: 'vip_isp',
    width: 100,
    isDefaultShow: true,
    render: ({ cell }: { cell: string }) => LB_ISP[cell] ?? (cell || '--'),
  },
  {
    label: '需求数量',
    isDefaultShow: true,
    field: 'require_count',
  },
  {
    label: 'IP版本',
    field: 'address_ip_version',
    width: 100,
    render: ({ cell }: { cell: IpVersionType }) => IP_VERSION_DISPLAY_NAME[cell.toLowerCase() as IpVersionType],
  },
  {
    label: '规格',
    field: 'sla_type',
    isDefaultShow: true,
    render: ({ cell }: { cell: any }) => CLB_SPECS[cell] ?? '--',
  },
  {
    label: '带宽上限',
    field: 'internet_max_bandwidth_out',
    isDefaultShow: true,
    render: ({ cell }: { cell: any }) => (cell ? `${cell}（Mbps）` : '--'),
  },
  {
    label: '安全组模式',
    field: 'load_balancer_pass_to_target',
    showOverflowTooltip: true,
    isDefaultShow: true,
    render: ({ cell }: { cell: boolean }) => (cell ? '默认放通' : '不启用默认放通'),
  },
  {
    label: '带宽计费模式',
    field: 'internet_charge_type',
    isDefaultShow: true,
    render: ({ cell }: { cell: any }) => NET_CHARGE_MAP[cell],
  },
  {
    label: '所属VPC',
    field: 'cloud_vpc_id',
    showOverflowTooltip: true,
    isDefaultShow: true,
    render({ cell }: { cell: string }) {
      return h('span', [cell || '--']);
    },
  },
  {
    label: '所属子网',
    showOverflowTooltip: true,
    field: 'cloud_subnet_id',
    isDefaultShow: true,
    render({ cell }: { cell: string }) {
      return h('span', [cell || '--']);
    },
  },
  {
    label: '实例名称',
    showOverflowTooltip: true,
    field: 'name',
    render({ cell }: { cell: string }) {
      return h('span', [cell || '--']);
    },
  },
];

const generateColumnsSettings = (columns: any) => {
  const fields = [];
  for (const column of columns) {
    if (column.field && column.label) {
      fields.push({
        label: column.label,
        field: column.field,
        isDefaultShow: !!column.isDefaultShow,
        notDisplayedInBusiness: !!column.notDisplayedInBusiness,
      });
    }
  }
  const settings: Ref<Settings> = ref({
    fields,
    checked: fields.filter((field) => field.isDefaultShow).map((field) => field.field),
  });

  return settings;
};

const settings = generateColumnsSettings(configListColumns);
configListColumns.push({
  label: '操作',
  width: 120,
  showOverflowTooltip: false,
  render: ({ data }: { data: any }) => {
    return h('div', { class: 'operation-column' }, [
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          class: 'mr10',
          onClick: () => handleClone(data),
        },
        '克隆',
      ),
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          class: 'mr10',
          onClick: () => handleRemove(data.rowKey),
        },
        '移除',
      ),
    ]);
  },
});

const showConfigureSlider = () => {
  emit('showConfigureSlider');
};

const handleClone = (data: ApplyClbModel) => {
  emit('cloneData', data);
};

const handleRemove = (key: string) => {
  emit('removeData', key);
};
</script>
