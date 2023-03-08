<script lang="ts" setup>
// import { defineComponent, reactive } from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  ref,
} from 'vue';
import FormSelect from '@/views/business/components/form-select.vue';
import { BusinessFormFilter } from '@/typings';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});
const { t } = useI18n();
const type = ref('tcloud');
const handleFormFilter = (value: BusinessFormFilter) => {
  console.log(value);
  type.value = value.vendor;
};
const formData = ref({ vpc: '' });
const accountList = ref([]);
const submit = () => {
  console.log('formData', formData);
};
const {
  datas,
  isLoading,
} = useQueryList(props, 'vpcs');
</script>

<template>
  <div class="business-dialog-warp">
    <form-select @change="handleFormFilter"></form-select>
    <bk-form class="form-subnet">
      <bk-form-item
        :label="t('所属的VPC网络')"
        class="item-warp"
      >
        <bk-select
          :loading="isLoading"
          class="item-warp-component"
          v-model="formData.vpc"
        >
          <bk-option
            v-for="(item, index) in datas"
            :key="index"
            :value="item.id"
            :label="item.name"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('名称')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" v-model="formData.vpc" :placeholder="t('请输入子网名称')" />
      </bk-form-item>
      <bk-form-item
        :label="t('可用区')"
        class="item-warp"
      >
        <!-- <bk-input class="item-warp-component" v-model="formData.vpc" /> -->
        <bk-button-group class="item-button-group">
          <bk-button>
            {{t('出站规则')}}
          </bk-button>
          <bk-button class="item-button">
            {{t('入站规则')}}
          </bk-button>
        </bk-button-group>
      </bk-form-item>
      <bk-form-item
        :label="t('IPv4 CIDR')"
        class="item-warp"
      >
        <bk-input class="item-warp-component" v-model="formData.vpc" :placeholder="t('请输入IP')" />
      </bk-form-item>
      <!-- aws 没有关联路由表 -->
      <bk-form-item
        v-if="type !== 'aws'"
        :label="t('关联路由表')"
        class="item-warp"
      >
        <bk-select
          class="item-warp-component"
          v-model="formData.vpc"
        >
          <bk-option
            v-for="(item, index) in accountList"
            :key="index"
            :value="item.id"
            :label="item.name"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        v-if="type === 'aws' || type === 'huawei' || type === 'azure'"
        :label="t('是否开启IPv6')"
        class="item-warp"
      >
        <bk-switcher
          v-model="formData.vpc"
        />
      </bk-form-item>
      <bk-form-item
        v-if="type === 'azure'"
        :label="t('关联NAT网关')"
        class="item-warp"
      >
        <bk-select
          class="item-warp-component"
          v-model="formData.vpc"
        >
          <bk-option
            v-for="(item, index) in accountList"
            :key="index"
            :value="item.id"
            :label="item.name"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        label-width="50"
        class="item-warp mt40"
      >
        <bk-button class="item-warp-button" theme="primary" @click="submit">
          {{t('提交创建')}}
        </bk-button>
        <bk-button class="ml20 item-warp-button">
          {{t('取消')}}
        </bk-button>
      </bk-form-item>
    </bk-form>
  </div>
</template>
<style lang="scss" scoped>
.form-subnet{
  border-top: 1px solid #C4C6CC;
  padding-top: 20px;
  .item-warp-component{
    width: 200px;
  }
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
