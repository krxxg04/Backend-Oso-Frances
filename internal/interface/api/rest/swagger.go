package rest

import "github.com/gin-gonic/gin"

const docsHTML = `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Backend OF Docs</title>
  <style>
    :root {
      --bg: #f4efe7;
      --paper: rgba(255,255,255,0.82);
      --ink: #1d2a3a;
      --muted: #5d6a7b;
      --line: rgba(29,42,58,0.12);
      --accent: #c65d2e;
      --accent-2: #0c7c59;
      --shadow: 0 24px 80px rgba(29,42,58,0.12);
      --radius: 24px;
      --mono: "IBM Plex Mono", "Consolas", monospace;
      --sans: "Segoe UI", "Helvetica Neue", sans-serif;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: var(--sans);
      color: var(--ink);
      background:
        radial-gradient(circle at top left, rgba(198,93,46,0.18), transparent 28%),
        radial-gradient(circle at top right, rgba(12,124,89,0.18), transparent 24%),
        linear-gradient(180deg, #f8f2ea 0%, #f4efe7 40%, #efe7dc 100%);
    }
    .shell {
      width: min(1180px, calc(100vw - 32px));
      margin: 24px auto 64px;
    }
    .hero {
      background: linear-gradient(135deg, rgba(255,255,255,0.88), rgba(255,248,240,0.92));
      border: 1px solid var(--line);
      border-radius: 32px;
      padding: 36px;
      box-shadow: var(--shadow);
      overflow: hidden;
      position: relative;
    }
    .hero:before {
      content: "";
      position: absolute;
      inset: auto -120px -120px auto;
      width: 280px;
      height: 280px;
      border-radius: 999px;
      background: radial-gradient(circle, rgba(198,93,46,0.18), transparent 70%);
    }
    .eyebrow {
      display: inline-flex;
      gap: 10px;
      align-items: center;
      padding: 8px 12px;
      border-radius: 999px;
      background: rgba(29,42,58,0.06);
      color: var(--muted);
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: 0.12em;
    }
    h1 {
      margin: 18px 0 10px;
      font-size: clamp(38px, 5vw, 64px);
      line-height: 0.98;
      letter-spacing: -0.04em;
      max-width: 900px;
    }
    .lead {
      margin: 0;
      max-width: 760px;
      color: var(--muted);
      font-size: 18px;
      line-height: 1.65;
    }
    .hero-grid {
      display: grid;
      grid-template-columns: 1.2fr 0.8fr;
      gap: 20px;
      margin-top: 26px;
    }
    .card {
      background: var(--paper);
      border: 1px solid var(--line);
      border-radius: var(--radius);
      padding: 22px;
      backdrop-filter: blur(12px);
    }
    .card h2, .section-title {
      margin: 0 0 12px;
      font-size: 16px;
      letter-spacing: 0.02em;
      text-transform: uppercase;
    }
    .mini-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 14px;
    }
    .metric {
      border: 1px solid var(--line);
      border-radius: 18px;
      padding: 16px;
      background: rgba(255,255,255,0.66);
    }
    .metric strong {
      display: block;
      font-size: 14px;
      color: var(--muted);
      margin-bottom: 8px;
      font-weight: 600;
    }
    .metric span {
      font-family: var(--mono);
      font-size: 14px;
      word-break: break-word;
    }
    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      margin-top: 20px;
    }
    .btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      padding: 14px 18px;
      border-radius: 14px;
      text-decoration: none;
      font-weight: 600;
      transition: transform .18s ease, box-shadow .18s ease, background .18s ease;
    }
    .btn:hover { transform: translateY(-1px); }
    .btn-primary {
      background: var(--ink);
      color: white;
      box-shadow: 0 12px 32px rgba(29,42,58,0.18);
    }
    .btn-secondary {
      background: rgba(255,255,255,0.78);
      color: var(--ink);
      border: 1px solid var(--line);
    }
    .stack {
      margin: 28px 0 18px;
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 14px;
    }
    .pill {
      border-radius: 18px;
      padding: 18px;
      border: 1px solid var(--line);
      background: rgba(255,255,255,0.7);
    }
    .pill strong {
      display: block;
      margin-bottom: 6px;
      font-size: 15px;
    }
    .pill span {
      color: var(--muted);
      font-size: 14px;
      line-height: 1.5;
    }
    .reference {
      margin-top: 28px;
      background: linear-gradient(180deg, rgba(255,255,255,0.9), rgba(255,252,247,0.92));
      border: 1px solid var(--line);
      border-radius: 32px;
      box-shadow: var(--shadow);
      overflow: hidden;
    }
    .reference-header {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
      padding: 24px 26px;
      border-bottom: 1px solid var(--line);
    }
    .reference-header p {
      margin: 6px 0 0;
      color: var(--muted);
    }
    .reference-badge {
      padding: 10px 14px;
      border-radius: 999px;
      background: rgba(12,124,89,0.12);
      color: var(--accent-2);
      font-weight: 700;
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: 0.12em;
    }
    #redoc-host {
      padding: 8px 10px 22px;
    }
    footer {
      padding: 18px 4px 0;
      color: var(--muted);
      font-size: 13px;
    }
    @media (max-width: 960px) {
      .hero-grid, .stack { grid-template-columns: 1fr; }
      .mini-grid { grid-template-columns: 1fr; }
      .reference-header { flex-direction: column; align-items: flex-start; }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="hero">
      <div class="eyebrow">Backend OF • Credit Simulation Platform</div>
      <h1>Documentacion API de Credito Vehicular con una vista seria, clara y lista para presentar.</h1>
      <p class="lead">
        Esta referencia cubre autenticacion, perfil del cliente, vehiculos, bancos y simulaciones
        de Compra Inteligente. Incluye contratos, ejemplos de payloads y la especificacion OpenAPI consumida por la UI.
      </p>

      <div class="hero-grid">
        <div class="card">
          <h2>Acceso Rapido</h2>
          <div class="actions">
            <a class="btn btn-primary" href="#reference">Ver referencia API</a>
            <a class="btn btn-secondary" href="/openapi.json" target="_blank" rel="noreferrer">Abrir OpenAPI JSON</a>
            <a class="btn btn-secondary" href="/health" target="_blank" rel="noreferrer">Health Check</a>
          </div>
          <div class="stack">
            <div class="pill">
              <strong>Auth con Cookies</strong>
              <span>El backend emite <code>access_token</code> y <code>refresh_token</code> como cookies httpOnly.</span>
            </div>
            <div class="pill">
              <strong>Simulaciones Reales</strong>
              <span>El flujo soporta cuota inicial, balloon, gracia, seguros y filtros de historial.</span>
            </div>
            <div class="pill">
              <strong>OpenAPI 3.0</strong>
              <span>La referencia tecnica se sirve desde <code>/openapi.json</code> y se renderiza abajo con Redoc.</span>
            </div>
          </div>
        </div>

        <aside class="card">
          <h2>Entorno</h2>
          <div class="mini-grid">
            <div class="metric">
              <strong>Base URL</strong>
              <span>http://localhost:8080</span>
            </div>
            <div class="metric">
              <strong>Version</strong>
              <span>v1</span>
            </div>
            <div class="metric">
              <strong>Formato</strong>
              <span>application/json</span>
            </div>
            <div class="metric">
              <strong>Auth</strong>
              <span>Cookie access_token</span>
            </div>
          </div>
        </aside>
      </div>
    </section>

    <section id="reference" class="reference">
      <div class="reference-header">
        <div>
          <div class="section-title">API Reference</div>
          <p>Referencia navegable con endpoints agrupados, request bodies, respuestas, ejemplos y esquemas.</p>
        </div>
        <div class="reference-badge">Interactive Spec</div>
      </div>
      <div id="redoc-host"></div>
    </section>

    <footer>
      Backend OF • Documentacion servida por el backend en <code>/swagger</code> y <code>/docs</code>.
    </footer>
  </main>

  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  <script>
    Redoc.init('/openapi.json', {
      expandResponses: '200,201',
      hideDownloadButton: false,
      hideHostname: false,
      pathInMiddlePanel: true,
      nativeScrollbars: true,
      requiredPropsFirst: true,
      sortPropsAlphabetically: false,
      theme: {
        colors: {
          primary: { main: '#1d2a3a' },
          success: { main: '#0c7c59' },
          warning: { main: '#c65d2e' },
          text: { primary: '#1d2a3a', secondary: '#5d6a7b' },
          border: { dark: 'rgba(29,42,58,0.12)', light: 'rgba(29,42,58,0.12)' },
          responses: {
            success: { color: '#0c7c59' },
            error: { color: '#c0392b' }
          }
        },
        typography: {
          fontFamily: 'Segoe UI, Helvetica Neue, sans-serif',
          headings: { fontFamily: 'Segoe UI, Helvetica Neue, sans-serif', fontWeight: '700' },
          code: { fontFamily: 'IBM Plex Mono, Consolas, monospace' }
        },
        sidebar: {
          backgroundColor: 'rgba(255,255,255,0.62)',
          textColor: '#1d2a3a',
          activeTextColor: '#c65d2e'
        },
        rightPanel: {
          backgroundColor: '#1f2630',
          textColor: '#f7f1e8'
        }
      }
    }, document.getElementById('redoc-host'));
  </script>
</body>
</html>`

func swaggerUI(c *gin.Context) {
	c.Data(200, "text/html; charset=utf-8", []byte(docsHTML))
}
