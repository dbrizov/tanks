# Tanks — Build Plan & Learning Roadmap

A 2D, browser-only tank arena shooter, built as a three-tier system to learn **Go**,
**Java/Spring Boot**, **Godot**, and **Kubernetes**. This document is a *roadmap*: it says
**what** to build, **why**, **what to learn**, and **how to know each step is done**. You write
every line yourself — the snippets here are teaching illustrations, not code to paste wholesale.

> **How to use this doc:** work milestone by milestone. Each has *Learning goals*, *What to build*,
> *Research pointers* (into your own notes in `docs/`), and a *Done when* checklist. Don't skip the
> "Web-target constraints" box — those rules shape every tier.

---

## 0. Overview & architecture

**The game:** enter an arena, drive a tank, shoot other tanks. Real-time, multiplayer, 2D.

**Three tiers:**

```
                    ┌─────────────────────────┐
   ┌── HTTPS ──────▶│  auth_java (Spring Boot) │  issues JWT, owns user identity
   │  login/register└─────────┬───────────────┘        │ PostgreSQL
   │                          │ JWT                     ▼
┌──┴───────────────┐          │                   ┌─────────┐
│  client_godot    │          │                   │Postgres │
│  (browser, thin) │          ▼                   └─────────┘
│  - captures input│   ┌──────────────────────────┐
│  - renders state │◀─▶│  server_go (authoritative │  owns 100% of the simulation
└──────────────────┘   │  arena server, WebSocket) │  (movement, collisions, hits)
        ▲   WSS         └──────────────────────────┘  stateless (match state in memory)
        │
   GitHub Pages (static hosting)          Backend runs on Oracle Cloud VM + k3s
```

**Data flow:** `client --HTTPS(login)--> auth (get JWT)` then
`client --WSS(JWT)--> Go server (play)`.

**The core principle — server-authoritative:** the **Go server owns the entire simulation**.
The Godot client is a *thin client*: it sends **input intents** ("I'm holding forward", "I fired")
and **renders the state the server sends back**. The client contains **no game logic and no
physics**. This prevents cheating (the client can't lie about positions/hits) and gives a single
source of truth. It also means the interesting engineering lives in Go — exactly what you want to
learn.

**Monorepo layout:**

```
tanks/
  client_godot/   Godot 4.7 project (2D, single-threaded Web export)
  server_go/      Go authoritative arena server (WebSocket)
  auth_java/      Spring Boot auth service (JWT issuer, Postgres)
  deploy/         Kubernetes (k3s) manifests
  docs/           Your learning notes (GO.md, JAVA.md, NETWORKING.md, SPRINGBOOT.md)
  PLAN.md         This file
```

---

## 1. Web-target constraints (read this first — these rules shape everything)

1. **Renderer must be `GL Compatibility`** (WebGL2). The project was created with `Forward Plus`
   + `d3d12`, which are desktop-only and won't serve a broad-browser web build. Change this in
   Project Settings → Rendering → Renderer.
2. **Fully server-authoritative.** All simulation — movement, collision, projectiles, hit
   detection, arena bounds — runs in the **Go server**. The Godot client runs **no physics**, so
   Godot's physics engine is unused and the `Jolt Physics` (3D) project setting is irrelevant.
   The server needs **no physics engine either** — hand-rolled 2D vector math (circle/AABB
   collision, velocity integration) on a fixed tick is plenty for a tank shooter.
3. **No raw sockets in the browser.** The client can't open UDP/TCP. It uses Godot's
   `WebSocketPeer`; the server speaks **WebSocket (WSS in production)**.
4. **Single-threaded Web export** (Godot 4.3+). This avoids needing `SharedArrayBuffer` /
   COOP-COEP headers, which is what lets the client live on **GitHub Pages** (Pages can't set
   custom HTTP headers). Single-threaded is fine for a 2D shooter.
5. **Secure transport from an HTTPS page.** Because the client is served over HTTPS (GitHub Pages),
   it must connect to the backend over **`wss://`** with a *valid* certificate. Plain `ws://`, a
   raw IP, or a self-signed cert is blocked as mixed content. (WebSocket is exempt from CORS, so
   the cross-origin Pages→backend connection itself is fine.)

---

## 2. Milestone 0 — Foundations & repo setup

**Learning goals:** get each toolchain installed and the repo organized; understand the Godot
project layout.

**What to build / do:**
- The monorepo folders already exist (`client_godot/ server_go/ auth_java/ deploy/ docs/`) and the
  Godot project + notes have been moved. Confirm the Godot project opens from `client_godot/`.
- `git init` at the repo root (this isn't a git repo yet). Sanity-check `.gitignore` covers
  `.godot/`, Go build output, and Java `target/`.
- Install toolchains and **write the versions down** (in this file or a `docs/TOOLCHAIN.md`):
  Godot 4.7, a JDK (21 LTS) + Maven/Gradle, Go (1.22+), Docker, `kubectl`.
- In Godot: set renderer to **Compatibility** (see constraint #1). The physics setting is moot.

**Done when:**
- `client_godot/` opens cleanly in Godot with the Compatibility renderer selected.
- `git status` works at the root; toolchain versions are recorded.

> **Build-order rationale:** you build the *game* (Go server + Godot client) first and get it
> playable in the browser, so the risky real-time netcode is proven early. **Authentication is
> bolted on later.** To make that cheap, design the Go server's connection handshake with an
> **auth seam** from day one — anonymous now, but a single choke point where JWT validation drops
> in later without touching the game loop.

---

## 3. Milestone 1 — Game server (Go), no auth yet — *owns the whole simulation*

**Learning goals:** Go concurrency (goroutines, channels, `select`), WebSocket handling, a
fixed-tick game loop, hand-rolled 2D simulation, broadcasting authoritative state.
→ Research: `docs/GO.md`, `docs/NETWORKING.md`.

**What to build (incremental — get each step working before the next):**
1. **WS echo server.** Accept a WebSocket upgrade, echo messages back. (Look at `nhooyr.io/websocket`
   or `gorilla/websocket`.) Test with a CLI like `wscat`.
2. **Connection registry + auth seam.** A hub that tracks connected clients. Route the accept
   through a single function that *currently* lets everyone in — this is the seam:
   ```go
   // authenticate is the ONE place auth will later live. For now it accepts anyone.
   func authenticate(r *http.Request) (PlayerID, error) {
       return PlayerID(randomID()), nil // M6 replaces this with JWT validation
   }
   ```
3. **Fixed tick loop.** A `time.Ticker` at e.g. 30 Hz. Each tick: drain queued inputs → advance the
   simulation → snapshot → broadcast.
   ```go
   tick := time.NewTicker(time.Second / 30)
   for range tick.C {
       world.applyInputs(inbox)     // consume intents since last tick
       world.step(1.0 / 30.0)       // integrate movement, resolve collisions, clamp to arena
       hub.broadcast(world.snapshot())
   }
   ```
4. **Input intents in, world state out.** Define a tiny message protocol (JSON is fine to start;
   you can move to a binary format later). Intents: movement axis + fire flag. Snapshot: list of
   tanks `{id, x, y, angle}`.
5. **Simulation.** Hand-rolled 2D: integrate velocity from intent, resolve tank/tank and
   tank/wall collisions with circle or AABB math, clamp to arena bounds. **No physics engine.**

**Concepts to research:** goroutine-per-connection vs. a single hub goroutine + channels (the hub
pattern avoids data races on world state); why the world should be mutated by **one** goroutine.

**Done when:** several `wscat`/browser clients connect **anonymously**, send input, and the server
ticks, simulates movement + collision, and broadcasts authoritative snapshots at a fixed rate.

---

## 4. Milestone 2 — Godot web client (2D, thin client), anonymous connect

**Learning goals:** Godot scenes/nodes, capturing input, `WebSocketPeer`, rendering server state
with no local simulation, single-threaded HTML5 export.

**What to build:**
- A `Main` scene that opens a `WebSocketPeer` to the Go server (no login yet).
- **Send input intents** each frame/tick (movement axis, fire) — do **not** move the tank locally.
- **Render from server state only:** on each snapshot, spawn/update tank sprites at the positions
  the server sent. The tank you control moves *because the server said so*, not because you pressed
  a key.
  ```gdscript
  func _process(_delta):
      var intent := {"ax": Input.get_axis("left","right"), "ay": Input.get_axis("up","down"),
                      "fire": Input.is_action_pressed("fire")}
      _socket.send_text(JSON.stringify(intent))
      while _socket.get_available_packet_count() > 0:
          _apply_snapshot(JSON.parse_string(_socket.get_packet().get_string_from_utf8()))
  ```
- Export to **Web (single-threaded)** and run the exported build in a browser locally (serve the
  export dir over http, e.g. `python -m http.server`, then open it).

**Note:** interpolation/prediction is a *rendering* concern deferred to M4. The client never
becomes a source of simulation truth.

**Done when:** the exported web build connects, sends input, and renders a tank that moves only in
response to server snapshots.

---

## 5. Milestone 3 — Vertical slice (game only, integration checkpoint)

**What to build:** wire client + server end to end, still no auth. Two browser tabs connect; each
sees the other's tank move under the server's authoritative simulation.

**Why this milestone matters:** it de-risks the hardest part (real-time netcode in a browser)
before you invest in gameplay or auth. If two tabs can see each other move smoothly, the
architecture is sound.

**Done when:** two browser tabs both render two tanks, and driving one is visible in the other.

---

## 6. Milestone 4 — Gameplay (all simulated in the Go server)

**What to build (all server-side):** shooting + projectiles, hit detection, health/respawn,
scoring, arena bounds. The client only renders results and plays effects (muzzle flash, hit spark).

**Optional netcode polish (render side only):**
- **Interpolation:** render between the last two server snapshots for smooth motion.
- **Client prediction + reconciliation** (optional, more advanced): predict your own tank locally
  for responsiveness, then correct when the server snapshot arrives. The server remains the sole
  authority. → `docs/NETWORKING.md`.
- WebRTC DataChannel (UDP-like, lower latency) is a *future* upgrade only — WebSocket first.

**Done when:** a full round is playable between two browsers, with every outcome decided
server-side.

---

## 7. Milestone 5 — Auth service (Spring Boot / Java)

**Learning goals:** Spring Boot basics, REST controllers, **Spring Data JPA + PostgreSQL**, JWT
issuance/validation, password hashing (BCrypt).
→ Research: `docs/JAVA.md`, `docs/SPRINGBOOT.md`.

**What to build:**
- Endpoints: `POST /register`, `POST /login` → returns a **signed JWT**; `GET /health`.
- A `User` JPA entity persisted to the database; hash passwords with BCrypt (never store plaintext).
  ```java
  @Entity
  class User {
      @Id @GeneratedValue Long id;
      @Column(unique = true) String username;
      String passwordHash;   // BCrypt
  }
  ```
- **Dev against H2 in-memory**, **Postgres for deploy** — same JPA code, switched by a Spring
  profile / datasource config, so you learn JPA without fighting infra early:
  ```
  application-dev.properties   -> H2 in-memory
  application-prod.properties  -> PostgreSQL
  ```

**Done when:** register persists a user; login returns a verifiable JWT (signature + claims check
out); the same code runs on H2 (dev profile) and Postgres (prod profile).

---

## 8. Milestone 6 — Wire auth into the game (fill the seam)

**What to build:**
- **Client:** add a login screen → call the auth service → obtain a JWT → send it on the WS
  handshake (e.g. as a query param or first message).
- **Server:** validate the JWT at the **auth seam** built in M1 (verify signature with the auth
  service's secret / JWKS), reject unauthenticated connections, and attach the player identity to
  the connection. The rest of the game loop is untouched — that's the payoff of the seam.
  ```go
  func authenticate(r *http.Request) (PlayerID, error) {
      claims, err := verifyJWT(r.URL.Query().Get("token")) // was: accept anyone
      if err != nil { return "", err }
      return PlayerID(claims.Subject), nil
  }
  ```

**Done when:** only authenticated players can join; identity flows login → server → gameplay
(e.g. the tank is labeled with the username from the token).

---

## 9. Milestone 7 — Deploy (client → GitHub Pages; backend → Oracle Cloud + k3s)

**Client (GitHub Pages):**
- Publish the **single-threaded** web export as static files (a GitHub Action on push, or a
  `gh-pages` branch). Free HTTPS on `*.github.io` (or a custom domain). No server, no Cloudflare
  needed for the client.
- Point the client's WebSocket URL at the backend's `wss://<domain>`.

**Backend (Oracle ARM VM + k3s):**
- *Learning goals:* Dockerfiles per service (**arm64**), Kubernetes objects (Deployment, Service,
  Ingress; **StatefulSet + PVC** for Postgres), provisioning an Oracle always-free ARM VM, k3s,
  Traefik ingress, TLS.
- Provision the Oracle **always-free ARM (aarch64)** Ubuntu VM. **Two gotchas:** build container
  images for **arm64**, and open ports 80/443 in **both** the Oracle **VCN security list** *and*
  the **instance iptables** (Oracle images ship restrictive iptables).
- Install **k3s** (single-binary Kubernetes, Traefik ingress built in). Dockerize `server_go` and
  `auth_java` as arm64 images. Deploy a **Postgres** StatefulSet with a PVC for the auth service.
  Write manifests in `deploy/`. Expose via Traefik Ingress on 443.
- **Domain + TLS is the only hard requirement:** an A record `<domain>` → the VM's public IP, plus
  a valid cert. Either **Cloudflare** (free DNS + edge TLS / origin cert; also hides the origin IP)
  **or** any DNS provider + **Let's Encrypt** via Traefik/cert-manager. Cloudflare is convenience,
  not a requirement.

**Done when:** the GitHub Pages client loads and connects over `wss://<domain>` to the Oracle
backend from an outside network, and the game plays.

> **Optional early deploy:** the game-only stack (client on Pages + Go server on the VM, no auth)
> can go live right after M4 to test real online play sooner; add the auth service + Postgres at
> M6. Deferring all deploy to here is fine too.

---

## 10. Appendix

**Glossary (terms new in this stack):**
- **Authoritative server** — the server, not the client, decides game state; clients only send
  input and render results.
- **Fixed tick** — the server advances the simulation at a constant rate (e.g. 30 Hz) regardless of
  render framerate.
- **JWT** — a signed token proving identity; the game server trusts it without calling the auth DB.
- **JWKS** — a published set of public keys used to verify JWT signatures.
- **k3s** — a lightweight, single-binary Kubernetes distribution.
- **Ingress / Traefik** — routes external HTTPS/WSS traffic to services inside the cluster.
- **PVC / StatefulSet** — Kubernetes objects for persistent storage and stateful pods (Postgres).
- **COOP/COEP** — headers enabling `SharedArrayBuffer`; needed only for *threaded* web exports
  (which we avoid).

**Where each note applies:** `docs/GO.md` → M1; `docs/NETWORKING.md` → M1, M3, M4;
`docs/JAVA.md` + `docs/SPRINGBOOT.md` → M5.

**Open questions to ask Claude (grow this list as you go):**
- _(add yours here — e.g. "best message format for snapshots?", "how to structure the Go hub?")_
