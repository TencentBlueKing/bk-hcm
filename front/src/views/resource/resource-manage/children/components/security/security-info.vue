<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import {
  useResourceStore,
} from '@/store';

import {
  useI18n,
} from 'vue-i18n';

import {
  PropType,
} from 'vue';

import {
  Message } from 'bkui-vue';
import { useRegionsStore } from '@/store/useRegionsStore';

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

const {
  t,
} = useI18n();

const resourceStore = useResourceStore();
const { getRegionName } = useRegionsStore();

const settingInfo: any[] = [
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
  },
  {
    name: t('地域'),
    prop: 'region',
    render: () => getRegionName(props.vendor, props.detail?.region),
  },
  {
    name: t('创建时间'),
    prop: 'created_at',
  },
  {
    name: t('修改时间'),
    prop: 'updated_at',
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
    render(val: any) {
      return val;
    },
  });
  if (props.vendor === 'aws') {
    settingInfo.splice(9, 0, {
      name: t('所属VPC'),
      prop: 'vpc_id',
      render(val: any) {
        return val;
      },
    }, {
      name: t('所属云VPC'),
      prop: 'cloud_vpc_id',
      render(val: any) {
        return val;
      },
    });
  }
} else if (props.vendor === 'azure') {
  settingInfo.splice(8, 0, {
    name: t('关联网络接口数'),
    prop: 'network_interface_count',
    render(val: any) {
      return val;
    },
  }, {
    name: t('关联子网数'),
    prop: 'subnet_count',
    render(val: any) {
      return val;
    },
  });
}

const handleChange = async (val: any) => {
  console.log(val);
  try {
    await resourceStore.updateSecurityInfo(props.id, val);
    Message({
      theme: 'success',
      message: t('更新成功'),
    });
    props.getDetail();
  } catch (error) {

  }
};
</script>

<template>
  <bk-loading
    :loading="props.loading"
  >
    <detail-info class="mt20" :fields="settingInfo" :detail="props.detail" @change="handleChange"></detail-info>
  </bk-loading>
</template>
