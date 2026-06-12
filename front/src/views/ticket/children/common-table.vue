<template>
  <CommonTable />
</template>

<script setup lang="ts">
import { h, inject } from 'vue';
import { APPLICATION_STATUS_MAP, APPLICATION_TYPE_MAP, searchData } from '../constants';
import { useRoute, useRouter } from 'vue-router';
import { Button } from 'bkui-vue';
import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusLoading from '@/assets/image/status_loading.png';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import { Spinner } from 'bkui-vue/lib/icon';
import { timeFormatter } from '@/common/util';
import { useTable } from '@/hooks/useTable/useTable';
import type { RulesItem } from '@/typings';
import useSearchUser from '@/hooks/use-search-user';
import { MENU_SERVICE_TICKET_DETAILS, MENU_BUSINESS_TICKET_DETAILS } from '@/constants/menu-symbol';
import UserValue from '@/components/display-value/user-value.vue';

interface IProps {
  rules: RulesItem[];
}

const props = withDefaults(defineProps<IProps>(), {});

const router = useRouter();
const route = useRoute();
const isBusinessPage = inject<boolean>('isBusinessPage');

const { search: searchUser } = useSearchUser();

const columns = [
  // {
  //   label: '申请ID',
  //   field: 'id',
  // },
  // {
  //   label: '来源',
  //   field: 'source',
  // },
  {
    label: '单号',
    field: 'sn',
    render: ({ data }: any) => {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick: () => {
            router.push({
              name: isBusinessPage ? MENU_BUSINESS_TICKET_DETAILS : MENU_SERVICE_TICKET_DETAILS,
              query: {
                ...route.query,
                id: data.id,
              },
            });
          },
        },
        data.sn,
      );
    },
  },
  {
    label: '申请类型',
    field: 'operation',
    render: ({ cell }: { cell: string }) => APPLICATION_TYPE_MAP[cell],
  },
  {
    label: '单据状态',
    field: 'status',
    render({ cell }: any) {
      let icon = StatusAbnormal;
      const txt = APPLICATION_STATUS_MAP[cell];
      switch (cell) {
        case 'pending':
        case 'delivering':
          icon = StatusLoading;
          break;
        case 'pass':
        case 'completed':
        case 'deliver_partial':
          icon = StatusSuccess;
          break;
        case 'rejected':
        case 'cancelled':
        case 'deliver_error':
          icon = StatusFailure;
          break;
      }
      return h('div', { class: 'cvm-status-container' }, [
        icon === StatusLoading
          ? h(Spinner, { fill: '#3A84FF', class: 'mr6', width: 14, height: 14 })
          : h('img', { src: icon, class: 'mr6', width: 14, height: 14 }),
        txt,
      ]);
    },
  },
  {
    label: '申请人',
    field: 'applicant',
    render: ({ cell }: { cell: string }) => h(UserValue, { value: cell }),
  },
  {
    label: '创建时间',
    field: 'created_at',
    render({ cell }: any) {
      return timeFormatter(cell);
    },
  },
  {
    label: '更新时间',
    field: 'updated_at',
    render({ cell }: any) {
      return timeFormatter(cell);
    },
  },
  {
    label: '备注',
    field: 'memo',
    render({ cell }: any) {
      return cell || '--';
    },
  },
];
const { CommonTable } = useTable({
  searchOptions: {
    searchData,
    extra: {
      getMenuList: async (item: ISearchItem, keyword: string): Promise<ISearchItem[]> => {
        const { id, async, children = [] } = item;

        if (!async) {
          return children;
        }

        if (keyword?.length < 2) {
          return [];
        }

        if (id === 'applicant') {
          const result = await searchUser(keyword);
          return result;
        }
      },
    },
  },
  tableOptions: {
    columns,
  },
  requestOption: {
    type: 'applications',
    filterOption: { rules: props.rules },
  },
});
</script>
