{
	"is_deleted": false,
	"name": "\u521b\u5efa\u5171\u4eab\u96c6\u7fa4\u547d\u540d\u7a7a\u95f4-20240308112119-20240914173410-20240914175319",
	"desc": "",
	"flow_type": "other",
	"is_enabled": true,
	"is_revocable": true,
	"revoke_config": {
		"state": 0,
		"type": 1
	},
	"is_draft": false,
	"is_builtin": false,
	"is_task_needed": false,
	"owners": "",
	"notify_rule": "ONCE",
	"notify_freq": 0,
	"is_biz_needed": false,
	"is_auto_approve": false,
	"is_iam_used": false,
	"is_supervise_needed": true,
	"supervise_type": "EMPTY",
	"supervisor": "",
	"engine_version": "PIPELINE_V1",
	"version_number": "20240920152131",
	"table": {
		"id": 36,
		"is_deleted": false,
		"name": "\u9ed8\u8ba4_20240319192628",
		"desc": "\u9ed8\u8ba4\u57fa\u7840\u6a21\u578b",
		"version": "EMPTY",
		"fields": [{
				"id": 1,
				"is_deleted": false,
				"is_builtin": true,
				"is_readonly": false,
				"is_valid": true,
				"display": true,
				"source_type": "CUSTOM",
				"source_uri": "",
				"api_instance_id": 0,
				"kv_relation": {},
				"type": "STRING",
				"key": "title",
				"name": "\u6807\u9898",
				"layout": "COL_12",
				"validate_type": "REQUIRE",
				"show_type": 1,
				"show_conditions": {},
				"regex": "EMPTY",
				"regex_config": {},
				"custom_regex": "",
				"desc": "\u8bf7\u8f93\u5165\u6807\u9898",
				"tips": "",
				"is_tips": false,
				"default": "",
				"choice": [],
				"related_fields": {},
				"meta": {},
				"flow_type": "DEFAULT",
				"project_key": "public",
				"source": "BASE-MODEL"
			},
			{
				"id": 2,
				"is_deleted": false,
				"is_builtin": true,
				"is_readonly": false,
				"is_valid": true,
				"display": true,
				"source_type": "DATADICT",
				"source_uri": "IMPACT",
				"api_instance_id": 0,
				"kv_relation": {},
				"type": "SELECT",
				"key": "impact",
				"name": "\u5f71\u54cd\u8303\u56f4",
				"layout": "COL_12",
				"validate_type": "REQUIRE",
				"show_type": 1,
				"show_conditions": {},
				"regex": "EMPTY",
				"regex_config": {},
				"custom_regex": "",
				"desc": "\u8bf7\u9009\u62e9\u5f71\u54cd\u8303\u56f4",
				"tips": "",
				"is_tips": false,
				"default": "",
				"choice": [],
				"related_fields": {},
				"meta": {},
				"flow_type": "DEFAULT",
				"project_key": "public",
				"source": "BASE-MODEL"
			},
			{
				"id": 3,
				"is_deleted": false,
				"is_builtin": true,
				"is_readonly": false,
				"is_valid": true,
				"display": true,
				"source_type": "DATADICT",
				"source_uri": "URGENCY",
				"api_instance_id": 0,
				"kv_relation": {},
				"type": "SELECT",
				"key": "urgency",
				"name": "\u7d27\u6025\u7a0b\u5ea6",
				"layout": "COL_12",
				"validate_type": "REQUIRE",
				"show_type": 1,
				"show_conditions": {},
				"regex": "EMPTY",
				"regex_config": {},
				"custom_regex": "",
				"desc": "\u8bf7\u9009\u62e9\u7d27\u6025\u7a0b\u5ea6",
				"tips": "",
				"is_tips": false,
				"default": "",
				"choice": [],
				"related_fields": {},
				"meta": {},
				"flow_type": "DEFAULT",
				"project_key": "public",
				"source": "BASE-MODEL"
			},
			{
				"id": 4,
				"is_deleted": false,
				"is_builtin": true,
				"is_readonly": true,
				"is_valid": true,
				"display": true,
				"source_type": "DATADICT",
				"source_uri": "PRIORITY",
				"api_instance_id": 0,
				"kv_relation": {},
				"type": "SELECT",
				"key": "priority",
				"name": "\u4f18\u5148\u7ea7",
				"layout": "COL_12",
				"validate_type": "REQUIRE",
				"show_type": 1,
				"show_conditions": {},
				"regex": "EMPTY",
				"regex_config": {},
				"custom_regex": "",
				"desc": "\u8bf7\u9009\u62e9\u4f18\u5148\u7ea7",
				"tips": "",
				"is_tips": false,
				"default": "",
				"choice": [],
				"related_fields": {
					"rely_on": [
						"urgency",
						"impact"
					]
				},
				"meta": {},
				"flow_type": "DEFAULT",
				"project_key": "public",
				"source": "BASE-MODEL"
			},
			{
				"id": 5,
				"is_deleted": false,
				"is_builtin": true,
				"is_readonly": false,
				"is_valid": true,
				"display": true,
				"source_type": "RPC",
				"source_uri": "ticket_status",
				"api_instance_id": 0,
				"kv_relation": {},
				"type": "SELECT",
				"key": "current_status",
				"name": "\u5de5\u5355\u72b6\u6001",
				"layout": "COL_12",
				"validate_type": "REQUIRE",
				"show_type": 1,
				"show_conditions": {},
				"regex": "EMPTY",
				"regex_config": {},
				"custom_regex": "",
				"desc": "\u8bf7\u9009\u62e9\u5de5\u5355\u72b6\u6001",
				"tips": "",
				"is_tips": false,
				"default": "",
				"choice": [],
				"related_fields": {},
				"meta": {},
				"flow_type": "DEFAULT",
				"project_key": "public",
				"source": "BASE-MODEL"
			}
		],
		"fields_order": [
			1,
			2,
			3,
			4,
			5
		],
		"field_key_order": [
			"title",
			"impact",
			"urgency",
			"priority",
			"current_status"
		]
	},
	"task_schemas": [],
	"creator": "",
	"updated_by": "",
	"workflow_id": 64,
	"version_message": "",
	"states": {
		"356": {
			"workflow": 64,
			"id": 356,
			"key": 356,
			"name": "\u5f00\u59cb",
			"desc": "",
			"distribute_type": "PROCESS",
			"axis": {
				"x": 150,
				"y": 150
			},
			"is_builtin": true,
			"variables": {
				"inputs": [],
				"outputs": []
			},
			"tag": "DEFAULT",
			"processors_type": "OPEN",
			"processors": "",
			"assignors": "",
			"assignors_type": "EMPTY",
			"delivers": "",
			"delivers_type": "EMPTY",
			"can_deliver": false,
			"extras": {},
			"is_draft": false,
			"is_terminable": false,
			"fields": [],
			"type": "START",
			"api_instance_id": 0,
			"is_sequential": false,
			"finish_condition": {},
			"is_multi": false,
			"is_allow_skip": false,
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null,
			"is_first_state": false
		},
		"357": {
			"workflow": 64,
			"id": 357,
			"key": 357,
			"name": "\u63d0\u5355",
			"desc": "",
			"distribute_type": "PROCESS",
			"axis": {
				"x": 285,
				"y": 150
			},
			"is_builtin": true,
			"variables": {
				"inputs": [],
				"outputs": [{
						"key": "CLUSTER_TYPE",
						"source": "field",
						"state": 2948,
						"type": "SELECT"
					},
					{
						"key": "CLUSTER_ID",
						"source": "field",
						"state": 2582,
						"type": "STRING"
					},
					{
						"key": "CPU_LIMITS",
						"source": "field",
						"state": 2718,
						"type": "INT"
					},
					{
						"key": "MEMORY_LIMITS",
						"source": "field",
						"state": 2718,
						"type": "INT"
					}
				]
			},
			"tag": "DEFAULT",
			"processors_type": "OPEN",
			"processors": "",
			"assignors": "",
			"assignors_type": "EMPTY",
			"delivers": "",
			"delivers_type": "EMPTY",
			"can_deliver": false,
			"extras": {
				"ticket_status": {
					"name": "",
					"type": "keep"
				}
			},
			"is_draft": false,
			"is_terminable": false,
			"fields": [
				607,
				612,
				611,
				613,
				614,
				615
			],
			"type": "NORMAL",
			"api_instance_id": 0,
			"is_sequential": false,
			"finish_condition": {},
			"is_multi": false,
			"is_allow_skip": false,
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": "admin",
			"update_at": "2024-09-20 15:03:16",
			"end_at": null,
			"is_first_state": true
		},
		"358": {
			"workflow": 64,
			"id": 358,
			"key": 358,
			"name": "\u7ed3\u675f",
			"desc": "",
			"distribute_type": "PROCESS",
			"axis": {
				"x": 1195,
				"y": 150
			},
			"is_builtin": true,
			"variables": {
				"inputs": [],
				"outputs": []
			},
			"tag": "DEFAULT",
			"processors_type": "OPEN",
			"processors": "",
			"assignors": "",
			"assignors_type": "EMPTY",
			"delivers": "",
			"delivers_type": "EMPTY",
			"can_deliver": false,
			"extras": {},
			"is_draft": false,
			"is_terminable": false,
			"fields": [],
			"type": "END",
			"api_instance_id": 0,
			"is_sequential": false,
			"finish_condition": {},
			"is_multi": false,
			"is_allow_skip": false,
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null,
			"is_first_state": false
		},
		"359": {
			"workflow": 64,
			"id": 359,
			"key": 359,
			"name": "\u8d1f\u8d23\u4eba\u5ba1\u6279",
			"desc": "",
			"distribute_type": "PROCESS",
			"axis": {
				"x": 540,
				"y": 155
			},
			"is_builtin": false,
			"variables": {
				"inputs": [],
				"outputs": [{
						"key": "Fd6380d03621747689b9776224da468d",
						"meta": {
							"choice": [{
									"key": "false",
									"name": "\u62d2\u7edd"
								},
								{
									"key": "true",
									"name": "\u901a\u8fc7"
								}
							],
							"code": "NODE_APPROVE_RESULT",
							"type": "SELECT"
						},
						"name": "\u5ba1\u6279\u7ed3\u679c",
						"source": "global",
						"state": 2956,
						"type": "STRING"
					},
					{
						"key": "O1af1a6c7fceb2bbe9243d0cfd871028",
						"meta": {
							"code": "NODE_APPROVER"
						},
						"name": "\u5ba1\u6279\u4eba",
						"source": "global",
						"state": 2956,
						"type": "STRING"
					},
					{
						"key": "dd93d6c0341ce48260408a2964448cb7",
						"meta": {
							"code": "PROCESS_COUNT"
						},
						"name": "\u5904\u7406\u4eba\u6570",
						"source": "global",
						"state": 2956,
						"type": "INT"
					},
					{
						"key": "c6619ac6399ebb6f4208406add9d971e",
						"meta": {
							"code": "PASS_COUNT"
						},
						"name": "\u901a\u8fc7\u4eba\u6570",
						"source": "global",
						"state": 2956,
						"type": "INT"
					},
					{
						"key": "l76a275fc8b01ceeb9a33f77ddb03679",
						"meta": {
							"code": "REJECT_COUNT"
						},
						"name": "\u62d2\u7edd\u4eba\u6570",
						"source": "global",
						"state": 2956,
						"type": "INT"
					},
					{
						"key": "f73b972755824685ca4cc7edd0a0bdab",
						"meta": {
							"code": "PASS_RATE",
							"unit": "PERCENT"
						},
						"name": "\u901a\u8fc7\u7387",
						"source": "global",
						"state": 2956,
						"type": "INT"
					},
					{
						"key": "e987844249bd935b6e2b0b2609da593f",
						"meta": {
							"code": "REJECT_RATE",
							"unit": "PERCENT"
						},
						"name": "\u62d2\u7edd\u7387",
						"source": "global",
						"state": 2956,
						"type": "INT"
					}
				]
			},
			"tag": "DEFAULT",
			"processors_type": "PERSON",
			"processors": "admin",
			"assignors": "",
			"assignors_type": "EMPTY",
			"delivers": "admin",
			"delivers_type": "PERSON",
			"can_deliver": false,
			"extras": {
				"enable_terminate_ticket_when_rejected": true,
				"ticket_status": {
					"name": "",
					"type": "keep"
				}
			},
			"is_draft": false,
			"is_terminable": false,
			"fields": [
				608,
				609,
				610
			],
			"type": "APPROVAL",
			"api_instance_id": 0,
			"is_sequential": false,
			"finish_condition": {
				"expressions": [],
				"type": "or"
			},
			"is_multi": false,
			"is_allow_skip": true,
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": "admin",
			"update_at": "2024-09-14 18:12:03",
			"end_at": null,
			"is_first_state": false
		},
		"360": {
			"workflow": 64,
			"id": 360,
			"key": 360,
			"name": "\u6210\u529f\u56de\u8c03",
			"desc": "",
			"distribute_type": "PROCESS",
			"axis": {
				"x": 860,
				"y": 85
			},
			"is_builtin": false,
			"variables": {
				"inputs": [],
				"outputs": []
			},
			"tag": "DEFAULT",
			"processors_type": "PERSON",
			"processors": "admin",
			"assignors": "",
			"assignors_type": "EMPTY",
			"delivers": "",
			"delivers_type": "EMPTY",
			"can_deliver": false,
			"extras": {
				"webhook_info": {
					"auth": {
						"auth_config": {
							"token": ""
						},
						"auth_type": "bearer_token"
					},
					"body": {
						"content": "{\n    \"title\": \"{{ticket_title}}\",\n    \"currentStatus\": \"{{ticket_current_status}}\",\n    \"sn\": \"{{ticket_sn}}\",\n    \"ticketUrl\": \"{{ticket_ticket_url}}\",\n    \"applyInCluster\": true,\n    \"approveResult\": true\n}",
						"raw_type": "JSON",
						"type": "raw"
					},
					"headers": [],
					"method": "POST",
					"query_params": [],
					"settings": {
						"timeout": 10
					},
					"success_exp": "resp.code==0",
					"url": "/bcsapi/v4/bcsproject/v1/projects/{{PROJECT_CODE}}/clusters/{{CLUSTER_ID}}/namespaces/{{NAMESPACE}}/callback/create"
				}
			},
			"is_draft": false,
			"is_terminable": false,
			"fields": [],
			"type": "WEBHOOK",
			"api_instance_id": 0,
			"is_sequential": false,
			"finish_condition": {},
			"is_multi": false,
			"is_allow_skip": false,
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null,
			"is_first_state": false
		},
		"361": {
			"workflow": 64,
			"id": 361,
			"key": 361,
			"name": "\u5931\u8d25\u56de\u8c03",
			"desc": "",
			"distribute_type": "PROCESS",
			"axis": {
				"x": 865,
				"y": 210
			},
			"is_builtin": false,
			"variables": {
				"inputs": [],
				"outputs": []
			},
			"tag": "DEFAULT",
			"processors_type": "PERSON",
			"processors": "admin",
			"assignors": "",
			"assignors_type": "EMPTY",
			"delivers": "",
			"delivers_type": "EMPTY",
			"can_deliver": false,
			"extras": {
				"webhook_info": {
					"auth": {
						"auth_config": {
							"token": ""
						},
						"auth_type": "bearer_token"
					},
					"body": {
						"content": "{\n    \"title\": \"{{ticket_title}}\",\n    \"currentStatus\": \"{{ticket_current_status}}\",\n    \"sn\": \"{{ticket_sn}}\",\n    \"ticketUrl\": \"{{ticket_ticket_url}}\",\n    \"applyInCluster\": false,\n    \"approveResult\": false\n}",
						"raw_type": "JSON",
						"type": "raw"
					},
					"headers": [],
					"method": "POST",
					"query_params": [],
					"settings": {
						"timeout": 10
					},
					"success_exp": "resp.code==0",
					"url": "/bcsapi/v4/bcsproject/v1/projects/{{PROJECT_CODE}}/clusters/{{CLUSTER_ID}}/namespaces/{{NAMESPACE}}/callback/create"
				}
			},
			"is_draft": false,
			"is_terminable": false,
			"fields": [],
			"type": "WEBHOOK",
			"api_instance_id": 0,
			"is_sequential": false,
			"finish_condition": {},
			"is_multi": false,
			"is_allow_skip": false,
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null,
			"is_first_state": false
		}
	},
	"transitions": {
		"349": {
			"workflow": 64,
			"id": 349,
			"from_state": 356,
			"to_state": 357,
			"name": "",
			"axis": {
				"end": "Left",
				"start": "Right"
			},
			"condition": {
				"expressions": [{
					"expressions": [{
						"condition": "==",
						"key": "G_INT_1",
						"value": 1
					}],
					"type": "and"
				}],
				"type": "and"
			},
			"condition_type": "default",
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null
		},
		"350": {
			"workflow": 64,
			"id": 350,
			"from_state": 359,
			"to_state": 360,
			"name": "\u5ba1\u6279\u901a\u8fc7",
			"axis": {
				"end": "Left",
				"start": "Right"
			},
			"condition": {
				"expressions": [{
					"checkInfo": false,
					"expressions": [{
						"choiceList": [],
						"condition": "==",
						"key": "Fd6380d03621747689b9776224da468d",
						"meta": {
							"choice": [{
									"key": "false",
									"name": "\u62d2\u7edd"
								},
								{
									"key": "true",
									"name": "\u901a\u8fc7"
								}
							],
							"code": "NODE_APPROVE_RESULT",
							"type": "SELECT"
						},
						"source": "field",
						"type": "SELECT",
						"value": "true"
					}],
					"type": "and"
				}],
				"type": "and"
			},
			"condition_type": "by_field",
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null
		},
		"351": {
			"workflow": 64,
			"id": 351,
			"from_state": 359,
			"to_state": 361,
			"name": "\u5ba1\u6279\u9a73\u56de",
			"axis": {
				"end": "Left",
				"start": "Right"
			},
			"condition": {
				"expressions": [{
					"checkInfo": false,
					"expressions": [{
						"choiceList": [],
						"condition": "==",
						"key": "Fd6380d03621747689b9776224da468d",
						"meta": {
							"choice": [{
									"key": "false",
									"name": "\u62d2\u7edd"
								},
								{
									"key": "true",
									"name": "\u901a\u8fc7"
								}
							],
							"code": "NODE_APPROVE_RESULT",
							"type": "SELECT"
						},
						"source": "field",
						"type": "SELECT",
						"value": "false"
					}],
					"type": "and"
				}],
				"type": "and"
			},
			"condition_type": "by_field",
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null
		},
		"352": {
			"workflow": 64,
			"id": 352,
			"from_state": 360,
			"to_state": 358,
			"name": "\u6d41\u7a0b\u7ed3\u675f",
			"axis": {
				"end": "Top",
				"start": "Right"
			},
			"condition": {
				"expressions": [{
					"expressions": [{
						"condition": "==",
						"key": "G_INT_1",
						"value": 1
					}],
					"type": "and"
				}],
				"type": "and"
			},
			"condition_type": "default",
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null
		},
		"353": {
			"workflow": 64,
			"id": 353,
			"from_state": 361,
			"to_state": 358,
			"name": "\u6d41\u7a0b\u7ed3\u675f",
			"axis": {
				"end": "Bottom",
				"start": "Right"
			},
			"condition": {
				"expressions": [{
					"expressions": [{
						"condition": "==",
						"key": "G_INT_1",
						"value": 1
					}],
					"type": "and"
				}],
				"type": "and"
			},
			"condition_type": "default",
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null
		},
		"354": {
			"workflow": 64,
			"id": 354,
			"from_state": 357,
			"to_state": 359,
			"name": "\u9ed8\u8ba4",
			"axis": {
				"end": "Left",
				"start": "Right"
			},
			"condition": {
				"expressions": [{
					"expressions": [{
						"condition": "==",
						"key": "G_INT_1",
						"value": 1
					}],
					"type": "and"
				}],
				"type": "and"
			},
			"condition_type": "default",
			"creator": null,
			"create_at": "2024-09-14 18:06:10",
			"updated_by": null,
			"update_at": "2024-09-14 18:06:10",
			"end_at": null
		}
	},
	"triggers": [{
		"rules": [{
			"name": "",
			"condition": "",
			"by_condition": false,
			"action_schemas": [{
				"id": 124,
				"creator": "",
				"updated_by": "",
				"is_deleted": false,
				"name": "",
				"display_name": "",
				"component_type": "automatic_announcement",
				"operate_type": "BACKEND",
				"delay_params": {
					"type": "custom",
					"value": 0
				},
				"can_repeat": false,
				"params": [{
						"key": "web_hook_id",
						"ref_type": "custom",
						"value": "BCS_CREATE_NAMESPACE_TICKET"
					},
					{
						"key": "chat_id",
						"ref_type": "custom",
						"value": ""
					},
					{
						"key": "content",
						"ref_type": "custom",
						"value": "\u60a8\u6709\u4e00\u6761\u5355\u636e\u5f85\u5904\u7406"
					},
					{
						"key": "mentioned_list",
						"ref_type": "import",
						"value": "${ticket_current_processors}"
					}
				],
				"inputs": {}
			}]
		}],
		"id": 124,
		"creator": "",
		"updated_by": "",
		"is_deleted": false,
		"name": "\u4f01\u5fae\u901a\u77e5",
		"desc": "",
		"signal": "THROUGH_TRANSITION",
		"sender": "3002",
		"inputs": [],
		"source_type": "workflow",
		"source_id": 64,
		"source_table_id": 0,
		"is_draft": false,
		"is_enabled": true,
		"icon": "message",
		"project_key": "alkaid-test"
	}],
	"fields": {
		"607": {
			"id": 607,
			"is_deleted": false,
			"is_builtin": true,
			"is_readonly": false,
			"is_valid": true,
			"display": true,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "STRING",
			"key": "title",
			"name": "\u6807\u9898",
			"layout": "COL_12",
			"validate_type": "REQUIRE",
			"show_type": 1,
			"show_conditions": {},
			"regex": "EMPTY",
			"regex_config": {},
			"custom_regex": "",
			"desc": "\u8bf7\u8f93\u5165\u6807\u9898",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": "",
			"source": "TABLE"
		},
		"608": {
			"id": 608,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": true,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "RADIO",
			"key": "bfaba606fe9be5d6596270a00c87d428",
			"name": "\u5ba1\u6279\u610f\u89c1",
			"layout": "COL_6",
			"validate_type": "REQUIRE",
			"show_type": 1,
			"show_conditions": {},
			"regex": "EMPTY",
			"regex_config": {},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "true",
			"choice": [{
					"key": "true",
					"name": "\u901a\u8fc7"
				},
				{
					"key": "false",
					"name": "\u62d2\u7edd"
				}
			],
			"related_fields": {},
			"meta": {
				"code": "APPROVE_RESULT"
			},
			"workflow_id": 64,
			"state_id": 359,
			"source": "CUSTOM"
		},
		"609": {
			"id": 609,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": false,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "TEXT",
			"key": "ff9e6f2b83c5ea1c47f36e10310980c3",
			"name": "\u5907\u6ce8",
			"layout": "COL_12",
			"validate_type": "OPTION",
			"show_type": 0,
			"show_conditions": {
				"expressions": [{
					"condition": "==",
					"key": "bfaba606fe9be5d6596270a00c87d428",
					"type": "RADIO",
					"value": "false"
				}],
				"type": "and"
			},
			"regex": "EMPTY",
			"regex_config": {},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": 359,
			"source": "CUSTOM"
		},
		"610": {
			"id": 610,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": false,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "TEXT",
			"key": "I60e9046a05cdff0951ee0acf07d4db8",
			"name": "\u5907\u6ce8",
			"layout": "COL_12",
			"validate_type": "REQUIRE",
			"show_type": 0,
			"show_conditions": {
				"expressions": [{
					"condition": "==",
					"key": "bfaba606fe9be5d6596270a00c87d428",
					"type": "RADIO",
					"value": "true"
				}],
				"type": "and"
			},
			"regex": "EMPTY",
			"regex_config": {},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": 359,
			"source": "CUSTOM"
		},
		"611": {
			"id": 611,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": false,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "STRING",
			"key": "APP",
			"name": "\u670d\u52a1",
			"layout": "COL_12",
			"validate_type": "REQUIRE",
			"show_type": 1,
			"show_conditions": {},
			"regex": "EMPTY",
			"regex_config": {
				"rule": {
					"expressions": [{
						"condition": "",
						"key": "",
						"source": "field",
						"type": "",
						"value": ""
					}],
					"type": "and"
				}
			},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": 357,
			"source": "CUSTOM"
		},
		"612": {
			"id": 612,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": false,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "STRING",
			"key": "BIZ",
			"name": "\u4e1a\u52a1",
			"layout": "COL_12",
			"validate_type": "REQUIRE",
			"show_type": 1,
			"show_conditions": {},
			"regex": "EMPTY",
			"regex_config": {
				"rule": {
					"expressions": [{
						"condition": "",
						"key": "",
						"source": "field",
						"type": "",
						"value": ""
					}],
					"type": "and"
				}
			},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": 357,
			"source": "CUSTOM"
		},
		"613": {
			"id": 613,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": false,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "STRING",
			"key": "VERSION_NAME",
			"name": "\u4e0a\u7ebf\u7248\u672c\u540d\u79f0",
			"layout": "COL_12",
			"validate_type": "REQUIRE",
			"show_type": 1,
			"show_conditions": {},
			"regex": "EMPTY",
			"regex_config": {
				"rule": {
					"expressions": [{
						"condition": "",
						"key": "",
						"source": "field",
						"type": "",
						"value": ""
					}],
					"type": "and"
				}
			},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": 357,
			"source": "CUSTOM"
		},
		"614": {
			"id": 614,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": false,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "STRING",
			"key": "SCOPE",
			"name": "\u4e0a\u7ebf\u8303\u56f4",
			"layout": "COL_12",
			"validate_type": "REQUIRE",
			"show_type": 1,
			"show_conditions": {},
			"regex": "EMPTY",
			"regex_config": {
				"rule": {
					"expressions": [{
						"condition": "",
						"key": "",
						"source": "field",
						"type": "",
						"value": ""
					}],
					"type": "and"
				}
			},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": 357,
			"source": "CUSTOM"
		},
		"615": {
			"id": 615,
			"is_deleted": false,
			"is_builtin": false,
			"is_readonly": false,
			"is_valid": true,
			"display": false,
			"source_type": "CUSTOM",
			"source_uri": "",
			"api_instance_id": 0,
			"kv_relation": {},
			"type": "STRING",
			"key": "COMPARE",
			"name": "\u7248\u672c\u5dee\u5f02\u5bf9\u6bd4",
			"layout": "COL_12",
			"validate_type": "REQUIRE",
			"show_type": 1,
			"show_conditions": {},
			"regex": "EMPTY",
			"regex_config": {
				"rule": {
					"expressions": [{
						"condition": "",
						"key": "",
						"source": "field",
						"type": "",
						"value": ""
					}],
					"type": "and"
				}
			},
			"custom_regex": "",
			"desc": "",
			"tips": "",
			"is_tips": false,
			"default": "",
			"choice": [],
			"related_fields": {},
			"meta": {},
			"workflow_id": 64,
			"state_id": 357,
			"source": "CUSTOM"
		}
	},
	"notify": [],
	"extras": {
		"biz_related": false,
		"need_urge": false,
		"urgers_type": "EMPTY",
		"urgers": "",
		"task_settings": []
	}
}