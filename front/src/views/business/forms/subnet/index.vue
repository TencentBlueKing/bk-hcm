<script lang="ts" setup>
// import { defineComponent, reactive } from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import {
  Message,
} from 'bkui-vue';
import {
  reactive,
  ref,
} from 'vue';
import {
  useBusinessStore,
  useAccountStore,
} from '@/store';
import FormSelect from '@/views/business/components/form-select.vue';
import { BusinessFormFilter } from '@/typings';
import VpcSelector from '@/components/vpc-selector/index.vue';
import ZoneSelector from '@/components/zone-selector/index.vue';
import RouteTableSelector from '@/components/route-table-selector/index.vue';
import ResourceGroupSelector from '@/components/resource-group/index.vue';

const emit = defineEmits(['cancel', 'success']);
const { t } = useI18n();
const formData = reactive<any>({});
const businessStore = useBusinessStore();
const accountStore = useAccountStore();
const submitLoading = ref(false);
const handleFormFilter = (data: BusinessFormFilter) => {
  console.log(data);
  formData.vendor = data.vendor;
  formData.region = data.region;
  formData.account_id = data.account_id;
};


// 提交
const submit = async () => {
  console.log('formData', formData);
  submitLoading.value = true;
  try {
    await businessStore.createSubnet(accountStore.bizs, formData);
    Message({
      message: t('新增子网成功'),
      theme: 'success',
    });
    emit('success');
  } catch (error) {
    console.log(error);
  } finally {
    submitLoading.value = false;
  }
};

// 方法
const cancel = async () => {
  emit('cancel');
};

</script>

<template>
  <div class="business-dialog-warp">
    <form-select @change="handleFormFilter"></form-select>
    <bk-form class="form-subnet">
      <bk-form-item
        :label="t('名称')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" v-model="formData.name" :placeholder="t('请输入子网名称')" />
      </bk-form-item>
      <bk-form-item
        :label="t('所属的VPC')"
        class="item-warp"
      >
        <vpc-selector :vendor="formData.vendor" v-model="formData.cloud_vpc_id"></vpc-selector>
      </bk-form-item>
      <bk-form-item
        :label="t('可用区')"
        class="item-warp"
        v-if="formData.vendor === 'tcloud' || formData.vendor === 'aws' || formData.vendor === 'huawei'"
      >
        <zone-selector
          :vendor="formData.vendor"
          :region="formData.region"
          v-model="formData.zone"></zone-selector>
      </bk-form-item>
      <!-- <bk-form-item
        :label="t('资源组')"
        class="item-warp"
      >
        <resource-group-selector v-model="formData.resource"></resource-group-selector>
      </bk-form-item> -->
      <bk-form-item
        :label="t('IPv4 CIDR')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" v-model="formData.ipv4_cidr" :placeholder="t('请输入IPV4')" />
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'aws' || formData.vendor === 'azure'"
        :label="t('IPv6 CIDR')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" v-model="formData.ipv6_cidr" :placeholder="t('请输入IPV6')" />
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'tcloud' || formData.vendor === 'azure'"
        :label="t('关联路由表')"
        class="item-warp"
      >
        <route-table-selector v-model="formData.cloud_route_table_id"></route-table-selector>
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'huawei'"
        :label="t('网关地址')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" v-model="formData.gateway_ip" :placeholder="t('请输入网关地址')" />
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'huawei'"
        :label="t('是否启用IPv6')"
        class="item-warp"
      >
        <bk-switcher
          v-model="formData.ipv6_enable"
        />
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'gcp'"
        :label="t('专用访问通道')"
        class="item-warp"
      >
        <bk-switcher
          disabled
          v-model="formData.private_ip_google_access"
        />
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'gcp'"
        :label="t('是否启用流日志')"
        class="item-warp"
      >
        <bk-switcher
          disabled
          v-model="formData.enable_flow_logs"
        />
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'azure'"
        :label="t('资源组')"
        class="item-warp"
      >
        <resource-group-selector
          v-model="formData.resource_group"></resource-group-selector>
      </bk-form-item>
      <bk-form-item
        :label="t('备注')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" type="textarea" v-model="formData.memo" :placeholder="t('请输入备注')" />
      </bk-form-item>
      <bk-form-item
        :label="t('备注')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" type="textarea" v-model="formData.memo" :placeholder="t('请输入备注')" />
      </bk-form-item>
      <bk-form-item
        label-width="50"
        class="item-warp mt40"
      >
        <bk-button
          class="item-warp-button" theme="primary"
          :disabled="!formData.vendor" @click="submit"
          :loading="submitLoading">
          {{t('提交创建')}}
        </bk-button>
        <bk-button class="ml20 item-warp-button" @click="cancel">
          {{t('取消')}}
        </bk-button>
      </bk-form-item>
    </bk-form>
  </div>
</template>
<style lang="scss" scoped>
.form-subnet{
  padding-right: 20px;
  .item-warp-button{
    width: 100px;
  }
  .item-button-group{
    .item-button{
      margin-left: 10px;
    }
  }
}
</style>
