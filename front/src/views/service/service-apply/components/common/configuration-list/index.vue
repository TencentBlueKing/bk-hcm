<template>
  <Button
    class="mb10"
    theme="primary"
    outline
    :disabled="!vendor || overLength"
    @click="showConfigureSlider"
    v-bk-tooltips="{
      content: disabledTip,
      disabled: tooltipsDisabled,
    }"
  >
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
  <div class="dropdown-menu" ref="menu" v-if="isShow">
    <div class="list" @click="() => handleEdit(actionData.rowKey)">编辑</div>
    <div class="list" @click="() => handleRemove(actionData.rowKey)">移除</div>
  </div>
</template>

<script setup lang="ts">
import { computed, h, ref, withDirectives } from 'vue';
import type { Ref } from 'vue';
import { Button, $bkPopover, bkTooltips } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import { useI18n } from 'vue-i18n';
import type { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import { useRegionsStore } from '@/store/useRegionsStore';
import { LB_ISP, VendorEnum, CLB_SPECS, NET_CHARGE_MAP } from '@/common/constant';
import { LB_NETWORK_TYPE_MAP } from '@/constants';
import { IP_VERSION_DISPLAY_NAME, IpVersionType } from '@/views/load-balancer/constants';
import { type Settings } from 'bkui-vue/lib/table/props';

const props = withDefaults(defineProps<IConfigurationListProps>(), {
  list: () => [],
  vendor: '',
});

const emit = defineEmits(['showConfigureSlider', 'cloneData', 'removeData', 'editData']);

const maxNum = 5; // 每次最大提交

const popInstance = ref(null);
const menu = ref(null);
const isShow = ref(false);

const { t } = useI18n();
const { getRegionName, getZoneName } = useRegionsStore();

const disabledTip = computed(() => {
  if (overLength.value) {
    return t(`一次性最大只能添加${maxNum}个`);
  }
  return t(`请先选择云账户`);
});
const tooltipsDisabled = computed(() => {
  if (!props.vendor) return false;
  if (overLength.value) return false;
  return true;
});
const overLength = computed(() => props.list.length >= maxNum);

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
    render({ cell, row }: { cell: string | string[]; row: { vendor: VendorEnum } }) {
      if (!cell) return h('span', '--');
      return h('span', [
        Array.isArray(cell)
          ? cell.map((zone) => getZoneName(zone, row.vendor)).join(',') || '--'
          : getZoneName(cell as string, row.vendor) || '--',
      ]);
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

const actionData = ref<ApplyClbModel>({
  bk_biz_id: 0,
  account_id: '',
  region: '',
  load_balancer_type: 'OPEN',
  name: '',
  zones: '',
  load_balancer_pass_to_target: false,
  cloud_vpc_id: '',
  require_count: 0,
  zoneType: '0',
  vendor: VendorEnum.TCLOUD,
  account_type: 'STANDARD',
  slaType: '0',
});

// 初始化popover, 并显示
const showDropdownList = (e: any, data: ApplyClbModel) => {
  popInstance.value?.close();
  actionData.value = data;
  isShow.value = true;

  popInstance.value = $bkPopover({
    isShow: isShow.value,
    trigger: 'manual',
    forceClickoutside: true,
    theme: 'light',
    renderType: 'shown',
    placement: 'bottom',
    arrow: false,
    allowHtml: true,
    extCls: 'more-action-dropdown-menu',
    target: e,
    content: menu,
    width: '',
    always: false,
    disabled: false,
    height: '',
    maxWidth: '',
    maxHeight: '',
    padding: 0,
    offset: 10,
    zIndex: 0,
    disableTeleport: false,
    autoPlacement: false,
    autoVisibility: false,
    disableOutsideClick: false,
    disableTransform: false,
    modifiers: [],
    popoverDelay: 0,
    componentEventDelay: 0,
    immediate: false,
  });
  popInstance.value?.show();
  popInstance.value?.update(e.target);
  popInstance.value?.show();
};

configListColumns.push({
  label: '操作',
  width: 120,
  fixed: 'right',
  showOverflowTooltip: false,
  render: ({ data }: { data: any; index: number }) => {
    return h('div', { class: 'operation-column' }, [
      withDirectives(
        h(
          Button,
          {
            text: true,
            theme: 'primary',
            class: 'mr10',
            onClick: () => handleClone(data),
            disabled: overLength.value,
          },
          '克隆',
        ),
        [
          [
            bkTooltips,
            {
              content: disabledTip.value,
              disabled: tooltipsDisabled.value,
            },
          ],
        ],
      ),
      h(
        'div',
        {
          class: ['more-action'],
          onClick: (e) => showDropdownList(e, data),
        },
        h('i', { class: 'hcm-icon bkhcm-icon-more-fill' }),
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

const handleEdit = (key: string) => {
  isShow.value = false;
  popInstance.value?.hide();
  emit('editData', key);
};

const handleRemove = (key: string) => {
  isShow.value = false;
  popInstance.value?.hide();
  emit('removeData', key);
};
</script>

<style lang="scss" scoped>
:deep(.operation-column) {
  height: 100%;
  display: flex;
  align-items: center;

  .more-action {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    cursor: pointer;

    & > i {
      position: absolute;
    }

    &:hover {
      background-color: #f0f1f5;
    }

    &.current-operate-row {
      background-color: #f0f1f5;
    }

    &.disabled {
      background-color: #fff;
      color: #dcdee5;
      cursor: not-allowed;
    }
  }
}
</style>

<style lang="scss">
.more-action-dropdown-menu {
  padding: 0 !important;
  background: red !important;

  .dropdown-menu {
    background: #fff;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    box-sizing: border-box;
    margin: 0;
    min-width: 100%;
    padding: 5px 0;

    .list {
      color: #63656e;
      cursor: pointer;
      display: block;
      font-size: 12px;
      height: 32px;
      line-height: 33px;
      list-style: none;
      padding: 0 16px;
      white-space: nowrap;

      &:hover {
        background: #f5f7fa;
      }
    }
  }
}
</style>
