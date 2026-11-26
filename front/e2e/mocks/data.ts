const longCertName = 'ieg-shared-certificate-with-a-very-long-remark-for-e2e';
const getMockCertName = (index: number) => {
  if (index < 2) return longCertName;
  if (index === 2) return 'expired-certificate';
  return `证书-${index}`;
};
/** 前 50 条用于首屏分页，隐藏目标证书放在第二页，验证后端检索能跨页命中。 */
export const mockCertificates = [
  ...Array.from({ length: 50 }, (_, index) => ({
    id: `cert-e2e-${index}`,
    cloud_id: `CERT-VISIBLE-${String(index).padStart(3, '0')}`,
    name: getMockCertName(index),
    vendor: 'tcloud',
    account_id: 'account-e2e',
    cert_type: 'SVR',
    cert_status: index === 2 ? '3' : '1',
    domain: ['e2e.example.com', '*.e2e.example.com'],
    encrypt_algorithm: index % 2 === 0 ? 'RSA 2048' : 'ECC 256',
    cloud_created_time: '2026-01-01T00:00:00Z',
    cloud_expired_time: index === 2 ? '2025-01-01T00:00:00Z' : '2027-04-29T00:00:00Z',
    bk_biz_id: -1,
  })),
  {
    id: 'cert-e2e-hidden',
    cloud_id: 'CERT-HIDDEN-999',
    name: 'ieg 目标证书',
    vendor: 'tcloud',
    account_id: 'account-e2e',
    cert_type: 'SVR',
    cert_status: '1',
    domain: ['hidden.example.com'],
    encrypt_algorithm: 'ECC 256',
    cloud_created_time: '2026-01-01T00:00:00Z',
    cloud_expired_time: '2028-01-01T00:00:00Z',
    bk_biz_id: -1,
  },
];

export const mockClbDetails = {
  id: 'clb-e2e',
  name: 'E2E-CLB',
  cloud_id: 'lb-e2e',
  account_id: 'account-e2e',
  vendor: 'tcloud',
  region: 'ap-test-1',
  cloud_vpc_id: 'vpc-e2e',
  lb_vip: [],
  bk_biz_id: 1,
  extension: {
    sla_type: 'shared',
    charge_type: '',
    load_balancer_pass_to_target: false,
    internet_charge_type: '',
    snat: false,
    snat_pro: false,
    snat_ips: [],
    delete_protect: false,
    egress: '',
    mix_ip_target: false,
    forward: 0,
    target_vpc: '',
    target_region: '',
  },
};

export const mockData = {
  certificates: mockCertificates,
  clbDetails: mockClbDetails,
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
