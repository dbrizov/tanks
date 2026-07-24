# Java for the Experienced Developer

*A fast-track guide for someone who already knows C#, C/C++, Python, and Lua.*

This isn't a "what is a variable" book. It assumes you understand OOP, generics, GC, closures, references vs. values, and how a compiler/runtime split works. Every concept is framed as **"here's how this differs from what you already know."** Java's syntax will feel like a slightly more verbose C#, so we skip the obvious and focus on the friction points.

**How to use this guide:** Each chapter has a short explanation, a worked example, and exercises. Type every exercise by hand — don't copy-paste. The friction of typing is where the muscle memory forms. Solutions are at the end of each chapter, but write yours first. Where a topic is a classic certification-exam trap, you'll see a **⚠️ Exam trap** callout — these are the things that bite everyone, drawn from the OCA/OCP study guides (see [Appendix B](#appendix-b--where-this-came-from--going-deeper)).

*This edition has been expanded using a 14-lecture Java Basics course and two Oracle certification study guides (SE 7 "K&B" and SE 8 Sybex). The original was Java-25-modern and skipped a lot of foundational mechanics; those are now folded in. Target runtime is still Temurin OpenJDK 25, but almost everything here is valid on Java 17+.*

---

## Table of Contents

1. [Setup & Your First Program](#chapter-1--setup--your-first-program)
2. [Types, Primitives, and the Boxing Trap](#chapter-2--types-primitives-and-the-boxing-trap)
3. [Strings, StringBuilder & Formatting](#chapter-3--strings-stringbuilder--formatting)
4. [Flow Control — the 5% That Differs](#chapter-4--flow-control--the-5-that-differs)
5. [Classes, Interfaces, and Records](#chapter-5--classes-interfaces-and-records)
6. [Constructors, Initialization & Access Control](#chapter-6--constructors-initialization--access-control)
7. [Enums (They're Real Classes)](#chapter-7--enums-theyre-real-classes)
8. [equals(), hashCode(), and Ordering](#chapter-8--equals-hashcode-and-ordering)
9. [Generics and Wildcards](#chapter-9--generics-and-wildcards)
10. [Collections Framework](#chapter-10--collections-framework)
11. [The Streams API (Your LINQ)](#chapter-11--the-streams-api-your-linq)
12. [Exceptions & Checked Exceptions](#chapter-12--exceptions--checked-exceptions)
13. [Optional, Null, and Defensive Code](#chapter-13--optional-null-and-defensive-code)
14. [Lambdas & Functional Interfaces](#chapter-14--lambdas--functional-interfaces)
15. [Nested & Anonymous Classes](#chapter-15--nested--anonymous-classes)
16. [Pattern Matching & Sealed Types](#chapter-16--pattern-matching--sealed-types)
17. [Dates, Times & Numbers](#chapter-17--dates-times--numbers)
18. [Files & I/O (NIO.2)](#chapter-18--files--io-nio2)
19. [Concurrency: Threads to Virtual Threads](#chapter-19--concurrency-threads-to-virtual-threads)
20. [Build Tools: Gradle](#chapter-20--build-tools-gradle)
21. [Capstone Project](#chapter-21--capstone-project)
- [Appendix A: Quick C# → Java Cheat Sheet](#appendix-a-quick-c--java-cheat-sheet)
- [Appendix B: Where This Came From & Going Deeper](#appendix-b--where-this-came-from--going-deeper)

---

## Chapter 1 — Setup & Your First Program

You've already installed Temurin OpenJDK 25 and the VS Code Java extension pack, so you're ready.

### The mental model

| C# | Java |
|---|---|
| `.cs` → IL → CLR JIT | `.java` → bytecode (`.class`) → JVM JIT |
| `namespace Foo {}` | `package foo;` (must match folder structure) |
| `dotnet run` | `java Main.java` (single-file mode, Java 11+) |
| `Console.WriteLine` | `System.out.println` |

One rule that trips up everyone from C#: **the public class name must match the filename exactly.** `Main.java` must contain `public class Main`. This is enforced by the compiler, not a convention. A single `.java` file can hold several classes, but **only one may be `public`**, and it names the file.

**Packages are folders.** `package com.denis.app;` means the file lives under `com/denis/app/`. This mirroring is enforced when you build with a tool. To use a public class from another package you either write its fully-qualified name or `import com.denis.app.Thing;` (and `import static` to pull in static members like `import static java.lang.Math.PI;`).

### Hello World

```java
public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, Java");
    }
}
```

Since Java 25 you can actually write a more compact form (compact source files with instance main methods), but learn the classic form first — it's what every codebase and tutorial uses:

```java
// Java 25 compact form — nice, but not yet universal
void main() {
    IO.println("Hello, Java");
}
```

Stick with the classic `public static void main(String[] args)` for now so nothing looks alien when you read real code. Note it has several legal spellings — `static public void main(...)`, `public static void main(String... args)`, `main(String args[])` — because `static`/`public` order is free, varargs is allowed, and C-style array brackets are legal (don't write them, but you'll see them on exams).

### Running it

In VS Code, click the **Run** codelens above `main()`, or from a terminal:

```
java Main.java
```

That single-file mode compiles and runs in one step — no explicit `javac` needed for quick scripts. The old two-step is `javac Main.java` (produces `Main.class`) then `java Main` (no extension, no `.class`). For anything multi-file, you'll use Gradle (Chapter 20).

### Exercise 1.1
Write a program that prints the numbers 1–10, but "Fizz" for multiples of 3, "Buzz" for multiples of 5, "FizzBuzz" for both. Yes, FizzBuzz — but do it to feel the `for` loop and `%` (identical to C#) and `System.out.println` rhythm.

### Exercise 1.2
Print `args` back to the user, one per line. Run it with `java Main.java one two three` and observe how CLI args arrive. Note: unlike C, `args[0]` is the **first argument**, not the program name.

---

## Chapter 2 — Types, Primitives, and the Boxing Trap

This is the single biggest day-one gotcha coming from C#.

### Primitives are NOT objects

Java has 8 primitive types: `byte` (8-bit), `short` (16-bit), `int` (32-bit), `long` (64-bit), `float` (32-bit), `double` (64-bit), `char` (16-bit), `boolean`. These are **not** objects — they have no methods, can't be `null`, and can't be used as generic type arguments. All numeric primitives are **signed** (see below), and each has a fixed default value (`0`, `0.0`, `false`, `'\u0000'`) that fields/arrays get automatically but **local variables do not** — a local must be assigned before use or the compiler rejects it.

Each primitive has a **boxed wrapper class**: `Integer`, `Long`, `Double`, `Boolean`, `Character`, etc.

```java
int a = 5;              // primitive, lives on the stack
Integer b = 5;          // boxed object, autoboxed from int
List<Integer> nums;     // MUST use Integer — List<int> does NOT compile
```

In C#, `int` *is* `System.Int32` — a struct, unified under `object`. In Java the divide is hard: `int` and `Integer` are genuinely different things, and the autoboxing that bridges them is a silent performance cost and a source of subtle bugs.

### The `==` trap (read this twice)

For primitives, `==` compares values (like C#). For **objects**, `==` compares references, NOT values. This includes boxed types and strings:

```java
Integer x = 1000, y = 1000;
x == y            // false! Two different objects
x.equals(y)       // true — value comparison

String s1 = new String("hi"), s2 = new String("hi");
s1 == s2          // false — different objects
s1.equals(s2)     // true
```

**Rule: use `.equals()` for object content comparison, always. Reserve `==` for primitives and reference-identity checks.** This is unlike C# where `==` is often overloaded to do the sensible thing (e.g. on `string`). Java does not overload operators at all.

> ⚠️ **Exam trap — the Integer cache.** Autoboxing caches `Integer` values from **−128 to 127**. So `Integer a = 100, b = 100; a == b` is `true`, but `Integer a = 200, b = 200; a == b` is `false`. Same code, different answer depending on the magnitude. This inconsistency is why you never use `==` on boxed types.

### Widening, narrowing, and casts

Conversions between numeric types follow C#-like rules, with Java vocabulary:

```java
int i = 5;
long l = i;            // widening — automatic (small → big container)
double d = l;          // widening — automatic
int back = (int) d;    // narrowing — explicit cast required, may lose data
byte bt = (byte) 300;  // narrowing wraps: 300 → 44 (silent overflow!)
```

- **Widening** (byte→short→int→long→float→double) is implicit.
- **Narrowing** requires an explicit `(cast)` and can silently truncate/overflow.
- `char` is an unsigned 16-bit UTF-16 code unit that participates in integer arithmetic: `'a' + 1` is `98` (an `int`), and `(char)('a' + 1)` is `'b'`.
- Integer overflow **wraps silently** — `Integer.MAX_VALUE + 1` is `Integer.MIN_VALUE`. Use `Math.addExact(...)` when you want an exception instead. (For money, use `BigDecimal` — Chapter 17.)

### No unsigned types

There's no `uint`/`ulong` in the C# sense. `long` is your 64-bit signed workhorse. (There are static helper methods like `Integer.toUnsignedString` and `Long.divideUnsigned` for the rare case, but there's no unsigned type.)

### var works, with limits

```java
var list = new ArrayList<String>();  // fine, Java 10+
var x = 5;                            // inferred int
// var y;                            // ERROR — needs an initializer
```

Same idea as C# `var`, only usable for local variables with an initializer.

### Exercise 2.1
Create two `Integer` variables both set to `500`. Print the result of `==` and `.equals()`. Then do the same with the value `100`. Explain (in a comment) why the `100` case behaves differently. *(Hint: Integer cache, −128..127.)*

### Exercise 2.2
Write a method `sum(List<Integer> nums)` that returns the total as an `int`. Notice where autoboxing/unboxing happens. Add a comment on each line where a box or unbox occurs.

---

## Chapter 3 — Strings, StringBuilder & Formatting

Strings feel like C# `string`, with two Java-specific wrinkles: the **string pool** and the fact that `String` has no operator support beyond `+`.

### Immutability and the string pool

`String` is immutable and `final` (can't be subclassed). Every method that "changes" a string returns a new one:

```java
String s = "abc";
s.concat("def");        // returns "abcdef" but DISCARDS it
System.out.println(s);  // still "abc"
s = s.concat("def");    // must reassign
```

Java keeps a **string pool** (an intern table). String *literals* with the same content share one object; `new String("abc")` deliberately creates an off-pool object. This is the mechanism behind the `==` trap from Chapter 2:

```java
String a = "abc";
String b = "abc";
String c = new String("abc");
a == b            // true  — both point at the pooled literal
a == c            // false — c is a fresh object
a.equals(c)       // true  — always compare content with equals()
a == c.intern()   // true  — intern() returns the pooled instance
```

C# also interns literals, but you rarely notice because `==` on `string` compares content. In Java the pool is visible precisely because `==` does not.

> **Rule: always use `.equals()` to compare string content — treat `==` on strings as a bug** unless you can articulate why you need reference identity (you almost never do). `==` only *accidentally* works when both sides are the same pooled object; the moment one comes from `new String(...)`, user input, a file, a network call, or runtime concatenation, `==` silently returns `false` for identical text. Refinements: put the literal on the left (`"yes".equals(input)`) or use `Objects.equals(a, b)` to stay null-safe, and use `.equalsIgnoreCase(...)` instead of lowercasing both sides. (This is the string-specific case of the general `==` trap in Chapter 2 and the equals contract in Chapter 8.)

### StringBuilder for mutation

Because `String` is immutable, building a string in a loop with `+` allocates a new object each iteration. Use `StringBuilder` (Java's `System.Text.StringBuilder`) — it's mutable:

```java
StringBuilder sb = new StringBuilder();
for (int i = 0; i < 5; i++) {
    sb.append(i).append(",");
}
sb.setLength(sb.length() - 1);   // drop trailing comma
String csv = sb.toString();       // "0,1,2,3,4"
```

`StringBuilder` also has `insert`, `delete`, `reverse`, `replace`. (There's a thread-safe `StringBuffer` too — ignore it unless you truly share one across threads.)

### Useful String methods

`length()`, `charAt(i)`, `substring(from, to)` (to is exclusive), `indexOf`, `contains`, `startsWith`/`endsWith`, `toUpperCase`/`toLowerCase`, `trim()`/`strip()` (`strip` is Unicode-aware, Java 11+), `isBlank()`, `replace`, `split(regex)`, `repeat(n)`, `String.join(sep, parts)`. Note `split` takes a **regex**, so `"a.b.c".split(".")` gives nothing useful — you need `split("\\.")`.

### Formatting

```java
String.format("%s is %d years old, %.2f%% done", name, age, pct);
System.out.printf("%-10s|%5d%n", "left", 42);   // like printf; %n = platform newline
```

`%s` string, `%d` integer, `%f` float (`%.2f` = 2 decimals), `%n` newline, `%%` a literal percent. This is your `$"{x}"` interpolation replacement — Java has no string interpolation operator (yet), so it's `format`/`printf` or `+` concatenation.

**Text blocks** (Java 15+) are multi-line string literals, like C# raw string literals:

```java
String json = """
    {
      "name": "%s",
      "age": %d
    }
    """.formatted(name, age);   // .formatted() == String.format on the receiver
```

### Exercise 3.1
Read a sentence (hardcode a `String`), reverse the **order of the words** (not the characters), and print the result. Build it with a `StringBuilder`.

### Exercise 3.2
Given `String[] rows` of names, print them as a right-aligned column exactly 15 characters wide using `printf("%15s%n", ...)`. Then produce the same output as a single text block joined with `String.join`.

---

## Chapter 4 — Flow Control — the 5% That Differs

`if/else`, `while`, `do-while`, `for`, `for-each`, and `?:` are identical to C#. Only a few things are worth flagging.

### No truthy integers

The condition of `if`/`while`/`for` **must be a `boolean`**. `if (x)` where `x` is an `int` does not compile — a classic C-habit error the compiler catches for you.

### switch has two eras

Classic (statement) switch still fall-through with `break`, and works on `int`/`byte`/`short`/`char`, their wrappers, `enum`, and — since Java 7 — `String`:

```java
switch (day) {
    case "MON": case "TUE":
        System.out.println("early week");
        break;                       // forget this and you fall through!
    default:
        System.out.println("other");
}
```

> ⚠️ **Exam trap.** A classic `case` label must be a **compile-time constant** (a literal or a `static final` set at compile time), and its type must match the selector. `long` is not a valid switch selector.

Modern **arrow switch** (Java 14+) has no fall-through and can be an expression that produces a value:

```java
int len = switch (day) {
    case "MON", "TUE", "WED", "THU", "FRI" -> 5;
    case "SAT", "SUN" -> 2;
    default -> throw new IllegalArgumentException(day);
};
```

Use `yield` when an arrow branch needs a block:

```java
String label = switch (score / 10) {
    case 10, 9 -> "A";
    default -> { String s = compute(score); yield s; }
};
```

Pattern-matching switch (on types) is Chapter 16.

### Labeled break/continue

Java **does** have labeled loops, which C# lacks — handy for breaking out of nested loops without a flag:

```java
outer:
for (int i = 0; i < rows; i++) {
    for (int j = 0; j < cols; j++) {
        if (grid[i][j] == target) break outer;   // jumps clear out of both loops
    }
}
```

`continue label;` restarts the labeled loop's next iteration.

### Exercise 4.1
Write a method that finds the first duplicated value in a 2-D `int[][]` and returns its `[row, col]`, using a labeled `break` to exit both loops. Return `null` (or an `int[]{-1,-1}`) if none.

### Exercise 4.2
Rewrite a small `if/else if` chain that maps an `int` grade band (0–10) to a letter as an **arrow switch expression**.

---

## Chapter 5 — Classes, Interfaces, and Records

### Classes

Familiar territory, with syntax notes:

```java
public class Account {
    private final String owner;   // final = readonly in C#
    private double balance;

    public Account(String owner, double balance) {
        this.owner = owner;
        this.balance = balance;
    }

    // No property syntax — write getters/setters by hand
    public String getOwner() { return owner; }
    public double getBalance() { return balance; }

    public void deposit(double amount) {
        if (amount <= 0) throw new IllegalArgumentException("must be positive");
        balance += amount;
    }
}
```

Key differences from C#:
- **No properties.** Write `getX()`/`setX()` by hand, or use the Lombok library's `@Getter`/`@Setter` to generate them. Java conventions expect the `get`/`set` naming — frameworks (including Spring) rely on it via reflection.
- `final` field ≈ C# `readonly`.
- Single inheritance for classes (`extends`), multiple interface implementation (`implements`) — same as C#.
- **Every method is virtual by default.** Where C# needs `virtual`/`override`, Java methods are overridable unless marked `final`. `@Override` is an optional annotation (compiler-checked — use it always; it catches typos in method signatures). Details and override rules are in Chapter 6.
- `this` refers to the current instance; `super` refers to the superclass (C# `base`).

### Interfaces

```java
public interface Shape {
    double area();
    double perimeter();

    // default methods — like C# default interface methods (Java 8+)
    default String describe() {
        return "area=" + area() + ", perimeter=" + perimeter();
    }

    // interface fields are implicitly public static final (constants)
    double PI = 3.14159;
}
```

An interface can `extends` multiple other interfaces. Before Java 8 it could only declare abstract methods and constants; since Java 8 it can carry `default` and `static` methods (and since 9, `private` helper methods).

### Records — your new best friend

Java `record` is nearly identical to C# `record`. Immutable data carrier, auto-generates constructor, accessors, `equals`, `hashCode`, `toString`:

```java
public record Point(double x, double y) {
    // compact constructor for validation
    public Point {
        if (Double.isNaN(x) || Double.isNaN(y))
            throw new IllegalArgumentException("NaN coord");
    }

    // you can add methods
    double distanceTo(Point other) {
        double dx = x - other.x, dy = y - other.y;
        return Math.sqrt(dx*dx + dy*dy);
    }
}
```

Note the accessor is `point.x()` — a method call, **not** a field access `point.x` and not a property. Records generate accessor methods named after the component, and they give you correct `equals`/`hashCode` for free (Chapter 8 explains why that matters so much).

### Exercise 5.1
Build the classic hierarchy: interface `Shape` with `area()` and `perimeter()`; classes `Circle` and `Rectangle` implementing it; a `record Point(double x, double y)`. In `main`, put a `Circle` and `Rectangle` into a `List<Shape>`, loop over them, and print each one's area formatted to 2 decimal places using `String.format("%.2f", value)`.

### Exercise 5.2
Add a `default` method `boolean isLargerThan(Shape other)` to the `Shape` interface that compares areas. Test it.

---

## Chapter 6 — Constructors, Initialization & Access Control

C# and Java agree on the big ideas here but differ in the fiddly rules the exams love — and in one genuinely new construct (initializer blocks).

### Constructor chaining: `this(...)` and `super(...)`

The **first statement** of a constructor is a call to another constructor: `this(...)` (another constructor of the same class) or `super(...)` (a superclass constructor). If you write neither, the compiler inserts `super()` — a call to the **no-arg** superclass constructor:

```java
class Animal {
    Animal()          { System.out.println("Animal()"); }
    Animal(String s)  { System.out.println("Animal(" + s + ")"); }
}
class Horse extends Animal {
    Horse() {
        super("from Horse");   // must be first; otherwise super() is auto-inserted
        System.out.println("Horse()");
    }
    Horse(int legs) {
        this();                // delegate to Horse()
        System.out.println("Horse(int)");
    }
}
```

> ⚠️ **Exam trap.** If a superclass declares *only* a constructor with arguments (no no-arg one), then every subclass constructor **must** explicitly call `super(args)` — the compiler can't insert `super()` because it doesn't exist. This is a common "won't compile" question.

Other rules: the compiler generates a **default no-arg constructor only if you write no constructor at all**. Constructors are **not inherited**. A constructor can't be `static`, `final`, or `abstract`.

### Initialization order

When you `new` an object, initialization runs in this order (superclass fully before subclass):

1. Static fields and `static { }` blocks — **once**, when the class is first loaded.
2. Instance fields and instance `{ }` blocks — in source order, each time an object is created.
3. The constructor body.

```java
class Demo {
    static int s = log("static field");
    static { log("static block"); }
    int i = log("instance field");
    { log("instance block"); }
    Demo() { log("constructor"); }
    static int log(String m) { System.out.println(m); return 0; }
}
// new Demo(); new Demo();  → static field/block once, then the instance items per object
```

Instance initializer blocks (`{ ... }`) have no C# equivalent; they run before the constructor body and are occasionally used to share setup across overloaded constructors. `final` instance fields must be assigned exactly once — inline, in an instance block, or in every constructor path.

### Access control — four levels, not two

| Modifier | Same class | Same package | Subclass (other pkg) | Everywhere | C# analog |
|---|:---:|:---:|:---:|:---:|---|
| `private` | ✅ | ❌ | ❌ | ❌ | `private` |
| *(none)* — package-private | ✅ | ✅ | ❌ | ❌ | `internal`-ish |
| `protected` | ✅ | ✅ | ✅ | ❌ | *(no exact match)* |
| `public` | ✅ | ✅ | ✅ | ✅ | `public` |

Two things surprise C# developers:
- **The default (no keyword) is package-private**, not private. Leaving off a modifier means "visible to everything in the same package."
- **`protected` in Java also grants package access**, and it grants subclass access even across packages. It's broader than C#'s `protected`.

### static members

`static` fields/methods belong to the class, not instances — like C#. Access them via the class name (`Math.max(...)`, `Integer.parseInt(...)`). A `static` method can't use `this` or instance members directly. **Static methods are not polymorphic**: redeclaring a static method in a subclass *hides* it (resolved by the reference's compile-time type), it doesn't override it.

### Varargs

Java's `params`:

```java
int sum(int... nums) {          // nums is an int[]
    int total = 0;
    for (int n : nums) total += n;
    return total;
}
sum();  sum(1, 2, 3);  sum(new int[]{1, 2});
```

The varargs parameter must be **last**. That's exactly why `main(String... args)` is legal.

### Pass-by-value — always

**Java is strictly pass-by-value.** There is no `ref`/`out`. For objects, the *reference* is passed by value — a copy of the pointer:

```java
void rename(Account a) {
    a.setOwner("Bob");        // ✅ mutates the shared object — visible to caller
    a = new Account("X", 0);  // ❌ only reassigns the local copy — caller unaffected
}
```

So a method can change the state of the object you handed it, but it can never make your variable point somewhere else. Primitives are copied outright.

### Exercise 6.1
Write a class with three overloaded constructors chained via `this(...)`, a static field, a `static` block, an instance field, and an instance block, each printing a message. Create two instances and predict the exact output order before running.

### Exercise 6.2
Write `void swap(int[] pair)` that swaps `pair[0]` and `pair[1]`, and `void swap(int a, int b)` that tries to swap two ints. Call both and explain in a comment why only the array version "works" (pass-by-value).

---

## Chapter 7 — Enums (They're Real Classes)

Java enums are far more powerful than C# enums. They're full classes that can have fields, constructors, and methods.

```java
public enum Planet {
    MERCURY(3.303e+23, 2.4397e6),
    EARTH(5.976e+24, 6.37814e6);

    private final double mass;
    private final double radius;

    Planet(double mass, double radius) {   // constructor is implicitly private
        this.mass = mass;
        this.radius = radius;
    }

    double surfaceGravity() {
        return 6.67300E-11 * mass / (radius * radius);
    }
}
```

Unlike C# where an enum is just a named integer, each Java enum constant is a **singleton object**. The constructor can't be called with `new` — the constants (`MERCURY(...)`) invoke it. Every enum gets `values()` (all constants), `valueOf(String)`, `name()`, and `ordinal()` for free, and enums can be used in `switch` and as `EnumSet`/`EnumMap` keys. You can even give each constant its own method implementation (constant-specific bodies), which makes enums the idiomatic way to build small state machines or strategy tables:

```java
public enum Operation {
    PLUS   { public double apply(double a, double b) { return a + b; } },
    MINUS  { public double apply(double a, double b) { return a - b; } },
    TIMES  { public double apply(double a, double b) { return a * b; } },
    DIVIDE { public double apply(double a, double b) { return a / b; } };

    public abstract double apply(double a, double b);
}
```

### Exercise 7.1
Build the `Operation` enum above and drive it from a loop that prints `a OP b = result` for every constant using `Operation.values()`.

---

## Chapter 8 — equals(), hashCode(), and Ordering

This chapter has no direct C# analog worth leaning on — the contracts are strict and the collections framework depends on them. This is the most common source of "why is my object not found in the HashSet" bugs.

### The default is identity

Every class inherits `equals(Object)` and `hashCode()` from `Object`. The defaults compare **references** (same as `==`) and derive the hash from the object's identity. So two "equal-looking" objects are unequal unless you override these:

```java
var a = new Point2D(1, 2);   // a plain class, NOT a record
var b = new Point2D(1, 2);
a.equals(b);   // false — default identity comparison
```

### The equals contract

Your `equals` must be:
- **Reflexive:** `x.equals(x)` is true.
- **Symmetric:** `x.equals(y)` ⇔ `y.equals(x)`.
- **Transitive:** if `x.equals(y)` and `y.equals(z)`, then `x.equals(z)`.
- **Consistent:** repeated calls return the same result (given unchanged state).
- `x.equals(null)` is **false** (never throws).

### The hashCode contract

- If `x.equals(y)`, then `x.hashCode() == y.hashCode()` **must** hold.
- Unequal objects *may* share a hash (collision) — but a good hash spreads objects across buckets. Returning a constant is legal but turns a `HashMap` into a linked list.
- **Always override both together.** `HashMap`/`HashSet` first pick a bucket by `hashCode`, then find the exact object with `equals`. Override only one and lookups break.

Do it with `java.util.Objects`:

```java
public class Point2D {
    private final int x, y;
    public Point2D(int x, int y) { this.x = x; this.y = y; }

    @Override public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof Point2D p)) return false;   // instanceof pattern (Ch16)
        return x == p.x && y == p.y;
    }
    @Override public int hashCode() { return Objects.hash(x, y); }
}
```

**Or just use a `record`** (Chapter 5) — it generates a correct, matching `equals`/`hashCode` for all components automatically. This is the number-one reason to reach for records for value types.

### Ordering: Comparable vs Comparator

Two interfaces, mirroring C#'s `IComparable<T>` and `IComparer<T>`:

- **`Comparable<T>`** — a type's *natural* order, via `int compareTo(T o)` (negative / zero / positive). Needed by `TreeSet`, `TreeMap`, `Collections.sort`, and `Arrays.binarySearch`.
- **`Comparator<T>`** — an *external*, swappable order, built fluently:

```java
class Person implements Comparable<Person> {
    String name; int age;
    public int compareTo(Person o) { return Integer.compare(this.age, o.age); }
}

// external orderings, composed like LINQ OrderBy/ThenBy:
Comparator<Person> byName = Comparator.comparing(p -> p.name);
Comparator<Person> byAgeThenName =
    Comparator.comparingInt((Person p) -> p.age)
              .thenComparing(p -> p.name)
              .reversed();

people.sort(byAgeThenName);
```

> ⚠️ **Exam trap.** `binarySearch` requires the collection to already be **sorted by the same ordering** you search with. If you sorted with a custom `Comparator`, you must pass the *same* comparator to `binarySearch`, or the result is undefined.

### Exercise 8.1
Give a plain `Book(String title, String author, int year)` class correct `equals`/`hashCode`. Put duplicates into a `HashSet` and confirm dedup works. Then delete your `equals`/`hashCode` and convert it to a `record` — confirm identical behavior.

### Exercise 8.2
Sort a `List<Person>` by age ascending, then by name ascending as a tiebreak, using `Comparator.comparing(...).thenComparing(...)`. Then reverse it.

---

## Chapter 9 — Generics and Wildcards

Generics look like C# but erase at runtime (**type erasure**) — a major conceptual difference.

### Type erasure

In C#, generics are reified: `typeof(List<int>)` is distinct from `List<string>` at runtime. In Java, generic type info is **erased** after compilation — at runtime `List<String>` and `List<Integer>` are both just `List`. Consequences:

- You **cannot** write `new T[]`, `T.class`, or `if (x instanceof List<String>)`.
- You cannot overload methods that differ only by generic parameter type.
- Runtime reflection can't recover the type argument (mostly).

Under the hood the JVM works with raw types exactly as it did before Java 5; the compiler inserts the casts and checks for you. That's why you can accidentally mix a raw `List` with a generic one and only get a warning.

This is the single biggest "why doesn't this work like C#" surprise. Keep it in mind whenever the compiler complains about generics.

### Wildcards: `? extends` and `? super`

Java uses wildcards for variance at the **use site**, where C# uses `in`/`out` at the declaration site:

```java
// "producer extends" — read from it
double sumOf(List<? extends Number> nums) { /* can read Number, can't add */ }

// "consumer super" — write into it
void addInts(List<? super Integer> list) { list.add(42); }
```

Mnemonic: **PECS — Producer Extends, Consumer Super.** If a collection produces values you read, use `? extends`. If it consumes values you write, use `? super`. Note that with `? extends Animal` you may **read** `Animal`s but may not **add** anything (the compiler can't know the exact subtype), and with `? super Integer` you may add `Integer`s but reads come back as `Object`.

> ⚠️ **Exam trap — generics aren't covariant, arrays are.** `List<Dog>` is **not** a `List<Animal>`, so you can't assign one to the other (this is what wildcards are for). Arrays *are* covariant: `Animal[] a = new Dog[3];` compiles — but storing a `Cat` into it throws `ArrayStoreException` at runtime. Generics were designed to move that error to compile time.

### Generic methods and bounds

```java
<T extends Comparable<T>> T max(List<T> list) { ... }   // bounded type parameter
static <K, V> Map<V, K> invert(Map<K, V> in) { ... }     // multiple type params
```

### Exercise 9.1
Write a generic method `<T extends Comparable<T>> T max(List<T> list)` returning the maximum element. Test with a `List<Integer>` and a `List<String>`.

### Exercise 9.2
Write `void copy(List<? extends T> src, List<? super T> dst)` that copies every element. Notice how PECS makes both the read and the write type-check.

---

## Chapter 10 — Collections Framework

Your mapping from C#:

| C# | Java (interface / impl) | Notes |
|---|---|---|
| `List<T>` | `List<T>` / `ArrayList<T>` | Program to the interface |
| `Dictionary<K,V>` | `Map<K,V>` / `HashMap<K,V>` | |
| `HashSet<T>` | `Set<T>` / `HashSet<T>` | |
| `Queue<T>` | `Queue<T>`/`Deque<T>` / `ArrayDeque<T>` | |
| `SortedDictionary` | `TreeMap` | Red-black tree, keys sorted |
| `SortedSet` | `TreeSet` | |
| `LinkedList<T>` | `LinkedList<T>` | |
| `PriorityQueue<T>` | `PriorityQueue<T>` | Not FIFO — orders by `Comparable`/`Comparator` |

Idiom: **declare by interface, instantiate by implementation.**

```java
List<String> names = new ArrayList<>();   // <> = diamond operator, infers type
Map<String, Integer> counts = new HashMap<>();

names.add("Denis");
counts.put("a", 1);
counts.getOrDefault("b", 0);                 // like TryGetValue-ish convenience
counts.computeIfAbsent("c", k -> new ...);   // very handy idiom
counts.merge("a", 1, Integer::sum);          // add-or-increment in one call
```

### Iterating a Map

```java
for (Map.Entry<String, Integer> e : counts.entrySet()) {
    System.out.println(e.getKey() + " = " + e.getValue());
}
// or
counts.forEach((k, v) -> System.out.println(k + " = " + v));
```

### Which Set/Map ordering?

- `HashSet`/`HashMap` — no order, O(1), needs correct `equals`/`hashCode` (Chapter 8).
- `LinkedHashSet`/`LinkedHashMap` — insertion order preserved.
- `TreeSet`/`TreeMap` — sorted; elements/keys must be `Comparable` or you supply a `Comparator`. `TreeSet` adds navigation: `first`, `last`, `lower`, `higher`, `floor`, `ceiling`.

> ⚠️ **Exam trap.** Put an object into a `HashSet` with no `hashCode`/`equals` override and you'll never find it again by a freshly-constructed equal key. Put a non-`Comparable` object into a `TreeSet` and you get a `ClassCastException` at runtime, not compile time.

### Helper classes

`Collections` (`sort`, `reverse`, `shuffle`, `unmodifiableList`, `binarySearch`, `min`/`max`) and `Arrays` (`sort`, `asList`, `fill`, `copyOf`, `toString`, `stream`). `List.of(...)`, `Set.of(...)`, `Map.of(...)` create **immutable** collections (throw on mutation) — use `new ArrayList<>(...)` when you need to modify.

### Exercise 10.1
Read a sentence (hardcode a `String`), split it on spaces, and build a `Map<String, Integer>` word-frequency count. Use `computeIfAbsent` or `merge`. Print sorted by count descending. *(You'll need Chapter 11's streams for the sort — or do it the manual way first with a `List<Map.Entry>` and a `Comparator`, then redo it with streams later.)*

---

## Chapter 11 — The Streams API (Your LINQ)

This is where you'll feel most at home conceptually — and most annoyed syntactically. Streams are LINQ, but more verbose and with different names.

| LINQ | Java Stream |
|---|---|
| `.Where(x => ...)` | `.filter(x -> ...)` |
| `.Select(x => ...)` | `.map(x -> ...)` |
| `.SelectMany` | `.flatMap` |
| `.OrderBy` | `.sorted(Comparator...)` |
| `.First()` | `.findFirst()` (returns `Optional`) |
| `.Any(...)` | `.anyMatch(...)` |
| `.All(...)` | `.allMatch(...)` |
| `.ToList()` | `.collect(Collectors.toList())` or `.toList()` (Java 16+) |
| `.Sum()` | `.mapToInt(...).sum()` |
| `.Distinct()` | `.distinct()` (uses `equals`) |
| `.GroupBy` | `.collect(Collectors.groupingBy(...))` |

```java
import java.util.List;
import java.util.stream.Collectors;

List<String> names = List.of("Denis", "Ana", "Bob", "Alice");

var result = names.stream()
    .filter(n -> n.length() > 3)
    .map(String::toUpperCase)          // method reference, like C# method group
    .sorted()
    .collect(Collectors.toList());
// [ALICE, DENIS]
```

Key differences from LINQ:
- Streams are **single-use** — once a terminal operation runs, the stream is consumed and can't be reused. LINQ is re-enumerable; Java streams are not.
- **Terminal** operations (`collect`, `forEach`, `count`, `findFirst`, `reduce`, `min`/`max`, `anyMatch`) trigger evaluation. **Intermediate** operations (`filter`, `map`, `sorted`, `distinct`, `limit`, `peek`) are lazy — nothing runs until a terminal operation pulls elements through. A pipeline with no terminal operation does nothing at all.
- Laziness has teeth on **infinite streams** (`Stream.iterate`, `Stream.generate`): `filter().findFirst()` terminates, but `sorted()` or `count()` on an infinite stream hangs forever because they must see every element.

### `groupingBy` and reductions

```java
Map<Integer, List<String>> byLength = names.stream()
    .collect(Collectors.groupingBy(String::length));

Map<Integer, Long> countByLength = names.stream()
    .collect(Collectors.groupingBy(String::length, Collectors.counting()));

double total = expenses.stream().mapToDouble(Expense::amount).sum();
Optional<String> longest = names.stream().max(Comparator.comparingInt(String::length));
```

### Primitive streams

`IntStream`, `LongStream`, `DoubleStream` avoid boxing and add `sum()`, `average()`, `range()`, `summaryStatistics()`. Convert with `mapToInt`/`mapToObj`, and note their optionals are `OptionalInt`/`OptionalDouble`/`OptionalLong` (Chapter 13).

### Exercise 11.1
Redo Exercise 10.1 (word frequency) entirely with streams: split, `Collectors.groupingBy` with `Collectors.counting()`, then sort entries by value descending and print the top 3.

### Exercise 11.2
Given a `List<Point>` (your record from Chapter 5), use streams to find the point farthest from the origin. Return an `Optional<Point>`.

---

## Chapter 12 — Exceptions & Checked Exceptions

Exceptions work like C#, **except** Java has **checked exceptions** — a category the compiler forces you to handle or declare.

### The hierarchy

Everything throwable descends from `Throwable`, which splits into:
- **`Error`** — serious, unrecoverable JVM problems (`OutOfMemoryError`, `StackOverflowError`). Don't catch these.
- **`Exception`** — recoverable problems. Its subclasses (except `RuntimeException`) are **checked**: `IOException`, `SQLException`, etc. The compiler forces catch-or-declare.
  - **`RuntimeException`** — **unchecked**: `NullPointerException`, `IllegalArgumentException`, `ClassCastException`, `ArrayIndexOutOfBoundsException`. No compiler enforcement, like all C# exceptions.

```java
// IOException is checked — this won't compile without handling
void readFile(String path) throws IOException {   // declare it...
    Files.readString(Path.of(path));
}

// ...or catch it
try {
    Files.readString(Path.of("x.txt"));
} catch (IOException e) {
    System.err.println("failed: " + e.getMessage());
}
```

Checked exceptions will feel bureaucratic coming from C#, which has none. The idiom is: catch what you can genuinely handle, declare (`throws`) what you can't, and wrap low-level checked exceptions in a domain `RuntimeException` at boundaries rather than propagating `throws IOException` through your whole call tree.

### try / catch / finally and multi-catch

```java
try {
    risky();
} catch (IOException | SQLException e) {   // multi-catch — one block, several types
    log(e);
} finally {
    cleanup();   // runs on any exit path: normal, exception, or return
}
```

Rules the exams probe: a `catch` for a more specific type must come **before** a broader one (unreachable catch is a compile error). `finally` runs even after a `return` in the `try`/`catch` (and a `return` in `finally` will override one from the `try` — don't do that). The only thing that skips `finally` is `System.exit()` or a JVM crash.

### try-with-resources = C# `using`

```java
try (var reader = Files.newBufferedReader(Path.of("x.txt"))) {
    return reader.readLine();
}   // reader.close() called automatically — reader implements AutoCloseable
```

`AutoCloseable` ≈ `IDisposable`; try-with-resources ≈ `using`. Multiple resources are separated by `;` and closed in **reverse** order. If both the body and `close()` throw, the body's exception wins and the `close()` exception is attached as a **suppressed** exception (retrievable via `getSuppressed()`).

### Throwing and custom exceptions

```java
throw new IllegalArgumentException("amount must be positive");

class InsufficientFundsException extends RuntimeException {   // unchecked by choice
    InsufficientFundsException(String msg) { super(msg); }
}
```

### Exercise 12.1
Write a method that reads the first line of a file with try-with-resources and returns it. Handle the missing-file case gracefully (return a default rather than crashing), and log the stack trace with `e.printStackTrace()` so you see the checked-exception flow.

---

## Chapter 13 — Optional, Null, and Defensive Code

Java has **no** nullable-reference-type compiler feature like C#'s `string?`. Instead, the idiom for "might be absent" return values is `Optional<T>`.

```java
Optional<String> findUser(int id) {
    if (id == 1) return Optional.of("Denis");
    return Optional.empty();
}

// consuming it
findUser(1)
    .map(String::toUpperCase)
    .ifPresentOrElse(
        name -> System.out.println("Found " + name),
        () -> System.out.println("Not found")
    );

String name = findUser(2).orElse("guest");
```

Important idiom rules:
- Use `Optional` for **return types**, not fields or parameters.
- Never call `.get()` without checking `.isPresent()` — it throws `NoSuchElementException` when empty. Prefer `.orElse`, `.orElseGet`, `.orElseThrow`, `.map`, `.filter`, `.ifPresent`.
- `Optional.of(null)` throws immediately; use `Optional.ofNullable(x)` when `x` might be null.
- For primitives there's `OptionalInt`/`OptionalLong`/`OptionalDouble` (used by primitive streams).
- `Optional` does not protect you from `null` everywhere — Java references can still be `null`, and `NullPointerException` (the "NPE") is the most common Java runtime error. Stay disciplined; `Objects.requireNonNull(x)` at boundaries documents and enforces intent.

### Exercise 13.1
Rewrite Exercise 11.2 (farthest point) so the caller uses `.orElseThrow()` to get a meaningful exception when the list is empty.

---

## Chapter 14 — Lambdas & Functional Interfaces

Java lambdas map to **functional interfaces** — any interface with exactly one abstract method (a "SAM" type). C# has `Func<>`/`Action<>`/delegates; Java has a set of built-in functional interfaces in `java.util.function`:

| C# | Java |
|---|---|
| `Func<T, R>` | `Function<T, R>` |
| `Func<T, bool>` | `Predicate<T>` |
| `Action<T>` | `Consumer<T>` |
| `Func<T>` | `Supplier<T>` |
| `Action` | `Runnable` |
| `Func<T1,T2,R>` | `BiFunction<T1,T2,R>` |
| `Func<T,T,T>` | `BinaryOperator<T>` |

```java
Function<Integer, Integer> square = x -> x * x;
Predicate<String> isLong = s -> s.length() > 5;
Supplier<String> greet = () -> "hi";
BiFunction<Integer, Integer, Integer> add = (a, b) -> a + b;

square.apply(5);       // 25 — note: .apply(), not direct invoke
isLong.test("hello");  // false
greet.get();           // "hi"
add.apply(2, 3);       // 5
```

Note the awkwardness vs C#: you invoke via a named method (`apply`/`test`/`get`/`accept`) rather than calling the variable directly like a C# delegate — it's an interface method under the hood. The optional `@FunctionalInterface` annotation makes the compiler enforce the single-abstract-method rule on your own interfaces.

**Syntax notes:** parentheses are optional for a single untyped parameter (`a -> ...`), required for zero or multiple (`() -> ...`, `(a, b) -> ...`) or when you write the type (`(Animal a) -> ...`). A block body needs `{ }` and an explicit `return`. Lambdas capture `this`, instance fields, statics, and **effectively-final** locals (like C#, but the local must not be reassigned).

**Method references** (`String::toUpperCase`, `System.out::println`, `Person::new`) ≈ C# method groups — four flavors: static (`Integer::parseInt`), bound instance (`obj::method`), unbound instance (`String::length`), and constructor (`ArrayList::new`).

### Exercise 14.1
Write a method `<T> List<T> filter(List<T> in, Predicate<T> pred)` that returns matching elements (reimplement `.filter` by hand to feel how `Predicate` works). Then call it with a lambda and with a method reference.

---

## Chapter 15 — Nested & Anonymous Classes

Java has four kinds of nested class. C# has nested classes too, but Java's *inner* (non-static) classes and *anonymous* classes behave differently enough to warrant their own chapter — and you'll meet anonymous classes constantly in older code and callback APIs.

### Static nested classes

Declared `static` inside another class. It's just a namespaced class with no link to an outer instance — the common, low-surprise case:

```java
class Outer {
    static class Node { int value; Node next; }   // no reference to Outer
}
new Outer.Node();   // no Outer instance needed
```

### Inner (non-static) classes

A non-static nested class holds an **implicit reference to an instance of the outer class**, so it can read the outer object's private fields. You need an outer instance to create one, and you reach the outer object with `Outer.this`:

```java
class Bank {
    private double rate;
    class Account {                     // inner: tied to a Bank instance
        double interest(double bal) { return bal * rate; }   // sees Bank.this.rate
    }
}
Bank bank = new Bank();
Bank.Account acc = bank.new Account();  // unusual syntax: outerInstance.new Inner()
```

### Local classes

A class declared inside a method — visible only there. It can capture the method's **effectively-final** locals (same rule as lambdas), because an instance may outlive the method call on the heap while the locals live on the stack.

### Anonymous classes

Declare *and* instantiate an unnamed class in one expression — the pre-lambda way to supply a one-off implementation of an interface or abstract class:

```java
Runnable r = new Runnable() {
    @Override public void run() { System.out.println("running"); }
};

button.addListener(new ClickListener() {
    @Override public void onClick(Event e) { handle(e); }
});
```

Since Java 8, a lambda replaces an anonymous class **when the target is a functional interface** (one abstract method). So `Runnable r = () -> System.out.println("running");` is the modern form. You still reach for an anonymous class when the type has **multiple** abstract methods, when you need instance fields/state, or when you must call `super`. (One subtle difference: inside a lambda, `this` refers to the enclosing instance; inside an anonymous class, `this` refers to the anonymous object itself.)

### Exercise 15.1
Implement a `Comparator<String>` that orders by string length, first as an **anonymous class**, then as a **lambda**. Sort the same list with each and confirm identical output.

### Exercise 15.2
Write a class `Counter` with an inner class `Tick` whose method increments and returns the outer counter's private field. Demonstrate `outer.new Tick()`.

---

## Chapter 16 — Pattern Matching & Sealed Types

Modern Java (17+) has pattern matching that will feel familiar from C#'s switch expressions.

### instanceof pattern

```java
Object o = "hello";
if (o instanceof String s) {        // binds s if match — like C# 'is string s'
    System.out.println(s.length());
}
if (o instanceof String s && s.length() > 3) { ... }   // guard in the same condition
```

This is why the `equals` override in Chapter 8 used `if (!(o instanceof Point2D p)) return false;` — the pattern both tests the type and binds the cast in one step.

### switch expressions & patterns

```java
String describe(Shape shape) {
    return switch (shape) {
        case Circle c    -> "circle r=" + c.radius();
        case Rectangle r -> "rect " + r.width() + "x" + r.height();
        case null        -> "no shape";           // switch can match null explicitly
        default          -> "unknown";
    };
}
```

### Record deconstruction

Patterns can destructure records (like C# positional patterns):

```java
case Circle(double r) -> Math.PI * r * r;
```

### Sealed types = exhaustive hierarchies

`sealed` restricts which types may extend/implement — like C#'s pattern for closed hierarchies. Combined with records and switch, you get exhaustive matching with no `default` needed:

```java
sealed interface Shape permits Circle, Rectangle {}
record Circle(double radius) implements Shape {}
record Rectangle(double w, double h) implements Shape {}

double area(Shape s) {
    return switch (s) {            // compiler knows it's exhaustive
        case Circle c    -> Math.PI * c.radius() * c.radius();
        case Rectangle r -> r.w() * r.h();
    };
}
```

If someone later adds a third permitted type, this switch **stops compiling** until you handle it — the exhaustiveness guarantee. This `sealed` + `record` + `switch` combo is the modern idiomatic Java way to model algebraic-data-type-style domains — closer to F# discriminated unions than to classic OOP.

### Exercise 16.1
Model a tiny expression tree: `sealed interface Expr` permitting `Num(double)`, `Add(Expr, Expr)`, `Mul(Expr, Expr)` (all records). Write an `eval(Expr)` using a switch expression with record-deconstruction patterns. Evaluate `Add(Num(2), Mul(Num(3), Num(4)))` → `14`.

---

## Chapter 17 — Dates, Times & Numbers

### java.time (the modern API)

Forget the legacy `java.util.Date`/`Calendar` — they're mutable, confusingly designed, and only kept for old code. Use **`java.time`** (JSR-310), which is immutable and fluent, much like C#'s `DateTime`/`DateTimeOffset` but split into focused types:

| Type | Holds |
|---|---|
| `LocalDate` | date only (no time, no zone) |
| `LocalTime` | time only |
| `LocalDateTime` | date + time, no zone |
| `ZonedDateTime` | date + time + zone |
| `Instant` | a point on the UTC timeline (≈ epoch millis) |

```java
LocalDate today = LocalDate.now();
LocalDate d = LocalDate.of(2026, 7, 16);
LocalDateTime dt = LocalDateTime.of(2026, 7, 16, 9, 30);
ZonedDateTime z = ZonedDateTime.now(ZoneId.of("Europe/Sofia"));

LocalDate next = today.plusDays(10).minusMonths(1);   // immutable → returns new objects
boolean before = d.isBefore(today);
```

> ⚠️ **Exam trap.** These types are **immutable**: `today.plusDays(1);` without assigning the result does nothing. Every "mutation" returns a new instance — just like `String`.

**Period vs Duration:** `Period` is date-based (years/months/days), `Duration` is time-based (hours/seconds/nanos). `Period` doesn't chain and doesn't apply to `LocalTime`; `Duration` doesn't apply to `LocalDate`.

```java
Period age = Period.between(birthDate, LocalDate.now());   // e.g. 34 years, 2 months
Duration d = Duration.ofMinutes(90);
```

**Formatting/parsing** with `DateTimeFormatter`:

```java
var fmt = DateTimeFormatter.ofPattern("dd MMM yyyy");
String s = today.format(fmt);           // "16 Jul 2026"
LocalDate back = LocalDate.parse("16 Jul 2026", fmt);
```

Pattern letters: `y` year, `M` month (`MMM` = Jul, `MMMM` = July), `d` day, `H`/`h` hour, `m` minute, `s` second. Formatting a `LocalTime` with a date pattern (or vice-versa) throws — the field must exist on the value.

### Numbers with precision

- **`BigDecimal` for money.** `double`/`float` are binary floating point and can't represent `0.1` exactly — never use them for currency. `new BigDecimal("0.10").add(new BigDecimal("0.20"))` is exactly `0.30`. Construct from a **String**, not a `double`, to avoid inheriting the float error, and specify a `RoundingMode` when dividing.
- **`BigInteger`** for arbitrary-precision integers (no overflow).
- **`NumberFormat`** / `DecimalFormat` for locale-aware number and currency formatting; `Integer.parseInt` / `Double.parseDouble` to parse.

### Exercise 17.1
Compute a person's age in whole years from a hardcoded birth `LocalDate` using `Period.between`. Then format today's date three ways with different `DateTimeFormatter` patterns.

### Exercise 17.2
Sum a list of prices given as `String`s (e.g. `"19.99"`, `"0.01"`) using `BigDecimal`, and print the total formatted as currency with `NumberFormat.getCurrencyInstance()`. Compare against doing the same with `double` and observe the rounding error.

---

## Chapter 18 — Files & I/O (NIO.2)

Java has two I/O generations: the old `java.io` streams (`FileInputStream`, `BufferedReader`, …) and the modern **NIO.2** (`java.nio.file`, Java 7+). Learn NIO.2 first — it's what you'll write.

### Paths

`Path` is a file/directory location (not the file itself). `Paths.get(...)` or `Path.of(...)` build one; `Files` operates on them:

```java
Path p = Path.of("data", "expenses.csv");   // platform-independent separators
p.getFileName();  p.getParent();  p.resolve("sub"); // join
p.toAbsolutePath();  p.normalize();                  // strip . and ..
```

### Reading and writing

```java
// small files, whole content:
String text = Files.readString(p);
List<String> lines = Files.readAllLines(p);       // loads all into memory
Files.writeString(p, "hello");
Files.write(p, lines);

// large files, lazily, as a stream (close it!):
try (Stream<String> stream = Files.lines(p)) {
    long count = stream.filter(l -> !l.isBlank()).count();
}

// buffered reader/writer (try-with-resources — Chapter 12):
try (BufferedWriter w = Files.newBufferedWriter(p)) {
    w.write("id,amount");
    w.newLine();
}
```

`Files.lines` returns a lazy `Stream<String>` backed by an open file handle — **use it in try-with-resources** so the handle closes. `readAllLines` reads everything into memory (fine for small files, dangerous for huge ones).

### File operations

`Files.exists(p)`, `Files.createDirectories(p)` (makes missing parents; `createDirectory` fails if parents are absent), `Files.copy`, `Files.move`, `Files.delete` (throws if absent) vs `Files.deleteIfExists`, `Files.size`, `Files.walk(p)` to traverse a tree lazily.

### Byte vs character streams (background)

The old API pairs are worth recognizing: `InputStream`/`OutputStream` move **bytes** (binary), `Reader`/`Writer` move **characters** (text, charset-decoded). "High-level" streams wrap "low-level" ones to add buffering (`BufferedReader`) or convenience (`PrintWriter`, with `print`/`println`/`printf`). Always specify a `Charset` (default `UTF-8` on modern JDKs) for text.

### Serialization (know it exists)

`implements Serializable` (a marker interface, no methods) lets an object be written to bytes; `transient` fields are skipped, as are `static` fields. It's brittle across class changes (guard with `serialVersionUID`) — for real persistence/interchange prefer JSON (Jackson/Gson) or a database. You mainly need to recognize it.

### Exercise 18.1
Write a small CSV of expenses (`id,description,amount` per line) with `Files.write`, then read it back with `Files.lines`, parse each line into a record, and print the total. Handle the header row and skip blank lines. Wrap the stream in try-with-resources.

---

## Chapter 19 — Concurrency: Threads to Virtual Threads

Java's concurrency toolkit is bigger and lower-level than C#'s `async`/`await`, but the modern parts (virtual threads, `CompletableFuture`) get you similar ergonomics. Learn the classic pieces first — you'll see them everywhere — then the modern approach.

### Threads and the shared-state problem

```java
Runnable task = () -> System.out.println("hi from " + Thread.currentThread());
Thread t = new Thread(task);
t.start();     // start(), NOT run() — run() would execute on the current thread
t.join();      // wait for it to finish
```

Each thread has its own stack; they share the heap. The moment two threads touch the same mutable field, you have a **race condition**:

```java
int count = 0;
// count++ is read-modify-write; interleaving two threads loses increments
```

### synchronized, volatile, atomic

- **`synchronized`** — every object has an intrinsic lock (monitor). A `synchronized` method or block lets only one thread in at a time. A `synchronized` *instance* method locks `this`; a `synchronized static` method locks the *class* object.

  ```java
  synchronized void inc() { count++; }
  void inc2() { synchronized (lock) { count++; } }   // block form, explicit lock object
  ```

- **`volatile`** — guarantees **visibility** (a write by one thread is seen by others; reads come from main memory, not a cached register) and prevents instruction reordering around it. It does **not** make compound actions like `count++` atomic.

- **Atomics** — `AtomicInteger`, `AtomicLong`, `AtomicReference` do lock-free compare-and-swap:

  ```java
  AtomicInteger count = new AtomicInteger();
  count.incrementAndGet();   // atomic, no synchronized needed
  ```

- **`wait()`/`notify()`/`notifyAll()`** (inherited from `Object`) are the low-level signaling primitives — must be called while holding the object's lock. You rarely write these directly now; prefer higher-level tools.

- **Deadlock** — two threads each holding a lock the other wants. Avoid by acquiring locks in a consistent global order, or use `ReentrantLock.tryLock()` with a timeout.

### java.util.concurrent (what you actually use)

- **`ExecutorService`** decouples *task submission* from *thread management*:

  ```java
  ExecutorService pool = Executors.newFixedThreadPool(4);
  Future<Integer> f = pool.submit(() -> expensiveCompute());   // Callable returns a value
  Integer result = f.get();     // blocks until done
  pool.shutdown();              // stop accepting tasks; app won't exit until you do this
  ```

  `Runnable` returns nothing; **`Callable<T>`** returns a value (and may throw checked exceptions). `Future<T>` ≈ `Task<T>` but you `.get()` (blocking) instead of `await`.

- **`CompletableFuture`** for async composition (closest to `await` chaining):

  ```java
  CompletableFuture
      .supplyAsync(() -> fetchData())
      .thenApply(data -> transform(data))
      .thenAccept(result -> System.out.println(result));
  ```

- **Concurrent collections** — `ConcurrentHashMap`, `CopyOnWriteArrayList`, `ConcurrentLinkedQueue` are built for multi-threaded access and far faster than wrapping a plain collection with `Collections.synchronizedMap(...)`.

- **`ReentrantLock`**/`Condition` are more flexible than `synchronized` (timed/interruptible acquisition, multiple wait sets). The **Fork/Join** framework (`RecursiveTask`, work-stealing) powers parallel streams for divide-and-conquer workloads.

### Virtual threads — the modern approach

Instead of C#-style `async` coloring your functions, Java 21+ lets you write **plain blocking code** on cheap virtual threads. You can spawn millions:

```java
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    for (int i = 0; i < 10_000; i++) {
        executor.submit(() -> {
            // blocking I/O here is fine — the JVM unmounts the virtual
            // thread from its carrier while it waits
            return fetchSomething();
        });
    }
}
```

The philosophical difference from C#: instead of `async`/`await` making non-blocking explicit in the type system, Java made blocking cheap so you can write straightforward sequential code and still scale. No function coloring.

### Exercise 19.1
Have two threads each increment a shared counter one million times. Run it first with a plain `int` (observe lost updates / wrong total), then fix it with `AtomicInteger`, then again with a `synchronized` block. Compare results.

### Exercise 19.2
Spawn 1,000 virtual threads, each sleeping 100ms then printing its index. Observe it finishes in roughly 100ms total, not 100 seconds. *(Use `Thread.sleep(100)` inside the task.)*

---

## Chapter 20 — Build Tools: Gradle

Once you go past single-file scripts, you need a build tool. **Gradle** is the modern choice (Maven is the older, XML-based alternative — you'll see both in the wild; Spring Boot supports either). *(The lecture course used Eclipse's built-in compiler and raw `javac`/`java` commands; the language content is identical, but real projects use a build tool.)*

### Why you need it
- Dependency management (like NuGet).
- Multi-file compilation, packaging into a runnable `.jar`.
- Standard project layout that IDEs and frameworks expect.

### Standard layout
```
my-app/
├── build.gradle          # like .csproj — build config + deps
├── settings.gradle       # project name / multi-module setup
└── src/
    ├── main/java/        # your code
    └── test/java/        # your tests
```

### Minimal build.gradle
```groovy
plugins {
    id 'java'
    id 'application'
}

group = 'com.denis'
version = '1.0'

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(25)
    }
}

repositories {
    mavenCentral()       // like nuget.org
}

dependencies {
    testImplementation 'org.junit.jupiter:junit-jupiter:5.11.0'
}

application {
    mainClass = 'com.denis.Main'
}

tasks.named('test') {
    useJUnitPlatform()
}
```

### Common commands
| C# / dotnet | Gradle |
|---|---|
| `dotnet build` | `./gradlew build` |
| `dotnet run` | `./gradlew run` |
| `dotnet test` | `./gradlew test` |
| `dotnet add package X` | add a line to `dependencies {}` |

`gradlew` is the **Gradle wrapper** — a checked-in script that downloads the correct Gradle version, so everyone on a team uses the same build. Always use `./gradlew`, not a globally installed `gradle`.

### Exercise 20.1
Convert one of your earlier exercises into a proper Gradle project. Add JUnit 5 and write one real unit test:

```java
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class MaxTest {
    @Test
    void findsMax() {
        assertEquals(9, max(List.of(3, 9, 1)));
    }
}
```

Run `./gradlew test`. JUnit 5 ≈ NUnit/xUnit; `@Test` ≈ `[Test]`/`[Fact]`; `assertEquals` argument order is **(expected, actual)** — the reverse of some frameworks, so watch out.

---

## Chapter 21 — Capstone Project

Build a small command-line **expense tracker** that exercises everything above. No Spring yet — pure Java + Gradle.

### Requirements
1. **Model** (records + sealed types): an `Expense` record with `id`, `description`, `amount` (`BigDecimal` — Chapter 17), `category` (an enum), and `date` (`java.time.LocalDate`).
2. **Storage**: read/write expenses from a CSV file (NIO.2 `Files`, try-with-resources, checked exceptions handled — Chapters 12 & 18).
3. **Queries** (streams): total by category (`groupingBy` + `Collectors.reducing`/`summingDouble`), the biggest expense (`Optional`), monthly totals, sorted output (`Comparator`).
4. **Equality**: rely on the record's generated `equals`/`hashCode` to dedup expenses in a `Set` (Chapter 8).
5. **CLI**: parse `args` to support `add`, `list`, `report` commands.
6. **Tests** (JUnit 5): test the query logic on an in-memory list.

### Stretch goals
- Use virtual threads to load multiple CSV files concurrently (Chapter 19).
- Add a `sealed interface Command permits AddCommand, ListCommand, ReportCommand` and pattern-match on it (Chapter 16).
- Format the report with text blocks (`"""..."""`) and `NumberFormat` currency.

When this is done and comfortable, **you're ready for Spring Boot** — you'll have the language, collections, streams, generics, build tooling, and testing all in muscle memory, so Spring's magic will layer cleanly on top instead of compounding unfamiliarity.

---

## Appendix A: Quick C# → Java Cheat Sheet

| C# | Java |
|---|---|
| `string` | `String` |
| `var` | `var` (locals only) |
| `x?.y` | no null-safe operator; use `Optional` or manual null checks |
| `x ?? y` | `x != null ? x : y` (no `??`); or `Optional.orElse` |
| `nameof(x)` | no equivalent |
| `$"{x} text"` | `String.format("%s text", x)`, `"x = " + x`, or `.formatted()` on a text block |
| `using (var r = ...)` | `try (var r = ...) { }` |
| `IDisposable` | `AutoCloseable` |
| `IEnumerable<T>` | `Iterable<T>` |
| `IComparable<T>` / `IComparer<T>` | `Comparable<T>` / `Comparator<T>` |
| `GetHashCode` / `Equals` | `hashCode()` / `equals(Object)` (override both) |
| `yield return` | no equivalent; return a `Stream` or collection |
| `readonly` | `final` |
| `const` | `static final` |
| `sealed class` (no inherit) | `final class` |
| `sealed`/DU hierarchies (closed set) | `sealed ... permits` |
| `abstract` | `abstract` |
| `virtual` (opt-in) | methods are virtual by default; `final` to opt out |
| `override` | `@Override` (annotation, optional but recommended) |
| `internal` | package-private (default, no keyword) |
| `protected` | `protected` (also grants same-package access) |
| `namespace` | `package` |
| `partial class` | no equivalent |
| `params` | varargs (`T...`) |
| `ref`/`out` | no equivalent; Java is always pass-by-value |
| `[Attribute]` | `@Annotation` |
| `Task<T>` | `Future<T>` / `CompletableFuture<T>` |
| `async`/`await` | virtual threads (write blocking code) |
| `DateTime` | `java.time.LocalDateTime` / `ZonedDateTime` / `Instant` |
| `decimal` | `BigDecimal` |
| properties | manual getters/setters (or Lombok) |
| extension methods | no equivalent (static helper methods instead) |
| `struct` (value type) | no equivalent; records are still reference types |
| operator overloading | not supported |
| `StringBuilder` | `StringBuilder` (identical idea) |

---

## Appendix B — Where This Came From & Going Deeper

This guide was expanded using materials from a friend's collection. If you want to go deeper than a fast-track allows, they're all worth your time:

**The 14-lecture "Java Basics" course** — a thorough, traditional progression (Intro/JVM → Structure → OOP → Constructors → Operators → Strings & Arrays → Flow Control → Exceptions → Dates & Numbers → Generics & Collections → Inner Classes → Lambda & Stream → IO/NIO.2 → Threads), plus a "Quick Help" cheat deck and seven practice tests. It's Eclipse- and `javac`-oriented; the language content maps 1:1 onto this guide's VS Code + Gradle setup. Great for worked examples and drills — the exercises there complement the ones here.

**OCA/OCP Java SE 7 Programmer I & II Study Guide (Exams 1Z0-803 & 1Z0-804)** — Kathy Sierra & Bert Bates (McGraw-Hill, 2015). The legendary "K&B" book. Unmatched on **language mechanics**: overriding vs. overloading rules, initialization order, access control, the exact behavior of the operators and flow control. If a subtle "does this compile?" question ever puzzles you, this is where the answer lives. The `⚠️ Exam trap` callouts in this guide are the flavor of thing it drills relentlessly.

**OCP Oracle Certified Professional Java SE 8 Programmer II Study Guide** — Jeanne Boyarsky & Scott Selikoff (Sybex/Wiley, 2016). The modern-API companion: **streams, lambdas, functional interfaces, `java.time`, NIO.2, concurrency, and generics/collections** in depth. This is the best structured reference for Chapters 8, 11, 14, 17, 18, and 19 here.

**The certification path itself** is a solid learning scaffold even if you never sit an exam: the SE 7 exams were 1Z0-803 (OCA) and 1Z0-804 (OCP); the SE 8 OCP track culminated in 1Z0-809. Oracle has since renumbered for Java 11/17/21, but the *topics* are stable, and these two books cover them thoroughly. Note both books predate records, sealed types, `var`, switch expressions, text blocks, and virtual threads — for those modern features, trust Chapters 5, 15, 16, and 19 here and the current JDK docs.

---

*Work through Chapters 1–21 in order. Budget roughly two weeks of evenings for 1–19, a day for 20–21. The new chapters (3, 4, 6, 8, 15, 17, 18) and the expanded concurrency chapter are the "traditional Java foundation" the original fast-track skipped — don't skim them; they're exactly what separates code that compiles from code that behaves. Then move to Spring Boot with confidence.*
