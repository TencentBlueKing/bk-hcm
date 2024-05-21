import { VendorEnum } from '@/common/constant';
import { IListResData, IQueryResData } from '@/typings';

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

// 方案编辑数据
export interface ISchemeEditingData {
  id?: string;
  bk_biz_id: number;
  name: string;
}

// 方案切换下拉框单条数据
export interface ISchemeSelectorItem {
  id: string | number;
  name: string;
  bk_biz_id: number | string;
  deployment_architecture: string[];
  vendors: string[];
  composite_score: number;
  net_score: number;
  cost_score: number;
}

// idc机房延迟数据列表单条数据
export interface IIdcLatencyListItem {
  name: string;
  children: { name: string; value: { [key: string]: number } }[];
}

export interface IUserDistributionItem {
  name: string;
  children: { name: string; value: number }[];
}

/**
 * 查询国家列表
 */
export type ICountriesListResData = IListResData<Array<string>>;

/**
 * 查询云选型用户分布占比
 */
export type IUserDistributionResData = IQueryResData<Array<IAreaInfo>>;

/**
 * 查询业务类型列表
 */
export interface IBizType {
  id: string;
  biz_type: string;
  cover_ping: number;
  deployment_architecture: Array<'distributed' | 'centralized'>;
  creator: string;
  reviset: string;
  created_at: string;
  updated_at: string;
}
export type IBizTypeList = Array<IBizType>;
export type IBizTypeResData = IListResData<IBizTypeList>;

/**
 * 云资源选型方案
 */
export interface IAreaInfo {
  name: string;
  value?: number;
  children?: Array<IAreaInfo>;
}
export interface IGenerateSchemesReqParams {
  selected_countries: Array<string>;
  cover_ping: number;
  deployment_architecture: Array<'distributed' | 'centralized'>;
  biz_type: string;
  user_distribution: Array<IAreaInfo>;
  user_distribution_mode: string;
}
interface IRecommendScheme {
  cover_rate: number;
  composite_score: number;
  net_score: number;
  cost_score: number;
  result_idc_ids: string[];
  id: string;
  name: string;
  vendors: [];
  deployment_architecture: [];
  bk_biz_id: string;
  isSaved: boolean;
}
export type IRecommendSchemeList = Array<IRecommendScheme>;
export type IGenerateSchemesResData = IQueryResData<IRecommendSchemeList>;

export interface IServiceArea {
  country_name: string;
  province_name: string;
  network_latency: number;
}

export interface IIdcServiceAreaRel {
  idc_id: string;
  service_areas: Array<IServiceArea>;
  avg_latency: number;
}

export interface IIdcInfo {
  name: string;
  vendor: VendorEnum;
  region: string;
  id: string;
  price: number;
  country?: string;
}

// idc机房列表数据
export interface IIdcListItem extends IIdcInfo {
  bk_biz_id: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}
