package loadbalancer

import (
	"fmt"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// BindRSPreview 批量绑定RS 预览接口
func (svc *lbSvc) BindRSPreview(cts *rest.Contexts) (interface{}, error) {

	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	req := new(cloud.BatchOperationImportExcelReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rows, err := parseExcelFileToRows(req.ExcelFileBase64)
	if err != nil {
		return nil, err
	}

	bindRSRecords, errList := parseBindRsRawInput(rows)
	return svc.bindRSPreview(cts, bindRSRecords, errList, bizID)
}

func (svc *lbSvc) bindRSPreview(cts *rest.Contexts, bindRSRecords []*lblogic.BindRSRecord,
	errList []*cloud.BatchOperationValidateError, bkBizID int64) (interface{}, error) {

	validateErrs := svc.validateBindRSRecord(cts.Kit, bindRSRecords, bkBizID)
	errList = append(errList, validateErrs...)

	validateRelationsErrs := svc.validateBindRSRecordRelations(bindRSRecords)
	errList = append(errList, validateRelationsErrs...)

	resourceLockErrs := svc.checkLoadBalanceResourceLock(cts.Kit, func() (map[string]string, []*cloud.BatchOperationValidateError) {
		return svc.getBindRSRecordsLoadBalanceMap(cts.Kit, bindRSRecords, bkBizID)
	})
	errList = append(errList, resourceLockErrs...)

	if len(errList) > 0 {
		return errList, nil
	}

	resp, err := svc.buildBindRSPreviewResp(cts.Kit, bindRSRecords, bkBizID)
	if err != nil {
		logs.Errorf("build bind rs preview response failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return resp, nil
}

func (svc *lbSvc) validateBindRSRecord(kt *kit.Kit, records []*lblogic.BindRSRecord,
	bkBizID int64) []*cloud.BatchOperationValidateError {

	errList := make([]*cloud.BatchOperationValidateError, 0)
	for _, record := range records {
		key := record.GetKey()
		err := record.Validate()
		if err != nil {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s %v", key, err),
			})
		}

		validateErrs := record.CheckWithDataService(kt, svc.client.DataService(), bkBizID)
		if len(validateErrs) > 0 {
			errList = append(errList, validateErrs...)
			continue
		}
	}
	return errList
}

// validateBindRSRecordRelations 校验 records的关联关系
func (svc *lbSvc) validateBindRSRecordRelations(records []*lblogic.BindRSRecord) []*cloud.BatchOperationValidateError {

	recordMap := make(map[string]struct{})
	fourLayerRSMap := make(map[string]string)
	errList := make([]*cloud.BatchOperationValidateError, 0)
	for _, record := range records {
		key := record.GetKey()
		if _, ok := recordMap[key]; ok {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s duplicate record", key),
			})
		}
		recordMap[key] = struct{}{}

		if !record.Protocol.IsLayer7Protocol() {
			// (vip+protocol+rsip+rsport) should be globally unique for fourth layer listeners"+
			for _, rsKey := range record.GetRSKeys() {
				k := fmt.Sprintf("%s:%s:%s", record.VIP, record.Protocol, rsKey)
				if v, ok := fourLayerRSMap[k]; ok {
					errList = append(errList, &cloud.BatchOperationValidateError{
						Reason: fmt.Sprintf("(vip+protocol+rsip+rsport) should be globally unique for"+
							" fourth layer listeners, %s already exists in listener %s", rsKey, v),
					})
					continue
				}
				fourLayerRSMap[k] = key
			}
		}
	}
	return errList
}

// getBindRSRecordsLoadBalanceMap return a map[loadBalancerID]vip
func (svc *lbSvc) getBindRSRecordsLoadBalanceMap(kt *kit.Kit, records []*lblogic.BindRSRecord, bkBizID int64) (
	map[string]string, []*cloud.BatchOperationValidateError) {

	errList := make([]*cloud.BatchOperationValidateError, 0)
	lbIDToVIPMap := make(map[string]string)
	for _, record := range records {
		lb, err := record.GetLoadBalancer(kt, svc.client.DataService(), bkBizID)
		if err != nil {
			logs.Errorf("get loadBalancer failed, err: %v, rid: %s", err, kt.Rid)
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s %v", record.GetKey(), err),
			})
			continue
		}
		lbIDToVIPMap[lb.ID] = record.VIP
	}
	return lbIDToVIPMap, errList
}

func (svc *lbSvc) buildBindRSPreviewResp(kt *kit.Kit, records []*lblogic.BindRSRecord,
	bkBizID int64) ([]*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord], error) {

	resultMap := make(map[string]*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord])
	for _, record := range records {
		lb, err := record.GetLoadBalancer(kt, svc.client.DataService(), bkBizID)
		if err != nil {
			logs.Errorf("get loadBalancer failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		previewResp, ok := resultMap[lb.ID]
		if !ok {
			previewResp = &cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord]{
				ClbID:        lb.ID,
				ClbName:      lb.Name,
				Vip:          record.VIP,
				IPDomainType: lblogic.IPDomainTypeMap[record.IPDomainType],
				Listeners:    make([]*lblogic.BindRSRecord, 0),
			}
			resultMap[lb.ID] = previewResp
		}
		previewResp.Listeners = append(previewResp.Listeners, record)
		previewResp.NewRsCount += len(record.RSInfos)
	}
	result := make([]*cloud.BatchOperationPreviewResult[*lblogic.BindRSRecord], 0, len(resultMap))
	for _, pr := range resultMap {
		result = append(result, pr)
	}
	return result, nil
}

func parseBindRsRawInput(rows [][]string) (
	[]*lblogic.BindRSRecord, []*cloud.BatchOperationValidateError) {

	errList := make([]*cloud.BatchOperationValidateError, 0, len(rows))
	bindRSRecords := make([]*lblogic.BindRSRecord, 0, len(rows))
	for i, row := range rows {
		rawInput, err := lblogic.ParseBindRSRawInput(row)
		if err != nil {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("parse row %d failed: %v", i+2, err),
			})
			continue
		}
		records, err := rawInput.SplitRecord()
		if err != nil {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("parse row %d failed: %v", i+2, err),
			})
			continue
		}

		bindRSRecords = append(bindRSRecords, records...)
	}
	return bindRSRecords, errList
}
