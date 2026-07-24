# Go for the Experienced Developer — Building an .io-Style Multiplayer Game Server

*A fast-track guide for someone who already knows C, C++, C#, Python, and Lua — and who does gamedev.*

The goal: **learn Go** by building a real-time multiplayer game server in the ".io" style — a simple shared-arena browser game (think Agar.io / Diep.io) where the server is authoritative, players connect over WebSocket, and the game logic is simple enough to live entirely in Go. No physics engine, no hand-rolled UDP, no cgo — every hard part that *isn't* Go is deliberately kept out, so your effort goes into learning the language.

**Why this project (and not the tank game):** The valuable, transferable thing in multiplayer servers is the *architecture* — goroutine-per-connection, channel-owned authoritative state, a fixed tick loop, snapshot broadcast, matchmaking. An .io-style game exercises all of that while keeping the *simulation* trivial (movement is `pos += velocity * dt`, collisions are circle overlaps). That means you spend your brain on **Go**, not on Box2D or netcode theory. It's also a genuine stepping stone: this exact architecture transfers to a physics-based tank game later (see the Phase 2 appendix), where you'd swap trivial movement for a headless game engine.

**What you'll build:** a top-down shooter arena. Each player is a circle that moves with WASD and shoots in the aim direction; bullets are circles with velocity; getting hit costs health / respawns; simple scoring. All movement and collision is plain arithmetic in Go. A tiny HTML5 canvas + JS page is the test client, so you *see it working* immediately without needing any engine.

**How to use this guide:**
- Part I (Ch.1-6) teaches Go itself, C/C++/C#-diff-framed. Type the exercises.
- Part II (Ch.7-13) builds the server incrementally, introducing each Go concept as the project needs it, plus the tiny browser test client.
- Transport is **WebSocket** via `coder/websocket` (the current recommended library; `gorilla/websocket` is archived).
- Messages are **JSON** (readable, debuggable, and perfect for a browser client). Simplicity over compactness — the point is learning Go, not optimizing bytes.

---

## Table of Contents

**Part I — The Go Language**
1. Setup & Mental Model
2. Types, Structs, and Pointers
3. Methods & Interfaces (Structural!)
4. Slices, Maps, and the (value, error) Idiom
5. Error Handling, defer, and JSON
6. Goroutines, Channels & select — The Heart

**Part II — Building the Game Server**
7. Project Setup & the WebSocket Echo Server
8. Many Players: Goroutine-per-Connection
9. The Game World: Channel-Owned State + Tick Loop
10. Game Logic: Movement, Shooting, Collisions
11. The Browser Test Client
12. Matchmaking & Rooms
13. Deployment

**Appendix**
- Phase 2: The Physics Tank Game (when you want to ship a real game)
- C/C++/C# -> Go Cheat Sheet

---

# Part I — The Go Language

## Chapter 1 — Setup & Mental Model

### Install
```
winget install GoLang.Go          # Windows
```
Verify `go version`. Install the official VS Code **Go** extension (wires up gopls, delve, formatting).

### Mental model vs. what you know

Go is closer to **C than C#/Java** in spirit, with GC, memory safety, and built-in concurrency.

| Concept | Go | Closest thing you know |
|---|---|---|
| Build | `go build` -> single static native binary | C++ but zero-config, instant |
| No classes | structs + methods + interfaces | C structs with methods |
| No inheritance | composition + interfaces only | (deliberate — forget hierarchies) |
| No exceptions | (value, error) returns | C error codes, ergonomic |
| Concurrency | goroutines + channels | green threads + message passing |
| GC | yes, low-latency | (unlike C++/Rust) |
| Pointers | *T / &x, no arithmetic | C pointers minus footguns |
| Formatting | gofmt — one true style | (no bikeshedding) |

The language is tiny; the depth is in idioms and the stdlib. For this project you'll lean on Go's **concurrency** (goroutine-per-player, channels) — which is exactly what Go is best at and what makes server code pleasant.

### Hello World
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go")
}
```
`go run main.go`.

### The unused-variable/import compile error
Go **won't compile** with unused variables or imports. Set up format-on-save now so it auto-fixes imports:
```jsonc
"[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": { "source.organizeImports": true }
}
```

### Exercise 1.1
Print 1-10 with a `for` loop. Go has only `for` — three forms: `for i := 0; i < 10; i++`, `for cond`, and bare `for {}` (infinite).

### Exercise 1.2
Print `os.Args`. `os.Args[0]` is the program name (like C).

---

## Chapter 2 — Types, Structs, and Pointers

### Basic types
`int`, sized ints, `uint` (Go has unsigned), `float32/64`, `bool`, `string`, `byte` (=uint8), `rune` (=int32). For this game you'll mostly use `float64` (positions/velocities) and `int` (health/score).

Declarations:
```go
var x int = 5
var y = 5
z := 5          // short form, inside functions — you'll use this most
const c = 3.14
```

### Structs — your game entities
```go
type Vec2 struct { X, Y float64 }

type Player struct {
    ID     string
    Pos    Vec2
    Vel    Vec2
    Angle  float64   // aim direction
    Health int
    Score  int
    Alive  bool
}

type Bullet struct {
    Owner string
    Pos   Vec2
    Vel   Vec2
}
```

**Capitalization = access control.** Uppercase = exported (public); lowercase = unexported (package-private). Applies to types, fields, functions, methods. (Bonus: JSON encoding only includes *exported* fields — Ch.5.)

### Pointers — C pointers, safety on
```go
func (p *Player) TakeDamage(d int) {
    p.Health -= d          // auto-deref, just . (no ->)
    if p.Health <= 0 { p.Alive = false }
}
```
No pointer arithmetic, no manual free (GC), `.` works on values and pointers. Passing a struct by value copies it; use `*Player` to mutate the original or avoid copying.

### Zero value
Every type has a zero value — no uninitialized memory. `var p Player` is fully zeroed and valid.

### Exercise 2.1
Define `Vec2` and free functions `Add`, `Sub`, `Scale(v, s)`, `Length`, `Normalize`. (No operator overloading in Go — you write these. You'll use them constantly for movement.)

### Exercise 2.2
Write `func (v Vec2) DistanceTo(o Vec2) float64`. You'll use it for collision checks (distance < sum of radii = overlap).

---

## Chapter 3 — Methods & Interfaces (Structural!)

### Methods — receiver before the name
```go
func (v Vec2) Length() float64 { return math.Sqrt(v.X*v.X + v.Y*v.Y) }
func (p *Player) Respawn(at Vec2) { p.Pos = at; p.Health = 100; p.Alive = true }
```
Value receiver (`v Vec2`) = copy, for small read-only-ish math types. Pointer receiver (`p *Player`) = mutate original / avoid copy. If any method needs a pointer receiver, make them all pointer receivers.

### Interfaces — implicit/structural satisfaction
The big mental shift from C#/Java: **no `implements` keyword.** If a type has the methods, it satisfies the interface automatically.
```go
type Entity interface {
    Update(dt float64)
    Position() Vec2
}
// Any type with Update + Position satisfies Entity. You never declare it.
```
Payoff: define interfaces at the point of use. "Accept interfaces, return structs." Keep interfaces small (often one method).

### Exercise 3.1
Give `Player` and `Bullet` both an `Update(dt float64)` (integrate position from velocity) and a `Position() Vec2`. Define the `Entity` interface. Write `func step(entities []Entity, dt float64)` that updates them all. Note you never declared the implements relationship — structural typing in action.

---

## Chapter 4 — Slices, Maps, and the (value, error) Idiom

### Slices — dynamic arrays
```go
var bullets []*Bullet
bullets = append(bullets, &Bullet{Owner: id})   // append returns new header — reassign!
for i, b := range bullets { _ = i; _ = b }       // range = foreach with index
for _, b := range bullets { b.Update(dt) }       // _ discards index
```
Always `s = append(s, x)`. You'll use a slice for bullets and a helper to remove dead ones (filter).

### Maps — your player table
```go
players := make(map[string]*Player)   // keyed by player ID
players[id] = &Player{ID: id}
p, ok := players[id]                   // comma-ok: ok=false if absent
delete(players, id)
```
**Maps are NOT concurrency-safe.** Concurrent read/write panics. This is why the game state lives in ONE goroutine (Ch.6/9) — the idiomatic Go fix, no mutexes.

### (value, error) idiom
```go
func parseInput(data []byte) (Input, error) {
    var in Input
    if err := json.Unmarshal(data, &in); err != nil {
        return Input{}, fmt.Errorf("bad input: %w", err)
    }
    return in, nil
}
in, err := parseInput(data)
if err != nil { return }
```
You'll type `if err != nil {` constantly. Explicit, visible failure.

### Exercise 4.1
Write `func removeDeadBullets(bullets []*Bullet) []*Bullet` returning a new slice with only live ones. This is your `.filter`/`.Where`, done with an explicit loop (idiomatic Go).

---

## Chapter 5 — Error Handling, defer, and JSON

### Errors are values
```go
if err != nil {
    return fmt.Errorf("loading arena: %w", err)   // %w wraps the underlying error
}
```
No exceptions for normal flow. `fmt.Errorf` with `%w` wraps an error, preserving the chain. `panic`/`recover` exist but are only for truly unrecoverable situations — NOT normal errors.

### defer — cleanup that always runs
```go
func handleConn(c *websocket.Conn) {
    defer c.CloseNow()      // runs when handleConn returns, no matter how
    // ...
}
```
`defer` schedules a call for when the surrounding function returns (like `finally`/RAII, but explicit). LIFO order. You'll use it for closing connections constantly.

### JSON — your wire format
Go's `encoding/json` maps structs <-> JSON using struct tags. Only **exported** fields are encoded.
```go
type Snapshot struct {
    Players []PlayerState `json:"players"`
    Bullets []BulletState `json:"bullets"`
    Tick    int           `json:"tick"`
}

data, err := json.Marshal(snap)          // struct -> JSON bytes
err = json.Unmarshal(data, &snap)        // JSON bytes -> struct
```
The `json:"players"` tag controls the JSON key name (lowercase is the JS convention your browser client wants). This is all you need — JSON is perfect for a browser client and readable while debugging.

### Exercise 5.1
Define `Input` (a JSON message from client: which keys are held, aim angle, firing bool) and `Snapshot` (server -> client world state). Round-trip both through `json.Marshal`/`Unmarshal` and confirm they survive. Add `json:"..."` tags with lowercase names.

---

## Chapter 6 — Goroutines, Channels & select — The Heart

**The reason Go suits this project.** Read twice.

### Goroutines
```go
go handlePlayer(conn)    // runs concurrently; the caller continues immediately
```
Cheap (start ~2KB), runtime-scheduled onto few OS threads. **One goroutine per connected player** is the core pattern — each blocks reading its player's socket; the runtime handles thousands cheaply.

### Channels — typed pipes between goroutines
```go
ch := make(chan Input)          // unbuffered: send blocks until receive
buffered := make(chan Input, 256)
ch <- in                        // send
in := <-ch                      // receive
```
Mantra: **"share memory by communicating."** Instead of shared game state guarded by locks, ONE goroutine owns the state, and player goroutines send it messages over channels. That owner processes them one at a time — no locks, no data races. This is the whole design:
- Each player goroutine reads network input and **sends** it on a channel to the game-world goroutine.
- The **game-world goroutine** owns all state, receives inputs one at a time, and applies them.

### select — the multiplexer
```go
for {
    select {
    case in := <-world.inputs:     // a player sent input
        world.applyInput(in)
    case p := <-world.join:        // a player joined
        world.addPlayer(p)
    case p := <-world.leave:       // a player left
        world.removePlayer(p)
    case <-ticker.C:               // time to advance the simulation
        world.step()
        world.broadcast()
    }
}
```
`select` blocks until one case can proceed. **This single loop is the skeleton of your entire game server** — it waits on "a player did something" or "time to tick" and handles whichever happens. Internalize it; Part II builds directly on it.

### Exercise 6.1
Spawn 5 goroutines, each sending its index on a shared channel; receive 5 in main and print. They arrive in any order — that's concurrency.

### Exercise 6.2
Build a tiny "world actor": a goroutine owning a `map[string]int` score table, receiving `addScore(id, n)` messages and `getScore(id)->reply` requests (reply via a channel embedded in the request struct). This is the owning-goroutine pattern — the exact shape of your game world.

### Exercise 6.3
A `select` loop that ticks every 50ms (20Hz, `time.NewTicker`) printing "tick", while also draining an `inputs` channel fed by a separate goroutine. Run a few seconds. **This is a stripped-down game loop** — you're most of the way to the server core.

---

# Part II — Building the Game Server

From here, one growing codebase. By the end: a playable browser multiplayer shooter arena.

## Chapter 7 — Project Setup & the WebSocket Echo Server

### Module setup
```
mkdir arena-server && cd arena-server
go mod init github.com/denis/arena-server
go get github.com/coder/websocket
```
`go mod init` creates `go.mod` (like `.csproj`/`package.json`). Layout:
```
arena-server/
├── go.mod
├── main.go
├── internal/game/     # world, players, bullets, tick loop
└── web/index.html     # the browser test client (Ch.11)
```

### The echo server — prove the pipe works
A WebSocket server that echoes JSON back. Uses `coder/websocket` (current recommended lib; gorilla is archived).
```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/coder/websocket"
    "github.com/coder/websocket/wsjson"
)

func main() {
    http.HandleFunc("/play", handlePlay)
    http.Handle("/", http.FileServer(http.Dir("web")))   // serve the test client
    log.Println("listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
    c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        InsecureSkipVerify: true,   // dev only: allow any origin; tighten for prod
    })
    if err != nil { return }
    defer c.CloseNow()

    for {
        ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
        var msg map[string]any
        err := wsjson.Read(ctx, c, &msg)
        cancel()
        if err != nil { return }        // client gone — end goroutine
        log.Printf("received: %v", msg)
        if err := wsjson.Write(r.Context(), c, msg); err != nil { return }
    }
}
```
`http.ListenAndServe` runs each connection handler in **its own goroutine** automatically — so `handlePlay` is already your per-player goroutine. You didn't even write `go`.

### Exercise 7.1
Run it. Connect from a browser console:
```js
const ws = new WebSocket("ws://localhost:8080/play");
ws.onmessage = e => console.log("got", e.data);
ws.onopen = () => ws.send(JSON.stringify({hello: "world"}));
```
Confirm the message bounces back. Milestone: **the pipe works.**

---

## Chapter 8 — Many Players: Goroutine-per-Connection

Each player gets a goroutine that reads their input, plus an outbound `send` channel drained by a write goroutine.

```go
type Player struct {
    ID   string
    conn *websocket.Conn
    send chan []byte        // outbound queue; the world pushes snapshots here
    // game state (Pos, Vel, Health, ...) lives in the world's map, not here
}
```

**Why a `send` channel + separate write goroutine:** the world goroutine can't safely block on `conn.Write` for a slow client — it would stall the whole simulation. So each player has an outbound `send` channel and a dedicated write goroutine that drains it to the socket. Two goroutines per connection: one reading, one writing.

```go
func (p *Player) writeLoop(ctx context.Context) {
    for data := range p.send {          // ranges until send is closed
        if err := p.conn.Write(ctx, websocket.MessageText, data); err != nil {
            return
        }
    }
}

func (p *Player) readLoop(ctx context.Context, world *World) {
    world.join <- p
    defer func() { world.leave <- p }()
    for {
        _, data, err := p.conn.Read(ctx)
        if err != nil { return }
        var in Input
        if json.Unmarshal(data, &in) == nil {
            in.PlayerID = p.ID
            world.inputs <- in           // hand off to the owning goroutine
        }
    }
}
```

### Exercise 8.1
Assign each player a unique ID (a random string or counter), spawn read + write goroutines, log join/leave. Connect two browser tabs; confirm both are handled independently.

---

## Chapter 9 — The Game World: Channel-Owned State + Tick Loop

The core: **one goroutine owns all game state.** All input funnels in over channels. No mutexes ever.

```go
type World struct {
    players map[string]*Player       // connection wrappers
    tanks   map[string]*PlayerState  // authoritative game state
    bullets []*Bullet
    inputs  chan Input
    join    chan *Player
    leave   chan *Player
    tick    int
}

func (w *World) run() {
    ticker := time.NewTicker(time.Second / 20)   // 20Hz simulation
    defer ticker.Stop()
    for {
        select {
        case p := <-w.join:
            w.players[p.ID] = p
            w.tanks[p.ID] = spawnState(p.ID)
        case p := <-w.leave:
            delete(w.players, p.ID)
            delete(w.tanks, p.ID)
        case in := <-w.inputs:
            w.applyInput(in)          // store this player's current input
        case <-ticker.C:
            w.step()                  // advance simulation one tick
            w.broadcast()             // push snapshot to every player.send
            w.tick++
        }
    }
}
```

Because ONLY this goroutine touches `players`/`tanks`/`bullets`, there are **no data races and no locks**. Player goroutines just drop inputs on `w.inputs`. This is Go's "share memory by communicating," and it's dramatically simpler than the lock-heavy shared-state model you'd write in C++/C#.

```go
func (w *World) broadcast() {
    snap := w.buildSnapshot()          // a Snapshot struct (Ch.5)
    data, _ := json.Marshal(snap)
    for _, p := range w.players {
        select {
        case p.send <- data:           // enqueue
        default:                       // slow client: drop this snapshot, don't stall
        }
    }
}
```
The `select/default` drops a snapshot for a backed-up client rather than blocking the simulation — for real-time state, the next snapshot supersedes it anyway.

### Exercise 9.1
Implement `World` with `run()`, `join`/`leave`/`inputs` channels, a 20Hz tick, and snapshot broadcast (players just sit still for now). Connect two tabs; confirm both receive a snapshot stream 20x/sec. Milestone: **the world is alive and streaming.**

---

## Chapter 10 — Game Logic: Movement, Shooting, Collisions

Now make it a game. All of this lives in the world goroutine — single owner, zero concurrency concerns.

### Input as state
```go
type Input struct {
    PlayerID string  `json:"-"`       // set server-side, not from client
    Up, Down, Left, Right bool        `json:"up","down","left","right"`
    Angle    float64                  `json:"angle"`   // aim direction
    Firing   bool                     `json:"firing"`
}
```
`applyInput` stores the player's current input flags; `step()` reads them each tick. Decouples input rate from tick rate.

### step(): movement, firing, bullets, collisions
```go
func (w *World) step() {
    const dt = 1.0 / 20.0
    const speed = 200.0     // units/sec

    for id, in := range w.currentInputs {
        ps := w.tanks[id]
        if ps == nil || !ps.Alive { continue }
        // movement: build a direction from held keys
        var dir Vec2
        if in.Up    { dir.Y -= 1 }
        if in.Down  { dir.Y += 1 }
        if in.Left  { dir.X -= 1 }
        if in.Right { dir.X += 1 }
        dir = dir.Normalize()
        ps.Pos = ps.Pos.Add(dir.Scale(speed * dt))
        ps.Angle = in.Angle
        // firing (with reload)
        if in.Firing && w.tick - ps.LastShotTick > reloadTicks {
            ps.LastShotTick = w.tick
            w.bullets = append(w.bullets, &Bullet{
                Owner: id,
                Pos:   ps.Pos,
                Vel:   Vec2{math.Cos(in.Angle), math.Sin(in.Angle)}.Scale(bulletSpeed),
            })
        }
    }

    // integrate bullets
    for _, b := range w.bullets {
        b.Pos = b.Pos.Add(b.Vel.Scale(dt))
    }
    // collisions: bullet vs player (circle overlap)
    for _, b := range w.bullets {
        for id, ps := range w.tanks {
            if id == b.Owner || !ps.Alive { continue }
            if b.Pos.DistanceTo(ps.Pos) < playerRadius {
                ps.TakeDamage(25)
                b.Dead = true
                if !ps.Alive { w.tanks[b.Owner].Score++ }
            }
        }
    }
    w.bullets = removeDeadBullets(w.bullets)
    // clamp to arena bounds, respawn dead players after a delay, etc.
}
```
Pure arithmetic — no physics engine, all native Go. This is the whole point: the *architecture* is real multiplayer, but the *simulation* is simple enough to be pure Go.

### Exercise 10.1
Implement movement, firing with reload, bullet integration, circle collision, health, scoring, arena bounds, and respawn. Test by driving with two browser tabs (send inputs from the console or wait for Ch.11's client). Milestone: **it's a game.**

---

## Chapter 11 — The Browser Test Client

A tiny HTML5 canvas + JS page so you can actually play. Put it in `web/index.html` (served by the file server from Ch.7). This is NOT Go — it's the throwaway client so you see your server working.

```html
<!DOCTYPE html>
<html>
<body style="margin:0">
<canvas id="c" width="800" height="600" style="background:#111"></canvas>
<script>
const ws = new WebSocket("ws://localhost:8080/play");
const ctx = document.getElementById("c").getContext("2d");
let world = { players: [], bullets: [] };
const keys = {};

ws.onmessage = e => { world = JSON.parse(e.data); };
addEventListener("keydown", e => { keys[e.key] = true; });
addEventListener("keyup",   e => { keys[e.key] = false; });

let angle = 0;
addEventListener("mousemove", e => {
    angle = Math.atan2(e.clientY - 300, e.clientX - 400);
});

// send input ~30x/sec
setInterval(() => {
    if (ws.readyState !== 1) return;
    ws.send(JSON.stringify({
        up: !!keys["w"], down: !!keys["s"],
        left: !!keys["a"], right: !!keys["d"],
        angle: angle, firing: !!keys[" "],
    }));
}, 33);

// render loop
function draw() {
    ctx.clearRect(0, 0, 800, 600);
    for (const p of world.players) {
        ctx.fillStyle = "#4af";
        ctx.beginPath(); ctx.arc(p.x, p.y, 15, 0, 7); ctx.fill();
    }
    for (const b of world.bullets) {
        ctx.fillStyle = "#fa4";
        ctx.beginPath(); ctx.arc(b.x, b.y, 4, 0, 7); ctx.fill();
    }
    requestAnimationFrame(draw);
}
draw();
</script>
</body>
</html>
```
Open `http://localhost:8080` in two tabs and play against yourself. (Adjust field names to match your JSON tags.) For smoother motion later you could interpolate between snapshots, but at 20Hz on localhost it's already fine.

### Exercise 11.1
Wire the client to your server. Two tabs, drive around, shoot each other, watch health/score. Milestone: **a playable browser multiplayer game.**

---

## Chapter 12 — Matchmaking & Rooms

One shared arena works; now support multiple concurrent games. A **manager goroutine** owns the set of worlds and routes players.

```go
type Manager struct {
    join  chan *Player
    rooms map[int]*World
    next  int
}

func (m *Manager) run() {
    var open *World
    for {
        p := <-m.join
        if open == nil || open.full() {
            open = m.newWorld()
            go open.run()                 // each world is its own goroutine
            m.rooms[open.ID] = open
        }
        open.join <- p
        if open.full() { open = nil }
    }
}
```
Each world runs independently on its own goroutine with its own tick and state. The manager just routes. This scales naturally to many concurrent matches — and it's remarkably little code. **This is Go flexing exactly what it's built for.**

### Exercise 12.1
Add the manager; route players into arenas of capacity 4; spin up a new world when one fills. Connect enough tabs to force a second arena; confirm two independent games tick concurrently.

---

## Chapter 13 — Deployment

A Go server is the easiest thing to deploy: one static binary, no runtime.

### Cross-compile for Linux
```
GOOS=linux GOARCH=amd64 go build -o arena-server .
```
This produces a single Linux binary from your Windows machine — no dependencies to install on the server.

### Host on a cheap VPS
A $5/month VPS (Hetzner, DigitalOcean, Vultr, Linode) is ideal, and it's the same Linux-admin skillset you already use for Apache/SVN:
1. `scp arena-server` and your `web/` folder to the box.
2. Run it under `systemd` so it restarts on crash / boot.
3. Open the port (e.g. 8080, or put it behind a reverse proxy on 80/443).

Minimal `systemd` unit (`/etc/systemd/system/arena.service`):
```ini
[Unit]
Description=Arena Server
After=network.target

[Service]
ExecStart=/opt/arena/arena-server
WorkingDirectory=/opt/arena
Restart=always
User=arena

[Install]
WantedBy=multi-user.target
```
`sudo systemctl enable --now arena`. Since WebSocket rides on HTTP/TCP, it works through normal firewalls and behind reverse proxies (nginx/Caddy) that can also terminate TLS — no special networking. (This is a nice contrast to raw UDP, which needs a real IP + open UDP port; WebSocket "just works" through standard web infrastructure.)

### Exercise 13.1
Cross-compile, deploy to a VPS (or test locally first with the binary), and connect from another machine/browser over the internet. Milestone: **shipped.**

---

# Appendix — Phase 2: The Physics Tank Game

*When your goal shifts from "learn Go" to "ship a real physics-based game," the tank arena you actually want is a different project with a different architecture. This note captures why, so future-you starts in the right place.*

**The key realization:** a tank game built on Box2D (capsule solver, collision/trigger events) needs the *same physics engine* deciding authoritative truth — you can't reimplement Box2D's solver in Go, and even if you did it wouldn't match. That forces one of two coherent architectures:

1. **Headless game-engine server (recommended, the "Godot model").** The server is your game engine (Godot, or a headless build of Hob2D) running the simulation with no window. Physics/gameplay written once, run natively on both sides; authority is automatic; no cross-language physics, no cgo. This is how engines with built-in multiplayer (Godot/Unreal/Unity dedicated servers) work, and it's the elegant answer for physics-heavy games. Cost: the server speaks the engine's language (C++ for headless Hob2D; GDScript/C# for Godot), so it's not a Go-learning project.

2. **Go server + real Box2D via cgo.** Keep the Go networking/RPC/room code, but link the actual Box2D 3.x (which is now written in C, making cgo bindings clean) and step a `b2World` each tick. One physics implementation (real Box2D) on both sides; authority preserved. Cost: cgo build complexity, cross-compile friction, and some loss of Go's clean-static-binary simplicity. Viable, but you're swimming somewhat upstream.

If the tank game is about *shipping the game*, Path 1 (headless engine) is almost certainly right. If it's about *also* learning Go at a deep systems level and you accept the friction, Path 2 works.

**What transfers from this guide:** the authoritative-server architecture — goroutine/actor-per-connection, channel-owned (or engine-owned) state, fixed tick loop, snapshot broadcast, matchmaking, interpolation/prediction concepts (see NETWORKING.md). You learned all of it here on easy mode; in Phase 2 you swap "trivial Go movement" for "engine physics" and (optionally) "WebSocket" for a lower-level transport. The skeleton is the same.

**Prerequisite reading for Phase 2:** NETWORKING.md Part III (authoritative servers, snapshots, interpolation, prediction, lag compensation) — all still applies, now with a real physics simulation underneath.

---

## Appendix — C/C++/C# -> Go Cheat Sheet

| You know | Go |
|---|---|
| class | struct + methods (no inheritance) |
| public/private | Uppercase / lowercase |
| inheritance | embedding + interfaces |
| interface (explicit implement) | interface (structural, implicit) |
| vector<T> / List<T> | []T (slice) |
| map / Dictionary | map[K]V (not concurrency-safe) |
| exceptions | (value, error) + if err != nil |
| finally / RAII | defer |
| throw (fatal) | panic (rare) |
| threads + locks | goroutines + channels |
| Task / async (C#) | goroutines (write blocking code) |
| -> and . | just . |
| pointer arithmetic | none |
| nullptr / null | nil |
| new / delete | GC; make / &T{} |
| null check | comma-ok: v, ok := m[k] |
| operator overloading | none (write Add, Sub, Scale) |
| JSON lib | encoding/json (struct tags, exported fields only) |
| namespaces | packages (one per directory) |
| #include | import (packages) |
| .csproj | go.mod |

---

*Work Part I in order over a few evenings — the language is small. Then Part II is the project, milestone by milestone; you'll have a playable browser game by Ch.11 and it deployed by Ch.13. Keep the tank/physics ambition for Phase 2, when shipping a game — not learning a language — is the goal.*
