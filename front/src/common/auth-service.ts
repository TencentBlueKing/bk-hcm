import * as authSymbol from '@/constants/auth-symbols';

// 权限标记，使用权限时提供的格式
export interface IAuthSign {
  type: symbol;
  relation?: [...(number | string)[]];
}

export type AuthActionType = 'create' | 'update' | 'find' | 'delete' | 'import' | 'access' | 'apply' | 'recycle';

export type AuthResourceType =
  | 'cloud_selection_scheme'
  | 'account'
  | 'biz'
  | 'cvm'
  | 'cert'
  | 'biz_audit'
  | 'recycle_bin'
  | 'root_account'
  | 'main_account'
  | 'account_bill'
  | 'load_balancer';

// 权限校验参数
export interface IVerifyResourceItem {
  action: AuthActionType;
  resource_type: AuthResourceType;
  bk_biz_id?: number;
  resource_id?: number | string;
  [key: string]: number | string;
}

export interface IVerifyParams {
  resources: IVerifyResourceItem[];
}

// 一个权限点的定义
export interface IAuthDefinition {
  id: string;
  action: AuthActionType;
  resourceType: AuthResourceType;
  transform?: (
    definition: IAuthDefinition,
    relation: IAuthSign['relation'],
  ) => IVerifyResourceItem | IVerifyResourceItem[];
}

const basicTransform = (
  definition: IAuthDefinition,
  meta?: { bk_biz_id?: number; resource_id?: number | string; [key: string]: number | string },
) => {
  const { action, resourceType } = definition;
  return {
    action,
    resource_type: resourceType,
    ...meta,
  };
};

export const getAuthDef = (type: symbol) => AUTH_DEFINITIONS[type];

export const getAuthDefs = (sign: IAuthSign | IAuthSign[]) => {
  const signs = Array.isArray(sign) ? sign : [sign];
  return signs.map((item) => getAuthDef(item.type));
};

export const getAuthResources = (sign: IAuthSign | IAuthSign[]) => {
  const signs = Array.isArray(sign) ? sign : [sign];

  const resources: IVerifyResourceItem[] = [];
  signs.forEach((sign) => {
    const { type, relation = [] } = sign;
    const definition = getAuthDef(type);

    if (!definition) {
      throw new Error(`未定义的权限类型：${type?.toString()}`);
    }

    let resource;
    if (definition.transform) {
      resource = definition.transform?.(definition, relation);
    } else {
      resource = basicTransform(definition);
    }

    if (Array.isArray(resource)) {
      resources.push(...resource);
    } else {
      resources.push(resource);
    }
  });

  return resources;
};

export const getVerifyParams = (sign: IAuthSign | IAuthSign[]) => {
  const resources = getAuthResources(sign);
  return { resources };
};

export const AUTH_DEFINITIONS = Object.freeze<Record<symbol, IAuthDefinition>>({
  [authSymbol.AUTH_CREATE_CLOUD_SELECTION_SCHEME]: {
    id: 'cloud_selection_recommend',
    action: 'create',
    resourceType: 'cloud_selection_scheme',
  },
  [authSymbol.AUTH_FIND_CLOUD_SELECTION_SCHEME]: {
    id: 'cloud_selection_find',
    action: 'find',
    resourceType: 'cloud_selection_scheme',
  },
  [authSymbol.AUTH_UPDATE_CLOUD_SELECTION_SCHEME]: {
    id: 'cloud_selection_edit',
    action: 'update',
    resourceType: 'cloud_selection_scheme',
  },
  [authSymbol.AUTH_DELETE_CLOUD_SELECTION_SCHEME]: {
    id: 'cloud_selection_delete',
    action: 'delete',
    resourceType: 'cloud_selection_scheme',
  },
  [authSymbol.AUTH_FIND_ACCOUNT]: {
    id: 'account_find',
    action: 'find',
    resourceType: 'account',
    transform: (definition, relation) => basicTransform(definition, { resource_id: relation[0] }),
  },
  [authSymbol.AUTH_IMPORT_ACCOUNT]: {
    id: 'account_import',
    action: 'import',
    resourceType: 'account',
  },
  [authSymbol.AUTH_UPDATE_ACCOUNT]: {
    id: 'account_edit',
    action: 'import',
    resourceType: 'account',
  },
  [authSymbol.AUTH_ACCESS_BIZ]: {
    id: 'biz_access',
    action: 'access',
    resourceType: 'biz',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_FIND_IAAS_RESOURCE]: {
    id: 'resource_find',
    action: 'find',
    resourceType: 'cvm',
  },
  [authSymbol.AUTH_CREATE_IAAS_RESOURCE]: {
    id: 'iaas_resource_create',
    action: 'create',
    resourceType: 'cvm',
    transform: (definition, relation) => basicTransform(definition, { resource_id: relation[0] }),
  },
  [authSymbol.AUTH_UPDATE_IAAS_RESOURCE]: {
    id: 'iaas_resource_operate',
    action: 'update',
    resourceType: 'cvm',
    transform: (definition, relation) => basicTransform(definition, { resource_id: relation[0] }),
  },
  [authSymbol.AUTH_DELETE_IAAS_RESOURCE]: {
    id: 'iaas_resource_delete',
    action: 'delete',
    resourceType: 'cvm',
    transform: (definition, relation) => basicTransform(definition, { resource_id: relation[0] }),
  },
  [authSymbol.AUTH_BIZ_FIND_IAAS_RESOURCE]: {
    id: 'biz_resource_find',
    action: 'find',
    resourceType: 'cvm',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_BIZ_CREATE_IAAS_RESOURCE]: {
    id: 'biz_iaas_resource_create',
    action: 'create',
    resourceType: 'cvm',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_BIZ_UPDATE_IAAS_RESOURCE]: {
    id: 'biz_iaas_resource_operate',
    action: 'update',
    resourceType: 'cvm',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_BIZ_DELETE_IAAS_RESOURCE]: {
    id: 'biz_iaas_resource_delete',
    action: 'delete',
    resourceType: 'cvm',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_BIZ_FIND_AUDIT]: {
    id: 'resource_audit_find',
    action: 'find',
    resourceType: 'biz_audit',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_FIND_RECYCLE_BIN]: {
    id: 'recycle_bin_find',
    action: 'find',
    resourceType: 'recycle_bin',
  },
  [authSymbol.AUTH_MANAGE_RECYCLE_BIN]: {
    id: 'recycle_bin_manage',
    action: 'recycle',
    resourceType: 'recycle_bin',
  },
  [authSymbol.AUTH_CREATE_CERT]: {
    id: 'cert_resource_create',
    action: 'create',
    resourceType: 'cert',
  },
  [authSymbol.AUTH_BIZ_CREATE_CERT]: {
    id: 'biz_cert_resource_create',
    action: 'create',
    resourceType: 'cert',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_DELETE_CERT]: {
    id: 'cert_resource_delete',
    action: 'delete',
    resourceType: 'cert',
  },
  [authSymbol.AUTH_BIZ_DELETE_CERT]: {
    id: 'biz_cert_resource_delete',
    action: 'delete',
    resourceType: 'cert',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_FIND_ROOT_ACCOUNT]: {
    id: 'root_account_find',
    action: 'find',
    resourceType: 'root_account',
  },
  [authSymbol.AUTH_FIND_MAIN_ACCOUNT]: {
    id: 'main_account_find',
    action: 'find',
    resourceType: 'main_account',
  },
  [authSymbol.AUTH_UPDATE_MAIN_ACCOUNT]: {
    id: 'main_account_edit',
    action: 'update',
    resourceType: 'main_account',
  },
  [authSymbol.AUTH_FIND_ACCOUNT_BILL]: {
    id: 'account_bill_find',
    action: 'find',
    resourceType: 'account_bill',
  },
  [authSymbol.AUTH_CREATE_CLB]: {
    id: 'clb_resource_create',
    action: 'create',
    resourceType: 'load_balancer',
    transform: (definition, relation) => basicTransform(definition, { resource_id: relation[0] }),
  },
  [authSymbol.AUTH_BIZ_CREATE_CLB]: {
    id: 'biz_clb_resource_create',
    action: 'apply',
    resourceType: 'load_balancer',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_UPDATE_CLB]: {
    id: 'clb_resource_operate',
    action: 'update',
    resourceType: 'load_balancer',
    transform: (definition, relation) => basicTransform(definition, { resource_id: relation[0] }),
  },
  [authSymbol.AUTH_BIZ_UPDATE_CLB]: {
    id: 'biz_clb_resource_operate',
    action: 'update',
    resourceType: 'load_balancer',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
  [authSymbol.AUTH_DELETE_CLB]: {
    id: 'clb_resource_delete',
    action: 'delete',
    resourceType: 'load_balancer',
    transform: (definition, relation) => basicTransform(definition, { resource_id: relation[0] }),
  },
  [authSymbol.AUTH_BIZ_DELETE_CLB]: {
    id: 'biz_clb_resource_delete',
    action: 'delete',
    resourceType: 'load_balancer',
    transform: (definition, relation) => basicTransform(definition, { bk_biz_id: relation[0] as number }),
  },
});
