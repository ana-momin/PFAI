<div align="center">

# Advanced Go Systems

### Three production-minded Go projects built by Anas at NUTECH

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Portfolio](https://img.shields.io/badge/Portfolio-Live-46E6A7?style=for-the-badge)](https://pfai-go-lab.vercel.app)
[![NUTECH](https://img.shields.io/badge/NUTECH-University_for_Industry-0F5F50?style=for-the-badge)](https://nutech.edu.pk/)

**Realtime communication. Secure APIs. Coordinated concurrency.**

[Explore Portfolio](https://pfai-go-lab.vercel.app) · [Signal Room](https://pfai-signal-chat.onrender.com) · [Keyline](https://pfai-keyline.vercel.app) · [Orbit](https://pfai-orbit.vercel.app)

</div>

---

## Overview

This repository contains three complete Advanced Go practice projects and a cinematic portfolio website that presents them. Every application has a real Go backend, an embedded responsive frontend, and a distinct visual identity.

| Project | Focus | Live Demo |
| --- | --- | --- |
| **7.1 Signal Room** | Gorilla WebSocket chat, goroutines, channels, presence, and message broadcasting | [Open Signal Room](https://pfai-signal-chat.onrender.com) |
| **7.2 Keyline Workspace** | Gin REST API, JWT middleware, bcrypt passwords, and per-user task CRUD | [Open Keyline](https://pfai-keyline.vercel.app) |
| **7.3 Orbit Crawler** | Concurrent worker pool, token-bucket rate limiting, cancellation, and SSRF protection | [Launch Orbit](https://pfai-orbit.vercel.app) |
| **Portfolio** | Interactive project showcase with motion, parallax, and NUTECH attribution | [View Portfolio](https://pfai-go-lab.vercel.app) |

## System Highlights

### Signal Room

- Persistent Gorilla WebSocket connections
- Goroutine-powered broadcast hub
- Live user presence and system events
- Bounded message history and heartbeat pings
- Automatic client reconnection

### Keyline Workspace

- Registration and login with bcrypt password hashing
- Signed 24-hour JWT access tokens
- Gin authentication middleware
- Isolated task resources for each user
- Complete create, read, update, and delete workflow

### Orbit Crawler

- Bounded concurrent worker pool
- Token-bucket request rate limiting
- Same-origin crawl discovery
- Context cancellation and request timeouts
- Private-network and redirect protections
- Live crawl telemetry interface

## Architecture

```text
PFAI/
├── chat/          Gorilla WebSocket server + embedded Signal Room UI
├── auth-api/      Gin JWT REST API + embedded Keyline UI
├── crawler/       Concurrent crawler + embedded Orbit UI
├── portfolio/     Interactive project showcase
├── render.yaml    Render deployment configuration
└── run-all.ps1    Local launcher for all Go services
```

Each Go application embeds its frontend assets into the compiled binary, making it portable and easy to deploy.

## Run Locally

Requirements:

- Go 1.24 or newer
- PowerShell for the optional combined launcher

Launch everything:

```powershell
.\run-all.ps1
```

Or run each application independently:

```powershell
cd chat
go run .
# http://localhost:8081

cd ../auth-api
go run .
# http://localhost:8082

cd ../crawler
go run .
# http://localhost:8083
```

Serve the portfolio with any static HTTP server:

```powershell
cd portfolio
python -m http.server 4173
```

## Verification

```powershell
cd chat; go test ./...; go build ./...
cd ../auth-api; go test ./...; go build ./...
cd ../crawler; go test ./...; go build ./...
```

The production release was verified with:

- Two-client public WebSocket message delivery
- JWT registration, authentication, and secured task CRUD
- Concurrent crawling against public websites
- Responsive desktop and mobile portfolio rendering

## Deployment

| Service | Platform | URL |
| --- | --- | --- |
| Portfolio | Vercel | https://pfai-go-lab.vercel.app |
| Signal Room | Render | https://pfai-signal-chat.onrender.com |
| Keyline Workspace | Vercel Go Runtime | https://pfai-keyline.vercel.app |
| Orbit Crawler | Vercel Go Runtime | https://pfai-orbit.vercel.app |

Signal Room runs on Render because WebSocket servers require a persistent process. Keyline currently uses an in-memory course-project store, so production data can reset when a Vercel instance is replaced.

## Author

**Anas**  
National University of Technology, Islamabad  
Advanced Go Practice · 2026

---

<div align="center">
Built with Go, careful engineering, and a strong preference for excellent visual design.
</div>
