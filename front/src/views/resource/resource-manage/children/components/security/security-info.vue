<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import { useResourceStore } from '@/store';

import { useI18n } from 'vue-i18n';

import { PropType } from 'vue';

import { Message } from 'bkui-vue';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { timeFormatter, parseTags } from '@/common/util';
import { FieldList } from '../../../common/info-list/types';

const props = defineProps({
  id: {
    type: String as PropType<any>,
  },
  vendor: {
    type: String as PropType<any>,
  },
  loading: {
    type: Boolean as PropType<boolean>,
  },
  detail: {
    type: Object as PropType<any>,
  },
  getDetail: {
    type: Function as PropType<() => void>,
  },
});

const { t } = useI18n();

const resourceStore = useResourceStore();
const { getRegionName } = useRegionsStore();
const { getNameFromBusinessMap } = useBusinessMapStore();

const settingInfo: FieldList = [
  {
    name: 'ID',
    prop: 'id',
  },
  {
    name: t('账号 ID'),
    prop: 'account_id',
  },
  {
    name: t('资源名称'),
    prop: 'name',
    edit: props.vendor !== 'azure' && props.vendor !== 'aws',
  },
  {
    name: t('云资源ID'),
    prop: 'cloud_id',
  },
  {
    name: t('云厂商'),
    prop: 'vendorName',
  },
  {
    name: t('业务'),
    prop: 'bk_biz_id',
    render: (val: number) => (val === -1 ? '未分配' : `${getNameFromBusinessMap(val)} (${val})`),
  },
  {
    name: t('地域'),
    prop: 'region',
    render: () => getRegionName(props.vendor, props.detail?.region),
  },
  {
    name: t('创建时间'),
    prop: 'created_at',
    render: (val: string) => timeFormatter(val),
  },
  {
    name: t('修改时间'),
    prop: 'updated_at',
    render: (val: string) => timeFormatter(val),
  },
  {
    name: t('标签'),
    prop: 'tags',
    render: (val: any) => parseTags(val),
  },
  {
    name: t('备注'),
    prop: 'memo',
    edit: props.vendor !== 'aws',
  },
];

if (props.vendor === 'tcloud' || props.vendor === 'aws' || props.vendor === 'huawei') {
  settingInfo.splice(8, 0, {
    name: t('关联CVM实例数'),
    prop: 'cvm_count',
    render(val: number) {
      return val;
    },
  });
  if (props.vendor === 'aws') {
    settingInfo.splice(
      9,
      0,
      {
        name: t('所属VPC'),
        prop: 'vpc_id',
        render(val: string) {
          return val;
        },
      },
      {
        name: t('所属云VPC'),
        prop: 'cloud_vpc_id',
        render(val: string) {
          return val;
        },
      },
    );
  }
} else if (props.vendor === 'azure') {
  settingInfo.splice(
    7,
    0,
    {
      name: t('关联网络接口数'),
      prop: 'network_interface_count',
      render(val: number) {
        return val;
      },
    },
    {
      name: t('关联子网数'),
      prop: 'subnet_count',
      render(val: number) {
        return val;
      },
    },
  );
}

const handleChange = async (val: any) => {
  try {
    await resourceStore.updateSecurityInfo(props.id, val);
    Message({
      theme: 'success',
      message: t('更新成功'),
    });
    props.getDetail();
  } catch (error) {}
};
</script>

<template>
  <bk-loading :loading="props.loading">
    <detail-info
      :fields="settingInfo"
      :detail="props.detail"
      @change="handleChange"
      label-width="130px"
      global-copyable
    />
  </bk-loading>
</template>
