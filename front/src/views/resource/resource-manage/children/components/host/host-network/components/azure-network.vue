<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import type {
  // PlainObject,
  FilterType,
} from '@/typings/resource';
import {
  ref,
  h,
  PropType,
} from 'vue';
import {
// Button,
// InfoBox,
} from 'bkui-vue';
import {
  useResourceStore,
} from '@/store/resource';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  data: {
    type: Object,
  },
});

const { t } = useI18n();
const resourceStore = useResourceStore();
const showBind = ref(false);
const tableData = ref<any>([]);
const isLoading = ref(false);

const columns = [
  {
    label: t('接口ID'),
    field: 'id',
  },
  {
    label: t('类型'),
    field: 'id',
    render({ data }: any) {
      return [
        h(
          'span',
          {},
          [
            data?.extension?.type || '--',
          ],
        ),
      ];
    },
  },
  {
    label: t('内网IP'),
    render({ data }: any) {
      return [
        h(
          'span',
          {},
          [
            data.private_ipv4.join(',') || data.private_ipv6.join(',') || '--',
          ],
        ),
      ];
    },
  },
  {
    label: t('普通公网IP/EIP'),
    field: 'public_ip',
    render({ data }: any) {
      return [
        h(
          'span',
          {},
          [
            data.public_ipv4.join(',') || data.public_ipv6.join(',') || '--',
          ],
        ),
      ];
    },
  },
  {
    label: t('地域'),
    field: 'region',
  },
  {
    label: t('备注'),
    field: 'id',
  },
  // {
  //   label: '操作',
  //   render() {
  //     return [
  //       h(
  //         Button,
  //         {
  //           text: true,
  //           theme: 'primary',
  //           class: 'mr10',
  //           onClick() {
  //             handleFreedIp();
  //           },
  //         },
  //         [
  //           '解绑',
  //         ],
  //       ),
  //     ];
  //   },
  // },
];


const handleToggleShow = () => {
  showBind.value = !showBind.value;
};

const handleConfirmBind = () => {
  handleToggleShow();
};

// const handleFreedIp = () => {
//   InfoBox({
//     title: '确定解绑此网络接口',
//     subTitle: '解绑网络接口',
//     headerAlign: 'center',
//     footerAlign: 'center',
//     contentAlign: 'center',
//     onConfirm() {
//       console.log('111');
//     },
//   });
// };

const handleRadio = (item: any) => {
  console.log(item);
};

const getNetWorkList = async () => {
  isLoading.value = true;
  try {
    const type = props.data.vendor;
    const { id } = props.data;
    const res = await resourceStore.getNetworkList(type, id);
    console.log('res', res);
    tableData.value = res.data;
  } catch (error) {
    console.log(error);
  } finally {
    isLoading.value = false;
  }
};

getNetWorkList();


</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <!-- <bk-button
      class="mt20"
      theme="primary"
      @click="handleToggleShow"
    >
      {{ t('绑定已有网络接口') }}
    </bk-button> -->
    <bk-table
      row-hover="auto"
      :columns="columns"
      :data="tableData"
      show-overflow-tooltip
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
        :columns="columns"
        show-overflow-tooltip
      >
        <bk-table-column
          label="内网IP"
        >
          <div class="cell-flex">
            <bk-radio
              label="" @click="() => {
                handleRadio(data)
              }" />
            <span class="pl10">{{ data.ip }}</span>
          </div>
        </bk-table-column>
        <bk-table-column
          label="已绑定的EIP"
          prop="ip"
        />
      </bk-table>
    </bk-dialog>
  </bk-loading>
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin-bottom: 8px;
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
