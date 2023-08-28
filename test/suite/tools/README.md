## 运行集成测试相关说明

### 1. 文件说明
```shell
.
├── cloud-server.test
├── *.test
├── start.sh
└── tools.sh
```

- *.test 是编译完的二进制测试文件，文件名代表是对什么类型资源进行测试。如 cloud-server.test 是对 cloud-server 资源相关接口的集成测试。
- testhelper 是对测试执行结果导出的 json 文件进行统计分析的工具，最终会生成一个 html 页面。
- start.sh 执行测试并进行统计分析的脚本，需要配置相关环境变量才可运行。

#### 1.1 start.sh 执行需设置的环境变量

start.sh 支持通过指定环境变量文件设置环境变量
下列环境变量如果不设置将以脚本中的默认值运行.

```shell

# 集成测试环境 cloud-server 地址
export ENV_SUITE_TEST_CLOUD_REQUEST_HOST=http://127.0.0.1:9602
# 集成测试环境 hc-service 地址
export ENV_SUITE_TEST_HC_REQUEST_HOST=http://127.0.0.1:9601
# 集成测试环境  data-service
export ENV_SUITE_TEST_DATA_REQUEST_HOST=http://127.0.0.1:9600

# mysql 相关配置， 运行集成测试之前会进行清库操作，否则会对测试结果造成影响
export ENV_SUITE_TEST_MYSQL_IP=127.0.0.1
export ENV_SUITE_TEST_MYSQL_PORT=3306
export ENV_SUITE_TEST_MYSQL_USER=hcm
export ENV_SUITE_TEST_MYSQL_PW=hcm_suit_test_pwd
export ENV_SUITE_TEST_MYSQL_DB=hcm_suite_test

# 测试结果导出的json文件存储目录
export ENV_SUITE_TEST_SAVE_DIR=./result
# 对测试结果统计分析生成的html页面存储路径
export ENV_SUITE_TEST_OUTPUT_PATH=./result/statistics.html
```