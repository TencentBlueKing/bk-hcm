package lblogic

import (
	"strings"
	"testing"

	corecvm "hcm/pkg/api/core/cloud/cvm"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"

	"github.com/stretchr/testify/require"
)

func TestBuildBatchGetCvmWithoutVpcExpr_ShouldUseJSONOverlapsForIPArrays(t *testing.T) {
	ips := []string{"10.0.0.1", "10.0.0.2"}
	expr := buildBatchGetCvmWithoutVpcExpr(ips, enumor.TCloud, 2, "acc-1")

	require.NotNil(t, expr)
	require.Equal(t, filter.And, expr.Op)
	require.Len(t, expr.Rules, 4)

	ipExpr, ok := expr.Rules[0].(*filter.Expression)
	require.True(t, ok)
	require.Equal(t, filter.Or, ipExpr.Op)
	require.Len(t, ipExpr.Rules, 4)

	for _, rule := range ipExpr.Rules {
		atom, ok := rule.(*filter.AtomRule)
		require.True(t, ok)
		require.Equal(t, filter.JSONOverlaps.Factory(), atom.Op)
		require.Equal(t, ips, atom.Value)
	}
}

func TestValidateCvmExist_ShouldFindMatchedCVMForDifferentIPsInSameBatch(t *testing.T) {
	cvmList := []corecvm.BaseCvm{
		{
			CloudID:              "cvm-1",
			PrivateIPv4Addresses: []string{"10.0.0.1"},
			CloudVpcIDs:          []string{"vpc-1"},
		},
		{
			CloudID:              "cvm-2",
			PrivateIPv4Addresses: []string{"10.0.0.2"},
			CloudVpcIDs:          []string{"vpc-1"},
		},
	}

	lb := corelb.LoadBalancerRaw{BaseLoadBalancer: corelb.BaseLoadBalancer{CloudVpcID: "vpc-1"}}
	kt := &kit.Kit{}

	cvm1, err := validateCvmExist(kt, nil, "10.0.0.1", lb, false, false, "", cvmList)
	require.NoError(t, err)
	require.NotNil(t, cvm1)
	require.Equal(t, "cvm-1", cvm1.CloudID)

	cvm2, err := validateCvmExist(kt, nil, "10.0.0.2", lb, false, false, "", cvmList)
	require.NoError(t, err)
	require.NotNil(t, cvm2)
	require.Equal(t, "cvm-2", cvm2.CloudID)
}

func TestBuildBatchGetCvmWithoutVpcExpr_SQLShouldUseJSONOverlaps(t *testing.T) {
	ips := []string{"10.0.0.1", "10.0.0.2"}
	expr := buildBatchGetCvmWithoutVpcExpr(ips, enumor.TCloud, 2, "acc-1")

	where, _, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(where, "WHERE "))
	require.Regexp(t, `JSON_OVERLAPS\(private_ipv4_addresses, JSON_ARRAY\(:private_ipv4_addresses_[A-Za-z0-9]+_0,:private_ipv4_addresses_[A-Za-z0-9]+_1\)\)`, where)
	require.Regexp(t, `JSON_OVERLAPS\(private_ipv6_addresses, JSON_ARRAY\(:private_ipv6_addresses_[A-Za-z0-9]+_0,:private_ipv6_addresses_[A-Za-z0-9]+_1\)\)`, where)
	require.Regexp(t, `JSON_OVERLAPS\(public_ipv4_addresses, JSON_ARRAY\(:public_ipv4_addresses_[A-Za-z0-9]+_0,:public_ipv4_addresses_[A-Za-z0-9]+_1\)\)`, where)
	require.Regexp(t, `JSON_OVERLAPS\(public_ipv6_addresses, JSON_ARRAY\(:public_ipv6_addresses_[A-Za-z0-9]+_0,:public_ipv6_addresses_[A-Za-z0-9]+_1\)\)`, where)
	require.Contains(t, where, "vendor = :vendor")
	require.Contains(t, where, "bk_biz_id = :bk_biz_id")
	require.Contains(t, where, "account_id = :account_id")
}
