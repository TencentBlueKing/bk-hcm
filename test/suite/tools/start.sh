#!/bin/bash

echo -e '\033[34;1mStart suite testing\033[0m'

for env_file in $1; do
  echo  -e "\033[33mSourcing environment file: ${env_file}\033[0m"
  source ${env_file}
done


# suite test env address.
etcd_endpoints=${ENV_SUITE_TEST_ETCD_ENDPOINTS:-"127.0.0.1:2379"}

# Warn: mysql used to clear the mysql of the suite test environment.
mysql_ip=${ENV_SUITE_TEST_MYSQL_IP:-"127.0.0.1"}
mysql_port=${ENV_SUITE_TEST_MYSQL_PORT:-"3306"}
mysql_user=${ENV_SUITE_TEST_MYSQL_USER:-"hcm"}
mysql_password=${ENV_SUITE_TEST_MYSQL_PW:-"hcm_suit_test_pwd"}
mysql_db=${ENV_SUITE_TEST_MYSQL_DB:-"hcm_suite_test"}

# go test result export json file save dir.
save_dir=${ENV_SUITE_TEST_SAVE_DIR:-"result"}
# go test result statistics result save file path.
output_path=${ENV_SUITE_TEST_OUTPUT_PATH:-"output.html"}

# exec go test.
if [ -d "${save_dir}" ]; then
  rm -r ${save_dir}
fi

mkdir ${save_dir}

./cloud-server.test -test.run TestCloudServer -convey-json=true \
  --etcd-endpoints=${etcd_endpoints}  \
  --mysql-ip=${mysql_ip} --mysql-port=${mysql_port} --mysql-user=${mysql_user} --mysql-passwd=${mysql_password} --mysql-db=${mysql_db} \
     > ${save_dir}/api.json

if [ $? -ne 0 ]; then
  echo -e '\033[31;1mSuite testing FAILED\033[0m'
else
  echo -e '\033[32;1mSuite testing SUCCEED\033[0m'
fi
# statistics.
./testhelper -input-dir=${save_dir} -output-path=${output_path}

