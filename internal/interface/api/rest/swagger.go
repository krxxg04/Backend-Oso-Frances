package rest

import "github.com/gin-gonic/gin"

const swaggerUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Backend OF Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        docExpansion: 'list',
        persistAuthorization: true
      });
    };
  </script>
</body>
</html>`

func swaggerUI(c *gin.Context) {
	c.Data(200, "text/html; charset=utf-8", []byte(swaggerUIHTML))
}
