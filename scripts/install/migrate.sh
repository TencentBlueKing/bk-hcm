#!/bin/bash

# 请修改下面的mysql相关配置. Please change the mysql configuration below.
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USERNAME=root
MYSQL_DATABASE=hcm
MYSQL_PASSWORD='password'

# 预定义变量，请勿修改. Predefined variables, do not modify.
SQLDIR=sql
TARGET_VERSION_FILE=../VERSION
TARGE_VERSION=
CURRENT_VERSION=
# 1.1.18前发布的文件中没有版本信息，通过硬编码表格来处理旧文件的版本对应关系
VERTABLE=(0.0.0 v1.0.0 v1.1.0 v1.1.3 v1.1.6 v1.1.7 v1.1.18)

MYSQL="mysql --host=$MYSQL_HOST --port=$MYSQL_PORT --user=$MYSQL_USERNAME --password=$MYSQL_PASSWORD \
        --database=$MYSQL_DATABASE --batch  --skip-column-names"

help() {
    cat <<EOF
Options:
    -i      first install, same as '-c v0.0.0'.
    -w [path]
            specify path for sql file.
    -t [target version]
            specify target version for migration. 
    -c [current version]
            specify current version for migration. 
    -a [path]
            target version file path.
    -h      show this help.
    
Previous sql file and hcm version relation:
    +------+---------+
    |  SQL |   HCM   |
    +------+---------+
    | 0001 | v1.0.0  |
    | 0002 | v1.1.0  |
    | 0003 | v1.1.3  |
    | 0004 | v1.1.6  |
    | 0005 | v1.1.7  |
    | 0006 | v1.1.18 |
    +------+---------+
you can use '-c' flag to specify current version for upgrading from old release.
EOF
}

mysqla() {
    # mysql Auth
    $MYSQL $@
}

mysqlx() {
    # mysql eXecute
    $MYSQL --execute="$1"
}

ver_ge() {
    # great or equal
    echo -e $1\\n$2 | sort -rV -C
}
ver_le() {
    # less or equal
    echo -e $1\\n$2 | sort -V -C
}

ver_gt() {
    # greater than
    ! ver_le $1 $2
}

get_current_version() {
    # 尝试从数据库获取当前版本并设置全局变量
    CURRENT_VERSION=$(mysqlx 'SELECT `hcm_ver` from hcm_version;')
    if [ $? -ne 0 ]; then
        echo "[ERROR] Fail to find version information! Check the error info below or use '-c' to specify current version."
        echo "If it is your first time installing hcm, just using '-i' flag."
        exit -1
    fi
}

get_hcm_ver() {
    ver=$(grep -o -E 'HCMVER=v[0-9]+\.[0-9]+\.[0-9]+' $1 | tail -c +8)
    # 处理没有版本信息的旧版本SQL文件
    if [ -z "$ver" ]; then
        # 从传入的参数重取前4个字符作为idx
        let idx=$(basename $1 | cut -c 4)
        echo ${VERTABLE[idx]}
    else
        echo $ver
    fi
}

arg_check() {
    #  参数检查

    while getopts 't:c:w:ha:i' flag; do
        case $flag in
        i) CURRENT_VERSION=v0.0.0 ;;
        c) CURRENT_VERSION=$OPTARG ;;
        t) TARGE_VERSION=$OPTARG ;;
        a) TARGET_VERSION_FILE=$OPTARG ;;
        w) SQLDIR=$OPTARG ;;
        h) help;exit 0;;
        *) help;exit -1;;
        esac
    done

    # TARGE_VERSION 未设置，尝试读取版本文件
    if [ -z "$TARGE_VERSION" ]; then
        TARGE_VERSION=$(head -n1 ${TARGET_VERSION_FILE})
        if [ -z "$TARGE_VERSION" ]; then
            echo "[ERROR] Fail to get target version! Please use -t to specify version, or -a to specify version file."
            exit -1
        fi
        echo got target version \($TARGE_VERSION\) from $TARGET_VERSION_FILE
    fi

    # CURRENT_VERSION 未设置，自动判断当前版本
    if [ -z "$CURRENT_VERSION" ]; then
        get_current_version
    fi

    echo target=$TARGE_VERSION current=$CURRENT_VERSION
    if ver_ge $CURRENT_VERSION $TARGE_VERSION; then
        echo "[ERROR] current($CURRENT_VERSION) >= target($TARGE_VERSION)"
        exit -1
    fi
}

main() {

    # 遍历每个sql文件，其中ls 默认按文件名的字符序排序
    for sqlfile in $(ls -1 $SQLDIR/*.sql); do
        # 获取匹配的版本信息
        hcmver_of_sql=$(get_hcm_ver $sqlfile)
        echo -n "[$hcmver_of_sql]" $sqlfile --\> 
        # 小于等于当前版本的pass
        if ver_le $hcmver_of_sql $CURRENT_VERSION; then
            echo pass
            continue
        fi
        #  大于当前版本的退出
        if ver_gt $hcmver_of_sql $TARGE_VERSION; then
            echo stop
            break
        fi
        echo exec
        mysqla <$sqlfile
        if [ $? -ne 0 ]; then
            echo "[ERROR] Fail to execute $sqlfile! Exiting..."
            exit -1
        fi
    done
}

arg_check $@
main
