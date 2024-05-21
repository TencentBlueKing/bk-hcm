<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { PropType, ref } from 'vue';
import useDetail from '../../../hooks/use-detail';

const props = defineProps({
  id: {
    type: String as PropType<string>,
  },
  vendor: {
    type: String as PropType<string>,
  },
});

const fields = ref([
  {
    name: '实例ID',
    prop: 'cloud_id',
  },
  {
    name: '名称',
    prop: 'cloud_id',
  },
  {
    name: '云厂商',
    prop: 'vendorName',
  },
  {
    name: '架构',
    prop: 'architecture',
  },
  {
    name: '状态',
    prop: 'state',
  },
  {
    name: '类型',
    prop: 'type',
  },
  {
    name: '平台',
    prop: 'platform',
  },
]);

const { loading, detail } = useDetail(
  'images',
  props.id,
  (detail: any) => {
    switch (detail.vendor) {
      case 'tcloud':
        fields.value.push(
          ...[
            {
              name: '地域',
              prop: 'region',
            },
            {
              name: '镜像来源',
              prop: 'image_source',
            },
            {
              name: '镜像大小',
              prop: 'image_size',
            },
          ],
        );
        break;
      // 其它类型的待补充
    }
  },
  props.vendor,
);
</script>

<template>
  <bk-loading :loading="loading">
    <detail-info :detail="detail" :fields="fields" />
  </bk-loading>
</template>
