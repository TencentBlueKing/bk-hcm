import { PropType, computed, defineComponent, ref, watchEffect } from 'vue';
import { Select, Tag } from 'bkui-vue';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useSingleList } from '@/hooks/useSingleList';
import { VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum } from '@/typings';
import './index.scss';

const { Option } = Select;

/**
 * 主备可用区选择器
 * @prop zones 主可用区
 * @prop backupZones 备可用区
 * @prop vendor 云厂商
 * @prop region 地域
 */
export default defineComponent({
  name: 'PrimaryStandZoneSelector',
  props: {
    zones: String,
    backupZones: String,
    vendor: String as PropType<VendorEnum>,
    region: String,
  },
  emits: ['update:zones', 'update:backupZones'],
  setup(props, { emit }) {
    // 引入一些可能用得着的 store
    const regionStore = useRegionsStore();

    const zones = ref('');
    const backupZones = ref('');
    const selectedValue = computed(() => {
      if (!zones.value && !backupZones.value) return [];
      return [zones.value, backupZones.value];
    });

    const { dataList, isDataLoad, handleScrollEnd } = useSingleList({
      url: () => `/api/v1/cloud/vendors/${props.vendor}/regions/${props.region}/zones/list`,
      rules: () => [
        { op: QueryRuleOPEnum.EQ, field: 'vendor', value: props.vendor },
        { op: QueryRuleOPEnum.EQ, field: 'state', value: 'AVAILABLE' },
      ],
      immediate: true,
    });

    // 钩子 - 选中option
    const handleSelect = (v: string) => {
      if (zones.value && !backupZones.value) {
        backupZones.value = v;
      } else {
        zones.value = v;
      }
    };
    // 钩子 - 取消选中option
    const handleDeSelect = (v: string) => {
      if (zones.value === v) {
        zones.value = '';
      } else {
        backupZones.value = '';
      }
    };
    // 钩子 - 清空选中
    const handleClear = () => {
      zones.value = '';
      backupZones.value = '';
    };

    // click-handler - 交换主备可用区
    const handleExchange = (e: MouseEvent) => {
      e.stopPropagation();
      const temp = zones.value;
      zones.value = backupZones.value;
      backupZones.value = temp;
    };

    // close-handler - tag取消选中
    const handleClose = (isBackup: boolean) => () => {
      (isBackup ? backupZones : zones).value = '';
    };

    // 更新 formData
    watchEffect(() => {
      emit('update:zones', zones.value);
      emit('update:backupZones', backupZones.value);
    });

    return () => (
      <Select
        class='primary-stand-zone-selector'
        modelValue={selectedValue.value}
        multiple
        multipleMode='tag'
        onScroll-end={handleScrollEnd}
        onSelect={handleSelect}
        onDeselect={handleDeSelect}
        onClear={handleClear}
        loading={isDataLoad.value}
        scrollLoading={isDataLoad.value}>
        {{
          default: () =>
            dataList.value.map(({ id, name, name_cn }) => (
              <Option key={id} id={name} name={name_cn || name}>
                <span>{name_cn || name}</span>
                {zones.value === name && (
                  <Tag class='ml12' theme='info'>
                    主可用区
                  </Tag>
                )}
                {backupZones.value === name && (
                  <Tag class='ml12' theme='warning'>
                    备可用区
                  </Tag>
                )}
              </Option>
            )),
          tag: () => (
            <div class='selected-tag-value-container'>
              {zones.value && (
                <Tag closable theme='info' onClose={handleClose(false)}>
                  主&nbsp;:&nbsp;{regionStore.getZoneName(zones.value, props.vendor)}
                </Tag>
              )}
              {zones.value && backupZones.value && (
                <i class='hcm-icon bkhcm-icon-exchange-line' onClick={handleExchange}></i>
              )}
              {backupZones.value && (
                <Tag closable theme='warning' onClose={handleClose(true)}>
                  备&nbsp;:&nbsp;{regionStore.getZoneName(backupZones.value, props.vendor)}
                </Tag>
              )}
            </div>
          ),
        }}
      </Select>
    );
  },
});
