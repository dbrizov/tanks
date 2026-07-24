# Networking Fundamentals — From Ports to Game Netcode

*A language-neutral primer, starting from near-zero, building toward the netcode you'll need for a real-time multiplayer game.*

This guide assumes you're a strong programmer but haven't done low-level networking. It starts from "what is a port" and ends at "here's how authoritative game servers, snapshots, interpolation, and lag compensation actually work." No specific language — these concepts are the same in Go, C++, C#, Rust, Python. Where it helps, examples use pseudocode.

**The arc:** addresses & ports → the layered model → TCP vs UDP → sockets (the API) → framing & serialization → the real-time problem → authoritative servers → client prediction & reconciliation → interpolation → lag compensation. By the end you'll understand every term in the tank-server guide and *why* each choice was made.

---

## Table of Contents

**Part I — How the Internet Actually Moves Bytes**
1. [Addresses, Ports, and Sockets](#chapter-1--addresses-ports-and-sockets)
2. [The Layered Model](#chapter-2--the-layered-model)
3. [TCP: The Reliable Stream](#chapter-3--tcp-the-reliable-stream)
4. [UDP: The Raw Datagram](#chapter-4--udp-the-raw-datagram)
5. [TCP vs UDP: The Decision](#chapter-5--tcp-vs-udp-the-decision)

**Part II — Programming Against the Network**
6. [The Sockets API](#chapter-6--the-sockets-api)
7. [Framing: Where Does a Message End?](#chapter-7--framing-where-does-a-message-end)
8. [Serialization: Structs to Bytes](#chapter-8--serialization-structs-to-bytes)
9. [Higher-Level Protocols: HTTP, WebSocket, QUIC](#chapter-9--higher-level-protocols-http-websocket-quic)

**Part III — The Real-Time Problem (Game Netcode)**
10. [Why Real-Time Is Hard: Latency, Jitter, Loss](#chapter-10--why-real-time-is-hard-latency-jitter-loss)
11. [Authoritative Servers](#chapter-11--authoritative-servers)
12. [Snapshots & Interpolation](#chapter-12--snapshots--interpolation)
13. [Client-Side Prediction & Reconciliation](#chapter-13--client-side-prediction--reconciliation)
14. [Lag Compensation & Putting It Together](#chapter-14--lag-compensation--putting-it-together)

---

# Part I — How the Internet Actually Moves Bytes

## Chapter 1 — Addresses, Ports, and Sockets

### IP address — *which machine*

An **IP address** identifies a machine on a network. Two flavors:
- **IPv4**: four bytes, written `192.168.1.10`. ~4 billion addresses (we ran out; hence IPv6).
- **IPv6**: sixteen bytes, written `2001:db8::1`. Effectively unlimited.

Special addresses you'll meet constantly:
- `127.0.0.1` (IPv4) / `::1` (IPv6) — **localhost**, "this same machine." Your server and client both on your PC talk over this while developing.
- `0.0.0.0` — "all interfaces on this machine." When a server "listens on `0.0.0.0:8080`," it accepts connections arriving on any of the machine's network interfaces.
- Private ranges (`192.168.x.x`, `10.x.x.x`) — LAN-only addresses, not routable on the public internet. Your home devices use these behind your router.

### Port — *which program on that machine*

An IP gets you to the machine; a **port** gets you to the right *program* on it. A port is just a 16-bit number (0–65535). One machine, one IP, but many programs each listening on their own port.

- Ports **0–1023**: "well-known" (HTTP=80, HTTPS=443, SSH=22). Often require admin privileges to bind.
- Ports **1024–49151**: "registered," used by apps. Your game server on `8080` lives here.
- Ports **49152–65535**: "ephemeral," temporarily assigned to *outgoing* connections by the OS.

Analogy: the IP is a building's street address; the port is the apartment number. You need both to deliver a message to the right recipient.

### Socket — *one endpoint of a connection*

A **socket** is the programming abstraction for one end of a network conversation. Concretely, a socket is identified by the **4-tuple**:

```
(source IP, source port, destination IP, destination port)
```

That 4-tuple uniquely identifies a connection. This is why one server port can handle thousands of simultaneous clients: each client has a *different* source IP+port, so each connection is a distinct 4-tuple even though the server's IP+port is the same for all of them. The OS uses the full tuple to route incoming bytes to the correct socket.

A socket, to your program, behaves like a file handle / stream you read from and write to. That's the core abstraction of Chapter 6.

### Exercise 1.1 (conceptual)
Your tank server listens on `0.0.0.0:8080`. Two players connect from the same house (same public IP, say `86.20.5.1`) through the same router. How does the server keep their two connections separate, given they share a source IP? *(Answer: different source ports — the router assigns each a distinct ephemeral port via NAT, so the 4-tuples differ. More on NAT in Chapter 10.)*

---

## Chapter 2 — The Layered Model

Networking is built in **layers**, each using the one below and providing service to the one above. You don't need all seven of the academic "OSI model" — the practical stack is:

```
┌────────────────────────────────────────────────────────────┐
│  APPLICATION   HTTP, WebSocket, your game protocol          │  what your code speaks
├────────────────────────────────────────────────────────────┤
│  (TLS)         encryption, optional                          │
├────────────────────────────────────────────────────────────┤
│  TRANSPORT     TCP  |  UDP        ← the two real primitives  │  reliability / ports
├────────────────────────────────────────────────────────────┤
│  NETWORK       IP                 ← addressing & routing     │  which machine
├────────────────────────────────────────────────────────────┤
│  LINK          Ethernet, Wi-Fi    ← physical-ish transport   │  actual wire/radio
└────────────────────────────────────────────────────────────┘
```

**Key idea — encapsulation:** when you send a game message, each layer wraps it in its own header, like nesting envelopes:

```
[ Ethernet [ IP [ TCP [ your bytes ] ] ] ]
```

Your "move forward" message gets a TCP header (ports, sequence numbers) wrapped around it, then an IP header (source/dest IP) around that, then an Ethernet frame around that. At the destination, each layer peels off its header and hands the payload up. You program at the **application** layer and mostly treat the layers below as a pipe — but understanding they're there explains overhead, MTU limits (Chapter 4), and why UDP vs TCP is a *transport*-layer choice.

**Where the tank-server pieces live:**
- Your JSON/binary game messages = **application** layer (you design this).
- WebSocket = **application** layer (rides on TCP).
- TCP / UDP = **transport** layer (the actual primitives — Chapters 3–5).
- IP = **network** layer (you basically never touch it directly).

### Exercise 2.1 (conceptual)
When your guide said "WebSocket is three layers up from the transport primitive," name the layers between your game message and TCP. *(WebSocket sits on HTTP-upgrade semantics, which sits on TCP. So: your message → WebSocket framing → TCP. WebSocket adds message framing that raw TCP lacks — Chapter 7.)*

---

## Chapter 3 — TCP: The Reliable Stream

**TCP (Transmission Control Protocol)** is one of the two transport primitives. Its job: turn the unreliable underlying network into a **reliable, ordered, connection-oriented byte stream**.

### What TCP guarantees

1. **Reliable delivery** — lost packets are detected and retransmitted. What you send arrives (or the connection errors out).
2. **Ordered delivery** — bytes arrive in the order you sent them, even if the underlying packets took different routes and arrived scrambled.
3. **Connection-oriented** — a handshake establishes a session before data flows; both sides track state.
4. **Flow control & congestion control** — TCP slows down if the receiver or network is overwhelmed.

### The handshake (SYN, SYN-ACK, ACK)

Before any data, TCP does a **three-way handshake**:
```
Client → Server:  SYN          "let's talk, my sequence starts at X"
Server → Client:  SYN-ACK      "ok, and mine starts at Y, I got your X"
Client → Server:  ACK          "got your Y, we're connected"
```
Only then does data flow. This round-trip is *setup latency* — one reason connectionless UDP can start faster.

### The critical downside for games: head-of-line blocking

Because TCP guarantees *order*, if packet #5 is lost, packets #6, #7, #8 **cannot be delivered to your application even though they already arrived** — TCP holds them in a buffer until #5 is retransmitted and arrives, so it can hand them up in order. This is **head-of-line (HOL) blocking**.

For a game this is often *the wrong tradeoff*: if a position update is lost, you don't want the newer position updates stuck waiting behind it — the newer ones *supersede* the lost one anyway. You'd rather have the latest and skip the stale. TCP can't do that; it insists on delivering everything in order. This single property is why real-time games often prefer UDP. Hold this thought for Chapter 5.

### "Stream, not messages"

Crucial TCP subtlety that bites everyone: TCP is a **byte stream, not a message stream.** If you `write("HELLO")` then `write("WORLD")`, the receiver might `read` `"HELLOWORLD"` in one call, or `"HELL"` then `"OWORLD"`, or any split. TCP does not preserve your message boundaries — it just delivers bytes in order. **You** must add framing to know where one message ends and the next begins (Chapter 7). This is the #1 beginner networking bug.

### Exercise 3.1 (conceptual)
You send three 30Hz position snapshots over TCP. The packet carrying snapshot #2 is lost. Describe what your game receives and when, and why that's bad for real-time rendering. *(You get snapshot #1, then a stall while #2 retransmits, then #2 and #3 together — a hitch. You'd have preferred to skip #2 entirely and show #3 immediately.)*

---

## Chapter 4 — UDP: The Raw Datagram

**UDP (User Datagram Protocol)** is the other transport primitive, and it's almost the *opposite* of TCP: minimal, connectionless, unreliable — a thin wrapper over raw IP packets.

### What UDP does (and doesn't) do

UDP gives you:
- **Datagrams** — you send discrete packets; each `send` is one packet, each `receive` is one packet. **Message boundaries are preserved** (unlike TCP's stream!).
- **Ports** — so multiple programs can share a machine (same as TCP).
- **A checksum** — corrupt packets are dropped.

UDP does **not** give you:
- **No delivery guarantee** — packets can vanish silently. No retransmission.
- **No ordering** — packet #5 can arrive before #3.
- **No connection** — no handshake; you just fire packets at an address.
- **No congestion control** — you can flood the network (your responsibility to be nice).

### Why games love it

UDP is the "give me the primitive, I'll build what I need" option. For real-time games:
- **No head-of-line blocking** — a lost packet doesn't stall later ones. Newer state arrives immediately.
- **You choose per-message reliability** — position snapshots can be fire-and-forget (latest wins, losing one is fine), while critical events (a player fired, a player died) get *your own* reliability layer (sequence numbers + acks + resend). You pay the reliability cost only where it's needed.
- **Lower latency** — no handshake setup, no forced retransmit waits.

The catch: everything TCP gave you for free — reliability where you need it, ordering where you need it, framing-into-messages, connection tracking — **you build yourself** on top of UDP. This is exactly why the tank-server guide called the UDP migration "a whole project": you're reimplementing the useful 20% of TCP that your game actually needs, and nothing more.

### Datagram size & MTU

A practical UDP constraint: packets shouldn't exceed the network's **MTU** (Maximum Transmission Unit, typically ~1500 bytes on Ethernet, safely ~1200 for internet use). Larger UDP datagrams get **fragmented** at the IP layer, and if *any* fragment is lost the whole datagram is lost — fragmentation multiplies loss probability. So game netcode keeps packets small (under ~1200 bytes), which shapes how much state you cram into one snapshot. TCP handles this for you (it segments the stream); with UDP it's your concern.

### Exercise 4.1 (conceptual)
Your tank game sends two kinds of messages: continuous position snapshots (30/sec) and "tank fired a shot" events. Over UDP, which needs a reliability layer and which is fine as fire-and-forget? Why? *(Snapshots: fire-and-forget — losing one is invisible, the next supersedes it. Fire events: need reliability — if lost, a shot never happens and state desyncs, so they need sequence+ack+resend.)*

---

## Chapter 5 — TCP vs UDP: The Decision

The whole comparison, and how it maps to your project's choices.

| Property | TCP | UDP |
|---|---|---|
| Reliability | guaranteed | none (build your own) |
| Ordering | guaranteed | none (build your own) |
| Connection | yes (handshake) | no |
| Message boundaries | **no** (byte stream) | yes (datagrams) |
| Head-of-line blocking | yes (bad for real-time) | no |
| Per-message reliability control | no (all-or-nothing) | yes (you choose) |
| Setup latency | handshake round-trip | none |
| Congestion control | built in | your responsibility |
| Complexity to use | low | high (you build the missing parts) |
| Firewall/NAT friendliness | high | lower (more often blocked) |

### The honest guidance

- **Turn-based games** (cards, chess, tactics): **TCP.** Reliability and ordering are exactly what you want; latency is a non-issue. Nothing to build.
- **Fast-twitch action** (competitive FPS, fighting games): **UDP.** Head-of-line blocking is unacceptable; you need latest-state-wins and per-message reliability. Worth the complexity.
- **In between** (your tank arena): **either can work.** Tanks are slower than an FPS, so TCP's occasional retransmit hitch is tolerable — which is exactly why your guide starts on **WebSocket (TCP)** to keep you focused on learning Go and architecture, then offers UDP as a later "real netcode" upgrade once the game works.

### Why start on TCP/WebSocket then migrate

This is the pedagogically correct order and worth stating plainly: you learn the **architecture** (authoritative server, snapshots, interpolation, prediction — Part III) on a transport that Just Works, so bugs are *your game logic's* bugs, not "did my hand-rolled UDP ack layer drop something" bugs. Once the architecture is solid, swapping the transport to UDP isolates the netcode-plumbing learning to one layer. Trying to learn both at once is how people burn out on multiplayer projects.

### Exercise 5.1 (conceptual)
Justify, in your own words, why an authoritative tank server on WebSocket/TCP is a reasonable *v1* even though "real" action games use UDP — and what specific symptom would tell you you've outgrown TCP and need the UDP migration. *(Symptom: under packet loss, players see periodic freezes/rubber-banding as TCP stalls on retransmits — the HOL-blocking tax. If that becomes noticeable at your target ping/loss, it's time for UDP.)*

---

# Part II — Programming Against the Network

## Chapter 6 — The Sockets API

Every language's networking, under the hood, wraps the same **BSD sockets** API (the C API from the 1980s that became the universal standard). Understanding the shape once means you understand it in every language.

### The server side

```
socket()   → create an endpoint
bind()     → attach it to an IP:port (e.g. 0.0.0.0:8080)
listen()   → mark it as accepting incoming connections (TCP)
accept()   → block until a client connects; returns a NEW socket for that client
read()/write() on that per-client socket
close()
```

The key insight: **`accept()` returns a *new* socket per client.** Your listening socket stays listening; each accepted connection is its own socket (its own 4-tuple). This is why a server handles many clients on one port — each gets a distinct socket. In your tank server, "one goroutine per connection" means one goroutine per accepted socket.

### The client side

```
socket()   → create an endpoint
connect()  → initiate connection to server IP:port (TCP handshake happens here)
read()/write()
close()
```

### Blocking vs non-blocking — the concurrency crux

By default, socket operations **block**: `read()` waits until data arrives; `accept()` waits until a client connects. A naive single-threaded server can therefore handle only one client at a time. The solutions, historically:

1. **Thread-per-connection** — spawn an OS thread per client. Simple, but OS threads are heavy (~1MB stack each); thousands of them strain the machine.
2. **Non-blocking + event loop** (`select`/`epoll`/`kqueue`/IOCP) — one thread watches many sockets, reacting as each becomes ready. Scales to huge connection counts but the code is callback-heavy and harder to write.
3. **Green threads / async runtime** — the language runtime multiplexes many lightweight "threads" onto few OS threads, giving you thread-per-connection's *simple blocking code* with the event loop's *scalability*.

**Option 3 is exactly what Go's goroutines give you**, and why Go is so pleasant for servers: you write straightforward blocking `read()` code, one goroutine per connection, and the runtime uses an epoll-style event loop underneath to scale it. You get the simple mental model and the performance. (This is also what Java's virtual threads and Rust's async aim at — Go just made it the default.)

### Exercise 6.1 (conceptual)
Explain why "one goroutine per connection, each blocking on read" doesn't fall over at 10,000 connections, whereas "one OS thread per connection" might. *(Goroutines start at ~2KB and are multiplexed by the runtime onto a small pool of OS threads via an event loop; 10k goroutines is fine. 10k OS threads is ~10GB of stack plus heavy context-switching.)*

---

## Chapter 7 — Framing: Where Does a Message End?

The #1 practical networking bug, and a concept every game-netcoder must own.

### The problem (TCP)

As Chapter 3 warned, TCP is a **byte stream** with no message boundaries. You send:
```
{"type":"fire"}       {"type":"move","x":5}
```
The receiver's `read()` might return:
```
{"type":"fire"}{"type":"mo         ← partial! second message cut in half
```
or both glued together, or one byte at a time. **You cannot assume one `read()` = one message.** You must *frame*: encode where each message starts and ends.

### The three framing strategies

**1. Length-prefixing (most common for binary/games).**
Before each message, send its length as a fixed-size integer:
```
[4-byte length][ ...that many bytes of message... ][4-byte length][ ...message... ]
```
The receiver reads exactly 4 bytes → learns N → reads exactly N more bytes → that's one complete message → repeat. Robust, efficient, the standard choice for game protocols.

**2. Delimiter-based (common for text).**
End each message with a sentinel byte, e.g. newline `\n`:
```
{"type":"fire"}\n{"type":"move"}\n
```
Read until you hit the delimiter. Simple, but breaks if the delimiter can appear *inside* a message (must escape it), so it's mostly for line-based text protocols.

**3. Fixed-size messages.**
If every message is exactly the same size, just read that many bytes. Rare, but simple when it applies.

### Why WebSocket saved you this in v1

**WebSocket has framing built in.** It's a *message*-oriented protocol layered on TCP's stream — the WebSocket layer does length-prefixing internally, so your `Read()` gives you exactly one complete message every time. That's a big reason the tank-server guide starts on WebSocket: you skip hand-writing framing while learning everything else. When you migrate to raw TCP or UDP, framing becomes your job again — length-prefixing for TCP; UDP datagrams are already message-bounded (one datagram = one message) so UDP gives framing back for free, but takes reliability away.

### Exercise 7.1 (conceptual)
You move your tank protocol from WebSocket to raw TCP. Which framing strategy fits binary game messages, and walk through how the receiver reads exactly one message. *(Length-prefix: read 4 bytes → N; loop reading until you've accumulated N bytes → one message; repeat. Must handle partial reads by accumulating in a buffer.)*

---

## Chapter 8 — Serialization: Structs to Bytes

The network moves **bytes**. Your game has **structs** (`Tank{Pos, Angle, Health}`). Serialization is the translation both ways: struct → bytes (to send) and bytes → struct (on receive). Also called marshaling/encoding.

### Text formats (readable, bigger, slower)

**JSON** — human-readable, self-describing, ubiquitous. `{"health":100,"x":5.0}`. Great for **development and debugging** because you can read the wire. Costs: verbose (field names repeated every message), floats become strings, parsing is comparatively slow. This is why the tank guide *starts* here — you can see everything — then moves off it.

### Binary formats (compact, fast, opaque)

**Hand-rolled binary** — you pack fields directly: a 1-byte type tag, then the raw bytes of each field in a fixed order. Smallest and fastest, but you write and maintain the encoder/decoder on both ends, and it's easy to desync the two sides. Excellent *learning* exercise.

**Schema-based (Protobuf, FlatBuffers, Cap'n Proto)** — you write a schema once (`.proto`), and a tool generates encoder/decoder code for every language you need. This is the pragmatic winner for **multi-language projects** — which yours is (Go server + C++ client). One schema, generated Go *and* C++ types, no hand-syncing serialization across two languages. FlatBuffers/Cap'n Proto additionally allow reading fields without fully parsing (zero-copy), which matters at high message rates.

### The cross-language trap

Two things must match *exactly* on both ends or you get garbage:
1. **Field order and types** — both sides must agree byte-for-byte on layout.
2. **Endianness (byte order)** — the order of bytes within a multi-byte number. **Big-endian** (most significant byte first, the "network byte order" convention) vs **little-endian** (least significant first, what x86/ARM use internally). If your Go server writes a little-endian int and your C++ client reads it big-endian, `100` becomes `1677721600`. **Pick one, document it, convert explicitly on both sides.** Schema tools handle this for you — another reason they shine for two-language projects.

### Exercise 8.1 (conceptual)
Your snapshot has 8 tanks × (2 floats position + 1 float angle + 1 int health). Estimate the byte size in naive JSON vs packed binary (floats = 4 bytes, int = 4 bytes, ignore JSON whitespace). Why does this gap matter at 30 snapshots/sec? *(Binary: 8 × (4+4+4+4) = 128 bytes. JSON: easily 400–600 bytes with field names/braces/string-encoded numbers. At 30Hz × many players the bandwidth and parse-time difference compounds fast.)*

---

## Chapter 9 — Higher-Level Protocols: HTTP, WebSocket, QUIC

Where the convenience layers sit, and the modern options.

### HTTP — request/response

The web's protocol: client sends a request, server sends a response, done. **Stateless, one-shot, client-initiated.** Bad for games (the server can't *push* updates; the client would have to constantly re-ask "anything new?"). But it's how connections *start* before upgrading, and how you'd do non-realtime bits (login, matchmaking REST calls, leaderboards).

### WebSocket — persistent bidirectional over TCP

Solves HTTP's limits for real-time-ish needs: after an HTTP "upgrade" handshake, the connection becomes a **persistent, bidirectional, message-framed** channel over the same TCP socket. Either side can send anytime. Message framing is built in (Chapter 7). Firewall-friendly (looks like web traffic). This is your v1 transport: reliable, framed, trivially usable from Go and C++, lets you learn architecture without transport plumbing. Its ceiling is TCP's ceiling — head-of-line blocking under loss.

### QUIC & WebTransport — the modern answer

**QUIC** is a newish transport built *on top of UDP* that bundles the good parts of TCP+TLS without the head-of-line blocking:
- Runs over UDP, so no kernel TCP semantics forcing in-order delivery.
- Provides **multiple independent streams** in one connection — a loss in one stream doesn't block others (fixes HOL blocking *between* streams).
- Supports both **reliable ordered** streams *and* **unreliable datagrams** in the same connection — so you can send snapshots unreliably and events reliably, over one connection, with encryption built in.
- Faster connection setup than TCP+TLS (fewer round-trips).

**WebTransport** is the browser-facing API built on QUIC — think "WebSocket's successor," giving web clients access to unreliable + reliable delivery.

**Why this matters for your migration path:** the tank guide frames the future upgrade as "WebSocket → raw UDP," which means hand-building reliability/ordering/framing. **QUIC/WebTransport is a middle path**: you get UDP-style unreliable datagrams for snapshots *and* reliable streams for events *and* framing *and* encryption, without hand-rolling an ack layer. For a modern game the realistic choices are (a) raw UDP with a custom reliability layer (max control, most work, what AAA engines historically did) or (b) QUIC/WebTransport (most of the benefit, far less plumbing, especially if you ever want browser clients). Given you're building a C++ client, both are reachable; QUIC libraries exist for C++ and Go. Worth evaluating QUIC before committing to hand-rolled UDP when the time comes.

### Exercise 9.1 (conceptual)
QUIC gives you "unreliable datagrams and reliable streams in one connection." Map your two tank message types onto those. *(Position snapshots → unreliable datagrams, latest-wins, no HOL stall. Fire/death events → reliable stream, guaranteed delivery. One connection, per-message-type reliability — the thing raw UDP makes you build by hand.)*

---

# Part III — The Real-Time Problem (Game Netcode)

## Chapter 10 — Why Real-Time Is Hard: Latency, Jitter, Loss

Everything in Part III exists to fight three enemies. Name them precisely:

### Latency (ping / RTT)

**Latency** is the time for data to travel one way; **RTT (round-trip time)** is there-and-back — what "ping" measures. Even at light-speed limits, cross-continent RTT is ~80–150ms; add real-world routing and it's often 30–100ms domestically, 150ms+ internationally. **You cannot eliminate latency — it's physics.** Every netcode technique is about *hiding* or *compensating* for it, never removing it.

Consequence: when a player presses fire, the server won't know for ~half-a-ping. When the server updates the world, the player won't see it for ~half-a-ping. The player is always acting on a slightly *old* world and their actions always apply slightly *late*. This delay is the root problem.

### Jitter

**Jitter** is *variation* in latency — packets arriving 40ms, then 45ms, then 38ms, then 90ms apart. Even if average latency is fine, inconsistent arrival makes motion stutter if you render packets the instant they arrive. The fix is a small **buffer** that smooths arrival (interpolation delay, Chapter 12) — you trade a little added latency for smooth motion.

### Packet loss

Packets vanish (congestion, Wi-Fi interference, routing hiccups). On UDP they're simply gone; on TCP they trigger a retransmit-and-stall (HOL blocking). Netcode must **degrade gracefully** under loss — a lost snapshot should cause at most a tiny visual hiccup, not a desync or freeze.

### The NAT complication (why P2P is hard)

Most players are behind a **NAT** (Network Address Translation) router that shares one public IP among many devices, rewriting ports on outgoing traffic. This is why Exercise 1.1's two housemates got distinct ports. NAT means a machine usually **can't receive unsolicited inbound connections** — the router doesn't know which internal device to route them to. Consequences:
- **Client→server works fine** (client initiates outbound; router remembers the mapping). This is why your **dedicated-server** model (client connects out to your server) sidesteps NAT entirely — a big practical reason to prefer it over peer-to-peer.
- **Peer-to-peer is hard** — two NAT'd peers can't directly connect without tricks (STUN/TURN/hole-punching). Another point for the authoritative-server architecture: it avoids the NAT-traversal rabbit hole.

### Exercise 10.1 (conceptual)
A player has 80ms RTT. They press fire. In a naive "send input, wait for server, then show result" design, how long until they see their own shot appear, and why does that feel terrible? *(~80ms — a full round trip — before their own shot renders. Input feeling laggy by 80ms is very noticeable. Chapter 13's client-side prediction fixes exactly this by showing the shot locally *immediately*.)*

---

## Chapter 11 — Authoritative Servers

The foundational architecture decision, and the one your tank server uses.

### The model

**One machine — the server — owns the true game state.** Clients don't decide what happens; they send *inputs* ("I'm holding forward," "I fired"), the server simulates, and the server's result is **authoritative truth**. Clients render a *copy* of that truth.

Contrast with the naive alternative — clients tell the server their positions ("I'm at x=500 now"). That's trivially cheatable: a hacked client just claims to be anywhere, claims every shot hits, claims infinite health. **If the client asserts state, the client can lie.** So: clients send *intent*, never *authoritative state*.

### Why it wins

- **Anti-cheat** — the server validates everything. A client claiming an impossible move or a shot through a wall is simply ignored; the server computes reality.
- **Single source of truth** — no "whose version is right" conflicts between peers. The server arbitrates.
- **NAT-friendly** — clients connect *out* to the server (Chapter 10), sidestepping inbound-NAT problems.

The cost: latency (Chapter 10). Because the server decides, and it's ~half-a-ping away, a client's actions don't take visible effect until the round trip completes — *unless* you add prediction (Chapter 13). The authoritative model creates the latency-hiding problem that the rest of Part III solves.

### The server tick

The server runs a **fixed-timestep simulation loop** — the "tick." Each tick (say 30/sec) it: ingests queued client inputs, advances the simulation by one fixed `dt`, resolves collisions/hits, and emits a state **snapshot** to clients. Fixed timestep matters for determinism and fairness — every client's input is applied against the same regular clock. This is exactly the `select`-loop-with-ticker in the tank-server guide, and the "authoritative" part is why the server, not the C++ client, runs the real physics.

### Exercise 11.1 (conceptual)
Why does "clients send inputs, server simulates" prevent an aimbot-style position hack that "clients send positions" cannot? *(If the client only sends inputs, it can't place itself anywhere — the server integrates movement from inputs and enforces speed/walls. A position-asserting protocol trusts the client's claimed location, which a hack fabricates freely.)*

---

## Chapter 12 — Snapshots & Interpolation

How clients display authoritative state smoothly despite low snapshot rates and jitter.

### Snapshots

The server sends **snapshots** — periodic captures of world state (all tank positions, angles, health) — at the tick rate, e.g. 30/sec. Between snapshots the client hears nothing new about remote tanks. Two problems: (1) 30 snapshots/sec but the client renders at 60–144fps, so most frames have no fresh data; (2) jitter makes snapshots arrive unevenly. Naively "snap each tank to its latest snapshot position the moment it arrives" produces stuttery, teleporting motion.

### Interpolation — render the past, smoothly

The fix: **don't render remote entities at "now" — render them slightly in the past, interpolating between two known snapshots.** Concretely:

1. Buffer the last few snapshots, each tagged with the server time/tick it represents.
2. Deliberately render at `now − interpolationDelay` (e.g. 100ms behind real-time — roughly 2 snapshot intervals at 30Hz).
3. Because you're rendering in the past, you almost always have *two* real snapshots straddling your render time — one just before, one just after. **Lerp** (linear-interpolate) each remote tank's position between them based on where render-time falls between the two snapshot timestamps.

You trade a small, constant, unnoticeable delay (~100ms on *remote* entities) for buttery-smooth motion that's immune to jitter and tolerant of the occasional lost snapshot (you interpolate across the gap). This is **entity interpolation**, and it's standard in essentially every networked game. It's the exact `[HOB2D SEAM]` interpolation-buffer contract in your tank guide.

### Why "the past" is OK

You're watching *other* players ~100ms behind reality. Human players don't notice a consistent 100ms delay on *others'* movement (it's smooth and predictable). What they *would* notice is lag on *their own* input — and that's handled separately by prediction (Chapter 13). Split responsibilities: **interpolation smooths remote entities; prediction hides latency on the local player.**

### Extrapolation (the riskier cousin)

If you have no future snapshot to interpolate toward (e.g. a snapshot was lost and none newer arrived yet), you can **extrapolate** — extend an entity's last known velocity forward, guessing where it is. Cheap but wrong when the entity changes direction, causing a visible "snap back" when the real snapshot arrives. Most games interpolate and tolerate a brief freeze on loss rather than extrapolate aggressively. Know it exists; prefer interpolation.

### Exercise 12.1 (conceptual)
Why render remote tanks ~100ms in the past instead of at their latest received position? What breaks if you set interpolation delay to zero? *(At zero delay you have no "next" snapshot to interpolate toward, so you're forced to either snap to each arriving snapshot (stutter) or extrapolate (guessing, snap-backs). A small buffer guarantees two straddling snapshots for smooth lerping and absorbs jitter/loss.)*

---

## Chapter 13 — Client-Side Prediction & Reconciliation

How to make the *local* player feel instant despite the server being half-a-ping away. This is the hardest core netcode concept — and why the tank guide defers it to an explicit optional v2.

### The problem it solves

With a pure authoritative model, when *you* press forward, the sequence is: input → server (½ ping) → server simulates → snapshot back (½ ping) → you finally move. That's a full RTT of input lag on *your own tank* — feels sluggish and awful even at moderate ping (Exercise 10.1).

### Client-side prediction

Don't wait for the server for your *own* actions. The client **runs the same movement simulation locally** and applies your input **immediately** — your tank moves the instant you press forward, zero perceived lag. You still send the input to the server (it remains authoritative); you've just *predicted* the result locally in the meantime, using the same deterministic movement code the server runs.

### The catch: you might predict wrong

Your prediction can diverge from the server's authority (you collided with something the server knew about, or another player's action changed things). When a server snapshot arrives, it might disagree with what you predicted. You can't just snap to it — that would cause constant jitter. You need **reconciliation**.

### Server reconciliation

The technique:
1. **Number every input** you send (input #1, #2, #3…) and **keep a local history** of them.
2. The server processes inputs and, in its snapshots, **acknowledges the last input number it applied** for you.
3. When a snapshot arrives saying "I've applied up to input #5, you're at position P," the client: **resets its local tank to the authoritative P**, then **re-applies inputs #6, #7, #8…** (everything sent after #5, still in the history and not yet acknowledged) on top of P.
4. Result: the local tank ends up at a corrected-but-still-responsive position, having reconciled server truth with your still-pending inputs.

If your prediction was correct (common case), reconciliation changes nothing visible. If it was wrong, you get a small correction rather than a wait-for-the-server stall. This **predict-and-reconcile** loop is the heart of responsive authoritative netcode (the model popularized by competitive shooters).

### Why it's deferred in your project

It requires: deterministic movement shared identically by client and server, input sequence numbering, input history buffers, server ack-of-last-input in snapshots, and the reset-and-replay logic — a meaningful chunk of machinery on both sides. Get the game **working correctly server-authoritative first** (local player rendered from snapshots, slightly laggy but simple and right), *then* add prediction as an isolated upgrade. Starting with prediction is a classic way to get a subtly-broken multiplayer game you can't debug, because you can't tell prediction bugs from logic bugs.

### Exercise 13.1 (conceptual)
Walk through what the client does when it has sent inputs #1–#8, and a snapshot arrives acknowledging input #5 with an authoritative position. Why re-apply #6–#8 instead of just accepting the position as-is? *(Reset local tank to the acked position (truth as of #5), then replay #6–#8 from the local history so the tank reflects inputs the server hasn't processed yet — otherwise the tank would visibly jump backward to where it was at #5, discarding 3 inputs' worth of your recent movement.)*

---

## Chapter 14 — Lag Compensation & Putting It Together

The last technique, then the full picture.

### The hit-detection problem

Interpolation (Chapter 12) means you *see* other tanks ~100ms in the past. So when you aim at an enemy and fire, you're aiming at where they *were*, not where the server currently thinks they are. Naively, the server checks the shot against the enemy's *current* position — which is ahead of what you saw — so your visually-perfect shot **misses**. Frustrating and feels broken.

### Lag compensation (server-side rewind)

The server fixes this by **rewinding time** for hit checks:
1. The server keeps a short **history of past positions** of every entity (the last several hundred ms of snapshots).
2. When a client fires, the client's message includes (or the server infers) **what client-time the shot was taken at** — i.e., the ~100ms-in-the-past world the player was actually looking at.
3. The server **rewinds the world to that moment** — reconstructs where every tank *was* when the player fired — and checks the shot against *those* positions.
4. If it hit in the player's view, it hits, authoritatively.

This "server rewind" makes shooting feel fair — you hit what you aimed at — at the cost of the occasional "shot me from behind cover" moment for the victim (because the server honored the shooter's slightly-old view). Every competitive shooter makes this tradeoff; it favors the shooter's experience.

### The complete picture — how it all fits

For your tank game, fully built out, the end-to-end loop is:

1. **Transport** (Ch.3–9) — reliable framed messages (WebSocket v1; UDP/QUIC later) carry inputs up and snapshots down.
2. **Authoritative server** (Ch.11) — owns state, runs a fixed 30Hz tick, ingests inputs, simulates, resolves hits, emits snapshots. Clients never assert state.
3. **Serialization** (Ch.8) — inputs and snapshots encode to bytes (JSON v1; binary/Protobuf later), matched byte-for-byte on the C++ and Go sides.
4. **Client prediction + reconciliation** (Ch.13) — the local tank responds instantly; corrections reconcile against server snapshots using numbered-input replay. *(v2 — deferred.)*
5. **Entity interpolation** (Ch.12) — remote tanks render ~100ms in the past, lerped between snapshots, smooth under jitter and loss.
6. **Lag compensation** (Ch.14) — the server rewinds to the shooter's view for fair hit detection. *(Advanced — add once the above works.)*

**The build order that keeps you sane** (mirrors the tank guide's milestones): get transport + authoritative server + interpolation working first — that's a complete, correct, playable game where remote tanks move smoothly and the local player is slightly laggy. *Then* layer on prediction (local player feels instant), *then* lag compensation (shooting feels fair). Each layer is isolated and independently testable. Never try to build all six at once.

### Exercise 14.1 (conceptual, capstone)
Trace a single "player fires and hits an enemy" event through the fully-built system: what the shooter's client does locally and sends, what the server does with it (including rewind), and how both the shooter and the victim eventually see the result. Name which chapter's technique handles each step. *(Rough path: client shows muzzle flash immediately via prediction (Ch.13) and sends a timestamped fire input (Ch.8 serialization, Ch.7 framing) over the transport (Ch.9); server receives on that player's goroutine, hands it to the authoritative room loop (Ch.11), rewinds enemy positions to the shooter's ~100ms-past view (Ch.14), checks the hit, applies damage authoritatively; next snapshot broadcasts the enemy's new health/death; both clients render it — the victim via interpolation (Ch.12), the shooter reconciling against the confirmed result (Ch.13).)*

---

## Appendix: Glossary of Every Term

| Term | Meaning |
|---|---|
| IP address | identifies a machine on a network |
| Port | identifies a program on that machine (0–65535) |
| Socket | one endpoint of a connection; the 4-tuple (src IP, src port, dst IP, dst port) |
| TCP | reliable, ordered, connection-oriented byte-stream transport |
| UDP | unreliable, unordered, connectionless datagram transport |
| Handshake | TCP's SYN/SYN-ACK/ACK connection setup |
| Head-of-line blocking | TCP stalling later data behind a lost packet |
| MTU | max packet size before fragmentation (~1500 bytes) |
| Framing | marking where one message ends and the next begins |
| Length-prefixing | framing by sending each message's byte-length first |
| Serialization | converting structs ↔ bytes (aka marshaling) |
| Endianness | byte order of multi-byte numbers (big- vs little-endian) |
| Latency / RTT | one-way / round-trip travel time ("ping" = RTT) |
| Jitter | variation in latency |
| Packet loss | packets vanishing in transit |
| NAT | router sharing one public IP among many devices; blocks unsolicited inbound |
| WebSocket | persistent, framed, bidirectional messaging over TCP |
| QUIC | modern UDP-based transport with reliable streams + unreliable datagrams, no HOL blocking |
| WebTransport | browser API over QUIC |
| Authoritative server | server owns true state; clients send inputs, not state |
| Tick | one fixed-timestep step of the server simulation |
| Snapshot | periodic capture of world state sent to clients |
| Interpolation | rendering remote entities smoothly between two past snapshots |
| Extrapolation | guessing an entity's position ahead of known data (risky) |
| Client-side prediction | applying local input immediately without waiting for the server |
| Reconciliation | correcting a mispredicted local state by replaying unacked inputs |
| Lag compensation | server rewinding to the shooter's view for fair hit detection |

---

*Read Part I to understand the pipes, Part II to program them, Part III to build real-time games on them. Then the tank-server guide's choices — WebSocket first, authoritative server, snapshots + interpolation, deferred prediction — will read as obvious consequences of everything here.*
