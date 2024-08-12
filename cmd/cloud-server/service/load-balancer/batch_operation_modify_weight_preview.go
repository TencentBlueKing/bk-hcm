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

// ModifyWeightPreview ...
func (svc *lbSvc) ModifyWeightPreview(cts *rest.Contexts) (interface{}, error) {

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
	updateWeightRecords, errList := parseModifyWeightRawInput(rows)
	return svc.modifyWeightPreview(cts, updateWeightRecords, errList, bizID)
}

func (svc *lbSvc) modifyWeightPreview(cts *rest.Contexts, updateWeightRecords []*lblogic.ModifyWeightRecord,
	errList []*cloud.BatchOperationValidateError, bkBizID int64) (interface{}, error) {

	validateErrors := svc.validateModifyWeightRecord(cts.Kit, updateWeightRecords, bkBizID)
	errList = append(errList, validateErrors...)

	validateRelationsErrs := svc.validateModifyWeightRecordRelations(updateWeightRecords)
	errList = append(errList, validateRelationsErrs...)

	resourceLockErrs := svc.checkLoadBalanceResourceLock(cts.Kit,
		func() (map[string]string, []*cloud.BatchOperationValidateError) {
			return svc.getModifyWeightRecordsLoadBalanceMap(cts.Kit, updateWeightRecords, bkBizID)
		})
	errList = append(errList, resourceLockErrs...)

	if len(errList) > 0 {
		return errList, nil
	}

	result, err := svc.buildModifyWeightPreviewResp(cts.Kit, updateWeightRecords, bkBizID)
	if err != nil {
		logs.Errorf("build modify weight preview response failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func (svc *lbSvc) validateModifyWeightRecord(kt *kit.Kit, records []*lblogic.ModifyWeightRecord,
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

func (svc *lbSvc) validateModifyWeightRecordRelations(
	records []*lblogic.ModifyWeightRecord) []*cloud.BatchOperationValidateError {

	recordMap := make(map[string]struct{})
	errList := make([]*cloud.BatchOperationValidateError, 0)
	for _, record := range records {
		flag := record.GetKey()
		if _, ok := recordMap[flag]; ok {
			errList = append(errList, &cloud.BatchOperationValidateError{
				Reason: fmt.Sprintf("%s duplicate record", flag),
			})
		}
		recordMap[flag] = struct{}{}
	}
	return errList
}

func (svc *lbSvc) getModifyWeightRecordsLoadBalanceMap(kt *kit.Kit, records []*lblogic.ModifyWeightRecord,
	bkBizID int64) (map[string]string, []*cloud.BatchOperationValidateError) {

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

func (svc *lbSvc) buildModifyWeightPreviewResp(kt *kit.Kit, records []*lblogic.ModifyWeightRecord,
	bkBizID int64) ([]*cloud.BatchOperationPreviewResult[*lblogic.ModifyWeightRecord], error) {

	resultMap := make(map[string]*cloud.BatchOperationPreviewResult[*lblogic.ModifyWeightRecord])
	for _, record := range records {
		lb, err := record.GetLoadBalancer(kt, svc.client.DataService(), bkBizID)
		if err != nil {
			logs.Errorf("get loadBalancer failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		previewResp, ok := resultMap[lb.ID]
		if !ok {
			previewResp = &cloud.BatchOperationPreviewResult[*lblogic.ModifyWeightRecord]{
				ClbID:        lb.ID,
				ClbName:      lb.Name,
				Vip:          record.VIP,
				IPDomainType: lblogic.IPDomainTypeMap[record.IPDomainType],
				Listeners:    make([]*lblogic.ModifyWeightRecord, 0),
			}
			resultMap[lb.ID] = previewResp
		}
		previewResp.Listeners = append(previewResp.Listeners, record)
		previewResp.NewRsCount += len(record.RSInfos)
	}
	result := make([]*cloud.BatchOperationPreviewResult[*lblogic.ModifyWeightRecord], 0, len(resultMap))
	for _, pr := range resultMap {
		result = append(result, pr)
	}
	return result, nil
}

func parseModifyWeightRawInput(rows [][]string) (
	[]*lblogic.ModifyWeightRecord, []*cloud.BatchOperationValidateError) {

	errList := make([]*cloud.BatchOperationValidateError, 0, len(rows))
	updateWeightRecords := make([]*lblogic.ModifyWeightRecord, 0, len(rows))
	for i, row := range rows {
		rawInput, err := lblogic.ParseModifyWeightRawInput(row)
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
		updateWeightRecords = append(updateWeightRecords, records...)
	}
	return updateWeightRecords, errList
}
