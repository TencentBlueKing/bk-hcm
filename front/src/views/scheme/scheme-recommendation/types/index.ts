/**
 * 资源选型 TS 类型
 */

interface AreaInfo {
  name: string;
  value?: number;
  children?: AreaInfoList;
}

type AreaInfoList = Array<AreaInfo>;

export interface GenerateSchemesParams {
  cover_ping: number;
  deployment_architecture: Array<'distributed' | 'centralized'>;
  biz_type: string;
  user_distribution: AreaInfoList;
}
