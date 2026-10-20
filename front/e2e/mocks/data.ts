export const mockData = {
  businesses: [{ id: 1, name: '测试业务' }],
  accounts: [
    {
      id: 'account-e2e',
      vendor: 'tcloud',
      name: '测试云账号',
      managers: [],
      type: 'resource',
      site: 'china',
      bk_biz_ids: [1],
    },
  ],
  region: { region_id: 'ap-test-1', region_name: '测试地域', vendor: 'tcloud' },
  zone: { name: 'ap-test-1a', name_cn: '测试一区' },  resource: {
    ZoneResourceSet: [
      {
        MasterZone: 'ap-test-1a',
        SlaveZone: '',
        ResourceSet: [
          {
            Type: ['CMCC'],
            Isp: 'CMCC',
            AvailabilitySet: [{ Type: 'CMCC', Availability: 'Available' }],
            TypeSet: [{ Type: 'CMCC', SpecAvailabilitySet: [{ SpecType: 'shared', Availability: 'Available' }] }],
          },
        ],
      },
    ],
    TotalCount: 1,
  },
  clusterTags: [
    {
      cluster_tag: 'l4-test',
      cluster_type: 'TGW',
      clusters: [
        {
          cloud_cluster_id: 'cluster-test',
          cluster_id: 'cluster-test',
          cluster_name: '测试四层集群',
          egress: 'test-egress',
          isp: 'CMCC',
          zone: 'ap-test-1a',
        },
      ],
    },
    {
      cluster_tag: 'l7-test',
      cluster_type: 'STGW',
      clusters: [
        {
          cloud_cluster_id: 'l7-cluster-test',
          cluster_id: 'l7-cluster-test',
          cluster_name: '测试七层集群',
          egress: 'test-egress',
          isp: 'CMCC',
          zone: 'ap-test-1a',
        },
      ],
    },
  ],
  /** 竞态回归：第一次（旧条件）请求延迟返回的过期标签 */
  staleClusterTags: [
    {
      cluster_tag: 'l4-stale',
      cluster_type: 'TGW',
      clusters: [
        {
          cloud_cluster_id: 'cluster-stale',
          cluster_id: 'cluster-stale',
          cluster_name: '过期四层集群',
          egress: 'stale-egress',
          isp: 'CMCC',
          zone: 'ap-test-1a',
        },
      ],
    },
    {
      cluster_tag: 'l7-stale',
      cluster_type: 'STGW',
      clusters: [
        {
          cloud_cluster_id: 'l7-cluster-stale',
          cluster_id: 'l7-cluster-stale',
          cluster_name: '过期七层集群',
          egress: 'stale-egress',
          isp: 'CMCC',
          zone: 'ap-test-1a',
        },
      ],
    },
  ],
};
