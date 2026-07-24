# Spring Boot for the ASP.NET Core Developer — Building the Game Backend

*A fast-track guide for someone who already knows ASP.NET Core, C#, and Entity Framework — and who is building the "meta" backend for a multiplayer game.*

The goal: build the **backend services** for your multiplayer game — accounts, authentication, matchmaking, leaderboards, and persistent stats — as a Spring Boot REST API backed by PostgreSQL, with real authentication via Spring Security. This is the Spring Boot half of the architecture; the Go server handles real-time gameplay (see GO.md), and later calls into this backend over HTTP.

```
┌─────────────────┐      ┌──────────────────────┐  HTTP/REST  ┌─────────────────────┐
│ Browser client  │  WS  │ Go game server       │◄───────────►│ Spring Boot backend │  ← THIS GUIDE
│                 │◄────►│ real-time gameplay   │             │ accounts / auth     │
└─────────────────┘      └──────────────────────┘             │ matchmaking         │
                                                              │ leaderboards / DB   │
                                                              └─────────────────────┘
```

**Why Spring Boot is the right tool here:** accounts/leaderboards/matchmaking are CRUD-and-database work — REST endpoints, persistence, auth, business logic — which is exactly Spring Boot's wheelhouse and exactly where your ASP.NET Core experience transfers almost one-to-one. This guide is framed throughout as **ASP.NET Core → Spring Boot translation**, because the concepts are the same; mostly you're learning Java's annotation names for patterns you already know.

**Prerequisite:** the core Java chapters (JAVA.md Ch.1-14). Spring Boot leans on Java fundamentals — generics, interfaces, records, streams, build tooling — and learning Spring's annotation magic while still fighting Java syntax is its own scope-creep trap. Get Java into muscle memory first; then Spring Boot layers on cleanly.

**Versions (verify before starting — Spring moves fast):** targets **Spring Boot 4.1** (the current stable as of mid-2026, built on Spring Framework 7 and Spring Security 7), **Java 25** (your install; Spring Boot 4.1 supports Java 17-26), **PostgreSQL**. Note Spring Boot has NO LTS — every minor ships on a 12-month support cycle, new minors every 6 months. Check start.spring.io for current versions when you scaffold.

---

## Table of Contents

1. The ASP.NET Core → Spring Boot Map
2. Scaffolding with Spring Initializr
3. Your First REST Controller
4. The Data Layer: JPA & PostgreSQL
5. Services, DI, and Project Structure
6. Accounts: Entities, Validation, Registration
7. Leaderboards: Queries, Sorting, Pagination
8. Matchmaking & Match Results
9. Spring Security: Real Authentication
10. Configuration & Profiles
11. Testing
12. The HTTP Contract for the Go Server
13. Deployment

---

## Chapter 1 — The ASP.NET Core → Spring Boot Map

Before any code, internalize this — it's most of what you need conceptually. You already know these patterns; you're learning new names.

| ASP.NET Core | Spring Boot |
|---|---|
| `[ApiController]` class | `@RestController` class |
| `[HttpGet]` / `[HttpPost]` | `@GetMapping` / `@PostMapping` |
| `[Route("api/players")]` | `@RequestMapping("/api/players")` |
| `[FromBody]` param | `@RequestBody` param |
| `[FromRoute]` / `{id}` | `@PathVariable` |
| `[FromQuery]` | `@RequestParam` |
| `IActionResult` / `Ok()` / `NotFound()` | `ResponseEntity<T>` / `.ok()` / `.notFound()` |
| built-in DI container | Spring's IoC container |
| constructor injection | constructor injection (same!) |
| `services.AddScoped<IFoo, Foo>()` | `@Service` + component scanning |
| `Program.cs` / `Startup.cs` | `@SpringBootApplication` + auto-config |
| Entity Framework `DbContext` | Spring Data JPA + Hibernate |
| `DbSet<Player>` | `JpaRepository<Player, Long>` |
| EF migrations | Flyway / Liquibase (or JPA auto-DDL in dev) |
| Kestrel (built-in server) | embedded Tomcat |
| `appsettings.json` | `application.yml` / `application.properties` |
| environment configs | Spring profiles (`application-dev.yml`) |
| `[Authorize]` | `@PreAuthorize` / Spring Security config |
| ASP.NET Identity | Spring Security + your user store |
| NuGet | Maven / Gradle |
| xUnit / NUnit | JUnit 5 |
| Middleware pipeline | Servlet filters / Spring interceptors |

The architecture you'll build is the one you already know: **controllers → services → repositories → database.** Controllers handle HTTP, services hold business logic, repositories talk to the DB, DI wires them together. Identical layering to a clean ASP.NET Core API.

---

## Chapter 2 — Scaffolding with Spring Initializr

The `dotnet new webapi` equivalent is **Spring Initializr** — a project generator at start.spring.io (also built into IntelliJ Ultimate and the VS Code Spring extension).

### Generate the project
Go to https://start.spring.io and select:
- **Project:** Gradle (Groovy or Kotlin DSL) — modern default. (Maven also fine; Spring supports both.)
- **Language:** Java
- **Spring Boot:** the current 4.1.x stable
- **Java:** 25 (or 17+ — match your install)
- **Dependencies** (the "starters" — curated dependency bundles):
  - **Spring Web** — REST controllers, embedded Tomcat
  - **Spring Data JPA** — the EF-equivalent data layer
  - **PostgreSQL Driver** — the DB driver
  - **Spring Security** — authentication (Ch.9)
  - **Validation** — request validation (`@Valid`)
  - **Spring Boot DevTools** — auto-restart on code change (like `dotnet watch`)

Download, unzip, open in your IDE. (IntelliJ IDEA is the Java-native choice and will feel like Rider; VS Code with the Spring Boot Extension Pack also works.)

### What you got
```
game-backend/
├── build.gradle              # deps + build config (≈ .csproj)
├── src/main/java/com/denis/gamebackend/
│   └── GameBackendApplication.java   # entry point
├── src/main/resources/
│   └── application.properties        # ≈ appsettings.json
└── src/test/java/...
```

### The entry point
```java
@SpringBootApplication
public class GameBackendApplication {
    public static void main(String[] args) {
        SpringApplication.run(GameBackendApplication.class, args);
    }
}
```
`@SpringBootApplication` bundles three annotations: `@Configuration` (this is a config class), `@EnableAutoConfiguration` (Spring configures itself based on what's on the classpath — see a PostgreSQL driver? set up a datasource), and `@ComponentScan` (find all `@Component`/`@Service`/`@RestController` classes in this package and below, and register them for DI). This auto-configuration is Spring Boot's core magic — the equivalent of ASP.NET Core's convention-based setup, but more aggressive.

Run it (`./gradlew bootRun` or the IDE run button). It starts embedded Tomcat on port 8080. You now have a running (empty) web server.

### Exercise 2.1
Scaffold the project with the dependencies above, run it, confirm Tomcat starts on 8080. (It'll 404 everything — no endpoints yet. That's next.)

---

## Chapter 3 — Your First REST Controller

```java
@RestController
@RequestMapping("/api/health")
public class HealthController {

    @GetMapping
    public Map<String, String> health() {
        return Map.of("status", "ok");
    }
}
```
`@RestController` = `[ApiController]` (implies responses are serialized to JSON, not view names). `@RequestMapping` sets the base path. `@GetMapping` maps HTTP GET. The returned `Map` is auto-serialized to JSON by Jackson (Spring's JSON library, ≈ System.Text.Json). Hit `http://localhost:8080/api/health` → `{"status":"ok"}`.

### Path variables, query params, request bodies
```java
@GetMapping("/{id}")
public Player getById(@PathVariable Long id) { ... }              // ≈ [FromRoute]

@GetMapping
public List<Player> search(@RequestParam String name) { ... }    // ≈ [FromQuery]

@PostMapping
public Player create(@RequestBody CreatePlayerRequest req) { ... } // ≈ [FromBody]
```

### ResponseEntity for status control
When you need to control status codes (≈ `IActionResult`):
```java
@GetMapping("/{id}")
public ResponseEntity<Player> getById(@PathVariable Long id) {
    return playerService.find(id)
        .map(ResponseEntity::ok)                    // 200 with body
        .orElse(ResponseEntity.notFound().build()); // 404
}
```
`ResponseEntity<T>` ≈ `IActionResult`; `.ok()`/`.notFound()`/`.status(...)` mirror `Ok()`/`NotFound()`/`StatusCode(...)`.

### Exercise 3.1
Build the `HealthController`, plus a `PlayerController` with a hardcoded in-memory list: `GET /api/players`, `GET /api/players/{id}` (404 if missing), `POST /api/players`. Test with curl or your IDE's HTTP client. No database yet — just the HTTP shape.

---

## Chapter 4 — The Data Layer: JPA & PostgreSQL

Spring Data JPA is the Entity Framework equivalent — an ORM (Hibernate under the hood) mapping Java objects to database tables.

### Set up PostgreSQL
Install Postgres locally (or run it in Docker: `docker run -e POSTGRES_PASSWORD=dev -p 5432:5432 postgres`). Create a database `gamedb`.

### Configure the datasource
`application.properties`:
```properties
spring.datasource.url=jdbc:postgresql://localhost:5432/gamedb
spring.datasource.username=postgres
spring.datasource.password=dev
spring.jpa.hibernate.ddl-auto=update    # dev only: auto-create/update tables from entities
spring.jpa.show-sql=true                # log SQL (like EF query logging)
```
`ddl-auto=update` auto-generates tables from your entities — great for dev, like EF's `EnsureCreated`. For production you'd use Flyway/Liquibase migrations instead (Ch.13 note).

### An entity (≈ EF entity)
```java
@Entity
@Table(name = "players")
public class Player {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String username;

    @Column(nullable = false)
    private String email;

    private int totalScore;
    private int matchesPlayed;

    // JPA requires a no-arg constructor
    protected Player() {}

    public Player(String username, String email) {
        this.username = username;
        this.email = email;
    }
    // getters/setters (or use Lombok @Getter/@Setter to generate them)
}
```
`@Entity` = an EF entity; `@Id` + `@GeneratedValue` = `[Key]` with identity; `@Column` ≈ `[Column]`. Note JPA's quirk: it requires a no-arg constructor (can be `protected`). Lombok (`@Getter`/`@Setter`/`@NoArgsConstructor`) eliminates the getter/setter/constructor boilerplate if you add it.

### The repository (≈ DbSet, but you don't write the class)
```java
public interface PlayerRepository extends JpaRepository<Player, Long> {
    Optional<Player> findByUsername(String username);      // Spring generates the query!
    List<Player> findByTotalScoreGreaterThan(int score);
}
```
This is the magic that surprises EF users: **you declare an interface, and Spring generates the implementation at runtime.** `JpaRepository<Player, Long>` gives you `save`, `findById`, `findAll`, `delete`, etc. for free (≈ `DbSet` methods). And **query derivation** — Spring parses method *names* like `findByUsername` and writes the SQL for you. `findByTotalScoreGreaterThan` becomes `WHERE total_score > ?`. For complex queries you drop to `@Query("...")` with JPQL or native SQL.

### Exercise 4.1
Create the `Player` entity and `PlayerRepository`. Wire the repository into `PlayerController` (constructor injection — next chapter) and make the CRUD endpoints hit the real database. Confirm rows appear in Postgres and `findByUsername` works.

---

## Chapter 5 — Services, DI, and Project Structure

### Constructor injection — identical to ASP.NET Core
```java
@Service
public class PlayerService {
    private final PlayerRepository repo;

    public PlayerService(PlayerRepository repo) {   // Spring injects it — no attribute needed
        this.repo = repo;
    }

    public Optional<Player> find(Long id) { return repo.findById(id); }

    public Player register(String username, String email) {
        repo.findByUsername(username).ifPresent(p -> {
            throw new IllegalStateException("username taken");
        });
        return repo.save(new Player(username, email));
    }
}
```
`@Service` marks it for component scanning (≈ `services.AddScoped`). Spring sees the constructor and injects the `PlayerRepository` automatically — **you don't register anything explicitly** (unlike ASP.NET Core's `AddScoped` calls). Component scanning + constructor injection just wires it. This is more implicit than ASP.NET Core; the annotations *are* the registration.

Constructor injection is preferred (over field injection with `@Autowired`) for the same reasons as C#: immutable dependencies, testability, no hidden state.

### Standard project structure
```
com.denis.gamebackend/
├── controller/    # @RestController — HTTP layer
├── service/       # @Service — business logic
├── repository/    # JpaRepository interfaces — data access
├── entity/        # @Entity — domain/DB models
├── dto/           # request/response records (not exposed entities)
└── config/        # @Configuration — security, beans, etc.
```
Same layering as a clean ASP.NET Core solution, just package folders instead of projects.

### DTOs — don't expose entities directly
Use Java `record`s for request/response shapes (≈ C# records for DTOs):
```java
public record CreatePlayerRequest(String username, String email) {}
public record PlayerResponse(Long id, String username, int totalScore) {}
```
Map entities ↔ DTOs in the service or controller. Keeps your DB schema decoupled from your API contract — same discipline you'd apply in ASP.NET Core.

### Exercise 5.1
Refactor: introduce `PlayerService` between controller and repository. Add `CreatePlayerRequest`/`PlayerResponse` records. Controller takes the request DTO, calls the service, returns the response DTO. Confirm the layering works end to end.

---

## Chapter 6 — Accounts: Entities, Validation, Registration

Now build the real accounts feature.

### Validation (≈ FluentValidation / DataAnnotations)
Add constraints to your request DTO:
```java
public record CreatePlayerRequest(
    @NotBlank @Size(min = 3, max = 20) String username,
    @NotBlank @Email String email,
    @NotBlank @Size(min = 8) String password
) {}
```
Then `@Valid` in the controller triggers validation (≈ `[ApiController]`'s automatic model validation):
```java
@PostMapping("/register")
public ResponseEntity<PlayerResponse> register(@Valid @RequestBody CreatePlayerRequest req) {
    ...
}
```
Invalid requests auto-return 400 with error details. `@NotBlank`, `@Email`, `@Size`, `@Min`, etc. mirror DataAnnotations.

### Global exception handling (≈ exception filters / middleware)
```java
@RestControllerAdvice
public class ApiExceptionHandler {
    @ExceptionHandler(IllegalStateException.class)
    public ResponseEntity<Map<String,String>> handleConflict(IllegalStateException e) {
        return ResponseEntity.status(HttpStatus.CONFLICT)
            .body(Map.of("error", e.getMessage()));
    }
}
```
`@RestControllerAdvice` is a cross-cutting exception handler (≈ ASP.NET Core exception-handling middleware / `IExceptionFilter`) — catch service exceptions once, translate to clean HTTP responses.

### Exercise 6.1
Build `POST /api/players/register` with validation, uniqueness check (username + email), and proper error responses (400 for validation, 409 for taken username). Store the password hashed — but *don't* hand-roll hashing; you'll do it properly with Spring Security's `PasswordEncoder` in Ch.9, so stub it for now and wire real hashing there.

---

## Chapter 7 — Leaderboards: Queries, Sorting, Pagination

A leaderboard is a great JPA exercise — sorting, limiting, pagination, aggregate queries.

### Pagination & sorting (built in)
Spring Data gives you `Pageable` for free:
```java
public interface PlayerRepository extends JpaRepository<Player, Long> {
    Page<Player> findAllByOrderByTotalScoreDesc(Pageable pageable);
}
```
Controller:
```java
@GetMapping("/leaderboard")
public Page<PlayerResponse> leaderboard(
        @RequestParam(defaultValue = "0") int page,
        @RequestParam(defaultValue = "20") int size) {
    return playerService.topPlayers(PageRequest.of(page, size))
        .map(p -> new PlayerResponse(p.getId(), p.getUsername(), p.getTotalScore()));
}
```
`Page<T>` returns the results plus total count, page number, etc. — a paginated leaderboard endpoint with almost no code. `PageRequest.of(page, size)` ≈ `.Skip().Take()` but with metadata.

### Custom queries with @Query
For anything derivation can't express, JPQL (entity-oriented) or native SQL:
```java
@Query("SELECT p FROM Player p WHERE p.matchesPlayed > :min ORDER BY p.totalScore DESC")
List<Player> rankedActivePlayers(@Param("min") int minMatches);

@Query(value = "SELECT * FROM players ORDER BY total_score DESC LIMIT 10", nativeQuery = true)
List<Player> top10Native();
```

### Exercise 7.1
Build `GET /api/leaderboard?page=&size=` returning a paginated, score-descending leaderboard. Add a `GET /api/leaderboard/top10` using a custom query. Seed some players and verify ordering + pagination.

---

## Chapter 8 — Matchmaking & Match Results

The game-integration features the Go server will eventually call.

### Match result entity + relationship
```java
@Entity
@Table(name = "matches")
public class Match {
    @Id @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    private Instant playedAt;
    private Long winnerId;

    @ElementCollection
    private List<Long> participantIds;
    // ...
}
```

### Recording results (the Go server posts here)
```java
@PostMapping("/api/matches")
public ResponseEntity<Void> recordMatch(@Valid @RequestBody MatchResult result) {
    matchService.record(result);   // save match, update each player's totalScore/matchesPlayed
    return ResponseEntity.status(HttpStatus.CREATED).build();
}
```
`matchService.record` is a transactional operation — saving the match and updating multiple players' stats atomically:
```java
@Service
public class MatchService {
    @Transactional                     // ≈ EF transaction / SaveChanges scope
    public void record(MatchResult r) {
        matchRepo.save(toEntity(r));
        for (var pid : r.participantIds()) {
            var p = playerRepo.findById(pid).orElseThrow();
            p.setMatchesPlayed(p.getMatchesPlayed() + 1);
            if (pid.equals(r.winnerId())) p.setTotalScore(p.getTotalScore() + 100);
            playerRepo.save(p);
        }
    }
}
```
`@Transactional` wraps the method in a DB transaction — if anything throws, it all rolls back (≈ a single EF `SaveChanges` unit of work, but declarative).

### Matchmaking endpoint
A simple version: `POST /api/matchmaking/queue` adds a player to a queue (a DB table or in-memory structure), returns a match assignment when enough players are queued. Keep it simple — the real-time coordination lives in the Go server; this backend just tracks who's looking for a game and hands out match assignments.

### Exercise 8.1
Build `POST /api/matches` (record a result, update stats transactionally) and a basic `POST /api/matchmaking/queue`. Verify that recording a match updates the leaderboard (stats flow through to Ch.7's endpoint).

---

## Chapter 9 — Spring Security: Real Authentication

Now do auth properly. Spring Security is powerful and initially opaque — here's the minimal real-world setup for a REST API with token auth.

### Password hashing
Spring Security provides `PasswordEncoder`:
```java
@Configuration
public class SecurityBeans {
    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();   // industry-standard hashing
    }
}
```
Inject it into your registration service to hash passwords (`encoder.encode(rawPassword)`) and verify them (`encoder.matches(raw, stored)`). Never store plaintext — this replaces the Ch.6 stub.

### The auth model for a game backend: JWT tokens
For a stateless REST API that a game client and Go server call, **JWT (JSON Web Tokens)** are the standard: the client logs in once, gets a signed token, and includes it on subsequent requests. Stateless, no server session, and the Go server can validate the same token.

The flow:
1. `POST /api/auth/login` with username+password → verify with `PasswordEncoder` → issue a signed JWT.
2. Client sends `Authorization: Bearer <token>` on protected requests.
3. A security filter validates the token on each request and sets the authenticated user.

### Security configuration (Spring Security 6/7 style)
```java
@Configuration
@EnableWebSecurity
public class SecurityConfig {
    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http
            .csrf(csrf -> csrf.disable())                 // stateless API, CSRF off
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/api/auth/**", "/api/players/register").permitAll()
                .requestMatchers("/api/health").permitAll()
                .anyRequest().authenticated())            // everything else needs a token
            .sessionManagement(s -> s.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .addFilterBefore(jwtFilter, UsernamePasswordAuthenticationFilter.class);
        return http.build();
    }
}
```
This is the modern lambda-DSL config (Spring Security 6+ dropped the old `WebSecurityConfigurerAdapter`). `authorizeHttpRequests` ≈ your `[Authorize]`/`[AllowAnonymous]` policy, centralized. `permitAll()` on login/register/health, `authenticated()` everywhere else.

### Method-level authorization (≈ [Authorize] attributes)
```java
@PreAuthorize("hasRole('ADMIN')")
@DeleteMapping("/api/players/{id}")
public void deletePlayer(@PathVariable Long id) { ... }
```
`@PreAuthorize` ≈ `[Authorize(Roles = "Admin")]`.

### Honest note on Spring Security's learning curve
Spring Security is the steepest part of Spring Boot — it's powerful but has a lot of moving parts (filter chains, authentication providers, JWT plumbing). The JWT filter itself (parsing/validating the token, loading the user) is ~40 lines you'll adapt from a reference. Budget real time here; it's the one chapter where "follow a known-good reference implementation closely" beats "figure it out from scratch." A well-regarded reference is the official Spring Security docs plus Baeldung's JWT tutorials — but verify against your Spring Security 7.x version since the API shifts between majors.

### Exercise 9.1
Implement: BCrypt password hashing in registration, `POST /api/auth/login` issuing a JWT, a JWT validation filter, and the security config locking down all endpoints except auth/register/health. Verify: register → login → get token → call a protected endpoint with the Bearer token → 401 without it.

---

## Chapter 10 — Configuration & Profiles

### application.yml and profiles
`application.yml` (≈ `appsettings.json`), with profile-specific overrides (≈ `appsettings.{Environment}.json`):
```yaml
# application.yml (base)
spring:
  application:
    name: game-backend
server:
  port: 8080

# application-dev.yml (dev profile)
spring:
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: true

# application-prod.yml (prod profile)
spring:
  jpa:
    hibernate:
      ddl-auto: validate      # never auto-alter schema in prod
    show-sql: false
```
Activate a profile with `SPRING_PROFILES_ACTIVE=prod` (env var) or `--spring.profiles.active=prod`. Profiles ≈ ASP.NET Core environments.

### Secrets & externalized config
Never hardcode DB passwords or JWT signing keys. Spring reads from environment variables automatically:
```yaml
spring:
  datasource:
    password: ${DB_PASSWORD}       # from env var
app:
  jwt:
    secret: ${JWT_SECRET}
```
`${VAR}` pulls from environment (≈ ASP.NET Core config providers). On your VPS, set these as env vars in the systemd unit.

### Exercise 10.1
Split config into base + dev + prod profiles. Externalize the DB password and JWT secret to env vars. Run under the dev profile locally; confirm prod uses `ddl-auto: validate`.

---

## Chapter 11 — Testing

JUnit 5 (≈ xUnit/NUnit) plus Spring's test support.

### Unit test a service (plain JUnit + Mockito)
```java
@ExtendWith(MockitoExtension.class)
class PlayerServiceTest {
    @Mock PlayerRepository repo;
    @InjectMocks PlayerService service;

    @Test
    void registerRejectsDuplicate() {
        when(repo.findByUsername("denis")).thenReturn(Optional.of(new Player("denis","e")));
        assertThrows(IllegalStateException.class,
            () -> service.register("denis", "e@x.com"));
    }
}
```
`@Mock`/`@InjectMocks` (Mockito) ≈ Moq. `assertThrows` ≈ `Assert.Throws`.

### Integration test a controller (Spring's @SpringBootTest / @WebMvcTest)
```java
@WebMvcTest(PlayerController.class)
class PlayerControllerTest {
    @Autowired MockMvc mvc;
    @MockBean PlayerService service;

    @Test
    void getReturns404WhenMissing() throws Exception {
        when(service.find(99L)).thenReturn(Optional.empty());
        mvc.perform(get("/api/players/99"))
           .andExpect(status().isNotFound());
    }
}
```
`@WebMvcTest` spins up just the web layer with a mock service (≈ ASP.NET Core `WebApplicationFactory` but lighter). `MockMvc` ≈ `TestServer`/`HttpClient` in integration tests.

### Testcontainers for real-DB tests
For repository tests against a real PostgreSQL (not H2), **Testcontainers** spins up a throwaway Postgres in Docker for the test run — the gold standard for DB integration tests, so you test against the real database you deploy on.

### Exercise 11.1
Write: a unit test for `PlayerService.register` (duplicate rejection), a `@WebMvcTest` for the 404 path, and (bonus) a Testcontainers-backed repository test for `findByUsername`.

---

## Chapter 12 — The HTTP Contract for the Go Server

This is the seam to GO.md — the endpoints the Go game server calls into this backend. Define it clearly so the two connect cleanly later (the same way NETWORKING.md and GO.md interlock).

### Endpoints the Go server consumes
```
POST /api/auth/validate
  Body: { "token": "<jwt>" }
  → 200 { "playerId": 123, "username": "denis" }  if valid
  → 401                                            if invalid
  Purpose: Go server validates a client's token when they connect over WS.

POST /api/matches
  Auth: service-to-service token (see below)
  Body: { "winnerId": 123, "participantIds": [123,456], "playedAt": "..." }
  → 201
  Purpose: Go server reports a finished match; backend updates stats/leaderboard.

GET /api/players/{id}/profile
  → 200 { "id":123, "username":"denis", "totalScore":4200, "matchesPlayed":37 }
  Purpose: Go server / client fetches a player's meta profile.

POST /api/matchmaking/queue
  Body: { "playerId": 123 }
  → 200 { "matchId": "...", "serverAddress": "..." }  when matched
  Purpose: client requests a game; backend assigns them to a Go server instance.
```

### Service-to-service auth
The Go server calling `POST /api/matches` shouldn't use a player's token — it's a trusted backend caller. Give the Go server its own **service credential** (a dedicated API key or service JWT) and authorize those endpoints for the service role. Note in the security config which endpoints are player-authenticated vs service-authenticated.

### Client token flow across both systems
1. Client → Spring Boot `POST /api/auth/login` → gets JWT.
2. Client → Go server (WebSocket), sends the JWT on connect.
3. Go server → Spring Boot `POST /api/auth/validate` → confirms identity, gets playerId.
4. Game happens (Go server).
5. Go server → Spring Boot `POST /api/matches` → records result (service auth).

This is the industry-standard split: **Spring Boot owns identity and persistence; the Go server owns real-time gameplay and trusts the backend for auth/stats.**

### Exercise 12.1
Implement `POST /api/auth/validate` and `GET /api/players/{id}/profile`. Write the contract doc (like the endpoint list above) into the repo so future-you (or Claude Code, when wiring the Go server) has the exact shapes. These are the endpoints GO.md's server will call.

---

## Chapter 13 — Deployment

### Build a runnable jar
```
./gradlew bootJar
```
Produces `build/libs/game-backend-1.0.jar` — a single executable jar with embedded Tomcat. Run it anywhere with a JRE:
```
java -jar game-backend-1.0.jar
```
Simpler than it sounds — no external app server, like your Go binary but needs a JVM present.

### On the VPS (same box family as your Go server)
1. Install a JRE (or bundle one — jlink/jpackage can make a self-contained image).
2. Install/provision PostgreSQL (managed DB or on the same box).
3. Run the jar under systemd with env vars for secrets and `SPRING_PROFILES_ACTIVE=prod`.

Minimal systemd unit:
```ini
[Unit]
Description=Game Backend
After=network.target postgresql.service

[Service]
ExecStart=/usr/bin/java -jar /opt/backend/game-backend.jar
Environment=SPRING_PROFILES_ACTIVE=prod
Environment=DB_PASSWORD=...
Environment=JWT_SECRET=...
WorkingDirectory=/opt/backend
Restart=always
User=backend

[Install]
WantedBy=multi-user.target
```

### Production database migrations
In prod, don't use `ddl-auto: update` (risky — Hibernate altering your schema). Use **Flyway** (add the dependency): versioned SQL migration files in `src/main/resources/db/migration/` run automatically on startup, giving you controlled, reviewable schema changes (≈ EF migrations). Set `ddl-auto: validate` so Hibernate only *checks* the schema matches your entities.

### Docker option
A `Dockerfile` (or Spring Boot's built-in `./gradlew bootBuildImage`, which makes an optimized layered image with no Dockerfile) containerizes the app — handy if you later run both services under Docker Compose or Kubernetes (ties into your K8s interest).

### Exercise 13.1
Build the jar, run it locally with the prod profile against your Postgres, add one Flyway migration for the players table, and (bonus) build a container image with `bootBuildImage`.

---

## Appendix: Quick ASP.NET Core → Spring Boot Cheat Sheet

| ASP.NET Core | Spring Boot |
|---|---|
| `dotnet new webapi` | Spring Initializr (start.spring.io) |
| `[ApiController]` | `@RestController` |
| `[HttpGet]`/`[HttpPost]` | `@GetMapping`/`@PostMapping` |
| `[Route]` | `@RequestMapping` |
| `[FromBody]`/`[FromRoute]`/`[FromQuery]` | `@RequestBody`/`@PathVariable`/`@RequestParam` |
| `IActionResult`, `Ok()`, `NotFound()` | `ResponseEntity`, `.ok()`, `.notFound()` |
| `services.AddScoped<I,T>()` | `@Service` + component scan |
| ctor injection | ctor injection (same) |
| `DbContext`/`DbSet<T>` | `JpaRepository<T,ID>` |
| EF entity `[Key]` | `@Entity`/`@Id` |
| EF migrations | Flyway/Liquibase |
| `[Authorize]` | `@PreAuthorize` / SecurityFilterChain |
| ASP.NET Identity | Spring Security + PasswordEncoder + JWT |
| `appsettings.json` | `application.yml` |
| environments | profiles |
| `dotnet watch` | Spring Boot DevTools |
| xUnit + Moq | JUnit 5 + Mockito |
| `WebApplicationFactory` | `@SpringBootTest`/`@WebMvcTest` + MockMvc |
| Kestrel | embedded Tomcat |
| `dotnet publish` | `./gradlew bootJar` |

---

*Prerequisite: JAVA.md Ch.1-14. Then work this in order — Ch.1's translation table is most of the mental model; the rest is applying patterns you know via new annotations. Ch.9 (Security) is the hard part — follow a known-good JWT reference closely. Ch.12 is the seam to the Go server (GO.md). Verify Spring Boot/Security versions at start.spring.io before scaffolding, since Spring ships every 6 months with no LTS.*
