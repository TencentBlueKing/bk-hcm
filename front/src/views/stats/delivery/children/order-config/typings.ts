import { IListResData } from '@/typings';

export interface IOrderStatistics {
  id?: string | number;
  stat_month?: string;
  bk_biz_id?: number | string;
  sub_order_ids?: string[];
  start_at?: string;
  end_at?: string;
  memo?: string;
}

export interface IYearMonths {
  stat_month: string;
}

export type OrderStatisticsListData = IListResData<IOrderStatistics[]>;
export type YearMonthsListData = IListResData<IYearMonths[]>;
