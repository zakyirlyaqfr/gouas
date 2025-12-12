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
        "/api/v1/auth/login": {
            "post": {
                "description": "Login user",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["5.1 Authentication"],
                "summary": "Login User",
                "parameters": [
                    {
                        "description": "Credentials",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "username": { "type": "string", "example": "admin" },
                                "password": { "type": "string", "example": "admin123" }
                            }
                        }
                    }
                ],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/auth/refresh": {
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.1 Authentication"],
                "summary": "Refresh Token",
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/auth/logout": {
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.1 Authentication"],
                "summary": "Logout",
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/auth/profile": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.1 Authentication"],
                "summary": "Get Current Profile",
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/users": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.2 Users (Admin)"],
                "summary": "List Users",
                "responses": { "200": { "description": "OK" } }
            },
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.2 Users (Admin)"],
                "summary": "Create User",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "username": { "type": "string" },
                                "email": { "type": "string" },
                                "password": { "type": "string" },
                                "fullName": { "type": "string" },
                                "roleName": { "type": "string", "enum": ["Mahasiswa", "Dosen Wali", "Admin"] }
                            }
                        }
                    }
                ],
                "responses": { "201": { "description": "Created" } }
            }
        },
        "/api/v1/users/{id}": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.2 Users (Admin)"],
                "summary": "Get User Detail",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            },
            "put": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.2 Users (Admin)"],
                "summary": "Update User",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            },
            "delete": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.2 Users (Admin)"],
                "summary": "Delete User",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/users/{id}/role": {
            "put": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.2 Users (Admin)"],
                "summary": "Update User Role",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/achievements": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "List Achievements (Filtered)",
                "responses": { "200": { "description": "OK" } }
            },
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Create Achievement (Mahasiswa)",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "title": { "type": "string" },
                                "achievementType": { "type": "string", "enum": ["Competition", "Organization", "Publication", "Certification"] },
                                "description": { "type": "string" }
                            }
                        }
                    }
                ],
                "responses": { "201": { "description": "Created" } }
            }
        },
        "/api/v1/achievements/{id}": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Get Achievement Detail",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            },
            "put": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Update Achievement",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            },
            "delete": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Delete Achievement",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/achievements/{id}/submit": {
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Submit for Verification",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/achievements/{id}/verify": {
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Verify Achievement (Dosen Wali)",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/achievements/{id}/reject": {
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Reject Achievement (Dosen Wali)",
                "parameters": [
                    { "name": "id", "in": "path", "required": true, "type": "string" },
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "note": { "type": "string" }
                            }
                        }
                    }
                ],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/achievements/{id}/history": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Get Status History",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/achievements/{id}/attachments": {
            "post": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.4 Achievements"],
                "summary": "Upload Attachment",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/students": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.5 Students & Lecturers"],
                "summary": "List Students",
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/students/{id}": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.5 Students & Lecturers"],
                "summary": "Get Student Detail",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/students/{id}/achievements": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.5 Students & Lecturers"],
                "summary": "Get Student Achievements",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/students/{id}/advisor": {
            "put": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.5 Students & Lecturers"],
                "summary": "Assign Advisor",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/lecturers": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.5 Students & Lecturers"],
                "summary": "List Lecturers",
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/lecturers/{id}/advisees": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.5 Students & Lecturers"],
                "summary": "Get Advisees (Mahasiswa Bimbingan)",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/reports/statistics": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.8 Reports & Analytics"],
                "summary": "Get General Statistics",
                "responses": { "200": { "description": "OK" } }
            }
        },
        "/api/v1/reports/student/{id}": {
            "get": {
                "security": [{"BearerAuth": []}],
                "tags": ["5.8 Reports & Analytics"],
                "summary": "Get Student Report",
                "parameters": [{ "name": "id", "in": "path", "required": true, "type": "string" }],
                "responses": { "200": { "description": "OK" } }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:3000",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Sistem Pelaporan Prestasi API",
	Description:      "API Documentation for GOUAS Project",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}