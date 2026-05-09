package http

import "github.com/gin-gonic/gin"

func openAPISpec() gin.H {
	return gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "API Credito Vehicular Compra Inteligente",
			"version":     "1.0.0",
			"description": "Backend para autenticacion, vehiculos, simulaciones, resumen financiero y cronograma de pagos.",
		},
		"paths": gin.H{
			"/api/v1/bancos":        gin.H{"get": gin.H{"summary": "Listar bancos y condiciones referenciales para credito vehicular"}},
			"/api/v1/auth/register": gin.H{"post": gin.H{"summary": "Registrar usuario con email, DNI y nombre completo"}},
			"/api/v1/auth/login":    gin.H{"post": gin.H{"summary": "Iniciar sesion"}},
			"/api/v1/auth/session":  gin.H{"get": gin.H{"summary": "Consultar sesion activa"}},
			"/api/v1/vehiculos": gin.H{
				"post": gin.H{"summary": "Registrar vehiculo del cliente"},
				"get":  gin.H{"summary": "Listar vehiculos del usuario autenticado"},
			},
			"/api/v1/vehiculos/{id}": gin.H{"get": gin.H{"summary": "Consultar vehiculo propio por ID"}},
			"/api/v1/simulaciones": gin.H{
				"post": gin.H{"summary": "Crear simulacion Compra Inteligente"},
				"get": gin.H{
					"summary": "Listar historial filtrable de simulaciones propias",
					"parameters": []gin.H{
						{"name": "fechaDesde", "in": "query", "schema": gin.H{"type": "string", "format": "date"}},
						{"name": "fechaHasta", "in": "query", "schema": gin.H{"type": "string", "format": "date"}},
						{"name": "moneda", "in": "query", "schema": gin.H{"type": "string", "enum": []string{"PEN", "USD"}}},
						{"name": "plazoMeses", "in": "query", "schema": gin.H{"type": "integer", "enum": []int{24, 36}}},
						{"name": "montoMin", "in": "query", "schema": gin.H{"type": "number"}},
						{"name": "montoMax", "in": "query", "schema": gin.H{"type": "number"}},
						{"name": "vehiculo", "in": "query", "schema": gin.H{"type": "string"}},
					},
				},
			},
			"/api/v1/simulaciones/{id}": gin.H{"get": gin.H{"summary": "Consultar simulacion propia por ID"}},
			"/api/v1/clientes/me":       gin.H{"get": gin.H{"summary": "Consultar perfil basico del cliente autenticado"}},
			"/health":                   gin.H{"get": gin.H{"summary": "Health check"}},
		},
		"components": gin.H{
			"schemas": gin.H{
				"Vehicle": gin.H{
					"type":     "object",
					"required": []string{"marca", "modelo", "precio", "moneda"},
					"properties": gin.H{
						"id":     gin.H{"type": "string"},
						"marca":  gin.H{"type": "string"},
						"modelo": gin.H{"type": "string"},
						"anio":   gin.H{"type": "integer"},
						"tipo":   gin.H{"type": "string"},
						"precio": gin.H{"type": "number"},
						"moneda": gin.H{"type": "string", "enum": []string{"PEN", "USD"}},
					},
				},
				"SimulationInput": gin.H{
					"type": "object",
					"properties": gin.H{
						"bancoId":                gin.H{"type": "string"},
						"moneda":                 gin.H{"type": "string"},
						"vehiculo":               gin.H{"$ref": "#/components/schemas/Vehicle"},
						"porcentajeCuotaInicial": gin.H{"type": "number"},
						"plazoMeses":             gin.H{"type": "integer", "enum": []int{24, 36}},
						"tasaEfectivaAnual":      gin.H{"type": "number"},
						"cuotaFinalBalloon":      gin.H{"type": "number"},
						"seguroVehicularMensual": gin.H{"type": "number"},
						"seguroDesgravamenAnual": gin.H{"type": "number"},
						"periodosGracia":         gin.H{"type": "integer"},
						"tipoGracia":             gin.H{"type": "string", "enum": []string{"sin_gracia", "parcial", "total"}},
						"fechaInicio":            gin.H{"type": "string", "format": "date"},
					},
				},
				"BankOption": gin.H{
					"type": "object",
					"properties": gin.H{
						"id":                               gin.H{"type": "string"},
						"nombre":                           gin.H{"type": "string"},
						"producto":                         gin.H{"type": "string"},
						"moneda":                           gin.H{"type": "string", "enum": []string{"PEN", "USD"}},
						"tasaEfectivaAnual":                gin.H{"type": "number"},
						"seguroDesgravamenMensual":         gin.H{"type": "number"},
						"seguroDesgravamenAnual":           gin.H{"type": "number"},
						"seguroVehicularMensualPorcentaje": gin.H{"type": "number"},
						"montoMin":                         gin.H{"type": "number"},
						"montoMax":                         gin.H{"type": "number"},
						"porcentajeCuotaInicialMin":        gin.H{"type": "number"},
						"porcentajeCuotaInicialMax":        gin.H{"type": "number"},
						"plazosMeses":                      gin.H{"type": "array", "items": gin.H{"type": "integer"}},
						"periodosGraciaMax":                gin.H{"type": "integer"},
						"fuente":                           gin.H{"type": "string"},
						"fuenteUrl":                        gin.H{"type": "string"},
						"notas":                            gin.H{"type": "string"},
					},
				},
				"FinancialSummary": gin.H{
					"type": "object",
					"properties": gin.H{
						"montoFinanciado":   gin.H{"type": "number"},
						"cuotaInicial":      gin.H{"type": "number"},
						"cuotaMensual":      gin.H{"type": "number"},
						"cuotaFinalBalloon": gin.H{"type": "number"},
						"totalIntereses":    gin.H{"type": "number"},
						"totalSeguros":      gin.H{"type": "number"},
						"totalPagado":       gin.H{"type": "number"},
						"tcea":              gin.H{"type": "number"},
						"van":               gin.H{"type": "number"},
						"tir":               gin.H{"type": "number"},
						"fechaFinalizacion": gin.H{"type": "string"},
					},
				},
			},
		},
	}
}
