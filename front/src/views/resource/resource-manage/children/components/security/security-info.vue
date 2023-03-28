<script lang="ts" setup>
// import InfoList from '../../../common/info-list/info-list';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import useDetail from '@/views/resource/resource-manage/hooks/use-detail';

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

const props = defineProps({
  id: {
    type: String as PropType<any>,
  },
  vendor: {
    type: String as PropType<any>,
  },
});

const {
  t,
} = useI18n();

const resourceStore = useResourceStore();

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
    edit: true,
  },
];

const {
  loading,
  detail,
  getDetail,
} = useDetail(
  'security_groups',
  props.id,
);
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
    getDetail();
  } catch (error) {

  }
};
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-info class="mt20" :fields="settingInfo" :detail="detail" @change="handleChange"></detail-info>
  </bk-loading>
</template>
