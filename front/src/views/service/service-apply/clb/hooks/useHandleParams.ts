import { nextTick, onMounted, onUnmounted, watch } from 'vue';
// import types
import { ApplyClbModel } from '@/api/load_balancers/apply-clb/types';
import bus from '@/common/bus';

export default (formModel: ApplyClbModel, formRef: any) => {
  // 重置参数
  const resetParams = (
    keys: string[] = ['zones', 'backup_zones', 'cloud_vpc_id', 'cloud_subnet_id', 'vip_isp', 'cloud_eip_id'],
  ) => {
    keys.forEach((key) => {
      switch (typeof formModel[key]) {
        case 'number':
          formModel[key] = 0;
          break;
        case 'string':
          formModel[key] = '';
      }
    });
  };
  // 清除校验结果
  const handleClearValidate = () => {
    nextTick(() => {
      formRef.value.clearValidate();
    });
  };

  watch([() => formModel.account_id, () => formModel.region], () => {
    // 当 account_id 或 region 改变时, 恢复默认状态
    formModel.load_balancer_type = 'OPEN';
    resetParams();
    Object.assign(formModel, {
      address_ip_version: 'IPV4',
      zoneType: 'single',
      sla_type: 'shared',
      internet_charge_type: 'TRAFFIC_POSTPAID_BY_HOUR',
    });
    handleClearValidate();
  });

  watch(
    () => formModel.load_balancer_type,
    (val) => {
      // 重置通用参数
      resetParams();
      if (val === 'INTERNAL') {
        resetParams(['address_ip_version', 'sla_type', 'internet_charge_type', 'internet_max_bandwidth_out']);
      } else {
        // 如果是公网, 则重置初始状态
        Object.assign(formModel, {
          address_ip_version: 'IPV4',
          zoneType: 'single',
          sla_type: 'shared',
          internet_charge_type: 'TRAFFIC_POSTPAID_BY_HOUR',
        });
      }
      handleClearValidate();
    },
  );

  watch(
    () => formModel.address_ip_version,
    () => {
      resetParams(['zones', 'backup_zones', 'vip_isp']);
      handleClearValidate();
    },
  );

  watch(
    () => formModel.zoneType,
    () => {
      resetParams(['zones', 'backup_zones', 'cloud_vpc_id']);
      handleClearValidate();
    },
  );

  onMounted(() => {
    bus.$on('resetParams', resetParams);
  });

  onUnmounted(() => {
    bus.$off('resetParams');
  });
};
