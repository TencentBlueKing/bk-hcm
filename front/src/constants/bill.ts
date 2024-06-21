// 账单类型 - 华为
export const BILL_TYPE__MAP_HW = {
  1: '消费-新购',
  2: '消费-续订',
  3: '消费-变更',
  4: '退款-退订',
  5: '消费-使用',
  8: '消费-自动续订',
  9: '调账-补偿',
  14: '消费-服务支持计划月末扣费',
  15: '消费-税金',
  16: '调账-扣费',
  17: '消费-保底差额',
  20: '退款-变更',
  23: '消费-节省计划抵扣',
  24: '退款-包年/包月转按需',
  100: '退款-退订税金',
  101: '调账-补偿税金',
  102: '调账-扣费税金',
};

// 调账状态
export const BILL_ADJUSTMENT_STATE__MAP = {
  unconfirmed: '未确认',
  confirmed: '已确认',
};

// 调账类型
export const BILL_ADJUSTMENT_TYPE__MAP = {
  increase: '增加',
  decrease: '减少',
};

// 币种
export const CURRENCY_MAP = {
  USD: '美元',
  RMB: '人民币',
};