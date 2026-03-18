<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import { Message } from 'bkui-vue';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import BusinessValue from '@/components/display-value/business-value.vue';
import {
  GPU_DEMAND_STATUS,
  GPU_DEMAND_STATUS_MAP,
  useGpuDemandStore,
  type IGpuDemandItem,
  type IGpuDemandSubOrder,
  type ITplConfig,
} from '@/store/resource-plan/gpu-demand';
import { evaluateFormula } from '../hooks/use-excel-preview';
import { useTerminateConfirm } from '../hooks/use-terminate-confirm';
import GpuDemandSlider from '../create/index.vue';
import { QueryRuleOPEnum } from '@/typings';
import QuanbuIcon from '@/assets/image/quanbu.svg';
import RejectCircleIcon from '@/assets/image/reject-circle.svg';
import StatusLoading from '@/assets/image/status_loading.png';

const route = useRoute();

const { isBusinessPage, isServicePage: _isServicePage } = useWhereAmI();

const { setTitle } = useBreadcrumb();
const gpuDemandStore = useGpuDemandStore();
const { confirmTerminateOrder, confirmTerminateSubOrder } = useTerminateConfirm();

// ==================== 数据状态 ====================
const orderId = computed(() => (route.query.id as string) || '');
const orderDetail = ref<IGpuDemandItem | null>(null);
const subOrders = ref<IGpuDemandSubOrder[]>([]);
const tplConfig = ref<ITplConfig | null>(null);

// ==================== 面包屑标题 ====================
const statusText = computed(() => {
  if (!orderDetail.value) return '';
  return GPU_DEMAND_STATUS_MAP[orderDetail.value.status] ?? '';
});

/** 单据状态图标类型映射（复用列表中的 dynamic-status 配置） */
const ORDER_STATUS_ICON_MAP: Record<string, string> = {
  [GPU_DEMAND_STATUS.DONE]: 'success',
  [GPU_DEMAND_STATUS.REJECT]: 'fail',
  [GPU_DEMAND_STATUS.REJECT_ALL]: 'fail',
  [GPU_DEMAND_STATUS.INIT]: 'wait',
  [GPU_DEMAND_STATUS.PENDING]: 'ing',
  [GPU_DEMAND_STATUS.TERMINATE]: 'stop',
};

const orderStatusIcon = computed(() => {
  if (!orderDetail.value) return 'unknown';
  return ORDER_STATUS_ICON_MAP[orderDetail.value.status] || 'unknown';
});

watch(
  [orderId, orderDetail],
  () => {
    if (orderId.value) {
      setTitle('GPU需求详情');
    }
  },
  { immediate: true },
);

// ==================== 评审状态汇总 ====================
const REVIEW_FILTERS = [
  { key: 'all', label: '全部' },
  { key: 'INIT', label: '待评审' },
  { key: 'PENDING', label: '评审中' },
  { key: 'DONE', label: '已评审' },
  { key: 'REJECT', label: '已驳回' },
  { key: 'TERMINATE', label: '已终止' },
] as const;

const activeFilter = ref<string>('all');

/** 子单评审状态 → 中文标签 */
const SUB_STATUS_LABEL: Record<string, string> = {
  INIT: '待评审',
  PENDING: '评审中',
  DONE: '已评审',
  REJECT: '已驳回',
  TERMINATE: '已终止',
};

/** 子单评审状态 → 标签颜色 */
const SUB_STATUS_COLOR: Record<string, { bg: string; text: string }> = {
  INIT: { bg: '#fdeed8', text: '#e38b02' },
  PENDING: { bg: '#e1ecff', text: '#1768ef' },
  DONE: { bg: '#daf6e5', text: '#299e56' },
  REJECT: { bg: '#ffebeb', text: '#e71818' },
  TERMINATE: { bg: '#f0f1f5', text: '#979ba5' },
};

const statusCounts = computed(() => {
  const counts: Record<string, number> = {};
  for (const sub of subOrders.value) {
    counts[sub.status] = (counts[sub.status] || 0) + 1;
  }
  counts.all = subOrders.value.length;
  return counts;
});

// 当前选中的筛选状态计数变为 0 时，自动回退到「全部」
watch(statusCounts, (counts) => {
  if (activeFilter.value !== 'all' && !(counts[activeFilter.value] > 0)) {
    activeFilter.value = 'all';
  }
});

// ==================== Tab 和表格 ====================
const SUMMARY_TAB = '__summary__';
const activeTab = ref(SUMMARY_TAB);

/** tpl_config 中的 sheets 列表 */
const tplSheets = computed(() => tplConfig.value?.sheets ?? []);

/** 从 tpl_config 动态生成的 tab 列表（即使无子单数据也展示） */
const sheetTabs = computed(() => {
  return tplSheets.value.map((sheet) => {
    const sheetSubs = filteredSubOrders.value.filter((s) => s.demand_type === sheet.name);
    const count = sheetSubs.length;
    // 判断该 sheet 中是否存在驳回状态的子单（基于全量子单，不受筛选影响）
    const hasReject = subOrders.value.some((s) => s.demand_type === sheet.name && s.status === 'REJECT');
    return { name: sheet.name, count, hasReject };
  });
});

/** 子单数据按当前 filter 过滤 */
const filteredSubOrders = computed(() => {
  if (activeFilter.value === 'all') return subOrders.value;
  return subOrders.value.filter((s) => s.status === activeFilter.value);
});

// ==================== Sheet 元信息解析 ====================

/**
 * 每个 sheet 的核心元信息：
 * - valueDbField: 第三列 fixed_header 的 db_field（gpu_num 或 qpm_max）
 * - valueFormula: 如果是公式计算列，保存公式
 * - excelFieldToDataIndex: excel 列号 → extension 中的索引映射（仅 headers 部分）
 * - fixedExcelFieldToRawIndex: 整体 excel 列号 → raw 索引（fixed_headers + headers 中所有 field 非 "-" 的列）
 */
interface ISheetMeta {
  sheetName: string;
  valueDbField: 'gpu_num' | 'qpm_max';
  valueFormula?: string;
  /** excel 列号 → 子单 extension 数组中的索引 */
  extensionFieldMap: Record<string, number>;
  /** 所有 excel 列号（fixed_headers + headers）→ 全局 raw 索引，用于公式计算 */
  allFieldToRawIndex: Record<string, number>;
  /** fixed_headers 中有 field 的列数（即 extension 开始之前的偏移量） */
  fixedFieldCount: number;
}

const sheetMetaMap = computed<Map<string, ISheetMeta>>(() => {
  const map = new Map<string, ISheetMeta>();
  for (const sheet of tplSheets.value) {
    const allFieldToRawIndex: Record<string, number> = {};
    let rawIdx = 0;

    // 遍历 fixed_headers，统计有 field 的列
    let fixedFieldCount = 0;
    let valueDbField: 'gpu_num' | 'qpm_max' = 'gpu_num';
    let valueFormula: string | undefined;

    for (const h of sheet.fixed_headers) {
      if (h.field && h.field !== '-') {
        allFieldToRawIndex[h.field] = rawIdx;
        rawIdx += 1;
        fixedFieldCount += 1;
      }
      // 找第三列（通常 db_field 为 gpu_num 或 qpm_max）
      if (h.db_field === 'gpu_num' || h.db_field === 'qpm_max') {
        valueDbField = h.db_field;
        if (h.formula) valueFormula = h.formula;
      }
    }

    // 遍历 headers
    const extensionFieldMap: Record<string, number> = {};
    let extIdx = 0;
    for (const h of sheet.headers) {
      if (h.field && h.field !== '-') {
        allFieldToRawIndex[h.field] = rawIdx;
        extensionFieldMap[h.field] = extIdx;
        rawIdx += 1;
        extIdx += 1;
      }
    }

    map.set(sheet.name, {
      sheetName: sheet.name,
      valueDbField,
      valueFormula,
      extensionFieldMap,
      allFieldToRawIndex,
      fixedFieldCount,
    });
  }
  return map;
});

// ==================== 计算子单的 gpu_num / qpm_max（处理公式） ====================

/**
 * 获取子单的实际数值（gpu_num 或 qpm_max），如果模板中该值是公式计算列，则前端计算
 * 注意：接口返回的 gpu_num / qpm_max 已经是后端计算好的，直接使用即可
 * 但如果后端未计算（值为0但有公式），则前端兜底计算
 */
function getSubOrderValue(sub: IGpuDemandSubOrder, meta: ISheetMeta): number {
  const dbVal = sub[meta.valueDbField] ?? 0;
  // 如果后端已有值，直接返回
  if (dbVal > 0) return dbVal;

  // 后端值为0，且该列有公式，尝试前端计算
  if (meta.valueFormula) {
    const getVal = (excelField: string): number => {
      const rawIdx = meta.allFieldToRawIndex[excelField];
      if (rawIdx === undefined) return 0;
      // raw_data 顺序：fixed_headers 中有 field 的列 + headers 中有 field 的列
      // 对于子单接口，fixed_headers 的数据在子单字段中，headers 的数据在 extension 中
      if (rawIdx < meta.fixedFieldCount) {
        // 从子单的固定字段取值：A→demand_year, B→demand_month, C→gpu_num/qpm_max
        // 但公式列本身就是 C 列，不应该自引用。这里从子单字段映射
        const sheet = tplSheets.value.find((s) => s.name === sub.demand_type);
        if (!sheet) return 0;
        const fixedHeaders = sheet.fixed_headers.filter((h) => h.field && h.field !== '-');
        const fh = fixedHeaders[rawIdx];
        if (fh?.db_field) {
          const val = Number((sub as any)[fh.db_field]);
          return Number.isNaN(val) ? 0 : val;
        }
        return 0;
      }
      // 从 extension 取值
      const extIdx = rawIdx - meta.fixedFieldCount;
      if (extIdx >= 0 && extIdx < (sub.extension?.length ?? 0)) {
        const val = Number(sub.extension[extIdx]);
        return Number.isNaN(val) ? 0 : val;
      }
      return 0;
    };
    const result = evaluateFormula(meta.valueFormula, getVal);
    return Math.ceil(result);
  }
  return dbVal;
}

// ==================== 数据汇总表 ====================

/** 汇总表中收集所有子单涉及的年月列表（排序后） */
const allYearMonths = computed(() => {
  const set = new Set<string>();
  for (const sub of filteredSubOrders.value) {
    const ym = `${sub.demand_year}-${String(sub.demand_month).padStart(2, '0')}`;
    set.add(ym);
  }
  return [...set].sort();
});

interface ISummaryRow {
  demand_type: string;
  gpu_num: number;
  qpm_max: number;
  [monthKey: string]: string | number;
}

const summaryColumns = computed(() => {
  const fixedCols = [
    { field: 'demand_type', label: 'GPU需求类别', minWidth: 150, fixed: 'left' },
    { field: 'gpu_num', label: 'GPU卡数', minWidth: 80, fixed: 'left' },
    { field: 'qpm_max', label: '需求QPM', minWidth: 80, fixed: 'left' },
  ];
  const monthCols = allYearMonths.value.map((ym) => ({
    field: `month_${ym}`,
    label: ym,
    minWidth: 60,
    fixed: undefined as string | undefined,
  }));
  return [...fixedCols, ...monthCols];
});

/**
 * 汇总表行数据：
 * - 每行 = 一个 demand_type（sheet 名称）
 * - GPU卡数 = 该类型所有子单 gpu_num 之和
 * - 需求QPM = 该类型所有子单 qpm_max 之和
 * - GPU卡数 和 需求QPM 互斥：根据该 sheet 的 fixed_headers 第三列 db_field 判断
 * - 年月列 = 按月聚合的 gpu_num 或 qpm_max 总和
 */
const summaryRows = computed<ISummaryRow[]>(() => {
  const rows: ISummaryRow[] = [];
  // 按照 tplSheets 的顺序遍历，即使无子单数据也展示该 sheet
  for (const sheet of tplSheets.value) {
    const meta = sheetMetaMap.value.get(sheet.name);
    if (!meta) continue;

    const sheetSubs = filteredSubOrders.value.filter((s) => s.demand_type === sheet.name);
    const isGpuType = meta.valueDbField === 'gpu_num';

    // 按月聚合
    const monthAgg: Record<string, number> = {};
    let totalGpu = 0;
    let totalQpm = 0;

    for (const sub of sheetSubs) {
      const val = getSubOrderValue(sub, meta);
      if (isGpuType) {
        totalGpu += val;
      } else {
        totalQpm += val;
      }

      const ym = `${sub.demand_year}-${String(sub.demand_month).padStart(2, '0')}`;
      monthAgg[ym] = (monthAgg[ym] || 0) + val;
    }

    const row: ISummaryRow = {
      demand_type: sheet.name,
      gpu_num: isGpuType ? totalGpu : 0,
      qpm_max: isGpuType ? 0 : totalQpm,
    };

    for (const ym of allYearMonths.value) {
      row[`month_${ym}`] = monthAgg[ym] || 0;
    }

    rows.push(row);
  }
  return rows;
});

// ==================== 各 sheet tab 的表格数据 ====================

/**
 * 复用 use-excel-preview 中的列构建思路：
 * 对于子单详情表格，展示 fixed_headers（非 hidden）+ headers（非 hidden）
 * extension 数组对应 headers 中 field 不为 "-" 的列
 */

interface ISheetTableColumn {
  field: string;
  label: string;
  minWidth: number;
  isFixed: boolean;
  dbField?: string;
  formula?: string;
  excelField?: string;
  fixed?: string;
}

const getSheetColumns = (sheetName: string): ISheetTableColumn[] => {
  const sheet = tplSheets.value.find((s) => s.name === sheetName);
  if (!sheet) return [];
  const cols: ISheetTableColumn[] = [];

  // ---- 前置固定列：评审状态 & 评审意见 ----
  cols.push({
    field: '_status',
    label: '评审状态',
    minWidth: 100,
    isFixed: true,
    fixed: 'left',
  });
  cols.push({
    field: '_comment',
    label: '评审意见',
    minWidth: 120,
    isFixed: true,
    fixed: 'left',
  });

  let colIdx = 0;

  // fixed_headers
  for (const h of sheet.fixed_headers) {
    if (!h.hidden) {
      cols.push({
        field: `col_${colIdx}`,
        label: h.name,
        minWidth: 120,
        isFixed: true,
        dbField: h.db_field,
        formula: h.formula,
        excelField: h.field,
      });
      colIdx += 1;
    }
  }
  // headers
  for (const h of sheet.headers) {
    if (!h.hidden) {
      cols.push({
        field: `col_${colIdx}`,
        label: h.name,
        minWidth: 120,
        isFixed: false,
        formula: h.formula,
        excelField: h.field,
      });
      colIdx += 1;
    }
  }
  return cols;
};

/**
 * 构建 sheet 表格行数据
 * - fixed_headers 中有 db_field 的列：从子单字段取值
 * - fixed_headers 中有 formula 的列：前端计算（或从子单字段取已计算的值）
 * - headers 的列：从 extension 数组按顺序取
 */
const getSheetRows = (sheetName: string): Record<string, any>[] => {
  const sheet = tplSheets.value.find((s) => s.name === sheetName);
  if (!sheet) return [];
  const meta = sheetMetaMap.value.get(sheetName);
  if (!meta) return [];

  const sheetSubs = filteredSubOrders.value.filter((s) => s.demand_type === sheetName);

  return sheetSubs.map((sub) => {
    const row: Record<string, any> = {
      _id: sub.id,
      _status: sub.status,
      _comment: Array.isArray(sub.comment) && sub.comment.length > 0 ? sub.comment.join('、') : sub.comment || '-',
    };
    let colIdx = 0;

    // fixed_headers
    for (const h of sheet.fixed_headers) {
      if (!h.hidden) {
        if (h.db_field) {
          let val = (sub as any)[h.db_field];
          // 如果是公式计算列且后端值为0，尝试前端计算
          if (h.formula && (val === 0 || val === undefined || val === null)) {
            val = getSubOrderValue(sub, meta);
          }
          row[`col_${colIdx}`] = val ?? '';
        } else if (h.formula) {
          // 无 db_field 但有 formula 的列
          row[`col_${colIdx}`] = getSubOrderValue(sub, meta);
        } else {
          row[`col_${colIdx}`] = '';
        }
        colIdx += 1;
      }
    }

    // headers：从 extension 按顺序取值
    let extIdx = 0;
    for (const h of sheet.headers) {
      const hasField = h.field && h.field !== '-';
      if (!h.hidden) {
        if (h.formula && !hasField) {
          // 公式计算列，需要用 evaluateFormula
          const getVal = (excelField: string): number => {
            const rawIdx = meta.allFieldToRawIndex[excelField];
            if (rawIdx === undefined) return 0;
            if (rawIdx < meta.fixedFieldCount) {
              const fixedHeaders = sheet.fixed_headers.filter((fh: { field?: string }) => fh.field && fh.field !== '-');
              const fh = fixedHeaders[rawIdx];
              if (fh?.db_field) {
                const v = Number((sub as any)[fh.db_field]);
                return Number.isNaN(v) ? 0 : v;
              }
              return 0;
            }
            const ei = rawIdx - meta.fixedFieldCount;
            if (ei >= 0 && ei < (sub.extension?.length ?? 0)) {
              const v = Number(sub.extension[ei]);
              return Number.isNaN(v) ? 0 : v;
            }
            return 0;
          };
          row[`col_${colIdx}`] = Math.ceil(evaluateFormula(h.formula, getVal));
        } else {
          row[`col_${colIdx}`] = hasField ? sub.extension?.[extIdx] ?? '' : '';
        }
        colIdx += 1;
      }
      // 只有有 excel 列号的 header 才消耗 extension 索引
      if (hasField) extIdx += 1;
    }

    return row;
  });
};

// ==================== 子单操作权限 ====================
const canSubEdit = (row: Record<string, any>) => row._status === 'INIT' || row._status === 'REJECT';
const canSubTerminate = (row: Record<string, any>) => row._status === 'REJECT';

// ==================== 子单操作（待开发） ====================
const handleSubEdit = (row: Record<string, any>) => {
  Message({ theme: 'warning', message: `编辑功能开发中（子单ID: ${row._id}）` });
};

const handleSubReview = (row: Record<string, any>) => {
  Message({ theme: 'warning', message: `评审功能开发中（子单ID: ${row._id}）` });
};

const handleSubReject = (row: Record<string, any>) => {
  Message({ theme: 'warning', message: `驳回功能开发中（子单ID: ${row._id}）` });
};

// ==================== 子单终止（InfoBox 确认弹窗） ====================
const handleSubTerminate = (row: Record<string, any>) => {
  confirmTerminateSubOrder(row._id, fetchDetail);
};

// ==================== 数据加载 ====================
const fetchDetail = async () => {
  if (!orderId.value) return;

  // 获取主单详情
  const detail = await gpuDemandStore.getGpuDemandDetail(orderId.value);
  orderDetail.value = detail;

  // 获取子单列表（真实接口）
  try {
    const data = await gpuDemandStore.getGpuSubOrderList({
      filter: { op: 'and', rules: [{ field: 'order_id', op: QueryRuleOPEnum.EQ, value: orderId.value }] },
      page: { count: false, start: 0, limit: 500 },
    });
    subOrders.value = data.details ?? [];
    if (data.tpl_config?.length > 0) {
      const [firstConfig] = data.tpl_config;
      tplConfig.value = firstConfig;
    }
  } catch {
    subOrders.value = [];
  }

  // 如果是通过"调整"操作进入的，默认切换到第一个 sheet tab（第二个 tab）
  // 需要等 nextTick，确保 tplSheets 变化后 bk-tab-panel 已渲染完成
  if (route.query.action === 'adjust' && tplSheets.value.length > 0) {
    await nextTick();
    activeTab.value = tplSheets.value[0].name;
  }
};

onMounted(fetchDetail);

// ==================== 重新导入弹窗 ====================
const isReimportShow = ref(false);

const handleReimport = () => {
  isReimportShow.value = true;
};

const handleStartReview = async () => {
  if (!orderId.value) return;
  try {
    await gpuDemandStore.batchPendingOrders({ order_ids: [orderId.value] });
    Message({ theme: 'success', message: '转为评审中成功' });
    fetchDetail();
  } catch {
    Message({ theme: 'error', message: '转为评审中失败' });
  }
};

const handleReimportSuccess = () => {
  // 重新导入成功后刷新详情数据
  fetchDetail();
};

// ==================== 重新导入数据权限 ====================
const canReimport = computed(() => {
  if (!orderDetail.value) return false;
  const { status } = orderDetail.value;
  return status === GPU_DEMAND_STATUS.INIT || status === GPU_DEMAND_STATUS.REJECT_ALL;
});

// ==================== 终止主单（InfoBox 确认弹窗） ====================
const canTerminate = computed(() => {
  if (!orderDetail.value) return false;
  const { status } = orderDetail.value;
  if (isBusinessPage) {
    return status === GPU_DEMAND_STATUS.INIT || status === GPU_DEMAND_STATUS.REJECT_ALL;
  }
  // 服务页面
  return status === GPU_DEMAND_STATUS.PENDING;
});

const handleTerminateOrder = () => {
  confirmTerminateOrder(orderId.value, fetchDetail);
};
</script>

<template>
  <div class="gpu-demand-detail">
    <!-- 面包屑头部：单据ID + 状态 -->
    <Teleport defer to="#breadcrumbHead">
      <div class="breadcrumb-head-info" v-if="orderId">
        <span class="breadcrumb-separator">|</span>
        <span class="breadcrumb-order-id">{{ orderId }}</span>
        <template v-if="statusText">
          <span class="breadcrumb-separator">|</span>
          <span class="breadcrumb-status">
            <span :class="['status-icon', orderStatusIcon]" v-if="orderStatusIcon !== 'ing'"></span>
            <img :src="StatusLoading" :class="['status-icon', orderStatusIcon]" alt="icon" v-else />
            <span>{{ statusText }}</span>
          </span>
        </template>
      </div>
    </Teleport>

    <!-- 右上角按钮：通过 Teleport 插入到面包屑区域 -->
    <Teleport defer to="#breadcrumbExtra">
      <div class="breadcrumb-actions">
        <bk-button v-if="isBusinessPage" :disabled="!canReimport" @click="handleReimport">
          <i class="hcm-icon bkhcm-icon-upload mr8"></i>
          重新导入数据
        </bk-button>
        <bk-button v-else :disabled="orderDetail?.status !== GPU_DEMAND_STATUS.INIT" @click="handleStartReview">
          转为评审中
        </bk-button>
        <bk-button :disabled="!canTerminate" @click="handleTerminateOrder">终止</bk-button>
      </div>
    </Teleport>

    <!-- 顶部信息栏 -->
    <div class="detail-header">
      <div class="info-col">
        <span class="info-label">运营产品</span>
        <span class="info-value">{{ orderDetail?.op_product_name ?? '-' }}</span>
      </div>
      <div class="info-col">
        <span class="info-label">业务</span>
        <span class="info-value">
          <BusinessValue v-if="orderDetail?.bk_biz_id" :value="orderDetail.bk_biz_id" />
          <span v-else>-</span>
        </span>
      </div>
      <div class="info-col review-col">
        <span class="info-label">评审状态汇总</span>
        <div class="filter-tags">
          <div
            v-for="filter in REVIEW_FILTERS"
            v-show="filter.key === 'all' || (statusCounts[filter.key] || 0) > 0"
            :key="filter.key"
            class="filter-tag"
            :class="[`filter-tag--${filter.key.toLowerCase()}`, { active: activeFilter === filter.key }]"
            @click="activeFilter = filter.key"
          >
            <img
              v-if="filter.key === 'REJECT' && (statusCounts['REJECT'] || 0) > 0"
              :src="RejectCircleIcon"
              alt="reject"
              class="filter-reject-icon"
            />
            <span class="filter-tag-text">{{ filter.label }}：{{ statusCounts[filter.key] || 0 }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab 表格区域 -->
    <div class="detail-table-panel">
      <bk-tab v-model:active="activeTab" type="card-tab">
        <!-- 数据汇总 Tab（固定在第一个） -->
        <bk-tab-panel :name="SUMMARY_TAB">
          <template #label>
            <div class="tab-label summary-label">
              <img :src="QuanbuIcon" alt="quanbu" class="tab-icon" />
              <span>数据汇总</span>
            </div>
          </template>
        </bk-tab-panel>
        <!-- 动态 sheet Tab -->
        <bk-tab-panel v-for="tab in sheetTabs" :key="tab.name" :name="tab.name">
          <template #label>
            <div class="tab-label">
              <img v-if="tab.hasReject" :src="RejectCircleIcon" alt="reject" class="tab-reject-icon" />
              <span>{{ tab.name }}</span>
              <span class="tab-count">{{ tab.count }}</span>
            </div>
          </template>
        </bk-tab-panel>
      </bk-tab>

      <!-- 数据汇总表 -->
      <bk-table
        v-if="activeTab === SUMMARY_TAB"
        :data="summaryRows"
        :max-height="500"
        row-hover="auto"
        show-overflow-tooltip
        :border="['row']"
        class="table-container"
      >
        <bk-table-column
          v-for="col in summaryColumns"
          :key="col.field"
          :prop="col.field"
          :label="col.label"
          :min-width="col.minWidth"
          :fixed="col.fixed"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span v-if="col.field === 'gpu_num' || col.field === 'qpm_max'">
              {{ row[col.field] > 0 ? row[col.field] : '-' }}
            </span>
            <span v-else-if="col.field.startsWith('month_')">
              {{ row[col.field] > 0 ? row[col.field] : '-' }}
            </span>
            <span v-else>{{ row[col.field] }}</span>
          </template>
        </bk-table-column>
      </bk-table>

      <!-- 各 sheet 的子单表格 -->
      <template v-for="tab in sheetTabs" :key="tab.name">
        <bk-table
          v-if="activeTab === tab.name"
          :data="getSheetRows(tab.name)"
          :max-height="500"
          row-hover="auto"
          show-overflow-tooltip
          :border="['row']"
          class="table-container"
        >
          <bk-table-column
            v-for="col in getSheetColumns(tab.name)"
            :key="col.field"
            :prop="col.field"
            :label="col.label"
            :min-width="col.minWidth"
            :fixed="col.fixed"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <span
                v-if="col.field === '_status'"
                class="status-tag"
                :style="{
                  background: SUB_STATUS_COLOR[row._status]?.bg || '#f0f1f5',
                  color: SUB_STATUS_COLOR[row._status]?.text || '#979ba5',
                }"
              >
                {{ SUB_STATUS_LABEL[row._status] || row._status }}
              </span>
              <span v-else>{{ row[col.field] }}</span>
            </template>
          </bk-table-column>
          <!-- 操作列 -->
          <bk-table-column label="操作" fixed="right" :width="isBusinessPage ? 120 : 160">
            <template #default="{ row }">
              <div class="sub-order-actions">
                <bk-button theme="primary" text :disabled="!canSubEdit(row)" @click="handleSubEdit(row)">
                  编辑
                </bk-button>
                <template v-if="isBusinessPage">
                  <bk-button theme="primary" text :disabled="!canSubTerminate(row)" @click="handleSubTerminate(row)">
                    终止
                  </bk-button>
                </template>
                <template v-else>
                  <bk-button theme="primary" text @click="handleSubReview(row)">评审</bk-button>
                  <bk-button theme="primary" text @click="handleSubReject(row)">驳回</bk-button>
                </template>
              </div>
            </template>
          </bk-table-column>
        </bk-table>
      </template>
    </div>
  </div>
  <GpuDemandSlider
    v-model="isReimportShow"
    mode="reimport"
    :order-detail="orderDetail"
    @success="handleReimportSuccess"
  />
</template>

<style lang="scss" scoped>
.gpu-demand-detail {
  height: 100%;
  padding: 0 24px 24px;
}

.breadcrumb-actions {
  display: flex;
  align-items: center;
  gap: 8px;

  i {
    font-size: 12px;
    color: #4d4f56;
  }

  :deep(.bk-button.is-disabled) i {
    color: inherit;
  }
}

.breadcrumb-head-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  margin-left: -8px;

  .breadcrumb-separator {
    color: #dcdee5;
  }

  .breadcrumb-order-id {
    color: #979ba5;
  }

  .breadcrumb-status {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #313238;
    font-size: 12px;

    .status-icon {
      width: 8px;
      height: 8px;
      border-radius: 50%;

      &.success {
        background: #cbf0da;
        border: 1px solid #2caf5e;
      }

      &.fail {
        background: #fdd;
        border: 1px solid #ea3636;
      }

      &.wait {
        background: #fce5c0;
        border: 1px solid #f59500;
      }

      &.stop {
        background: #f0f1f5;
        border: 1px solid #c4c6cc;
      }

      &.ing {
        width: 10px;
        height: 10px;
        animation: spin 2s linear infinite;
      }
    }
  }
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }

  100% {
    transform: rotate(360deg);
  }
}

.detail-header {
  display: flex;
  align-items: flex-start;
  gap: 48px;
  padding: 16px 0;
}

.info-col {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;

  .info-label {
    font-size: 12px;
    line-height: 32px;
    color: #4d4f56;
    white-space: nowrap;
  }

  .info-value {
    font-size: 12px;
    line-height: 20px;
    color: #313238;
  }
}

.review-col {
  flex-shrink: 1;
  min-width: 0;
}

// ==================== 筛选标签组 ====================
.filter-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-tag {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  font-size: 12px;
  line-height: 20px;
  border: 1px solid transparent;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
  outline: none;
  gap: 4px;

  .filter-reject-icon {
    width: 14px;
    height: 14px;
  }

  .filter-tag-text {
    font-size: 12px;
    line-height: 20px;
  }

  // ---------- 各状态颜色 ----------
  // 全部：白色背景
  &--all {
    background: #fff;

    .filter-tag-text {
      color: #4d4f56;
    }

    &.active {
      border-color: #dcdee5;
    }
  }

  // 待评审
  &--init {
    background: #fdeed8;

    .filter-tag-text {
      color: #e38b02;
    }

    &.active {
      border-color: #f5c68a;
    }
  }

  // 评审中
  &--pending {
    background: #e1ecff;

    .filter-tag-text {
      color: #1768ef;
    }

    &.active {
      border-color: #a3c5fd;
    }
  }

  // 已评审
  &--done {
    background: #daf6e5;

    .filter-tag-text {
      color: #299e56;
    }

    &.active {
      border-color: #a1e3ba;
    }
  }

  // 已驳回
  &--reject {
    background: #ffebeb;

    .filter-tag-text {
      color: #e71818;
    }

    &.active {
      border-color: #f8b4b4;
    }
  }

  // 已终止
  &--terminate {
    background: #f0f1f5;

    .filter-tag-text {
      color: #979ba5;
    }

    &.active {
      border-color: #c4c6cc;
    }
  }
}

// ==================== Tab 表格区域 ====================
.detail-table-panel {
  background: #fff;
  border-radius: 2px;
  box-shadow: 0 2px 4px 0 #1919290d;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;

  .tab-icon {
    width: 14px;
    height: 14px;
  }

  .tab-reject-icon {
    width: 14px;
    height: 14px;
  }

  .tab-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    padding: 0 6px;
    font-size: 10px;
    line-height: 18px;
    color: #4d4f56;
    background: #fff;
    height: 16px;
    border-radius: 8px;
  }
}

.summary-label {
  .tab-icon {
    margin-right: 2px;
  }
}

.table-container {
  padding: 24px;
  padding-top: 4px;
}

.status-tag {
  display: inline-block;
  padding: 0 8px;
  font-size: 12px;
  line-height: 22px;
  border-radius: 2px;
  white-space: nowrap;
}

.sub-order-actions {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}
</style>
