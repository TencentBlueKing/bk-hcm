<script lang="ts" setup>
// import { defineComponent, reactive } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { reactive, ref, watch } from 'vue';
import { useBusinessStore, useAccountStore, useResourceStore } from '@/store';
import FormSelect from '@/views/business/components/form-select.vue';
import { BusinessFormFilter } from '@/typings';
import VpcSelector from '@/components/vpc-selector/index.vue';
import ZoneSelector from '@/components/zone-selector/index.vue';
import RouteTableSelector from '@/components/route-table-selector/index.vue';
import ResourceGroupSelector from '@/components/resource-group/index.vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const emit = defineEmits(['cancel', 'success']);
const { t } = useI18n();
const formData = reactive<any>({ ipv4_cidr: '' });
const businessStore = useBusinessStore();
const accountStore = useAccountStore();
const resourceStore = useResourceStore();
const submitLoading = ref(false);
const ipv4Cidr = ref('');
const isVPCSupportedIpv6 = ref(false);
const isNeedIPV6 = ref(false);
const { isResourcePage } = useWhereAmI();

const handleFormFilter = (data: BusinessFormFilter) => {
  formData.vendor = data.vendor;
  formData.region = data.region;
  formData.account_id = data.account_id;
};

// 提交
const submit = async () => {
  console.log('formData', formData);
  submitLoading.value = true;
  try {
    await businessStore.createSubnet(accountStore.bizs, formData, isResourcePage);
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

const getVpcDetail = async (vpcId: string) => {
  console.log('vpcId', vpcId);
  if (!vpcId) return;
  const res = await resourceStore.detail('vpcs', vpcId);
  ipv4Cidr.value = res?.data?.extension?.cidr
    ?.filter((e: any) => e.type === 'ipv4')
    ?.map((item: any) => item.cidr)
    ?.join('｜');
  isVPCSupportedIpv6.value = res?.data?.extension?.cidr?.some((e: any) => e.type === 'ipv6');
};

// 方法
const cancel = async () => {
  emit('cancel');
};

watch(
  () => formData.vendor,
  (val) => {
    if (val === 'azure') {
      formData.ipv4_cidr = [];
    } else {
      formData.ipv4_cidr = '';
    }
  },
);
</script>

<template>
  <div class="business-dialog-warp">
    <form-select @change="handleFormFilter"></form-select>
    <bk-form class="form-subnet">
      <bk-form-item :label="t('名称')" class="item-warp">
        <bk-input class="item-warp-component" v-model="formData.name" :placeholder="t('请输入子网名称')" />
      </bk-form-item>
      <bk-form-item v-if="formData.vendor === 'azure'" :label="t('资源组')" class="item-warp">
        <resource-group-selector v-model="formData.resource_group" :vendor="formData.vendor"></resource-group-selector>
      </bk-form-item>
      <bk-form-item :label="t('所属的VPC')" class="item-warp">
        <vpc-selector
          :vendor="formData.vendor"
          :region="formData.region"
          v-model="formData.cloud_vpc_id"
          @handleVpcDetail="getVpcDetail"
        ></vpc-selector>
      </bk-form-item>
      <bk-form-item
        :label="t('可用区')"
        class="item-warp"
        v-if="formData.vendor === 'tcloud' || formData.vendor === 'aws'"
      >
        <zone-selector :vendor="formData.vendor" :region="formData.region" v-model="formData.zone"></zone-selector>
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
        :description="ipv4Cidr ? `请输入在${ipv4Cidr}中的CIDR` : ''"
      >
        <bk-tag-input
          v-if="formData.vendor === 'azure'"
          v-model="formData.ipv4_cidr"
          allow-create
          allow-auto-match
          :list="[]"
          :placeholder="t('请输入IPV4')"
        />
        <bk-input v-else class="item-warp-component" v-model="formData.ipv4_cidr" :placeholder="t('请输入IPV4 CIDR')" />
      </bk-form-item>
      <bk-form-item v-if="isVPCSupportedIpv6" label="是否开启 IPV6" class="item-wrap">
        <bk-switcher v-model="isNeedIPV6" />
      </bk-form-item>
      <bk-form-item
        v-if="formData.vendor === 'aws' && isVPCSupportedIpv6 && isNeedIPV6"
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
        <route-table-selector
          :cloud-vpc-id="formData.cloud_vpc_id"
          v-model="formData.cloud_route_table_id"
        ></route-table-selector>
      </bk-form-item>
      <bk-form-item v-if="formData.vendor === 'huawei'" :label="t('网关地址')" class="item-warp">
        <bk-input class="item-warp-component" v-model="formData.gateway_ip" :placeholder="t('请输入网关地址')" />
      </bk-form-item>
      <bk-form-item v-if="formData.vendor === 'huawei'" :label="t('是否启用IPv6')" class="item-warp">
        <bk-switcher v-model="formData.ipv6_enable" />
      </bk-form-item>
      <bk-form-item v-if="formData.vendor === 'gcp'" :label="t('专用访问通道')" class="item-warp">
        <bk-switcher disabled v-model="formData.private_ip_google_access" />
      </bk-form-item>
      <bk-form-item v-if="formData.vendor === 'gcp'" :label="t('是否启用流日志')" class="item-warp">
        <bk-switcher disabled v-model="formData.enable_flow_logs" />
      </bk-form-item>
      <bk-form-item :label="t('备注')" class="item-warp">
        <bk-input
          class="item-warp-component"
          type="textarea"
          v-model="formData.memo"
          :resize="false"
          :placeholder="t('请输入备注')"
        />
      </bk-form-item>
      <bk-form-item label-width="50" class="item-warp mt40">
        <bk-button
          class="item-warp-button"
          theme="primary"
          :disabled="!formData.vendor"
          @click="submit"
          :loading="submitLoading"
        >
          {{ t('提交创建') }}
        </bk-button>
        <bk-button class="ml20 item-warp-button" @click="cancel">
          {{ t('取消') }}
        </bk-button>
      </bk-form-item>
    </bk-form>
  </div>
</template>
<style lang="scss" scoped>
.form-subnet {
  padding-right: 20px;
  .item-warp-button {
    width: 100px;
  }
  .item-button-group {
    .item-button {
      margin-left: 10px;
    }
  }
}
</style>
