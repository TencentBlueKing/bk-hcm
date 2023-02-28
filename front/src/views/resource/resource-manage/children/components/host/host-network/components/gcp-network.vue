<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import {
  ref,
} from 'vue';

const { t } = useI18n();
const showBind = ref(false);
const columns = [
  {
    label: '接口名称',
    field: 'id',
  },
  {
    label: '内网IP',
    field: 'id',
  },
  {
    label: '公网IP',
    field: 'id',
  },
  {
    label: '所属网络(VPC)',
    field: 'id',
  },
  {
    label: '所属子网',
    field: 'id',
  },
  {
    label: '网络层级',
    field: 'id',
  },
  {
    label: 'IP转发',
    field: 'id',
  },
];
const tableData = [
  {
    id: 233,
  },
];


const handleToggleShow = () => {
  showBind.value = !showBind.value;
};

const handleConfirmBind = () => {
  handleToggleShow();
};

const handleRadio = (item: any) => {
  console.log(item);
};


</script>

<template>
  <bk-button
    class="mt20"
    theme="primary"
    @click="handleToggleShow"
  >
    {{ t('绑定已有网络接口') }}
  </bk-button>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="tableData"
  />


  <bk-dialog
    :is-show="showBind"
    width="620"
    title="绑定虚拟IP"
    theme="primary"
    quick-close
    @closed="handleToggleShow"
    @confirm="handleConfirmBind">
    <bk-table
      class="mt20"
      dark-header
      :data="[{ ip: 'testetstt' }]"
      :outer-border="false"
    >
      <bk-table-column
        label="内网IP"
      >
        <template #default="{ data } ">
          <div class="cell-flex">
            <bk-radio
              label="" @click="() => {
                handleRadio(data)
              }" />
            <span class="pl10">{{ data.ip }}</span>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column
        label="已绑定的EIP"
        prop="ip"
      />
    </bk-table>
  </bk-dialog>
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin: 20px 0 5px;
  }
  .sub-title{
    font-size: 12px;
  }
  .cell-flex{
    display: flex;
    align-items: center;
  }
  .table-warp{
    padding: 20px;
    border: 1px dashed rgb(225, 221, 221);
    .table-flex{
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }
  .flex{
    display: flex;
    align-items: center;
  }
</style>
