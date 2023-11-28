// 已收藏方案
export interface ICollectedSchemeItem {
  id: string;
  user: string;
  res_type: string;
  res_id: string;
  creator: string;
  created_at: string;
}

// 方案列表单条数据
export interface ISchemeListItem {
  id: string;
  bk_biz_id: number;
  name: string;
  biz_type: string;
  vendors: string[];
  deployment_architecture: string[];
  cover_ping: number;
  composite_score: number;
  net_score: number;
  cost_score: number;
  cover_rate: number;
  user_distribution: IUserDistributionItem[];
  result_idc_ids: string[];
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export interface IUserDistributionItem {
  name: string;
  children: { name: string; value: number; }[];
}