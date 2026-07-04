package rest

import "github.com/gin-gonic/gin"

func openAPISpec(serverURL string) gin.H {
	return gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "Backend OF API",
			"version":     "1.0.0",
			"description": "API para autenticacion, perfil de cliente, catalogo de bancos, registro de vehiculos y simulaciones de credito vehicular Compra Inteligente.",
			"contact": gin.H{
				"name": "Backend OF",
			},
		},
		"servers": []gin.H{
			{"url": serverURL, "description": "Servidor actual"},
		},
		"tags": []gin.H{
			{"name": "Auth", "description": "Registro, login, refresh de sesion y consulta de autenticacion."},
			{"name": "Bancos", "description": "Catalogo referencial de entidades financieras y condiciones base."},
			{"name": "Clientes", "description": "Informacion basica del cliente autenticado."},
			{"name": "Vehiculos", "description": "Alta y consulta de vehiculos asociados al usuario autenticado."},
			{"name": "Simulaciones", "description": "Creacion y consulta del historial de simulaciones de credito vehicular."},
			{"name": "System", "description": "Endpoints tecnicos de salud y especificacion."},
		},
		"paths": gin.H{
			"/api/v1/auth/register": gin.H{
				"post": gin.H{
					"tags":        []string{"Auth"},
					"operationId": "registerUser",
					"summary":     "Registrar usuario",
					"description": "Crea una cuenta de usuario, registra su perfil base y devuelve sesion autenticada mediante cookies httpOnly.",
					"requestBody": jsonBodyRef("#/components/schemas/RegisterRequest", true, registerExample()),
					"responses": gin.H{
						"201": jsonResponse("Usuario registrado correctamente.", "#/components/schemas/RegisterResponse"),
						"400": errorResponse("Payload invalido o datos de validacion incorrectos."),
						"409": errorResponse("El username ya existe."),
					},
				},
			},
			"/api/v1/auth/login": gin.H{
				"post": gin.H{
					"tags":        []string{"Auth"},
					"operationId": "loginUser",
					"summary":     "Iniciar sesion",
					"description": "Autentica al usuario con username y password y emite cookies de access y refresh token.",
					"requestBody": jsonBodyRef("#/components/schemas/LoginRequest", true, loginExample()),
					"responses": gin.H{
						"200": jsonResponse("Sesion iniciada correctamente.", "#/components/schemas/LoginResponse"),
						"400": errorResponse("Faltan credenciales requeridas."),
						"401": errorResponse("Credenciales invalidas."),
						"429": errorResponse("Demasiados intentos de login."),
					},
				},
			},
			"/api/v1/auth/google/login": gin.H{
				"get": gin.H{
					"tags":        []string{"Auth"},
					"operationId": "startGoogleLogin",
					"summary":     "Iniciar login con Google",
					"description": "Redirige al usuario a Google OAuth para autenticarse o registrarse.",
					"responses": gin.H{
						"302": gin.H{"description": "Redireccion a Google OAuth."},
						"501": errorResponse("Google OAuth no configurado."),
					},
				},
			},
			"/api/v1/auth/google/callback": gin.H{
				"get": gin.H{
					"tags":        []string{"Auth"},
					"operationId": "handleGoogleCallback",
					"summary":     "Callback de Google OAuth",
					"description": "Procesa el code de Google, crea o vincula el usuario y emite cookies de sesion.",
					"parameters": []gin.H{
						queryParam("code", "string", "", "Codigo de autorizacion de Google."),
						queryParam("state", "string", "", "State anti-CSRF generado por el backend."),
					},
					"responses": gin.H{
						"302": gin.H{"description": "Redireccion al frontend con resultado del login."},
						"501": errorResponse("Google OAuth no configurado."),
					},
				},
			},
			"/api/v1/auth/logout": gin.H{
				"post": gin.H{
					"tags":        []string{"Auth"},
					"operationId": "logoutUser",
					"summary":     "Cerrar sesion",
					"description": "Elimina las cookies de autenticacion del navegador.",
					"responses": gin.H{
						"200": jsonResponse("Sesion cerrada correctamente.", "#/components/schemas/OkResponse"),
					},
				},
			},
			"/api/v1/auth/refresh": gin.H{
				"post": gin.H{
					"tags":        []string{"Auth"},
					"operationId": "refreshSession",
					"summary":     "Renovar sesion",
					"description": "Renueva las cookies de sesion usando el refresh token almacenado en cookie.",
					"responses": gin.H{
						"200": jsonResponse("Sesion renovada correctamente.", "#/components/schemas/OkResponse"),
						"401": errorResponse("Refresh token invalido o ausente."),
					},
				},
			},
			"/api/v1/auth/session": gin.H{
				"get": gin.H{
					"tags":        []string{"Auth"},
					"operationId": "getSession",
					"summary":     "Consultar sesion activa",
					"description": "Devuelve si el usuario autenticado mantiene una sesion valida.",
					"security":    cookieSecurity(),
					"responses": gin.H{
						"200": jsonResponse("Sesion valida.", "#/components/schemas/SessionResponse"),
						"401": errorResponse("Sesion no encontrada o invalida."),
					},
				},
			},
			"/api/v1/bancos": gin.H{
				"get": gin.H{
					"tags":        []string{"Bancos"},
					"operationId": "listBankOptions",
					"summary":     "Listar bancos",
					"description": "Devuelve el catalogo referencial de bancos y parametros base usados para precargar simulaciones.",
					"responses": gin.H{
						"200": jsonResponse("Catalogo de bancos.", "#/components/schemas/BankOptionListResponse"),
					},
				},
			},
			"/api/v1/clientes/me": gin.H{
				"get": gin.H{
					"tags":        []string{"Clientes"},
					"operationId": "getClientProfile",
					"summary":     "Consultar perfil del cliente autenticado",
					"description": "Retorna informacion basica del usuario autenticado.",
					"security":    cookieSecurity(),
					"responses": gin.H{
						"200": jsonResponse("Perfil del usuario autenticado.", "#/components/schemas/ClientProfile"),
						"401": errorResponse("Sesion invalida."),
						"404": errorResponse("Cliente no encontrado."),
					},
				},
				"put": gin.H{
					"tags":        []string{"Clientes"},
					"operationId": "updateClientProfile",
					"summary":     "Actualizar perfil del cliente autenticado",
					"description": "Permite actualizar email, dni, nombre completo y pictureUrl del usuario autenticado. La foto debe ser PNG.",
					"security":    cookieSecurity(),
					"requestBody": jsonBodyRef("#/components/schemas/ClientProfileUpdateRequest", true, clientProfileUpdateExample()),
					"responses": gin.H{
						"200": jsonResponse("Perfil actualizado.", "#/components/schemas/ClientProfile"),
						"400": errorResponse("Payload invalido o datos inconsistentes."),
						"401": errorResponse("Sesion invalida."),
						"404": errorResponse("Cliente no encontrado."),
					},
				},
			},
			"/api/v1/vehiculos": gin.H{
				"get": gin.H{
					"tags":        []string{"Vehiculos"},
					"operationId": "listVehicles",
					"summary":     "Listar vehiculos",
					"description": "Lista los vehiculos asociados al usuario autenticado.",
					"security":    cookieSecurity(),
					"responses": gin.H{
						"200": jsonResponse("Listado de vehiculos.", "#/components/schemas/VehicleListResponse"),
						"401": errorResponse("Sesion invalida."),
					},
				},
				"post": gin.H{
					"tags":        []string{"Vehiculos"},
					"operationId": "createVehicle",
					"summary":     "Registrar vehiculo",
					"description": "Registra un vehiculo para el usuario autenticado.",
					"security":    cookieSecurity(),
					"requestBody": jsonBodyRef("#/components/schemas/VehicleCreateRequest", true, vehicleExample()),
					"responses": gin.H{
						"201": jsonResponse("Vehiculo registrado.", "#/components/schemas/Vehicle"),
						"400": errorResponse("Errores de validacion del vehiculo."),
						"401": errorResponse("Sesion invalida."),
					},
				},
			},
			"/api/v1/vehiculos/{id}": gin.H{
				"get": gin.H{
					"tags":        []string{"Vehiculos"},
					"operationId": "getVehicleByID",
					"summary":     "Consultar vehiculo por ID",
					"description": "Obtiene un vehiculo del usuario autenticado por identificador.",
					"security":    cookieSecurity(),
					"parameters": []gin.H{
						pathParam("id", "ID del vehiculo."),
					},
					"responses": gin.H{
						"200": jsonResponse("Vehiculo encontrado.", "#/components/schemas/Vehicle"),
						"401": errorResponse("Sesion invalida."),
						"404": errorResponse("Vehiculo no encontrado."),
					},
				},
				"put": gin.H{
					"tags":        []string{"Vehiculos"},
					"operationId": "updateVehicle",
					"summary":     "Actualizar vehiculo por ID",
					"description": "Modifica los datos de un vehiculo perteneciente al usuario autenticado.",
					"security":    cookieSecurity(),
					"parameters": []gin.H{
						pathParam("id", "ID del vehiculo."),
					},
					"requestBody": jsonBodyRef("#/components/schemas/VehicleCreateRequest", true, vehicleUpdateExample()),
					"responses": gin.H{
						"200": jsonResponse("Vehiculo actualizado.", "#/components/schemas/Vehicle"),
						"400": errorResponse("Errores de validacion del vehiculo."),
						"401": errorResponse("Sesion invalida."),
						"404": errorResponse("Vehiculo no encontrado."),
					},
				},
			},
			"/api/v1/simulaciones": gin.H{
				"get": gin.H{
					"tags":        []string{"Simulaciones"},
					"operationId": "listSimulations",
					"summary":     "Listar simulaciones",
					"description": "Lista el historial del usuario autenticado y admite filtros por fecha, moneda, plazo, monto y vehiculo.",
					"security":    cookieSecurity(),
					"parameters": []gin.H{
						queryParam("fechaDesde", "string", "date", "Fecha minima del historial."),
						queryParam("fechaHasta", "string", "date", "Fecha maxima del historial."),
						enumQueryParam("moneda", "string", []string{"PEN", "USD"}, "Moneda de la simulacion."),
						enumQueryParam("plazoMeses", "integer", []int{24, 36}, "Plazo de la simulacion."),
						queryParam("montoMin", "number", "", "Monto financiado minimo."),
						queryParam("montoMax", "number", "", "Monto financiado maximo."),
						queryParam("vehiculo", "string", "", "Texto libre para marca, modelo o tipo."),
					},
					"responses": gin.H{
						"200": jsonResponse("Historial de simulaciones.", "#/components/schemas/SimulationListResponse"),
						"401": errorResponse("Sesion invalida."),
					},
				},
				"post": gin.H{
					"tags":        []string{"Simulaciones"},
					"operationId": "createSimulation",
					"summary":     "Crear simulacion",
					"description": "Genera una simulacion de credito vehicular Compra Inteligente y la guarda en el historial del usuario autenticado. bancoId es opcional; si no se envia, se usan las tasas y seguros manuales del payload.",
					"security":    cookieSecurity(),
					"requestBody": jsonBodyRef("#/components/schemas/SimulationInput", true, simulationExample()),
					"responses": gin.H{
						"201": jsonResponse("Simulacion creada.", "#/components/schemas/Simulation"),
						"400": errorResponse("Errores de validacion de la simulacion."),
						"401": errorResponse("Sesion invalida."),
					},
				},
			},
			"/api/v1/simulaciones/{id}": gin.H{
				"get": gin.H{
					"tags":        []string{"Simulaciones"},
					"operationId": "getSimulationByID",
					"summary":     "Consultar simulacion por ID",
					"description": "Recupera una simulacion del historial del usuario autenticado.",
					"security":    cookieSecurity(),
					"parameters": []gin.H{
						pathParam("id", "ID de la simulacion."),
					},
					"responses": gin.H{
						"200": jsonResponse("Simulacion encontrada.", "#/components/schemas/Simulation"),
						"401": errorResponse("Sesion invalida."),
						"404": errorResponse("Simulacion no encontrada."),
					},
				},
			},
			"/health": gin.H{
				"get": gin.H{
					"tags":        []string{"System"},
					"operationId": "healthCheck",
					"summary":     "Health check",
					"description": "Permite validar si el backend responde correctamente.",
					"responses": gin.H{
						"200": jsonResponse("Servicio disponible.", "#/components/schemas/OkResponse"),
					},
				},
			},
		},
		"components": gin.H{
			"securitySchemes": gin.H{
				"cookieAuth": gin.H{
					"type":        "apiKey",
					"in":          "cookie",
					"name":        "access_token",
					"description": "Cookie httpOnly emitida por el backend despues del login.",
				},
			},
			"schemas": gin.H{
				"OkResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"ok": gin.H{"type": "boolean", "example": true},
					},
				},
				"APIError": gin.H{
					"type": "object",
					"properties": gin.H{
						"code":    gin.H{"type": "string", "example": "validation_error"},
						"message": gin.H{"type": "string", "example": "errores de validacion"},
						"field":   gin.H{"type": "string", "example": "username"},
					},
				},
				"ErrorResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"error":  gin.H{"$ref": "#/components/schemas/APIError"},
						"errors": gin.H{"type": "array", "items": gin.H{"$ref": "#/components/schemas/APIError"}},
					},
				},
				"LoginRequest": gin.H{
					"type":     "object",
					"required": []string{"username", "password"},
					"properties": gin.H{
						"username": gin.H{"type": "string", "example": "cliente01"},
						"password": gin.H{"type": "string", "example": "secret123"},
					},
				},
				"LoginResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"user": gin.H{
							"type": "object",
							"properties": gin.H{
								"username": gin.H{"type": "string", "example": "cliente01"},
							},
						},
					},
				},
				"RegisterRequest": gin.H{
					"type":     "object",
					"required": []string{"username", "fullName", "gmail", "dni", "password", "repeatPassword"},
					"properties": gin.H{
						"username":       gin.H{"type": "string", "example": "cliente01"},
						"fullName":       gin.H{"type": "string", "example": "Cliente Demo"},
						"gmail":          gin.H{"type": "string", "example": "cliente01@email.com"},
						"dni":            gin.H{"type": "string", "example": "12345678"},
						"password":       gin.H{"type": "string", "example": "secret123"},
						"repeatPassword": gin.H{"type": "string", "example": "secret123"},
					},
				},
				"RegisterResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"user": gin.H{
							"type": "object",
							"properties": gin.H{
								"username": gin.H{"type": "string", "example": "cliente01"},
								"fullName": gin.H{"type": "string", "example": "Cliente Demo"},
								"gmail":    gin.H{"type": "string", "example": "cliente01@email.com"},
								"dni":      gin.H{"type": "string", "example": "12345678"},
							},
						},
					},
				},
				"SessionResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"authenticated": gin.H{"type": "boolean", "example": true},
						"user": gin.H{
							"type": "object",
							"properties": gin.H{
								"username": gin.H{"type": "string", "example": "cliente01"},
							},
						},
					},
				},
				"ClientProfile": gin.H{
					"type": "object",
					"properties": gin.H{
						"username":   gin.H{"type": "string", "example": "cliente01"},
						"email":      gin.H{"type": "string", "example": "cliente01@email.com"},
						"dni":        gin.H{"type": "string", "example": "12345678"},
						"fullName":   gin.H{"type": "string", "example": "Cliente Demo"},
						"pictureUrl": gin.H{"type": "string", "example": "https://example.com/avatar.png", "description": "URL o data URL de una imagen PNG."},
						"role":       gin.H{"type": "string", "example": "user"},
					},
				},
				"ClientProfileUpdateRequest": gin.H{
					"type": "object",
					"properties": gin.H{
						"email":      gin.H{"type": "string", "example": "cliente01@email.com"},
						"dni":        gin.H{"type": "string", "example": "12345678"},
						"fullName":   gin.H{"type": "string", "example": "Cliente Demo"},
						"pictureUrl": gin.H{"type": "string", "example": "https://example.com/avatar.png", "description": "Solo se acepta PNG."},
					},
				},
				"Vehicle": gin.H{
					"type":     "object",
					"required": []string{"marca", "modelo", "precio", "moneda"},
					"properties": gin.H{
						"id":       gin.H{"type": "string", "example": "veh_20260622183000_1"},
						"userId":   gin.H{"type": "string", "example": "9f4b47e1-3ed2-4f3a-a650-4fb457001122"},
						"marca":    gin.H{"type": "string", "example": "Toyota"},
						"modelo":   gin.H{"type": "string", "example": "Yaris"},
						"anio":     gin.H{"type": "integer", "example": 2025},
						"tipo":     gin.H{"type": "string", "example": "sedan"},
						"precio":   gin.H{"type": "number", "example": 8000, "minimum": 2000, "maximum": 500000},
						"moneda":   gin.H{"type": "string", "enum": []string{"PEN", "USD"}, "example": "PEN"},
						"creadoEn": gin.H{"type": "string", "format": "date-time"},
					},
				},
				"VehicleCreateRequest": gin.H{
					"type":     "object",
					"required": []string{"marca", "modelo", "precio", "moneda"},
					"properties": gin.H{
						"marca":  gin.H{"type": "string", "example": "Toyota"},
						"modelo": gin.H{"type": "string", "example": "Yaris"},
						"anio":   gin.H{"type": "integer", "example": 2025},
						"tipo":   gin.H{"type": "string", "example": "sedan"},
						"precio": gin.H{"type": "number", "example": 8000, "minimum": 2000, "maximum": 500000},
						"moneda": gin.H{"type": "string", "enum": []string{"PEN", "USD"}, "example": "PEN"},
					},
				},
				"VehicleListResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"items": gin.H{
							"type":  "array",
							"items": gin.H{"$ref": "#/components/schemas/Vehicle"},
						},
					},
				},
				"BankOption": gin.H{
					"type": "object",
					"properties": gin.H{
						"id":                               gin.H{"type": "string", "example": "bbva-vehicular-sostenible"},
						"nombre":                           gin.H{"type": "string", "example": "BBVA"},
						"producto":                         gin.H{"type": "string", "example": "Prestamo Vehicular Sostenible"},
						"moneda":                           gin.H{"type": "string", "enum": []string{"PEN", "USD"}, "example": "PEN"},
						"tasaEfectivaAnual":                gin.H{"type": "number", "example": 11.49},
						"seguroDesgravamenMensual":         gin.H{"type": "number", "example": 0.069},
						"seguroDesgravamenAnual":           gin.H{"type": "number", "example": 0.0083},
						"seguroVehicularMensualPorcentaje": gin.H{"type": "number", "example": 0.50},
						"montoMin":                         gin.H{"type": "number", "example": 28800},
						"montoMax":                         gin.H{"type": "number", "example": 120000},
						"porcentajeCuotaInicialMin":        gin.H{"type": "number", "example": 0},
						"porcentajeCuotaInicialMax":        gin.H{"type": "number", "example": 50},
						"plazosMeses":                      gin.H{"type": "array", "items": gin.H{"type": "integer"}, "example": []int{24, 36}},
						"periodosGraciaMax":                gin.H{"type": "integer", "example": 1},
						"fuente":                           gin.H{"type": "string", "example": "BBVA Peru"},
						"fuenteUrl":                        gin.H{"type": "string", "example": "https://www.bbva.pe/"},
						"notas":                            gin.H{"type": "string", "example": "Valores referenciales para precarga del simulador."},
					},
				},
				"BankOptionListResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"items": gin.H{
							"type":  "array",
							"items": gin.H{"$ref": "#/components/schemas/BankOption"},
						},
					},
				},
				"Pago": gin.H{
					"type": "object",
					"properties": gin.H{
						"mes":                 gin.H{"type": "integer", "example": 1},
						"periodo":             gin.H{"type": "integer", "example": 1},
						"fecha":               gin.H{"type": "string", "example": "2026-07-01"},
						"saldoInicial":        gin.H{"type": "number", "example": 64000},
						"cuota":               gin.H{"type": "number", "example": 2381.42},
						"cuotaCapitalInteres": gin.H{"type": "number", "example": 2135.12},
						"interes":             gin.H{"type": "number", "example": 879.44},
						"seguroVehicular":     gin.H{"type": "number", "example": 180},
						"seguroDesgravamen":   gin.H{"type": "number", "example": 66.30},
						"seguro":              gin.H{"type": "number", "example": 246.30},
						"amortizacion":        gin.H{"type": "number", "example": 1255.68},
						"saldoFinal":          gin.H{"type": "number", "example": 62744.32},
						"saldoDeudor":         gin.H{"type": "number", "example": 62744.32},
						"tipoGracia":          gin.H{"type": "string", "enum": []string{"sin_gracia", "parcial", "total"}, "example": "sin_gracia"},
					},
				},
				"FinancialSummary": gin.H{
					"type": "object",
					"properties": gin.H{
						"montoFinanciado":   gin.H{"type": "number", "example": 64000},
						"cuotaInicial":      gin.H{"type": "number", "example": 16000},
						"cuotaMensual":      gin.H{"type": "number", "example": 2381.42},
						"cuotaFinalBalloon": gin.H{"type": "number", "example": 24000},
						"totalIntereses":    gin.H{"type": "number", "example": 12980.50},
						"totalSeguros":      gin.H{"type": "number", "example": 5321.20},
						"totalPagado":       gin.H{"type": "number", "example": 82301.70},
						"van":               gin.H{"type": "number", "example": 0.01},
						"tir":               gin.H{"type": "number", "example": 0.014},
						"tcea":              gin.H{"type": "number", "example": 0.181},
						"fechaFinalizacion": gin.H{"type": "string", "example": "2029-06-01"},
					},
				},
				"SimulationResult": gin.H{
					"type": "object",
					"properties": gin.H{
						"banco":           gin.H{"$ref": "#/components/schemas/BankOption"},
						"tasaPeriodo":     gin.H{"type": "number", "example": 0.013},
						"cuotaBase":       gin.H{"type": "number", "example": 2135.12},
						"van":             gin.H{"type": "number", "example": 0.01},
						"tir":             gin.H{"type": "number", "example": 0.014},
						"tcea":            gin.H{"type": "number", "example": 0.181},
						"costoTotal":      gin.H{"type": "number", "example": 82301.70},
						"totalIntereses":  gin.H{"type": "number", "example": 12980.50},
						"totalSeguros":    gin.H{"type": "number", "example": 5321.20},
						"cuotaInicial":    gin.H{"type": "number", "example": 16000},
						"montoNeto":       gin.H{"type": "number", "example": 64000},
						"montoFinanciado": gin.H{"type": "number", "example": 64000},
						"resumen":         gin.H{"$ref": "#/components/schemas/FinancialSummary"},
						"cronograma": gin.H{
							"type":  "array",
							"items": gin.H{"$ref": "#/components/schemas/Pago"},
						},
					},
				},
				"SimulationInput": gin.H{
					"type":     "object",
					"required": []string{"porcentajeCuotaInicial", "plazoMeses"},
					"properties": gin.H{
						"nombreCliente":            gin.H{"type": "string", "example": "cliente01"},
						"bancoId":                  gin.H{"type": "string", "example": "manual", "description": "Opcional. Usa un id del catalogo o 'manual' para ingresar tasas y seguros manualmente."},
						"moneda":                   gin.H{"type": "string", "enum": []string{"PEN", "USD"}, "example": "PEN"},
						"vehiculo":                 gin.H{"$ref": "#/components/schemas/VehicleCreateRequest"},
						"fechaInicio":              gin.H{"type": "string", "format": "date", "example": "2026-06-01"},
						"precioVehiculo":           gin.H{"type": "number", "example": 80000, "minimum": 2000, "maximum": 500000},
						"porcentajeCuotaInicial":   gin.H{"type": "number", "example": 20},
						"plazoMeses":               gin.H{"type": "integer", "enum": []int{24, 36}, "example": 36},
						"tasaAnual":                gin.H{"type": "number", "example": 18},
						"tipoTasa":                 gin.H{"type": "string", "enum": []string{"efectiva", "nominal"}, "example": "nominal"},
						"tasaEfectivaAnual":        gin.H{"type": "number", "example": 18},
						"frecuenciaCapitalizacion": gin.H{"type": "integer", "example": 12},
						"periodosPorAnio":          gin.H{"type": "integer", "example": 12},
						"periodosGracia":           gin.H{"type": "integer", "example": 0, "minimum": 0, "maximum": 6},
						"tipoGracia":               gin.H{"type": "string", "enum": []string{"sin_gracia", "parcial", "total"}, "example": "sin_gracia"},
						"valorFinal":               gin.H{"type": "number", "example": 24000, "description": "No puede superar el 50% del precio del vehiculo."},
						"cuotaFinalBalloon":        gin.H{"type": "number", "example": 24000, "description": "No puede superar el 50% del precio del vehiculo."},
						"seguroVehicularMensual":   gin.H{"type": "number", "example": 180, "minimum": 0, "maximum": 5000},
						"seguroDesgravamenAnual":   gin.H{"type": "number", "example": 1.2},
						"costosFinanciados":        gin.H{"type": "number", "example": 0},
						"costosIniciales":          gin.H{"type": "number", "example": 0},
					},
				},
				"Simulation": gin.H{
					"type": "object",
					"properties": gin.H{
						"id":       gin.H{"type": "string", "example": "sim_20260622183500_1"},
						"userId":   gin.H{"type": "string", "example": "9f4b47e1-3ed2-4f3a-a650-4fb457001122"},
						"creadoEn": gin.H{"type": "string", "format": "date-time"},
						"input":    gin.H{"$ref": "#/components/schemas/SimulationInput"},
						"result":   gin.H{"$ref": "#/components/schemas/SimulationResult"},
					},
				},
				"SimulationListResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"items": gin.H{
							"type":  "array",
							"items": gin.H{"$ref": "#/components/schemas/Simulation"},
						},
					},
				},
			},
		},
	}
}

func cookieSecurity() []gin.H {
	return []gin.H{{"cookieAuth": []string{}}}
}

func pathParam(name, description string) gin.H {
	return gin.H{
		"name":        name,
		"in":          "path",
		"required":    true,
		"description": description,
		"schema":      gin.H{"type": "string"},
	}
}

func queryParam(name, dataType, format, description string) gin.H {
	schema := gin.H{"type": dataType}
	if format != "" {
		schema["format"] = format
	}
	return gin.H{
		"name":        name,
		"in":          "query",
		"required":    false,
		"description": description,
		"schema":      schema,
	}
}

func enumQueryParam(name, dataType string, values any, description string) gin.H {
	return gin.H{
		"name":        name,
		"in":          "query",
		"required":    false,
		"description": description,
		"schema": gin.H{
			"type": dataType,
			"enum": values,
		},
	}
}

func jsonBodyRef(schemaRef string, required bool, example gin.H) gin.H {
	return gin.H{
		"required": required,
		"content": gin.H{
			"application/json": gin.H{
				"schema":  gin.H{"$ref": schemaRef},
				"example": example,
			},
		},
	}
}

func jsonResponse(description, schemaRef string) gin.H {
	return gin.H{
		"description": description,
		"content": gin.H{
			"application/json": gin.H{
				"schema": gin.H{"$ref": schemaRef},
			},
		},
	}
}

func errorResponse(description string) gin.H {
	return jsonResponse(description, "#/components/schemas/ErrorResponse")
}

func loginExample() gin.H {
	return gin.H{
		"username": "cliente01",
		"password": "secret123",
	}
}

func registerExample() gin.H {
	return gin.H{
		"username":       "cliente01",
		"fullName":       "Cliente Demo",
		"gmail":          "cliente01@email.com",
		"dni":            "12345678",
		"password":       "secret123",
		"repeatPassword": "secret123",
	}
}

func vehicleExample() gin.H {
	return gin.H{
		"marca":  "Toyota",
		"modelo": "Yaris",
		"anio":   2025,
		"tipo":   "sedan",
		"precio": 80000,
		"moneda": "PEN",
	}
}

func vehicleUpdateExample() gin.H {
	return gin.H{
		"marca":  "Toyota",
		"modelo": "Corolla Cross",
		"anio":   2026,
		"tipo":   "suv",
		"precio": 95000,
		"moneda": "PEN",
	}
}

func clientProfileUpdateExample() gin.H {
	return gin.H{
		"email":      "cliente01@email.com",
		"dni":        "12345678",
		"fullName":   "Cliente Demo",
		"pictureUrl": "https://example.com/avatar.png",
	}
}

func simulationExample() gin.H {
	return gin.H{
		"bancoId": "manual",
		"moneda":  "PEN",
		"vehiculo": gin.H{
			"marca":  "Toyota",
			"modelo": "Yaris",
			"anio":   2025,
			"tipo":   "sedan",
			"precio": 80000,
			"moneda": "PEN",
		},
		"porcentajeCuotaInicial": 20,
		"plazoMeses":             36,
		"periodosPorAnio":        12,
		"cuotaFinalBalloon":      24000,
		"seguroVehicularMensual": 180,
		"seguroDesgravamenAnual": 1.2,
		"periodosGracia":         0,
		"tipoGracia":             "sin_gracia",
		"fechaInicio":            "2026-06-01",
	}
}
