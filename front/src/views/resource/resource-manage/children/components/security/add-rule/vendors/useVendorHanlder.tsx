import { VendorEnum } from '@/common/constant';
import { defineComponent, ref, VNode, watch } from 'vue';
import { TcloudRecord, TcloudRenderRow, tcloudTitles } from './tcloud';
import { HuaweiRecord, HuaweiRenderRow, huaweiTitles } from './huawei';
import { AwsRecord, AwsRenderRow, awsTitles } from './aws';
import { AzureRecord, AzureRenderRow, azureTitles } from './azure';
import { tcloudHandler, tcloudPreHandler } from './tcloud/DataHandler';
import { huaweiHandler, huaweiPreHandler } from './huawei/DataHandler';
import { awsHandler, awsPreHandler } from './aws/DataHandler';
import { azureHandler, azurePreHandler } from './azure/DataHandler';

export interface IHead {
  minWidth?: number;
  title: string;
  width: number;
  renderAppend?: () => VNode;
  required?: boolean;
  memo?: string;
}

interface VendorHandlerValue {
  titles: IHead[];
  row: ReturnType<typeof defineComponent>;
  Record: () => Object & { key: string };
  handleData: Function;
  preHandle: Function;
}

export type SecurityRuleType = 'ingress' | 'engress';

export type Ext<T extends object> = T & { [key: string]: any };

export const useVendorHandler = (vendor: VendorEnum, type: SecurityRuleType) => {
  const map: Map<VendorEnum, VendorHandlerValue> = new Map();

  map.set(VendorEnum.TCLOUD, {
    titles: tcloudTitles(type),
    row: TcloudRenderRow,
    Record: TcloudRecord,
    handleData: tcloudHandler,
    preHandle: tcloudPreHandler,
  });

  map.set(VendorEnum.HUAWEI, {
    titles: huaweiTitles(type),
    row: HuaweiRenderRow,
    Record: HuaweiRecord,
    handleData: huaweiHandler,
    preHandle: huaweiPreHandler,
  });

  map.set(VendorEnum.AWS, {
    titles: awsTitles(type),
    row: AwsRenderRow,
    Record: AwsRecord,
    handleData: awsHandler,
    preHandle: awsPreHandler,
  });

  map.set(VendorEnum.AZURE, {
    titles: azureTitles,
    row: AzureRenderRow,
    Record: AzureRecord,
    handleData: azureHandler,
    preHandle: azurePreHandler,
  });

  const handler = ref(map.get(VendorEnum.TCLOUD));

  watch(
    () => vendor,
    () => {
      handler.value = map.get(vendor);
    },
    {
      immediate: true,
    },
  );

  return {
    handler,
  };
};
