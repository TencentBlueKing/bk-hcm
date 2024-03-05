/* eslint-disable max-len */
import http from '@/http';
import { defineStore } from 'pinia';
import { QueryFilterType, IPageQuery, IQueryResData } from '@/typings/common';
import { IAreaInfo, IBizTypeResData, ICountriesListResData, IGenerateSchemesResData, IUserDistributionResData, IGenerateSchemesReqParams, IRecommendSchemeList, IIdcServiceAreaRel, IIdcInfo, ISchemeSelectorItem } from '@/typings/scheme';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;


// 资源选型模块相关状态管理和接口定义
export const useSchemeStore = defineStore({
  id: 'schemeStore',
  state: () => ({
    userDistribution: [] as Array<IAreaInfo>,
    recommendationSchemes: [] as IRecommendSchemeList,
    // 用户选择的配置
    schemeConfig: {
      cover_ping: 0,
      biz_type: '',
      deployment_architecture: '',
    },
    schemeList: [] as ISchemeSelectorItem[],
    schemeData: {
      deployment_architecture: [],
      vendors: [],
      composite_score: 0,
      net_score: 0,
      cost_score: 0,
      name: '',
      idcList: [],
      id: '0',
    },
    selectedSchemeIdx: 0,
  }),
  actions: {
    setUserDistribution(data: Array<IAreaInfo>) {
      this.userDistribution = data;
    },
    setSelectedSchemeIdx(idx: number) {
      this.selectedSchemeIdx = idx;
    },
    setRecommendationSchemes(data: IRecommendSchemeList) {
      this.recommendationSchemes = data;
    },
    setSchemeConfig(cover_ping: number, biz_type: string, deployment_architecture: string) {
      this.schemeConfig.cover_ping = cover_ping;
      this.schemeConfig.biz_type = biz_type;
      this.schemeConfig.deployment_architecture = deployment_architecture;
    },
    setSchemeList(list: ISchemeSelectorItem[]) {
      this.schemeList = list;
    },
    setSchemeData(data: typeof this.schemeData) {
      this.schemeData = data;
    },
    sortSchemes(choice: string, isDes = true) {
      this.recommendationSchemes = this.recommendationSchemes.sort((a, b) => (b[choice] - a[choice]) * (isDes ? 1 : -1));
    },
    /**
     * 获取资源选型方案列表
     * @param filter 过滤参数
     * @param page 分页参数
     * @returns
     */
    listCloudSelectionScheme(filter: QueryFilterType, page: IPageQuery) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/schemes/list`, { filter, page });
    },
    /**
     * 删除资源选型方案
     * @param ids 方案id列表
     * @returns
     */
    deleteCloudSelectionScheme(ids: string[]) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/schemes/batch`, { data: { ids } });
    },
    /**
     * 获取资源选型方案详情
     * @param id 方案id
     * @returns
     */
    getCloudSelectionScheme(id: string) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/schemes/${id}`);
    },
    /**
     * 更新资源选型方案
     * @param id 方案id
     * @param data 方案数据
     */
    updateCloudSelectionScheme(id: string, data: { name: string; bk_biz_id?: number; }) {
      const { name, bk_biz_id } = data;
      const params: { name: string; bk_biz_id?: number; } = { name };
      if (bk_biz_id) {
        params.bk_biz_id = bk_biz_id;
      }
      return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/schemes/${id}`, params);
    },
    /**
     * 获取收藏的资源选型方案列表
     * @returns
     */
    listCollection() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/collections/cloud_selection_scheme/list`);
    },
    /** 添加收藏
    * @param id 方案id
    * @returns
    */
    createCollection(id: string) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/collections/create`, { res_type: 'cloud_selection_scheme', res_id: id });
    },
    /**
      * 取消收藏
      * @param id 收藏id
      * @returns
      */
    deleteCollection(id: number) {
      return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/collections/${id}`);
    },
    /**
     * 查询IDC机房列表
     * @param filter 过滤参数
     * @param page 分页参数
     * @returns
     */
    // listIdc (filter: QueryFilterType, page: IPageQuery) {
    //   return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/idcs/list`, { filter, page });
    // },
    /**
     * 查询业务延迟数据
     * @param topo 拓扑列表
     * @param ids idc列表
     */
    queryBizLatency(topo: IAreaInfo[], ids: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/latency/biz/query`, { area_topo: topo, idc_ids: ids });
    },
    /**
     * 查询ping延迟数据
     * @param topo 拓扑列表
     * @param ids idc列表
     */
    queryPingLatency(topo: IAreaInfo[], ids: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/latency/ping/query`, { area_topo: topo, idc_ids: ids });
    },
    /**
     * 获取云选型数据支持的国家列表
     * @returns
     */
    listCountries(): Promise<ICountriesListResData> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/countries/list`);
    },
    /**
     * 获取业务类型列表
     * @param page 分页参数
     * @returns
     */
    listBizTypes(page: IPageQuery): Promise<IBizTypeResData> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/biz_types/list`, { page });
    },
    /**
     * 获取云选型用户分布占比
     * @param area_topo 需要查询的国家列表
     * @returns
     */
    queryUserDistributions(area_topo: Array<IAreaInfo>): Promise<IUserDistributionResData> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/user_distributions/query`, { area_topo });
    },
    /**
     * 生成云资源选型方案
     * @param formData 业务属性
     * @returns
     */
    generateSchemes(data: IGenerateSchemesReqParams): Promise<IGenerateSchemesResData> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/schemes/generate`, data);
    },
    /**
     * 获取服务区
     * @param datasource 数据来源：ping（裸ping数据）、biz（业务数据）
     * @param idc_ids 机房 id 列表
     * @param area_topo 国家城市拓扑
     * @returns 服务区
     */
    queryIdcServiceArea(datasource: string, idc_ids: Array<string>, area_topo: Array<IAreaInfo>): IQueryResData<Array<IIdcServiceAreaRel>> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/idcs/service_areas/${datasource}/query`, {
        idc_ids,
        area_topo,
      });
    },
    /**
     * 查询idc列表对应的详细信息
     */
    listIdc(filter: QueryFilterType, page: IPageQuery): IQueryResData<Array<IIdcInfo>> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/idcs/list`, {
        filter,
        page,
      });
    },
    createScheme(data: any) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/schemes/create`, data);
    },
  },
});
