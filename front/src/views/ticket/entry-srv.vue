<template>
  <div class="tab-container">
    <bk-tab type="card-grid" v-model:active="applyType" class="header-tab" @update:active="saveActiveType">
      <bk-tab-panel v-for="(item, index) in tabList" :name="item.name" :label="item.label" :key="index">
        <component
          v-if="item.name === applyType"
          :is="item.Component"
          :rules="item.rules"
          v-bind="item.props"
        ></component>
      </bk-tab-panel>
    </bk-tab>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ApplicationsType } from './typings';
import CommonTable from './children/common-table.vue';
import { QueryRuleOPEnum } from '@/typings';

const router = useRouter();
const route = useRoute();
const { t } = useI18n();

const applyType = ref(route.query?.type || 'all');

const saveActiveType = (val: string) => {
  router.replace({ query: { type: val } });
};

const tabList = ref<ApplicationsType[]>([
  {
    label: t('全部'),
    name: 'all',
    rules: [],
    Component: CommonTable,
  },
  {
    label: t('云主机'),
    name: 'cloudMachines',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_cvm'],
      },
    ],
    Component: CommonTable,
  },
  {
    label: t('账号'),
    name: 'account',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['add_account', 'create_main_account', 'update_main_account'],
      },
    ],
    Component: CommonTable,
  },
  {
    label: t('硬盘'),
    name: 'disk',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_disk'],
      },
    ],
    Component: CommonTable,
  },
  {
    label: t('VPC'),
    name: 'vpc',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_disk'],
      },
    ],
    Component: CommonTable,
  },
  {
    label: '安全组',
    name: 'securityGroup',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          'create_security_group',
          'update_security_group',
          'delete_security_group',
          'associate_security_group',
          'disassociate_security_group',
          'create_security_group_rule',
          'update_security_group_rule',
          'delete_security_group_rule',
        ],
      },
    ],
    Component: CommonTable,
  },
  {
    label: '负载均衡',
    name: 'load_balancer',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_load_balancer'],
      },
    ],
    Component: CommonTable,
  },
]);
</script>

<style lang="scss" scoped>
.tab-container {
  height: 100%;
  padding: 24px;
}

:global(.bk-tab) {
  height: 100%;

  :global(.bk-tab-content) {
    height: calc(100% - 40px);
    padding: 0;
  }
}
</style>
