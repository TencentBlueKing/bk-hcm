#!/bin/bash
#
# TencentBlueKing is pleased to support the open source community by making
# 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
# Copyright (C) 2022 THL A29 Limited,
# a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
# either express or implied. See the License for the
# specific language governing permissions and limitations under the License.
#
# We undertake not to change the open source license (MIT license) applicable
#
# to the current version of the project delivered to anyone in the future.
#

RUNDIR=$(realpath ${RUNDIR:=".run"})

BINDIR=$(realpath ${BINDIR:="$RUNDIR/bin"})
ETCDIR=$(realpath ${ETCDIR:="$RUNDIR/etc"})

echo running in ${RUNDIR},
echo BINDIR=${BINDIR}, ETCDIR=${ETCDIR}

cd $RUNDIR

$BINDIR/bk-hcm-dataservice  -c $ETCDIR/data_service.yaml &
$BINDIR/bk-hcm-cloudserver -c  $ETCDIR/cloud_server.yaml &
$BINDIR/bk-hcm-hcservice -c  $ETCDIR/hc_service.yaml --mock &
$BINDIR/bk-hcm-authserver -c  $ETCDIR/auth_server.yaml --disable-auth &


function ctrl_c() {
    echo -e '\033[32;1mKilling...\033[0m'
    pkill bk-hcm
}
trap ctrl_c INT

# wait all service exit
wait



