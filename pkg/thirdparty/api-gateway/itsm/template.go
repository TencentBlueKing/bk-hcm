/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package itsm

const (
	// processInitTemplate 流程初始化模版
	processInitTemplate = `{
    "system": {
        "code": "{{ .systemID }}",
        "name": "海垒_HCM"
    },
    "form_models": [
        {
            "key": "$FormModel20250605162400001301",
            "name": "hcm_main_account",
            "desc": "HCM2.0 二级账号申请",
            "portal_id": "DEFAULT",
            "app_id": "core",
            "translations": {
                "name": "hcm_main_account",
                "name_en": "add_account",
                "desc": "HCM2.0 二级账号申请",
                "desc_en": "HCM2.0 二级账号申请"
            },
            "meta": {
                "fields": {
                    "platform_manager": {
                        "translations": {
                            "name": "平台管理员",
                            "name_en": "平台管理员",
                            "name_zh_hans": "平台管理员"
                        },
                        "key": "platform_manager",
                        "name": "平台管理员",
                        "type": "multiUser",
                        "is_builtin": false,
                        "jsonschema": {
                            "translations": {
                                "title": "",
                                "title_en": "",
                                "title_zh_hans": ""
                            },
                            "type": "array",
                            "title": "",
                            "number_unit": "",
                            "table_relation": null,
                            "into_todo": [

                            ],
                            "out_todo": [

                            ],
                            "attr_relation": null,
                            "itsm_jmespath": null,
                            "itsm_options": null,
                            "itsm_options_type": null,
                            "format": null,
                            "columns": null,
                            "properties": null,
                            "items": {
                                "translations": {
                                    "title": "",
                                    "title_en": "",
                                    "title_zh_hans": ""
                                },
                                "type": "string",
                                "title": "",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": "user",
                                "columns": null,
                                "properties": null,
                                "items": null
                            }
                        },
                        "meta": {
                            "key": "platform_manager",
                            "desc": "",
                            "tips": "请输入",
                            "type": "multiUser",
                            "title": {
                                "value": "平台管理员",
                                "isHide": false
                            },
                            "permission": {
                                "readonly": [
                                    "title",
                                    "key",
                                    "tips",
                                    "desc"
                                ]
                            },
                            "translations": {
                                "title": {
                                    "name": "平台管理员",
                                    "name_en": "平台管理员",
                                    "name_zh_hans": "平台管理员"
                                }
                            }
                        }
                    },
                    "op_product_manager": {
                        "translations": {
                            "name": "运营产品负责人",
                            "name_en": "运营产品负责人",
                            "name_zh_hans": "运营产品负责人"
                        },
                        "key": "op_product_manager",
                        "name": "运营产品负责人",
                        "type": "multiUser",
                        "is_builtin": false,
                        "jsonschema": {
                            "translations": {
                                "title": "",
                                "title_en": "",
                                "title_zh_hans": ""
                            },
                            "type": "array",
                            "title": "",
                            "number_unit": "",
                            "table_relation": null,
                            "into_todo": [

                            ],
                            "out_todo": [

                            ],
                            "attr_relation": null,
                            "itsm_jmespath": null,
                            "itsm_options": null,
                            "itsm_options_type": null,
                            "format": null,
                            "columns": null,
                            "properties": null,
                            "items": {
                                "translations": {
                                    "title": "",
                                    "title_en": "",
                                    "title_zh_hans": ""
                                },
                                "type": "string",
                                "title": "",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": "user",
                                "columns": null,
                                "properties": null,
                                "items": null
                            }
                        },
                        "meta": {
                            "key": "op_product_manager",
                            "desc": "",
                            "tips": "请输入",
                            "type": "multiUser",
                            "title": {
                                "value": "运营产品负责人",
                                "isHide": false
                            },
                            "permission": {
                                "readonly": [
                                    "title",
                                    "key",
                                    "tips",
                                    "desc"
                                ]
                            },
                            "translations": {
                                "title": {
                                    "name": "运营产品负责人",
                                    "name_en": "运营产品负责人",
                                    "name_zh_hans": "运营产品负责人"
                                }
                            }
                        }
                    },
                    "application_content": {
                        "translations": {

                        },
                        "key": "application_content",
                        "name": "申请内容",
                        "type": "textarea",
                        "is_builtin": false,
                        "jsonschema": {
                            "translations": {
                                "title": "",
                                "title_en": "",
                                "title_zh_hans": ""
                            },
                            "type": "string",
                            "title": "",
                            "number_unit": "",
                            "table_relation": null,
                            "into_todo": [

                            ],
                            "out_todo": [

                            ],
                            "attr_relation": null,
                            "itsm_jmespath": null,
                            "itsm_options": null,
                            "itsm_options_type": null,
                            "format": null,
                            "columns": null,
                            "properties": null,
                            "items": null
                        },
                        "meta": {
                            "key": "application_content",
                            "desc": "",
                            "tips": "请输入",
                            "type": "textarea",
                            "title": {
                                "value": "申请内容",
                                "isHide": false
                            },
                            "permission": {
                                "readonly": [
                                    "title",
                                    "key",
                                    "tips",
                                    "desc"
                                ]
                            }
                        }
                    }
                },
                "fields_order": [
                    "application_content",
                    "op_product_manager",
                    "platform_manager"
                ],
                "components": null
            }
        },
        {
            "key": "$FormModel20250616165800003101",
            "name": "hcm_create_resource",
            "desc": "HCM2.0 账号&资源接入",
            "portal_id": "DEFAULT",
            "app_id": "core",
            "translations": {
                "name": "hcm_create_resource",
                "name_en": "add_account",
                "desc": "HCM2.0 账号&资源接入",
                "desc_en": "HCM2.0 登记账号&资源接入账号录入"
            },
            "meta": {
                "fields": {
                    "account_manager": {
                        "translations": {

                        },
                        "key": "account_manager",
                        "name": "帐号负责人",
                        "type": "multiUser",
                        "is_builtin": false,
                        "jsonschema": {
                            "translations": {
                                "title": "",
                                "title_en": "",
                                "title_zh_hans": ""
                            },
                            "type": "array",
                            "title": "",
                            "number_unit": "",
                            "table_relation": null,
                            "into_todo": [

                            ],
                            "out_todo": [

                            ],
                            "attr_relation": null,
                            "itsm_jmespath": null,
                            "itsm_options": null,
                            "itsm_options_type": null,
                            "format": null,
                            "columns": null,
                            "properties": null,
                            "items": {
                                "translations": {
                                    "title": "",
                                    "title_en": "",
                                    "title_zh_hans": ""
                                },
                                "type": "string",
                                "title": "",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": "user",
                                "columns": null,
                                "properties": null,
                                "items": null
                            }
                        },
                        "meta": {
                            "key": "account_manager",
                            "desc": "",
                            "tips": "请输入",
                            "type": "multiUser",
                            "title": {
                                "value": "帐号负责人",
                                "isHide": false
                            },
                            "permission": {
                                "readonly": [
                                    "title",
                                    "key",
                                    "tips",
                                    "desc"
                                ]
                            }
                        }
                    },
                    "platform_manager": {
                        "translations": {
                            "name": "平台管理员",
                            "name_en": "平台管理员",
                            "name_zh_hans": "平台管理员"
                        },
                        "key": "platform_manager",
                        "name": "平台管理员",
                        "type": "multiUser",
                        "is_builtin": false,
                        "jsonschema": {
                            "translations": {
                                "title": "",
                                "title_en": "",
                                "title_zh_hans": ""
                            },
                            "type": "array",
                            "title": "",
                            "number_unit": "",
                            "table_relation": null,
                            "into_todo": [

                            ],
                            "out_todo": [

                            ],
                            "attr_relation": null,
                            "itsm_jmespath": null,
                            "itsm_options": null,
                            "itsm_options_type": null,
                            "format": null,
                            "columns": null,
                            "properties": null,
                            "items": {
                                "translations": {
                                    "title": "",
                                    "title_en": "",
                                    "title_zh_hans": ""
                                },
                                "type": "string",
                                "title": "",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": "user",
                                "columns": null,
                                "properties": null,
                                "items": null
                            }
                        },
                        "meta": {
                            "key": "platform_manager",
                            "desc": "",
                            "tips": "请输入",
                            "type": "multiUser",
                            "title": {
                                "value": "平台管理员",
                                "isHide": false
                            },
                            "permission": {
                                "readonly": [
                                    "title",
                                    "key",
                                    "tips",
                                    "desc"
                                ]
                            },
                            "translations": {
                                "title": {
                                    "name": "平台管理员",
                                    "name_en": "平台管理员",
                                    "name_zh_hans": "平台管理员"
                                }
                            }
                        }
                    },
                    "application_content": {
                        "translations": {

                        },
                        "key": "application_content",
                        "name": "申请内容",
                        "type": "textarea",
                        "is_builtin": false,
                        "jsonschema": {
                            "translations": {
                                "title": "",
                                "title_en": "",
                                "title_zh_hans": ""
                            },
                            "type": "string",
                            "title": "",
                            "number_unit": "",
                            "table_relation": null,
                            "into_todo": [

                            ],
                            "out_todo": [

                            ],
                            "attr_relation": null,
                            "itsm_jmespath": null,
                            "itsm_options": null,
                            "itsm_options_type": null,
                            "format": null,
                            "columns": null,
                            "properties": null,
                            "items": null
                        },
                        "meta": {
                            "id": "tmuSWGTE",
                            "key": "application_content",
                            "desc": "",
                            "tips": "请输入",
                            "type": "textarea",
                            "title": {
                                "value": "申请内容",
                                "isHide": false
                            },
                            "permission": {
                                "readonly": [
                                    "title",
                                    "key",
                                    "tips",
                                    "desc"
                                ]
                            }
                        }
                    }
                },
                "fields_order": [
                    "application_content",
                    "account_manager",
                    "platform_manager"
                ],
                "components": null
            }
        }
    ],
    "workflow_categories": [
        {
            "key": "$WorkflowCategory20250605162300000201",
            "name": "海垒-HCM",
            "app_id": "core",
            "portal_id": "DEFAULT",
            "ticket_sn_prefix": "HCM",
            "translations": {
                "name": "海垒-HCM",
                "name_en": "海垒-HCM"
            }
        }
    ],
    "workflows": [
        {
            "workflow": {
                "key": "$Workflow20250605162500002001",
                "name": "HCM2.0 二级账号申请",
                "portal_id": "DEFAULT",
                "desc": "HCM",
                "category": "$WorkflowCategory20250605162300000201",
                "translations": {
                    "name": "HCM2.0 二级账号申请",
                    "name_en": "HCM2.0 二级账号申请",
                    "desc": "HCM",
                    "desc_en": ""
                },
                "engine_pattern": "FORMAL",
                "form_model_key": "$FormModel20250605162400001301",
                "app_id": "core",
                "meta": {
                    "workflow_button": [
                        {
                            "translations": {
                                "name": "关闭",
                                "name_en": "Close",
                                "name_zh_hant": "關閉"
                            },
                            "key": "close",
                            "name": "关闭",
                            "enable": true,
                            "meta": {
                                "button_permission": [
                                    {
                                        "type": "ticket_role",
                                        "value": [
                                            "creator",
                                            "current_processors"
                                        ]
                                    }
                                ],
                                "urging_time": null
                            },
                            "extra": null
                        },
                        {
                            "translations": {
                                "name": "终止",
                                "name_en": "Terminate",
                                "name_zh_hant": "終止"
                            },
                            "key": "terminate",
                            "name": "终止",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": null
                        },
                        {
                            "translations": {
                                "name": "重新打开",
                                "name_en": "Reopen",
                                "name_zh_hant": "重新打開"
                            },
                            "key": "restart",
                            "name": "重新打开",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": null
                        },
                        {
                            "translations": {
                                "name": "催办",
                                "name_en": "Urge",
                                "name_zh_hant": "催辦"
                            },
                            "key": "urging",
                            "name": "催办",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": null
                        }
                    ]
                }
            },
            "version": {
                "key": "20250605162500004801",
                "workflow_key": "$Workflow20250605162500002001",
                "desc": null,
                "workflows": {
                    "$Workflow20250605162500002001": {
                        "translations": {
                            "name": "HCM2.0 二级账号申请",
                            "name_en": "HCM2.0 二级账号申请",
                            "name_zh_hans": "HCM2.0 二级账号申请"
                        },
                        "key": "$Workflow20250605162500002001",
                        "name": "HCM2.0 二级账号申请",
                        "desc": "",
                        "type": "",
                        "is_sub": false,
                        "activity_key": null,
                        "connecting_objects": {
                            "connectingobject_20250605162547_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250605162547_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject20250605162500003301",
                                "source_type": "activity",
                                "dest_key": "activityobject_20250605162540_1",
                                "dest_type": "activity"
                            },
                            "connectingobject_20250605162549_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250605162549_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject_20250605162540_1",
                                "source_type": "activity",
                                "dest_key": "activityobject_20250605162541_1",
                                "dest_type": "activity"
                            },
                            "connectingobject_20250605162551_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250605162551_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject_20250605162541_1",
                                "source_type": "activity",
                                "dest_key": "eventobject_20250605162543_1",
                                "dest_type": "event"
                            },
                            "connectingobject20250605162500006403": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject20250605162500006403",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "eventobject20250605162500003301",
                                "source_type": "event",
                                "dest_key": "activityobject20250605162500003301",
                                "dest_type": "activity"
                            }
                        },
                        "relations": [

                        ]
                    }
                },
                "activities": {
                    "activityobject_20250605162540_1": {
                        "translations": {
                            "name": "直接上级审批",
                            "name_en": "Approve Node",
                            "name_zh_hant": "審批節點"
                        },
                        "key": "activityobject_20250605162540_1",
                        "workflow_key": "$Workflow20250605162500002001",
                        "name": "直接上级审批",
                        "desc": "",
                        "type": "APPROVE_TASK",
                        "incomings": [
                            "connectingobject_20250605162547_1"
                        ],
                        "outgoings": [
                            "connectingobject_20250605162549_1"
                        ],
                        "meta": {
                            "label": "approve",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "update",
                                    "name": "更新",
                                    "enable": false,
                                    "translations": {
                                        "name": "更新",
                                        "name_en": "Update",
                                        "name_zh_hant": "更新"
                                    }
                                },
                                {
                                    "key": "save",
                                    "name": "保存",
                                    "enable": false,
                                    "translations": {
                                        "name": "保存",
                                        "name_en": "Save",
                                        "name_zh_hant": "保存"
                                    }
                                },
                                {
                                    "key": "deliver",
                                    "meta": {
                                        "ranges": [
                                            {
                                                "processors": [

                                                ],
                                                "processors_type": ""
                                            }
                                        ]
                                    },
                                    "name": "转单",
                                    "enable": true,
                                    "translations": {
                                        "name": "转单",
                                        "name_en": "Deliver",
                                        "name_zh_hant": "轉單"
                                    }
                                },
                                {
                                    "key": "signature",
                                    "meta": {
                                        "tips": "",
                                        "ranges": [
                                            {
                                                "processors": [

                                                ],
                                                "processors_type": ""
                                            }
                                        ],
                                        "patterns": [

                                        ],
                                        "translations": {

                                        }
                                    },
                                    "name": "加签",
                                    "enable": false,
                                    "translations": {
                                        "name": "加签",
                                        "name_en": "Signature",
                                        "name_zh_hant": "加簽"
                                    }
                                },
                                {
                                    "key": "back",
                                    "meta": {
                                        "pattern": "again",
                                        "activities": [
                                            "all"
                                        ]
                                    },
                                    "name": "退回",
                                    "enable": false,
                                    "translations": {
                                        "name": "退回",
                                        "name_en": "Return",
                                        "name_zh_hant": "退回"
                                    }
                                },
                                {
                                    "key": "approve",
                                    "meta": {
                                        "placeholder": "请输入"
                                    },
                                    "name": "同意",
                                    "enable": true,
                                    "translations": {
                                        "name": "同意",
                                        "name_en": "Approve",
                                        "name_zh_hant": "同意"
                                    }
                                },
                                {
                                    "key": "refuse",
                                    "meta": {
                                        "placeholder": "请输入"
                                    },
                                    "name": "拒绝",
                                    "enable": true,
                                    "translations": {
                                        "name": "拒绝",
                                        "name_en": "Refuse",
                                        "name_zh_hant": "拒絕"
                                    }
                                }
                            ],
                            "processors": "get_variable(ticket_id, \"TICKET\", \"creator\", \"user_leader($tag)\" ,\"List[User]\")",
                            "working_mode": "cooperate",
                            "processors_type": "feel"
                        },
                        "hooks": [

                        ]
                    },
                    "activityobject_20250605162541_1": {
                        "translations": {
                            "name": "平台管理员审批",
                            "name_en": "Approve Node",
                            "name_zh_hant": "審批節點"
                        },
                        "key": "activityobject_20250605162541_1",
                        "workflow_key": "$Workflow20250605162500002001",
                        "name": "平台管理员审批",
                        "desc": "",
                        "type": "APPROVE_TASK",
                        "incomings": [
                            "connectingobject_20250605162549_1"
                        ],
                        "outgoings": [
                            "connectingobject_20250605162551_1"
                        ],
                        "meta": {
                            "label": "approve",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "approve",
                                    "meta": {

                                    },
                                    "name": "同意",
                                    "enable": true,
                                    "translations": {
                                        "name": "同意",
                                        "name_en": "Approve",
                                        "name_zh_hant": "同意"
                                    }
                                },
                                {
                                    "key": "refuse",
                                    "meta": {

                                    },
                                    "name": "拒绝",
                                    "enable": true,
                                    "translations": {
                                        "name": "拒绝",
                                        "name_en": "Refuse",
                                        "name_zh_hant": "拒絕"
                                    }
                                }
                            ],
                            "processors": [
                                {
                                    "id": "DATA_TABLE[platform_manager]",
                                    "key": "DATA_TABLE",
                                    "feel": "",
                                    "name": "平台管理员",
                                    "path": "platform_manager",
                                    "type": "List[User]",
                                    "default": null,
                                    "variables": null,
                                    "jsonschema": null,
                                    "translations": {
                                        "name": "平台管理员",
                                        "name_en": "平台管理员",
                                        "name_zh_hans": "平台管理员"
                                    }
                                }
                            ],
                            "working_mode": "cooperate",
                            "processors_type": "user"
                        },
                        "hooks": [

                        ]
                    },
                    "activityobject20250605162500003301": {
                        "translations": {
                            "name": "提单",
                            "name_en": "Submit",
                            "name_zh_hans": "提单"
                        },
                        "key": "activityobject20250605162500003301",
                        "workflow_key": "$Workflow20250605162500002001",
                        "name": "提单",
                        "desc": "",
                        "type": "SUBMIT",
                        "incomings": [
                            "connectingobject20250605162500006403"
                        ],
                        "outgoings": [
                            "connectingobject_20250605162547_1"
                        ],
                        "meta": {
                            "label": "submit",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "submit",
                                    "name": "提交",
                                    "enable": true,
                                    "translations": {
                                        "name": "提交",
                                        "name_en": "Submit",
                                        "name_zh_hant": "提交"
                                    }
                                },
                                {
                                    "key": "save_draft",
                                    "name": "保存草稿",
                                    "enable": true,
                                    "translations": {
                                        "name": "保存草稿",
                                        "name_en": "Save Draft",
                                        "name_zh_hant": "保存草稿"
                                    }
                                },
                                {
                                    "key": "save_template",
                                    "name": "保存模板",
                                    "enable": true,
                                    "translations": {
                                        "name": "保存模板",
                                        "name_en": "Save Template",
                                        "name_zh_hant": "保存模板"
                                    }
                                }
                            ]
                        },
                        "hooks": [

                        ]
                    }
                },
                "events": {
                    "eventobject_20250605162543_1": {
                        "translations": {
                            "name": "结束",
                            "name_en": "End",
                            "name_zh_hant": "結束"
                        },
                        "key": "eventobject_20250605162543_1",
                        "workflow_key": "$Workflow20250605162500002001",
                        "name": "结束",
                        "desc": "",
                        "type": "end",
                        "incomings": [
                            "connectingobject_20250605162551_1"
                        ],
                        "outgoings": [

                        ],
                        "meta": {

                        }
                    },
                    "eventobject20250605162500003301": {
                        "translations": {
                            "name": "开始",
                            "name_en": "Start",
                            "name_zh_hans": "开始"
                        },
                        "key": "eventobject20250605162500003301",
                        "workflow_key": "$Workflow20250605162500002001",
                        "name": "开始",
                        "desc": "",
                        "type": "start",
                        "incomings": [

                        ],
                        "outgoings": [
                            "connectingobject20250605162500006403"
                        ],
                        "meta": {

                        }
                    }
                },
                "gateways": {

                },
                "meta": {
                    "ticket_button": {
                        "action_button": [
                            {
                                "translations": {
                                    "name": "撤回",
                                    "name_en": "Withdraw",
                                    "name_zh_hant": "撤回"
                                },
                                "key": "withdraw",
                                "name": "撤回",
                                "enable": true,
                                "extra": null,
                                "meta": {
                                    "can_withdraw_activity": [
                                        "activityobject_20250605162540_1",
                                        "activityobject_20250605162541_1"
                                    ]
                                }
                            },
                            {
                                "translations": {
                                    "name": "挂起",
                                    "name_en": "Suspend",
                                    "name_zh_hant": "掛起"
                                },
                                "key": "suspend",
                                "name": "挂起",
                                "enable": true,
                                "extra": null
                            },
                            {
                                "translations": {
                                    "name": "恢复",
                                    "name_en": "Restore",
                                    "name_zh_hant": "恢復"
                                },
                                "key": "recovery",
                                "name": "恢复",
                                "enable": true,
                                "extra": null
                            },
                            {
                                "translations": {
                                    "name": "转建文章",
                                    "name_en": "Dump Article",
                                    "name_zh_hant": "轉建文章"
                                },
                                "key": "convert",
                                "name": "转建文章",
                                "enable": false,
                                "extra": null,
                                "meta": {
                                    "can_convert_activity": [

                                    ]
                                }
                            }
                        ]
                    },
                    "custom_button": [

                    ],
                    "vips": [

                    ],
                    "stage": {
                        "is_enable": false,
                        "model": "",
                        "config": {

                        }
                    },
                    "variables": [

                    ]
                },
                "form_canvas_data": {
                    "form_data": {
                        "id": "form_xYCKNjUwS0",
                        "type": "form",
                        "align": "top",
                        "class": [

                        ],
                        "rules": [

                        ],
                        "layout": [
                            {
                                "list": [
                                    {
                                        "key": "ticket__title",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "text",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "标题",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "ticket__title",
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "verification.required"
                                            ],
                                            "noOperate": [
                                                "delete",
                                                "copy"
                                            ]
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "标题",
                                                "name_en": "Title",
                                                "name_zh_hant": "標題"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": true,
                                                "enabled": true
                                            },
                                            "wordLimit": {
                                                "value": {
                                                    "max": "",
                                                    "min": 0
                                                },
                                                "enabled": false
                                            },
                                            "formatLimit": {
                                                "value": {
                                                    "errorTips": "",
                                                    "expression": ""
                                                },
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "key": "application_content",
                                        "desc": "",
                                        "rows": 4,
                                        "tips": "请输入",
                                        "type": "textarea",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "申请内容",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "hAllpwXD",
                                        "location": "form",
                                        "sceneKey": "application_content",
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "多行文本",
                                                "name_en": "Plain text area",
                                                "name_zh_hant": "多行文本"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": false,
                                                "enabled": false
                                            },
                                            "wordLimit": {
                                                "value": {
                                                    "max": "",
                                                    "min": 0
                                                },
                                                "enabled": false
                                            },
                                            "formatLimit": {
                                                "value": "",
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "key": "op_product_manager",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "multiUser",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "运营产品负责人",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "FhEZeZQf",
                                        "location": "form",
                                        "sceneKey": "op_product_manager",
                                        "userScope": {
                                            "type": "custom",
                                            "value": {

                                            },
                                            "isRecursive": false
                                        },
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "userConfig": {
                                            "setSelf": false,
                                            "multiple": true,
                                            "userInfo": [

                                            ],
                                            "showStyle": "flat",
                                            "showVipIcon": true,
                                            "isShowUserInfo": false
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "运营产品负责人",
                                                "name_en": "运营产品负责人",
                                                "name_zh_hans": "运营产品负责人"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": false,
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "key": "platform_manager",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "multiUser",
                                        "class": [

                                        ],
                                        "state": "hide",
                                        "title": {
                                            "value": "平台管理员",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "K8jGcLkF",
                                        "location": "form",
                                        "sceneKey": "platform_manager",
                                        "userScope": {
                                            "type": "custom",
                                            "value": {

                                            },
                                            "isRecursive": false
                                        },
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "userConfig": {
                                            "setSelf": false,
                                            "multiple": true,
                                            "userInfo": [

                                            ],
                                            "showStyle": "flat",
                                            "showVipIcon": true,
                                            "isShowUserInfo": false
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "平台管理员",
                                                "name_en": "平台管理员",
                                                "name_zh_hans": "平台管理员"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": false,
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            }
                        ],
                        "styleCode": "",
                        "dataLinkage": [

                        ],
                        "verification": [

                        ]
                    },
                    "jsonschema": {
                        "type": "object",
                        "properties": {
                            "ticket__title": {
                                "translations": {
                                    "title": "标题",
                                    "title_en": "Title",
                                    "title_zh_hans": "标题"
                                },
                                "type": "string",
                                "title": "标题",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": null
                            },
                            "platform_manager": {
                                "translations": {
                                    "title": "平台管理员",
                                    "title_en": "平台管理员",
                                    "title_zh_hans": "平台管理员"
                                },
                                "type": "array",
                                "title": "平台管理员",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": {
                                    "translations": {
                                        "title": "",
                                        "title_en": "",
                                        "title_zh_hans": ""
                                    },
                                    "type": "string",
                                    "title": "",
                                    "number_unit": "",
                                    "table_relation": null,
                                    "into_todo": [

                                    ],
                                    "out_todo": [

                                    ],
                                    "attr_relation": null,
                                    "itsm_jmespath": null,
                                    "itsm_options": null,
                                    "itsm_options_type": null,
                                    "format": "user",
                                    "columns": null,
                                    "properties": null,
                                    "items": null
                                }
                            },
                            "op_product_manager": {
                                "translations": {
                                    "title": "运营产品负责人",
                                    "title_en": "运营产品负责人",
                                    "title_zh_hans": "运营产品负责人"
                                },
                                "type": "array",
                                "title": "运营产品负责人",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": {
                                    "translations": {
                                        "title": "",
                                        "title_en": "",
                                        "title_zh_hans": ""
                                    },
                                    "type": "string",
                                    "title": "",
                                    "number_unit": "",
                                    "table_relation": null,
                                    "into_todo": [

                                    ],
                                    "out_todo": [

                                    ],
                                    "attr_relation": null,
                                    "itsm_jmespath": null,
                                    "itsm_options": null,
                                    "itsm_options_type": null,
                                    "format": "user",
                                    "columns": null,
                                    "properties": null,
                                    "items": null
                                }
                            },
                            "application_content": {
                                "translations": {
                                    "title": "多行文本",
                                    "title_en": "Plain text area",
                                    "title_zh_hans": "多行文本"
                                },
                                "type": "string",
                                "title": "多行文本",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": null
                            }
                        },
                        "additionalProperties": false
                    },
                    "decision_table_relations": [

                    ],
                    "datasheet_table_relations": [

                    ]
                },
                "flow_canvas_data": {
                    "data": [
                        {
                            "id": "connectingobject20250605162500006403",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "eventobject20250605162500003301",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject20250605162500003301",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250605162547_1",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject20250605162500003301",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject_20250605162540_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250605162549_1",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject_20250605162540_1",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject_20250605162541_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250605162551_1",
                            "attrs": {
                                "line": {
                                    "stroke": "#1272FF"
                                }
                            },
                            "shape": "sequence_flow",
                            "tools": {
                                "name": null,
                                "items": [
                                    {
                                        "args": {
                                            "markup": [
                                                {
                                                    "attrs": {
                                                        "width": 24,
                                                        "height": 24
                                                    },
                                                    "tagName": "foreignObject",
                                                    "children": [
                                                        {
                                                            "ns": "http://www.w3.org/1999/xhtml",
                                                            "style": {
                                                                "width": "100%",
                                                                "height": "100%",
                                                                "min-width": 0
                                                            },
                                                            "tagName": "body",
                                                            "children": [
                                                                {
                                                                    "style": {
                                                                        "color": "#fff",
                                                                        "width": "100%",
                                                                        "height": "100%",
                                                                        "display": "flex",
                                                                        "background": "#1272FF",
                                                                        "align-items": "center",
                                                                        "line-height": 1,
                                                                        "border-radius": "4px",
                                                                        "justify-content": "center"
                                                                    },
                                                                    "tagName": "div",
                                                                    "className": "cw-icon cw-icon-delete"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "style": {
                                                        "display": "none"
                                                    },
                                                    "tagName": "p"
                                                }
                                            ],
                                            "offset": {
                                                "x": 0,
                                                "y": -12
                                            },
                                            "rotate": true,
                                            "distance": "50%"
                                        },
                                        "name": "button-remove"
                                    },
                                    {
                                        "args": {
                                            "attrs": {
                                                "fill": "#1272FF"
                                            }
                                        },
                                        "name": "target-arrowhead"
                                    },
                                    {
                                        "args": {
                                            "attrs": {
                                                "fill": "#1272FF"
                                            }
                                        },
                                        "name": "vertices"
                                    },
                                    {
                                        "args": {
                                            "markup": [
                                                {
                                                    "attrs": {
                                                        "width": 24,
                                                        "height": 24
                                                    },
                                                    "tagName": "foreignObject",
                                                    "children": [
                                                        {
                                                            "ns": "http://www.w3.org/1999/xhtml",
                                                            "style": {
                                                                "width": "100%",
                                                                "height": "100%",
                                                                "min-width": 0
                                                            },
                                                            "tagName": "body",
                                                            "children": [
                                                                {
                                                                    "style": {
                                                                        "color": "#fff",
                                                                        "width": "100%",
                                                                        "cursor": "pointer",
                                                                        "height": "100%",
                                                                        "display": "flex",
                                                                        "background": "#1272FF",
                                                                        "align-items": "center",
                                                                        "line-height": 1,
                                                                        "border-radius": "4px",
                                                                        "justify-content": "center"
                                                                    },
                                                                    "tagName": "div",
                                                                    "className": "cw-icon cw-icon-edit"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "style": {
                                                        "display": "none"
                                                    },
                                                    "tagName": "p"
                                                }
                                            ],
                                            "offset": {
                                                "x": -26,
                                                "y": -12
                                            },
                                            "rotate": true,
                                            "distance": "50%"
                                        },
                                        "name": "button"
                                    },
                                    {
                                        "args": {
                                            "markup": [
                                                {
                                                    "attrs": {
                                                        "width": 24,
                                                        "height": 24
                                                    },
                                                    "tagName": "foreignObject",
                                                    "children": [
                                                        {
                                                            "ns": "http://www.w3.org/1999/xhtml",
                                                            "style": {
                                                                "width": "100%",
                                                                "height": "100%",
                                                                "min-width": 0
                                                            },
                                                            "tagName": "body",
                                                            "children": [
                                                                {
                                                                    "style": {
                                                                        "color": "#fff",
                                                                        "width": "100%",
                                                                        "height": "100%",
                                                                        "display": "flex",
                                                                        "background": "#1272FF",
                                                                        "align-items": "center",
                                                                        "line-height": 1,
                                                                        "border-radius": "4px",
                                                                        "justify-content": "center"
                                                                    },
                                                                    "tagName": "div",
                                                                    "className": "cw-icon cw-icon-delete"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "style": {
                                                        "display": "none"
                                                    },
                                                    "tagName": "p"
                                                }
                                            ],
                                            "offset": {
                                                "x": 0,
                                                "y": -12
                                            },
                                            "rotate": true,
                                            "distance": "50%"
                                        },
                                        "name": "button-remove"
                                    },
                                    {
                                        "args": {
                                            "attrs": {
                                                "fill": "#1272FF"
                                            }
                                        },
                                        "name": "target-arrowhead"
                                    },
                                    {
                                        "args": {
                                            "attrs": {
                                                "fill": "#1272FF"
                                            }
                                        },
                                        "name": "vertices"
                                    },
                                    {
                                        "args": {
                                            "markup": [
                                                {
                                                    "attrs": {
                                                        "width": 24,
                                                        "height": 24
                                                    },
                                                    "tagName": "foreignObject",
                                                    "children": [
                                                        {
                                                            "ns": "http://www.w3.org/1999/xhtml",
                                                            "style": {
                                                                "width": "100%",
                                                                "height": "100%",
                                                                "min-width": 0
                                                            },
                                                            "tagName": "body",
                                                            "children": [
                                                                {
                                                                    "style": {
                                                                        "color": "#fff",
                                                                        "width": "100%",
                                                                        "cursor": "pointer",
                                                                        "height": "100%",
                                                                        "display": "flex",
                                                                        "background": "#1272FF",
                                                                        "align-items": "center",
                                                                        "line-height": 1,
                                                                        "border-radius": "4px",
                                                                        "justify-content": "center"
                                                                    },
                                                                    "tagName": "div",
                                                                    "className": "cw-icon cw-icon-edit"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "style": {
                                                        "display": "none"
                                                    },
                                                    "tagName": "p"
                                                }
                                            ],
                                            "offset": {
                                                "x": -26,
                                                "y": -12
                                            },
                                            "rotate": true,
                                            "distance": "50%"
                                        },
                                        "name": "button"
                                    }
                                ]
                            },
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject_20250605162541_1",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "eventobject_20250605162543_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "eventobject20250605162500003301",
                            "data": {
                                "x": 20,
                                "y": 180,
                                "id": "eventobject20250605162500003301",
                                "icon": "cw-icon cw-icon-kai-shi",
                                "meta": "$.events.eventobject20250605162500003301.meta",
                                "name": "$.events.eventobject20250605162500003301.name",
                                "type": "start",
                                "width": 80,
                                "height": 40,
                                "zIndex": 2,
                                "isError": [

                                ],
                                "isFinished": false,
                                "isSelected": false,
                                "translations": "$.events.eventobject20250605162500003301.translations"
                            },
                            "size": {
                                "width": 80,
                                "height": 40
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-right": {
                                        "position": "right"
                                    }
                                }
                            },
                            "shape": "events",
                            "zIndex": 2,
                            "position": {
                                "x": 20,
                                "y": 180
                            }
                        },
                        {
                            "id": "activityobject20250605162500003301",
                            "data": {
                                "x": 180,
                                "y": 160,
                                "id": "activityobject20250605162500003301",
                                "icon": "cw-icon cw-icon-shen-qing",
                                "meta": "$.activities.activityobject20250605162500003301.meta",
                                "name": "$.activities.activityobject20250605162500003301.name",
                                "type": "SUBMIT",
                                "width": 200,
                                "height": 80,
                                "zIndex": 2,
                                "isError": [

                                ],
                                "nodeType": "activities",
                                "isFinished": true,
                                "isSelected": false,
                                "configurable": true,
                                "translations": "$.activities.activityobject20250605162500003301.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 180,
                                "y": 160
                            }
                        },
                        {
                            "id": "activityobject_20250605162540_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "activityobject_20250605162540_1",
                                "code": "APPROVE_TASK",
                                "icon": "cw-icon cw-icon-shen-pi",
                                "meta": "$.activities.activityobject_20250605162540_1.meta",
                                "name": "$.activities.activityobject_20250605162540_1.name",
                                "type": "APPROVE_TASK",
                                "color": [
                                    "#FFE5C7",
                                    "#FD9D2C"
                                ],
                                "label": "审批节点",
                                "width": 200,
                                "config": {
                                    "type": "tab",
                                    "tabList": [
                                        {
                                            "type": "approval",
                                            "label": "审批对象",
                                            "isError": false
                                        },
                                        {
                                            "meta": {
                                                "buttons": [
                                                    {
                                                        "key": "approve",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "同意",
                                                        "label": "同意",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "refuse",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "拒绝",
                                                        "label": "拒绝",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "update",
                                                        "name": "更新",
                                                        "label": "更新",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "save",
                                                        "name": "保存",
                                                        "label": "保存",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "deliver",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ]
                                                        },
                                                        "name": "转单",
                                                        "label": "转单",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "signature",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ],
                                                            "patterns": [

                                                            ]
                                                        },
                                                        "name": "加签",
                                                        "label": "加签",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "back",
                                                        "meta": {
                                                            "pattern": "again",
                                                            "activities": [
                                                                "all"
                                                            ]
                                                        },
                                                        "name": "退回",
                                                        "label": "退回",
                                                        "switch": false,
                                                        "disabled": false
                                                    }
                                                ]
                                            },
                                            "type": "operate",
                                            "label": "操作按钮",
                                            "isError": false
                                        },
                                        {
                                            "type": "fields",
                                            "label": "字段配置",
                                            "isError": false
                                        }
                                    ]
                                },
                                "height": 80,
                                "isError": [

                                ],
                                "toolbar": [
                                    "copy",
                                    "delete"
                                ],
                                "dataType": "activities",
                                "nodeType": "activities",
                                "isDisabled": false,
                                "isFinished": true,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {
                                        "label": "approve",
                                        "fields": {

                                        },
                                        "buttons": [
                                            {
                                                "key": "approve",
                                                "name": "同意",
                                                "translations": {
                                                    "name": "同意",
                                                    "name_en": "Approve",
                                                    "name_zh_hant": "同意"
                                                }
                                            },
                                            {
                                                "key": "refuse",
                                                "name": "拒绝",
                                                "translations": {
                                                    "name": "拒绝",
                                                    "name_en": "Refuse",
                                                    "name_zh_hant": "拒絕"
                                                }
                                            }
                                        ],
                                        "processors": [

                                        ],
                                        "working_mode": "serial",
                                        "processors_type": ""
                                    },
                                    "name": "审批节点",
                                    "translations": {
                                        "name": "审批节点",
                                        "name_en": "Approve Node",
                                        "name_zh_hant": "審批節點"
                                    }
                                },
                                "configurable": true,
                                "translations": "$.activities.activityobject_20250605162540_1.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 460,
                                "y": 160
                            }
                        },
                        {
                            "id": "activityobject_20250605162541_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "activityobject_20250605162541_1",
                                "code": "APPROVE_TASK",
                                "icon": "cw-icon cw-icon-shen-pi",
                                "meta": "$.activities.activityobject_20250605162541_1.meta",
                                "name": "$.activities.activityobject_20250605162541_1.name",
                                "type": "APPROVE_TASK",
                                "color": [
                                    "#FFE5C7",
                                    "#FD9D2C"
                                ],
                                "label": "审批节点",
                                "width": 200,
                                "config": {
                                    "type": "tab",
                                    "tabList": [
                                        {
                                            "type": "approval",
                                            "label": "审批对象",
                                            "isError": false
                                        },
                                        {
                                            "meta": {
                                                "buttons": [
                                                    {
                                                        "key": "approve",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "同意",
                                                        "label": "同意",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "refuse",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "拒绝",
                                                        "label": "拒绝",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "update",
                                                        "name": "更新",
                                                        "label": "更新",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "save",
                                                        "name": "保存",
                                                        "label": "保存",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "deliver",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ]
                                                        },
                                                        "name": "转单",
                                                        "label": "转单",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "signature",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ],
                                                            "patterns": [

                                                            ]
                                                        },
                                                        "name": "加签",
                                                        "label": "加签",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "back",
                                                        "meta": {
                                                            "pattern": "again",
                                                            "activities": [
                                                                "all"
                                                            ]
                                                        },
                                                        "name": "退回",
                                                        "label": "退回",
                                                        "switch": false,
                                                        "disabled": false
                                                    }
                                                ]
                                            },
                                            "type": "operate",
                                            "label": "操作按钮",
                                            "isError": false
                                        },
                                        {
                                            "type": "fields",
                                            "label": "字段配置",
                                            "isError": false
                                        }
                                    ]
                                },
                                "height": 80,
                                "isError": [

                                ],
                                "toolbar": [
                                    "copy",
                                    "delete"
                                ],
                                "dataType": "activities",
                                "nodeType": "activities",
                                "isDisabled": false,
                                "isFinished": true,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {
                                        "label": "approve",
                                        "fields": {

                                        },
                                        "buttons": [
                                            {
                                                "key": "approve",
                                                "name": "同意",
                                                "translations": {
                                                    "name": "同意",
                                                    "name_en": "Approve",
                                                    "name_zh_hant": "同意"
                                                }
                                            },
                                            {
                                                "key": "refuse",
                                                "name": "拒绝",
                                                "translations": {
                                                    "name": "拒绝",
                                                    "name_en": "Refuse",
                                                    "name_zh_hant": "拒絕"
                                                }
                                            }
                                        ],
                                        "processors": [

                                        ],
                                        "working_mode": "serial",
                                        "processors_type": ""
                                    },
                                    "name": "审批节点",
                                    "translations": {
                                        "name": "审批节点",
                                        "name_en": "Approve Node",
                                        "name_zh_hant": "審批節點"
                                    }
                                },
                                "configurable": true,
                                "translations": "$.activities.activityobject_20250605162541_1.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 740,
                                "y": 160
                            }
                        },
                        {
                            "id": "eventobject_20250605162543_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "eventobject_20250605162543_1",
                                "code": "end",
                                "icon": "cw-icon cw-icon-jie-shu",
                                "meta": "$.events.eventobject_20250605162543_1.meta",
                                "name": "$.events.eventobject_20250605162543_1.name",
                                "type": "end",
                                "label": "结束",
                                "width": 80,
                                "height": 40,
                                "isError": [

                                ],
                                "toolbar": [
                                    "delete"
                                ],
                                "dataType": "events",
                                "nodeType": "events",
                                "isDisabled": false,
                                "isFinished": false,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {

                                    },
                                    "name": "结束",
                                    "translations": {
                                        "name": "结束",
                                        "name_en": "End",
                                        "name_zh_hant": "結束"
                                    }
                                },
                                "translations": "$.events.eventobject_20250605162543_1.translations"
                            },
                            "size": {
                                "width": 80,
                                "height": 40
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "events",
                            "zIndex": 2,
                            "position": {
                                "x": 1020,
                                "y": 180
                            }
                        }
                    ]
                },
                "normal_pattern_meta": null
            }
        },
        {
            "workflow": {
                "key": "$Workflow20250616170200004001",
                "name": "HCM2.0 登记账号&资源接入账号录入",
                "portal_id": "DEFAULT",
                "desc": "HCM",
                "category": "$WorkflowCategory20250605162300000201",
                "translations": {
                    "name": "HCM2.0 登记账号&资源接入账号录入",
                    "name_en": "HCM2.0 登记账号&资源接入账号录入",
                    "desc": "HCM",
                    "desc_en": "HCM"
                },
                "engine_pattern": "FORMAL",
                "form_model_key": "$FormModel20250616165800003101",
                "app_id": "core",
                "meta": {
                    "workflow_button": [
                        {
                            "translations": {
                                "name": "关闭",
                                "name_en": "Close",
                                "name_zh_hant": "關閉"
                            },
                            "key": "close",
                            "name": "关闭",
                            "enable": true,
                            "meta": {
                                "button_permission": [
                                    {
                                        "type": "ticket_role",
                                        "value": [
                                            "creator",
                                            "current_processors"
                                        ]
                                    }
                                ],
                                "urging_time": null
                            },
                            "extra": {

                            }
                        },
                        {
                            "translations": {
                                "name": "终止",
                                "name_en": "Cancel",
                                "name_zh_hant": "終止"
                            },
                            "key": "terminate",
                            "name": "终止",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": {

                            }
                        },
                        {
                            "translations": {
                                "name": "重新打开",
                                "name_en": "Reopen",
                                "name_zh_hant": "重新打開"
                            },
                            "key": "restart",
                            "name": "重新打开",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": {

                            }
                        },
                        {
                            "translations": {
                                "name": "催办",
                                "name_en": "Urge",
                                "name_zh_hant": "催辦"
                            },
                            "key": "urging",
                            "name": "催办",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": {

                            }
                        }
                    ]
                }
            },
            "version": {
                "key": "20250616170200008901",
                "workflow_key": "$Workflow20250616170200004001",
                "desc": null,
                "workflows": {
                    "$Workflow20250616170200004001": {
                        "translations": {
                            "name": "HCM2.0 登记账号&资源接入账号录入",
                            "name_en": "HCM2.0 登记账号&资源接入账号录入",
                            "name_zh_hans": "HCM2.0 登记账号&资源接入账号录入"
                        },
                        "key": "$Workflow20250616170200004001",
                        "name": "HCM2.0 登记账号&资源接入账号录入",
                        "desc": "HCM",
                        "type": "",
                        "is_sub": false,
                        "activity_key": null,
                        "connecting_objects": {
                            "connectingobject_20250616170309_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250616170309_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject20250616170200005901",
                                "source_type": "activity",
                                "dest_key": "activityobject_20250616170307_1",
                                "dest_type": "activity"
                            },
                            "connectingobject_20250616170322_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250616170322_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject_20250616170307_1",
                                "source_type": "activity",
                                "dest_key": "eventobject_20250616170319_1",
                                "dest_type": "event"
                            },
                            "connectingobject20250616170200013101": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject20250616170200013101",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "eventobject20250616170200005901",
                                "source_type": "event",
                                "dest_key": "activityobject20250616170200005901",
                                "dest_type": "activity"
                            }
                        },
                        "relations": [

                        ]
                    }
                },
                "activities": {
                    "activityobject_20250616170307_1": {
                        "translations": {
                            "name": "平台管理员审批",
                            "name_en": "Approve Node",
                            "name_zh_hans": "平台管理员审批"
                        },
                        "key": "activityobject_20250616170307_1",
                        "workflow_key": "$Workflow20250616170200004001",
                        "name": "平台管理员审批",
                        "desc": "",
                        "type": "APPROVE_TASK",
                        "incomings": [
                            "connectingobject_20250616170309_1"
                        ],
                        "outgoings": [
                            "connectingobject_20250616170322_1"
                        ],
                        "meta": {
                            "label": "approve",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "approve",
                                    "meta": {

                                    },
                                    "name": "同意",
                                    "extra": {
                                        "theme": "success"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "同意",
                                        "name_en": "Approve",
                                        "name_zh_hant": "同意"
                                    }
                                },
                                {
                                    "key": "refuse",
                                    "meta": {

                                    },
                                    "name": "拒绝",
                                    "extra": {
                                        "theme": "danger"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "拒绝",
                                        "name_en": "Reject",
                                        "name_zh_hant": "拒絕"
                                    }
                                }
                            ],
                            "processors": [
                                {
                                    "id": "DATA_TABLE[platform_manager]",
                                    "key": "DATA_TABLE",
                                    "feel": "",
                                    "name": "平台管理员",
                                    "path": "platform_manager",
                                    "type": "List[User]",
                                    "default": null,
                                    "variables": null,
                                    "jsonschema": null,
                                    "translations": {
                                        "name": "平台管理员",
                                        "name_en": "平台管理员",
                                        "name_zh_hans": "平台管理员"
                                    }
                                }
                            ],
                            "working_mode": "cooperate",
                            "processors_type": "user",
                            "advanced_settings": {
                                "auto_terminate": false
                            }
                        },
                        "hooks": [

                        ]
                    },
                    "activityobject20250616170200005901": {
                        "translations": {
                            "name": "提单",
                            "name_en": "Submit",
                            "name_zh_hans": "提单"
                        },
                        "key": "activityobject20250616170200005901",
                        "workflow_key": "$Workflow20250616170200004001",
                        "name": "提单",
                        "desc": "",
                        "type": "SUBMIT",
                        "incomings": [
                            "connectingobject20250616170200013101"
                        ],
                        "outgoings": [
                            "connectingobject_20250616170309_1"
                        ],
                        "meta": {
                            "label": "submit",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "submit",
                                    "name": "提交",
                                    "extra": {
                                        "theme": "primary"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "提交",
                                        "name_en": "Submit",
                                        "name_zh_hant": "提交"
                                    }
                                },
                                {
                                    "key": "save_draft",
                                    "name": "保存草稿",
                                    "extra": null,
                                    "enable": true,
                                    "translations": {
                                        "name": "保存草稿",
                                        "name_en": "Save Draft",
                                        "name_zh_hant": "保存草稿"
                                    }
                                },
                                {
                                    "key": "save_template",
                                    "name": "保存模板",
                                    "extra": null,
                                    "enable": true,
                                    "translations": {
                                        "name": "保存模板",
                                        "name_en": "Save Template",
                                        "name_zh_hant": "保存模板"
                                    }
                                }
                            ]
                        },
                        "hooks": [

                        ]
                    }
                },
                "events": {
                    "eventobject_20250616170319_1": {
                        "translations": {
                            "name": "结束",
                            "name_en": "End",
                            "name_zh_hant": "結束"
                        },
                        "key": "eventobject_20250616170319_1",
                        "workflow_key": "$Workflow20250616170200004001",
                        "name": "结束",
                        "desc": "",
                        "type": "end",
                        "incomings": [
                            "connectingobject_20250616170322_1"
                        ],
                        "outgoings": [

                        ],
                        "meta": {

                        }
                    },
                    "eventobject20250616170200005901": {
                        "translations": {
                            "name": "开始",
                            "name_en": "Start",
                            "name_zh_hans": "开始"
                        },
                        "key": "eventobject20250616170200005901",
                        "workflow_key": "$Workflow20250616170200004001",
                        "name": "开始",
                        "desc": "",
                        "type": "start",
                        "incomings": [

                        ],
                        "outgoings": [
                            "connectingobject20250616170200013101"
                        ],
                        "meta": {

                        }
                    }
                },
                "gateways": {

                },
                "meta": {
                    "ticket_button": {
                        "action_button": [
                            {
                                "translations": {
                                    "name": "撤回",
                                    "name_en": "Withdraw",
                                    "name_zh_hant": "撤回"
                                },
                                "key": "withdraw",
                                "name": "撤回",
                                "enable": true,
                                "extra": {

                                },
                                "meta": {
                                    "can_withdraw_activity": [
                                        "activityobject_20250616170307_1"
                                    ]
                                }
                            },
                            {
                                "translations": {
                                    "name": "挂起",
                                    "name_en": "Suspend",
                                    "name_zh_hant": "掛起"
                                },
                                "key": "suspend",
                                "name": "挂起",
                                "enable": true,
                                "extra": {

                                }
                            },
                            {
                                "translations": {
                                    "name": "恢复",
                                    "name_en": "Restore",
                                    "name_zh_hant": "恢復"
                                },
                                "key": "recovery",
                                "name": "恢复",
                                "enable": true,
                                "extra": {

                                }
                            },
                            {
                                "translations": {
                                    "name": "转建文章",
                                    "name_en": "Create article",
                                    "name_zh_hant": "轉建文章"
                                },
                                "key": "convert",
                                "name": "转建文章",
                                "enable": false,
                                "extra": {

                                },
                                "meta": {
                                    "can_convert_activity": [

                                    ]
                                }
                            }
                        ]
                    },
                    "custom_button": [

                    ],
                    "vips": [

                    ],
                    "stage": {
                        "is_enable": false,
                        "model": "",
                        "config": {

                        }
                    },
                    "variables": [

                    ]
                },
                "form_canvas_data": {
                    "form_data": {
                        "id": "form_ZklY9V72DA",
                        "type": "form",
                        "align": "top",
                        "class": [

                        ],
                        "rules": [

                        ],
                        "layout": [
                            {
                                "list": [
                                    {
                                        "key": "ticket__title",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "text",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "标题",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "ticket__title",
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "verification.required"
                                            ],
                                            "noOperate": [
                                                "delete",
                                                "copy"
                                            ]
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "标题",
                                                "name_en": "Short description",
                                                "name_zh_hant": "標題"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": true,
                                                "enabled": true
                                            },
                                            "wordLimit": {
                                                "value": {
                                                    "max": "",
                                                    "min": 0
                                                },
                                                "enabled": false
                                            },
                                            "formatLimit": {
                                                "value": {
                                                    "errorTips": "",
                                                    "expression": ""
                                                },
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "id": "tmuSWGTE",
                                        "key": "application_content",
                                        "desc": "",
                                        "rows": 4,
                                        "tips": "请输入",
                                        "type": "textarea",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "申请内容",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "Y85Z142o",
                                        "location": "form",
                                        "sceneKey": "application_content",
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "defaultValue": "",
                                        "translations": {

                                        },
                                        "verification": {
                                            "required": {
                                                "value": true,
                                                "enabled": true
                                            },
                                            "wordLimit": {
                                                "value": {
                                                    "max": "",
                                                    "min": 0
                                                },
                                                "enabled": false
                                            },
                                            "formatLimit": {
                                                "value": "",
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "key": "account_manager",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "multiUser",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "帐号负责人",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "w6bcmVPf",
                                        "location": "form",
                                        "sceneKey": "account_manager",
                                        "userScope": {
                                            "type": "custom",
                                            "value": {

                                            },
                                            "isRecursive": false
                                        },
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "userConfig": {
                                            "setSelf": false,
                                            "multiple": true,
                                            "userInfo": [

                                            ],
                                            "showStyle": "flat",
                                            "showVipIcon": true,
                                            "isShowUserInfo": false
                                        },
                                        "defaultValue": "",
                                        "translations": {

                                        },
                                        "verification": {
                                            "required": {
                                                "value": false,
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "key": "platform_manager",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "multiUser",
                                        "class": [

                                        ],
                                        "state": "hide",
                                        "title": {
                                            "value": "平台管理员",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "jB42VMii",
                                        "location": "form",
                                        "sceneKey": "platform_manager",
                                        "userScope": {
                                            "type": "custom",
                                            "value": {

                                            },
                                            "isRecursive": false
                                        },
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "userConfig": {
                                            "setSelf": false,
                                            "multiple": true,
                                            "userInfo": [

                                            ],
                                            "showStyle": "flat",
                                            "showVipIcon": true,
                                            "isShowUserInfo": false
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "平台管理员",
                                                "name_en": "平台管理员",
                                                "name_zh_hans": "平台管理员"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": false,
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            }
                        ],
                        "styleCode": "",
                        "dataLinkage": [

                        ],
                        "verification": [

                        ]
                    },
                    "jsonschema": {
                        "type": "object",
                        "properties": {
                            "ticket__title": {
                                "translations": {
                                    "title": "标题",
                                    "title_en": "Short description",
                                    "title_zh_hans": "标题"
                                },
                                "type": "string",
                                "title": "标题",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": null
                            },
                            "account_manager": {
                                "translations": {
                                    "title": "帐号负责人",
                                    "title_en": "帐号负责人",
                                    "title_zh_hans": "帐号负责人"
                                },
                                "type": "array",
                                "title": "帐号负责人",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": {
                                    "translations": {
                                        "title": "",
                                        "title_en": "",
                                        "title_zh_hans": ""
                                    },
                                    "type": "string",
                                    "title": "",
                                    "number_unit": "",
                                    "table_relation": null,
                                    "into_todo": [

                                    ],
                                    "out_todo": [

                                    ],
                                    "attr_relation": null,
                                    "itsm_jmespath": null,
                                    "itsm_options": null,
                                    "itsm_options_type": null,
                                    "format": "user",
                                    "columns": null,
                                    "properties": null,
                                    "items": null
                                }
                            },
                            "platform_manager": {
                                "translations": {
                                    "title": "平台管理员",
                                    "title_en": "平台管理员",
                                    "title_zh_hans": "平台管理员"
                                },
                                "type": "array",
                                "title": "平台管理员",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": {
                                    "translations": {
                                        "title": "",
                                        "title_en": "",
                                        "title_zh_hans": ""
                                    },
                                    "type": "string",
                                    "title": "",
                                    "number_unit": "",
                                    "table_relation": null,
                                    "into_todo": [

                                    ],
                                    "out_todo": [

                                    ],
                                    "attr_relation": null,
                                    "itsm_jmespath": null,
                                    "itsm_options": null,
                                    "itsm_options_type": null,
                                    "format": "user",
                                    "columns": null,
                                    "properties": null,
                                    "items": null
                                }
                            },
                            "application_content": {
                                "translations": {
                                    "title": "申请内容",
                                    "title_en": "申请内容",
                                    "title_zh_hans": "申请内容"
                                },
                                "type": "string",
                                "title": "申请内容",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": null
                            }
                        },
                        "additionalProperties": false
                    },
                    "decision_table_relations": [

                    ],
                    "datasheet_table_relations": [

                    ]
                },
                "flow_canvas_data": {
                    "data": [
                        {
                            "id": "connectingobject20250616170200013101",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "eventobject20250616170200005901",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject20250616170200005901",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250616170309_1",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject20250616170200005901",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject_20250616170307_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250616170322_1",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject_20250616170307_1",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "eventobject_20250616170319_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "eventobject20250616170200005901",
                            "data": {
                                "x": 20,
                                "y": 180,
                                "id": "eventobject20250616170200005901",
                                "icon": "cw-icon cw-icon-kai-shi",
                                "meta": "$.events.eventobject20250616170200005901.meta",
                                "name": "$.events.eventobject20250616170200005901.name",
                                "type": "start",
                                "width": 80,
                                "height": 40,
                                "zIndex": 2,
                                "isError": [

                                ],
                                "isFinished": false,
                                "isSelected": false,
                                "translations": "$.events.eventobject20250616170200005901.translations"
                            },
                            "size": {
                                "width": 80,
                                "height": 40
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-right": {
                                        "position": "right"
                                    }
                                }
                            },
                            "shape": "events",
                            "zIndex": 2,
                            "position": {
                                "x": 20,
                                "y": 180
                            }
                        },
                        {
                            "id": "activityobject20250616170200005901",
                            "data": {
                                "x": 180,
                                "y": 160,
                                "id": "activityobject20250616170200005901",
                                "icon": "cw-icon cw-icon-shen-qing",
                                "meta": "$.activities.activityobject20250616170200005901.meta",
                                "name": "$.activities.activityobject20250616170200005901.name",
                                "type": "SUBMIT",
                                "width": 200,
                                "height": 80,
                                "zIndex": 2,
                                "isError": [

                                ],
                                "nodeType": "activities",
                                "isFinished": true,
                                "isSelected": false,
                                "configurable": true,
                                "translations": "$.activities.activityobject20250616170200005901.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 180,
                                "y": 160
                            }
                        },
                        {
                            "id": "activityobject_20250616170307_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "activityobject_20250616170307_1",
                                "code": "APPROVE_TASK",
                                "icon": "cw-icon cw-icon-shen-pi",
                                "meta": "$.activities.activityobject_20250616170307_1.meta",
                                "name": "$.activities.activityobject_20250616170307_1.name",
                                "type": "APPROVE_TASK",
                                "color": [
                                    "#FFE5C7",
                                    "#FD9D2C"
                                ],
                                "label": "审批节点",
                                "width": 200,
                                "config": {
                                    "type": "tab",
                                    "tabList": [
                                        {
                                            "type": "approval",
                                            "label": "审批对象",
                                            "isError": false
                                        },
                                        {
                                            "meta": {
                                                "buttons": [
                                                    {
                                                        "key": "approve",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "同意",
                                                        "label": "同意",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "refuse",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "拒绝",
                                                        "label": "拒绝",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "update",
                                                        "name": "更新",
                                                        "label": "更新",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "save",
                                                        "name": "保存",
                                                        "label": "保存",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "deliver",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ]
                                                        },
                                                        "name": "转单",
                                                        "label": "转单",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "signature",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ],
                                                            "patterns": [

                                                            ]
                                                        },
                                                        "name": "加签",
                                                        "label": "加签",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "back",
                                                        "meta": {
                                                            "pattern": "again",
                                                            "activities": [
                                                                "all"
                                                            ]
                                                        },
                                                        "name": "退回",
                                                        "label": "退回",
                                                        "switch": false,
                                                        "disabled": false
                                                    }
                                                ]
                                            },
                                            "type": "operate",
                                            "label": "操作按钮",
                                            "isError": false
                                        },
                                        {
                                            "type": "fields",
                                            "label": "字段配置",
                                            "isError": false
                                        }
                                    ]
                                },
                                "height": 80,
                                "isError": [

                                ],
                                "toolbar": [
                                    "copy",
                                    "delete"
                                ],
                                "dataType": "activities",
                                "nodeType": "activities",
                                "isDisabled": false,
                                "isFinished": true,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {
                                        "label": "approve",
                                        "fields": {

                                        },
                                        "buttons": [
                                            {
                                                "key": "approve",
                                                "name": "同意",
                                                "extra": {
                                                    "theme": "success"
                                                },
                                                "translations": {
                                                    "name": "同意",
                                                    "name_en": "Approve",
                                                    "name_zh_hant": "同意"
                                                }
                                            },
                                            {
                                                "key": "refuse",
                                                "name": "拒绝",
                                                "extra": {
                                                    "theme": "danger"
                                                },
                                                "translations": {
                                                    "name": "拒绝",
                                                    "name_en": "Reject",
                                                    "name_zh_hant": "拒絕"
                                                }
                                            }
                                        ],
                                        "processors": [

                                        ],
                                        "working_mode": "serial",
                                        "processors_type": "",
                                        "advanced_settings": {
                                            "auto_terminate": false
                                        }
                                    },
                                    "name": "审批节点",
                                    "translations": {
                                        "name": "审批节点",
                                        "name_en": "Approve Node",
                                        "name_zh_hant": "審批節點"
                                    }
                                },
                                "configurable": true,
                                "translations": "$.activities.activityobject_20250616170307_1.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 500,
                                "y": 160
                            }
                        },
                        {
                            "id": "eventobject_20250616170319_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "eventobject_20250616170319_1",
                                "code": "end",
                                "icon": "cw-icon cw-icon-jie-shu",
                                "meta": "$.events.eventobject_20250616170319_1.meta",
                                "name": "$.events.eventobject_20250616170319_1.name",
                                "type": "end",
                                "label": "结束",
                                "width": 80,
                                "height": 40,
                                "isError": [

                                ],
                                "toolbar": [
                                    "delete"
                                ],
                                "dataType": "events",
                                "nodeType": "events",
                                "isDisabled": false,
                                "isFinished": false,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {

                                    },
                                    "name": "结束",
                                    "translations": {
                                        "name": "结束",
                                        "name_en": "End",
                                        "name_zh_hant": "結束"
                                    }
                                },
                                "translations": "$.events.eventobject_20250616170319_1.translations"
                            },
                            "size": {
                                "width": 80,
                                "height": 40
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "events",
                            "zIndex": 2,
                            "position": {
                                "x": 860,
                                "y": 180
                            }
                        }
                    ]
                },
                "normal_pattern_meta": null
            }
        },
        {
            "workflow": {
                "key": "$Workflow20250616171100004102",
                "name": "HCM2.0 资源申请通用流程",
                "portal_id": "DEFAULT",
                "desc": "HCM",
                "category": "$WorkflowCategory20250605162300000201",
                "translations": {
                    "name": "HCM2.0 资源申请通用流程",
                    "name_en": "HCM2.0 资源申请通用流程",
                    "desc": "HCM",
                    "desc_en": "HCM"
                },
                "engine_pattern": "FORMAL",
                "form_model_key": "$FormModel20250616165800003101",
                "app_id": "core",
                "meta": {
                    "workflow_button": [
                        {
                            "translations": {
                                "name": "关闭",
                                "name_en": "Close",
                                "name_zh_hant": "關閉"
                            },
                            "key": "close",
                            "name": "关闭",
                            "enable": true,
                            "meta": {
                                "button_permission": [
                                    {
                                        "type": "ticket_role",
                                        "value": [
                                            "creator",
                                            "current_processors"
                                        ]
                                    }
                                ],
                                "urging_time": null
                            },
                            "extra": {

                            }
                        },
                        {
                            "translations": {
                                "name": "终止",
                                "name_en": "Cancel",
                                "name_zh_hant": "終止"
                            },
                            "key": "terminate",
                            "name": "终止",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": {

                            }
                        },
                        {
                            "translations": {
                                "name": "重新打开",
                                "name_en": "Reopen",
                                "name_zh_hant": "重新打開"
                            },
                            "key": "restart",
                            "name": "重新打开",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": {

                            }
                        },
                        {
                            "translations": {
                                "name": "催办",
                                "name_en": "Urge",
                                "name_zh_hant": "催辦"
                            },
                            "key": "urging",
                            "name": "催办",
                            "enable": false,
                            "meta": {
                                "button_permission": null,
                                "urging_time": null
                            },
                            "extra": {

                            }
                        }
                    ]
                }
            },
            "version": {
                "key": "20250616171100009001",
                "workflow_key": "$Workflow20250616171100004102",
                "desc": null,
                "workflows": {
                    "$Workflow20250616171100004102": {
                        "translations": {
                            "name": "HCM2.0 资源申请通用流程",
                            "name_en": "HCM2.0 资源申请通用流程",
                            "name_zh_hans": "HCM2.0 资源申请通用流程"
                        },
                        "key": "$Workflow20250616171100004102",
                        "name": "HCM2.0 资源申请通用流程",
                        "desc": "HCM",
                        "type": "",
                        "is_sub": false,
                        "activity_key": null,
                        "connecting_objects": {
                            "connectingobject_20250616171158_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250616171158_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject20250616171100006001",
                                "source_type": "activity",
                                "dest_key": "activityobject_20250616171151_1",
                                "dest_type": "activity"
                            },
                            "connectingobject_20250616171201_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250616171201_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject_20250616171151_1",
                                "source_type": "activity",
                                "dest_key": "activityobject_20250616171153_1",
                                "dest_type": "activity"
                            },
                            "connectingobject_20250616171203_1": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject_20250616171203_1",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "activityobject_20250616171153_1",
                                "source_type": "activity",
                                "dest_key": "eventobject_20250616171155_1",
                                "dest_type": "event"
                            },
                            "connectingobject20250616171100013202": {
                                "translations": {
                                    "name": "",
                                    "name_en": "",
                                    "name_zh_hans": ""
                                },
                                "key": "connectingobject20250616171100013202",
                                "name": "",
                                "type": "sequence_flow",
                                "source_key": "eventobject20250616171100006001",
                                "source_type": "event",
                                "dest_key": "activityobject20250616171100006001",
                                "dest_type": "activity"
                            }
                        },
                        "relations": [

                        ]
                    }
                },
                "activities": {
                    "activityobject_20250616171151_1": {
                        "translations": {
                            "name": "直接上级审批",
                            "name_en": "Approve Node",
                            "name_zh_hans": "直接上级审批"
                        },
                        "key": "activityobject_20250616171151_1",
                        "workflow_key": "$Workflow20250616171100004102",
                        "name": "直接上级审批",
                        "desc": "",
                        "type": "APPROVE_TASK",
                        "incomings": [
                            "connectingobject_20250616171158_1"
                        ],
                        "outgoings": [
                            "connectingobject_20250616171201_1"
                        ],
                        "meta": {
                            "label": "approve",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "approve",
                                    "meta": {

                                    },
                                    "name": "同意",
                                    "extra": {
                                        "theme": "success"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "同意",
                                        "name_en": "Approve",
                                        "name_zh_hant": "同意"
                                    }
                                },
                                {
                                    "key": "refuse",
                                    "meta": {

                                    },
                                    "name": "拒绝",
                                    "extra": {
                                        "theme": "danger"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "拒绝",
                                        "name_en": "Reject",
                                        "name_zh_hant": "拒絕"
                                    }
                                }
                            ],
                            "processors": "get_variable(ticket_id, \"TICKET\", \"creator\", \"user_leader($tag)\" ,\"List[User]\")",
                            "working_mode": "cooperate",
                            "processors_type": "feel",
                            "advanced_settings": {
                                "auto_terminate": false
                            }
                        },
                        "hooks": [

                        ]
                    },
                    "activityobject_20250616171153_1": {
                        "translations": {
                            "name": "平台管理员审批",
                            "name_en": "Approve Node",
                            "name_zh_hans": "平台管理员审批"
                        },
                        "key": "activityobject_20250616171153_1",
                        "workflow_key": "$Workflow20250616171100004102",
                        "name": "平台管理员审批",
                        "desc": "",
                        "type": "APPROVE_TASK",
                        "incomings": [
                            "connectingobject_20250616171201_1"
                        ],
                        "outgoings": [
                            "connectingobject_20250616171203_1"
                        ],
                        "meta": {
                            "label": "approve",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "approve",
                                    "meta": {

                                    },
                                    "name": "同意",
                                    "extra": {
                                        "theme": "success"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "同意",
                                        "name_en": "Approve",
                                        "name_zh_hant": "同意"
                                    }
                                },
                                {
                                    "key": "refuse",
                                    "meta": {

                                    },
                                    "name": "拒绝",
                                    "extra": {
                                        "theme": "danger"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "拒绝",
                                        "name_en": "Reject",
                                        "name_zh_hant": "拒絕"
                                    }
                                }
                            ],
                            "processors": [
                                {
                                    "id": "DATA_TABLE[platform_manager]",
                                    "key": "DATA_TABLE",
                                    "feel": "",
                                    "name": "平台管理员",
                                    "path": "platform_manager",
                                    "type": "List[User]",
                                    "default": null,
                                    "variables": null,
                                    "jsonschema": null,
                                    "translations": {
                                        "name": "平台管理员",
                                        "name_en": "平台管理员",
                                        "name_zh_hans": "平台管理员"
                                    }
                                }
                            ],
                            "working_mode": "cooperate",
                            "processors_type": "user",
                            "advanced_settings": {
                                "auto_terminate": false
                            }
                        },
                        "hooks": [

                        ]
                    },
                    "activityobject20250616171100006001": {
                        "translations": {
                            "name": "提单",
                            "name_en": "Submit",
                            "name_zh_hans": "提单"
                        },
                        "key": "activityobject20250616171100006001",
                        "workflow_key": "$Workflow20250616171100004102",
                        "name": "提单",
                        "desc": "",
                        "type": "SUBMIT",
                        "incomings": [
                            "connectingobject20250616171100013202"
                        ],
                        "outgoings": [
                            "connectingobject_20250616171158_1"
                        ],
                        "meta": {
                            "label": "submit",
                            "fields": {

                            },
                            "buttons": [
                                {
                                    "key": "submit",
                                    "name": "提交",
                                    "extra": {
                                        "theme": "primary"
                                    },
                                    "enable": true,
                                    "translations": {
                                        "name": "提交",
                                        "name_en": "Submit",
                                        "name_zh_hant": "提交"
                                    }
                                },
                                {
                                    "key": "save_draft",
                                    "name": "保存草稿",
                                    "extra": null,
                                    "enable": true,
                                    "translations": {
                                        "name": "保存草稿",
                                        "name_en": "Save Draft",
                                        "name_zh_hant": "保存草稿"
                                    }
                                },
                                {
                                    "key": "save_template",
                                    "name": "保存模板",
                                    "extra": null,
                                    "enable": true,
                                    "translations": {
                                        "name": "保存模板",
                                        "name_en": "Save Template",
                                        "name_zh_hant": "保存模板"
                                    }
                                }
                            ]
                        },
                        "hooks": [

                        ]
                    }
                },
                "events": {
                    "eventobject_20250616171155_1": {
                        "translations": {
                            "name": "结束",
                            "name_en": "End",
                            "name_zh_hant": "結束"
                        },
                        "key": "eventobject_20250616171155_1",
                        "workflow_key": "$Workflow20250616171100004102",
                        "name": "结束",
                        "desc": "",
                        "type": "end",
                        "incomings": [
                            "connectingobject_20250616171203_1"
                        ],
                        "outgoings": [

                        ],
                        "meta": {

                        }
                    },
                    "eventobject20250616171100006001": {
                        "translations": {
                            "name": "开始",
                            "name_en": "Start",
                            "name_zh_hans": "开始"
                        },
                        "key": "eventobject20250616171100006001",
                        "workflow_key": "$Workflow20250616171100004102",
                        "name": "开始",
                        "desc": "",
                        "type": "start",
                        "incomings": [

                        ],
                        "outgoings": [
                            "connectingobject20250616171100013202"
                        ],
                        "meta": {

                        }
                    }
                },
                "gateways": {

                },
                "meta": {
                    "ticket_button": {
                        "action_button": [
                            {
                                "translations": {
                                    "name": "撤回",
                                    "name_en": "Withdraw",
                                    "name_zh_hant": "撤回"
                                },
                                "key": "withdraw",
                                "name": "撤回",
                                "enable": true,
                                "extra": {

                                },
                                "meta": {
                                    "can_withdraw_activity": [
                                        "activityobject_20250616171151_1",
                                        "activityobject_20250616171153_1"
                                    ]
                                }
                            },
                            {
                                "translations": {
                                    "name": "挂起",
                                    "name_en": "Suspend",
                                    "name_zh_hant": "掛起"
                                },
                                "key": "suspend",
                                "name": "挂起",
                                "enable": true,
                                "extra": {

                                }
                            },
                            {
                                "translations": {
                                    "name": "恢复",
                                    "name_en": "Restore",
                                    "name_zh_hant": "恢復"
                                },
                                "key": "recovery",
                                "name": "恢复",
                                "enable": true,
                                "extra": {

                                }
                            },
                            {
                                "translations": {
                                    "name": "转建文章",
                                    "name_en": "Create article",
                                    "name_zh_hant": "轉建文章"
                                },
                                "key": "convert",
                                "name": "转建文章",
                                "enable": false,
                                "extra": {

                                },
                                "meta": {
                                    "can_convert_activity": [

                                    ]
                                }
                            }
                        ]
                    },
                    "custom_button": [

                    ],
                    "vips": [

                    ],
                    "stage": {
                        "is_enable": false,
                        "model": "",
                        "config": {

                        }
                    },
                    "variables": [

                    ]
                },
                "form_canvas_data": {
                    "form_data": {
                        "id": "form_YsdvrRtqfe",
                        "type": "form",
                        "align": "top",
                        "class": [

                        ],
                        "rules": [

                        ],
                        "layout": [
                            {
                                "list": [
                                    {
                                        "key": "ticket__title",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "text",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "标题",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "ticket__title",
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "verification.required"
                                            ],
                                            "noOperate": [
                                                "delete",
                                                "copy"
                                            ]
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "标题",
                                                "name_en": "Short description",
                                                "name_zh_hant": "標題"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": true,
                                                "enabled": true
                                            },
                                            "wordLimit": {
                                                "value": {
                                                    "max": "",
                                                    "min": 0
                                                },
                                                "enabled": false
                                            },
                                            "formatLimit": {
                                                "value": {
                                                    "errorTips": "",
                                                    "expression": ""
                                                },
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "id": "tmuSWGTE",
                                        "key": "application_content",
                                        "desc": "",
                                        "rows": 4,
                                        "tips": "请输入",
                                        "type": "textarea",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "申请内容",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "R8MR3M3e",
                                        "location": "form",
                                        "sceneKey": "application_content",
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "defaultValue": "",
                                        "translations": {

                                        },
                                        "verification": {
                                            "required": {
                                                "value": true,
                                                "enabled": true
                                            },
                                            "wordLimit": {
                                                "value": {
                                                    "max": "",
                                                    "min": 0
                                                },
                                                "enabled": false
                                            },
                                            "formatLimit": {
                                                "value": "",
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "key": "account_manager",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "multiUser",
                                        "class": [

                                        ],
                                        "state": "default",
                                        "title": {
                                            "value": "帐号负责人",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "WJE6DjS2",
                                        "location": "form",
                                        "sceneKey": "account_manager",
                                        "userScope": {
                                            "type": "custom",
                                            "value": {

                                            },
                                            "isRecursive": false
                                        },
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "userConfig": {
                                            "setSelf": false,
                                            "multiple": true,
                                            "userInfo": [

                                            ],
                                            "showStyle": "flat",
                                            "showVipIcon": true,
                                            "isShowUserInfo": false
                                        },
                                        "defaultValue": "",
                                        "translations": {

                                        },
                                        "verification": {
                                            "required": {
                                                "value": false,
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            },
                            {
                                "list": [
                                    {
                                        "key": "platform_manager",
                                        "desc": "",
                                        "tips": "请输入",
                                        "type": "multiUser",
                                        "class": [

                                        ],
                                        "state": "hide",
                                        "title": {
                                            "value": "平台管理员",
                                            "isHide": false
                                        },
                                        "width": "COL_6",
                                        "columnId": "pDDKPCVw",
                                        "location": "form",
                                        "sceneKey": "platform_manager",
                                        "userScope": {
                                            "type": "custom",
                                            "value": {

                                            },
                                            "isRecursive": false
                                        },
                                        "permission": {
                                            "readonly": [
                                                "title",
                                                "key",
                                                "tips",
                                                "desc"
                                            ]
                                        },
                                        "userConfig": {
                                            "setSelf": false,
                                            "multiple": true,
                                            "userInfo": [

                                            ],
                                            "showStyle": "flat",
                                            "showVipIcon": true,
                                            "isShowUserInfo": false
                                        },
                                        "defaultValue": "",
                                        "translations": {
                                            "title": {
                                                "name": "平台管理员",
                                                "name_en": "平台管理员",
                                                "name_zh_hans": "平台管理员"
                                            }
                                        },
                                        "verification": {
                                            "required": {
                                                "value": false,
                                                "enabled": false
                                            }
                                        }
                                    }
                                ],
                                "type": "row"
                            }
                        ],
                        "styleCode": "",
                        "dataLinkage": [

                        ],
                        "verification": [

                        ]
                    },
                    "jsonschema": {
                        "type": "object",
                        "properties": {
                            "ticket__title": {
                                "translations": {
                                    "title": "标题",
                                    "title_en": "Short description",
                                    "title_zh_hans": "标题"
                                },
                                "type": "string",
                                "title": "标题",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": null
                            },
                            "account_manager": {
                                "translations": {
                                    "title": "帐号负责人",
                                    "title_en": "帐号负责人",
                                    "title_zh_hans": "帐号负责人"
                                },
                                "type": "array",
                                "title": "帐号负责人",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": {
                                    "translations": {
                                        "title": "",
                                        "title_en": "",
                                        "title_zh_hans": ""
                                    },
                                    "type": "string",
                                    "title": "",
                                    "number_unit": "",
                                    "table_relation": null,
                                    "into_todo": [

                                    ],
                                    "out_todo": [

                                    ],
                                    "attr_relation": null,
                                    "itsm_jmespath": null,
                                    "itsm_options": null,
                                    "itsm_options_type": null,
                                    "format": "user",
                                    "columns": null,
                                    "properties": null,
                                    "items": null
                                }
                            },
                            "platform_manager": {
                                "translations": {
                                    "title": "平台管理员",
                                    "title_en": "平台管理员",
                                    "title_zh_hans": "平台管理员"
                                },
                                "type": "array",
                                "title": "平台管理员",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": {
                                    "translations": {
                                        "title": "",
                                        "title_en": "",
                                        "title_zh_hans": ""
                                    },
                                    "type": "string",
                                    "title": "",
                                    "number_unit": "",
                                    "table_relation": null,
                                    "into_todo": [

                                    ],
                                    "out_todo": [

                                    ],
                                    "attr_relation": null,
                                    "itsm_jmespath": null,
                                    "itsm_options": null,
                                    "itsm_options_type": null,
                                    "format": "user",
                                    "columns": null,
                                    "properties": null,
                                    "items": null
                                }
                            },
                            "application_content": {
                                "translations": {
                                    "title": "申请内容",
                                    "title_en": "申请内容",
                                    "title_zh_hans": "申请内容"
                                },
                                "type": "string",
                                "title": "申请内容",
                                "number_unit": "",
                                "table_relation": null,
                                "into_todo": [

                                ],
                                "out_todo": [

                                ],
                                "attr_relation": null,
                                "itsm_jmespath": null,
                                "itsm_options": null,
                                "itsm_options_type": null,
                                "format": null,
                                "columns": null,
                                "properties": null,
                                "items": null
                            }
                        },
                        "additionalProperties": false
                    },
                    "decision_table_relations": [

                    ],
                    "datasheet_table_relations": [

                    ]
                },
                "flow_canvas_data": {
                    "data": [
                        {
                            "id": "connectingobject20250616171100013202",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "eventobject20250616171100006001",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject20250616171100006001",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250616171158_1",
                            "attrs": {
                                "line": {
                                    "stroke": "#1272FF"
                                }
                            },
                            "shape": "sequence_flow",
                            "tools": {
                                "name": null,
                                "items": [
                                    {
                                        "args": {
                                            "markup": [
                                                {
                                                    "attrs": {
                                                        "width": 24,
                                                        "height": 24
                                                    },
                                                    "tagName": "foreignObject",
                                                    "children": [
                                                        {
                                                            "ns": "http://www.w3.org/1999/xhtml",
                                                            "style": {
                                                                "width": "100%",
                                                                "height": "100%",
                                                                "min-width": 0
                                                            },
                                                            "tagName": "body",
                                                            "children": [
                                                                {
                                                                    "style": {
                                                                        "color": "#fff",
                                                                        "width": "100%",
                                                                        "height": "100%",
                                                                        "display": "flex",
                                                                        "background": "#1272FF",
                                                                        "align-items": "center",
                                                                        "line-height": 1,
                                                                        "border-radius": "4px",
                                                                        "justify-content": "center"
                                                                    },
                                                                    "tagName": "div",
                                                                    "className": "cw-icon cw-icon-delete"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "style": {
                                                        "display": "none"
                                                    },
                                                    "tagName": "p"
                                                }
                                            ],
                                            "offset": {
                                                "x": 0,
                                                "y": -12
                                            },
                                            "rotate": true,
                                            "distance": "50%"
                                        },
                                        "name": "button-remove"
                                    },
                                    {
                                        "args": {
                                            "attrs": {
                                                "fill": "#1272FF"
                                            }
                                        },
                                        "name": "target-arrowhead"
                                    },
                                    {
                                        "args": {
                                            "attrs": {
                                                "fill": "#1272FF"
                                            }
                                        },
                                        "name": "vertices"
                                    },
                                    {
                                        "args": {
                                            "markup": [
                                                {
                                                    "attrs": {
                                                        "width": 24,
                                                        "height": 24
                                                    },
                                                    "tagName": "foreignObject",
                                                    "children": [
                                                        {
                                                            "ns": "http://www.w3.org/1999/xhtml",
                                                            "style": {
                                                                "width": "100%",
                                                                "height": "100%",
                                                                "min-width": 0
                                                            },
                                                            "tagName": "body",
                                                            "children": [
                                                                {
                                                                    "style": {
                                                                        "color": "#fff",
                                                                        "width": "100%",
                                                                        "cursor": "pointer",
                                                                        "height": "100%",
                                                                        "display": "flex",
                                                                        "background": "#1272FF",
                                                                        "align-items": "center",
                                                                        "line-height": 1,
                                                                        "border-radius": "4px",
                                                                        "justify-content": "center"
                                                                    },
                                                                    "tagName": "div",
                                                                    "className": "cw-icon cw-icon-edit"
                                                                }
                                                            ]
                                                        }
                                                    ]
                                                },
                                                {
                                                    "style": {
                                                        "display": "none"
                                                    },
                                                    "tagName": "p"
                                                }
                                            ],
                                            "offset": {
                                                "x": -26,
                                                "y": -12
                                            },
                                            "rotate": true,
                                            "distance": "50%"
                                        },
                                        "name": "button"
                                    }
                                ]
                            },
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject20250616171100006001",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject_20250616171151_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250616171201_1",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject_20250616171151_1",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "activityobject_20250616171153_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "connectingobject_20250616171203_1",
                            "shape": "sequence_flow",
                            "router": {
                                "args": {
                                    "step": 20
                                },
                                "name": "manhattan"
                            },
                            "source": {
                                "cell": "activityobject_20250616171153_1",
                                "port": "p-right"
                            },
                            "target": {
                                "cell": "eventobject_20250616171155_1",
                                "port": "p-left"
                            },
                            "zIndex": 0,
                            "connector": {
                                "args": {
                                    "radius": 8
                                },
                                "name": "rounded"
                            }
                        },
                        {
                            "id": "eventobject20250616171100006001",
                            "data": {
                                "x": 20,
                                "y": 180,
                                "id": "eventobject20250616171100006001",
                                "icon": "cw-icon cw-icon-kai-shi",
                                "meta": "$.events.eventobject20250616171100006001.meta",
                                "name": "$.events.eventobject20250616171100006001.name",
                                "type": "start",
                                "width": 80,
                                "height": 40,
                                "zIndex": 2,
                                "isError": [

                                ],
                                "isFinished": false,
                                "isSelected": false,
                                "translations": "$.events.eventobject20250616171100006001.translations"
                            },
                            "size": {
                                "width": 80,
                                "height": 40
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-right": {
                                        "position": "right"
                                    }
                                }
                            },
                            "shape": "events",
                            "zIndex": 2,
                            "position": {
                                "x": 20,
                                "y": 180
                            }
                        },
                        {
                            "id": "activityobject20250616171100006001",
                            "data": {
                                "x": 180,
                                "y": 160,
                                "id": "activityobject20250616171100006001",
                                "icon": "cw-icon cw-icon-shen-qing",
                                "meta": "$.activities.activityobject20250616171100006001.meta",
                                "name": "$.activities.activityobject20250616171100006001.name",
                                "type": "SUBMIT",
                                "width": 200,
                                "height": 80,
                                "zIndex": 2,
                                "isError": [

                                ],
                                "nodeType": "activities",
                                "isFinished": true,
                                "isSelected": false,
                                "configurable": true,
                                "translations": "$.activities.activityobject20250616171100006001.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 180,
                                "y": 160
                            }
                        },
                        {
                            "id": "activityobject_20250616171151_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "activityobject_20250616171151_1",
                                "code": "APPROVE_TASK",
                                "icon": "cw-icon cw-icon-shen-pi",
                                "meta": "$.activities.activityobject_20250616171151_1.meta",
                                "name": "$.activities.activityobject_20250616171151_1.name",
                                "type": "APPROVE_TASK",
                                "color": [
                                    "#FFE5C7",
                                    "#FD9D2C"
                                ],
                                "label": "审批节点",
                                "width": 200,
                                "config": {
                                    "type": "tab",
                                    "tabList": [
                                        {
                                            "type": "approval",
                                            "label": "审批对象",
                                            "isError": false
                                        },
                                        {
                                            "meta": {
                                                "buttons": [
                                                    {
                                                        "key": "approve",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "同意",
                                                        "label": "同意",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "refuse",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "拒绝",
                                                        "label": "拒绝",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "update",
                                                        "name": "更新",
                                                        "label": "更新",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "save",
                                                        "name": "保存",
                                                        "label": "保存",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "deliver",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ]
                                                        },
                                                        "name": "转单",
                                                        "label": "转单",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "signature",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ],
                                                            "patterns": [

                                                            ]
                                                        },
                                                        "name": "加签",
                                                        "label": "加签",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "back",
                                                        "meta": {
                                                            "pattern": "again",
                                                            "activities": [
                                                                "all"
                                                            ]
                                                        },
                                                        "name": "退回",
                                                        "label": "退回",
                                                        "switch": false,
                                                        "disabled": false
                                                    }
                                                ]
                                            },
                                            "type": "operate",
                                            "label": "操作按钮",
                                            "isError": false
                                        },
                                        {
                                            "type": "fields",
                                            "label": "字段配置",
                                            "isError": false
                                        }
                                    ]
                                },
                                "height": 80,
                                "isError": [

                                ],
                                "toolbar": [
                                    "copy",
                                    "delete"
                                ],
                                "dataType": "activities",
                                "nodeType": "activities",
                                "isDisabled": false,
                                "isFinished": true,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {
                                        "label": "approve",
                                        "fields": {

                                        },
                                        "buttons": [
                                            {
                                                "key": "approve",
                                                "name": "同意",
                                                "extra": {
                                                    "theme": "success"
                                                },
                                                "translations": {
                                                    "name": "同意",
                                                    "name_en": "Approve",
                                                    "name_zh_hant": "同意"
                                                }
                                            },
                                            {
                                                "key": "refuse",
                                                "name": "拒绝",
                                                "extra": {
                                                    "theme": "danger"
                                                },
                                                "translations": {
                                                    "name": "拒绝",
                                                    "name_en": "Reject",
                                                    "name_zh_hant": "拒絕"
                                                }
                                            }
                                        ],
                                        "processors": [

                                        ],
                                        "working_mode": "serial",
                                        "processors_type": "",
                                        "advanced_settings": {
                                            "auto_terminate": false
                                        }
                                    },
                                    "name": "审批节点",
                                    "translations": {
                                        "name": "审批节点",
                                        "name_en": "Approve Node",
                                        "name_zh_hant": "審批節點"
                                    }
                                },
                                "configurable": true,
                                "translations": "$.activities.activityobject_20250616171151_1.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 480,
                                "y": 160
                            }
                        },
                        {
                            "id": "activityobject_20250616171153_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "activityobject_20250616171153_1",
                                "code": "APPROVE_TASK",
                                "icon": "cw-icon cw-icon-shen-pi",
                                "meta": "$.activities.activityobject_20250616171153_1.meta",
                                "name": "$.activities.activityobject_20250616171153_1.name",
                                "type": "APPROVE_TASK",
                                "color": [
                                    "#FFE5C7",
                                    "#FD9D2C"
                                ],
                                "label": "审批节点",
                                "width": 200,
                                "config": {
                                    "type": "tab",
                                    "tabList": [
                                        {
                                            "type": "approval",
                                            "label": "审批对象",
                                            "isError": false
                                        },
                                        {
                                            "meta": {
                                                "buttons": [
                                                    {
                                                        "key": "approve",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "同意",
                                                        "label": "同意",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "refuse",
                                                        "meta": {
                                                            "placeholder": "请输入"
                                                        },
                                                        "name": "拒绝",
                                                        "label": "拒绝",
                                                        "switch": true,
                                                        "disabled": true
                                                    },
                                                    {
                                                        "key": "update",
                                                        "name": "更新",
                                                        "label": "更新",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "save",
                                                        "name": "保存",
                                                        "label": "保存",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "deliver",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ]
                                                        },
                                                        "name": "转单",
                                                        "label": "转单",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "signature",
                                                        "meta": {
                                                            "ranges": [
                                                                {
                                                                    "processors": [

                                                                    ],
                                                                    "processors_type": ""
                                                                }
                                                            ],
                                                            "patterns": [

                                                            ]
                                                        },
                                                        "name": "加签",
                                                        "label": "加签",
                                                        "switch": false,
                                                        "disabled": false
                                                    },
                                                    {
                                                        "key": "back",
                                                        "meta": {
                                                            "pattern": "again",
                                                            "activities": [
                                                                "all"
                                                            ]
                                                        },
                                                        "name": "退回",
                                                        "label": "退回",
                                                        "switch": false,
                                                        "disabled": false
                                                    }
                                                ]
                                            },
                                            "type": "operate",
                                            "label": "操作按钮",
                                            "isError": false
                                        },
                                        {
                                            "type": "fields",
                                            "label": "字段配置",
                                            "isError": false
                                        }
                                    ]
                                },
                                "height": 80,
                                "isError": [

                                ],
                                "toolbar": [
                                    "copy",
                                    "delete"
                                ],
                                "dataType": "activities",
                                "nodeType": "activities",
                                "isDisabled": false,
                                "isFinished": true,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {
                                        "label": "approve",
                                        "fields": {

                                        },
                                        "buttons": [
                                            {
                                                "key": "approve",
                                                "name": "同意",
                                                "extra": {
                                                    "theme": "success"
                                                },
                                                "translations": {
                                                    "name": "同意",
                                                    "name_en": "Approve",
                                                    "name_zh_hant": "同意"
                                                }
                                            },
                                            {
                                                "key": "refuse",
                                                "name": "拒绝",
                                                "extra": {
                                                    "theme": "danger"
                                                },
                                                "translations": {
                                                    "name": "拒绝",
                                                    "name_en": "Reject",
                                                    "name_zh_hant": "拒絕"
                                                }
                                            }
                                        ],
                                        "processors": [

                                        ],
                                        "working_mode": "serial",
                                        "processors_type": "",
                                        "advanced_settings": {
                                            "auto_terminate": false
                                        }
                                    },
                                    "name": "审批节点",
                                    "translations": {
                                        "name": "审批节点",
                                        "name_en": "Approve Node",
                                        "name_zh_hant": "審批節點"
                                    }
                                },
                                "configurable": true,
                                "translations": "$.activities.activityobject_20250616171153_1.translations"
                            },
                            "size": {
                                "width": 200,
                                "height": 80
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "activities",
                            "zIndex": 2,
                            "position": {
                                "x": 780,
                                "y": 160
                            }
                        },
                        {
                            "id": "eventobject_20250616171155_1",
                            "data": {
                                "x": 130,
                                "y": 30,
                                "id": "eventobject_20250616171155_1",
                                "code": "end",
                                "icon": "cw-icon cw-icon-jie-shu",
                                "meta": "$.events.eventobject_20250616171155_1.meta",
                                "name": "$.events.eventobject_20250616171155_1.name",
                                "type": "end",
                                "label": "结束",
                                "width": 80,
                                "height": 40,
                                "isError": [

                                ],
                                "toolbar": [
                                    "delete"
                                ],
                                "dataType": "events",
                                "nodeType": "events",
                                "isDisabled": false,
                                "isFinished": false,
                                "isSelected": false,
                                "defaultData": {
                                    "meta": {

                                    },
                                    "name": "结束",
                                    "translations": {
                                        "name": "结束",
                                        "name_en": "End",
                                        "name_zh_hant": "結束"
                                    }
                                },
                                "translations": "$.events.eventobject_20250616171155_1.translations"
                            },
                            "size": {
                                "width": 80,
                                "height": 40
                            },
                            "view": "vue-shape-view",
                            "ports": {
                                "items": [
                                    {
                                        "id": "p-top",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-top",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-right",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-right",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-bottom",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-bottom",
                                        "zIndex": 20
                                    },
                                    {
                                        "id": "p-left",
                                        "attrs": {
                                            "circle": {
                                                "r": 4,
                                                "fill": "#fff",
                                                "style": {
                                                    "visibility": "hidden"
                                                },
                                                "magnet": true,
                                                "stroke": "#1272FF",
                                                "strokeWidth": 2
                                            }
                                        },
                                        "group": "port-left",
                                        "zIndex": 20
                                    }
                                ],
                                "groups": {
                                    "port-top": {
                                        "position": "top"
                                    },
                                    "port-left": {
                                        "position": "left"
                                    },
                                    "port-right": {
                                        "position": "right"
                                    },
                                    "port-bottom": {
                                        "position": "bottom"
                                    }
                                }
                            },
                            "shape": "events",
                            "zIndex": 2,
                            "position": {
                                "x": 1100,
                                "y": 180
                            }
                        }
                    ]
                },
                "normal_pattern_meta": null
            }
        }
    ],
    "key_mapping": {
        "form_models": {
            "{{ .tenantID }}_hcm_main_account": "$FormModel20250605162400001301",
            "{{ .tenantID }}_hcm_common": "$FormModel20250616165800003101"
        },
        "workflow_categories": {
            "{{ .tenantID }}_hcm": "$WorkflowCategory20250605162300000201"
        },
        "workflows": {
            "{{ .tenantID }}_hcm_main_account": "$Workflow20250605162500002001",
            "{{ .tenantID }}_hcm_add_account": "$Workflow20250616170200004001",
            "{{ .tenantID }}_hcm_common": "$Workflow20250616171100004102"
        }
    }
}`
)
