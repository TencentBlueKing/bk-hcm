<script setup lang="ts">
import { computed, h, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';
import { ILoadBalancerBatchImportModel, LoadBalancerBatchImportPreviewDetails, Status } from '../typings';
import {
  LoadBalancerBatchImportOperationType,
  Scheduler,
  SCHEDULER_NAME,
  SSL_MODE_NAME,
  SSLMode,
} from '@/views/load-balancer/constants';

import Step from '../../step.vue';

defineOptions({ name: 'LbBatchImportPreviewComp' });
const props = defineProps<{
  formModel: ILoadBalancerBatchImportModel;
  data: LoadBalancerBatchImportPreviewDetails;
  isBaseInfoEmpty: boolean;
}>();

const { t } = useI18n();

// info
const info = computed(() => {
  const tmp = { totalCount: 0, executableCount: 0, notExecutableCount: 0, existingCount: 0 };
  props.data?.forEach(({ status }) => {
    if (status === Status.executable) tmp.executableCount += 1;
    else if (status === Status.not_executable) tmp.notExecutableCount += 1;
    else if (status === Status.existing) tmp.existingCount += 1;
    tmp.totalCount += 1;
  });
  return tmp;
});

// table
const pagination = ref({ count: props.data?.length || 0, limit: 10 });
const baseColumns: any[] = [
  { label: t('CLB IP/域名'), field: 'clb_vip_domain' },
  { label: t('CLB ID'), field: 'cloud_clb_id' },
  { label: t('协议类型'), field: 'protocol' },
  {
    label: t('监听器端口'),
    field: 'listener_port',
    render: ({ cell }: { cell: number[] }) => (cell?.length === 2 ? `${cell[0]}-${cell[1]}` : cell[0]),
  },
  {
    label: t('参数校验结果'),
    field: 'validate_result',
    width: 350,
    render: ({ cell, data }: { cell: string[]; data: any }) => {
      const display = cell?.join('; ');
      if (data?.status === Status.executable) return h('span', { class: 'text-success' }, t('校验通过'));
      if (data?.status === Status.existing) return h('span', { class: 'text-warning' }, display);
      return h('span', { class: 'text-danger' }, h('span', { class: 'text-danger' }, display));
    },
  },
];
const columns = ref([]);
// 动态更新表格字段
watch(
  [() => props.formModel.vendor, () => props.formModel.operation_type],
  ([vendor, operationType]) => {
    // 根据不同云厂商、不同操作类型，动态更新表格字段
    const renderColumns = baseColumns.slice();
    switch (vendor) {
      case VendorEnum.TCLOUD:
        if (operationType === LoadBalancerBatchImportOperationType.create_layer4_listener) {
          renderColumns.splice(
            4,
            0,
            {
              label: t('均衡方式'),
              field: 'scheduler',
              isDefaultShow: true,
              render: ({ cell }: { cell: Scheduler }) => SCHEDULER_NAME[cell],
            },
            {
              label: t('健康检查'),
              field: 'health_check',
              isDefaultShow: true,
              render: ({ cell }: { cell: boolean }) => (cell ? '开启' : '关闭'),
            },
            { label: t('会话保持'), field: 'session' },
          );
        } else if (operationType === LoadBalancerBatchImportOperationType.create_layer7_listener) {
          renderColumns.splice(
            4,
            0,
            {
              label: t('证书认证方式'),
              field: 'ssl_mode',
              isDefaultShow: true,
              render: ({ cell }: { cell: SSLMode }) => SSL_MODE_NAME[cell],
            },
            {
              label: () => h('div', { class: 'text-center' }, [t('服务器证书'), h('br'), t('（HTTPS专用）')]),
              field: 'cert_cloud_ids',
              isDefaultShow: true,
              render: ({ cell }: { cell: string[] }) => cell?.join(',') || '--',
            },
            {
              label: () => h('div', { class: 'text-center' }, [t('CA证书'), h('br'), t('（HTTPS专用）')]),
              field: 'ca_cloud_id',
              isDefaultShow: true,
            },
          );
        } else if (operationType === LoadBalancerBatchImportOperationType.create_layer7_rule) {
          renderColumns.splice(
            4,
            0,
            { label: t('域名'), field: 'domain', isDefaultShow: true },
            { label: t('URL'), field: 'url_path', isDefaultShow: true },
            {
              label: t('均衡方式'),
              field: 'scheduler',
              isDefaultShow: true,
              render: ({ cell }: { cell: Scheduler }) => SCHEDULER_NAME[cell],
            },
            {
              label: t('健康检查'),
              field: 'health_check',
              isDefaultShow: true,
              render: ({ cell }: { cell: boolean }) => (cell ? '开启' : '关闭'),
            },
            { label: t('会话保持'), field: 'session' },
          );
        } else if (operationType === LoadBalancerBatchImportOperationType.binding_layer4_rs) {
          renderColumns.splice(
            4,
            0,
            { label: t('RS类型'), field: 'inst_type', isDefaultShow: true },
            { label: t('rsip'), field: 'rs_ip', isDefaultShow: true },
            {
              label: t('rsport'),
              field: 'rs_port',
              isDefaultShow: true,
              render: ({ cell }: { cell: number[] }) => cell?.join(',') || '--',
            },
            { label: t('权重'), field: 'weight', isDefaultShow: true },
          );
        } else {
          renderColumns.splice(
            4,
            0,
            { label: t('域名'), field: 'domain', isDefaultShow: true },
            { label: t('URL'), field: 'url_path', isDefaultShow: true },
            { label: t('RS类型'), field: 'inst_type', isDefaultShow: true },
            { label: t('rsip'), field: 'rs_ip', isDefaultShow: true },
            {
              label: t('rsport'),
              field: 'rs_port',
              isDefaultShow: true,
              render: ({ cell }: { cell: number[] }) => cell?.join(',') || '--',
            },
            { label: t('权重'), field: 'weight', isDefaultShow: true },
          );
        }
        break;
    }
    columns.value = renderColumns;
  },
  { immediate: true },
);

const status = ref<Status>();
const renderData = computed(() => {
  if (!status.value) return props.data || [];
  return props.data?.filter((item) => item.status === status.value) || [];
});
watch(
  renderData,
  (val) => {
    pagination.value.count = val.length;
  },
  { deep: true },
);

defineExpose({ info });
</script>

<template>
  <step :step="3" :title="t('结果预览')">
    <template v-if="data">
      <ul class="info-wrapper">
        <li>
          <span>{{ t('总数：') }}</span>
          <span class="count" @click="status = undefined">{{ info.totalCount }}</span>
        </li>
        <li>
          <span>{{ t('可执行数：') }}</span>
          <span class="count success" @click="status = Status.executable">{{ info.executableCount }}</span>
        </li>
        <li>
          <span>{{ t('不可执行数：') }}</span>
          <span class="count danger" @click="status = Status.not_executable">{{ info.notExecutableCount }}</span>
        </li>
        <li>
          <span>{{ t('已存在：') }}</span>
          <span class="count warning" @click="status = Status.existing">{{ info.existingCount }}</span>
        </li>
      </ul>
      <!-- TODO-CLB：virtual-enabled有高度问题，等组件升级 -->
      <bk-table :data="renderData" :columns="columns" :pagination="pagination" show-overflow-tooltip />
    </template>
    <bk-exception
      v-else-if="isBaseInfoEmpty"
      type="empty"
      description="请录入云账号、云地域、操作类型等信息"
      scene="part"
    />
    <bk-exception v-else type="empty" description="请上传文件" scene="part" />
  </step>
</template>

<style scoped lang="scss">
.info-wrapper {
  display: flex;
  justify-content: flex-end;
  align-items: center;

  li {
    margin-right: 20px;
    line-height: 32px;
    color: #313238;

    &:last-of-type {
      margin-right: 0;
    }

    .count {
      cursor: pointer;

      &.success {
        color: $success-color;
      }

      &.danger {
        color: $danger-color;
      }

      &.warning {
        color: $warning-color;
      }
    }
  }
}
</style>
