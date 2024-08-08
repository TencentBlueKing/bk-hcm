package loadbalancer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"

	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"

	"github.com/xuri/excelize/v2"
)

func (svc *lbSvc) checkLoadBalanceResourceLock(kt *kit.Kit,
	getLoadBalancerMapFunc func() (map[string]string, []*cloud.BatchOperationValidateError)) []*cloud.BatchOperationValidateError {

	lbIDToVIPMap, errList := getLoadBalancerMapFunc()
	for lbID, vip := range lbIDToVIPMap {
		flow, err := svc.checkResFlowRel(kt, lbID, enumor.LoadBalancerCloudResType)
		if err != nil {
			if flow != nil {
				errList = append(errList, &cloud.BatchOperationValidateError{
					Reason: fmt.Sprintf("loadBalancer[%s] exist task executing, not support modify", lbID),
					Ext:    flow.Owner,
				})
			} else {
				errList = append(errList, &cloud.BatchOperationValidateError{
					Reason: fmt.Sprintf("loadBalancer[%s] check resource flow failed: %v", vip, err),
				})
			}
		}
	}
	return errList
}

func parseExcelFileToRows(body cloud.Base64String) ([][]string, error) {
	buf := convertBase64StrToReader(body)
	file, err := excelize.OpenReader(buf)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获取第一个工作表的名称
	sheetName := file.GetSheetName(0)

	// 获取工作表中所有行
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 跳过标题行
	return rows[1:], nil
}

func convertBase64StrToReader(str cloud.Base64String) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(str)))
}
