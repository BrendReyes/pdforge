package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var servePort int
var serveNoOpen bool

var serveCmd = &cobra.Command{
  Use:   "serve",
  Short: "Serve the local coss ui web app for pdforge",
  Long: `The serve command starts the local pdforge web UI.
When a built coss ui frontend exists under web/dist, it is served first.
Otherwise, pdforge falls back to the secure server-rendered preview used for early development.`,
	Example: `  pdforge serve
  pdforge serve --port 3000`,
	Args: cobra.NoArgs,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port for the local web server")
  serveCmd.Flags().BoolVar(&serveNoOpen, "no-open", false, "Do not open the browser automatically")
}

func runServe(cmd *cobra.Command, args []string) error {
	if servePort < 1 || servePort > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", servePort)
	}

  server, err := newWebServer(servePort)
  if err != nil {
    return err
  }
  defer server.Close()

  if serveNoOpen {
    return server.Run(cmd.OutOrStdout())
  }

  return server.RunWithAutoOpen(cmd.OutOrStdout())
}

const heroUISampleHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>pdforge HeroUI Sample</title>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
  <link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;700&display=swap" rel="stylesheet" />
  <style>
    :root {
      --bg-0: #f8fbff;
      --bg-1: #ecf4ff;
      --surface: rgba(255, 255, 255, 0.78);
      --surface-strong: #ffffff;
      --ink: #0f172a;
      --muted: #475569;
      --line: rgba(148, 163, 184, 0.25);
      --brand: #0ea5e9;
      --brand-2: #2563eb;
      --accent: #14b8a6;
      --ok: #16a34a;
      --warn: #f59e0b;
      --radius-xl: 26px;
      --radius-lg: 18px;
      --radius-md: 12px;
      --shadow-soft: 0 16px 40px rgba(15, 23, 42, 0.12);
      --shadow-card: 0 10px 22px rgba(37, 99, 235, 0.14);
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      font-family: "Space Grotesk", "Segoe UI", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(1100px 500px at 5% -15%, #c8ecff 0%, transparent 60%),
        radial-gradient(1000px 460px at 100% -10%, #b7d0ff 0%, transparent 63%),
        linear-gradient(160deg, var(--bg-0) 0%, var(--bg-1) 100%);
      min-height: 100vh;
    }

    .wrap {
      max-width: 1180px;
      margin: 0 auto;
      padding: 34px 20px 62px;
    }

    .topbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 14px;
      padding: 12px 14px;
      border: 1px solid var(--line);
      background: var(--surface);
      backdrop-filter: blur(14px);
      border-radius: 999px;
      box-shadow: var(--shadow-soft);
      position: sticky;
      top: 14px;
      z-index: 10;
    }

    .brand {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 800;
      letter-spacing: 0.015em;
    }

    .logo {
      width: 30px;
      height: 30px;
      display: inline-grid;
      place-items: center;
      border-radius: 10px;
      background: linear-gradient(145deg, var(--brand) 0%, var(--brand-2) 75%);
      color: #fff;
      font-size: 13px;
      font-weight: 900;
      box-shadow: 0 8px 18px rgba(37, 99, 235, 0.45);
    }

    .nav-pills {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }

    .pill {
      border: 1px solid var(--line);
      background: rgba(255, 255, 255, 0.92);
      border-radius: 999px;
      font-size: 13px;
      padding: 7px 11px;
      color: var(--muted);
    }

    .hero {
      margin-top: 28px;
      display: grid;
      grid-template-columns: 1.25fr 0.75fr;
      gap: 16px;
      align-items: stretch;
    }

    .card {
      border: 1px solid var(--line);
      background: var(--surface);
      border-radius: var(--radius-xl);
      box-shadow: var(--shadow-soft);
      padding: 20px;
      backdrop-filter: blur(10px);
    }

    h1 {
      font-size: clamp(2rem, 4vw, 3.25rem);
      line-height: 1.05;
      margin: 8px 0 10px;
      letter-spacing: -0.03em;
    }

    p {
      margin: 0;
      color: var(--muted);
      line-height: 1.6;
    }

    .label {
      display: inline-flex;
      align-items: center;
      gap: 7px;
      width: fit-content;
      border-radius: 999px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.02em;
      padding: 6px 10px;
      color: #0b4f7a;
      background: linear-gradient(145deg, #d7f0ff 0%, #cbe3ff 100%);
      border: 1px solid rgba(14, 165, 233, 0.28);
    }

    .cta-row {
      margin-top: 18px;
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }

    .btn {
      border-radius: 14px;
      border: 1px solid var(--line);
      background: rgba(255, 255, 255, 0.94);
      color: var(--ink);
      padding: 11px 15px;
      font-weight: 700;
      font-size: 14px;
      cursor: pointer;
      transition: transform 180ms ease, box-shadow 180ms ease, border-color 180ms ease;
    }

    .btn:hover {
      transform: translateY(-1px);
      box-shadow: var(--shadow-card);
      border-color: rgba(37, 99, 235, 0.45);
    }

    .btn.primary {
      background: linear-gradient(145deg, var(--brand) 0%, var(--brand-2) 78%);
      color: #fff;
      border-color: transparent;
      box-shadow: 0 10px 22px rgba(14, 165, 233, 0.4);
    }

    .stack {
      display: grid;
      gap: 12px;
    }

    .stat {
      border: 1px solid var(--line);
      border-radius: var(--radius-lg);
      padding: 14px 14px 12px;
      background: var(--surface-strong);
    }

    .stat h3 {
      margin: 0;
      font-weight: 700;
      font-size: 14px;
    }

    .meter {
      margin-top: 9px;
      height: 9px;
      border-radius: 999px;
      background: #e2ecf8;
      overflow: hidden;
    }

    .meter > span {
      display: block;
      height: 100%;
      width: 72%;
      background: linear-gradient(90deg, #0ea5e9, #2563eb);
    }

    .features {
      margin-top: 18px;
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 12px;
    }

    .feature {
      border: 1px solid var(--line);
      border-radius: var(--radius-lg);
      background: rgba(255, 255, 255, 0.88);
      padding: 15px 14px;
      display: grid;
      gap: 7px;
      transition: transform 180ms ease, box-shadow 180ms ease;
    }

    .feature:hover {
      transform: translateY(-2px);
      box-shadow: var(--shadow-card);
    }

    .feature h3 {
      margin: 0;
      font-size: 0.98rem;
      letter-spacing: -0.01em;
    }

    .tiny {
      display: inline-flex;
      font-size: 12px;
      font-weight: 700;
      color: #0369a1;
    }

    .ok {
      color: var(--ok);
      font-weight: 700;
      font-size: 12px;
    }

    .warn {
      color: var(--warn);
      font-size: 12px;
      font-weight: 700;
    }

    .fade-in {
      opacity: 0;
      transform: translateY(8px);
      animation: rise 500ms ease forwards;
    }

    .delay-1 { animation-delay: 80ms; }
    .delay-2 { animation-delay: 160ms; }
    .delay-3 { animation-delay: 220ms; }

    @keyframes rise {
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }

    @media (max-width: 920px) {
      .hero {
        grid-template-columns: 1fr;
      }

      .features {
        grid-template-columns: repeat(2, minmax(0, 1fr));
      }

      .nav-pills {
        display: none;
      }
    }

    @media (max-width: 620px) {
      .features {
        grid-template-columns: 1fr;
      }

      .topbar {
        border-radius: 18px;
      }

      .wrap {
        padding: 20px 14px 46px;
      }
    }
  </style>
</head>
<body>
  <div class="wrap">
    <header class="topbar fade-in">
      <div class="brand">
        <span class="logo">PF</span>
        <span>pdforge web</span>
      </div>
      <div class="nav-pills">
        <span class="pill">HeroUI inspired</span>
        <span class="pill">local-first processing</span>
        <span class="pill">MVP preview</span>
      </div>
    </header>

    <section class="hero fade-in delay-1">
      <article class="card">
        <span class="label">Web MVP Platform</span>
        <h1>Privacy-first PDF workflows, now in your browser.</h1>
        <p>This draft uses a HeroUI-like visual style with soft glass surfaces, high-contrast action buttons, and compact cards. It is served directly by pdforge for local experimentation.</p>
        <div class="cta-row">
          <button class="btn primary">Upload Files</button>
          <button class="btn">Open Work Queue</button>
          <button class="btn">Inspect Endpoints</button>
        </div>
        <div class="features">
          <article class="feature">
            <span class="tiny">01 Toolchain</span>
            <h3>Merge PDFs</h3>
            <p>Combine multiple files into one ordered output document.</p>
          </article>
          <article class="feature">
            <span class="tiny">02 Toolchain</span>
            <h3>Split by Range</h3>
            <p>Extract pages with selector syntax and parity filters.</p>
          </article>
          <article class="feature">
            <span class="tiny">03 Toolchain</span>
            <h3>Compression</h3>
            <p>Reduce size with standard optimization and image modes.</p>
          </article>
        </div>
      </article>

      <aside class="card stack fade-in delay-2">
        <div class="stat">
          <h3>Batch Merge</h3>
          <div class="meter"><span></span></div>
          <p style="margin-top:8px; font-size:13px;">3 of 5 files prepared</p>
        </div>
        <div class="stat">
          <h3>Server Health</h3>
          <p><span class="ok">Healthy</span> at /healthz</p>
          <p style="font-size:13px; margin-top:8px;">No cloud uploads. No telemetry by default.</p>
        </div>
        <div class="stat">
          <h3>Web Adapter</h3>
          <p>Route stubs are ready for merge, split, and compress APIs.</p>
          <p style="margin-top:8px;"><span class="warn">Preview mode</span> keeps this UI static.</p>
        </div>
      </aside>
    </section>
  </div>
</body>
</html>
`
