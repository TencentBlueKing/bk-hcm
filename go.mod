module hcm

go 1.21

require (
	cloud.google.com/go/asset v1.19.0
	cloud.google.com/go/bigquery v1.59.1
	cloud.google.com/go/iam v1.1.7
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.2.2
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5 v5.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/consumption/armconsumption v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2 v2.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.1.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription v1.0.0
	github.com/TencentBlueKing/gopkg v1.1.0
	github.com/aws/aws-sdk-go v1.44.334
	github.com/emicklei/go-restful/v3 v3.10.2
	github.com/go-playground/validator/v10 v10.11.2
	github.com/go-sql-driver/mysql v1.7.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang/protobuf v1.5.4
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.40
	github.com/jmoiron/sqlx v1.3.5
	github.com/json-iterator/go v1.1.12
	github.com/microsoftgraph/msgraph-sdk-go v1.26.0
	github.com/pborman/uuid v1.2.1
	github.com/prometheus/client_golang v1.14.0
	github.com/smartystreets/goconvey v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing v1.0.908
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam v1.0.908
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs v1.0.908
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb v1.0.908
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.908
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm v1.0.908
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl v1.0.908
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc v1.0.908
	github.com/tidwall/gjson v1.14.4
	go.etcd.io/etcd/api/v3 v3.5.13
	go.etcd.io/etcd/client/v3 v3.5.13
	go.uber.org/atomic v1.10.0
	go.uber.org/mock v0.2.0
	golang.org/x/time v0.5.0
	google.golang.org/api v0.172.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go v0.112.2 // indirect
	cloud.google.com/go/accesscontextmanager v1.8.6 // indirect
	cloud.google.com/go/compute v1.24.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/longrunning v0.5.6 // indirect
	cloud.google.com/go/orgpolicy v1.12.2 // indirect
	cloud.google.com/go/osconfig v1.12.6 // indirect
	github.com/apache/arrow/go/v14 v14.0.2 // indirect
	github.com/cjlapao/common-go v0.0.39 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/microsoft/kiota-abstractions-go v1.6.0 // indirect
	github.com/microsoft/kiota-authentication-azure-go v1.0.2 // indirect
	github.com/microsoft/kiota-http-go v1.3.3 // indirect
	github.com/microsoft/kiota-serialization-form-go v1.0.0 // indirect
	github.com/microsoft/kiota-serialization-json-go v1.0.7 // indirect
	github.com/microsoft/kiota-serialization-multipart-go v1.0.0 // indirect
	github.com/microsoft/kiota-serialization-text-go v1.0.0 // indirect
	github.com/microsoftgraph/msgraph-sdk-go-core v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smarty/assertions v1.15.0 // indirect
	github.com/std-uritemplate/std-uritemplate/go v0.0.57 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240415180920-8c6c420018be // indirect
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.2 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.9.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/flatbuffers v23.5.26+incompatible // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.13 // indirect
	go.mongodb.org/mongo-driver v1.12.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect; indirectd
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)
