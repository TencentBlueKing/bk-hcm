<script setup lang="ts">
import { ref, watch, defineEmits, defineExpose, defineProps } from 'vue';

const emit = defineEmits(['change']);
defineProps({
  region: {
    type: String,
  },
  vendor: {
    type: String,
  },
});
const formData = ref<any>({
  eip_name: '', // eip名称
  eip_type: '5_bgp', // 线路类型. 5_bgp（全动态BGP） |5_sbgp（静态BGP）
  internet_charge_type: 'postPaid', // 计费模式，取值 prePaid(包年/包月) | postPaid(按需计费)
  eip_count: 1,
  bandwidth_option: {
    share_type: 'PER', // 带宽类型， 取值范围：PER，WHOLE（PER为独占带宽，WHOLE是共享带宽）
    charge_mode: 'bandwidth', // 带宽模式。 bandwidth（按照带宽）|traffic（按照流量）
    size: 1, // 带宽大小
    name: `bandwidth-${Math.floor(Math.random() * (9999 - 1000)) + 1000}`,
  },
});
const internet_charge_prepaid = ref({
  period_num: 1,
  period_type: 'month', // 取值 month(按月)| year(按年)
  is_auto_renew: true, // 是否自动续费。0表示手动续费，1表示自动续费
});
const formRef = ref(null);
const rules = {
  eip_name: [
    {
      validator: (value: string) => value.length > 0,
      message: '弹性公网IP名称必填',
      trigger: 'blur',
    },
  ],
  'bandwidth_option.name': [
    {
      validator: (value: string) => value.length > 0,
      message: '带宽名称必填',
      trigger: 'blur',
    },
  ],
};

const handleChange = () => {
  if (formData.value.internet_charge_type === 'prePaid') {
    formData.value.internet_charge_prepaid = internet_charge_prepaid.value;
  } else {
    delete formData.value.internet_charge_prepaid;
  }
  emit('change', formData.value);
};

const validate = () => {
  return formRef.value.validate();
};

watch(
  () => formData,
  () => {
    handleChange();
  },
  {
    immediate: true,
    deep: true,
  },
);

defineExpose([validate]);
</script>

<template>
  <bk-form label-width="150" ref="formRef" :model="formData" :rules="rules">
    <bk-form-item label="弹性公网IP名称" property="eip_name" required>
      <bk-input v-model="formData.eip_name" placeholder="请输入名称" />
    </bk-form-item>
    <bk-form-item label="计费模式">
      <!-- <bk-radio
        v-model="formData.internet_charge_type"
        label="prePaid"
      >
        包年包月
      </bk-radio> -->
      <bk-radio v-model="formData.internet_charge_type" label="postPaid">按量计费</bk-radio>
    </bk-form-item>
    <bk-form-item
      label="线路类型"
      description="查看全动态BGP与静态BGP的区别（https://support.huaweicloud.com/intl/zh-cn/eip_faq/faq_bandwidth_0013.html）查看服务等级协议(SLA)（https://www.huaweicloud.com/intl/zh-cn/declaration/sla.html）"
    >
      <bk-radio v-model="formData.eip_type" label="5_bgp">全动态BGP</bk-radio>
      <bk-radio v-model="formData.eip_type" label="5_sbgp">优选BGP</bk-radio>
    </bk-form-item>
    <bk-form-item v-if="formData.internet_charge_type === 'postPaid'" label="带宽计费模式" required>
      <bk-radio v-model="formData.bandwidth_option.charge_mode" label="bandwidth">按带宽计费</bk-radio>
      <bk-radio v-model="formData.bandwidth_option.charge_mode" label="traffic">按流量计费</bk-radio>
    </bk-form-item>
    <bk-form-item label="带宽名称" property="bandwidth_option.name" required>
      <bk-input v-model="formData.bandwidth_option.name" />
    </bk-form-item>
    <bk-form-item
      label="带宽大小"
      property="bandwidth_option.size"
      :description="
        formData.internet_charge_type === 'postPaid' && formData.bandwidth_option.charge_mode === 'traffic'
          ? '值的范围是1-300Mbit/s'
          : '值的范围是1-500Mbit/s'
      "
      required
    >
      <bk-compose-form-item>
        <bk-input
          v-model.number="formData.bandwidth_option.size"
          placeholder="请输入带宽大小"
          type="number"
          :min="1"
          :max="
            formData.internet_charge_type === 'postPaid' && formData.bandwidth_option.charge_mode === 'traffic'
              ? 300
              : 500
          "
          class="mr10"
        />
        Mbit/s
      </bk-compose-form-item>
    </bk-form-item>
    <bk-form-item
      v-if="formData.internet_charge_type === 'prePaid'"
      :description="formData.eip_type === '5_bgp' ? '取值范围（月：1-9，年：1）' : '取值范围（月：1-9，年：1-3）'"
      label="购买时长"
      property="internet_charge_prepaid.period_num"
      required
    >
      <bk-compose-form-item>
        <bk-input v-model.number="internet_charge_prepaid.period_num" placeholder="请输入时间" type="number" />
        <bk-select v-model="internet_charge_prepaid.period_type" :clearable="false">
          <bk-option value="month" label="月" />
          <bk-option value="year" label="年" />
        </bk-select>
      </bk-compose-form-item>
    </bk-form-item>
    <bk-form-item v-if="formData.internet_charge_type === 'prePaid'" label="自动续费">
      <bk-checkbox
        v-model="internet_charge_prepaid.is_auto_renew"
        :true-label="true"
        :false-label="false"
      ></bk-checkbox>
    </bk-form-item>
    <bk-form-item label="数量" required>
      {{ formData.eip_count }}
    </bk-form-item>
  </bk-form>
</template>
