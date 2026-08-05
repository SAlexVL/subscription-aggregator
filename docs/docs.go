package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/subscriptions": {
            "get": {
                "produces": ["application/json"],
                "tags": ["subscriptions"],
                "summary": "List subscriptions for a user",
                "parameters": [
                    {"type":"string","description":"User UUID (required)","name":"user_id","in":"query","required":true},
                    {"type":"string","description":"Filter by service name","name":"service_name","in":"query"},
                    {"type":"integer","description":"Page size (default 20, max 100)","name":"limit","in":"query"},
                    {"type":"integer","description":"Offset (default 0)","name":"offset","in":"query"}
                ],
                "responses": {
                    "200": {"description":"OK","schema":{"$ref":"#/definitions/model.ListResponse"}},
                    "400": {"description":"Bad Request","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "500": {"description":"Internal Server Error","schema":{"$ref":"#/definitions/handler.errorResponse"}}
                }
            },
            "post": {
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["subscriptions"],
                "summary": "Create subscription",
                "parameters": [
                    {"description":"Subscription","name":"body","in":"body","required":true,"schema":{"$ref":"#/definitions/model.CreateSubscriptionRequest"}}
                ],
                "responses": {
                    "201": {"description":"Created","schema":{"$ref":"#/definitions/model.Subscription"}},
                    "400": {"description":"Bad Request","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "500": {"description":"Internal Server Error","schema":{"$ref":"#/definitions/handler.errorResponse"}}
                }
            }
        },
        "/subscriptions/sum": {
            "get": {
                "description": "Calculates total cost as price * overlapping months",
                "produces": ["application/json"],
                "tags": ["subscriptions"],
                "summary": "Sum subscription costs for a period",
                "parameters": [
                    {"type":"string","description":"Period start (MM-YYYY)","name":"from","in":"query","required":true},
                    {"type":"string","description":"Period end (MM-YYYY)","name":"to","in":"query","required":true},
                    {"type":"string","description":"Filter by user UUID","name":"user_id","in":"query"},
                    {"type":"string","description":"Filter by service name","name":"service_name","in":"query"}
                ],
                "responses": {
                    "200": {"description":"OK","schema":{"$ref":"#/definitions/model.SumResponse"}},
                    "400": {"description":"Bad Request","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "500": {"description":"Internal Server Error","schema":{"$ref":"#/definitions/handler.errorResponse"}}
                }
            }
        },
        "/subscriptions/{id}": {
            "get": {
                "produces": ["application/json"],
                "tags": ["subscriptions"],
                "summary": "Get subscription by ID",
                "parameters": [
                    {"type":"string","description":"Subscription ID","name":"id","in":"path","required":true}
                ],
                "responses": {
                    "200": {"description":"OK","schema":{"$ref":"#/definitions/model.Subscription"}},
                    "400": {"description":"Bad Request","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "404": {"description":"Not Found","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "500": {"description":"Internal Server Error","schema":{"$ref":"#/definitions/handler.errorResponse"}}
                }
            },
            "put": {
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["subscriptions"],
                "summary": "Update subscription (partial)",
                "parameters": [
                    {"type":"string","description":"Subscription ID","name":"id","in":"path","required":true},
                    {"description":"Fields to update","name":"body","in":"body","required":true,"schema":{"$ref":"#/definitions/model.UpdateSubscriptionRequest"}}
                ],
                "responses": {
                    "200": {"description":"OK","schema":{"$ref":"#/definitions/model.Subscription"}},
                    "400": {"description":"Bad Request","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "404": {"description":"Not Found","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "500": {"description":"Internal Server Error","schema":{"$ref":"#/definitions/handler.errorResponse"}}
                }
            },
            "patch": {
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["subscriptions"],
                "summary": "Update subscription (partial)",
                "parameters": [
                    {"type":"string","description":"Subscription ID","name":"id","in":"path","required":true},
                    {"description":"Fields to update","name":"body","in":"body","required":true,"schema":{"$ref":"#/definitions/model.UpdateSubscriptionRequest"}}
                ],
                "responses": {
                    "200": {"description":"OK","schema":{"$ref":"#/definitions/model.Subscription"}},
                    "400": {"description":"Bad Request","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "404": {"description":"Not Found","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "500": {"description":"Internal Server Error","schema":{"$ref":"#/definitions/handler.errorResponse"}}
                }
            },
            "delete": {
                "produces": ["application/json"],
                "tags": ["subscriptions"],
                "summary": "Delete subscription",
                "parameters": [
                    {"type":"string","description":"Subscription ID","name":"id","in":"path","required":true}
                ],
                "responses": {
                    "204": {"description":"No Content"},
                    "400": {"description":"Bad Request","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "404": {"description":"Not Found","schema":{"$ref":"#/definitions/handler.errorResponse"}},
                    "500": {"description":"Internal Server Error","schema":{"$ref":"#/definitions/handler.errorResponse"}}
                }
            }
        }
    },
    "definitions": {
        "handler.errorResponse": {
            "type": "object",
            "properties": {"error": {"type": "string"}}
        },
        "model.CreateSubscriptionRequest": {
            "type": "object",
            "required": ["service_name", "price", "user_id", "start_date"],
            "properties": {
                "service_name": {"type": "string", "example": "Yandex Plus"},
                "price": {"type": "integer", "example": 400},
                "user_id": {"type": "string", "example": "60601fee-2bf1-4721-ae6f-7636e79a0cba"},
                "start_date": {"type": "string", "example": "07-2025"},
                "end_date": {"type": "string", "example": "12-2025"}
            }
        },
        "model.UpdateSubscriptionRequest": {
            "type": "object",
            "properties": {
                "service_name": {"type": "string"},
                "price": {"type": "integer"},
                "user_id": {"type": "string"},
                "start_date": {"type": "string"},
                "end_date": {"type": "string"}
            }
        },
        "model.Subscription": {
            "type": "object",
            "properties": {
                "id": {"type": "string"},
                "service_name": {"type": "string"},
                "price": {"type": "integer"},
                "user_id": {"type": "string"},
                "start_date": {"type": "string"},
                "end_date": {"type": "string"},
                "created_at": {"type": "string"},
                "updated_at": {"type": "string"}
            }
        },
        "model.ListResponse": {
            "type": "object",
            "properties": {
                "items": {"type": "array", "items": {"$ref": "#/definitions/model.Subscription"}},
                "total": {"type": "integer"},
                "limit": {"type": "integer"},
                "offset": {"type": "integer"}
            }
        },
        "model.SumResponse": {
            "type": "object",
            "properties": {
                "total": {"type": "integer"},
                "from": {"type": "string"},
                "to": {"type": "string"},
                "user_id": {"type": "string"},
                "service_name": {"type": "string"}
            }
        }
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Subscription Aggregator API",
	Description:      "REST API for aggregating online subscription data",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
